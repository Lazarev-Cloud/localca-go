package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/security"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// setupAPIRoutes adds API routes for the Next.js frontend
func SetupAPIRoutes(router *gin.Engine, certSvc certificates.CertificateServiceInterface, store *storage.Storage) {
	api := router.Group("/api")
	{
		// CORS middleware for API routes
		api.Use(corsMiddleware())

		// Security middleware for API routes
		api.Use(apiSecurityMiddleware())

		// Authentication endpoints
		api.POST("/login", apiLoginHandler(certSvc, store))
		api.GET("/setup", apiSetupHandler(certSvc, store))
		api.POST("/setup", apiSetupHandler(certSvc, store))
		api.GET("/auth/status", apiAuthStatusHandler(store))

		// Certificate endpoints
		api.GET("/certificates", apiGetCertificatesHandler(certSvc, store))
		api.POST("/certificates", apiCreateCertificateHandler(certSvc, store))

		// CA info endpoint
		api.GET("/ca-info", apiGetCAInfoHandler(certSvc, store))

		// Certificate operations
		api.POST("/revoke", apiRevokeCertificateHandler(certSvc, store))
		api.POST("/renew", apiRenewCertificateHandler(certSvc, store))
		api.POST("/delete", apiDeleteCertificateHandler(certSvc, store))

		// Settings endpoints
		api.GET("/settings", apiGetSettingsHandler(certSvc, store))
		api.POST("/settings", apiSaveSettingsHandler(certSvc, store))
		api.POST("/test-email", apiTestEmailHandler(certSvc, store))

		// Logout endpoint
		api.POST("/logout", apiLogoutHandler(store))

		// Download endpoints
		api.GET("/download/ca", downloadCAHandler(certSvc, store))
		api.GET("/download/crl", downloadCRLHandler(certSvc, store))
		api.GET("/download/:name/:type", downloadCertificateHandler(certSvc, store))
	}
}

// corsMiddleware handles CORS for API requests
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment variable
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			// If not specified, allow localhost development servers on common ports
			allowedOrigins = "http://localhost:3000,http://localhost:8080,https://localhost:3000,https://localhost:8080"
		}

		// Get allowed methods from environment variable
		allowedMethods := os.Getenv("CORS_ALLOWED_METHODS")
		if allowedMethods == "" {
			allowedMethods = "GET, POST, PUT, DELETE, OPTIONS" // Default fallback
		}

		// Get allowed headers from environment variable
		allowedHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
		if allowedHeaders == "" {
			allowedHeaders = "Content-Type, Authorization, X-CSRF-Token" // Default fallback
		}

		// Check if the origin is allowed
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			// If specific origins are defined (comma-separated list)
			if allowedOrigins != "*" {
				origins := strings.Split(allowedOrigins, ",")
				allowed := false
				for _, allowedOrigin := range origins {
					// Check for exact match
					if allowedOrigin == origin {
						allowed = true
						break
					}
				}
				if allowed {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				}
			} else {
				// Wildcard origin
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// apiSecurityMiddleware adds additional security headers and validation for API routes
func apiSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers specific to API
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// Rate limiting for API endpoints (basic implementation)
		// In production, you'd want a more sophisticated rate limiter
		userAgent := c.GetHeader("User-Agent")
		if userAgent == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "User-Agent header is required",
			})
			c.Abort()
			return
		}

		// Validate Content-Type for POST requests
		if c.Request.Method == "POST" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" &&
				!strings.Contains(contentType, "application/json") &&
				!strings.Contains(contentType, "application/x-www-form-urlencoded") &&
				!strings.Contains(contentType, "multipart/form-data") {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Message: "Unsupported Content-Type",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// apiGetCertificatesHandler returns all certificates as JSON
func apiGetCertificatesHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
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
func apiCreateCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate form data
		commonName := security.ValidateCommonName(c.PostForm("common_name"))
		password := security.SanitizeInput(c.PostForm("password"))
		isClient := c.PostForm("is_client") == "true"
		additionalDomains := security.SanitizeInput(c.PostForm("additional_domains"))

		// Validate input
		if commonName == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Common Name is required",
			})
			return
		}

		// Additional validation for common name
		if len(commonName) > 64 {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Common Name must be 64 characters or less",
			})
			return
		}

		// Validate password for client certificates
		if isClient && (password == "" || len(password) < 8) {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Password is required for client certificates and must be at least 8 characters",
			})
			return
		}

		// Check if certificate already exists
		existingCerts, err := store.ListCertificates()
		if err != nil {
			log.Printf("Failed to list existing certificates: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to check existing certificates",
			})
			return
		}

		for _, existingCert := range existingCerts {
			if existingCert == commonName {
				c.JSON(http.StatusConflict, APIResponse{
					Success: false,
					Message: "Certificate with this Common Name already exists",
				})
				return
			}
		}

		// Process additional domains
		var domains []string
		if additionalDomains != "" {
			domains = parseCSVList(additionalDomains)
			// Validate each domain
			for _, domain := range domains {
				if len(domain) > 255 {
					c.JSON(http.StatusBadRequest, APIResponse{
						Success: false,
						Message: "Domain names must be 255 characters or less",
					})
					return
				}
			}
		}

		var err2 error
		if isClient {
			// Create client certificate
			err2 = certSvc.CreateClientCertificate(commonName, password)
		} else {
			// Create server certificate
			err2 = certSvc.CreateServerCertificate(commonName, domains)
		}

		if err2 != nil {
			log.Printf("Failed to create certificate: %v", err2)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create certificate: %v", err2),
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
func apiGetCAInfoHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
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
func apiRevokeCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate serial number
		serialNumber := security.ValidateSerialNumber(c.PostForm("serial_number"))
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
func apiRenewCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate serial number
		serialNumber := security.ValidateSerialNumber(c.PostForm("serial_number"))
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
func apiDeleteCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate serial number
		serialNumber := security.ValidateSerialNumber(c.PostForm("serial_number"))
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

// apiGetSettingsHandler handles GET /api/settings
func apiGetSettingsHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, return mock settings data that matches the frontend interface
		settings := map[string]interface{}{
			"general": map[string]interface{}{
				"caName":       "LocalCA in.lc",
				"organization": "LocalCA",
				"country":      "US",
				"tlsEnabled":   true,
			},
			"email": map[string]interface{}{
				"emailNotify":  false,
				"smtpServer":   "",
				"smtpPort":     "25",
				"smtpUser":     "",
				"smtpPassword": "",
				"smtpUseTLS":   false,
				"emailFrom":    "",
				"emailTo":      "",
			},
			"storage": map[string]interface{}{
				"storagePath": "/app/data",
				"backupPath":  "",
				"autoBackup":  false,
			},
			"ca": map[string]interface{}{
				"caKeyPassword": "",
				"crlExpiryDays": "30",
			},
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Settings retrieved successfully",
			Data:    settings,
		})
	}
}

