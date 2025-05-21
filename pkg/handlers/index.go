package handlers

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// indexHandler handles the home page
func indexHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get CA info
		caName, _, organization, country, err := store.GetCAInfo()
		if err != nil {
			log.Printf("Failed to get CA info: %v", err)
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Error": "Failed to get CA information",
			})
			return
		}

		// Get CA certificate details
		caInfo, err := getCAInfo(store)
		if err != nil {
			log.Printf("Failed to get CA certificate details: %v", err)
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Error": "Failed to get CA certificate details",
			})
			return
		}

		// List all certificates
		certNames, err := store.ListCertificates()
		if err != nil {
			log.Printf("Failed to list certificates: %v", err)
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Error": "Failed to list certificates",
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

		// Render template
		c.HTML(http.StatusOK, "index.html", gin.H{
			"CAName":       caName,
			"Organization": organization,
			"Country":      country,
			"CAInfo":       caInfo,
			"Certificates": certificates,
			"CSRFToken":    c.GetString("csrf_token"),
		})
	}
}

// createCertificateHandler handles certificate creation
func createCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get form data
		commonName := c.PostForm("cn")
		password := c.PostForm("password")
		isClient := c.PostForm("client") == "on"
		additionalDomains := c.PostForm("additional_domains")

		// Validate input
		if commonName == "" {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Error": "Common Name is required",
			})
			return
		}

		// Process additional domains
		var domains []string
		if additionalDomains != "" {
			for _, domain := range strings.Split(additionalDomains, ",") {
				domain = strings.TrimSpace(domain)
				if domain != "" {
					domains = append(domains, domain)
				}
			}
		}

		var err error
		if isClient {
			// Client certificate requires password
			if password == "" {
				c.HTML(http.StatusBadRequest, "index.html", gin.H{
					"Error": "Password is required for client certificates",
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
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Error": fmt.Sprintf("Failed to create certificate: %v", err),
			})
			return
		}

		// Redirect to home page
		c.Redirect(http.StatusSeeOther, "/")
	}
}

// renewCertificateHandler handles certificate renewal
func renewCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get certificate name
		name := c.PostForm("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Certificate name is required",
			})
			return
		}

		// Check if it's a client certificate
		p12Path := store.GetCertificateP12Path(name)
		isClient := false
		if _, err := os.Stat(p12Path); err == nil {
			isClient = true
		}

		var err error
		if isClient {
			err = certSvc.RenewClientCertificate(name)
		} else {
			err = certSvc.RenewServerCertificate(name)
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

// deleteCertificateHandler handles certificate deletion
func deleteCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get certificate name
		name := c.PostForm("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Certificate name is required",
			})
			return
		}

		// Delete certificate
		if err := store.DeleteCertificate(name); err != nil {
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

// renewCAHandler handles CA certificate renewal
func renewCAHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Renew CA certificate
		if err := certSvc.RenewCA(); err != nil {
			log.Printf("Failed to renew CA certificate: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to renew CA certificate: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "CA certificate renewed successfully",
		})
	}
}

// revokeCertificateHandler handles certificate revocation
func revokeCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get certificate name
		name := c.PostForm("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Certificate name is required",
			})
			return
		}

		// Revoke certificate
		if err := certSvc.RevokeCertificate(name); err != nil {
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

// downloadCAHandler handles CA certificate download
func downloadCAHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		caPath := store.GetCAPublicCopyPath()
		c.FileAttachment(caPath, "ca.pem")
	}
}

// downloadCRLHandler handles CRL download
func downloadCRLHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		crlPath := filepath.Join(store.GetBasePath(), "ca.crl")
		if _, err := os.Stat(crlPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "CRL not found",
			})
			return
		}
		c.FileAttachment(crlPath, "ca.crl")
	}
}

