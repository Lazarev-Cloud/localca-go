package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// setupAPIRoutes adds API routes for the Next.js frontend
func SetupAPIRoutes(router *gin.Engine, certSvc *certificates.CertificateService, store *storage.Storage) {
	api := router.Group("/api")
	{
		// CORS middleware for API routes
		api.Use(corsMiddleware())

		// Certificate endpoints
		api.GET("/certificates", apiGetCertificatesHandler(certSvc, store))
		api.POST("/certificates", apiCreateCertificateHandler(certSvc, store))

		// CA info endpoint
		api.GET("/ca-info", apiGetCAInfoHandler(certSvc, store))

		// Certificate operations
		api.POST("/revoke", apiRevokeCertificateHandler(certSvc, store))
		api.POST("/renew", apiRenewCertificateHandler(certSvc, store))
		api.POST("/delete", apiDeleteCertificateHandler(certSvc, store))
	}
}

// corsMiddleware handles CORS for API requests
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// apiGetCertificatesHandler returns all certificates as JSON
func apiGetCertificatesHandler(certSvc *certificates.CertificateService, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// List all certificates
		certNames, err := store.ListCertificates()
		if err != nil {
			log.Printf("Failed to list certificates: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to list certificates",
			})
			return
		}

		// Get certificate details
		certificates := make([]CertificateInfo, 0, len(certNames))
		for _, name := range certNames {
			certInfo, err := getCertificateInfo(store, name)
			if err != nil {
				log.Printf("Failed to get certificate info for %s: %v", name, err)
				continue
			}
			certificates = append(certificates, certInfo)
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Certificates retrieved successfully",
			Data: map[string]interface{}{
				"certificates": certificates,
			},
		})
	}
}

// apiCreateCertificateHandler creates a new certificate via API
func apiCreateCertificateHandler(certSvc *certificates.CertificateService, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get form data
		commonName := c.PostForm("common_name")
		password := c.PostForm("password")
		isClient := c.PostForm("is_client") == "true"
		additionalDomains := c.PostForm("additional_domains")

		// Validate input
		if commonName == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Common Name is required",
			})
			return
		}

		// Process additional domains
		var domains []string
		if additionalDomains != "" {
			domains = parseCSVList(additionalDomains)
		}

		var err error
		if isClient {
			// Client certificate requires password
			if password == "" {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Message: "Password is required for client certificates",
				})
				return
			}

			// Create client certificate
			err = certSvc.CreateClientCertificate(commonName, password)
		} else {
			// Create server certificate
			err = certSvc.CreateServerCertificate(commonName, domains)
		}

		if err != nil {
			log.Printf("Failed to create certificate: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create certificate: %v", err),
			})
			return
		}

		// Return success
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Certificate created successfully",
		})
	}
}

// apiGetCAInfoHandler returns CA information as JSON
func apiGetCAInfoHandler(certSvc *certificates.CertificateService, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get CA info
		caInfo, err := getCAInfo(store)
		if err != nil {
			log.Printf("Failed to get CA info: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to get CA information",
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "CA info retrieved successfully",
			Data:    caInfo,
		})
	}
}

// apiRevokeCertificateHandler revokes a certificate via API
func apiRevokeCertificateHandler(certSvc *certificates.CertificateService, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get serial number
		serialNumber := c.PostForm("serial_number")
		if serialNumber == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Serial number is required",
			})
			return
		}

		// Find certificate by serial number
		certName, err := store.GetCertificateNameBySerial(serialNumber)
		if err != nil {
			log.Printf("Failed to find certificate with serial %s: %v", serialNumber, err)
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Certificate not found",
			})
			return
		}

		// Revoke certificate
		if err := certSvc.RevokeCertificate(certName); err != nil {
			log.Printf("Failed to revoke certificate: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to revoke certificate: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Certificate revoked successfully",
		})
	}
}

// apiRenewCertificateHandler renews a certificate via API
func apiRenewCertificateHandler(certSvc *certificates.CertificateService, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get serial number
		serialNumber := c.PostForm("serial_number")
		if serialNumber == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Serial number is required",
			})
			return
		}

		// Find certificate by serial number
		certName, err := store.GetCertificateNameBySerial(serialNumber)
		if err != nil {
			log.Printf("Failed to find certificate with serial %s: %v", serialNumber, err)
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Certificate not found",
			})
			return
		}

		// Check if it's a client certificate
		p12Path := store.GetCertificateP12Path(certName)
		isClient := false
		if _, err := os.Stat(p12Path); err == nil {
			isClient = true
		}

		// Renew certificate
		if isClient {
			err = certSvc.RenewClientCertificate(certName)
		} else {
			err = certSvc.RenewServerCertificate(certName)
		}

		if err != nil {
			log.Printf("Failed to renew certificate: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to renew certificate: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Certificate renewed successfully",
		})
	}
}

// apiDeleteCertificateHandler deletes a certificate via API
func apiDeleteCertificateHandler(certSvc *certificates.CertificateService, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get serial number
		serialNumber := c.PostForm("serial_number")
		if serialNumber == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Serial number is required",
			})
			return
		}

		// Find certificate by serial number
		certName, err := store.GetCertificateNameBySerial(serialNumber)
		if err != nil {
			log.Printf("Failed to find certificate with serial %s: %v", serialNumber, err)
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Certificate not found",
			})
			return
		}

		// Delete certificate
		if err := store.DeleteCertificate(certName); err != nil {
			log.Printf("Failed to delete certificate: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to delete certificate: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Certificate deleted successfully",
		})
	}
}

// Helper function to parse CSV list
func parseCSVList(csv string) []string {
	parts := strings.Split(csv, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