// apiSaveSettingsHandler handles POST /api/settings
func apiSaveSettingsHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings map[string]interface{}
		if err := c.BindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request format",
			})
			return
		}

		// For now, just return success
		// In a real implementation, you would save the settings to a config file or database
		log.Printf("Settings update request received: %+v", settings)

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Settings saved successfully",
		})
	}
}

// apiTestEmailHandler handles POST /api/test-email
func apiTestEmailHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var emailConfig map[string]interface{}
		if err := c.BindJSON(&emailConfig); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request format",
			})
			return
		}

		// For now, just simulate sending a test email
		log.Printf("Test email request received: %+v", emailConfig)

		// In a real implementation, you would actually send a test email
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Test email sent successfully",
		})
	}
}

// apiAuthStatusHandler handles authentication status checks
func apiAuthStatusHandler(store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if setup is completed
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}

		// If setup is not completed, return setup required
		if !authConfig.SetupCompleted {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Setup required",
				Data: map[string]interface{}{
					"setup_required": true,
					"authenticated":  false,
				},
			})
			return
		}

		// Check if user is authenticated
		session, err := c.Cookie("session")
		if err != nil || !validateSession(session, store) {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Authentication required",
				Data: map[string]interface{}{
					"setup_required": false,
					"authenticated":  false,
				},
			})
			return
		}

		// User is authenticated
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "User is authenticated",
			Data: map[string]interface{}{
				"setup_required": false,
				"authenticated":  true,
			},
		})
	}
}

// apiLogoutHandler handles API logout
func apiLogoutHandler(store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current session token to clean up server-side session
		sessionToken, err := c.Cookie("session")
		if err == nil && sessionToken != "" {
			// Clean up server-side session file securely
			sessionsDir := filepath.Join(store.GetBasePath(), "sessions")
			sessionFileBase := base64.URLEncoding.EncodeToString([]byte(sessionToken))
			if len(sessionFileBase) > 100 {
				sessionFileBase = sessionFileBase[:100]
			}
			sessionFile := filepath.Join(sessionsDir, sessionFileBase)

			// Validate the session file path before deletion
			if strings.HasPrefix(sessionFile, sessionsDir) {
				if err := os.Remove(sessionFile); err != nil && !os.IsNotExist(err) {
					log.Printf("Failed to remove session file: %v", err)
				}
			}
		}

		// Clear session cookie with secure parameters
		c.SetCookie(
			"session",
			"",
			-1, // expire immediately
			"/",
			"",                   // domain - empty for current domain
			c.Request.TLS != nil, // secure if using HTTPS
			true,                 // httpOnly
		)

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Logout successful",
		})
	}
}