// downloadCertificateHandler handles certificate download
func downloadCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		fileType := c.Param("type")

		var filePath string
		var fileName string

		switch fileType {
		case "crt":
			filePath = store.GetCertificatePath(name)
			fileName = name + ".crt"
		case "key":
			filePath = store.GetCertificateKeyPath(name)
			fileName = name + ".key"
		case "p12":
			filePath = store.GetCertificateP12Path(name)
			fileName = name + ".p12"
		case "bundle":
			filePath = store.GetCertificateBundlePath(name)
			fileName = name + ".bundle.crt"
		default:
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid file type",
			})
			return
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "File not found",
			})
			return
		}

		c.FileAttachment(filePath, fileName)
	}
}

// getCAInfo gets information about the CA certificate
func getCAInfo(store *storage.Storage) (CAInfo, error) {
	caInfo := CAInfo{}

	// Read CA certificate
	certPath := store.GetCAPublicKeyPath()
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return caInfo, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// Parse certificate
	block, _ := pem.Decode(certData)
	if block == nil {
		return caInfo, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return caInfo, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Fill in CA info
	caInfo.CommonName = cert.Subject.CommonName
	if len(cert.Subject.Organization) > 0 {
		caInfo.Organization = cert.Subject.Organization[0]
	}
	if len(cert.Subject.Country) > 0 {
		caInfo.Country = cert.Subject.Country[0]
	}
	caInfo.ExpiryDate = cert.NotAfter.Format("2006-01-02")
	caInfo.IsExpired = time.Now().After(cert.NotAfter)

	return caInfo, nil
}

// getCertificateInfo gets information about a certificate
func getCertificateInfo(store *storage.Storage, name string) (CertificateInfo, error) {
	certInfo := CertificateInfo{
		CommonName: name,
	}

	// Check if it's a client certificate
	p12Path := store.GetCertificateP12Path(name)
	certInfo.IsClient = false
	if _, err := os.Stat(p12Path); err == nil {
		certInfo.IsClient = true
	}

	// Check if certificate is revoked
	revokedPath := filepath.Join(store.GetCertificateDirectory(name), "revoked")
	if _, err := os.Stat(revokedPath); err == nil {
		certInfo.IsRevoked = true
	}

	// Read certificate
	certPath := store.GetCertificatePath(name)

	// Validate certificate path to prevent command injection
	if !filepath.IsAbs(certPath) {
		return certInfo, fmt.Errorf("certificate path must be absolute")
	}

	// Check if the file exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return certInfo, fmt.Errorf("certificate file does not exist: %w", err)
	}

	// Find OpenSSL path
	opensslPath, err := exec.LookPath("openssl")
	if err != nil {
		return certInfo, fmt.Errorf("failed to find openssl executable: %w", err)
	}

	cmd := exec.Command(
		opensslPath, "x509",
		"-in", certPath,
		"-noout",
		"-text",
	)
	output, err := cmd.Output()
	if err != nil {
		return certInfo, fmt.Errorf("failed to get certificate info: %w", err)
	}

	// Parse output
	outputStr := string(output)

	// Get serial number
	serialIndex := strings.Index(outputStr, "Serial Number:")
	if serialIndex != -1 {
		serialPart := outputStr[serialIndex:]
		endIndex := strings.Index(serialPart, "\n")
		if endIndex != -1 {
			serialLine := serialPart[:endIndex]
			certInfo.SerialNumber = strings.TrimSpace(strings.TrimPrefix(serialLine, "Serial Number:"))
		}
	}

	// Get expiry date
	expiryIndex := strings.Index(outputStr, "Not After:")
	if expiryIndex != -1 {
		expiryPart := outputStr[expiryIndex:]
		endIndex := strings.Index(expiryPart, "\n")
		if endIndex != -1 {
			expiryLine := expiryPart[:endIndex]
			expiryDateStr := strings.TrimSpace(strings.TrimPrefix(expiryLine, "Not After:"))

			// Parse date
			expiryDate, err := time.Parse("Jan 2 15:04:05 2006 MST", expiryDateStr)
			if err == nil {
				certInfo.ExpiryDate = expiryDate.Format("2006-01-02")

				// Check if expired or expiring soon
				now := time.Now()
				certInfo.IsExpired = now.After(expiryDate)
				certInfo.IsExpiringSoon = !certInfo.IsExpired && now.Add(30*24*time.Hour).After(expiryDate)
			}
		}
	}

	return certInfo, nil
}
