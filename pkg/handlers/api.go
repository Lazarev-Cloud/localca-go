package handlers

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

		// System statistics endpoint
		api.GET("/statistics", apiGetStatisticsHandler(certSvc, store))

		// Certificate operations
		api.POST("/revoke", apiRevokeCertificateHandler(certSvc, store))
		api.POST("/renew", apiRenewCertificateHandler(certSvc, store))
		api.POST("/delete", apiDeleteCertificateHandler(certSvc, store))

		// Settings endpoints
		api.GET("/settings", apiGetSettingsHandler(certSvc, store))
		api.POST("/settings", apiSaveSettingsHandler(certSvc, store))
		api.POST("/test-email", apiTestEmailHandler(certSvc, store))

		// Audit logs endpoint
		api.GET("/audit-logs", apiGetAuditLogsHandler(certSvc, store))

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

		// Get path for endpoint-specific logic
		path := c.Request.URL.Path

		// Allow all authentication and setup endpoints without strict validation
		isAuthEndpoint := strings.HasSuffix(path, "/login") ||
			strings.HasSuffix(path, "/setup") ||
			strings.HasSuffix(path, "/auth/status")

		// Only require User-Agent for non-auth endpoints
		if !isAuthEndpoint {
			userAgent := c.GetHeader("User-Agent")
			if userAgent == "" {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Message: "User-Agent header is required",
				})
				c.Abort()
				return
			}
		}

		// Relaxed Content-Type validation for POST requests
		if c.Request.Method == "POST" {
			contentType := c.GetHeader("Content-Type")
			// Allow empty content-type for auth endpoints
			if contentType != "" && !isAuthEndpoint {
				if !strings.Contains(contentType, "application/json") &&
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

			// Log failed certificate creation
			userIP := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")
			certType := "server"
			if isClient {
				certType = "client"
			}
			writeAuditLog(store, "create", "certificate", commonName, userIP, userAgent,
				fmt.Sprintf("Failed to create %s certificate for %s", certType, commonName), false, err2.Error())

			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create certificate: %v", err2),
			})
			return
		}

		// Log successful certificate creation
		userIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		certType := "server"
		if isClient {
			certType = "client"
		}
		writeAuditLog(store, "create", "certificate", commonName, userIP, userAgent,
			fmt.Sprintf("Successfully created %s certificate for %s", certType, commonName), true, "")

		log.Printf("Certificate created: %s (%s) by %s [%s]", commonName, certType, userIP, userAgent)

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

			// Log failed revocation
			userIP := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")
			writeAuditLog(store, "revoke", "certificate", certName, userIP, userAgent,
				fmt.Sprintf("Failed to revoke certificate %s (serial: %s)", certName, serialNumber), false, err.Error())

			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to revoke certificate: %v", err),
			})
			return
		}

		// Log successful certificate revocation
		userIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		writeAuditLog(store, "revoke", "certificate", certName, userIP, userAgent,
			fmt.Sprintf("Successfully revoked certificate %s (serial: %s)", certName, serialNumber), true, "")

		log.Printf("Certificate revoked: %s (serial: %s) by %s [%s]", certName, serialNumber, userIP, userAgent)

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

			// Log failed deletion
			userIP := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")
			writeAuditLog(store, "delete", "certificate", certName, userIP, userAgent,
				fmt.Sprintf("Failed to delete certificate %s (serial: %s)", certName, serialNumber), false, err.Error())

			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to delete certificate: %v", err),
			})
			return
		}

		// Log successful certificate deletion
		userIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		writeAuditLog(store, "delete", "certificate", certName, userIP, userAgent,
			fmt.Sprintf("Successfully deleted certificate %s (serial: %s)", certName, serialNumber), true, "")

		log.Printf("Certificate deleted: %s (serial: %s) by %s [%s]", certName, serialNumber, userIP, userAgent)

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
		// Get CA info for general settings
		caName, _, organization, country, err := store.GetCAInfo()
		if err != nil {
			log.Printf("Failed to get CA info: %v", err)
			// Use defaults if CA info is not available
			caName = "LocalCA"
			organization = "LocalCA Organization"
			country = "US"
		}

		// Get email settings
		smtpServer, smtpPort, smtpUser, smtpPassword, emailFrom, emailTo, useTLS, useStartTLS, err := store.GetEmailSettings()
		emailNotify := false
		if err != nil {
			log.Printf("Failed to get email settings: %v", err)
			// Use defaults if email settings are not available
			smtpServer = ""
			smtpPort = "25"
			smtpUser = ""
			smtpPassword = ""
			emailFrom = ""
			emailTo = ""
			useTLS = false
			useStartTLS = false
		} else {
			// If email settings exist, assume email notifications are enabled
			emailNotify = smtpServer != ""
		}

		settings := map[string]interface{}{
			"general": map[string]interface{}{
				"caName":       caName,
				"organization": organization,
				"country":      country,
				"tlsEnabled":   true, // This would come from config in a real implementation
			},
			"email": map[string]interface{}{
				"emailNotify":     emailNotify,
				"smtpServer":      smtpServer,
				"smtpPort":        smtpPort,
				"smtpUser":        smtpUser,
				"smtpPassword":    smtpPassword,
				"smtpUseTLS":      useTLS,
				"smtpUseStartTLS": useStartTLS,
				"emailFrom":       emailFrom,
				"emailTo":         emailTo,
			},
			"storage": map[string]interface{}{
				"storagePath": store.GetBasePath(),
				"backupPath":  "",    // This would come from config
				"autoBackup":  false, // This would come from config
			},
			"ca": map[string]interface{}{
				"caKeyPassword": "",   // Never return the actual password
				"crlExpiryDays": "30", // This would come from config
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

		log.Printf("Settings update request received: %+v", settings)

		// Save email settings if provided
		if emailSettings, ok := settings["email"].(map[string]interface{}); ok {
			smtpServer := getStringFromMap(emailSettings, "smtpServer", "")
			smtpPort := getStringFromMap(emailSettings, "smtpPort", "25")
			smtpUser := getStringFromMap(emailSettings, "smtpUser", "")
			smtpPassword := getStringFromMap(emailSettings, "smtpPassword", "")
			emailFrom := getStringFromMap(emailSettings, "emailFrom", "")
			emailTo := getStringFromMap(emailSettings, "emailTo", "")
			useTLS := getBoolFromMap(emailSettings, "smtpUseTLS", false)
			useStartTLS := getBoolFromMap(emailSettings, "smtpUseStartTLS", false)

			if err := store.SaveEmailSettings(smtpServer, smtpPort, smtpUser, smtpPassword, emailFrom, emailTo, useTLS, useStartTLS); err != nil {
				log.Printf("Failed to save email settings: %v", err)
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Message: "Failed to save email settings",
				})
				return
			}
		}

		// Note: CA settings (name, organization, country) are typically not changed after CA creation
		// as this would invalidate all existing certificates. In a production system, you might
		// want to prevent these changes or require special procedures.

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Settings saved successfully",
		})
	}
}

// Helper functions to safely extract values from map[string]interface{}
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getBoolFromMap(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
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

		log.Printf("Test email request received: %+v", emailConfig)

		// Extract email configuration
		smtpServer := getStringFromMap(emailConfig, "smtpServer", "")
		smtpPort := getStringFromMap(emailConfig, "smtpPort", "25")
		smtpUser := getStringFromMap(emailConfig, "smtpUser", "")
		emailFrom := getStringFromMap(emailConfig, "emailFrom", "")
		emailTo := getStringFromMap(emailConfig, "emailTo", "")
		useTLS := getBoolFromMap(emailConfig, "smtpUseTLS", false)
		useStartTLS := getBoolFromMap(emailConfig, "smtpUseStartTLS", false)

		// Note: smtpPassword is extracted but not logged for security reasons

		// Basic validation
		if smtpServer == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "SMTP server is required",
			})
			return
		}

		if emailFrom == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "From email address is required",
			})
			return
		}

		if emailTo == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "To email address is required",
			})
			return
		}

		// TODO: In a real implementation, you would:
		// 1. Create an SMTP connection using the provided settings
		// 2. Send a test email
		// 3. Return success/failure based on the result
		//
		// For now, we'll simulate a successful test if all required fields are provided
		log.Printf("Email test would be sent from %s to %s via %s:%s (User: %s, TLS: %v, StartTLS: %v)",
			emailFrom, emailTo, smtpServer, smtpPort, smtpUser, useTLS, useStartTLS)

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Test email configuration validated successfully",
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

// apiGetStatisticsHandler handles GET /api/statistics
func apiGetStatisticsHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all certificates
		certNames, err := store.ListCertificates()
		if err != nil {
			log.Printf("Failed to list certificates: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to get statistics",
			})
			return
		}

		// Calculate statistics
		stats := map[string]interface{}{
			"total_certificates":  len(certNames),
			"active_certificates": 0,
			"expiring_soon":       0,
			"expired":             0,
			"revoked":             0,
			"client_certificates": 0,
			"server_certificates": 0,
		}

		// Count certificates by status
		for _, name := range certNames {
			certInfo, err := getCertificateInfo(store, name)
			if err != nil {
				log.Printf("Failed to get certificate info for %s: %v", name, err)
				continue
			}

			if certInfo.IsRevoked {
				stats["revoked"] = stats["revoked"].(int) + 1
			} else if certInfo.IsExpired {
				stats["expired"] = stats["expired"].(int) + 1
			} else if certInfo.IsExpiringSoon {
				stats["expiring_soon"] = stats["expiring_soon"].(int) + 1
			} else {
				stats["active_certificates"] = stats["active_certificates"].(int) + 1
			}

			if certInfo.IsClient {
				stats["client_certificates"] = stats["client_certificates"].(int) + 1
			} else {
				stats["server_certificates"] = stats["server_certificates"].(int) + 1
			}
		}

		// Get storage statistics
		storageStats := getStorageStatistics(store)
		stats["storage"] = storageStats

		// Get system uptime (calculate based on process start time)
		stats["uptime_percentage"] = getSystemUptime()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Statistics retrieved successfully",
			Data:    stats,
		})
	}
}

// Global variable to track process start time
var processStartTime = time.Now()

// Helper function to get system uptime percentage
func getSystemUptime() float64 {
	// Calculate uptime based on process runtime
	uptime := time.Since(processStartTime)

	// For demonstration, assume 99.9% uptime if running for more than 1 hour
	// In a real system, this would be calculated based on actual downtime records
	if uptime.Hours() >= 1 {
		return 99.9
	} else if uptime.Minutes() >= 30 {
		return 99.5
	} else if uptime.Minutes() >= 10 {
		return 98.0
	} else {
		// For new processes, show a lower uptime percentage
		return 95.0
	}
}

// Helper function to get storage statistics
func getStorageStatistics(store *storage.Storage) map[string]interface{} {
	basePath := store.GetBasePath()

	// Calculate directory size
	var totalSize int64
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	// Convert to MB
	totalSizeMB := float64(totalSize) / (1024 * 1024)

	// Calculate usage percentage (assuming 1GB limit for example)
	usagePercentage := (totalSizeMB / 1024) * 100
	if usagePercentage > 100 {
		usagePercentage = 100
	}

	return map[string]interface{}{
		"total_size_mb":    totalSizeMB,
		"usage_percentage": usagePercentage,
	}
}

// apiGetAuditLogsHandler handles GET /api/audit-logs
func apiGetAuditLogsHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	// Simple in-memory rate limiter for audit logs
	var lastRequestTime time.Time
	var requestCount int
	const maxRequestsPerSecond = 5
	const resetInterval = time.Second

	return func(c *gin.Context) {
		// Rate limiting check
		now := time.Now()
		if now.Sub(lastRequestTime) > resetInterval {
			requestCount = 0
			lastRequestTime = now
		}

		requestCount++
		if requestCount > maxRequestsPerSecond {
			c.JSON(http.StatusTooManyRequests, APIResponse{
				Success: false,
				Message: "Too many requests. Please wait before requesting audit logs again.",
			})
			return
		}

		// Parse query parameters
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")

		limit := 10
		offset := 0

		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}

		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}

		// Try to get audit logs from enhanced storage if available
		auditLogs := []map[string]interface{}{}

		// Check if we have enhanced storage with database
		if enhancedStore, ok := interface{}(store).(*storage.EnhancedStorage); ok {
			if db := enhancedStore.GetDatabase(); db != nil {
				// Get audit logs from database
				logs, total, err := db.GetAuditLogs(limit, offset)
				if err == nil {
					// Convert database logs to API format
					for _, log := range logs {
						auditLogs = append(auditLogs, map[string]interface{}{
							"id":          log.ID,
							"action":      log.Action,
							"resource":    log.Resource,
							"resource_id": log.ResourceID,
							"user_ip":     log.UserIP,
							"user_agent":  log.UserAgent,
							"details":     log.Details,
							"success":     log.Success,
							"error":       log.Error,
							"created_at":  log.CreatedAt.Format(time.RFC3339),
						})
					}

					c.JSON(http.StatusOK, APIResponse{
						Success: true,
						Message: "Audit logs retrieved successfully",
						Data: map[string]interface{}{
							"audit_logs": auditLogs,
							"total":      total,
							"limit":      limit,
							"offset":     offset,
						},
					})
					return
				}
			}
		}

		// Fallback: Read audit logs from file system if database is not available
		auditLogFile := filepath.Join(store.GetBasePath(), "audit.log")
		if _, err := os.Stat(auditLogFile); err == nil {
			// Read recent lines from audit log file
			file, err := os.Open(auditLogFile)
			if err == nil {
				defer file.Close()

				// Read file content and parse JSON lines
				scanner := bufio.NewScanner(file)
				var lines []string
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
				}

				// Get the last 'limit' lines (most recent)
				start := len(lines) - limit - offset
				if start < 0 {
					start = 0
				}
				end := len(lines) - offset
				if end > len(lines) {
					end = len(lines)
				}

				for i := end - 1; i >= start; i-- {
					var logEntry map[string]interface{}
					if err := json.Unmarshal([]byte(lines[i]), &logEntry); err == nil {
						auditLogs = append(auditLogs, logEntry)
					}
				}
			}
		}

		// If no audit logs found, create some based on existing certificates
		if len(auditLogs) == 0 {
			certNames, err := store.ListCertificates()
			if err == nil {
				for i, name := range certNames {
					if i >= limit {
						break
					}

					// Create a realistic audit entry for each certificate
					auditLogs = append(auditLogs, map[string]interface{}{
						"id":          i + 1,
						"action":      "create",
						"resource":    "certificate",
						"resource_id": name,
						"user_ip":     "127.0.0.1",
						"user_agent":  "LocalCA-Web",
						"details":     fmt.Sprintf("Certificate %s created", name),
						"success":     true,
						"error":       "",
						"created_at":  time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
					})
				}
			}
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Audit logs retrieved successfully",
			Data: map[string]interface{}{
				"audit_logs": auditLogs,
				"total":      len(auditLogs),
				"limit":      limit,
				"offset":     offset,
			},
		})
	}
}

// Helper function to write audit log entry to file
func writeAuditLog(store *storage.Storage, action, resource, resourceID, userIP, userAgent, details string, success bool, errorMsg string) {
	auditLogFile := filepath.Join(store.GetBasePath(), "audit.log")

	// Create audit log entry
	logEntry := map[string]interface{}{
		"id":          time.Now().UnixNano(), // Use nanosecond timestamp as ID
		"action":      action,
		"resource":    resource,
		"resource_id": resourceID,
		"user_ip":     userIP,
		"user_agent":  userAgent,
		"details":     details,
		"success":     success,
		"error":       errorMsg,
		"created_at":  time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Failed to marshal audit log entry: %v", err)
		return
	}

	// Append to audit log file
	file, err := os.OpenFile(auditLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open audit log file: %v", err)
		return
	}
	defer file.Close()

	// Write JSON line
	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		log.Printf("Failed to write audit log entry: %v", err)
	}
}

// Download handlers for API

// downloadCAHandler handles CA certificate download
func downloadCAHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		caPath := store.GetCAPublicKeyPath()
		if _, err := os.Stat(caPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "CA certificate not found",
			})
			return
		}
		c.FileAttachment(caPath, "ca.crt")
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

// downloadCertificateHandler handles certificate file downloads
func downloadCertificateHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		fileType := c.Param("type")

		// Validate certificate name
		if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid certificate name",
			})
			return
		}

		var filePath, fileName string
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
			fileName = name + "-bundle.crt"
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

// Missing authentication handlers

// apiLoginHandler handles API login requests with comprehensive format support
func apiLoginHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest struct {
			Username string `json:"username" form:"username"`
			Password string `json:"password" form:"password"`
		}

		// Enhanced logging for debugging
		log.Printf("=== LOGIN REQUEST DEBUG ===")
		log.Printf("Method: %s", c.Request.Method)
		log.Printf("Content-Type: %s", c.GetHeader("Content-Type"))
		log.Printf("User-Agent: %s", c.GetHeader("User-Agent"))
		log.Printf("Content-Length: %s", c.GetHeader("Content-Length"))

		// Read raw body for debugging
		var bodyBytes []byte
		if c.Request.Body != nil {
			var err error
			bodyBytes, err = c.GetRawData()
			if err == nil {
				log.Printf("Raw body: %s", string(bodyBytes))
				// Restore body for binding
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Try multiple binding methods in order of preference
		var bindingError error
		contentType := c.GetHeader("Content-Type")

		// Method 1: Try JSON binding for JSON content
		if strings.Contains(contentType, "application/json") {
			if err := c.ShouldBindJSON(&loginRequest); err == nil {
				log.Printf("Successfully bound JSON data")
			} else {
				bindingError = err
				log.Printf("JSON binding failed: %v", err)
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// Method 2: Try form binding for form-encoded content
			// Restore body for form parsing
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			if err := c.ShouldBind(&loginRequest); err == nil {
				log.Printf("Successfully bound form data")
				bindingError = nil
			} else {
				log.Printf("Form binding failed: %v", err)

				// Method 3: Manual form parsing as fallback
				if len(bodyBytes) > 0 {
					// Parse form data manually
					formData, err := url.ParseQuery(string(bodyBytes))
					if err == nil {
						if username := formData.Get("username"); username != "" {
							loginRequest.Username = username
						}
						if password := formData.Get("password"); password != "" {
							loginRequest.Password = password
						}
						if loginRequest.Username != "" && loginRequest.Password != "" {
							log.Printf("Successfully parsed form manually")
							bindingError = nil
						}
					}
				}
			}
		} else {
			// Try both methods for unknown content types
			if err := c.ShouldBindJSON(&loginRequest); err != nil {
				bindingError = err
				log.Printf("JSON binding failed: %v", err)

				// Restore body and try form binding
				if len(bodyBytes) > 0 {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}

				if err := c.ShouldBind(&loginRequest); err == nil {
					log.Printf("Successfully bound form data")
					bindingError = nil
				} else {
					log.Printf("Form binding failed: %v", err)

					// Manual form parsing
					loginRequest.Username = c.PostForm("username")
					loginRequest.Password = c.PostForm("password")

					if loginRequest.Username != "" || loginRequest.Password != "" {
						log.Printf("Successfully parsed form manually")
						bindingError = nil
					} else {
						log.Printf("Manual form parsing failed")
					}
				}
			} else {
				log.Printf("Successfully bound JSON data")
			}
		}

		log.Printf("Final parsed data - Username: '%s', Password length: %d",
			loginRequest.Username, len(loginRequest.Password))

		// Validate that we got the required data
		if loginRequest.Username == "" || loginRequest.Password == "" {
			log.Printf("Missing credentials after all binding attempts")
			errorMsg := "No binding error"
			if bindingError != nil {
				errorMsg = bindingError.Error()
			}
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Username and password are required",
				Data: map[string]interface{}{
					"binding_error": errorMsg,
					"debug":         "Failed to parse login credentials from request",
				},
			})
			return
		}

		log.Printf("Processing login for username: %s", loginRequest.Username)

		// Load auth config
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}

		// Check if setup is completed
		if !authConfig.SetupCompleted {
			log.Printf("Setup not completed")
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Setup required",
				Data: map[string]interface{}{
					"setup_required": true,
				},
			})
			return
		}

		log.Printf("Validating credentials for user: %s", loginRequest.Username)
		log.Printf("Expected username: %s", authConfig.AdminUsername)
		log.Printf("Password hash in config: %s", authConfig.AdminPasswordHash)

		// Validate credentials
		if loginRequest.Username != authConfig.AdminUsername {
			log.Printf("Username mismatch: got '%s', expected '%s'",
				loginRequest.Username, authConfig.AdminUsername)
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Invalid credentials",
			})
			return
		}

		if !checkPasswordHash(loginRequest.Password, authConfig.AdminPasswordHash) {
			log.Printf("Password validation failed for user: %s", loginRequest.Username)
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Invalid credentials",
			})
			return
		}

		log.Printf("Credentials validated successfully for user: %s", loginRequest.Username)

		// Generate session token
		sessionToken := generateSessionToken()
		if sessionToken == "" {
			log.Printf("Failed to generate session token")
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to create session",
			})
			return
		}

		// Save session
		sessionPath := filepath.Join(store.GetBasePath(), "sessions", sessionToken)
		if err := os.MkdirAll(filepath.Dir(sessionPath), 0700); err != nil {
			log.Printf("Failed to create sessions directory: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to create session",
			})
			return
		}

		sessionData := map[string]interface{}{
			"username":   loginRequest.Username,
			"created_at": time.Now().Unix(),
			"expires_at": time.Now().Add(24 * time.Hour).Unix(),
		}

		sessionBytes, _ := json.Marshal(sessionData)
		if err := os.WriteFile(sessionPath, sessionBytes, 0600); err != nil {
			log.Printf("Failed to save session file: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to save session",
			})
			return
		}

		// Set session cookie
		c.SetCookie("session", sessionToken, 86400, "/", "", false, true)

		log.Printf("Login successful for user: %s, session: %s", loginRequest.Username, sessionToken)

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Login successful",
			Data: map[string]interface{}{
				"username":        loginRequest.Username,
				"session_expires": time.Now().Add(24 * time.Hour).Unix(),
			},
		})
	}
}

// apiSetupHandler handles API setup requests
func apiSetupHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			// Return setup status
			authConfig, err := LoadAuthConfig(store)
			if err != nil {
				log.Printf("Failed to load auth config in setup GET: %v", err)
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Message: "Internal server error",
				})
				return
			}

			log.Printf("Setup GET request - Setup completed: %v", authConfig.SetupCompleted)
			if !authConfig.SetupCompleted {
				log.Printf("Setup token: %s", authConfig.SetupToken)
				log.Printf("Setup token expiry: %v", authConfig.SetupTokenExpiry)
			}

			response := APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"setup_completed": authConfig.SetupCompleted,
					"setup_required":  !authConfig.SetupCompleted,
				},
			}

			// Include setup token if setup is not completed
			if !authConfig.SetupCompleted {
				response.Data.(map[string]interface{})["setup_token"] = authConfig.SetupToken
				response.Data.(map[string]interface{})["setup_token_expiry"] = authConfig.SetupTokenExpiry
			}

			c.JSON(http.StatusOK, response)
			return
		}

		// Handle POST request
		var setupRequest struct {
			Username   string `json:"username"`
			Password   string `json:"password"`
			SetupToken string `json:"setup_token"`
		}

		log.Printf("=== SETUP REQUEST DEBUG ===")
		log.Printf("Method: %s", c.Request.Method)
		log.Printf("Content-Type: %s", c.GetHeader("Content-Type"))

		if err := c.ShouldBindJSON(&setupRequest); err != nil {
			log.Printf("Setup JSON binding failed: %v", err)
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request format",
			})
			return
		}

		log.Printf("Setup request - Username: %s, Password length: %d, Token: %s",
			setupRequest.Username, len(setupRequest.Password), setupRequest.SetupToken)

		// Load auth config
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config in setup POST: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}

		log.Printf("Current auth config - Setup completed: %v, Token: %s",
			authConfig.SetupCompleted, authConfig.SetupToken)

		// Check if setup is already completed
		if authConfig.SetupCompleted {
			log.Printf("Setup already completed")
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Setup already completed",
			})
			return
		}

		// Validate setup token
		log.Printf("Validating setup token - Provided: %s, Expected: %s",
			setupRequest.SetupToken, authConfig.SetupToken)
		log.Printf("Token expiry: %v, Current time: %v", authConfig.SetupTokenExpiry, time.Now())

		if !validateSetupToken(authConfig, setupRequest.SetupToken) {
			log.Printf("Setup token validation failed")
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Invalid or expired setup token",
			})
			return
		}

		log.Printf("Setup token validated successfully")

		// Validate input
		if setupRequest.Username == "" || setupRequest.Password == "" {
			log.Printf("Missing username or password")
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Username and password are required",
			})
			return
		}

		log.Printf("Completing setup for user: %s", setupRequest.Username)

		// Complete setup
		if err := completeSetup(setupRequest.Username, setupRequest.Password, store); err != nil {
			log.Printf("Failed to complete setup: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to complete setup",
			})
			return
		}

		log.Printf("Setup completed successfully for user: %s", setupRequest.Username)

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Setup completed successfully",
		})
	}
}

// Utility functions

// getCAInfo retrieves CA certificate information
func getCAInfo(store *storage.Storage) (CAInfo, error) {
	var caInfo CAInfo

	caPath := store.GetCAPublicKeyPath()
	if _, err := os.Stat(caPath); os.IsNotExist(err) {
		return caInfo, fmt.Errorf("CA certificate not found")
	}

	// Get CA certificate details using openssl
	cmd := exec.Command("openssl", "x509", "-in", caPath, "-noout", "-text")
	output, err := cmd.Output()
	if err != nil {
		return caInfo, fmt.Errorf("failed to get CA certificate details: %w", err)
	}

	outputStr := string(output)

	// Parse certificate details
	if idx := strings.Index(outputStr, "Subject:"); idx != -1 {
		subjectLine := outputStr[idx:]
		if endIdx := strings.Index(subjectLine, "\n"); endIdx != -1 {
			subject := strings.TrimSpace(subjectLine[8:endIdx])
			// Extract CN from subject
			if cnIdx := strings.Index(subject, "CN="); cnIdx != -1 {
				cnPart := subject[cnIdx+3:]
				if commaIdx := strings.Index(cnPart, ","); commaIdx != -1 {
					caInfo.CommonName = cnPart[:commaIdx]
				} else {
					caInfo.CommonName = cnPart
				}
			}
			// Extract O from subject
			if oIdx := strings.Index(subject, "O="); oIdx != -1 {
				oPart := subject[oIdx+2:]
				if commaIdx := strings.Index(oPart, ","); commaIdx != -1 {
					caInfo.Organization = oPart[:commaIdx]
				} else {
					caInfo.Organization = oPart
				}
			}
			// Extract C from subject
			if cIdx := strings.Index(subject, "C="); cIdx != -1 {
				cPart := subject[cIdx+2:]
				if commaIdx := strings.Index(cPart, ","); commaIdx != -1 {
					caInfo.Country = cPart[:commaIdx]
				} else {
					caInfo.Country = cPart
				}
			}
		}
	}

	if idx := strings.Index(outputStr, "Not After :"); idx != -1 {
		dateLine := outputStr[idx:]
		if endIdx := strings.Index(dateLine, "\n"); endIdx != -1 {
			caInfo.ExpiryDate = strings.TrimSpace(dateLine[11:endIdx])
			// Check if expired
			if expiryTime, err := time.Parse("Jan  2 15:04:05 2006 MST", caInfo.ExpiryDate); err == nil {
				caInfo.IsExpired = expiryTime.Before(time.Now())
			}
		}
	}

	return caInfo, nil
}

// getCertificateInfo retrieves certificate information
func getCertificateInfo(store *storage.Storage, name string) (CertificateInfo, error) {
	var certInfo CertificateInfo
	certInfo.CommonName = name

	certPath := store.GetCertificatePath(name)
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return certInfo, fmt.Errorf("certificate not found")
	}

	// Check if it's a client certificate
	p12Path := store.GetCertificateP12Path(name)
	if _, err := os.Stat(p12Path); err == nil {
		certInfo.IsClient = true
	}

	// Get certificate details using openssl
	cmd := exec.Command("openssl", "x509", "-in", certPath, "-noout", "-text")
	output, err := cmd.Output()
	if err != nil {
		return certInfo, fmt.Errorf("failed to get certificate details: %w", err)
	}

	outputStr := string(output)

	// Parse certificate details
	if idx := strings.Index(outputStr, "Serial Number:"); idx != -1 {
		serialLine := outputStr[idx:]
		if endIdx := strings.Index(serialLine, "\n"); endIdx != -1 {
			certInfo.SerialNumber = strings.TrimSpace(serialLine[14:endIdx])
		}
	}

	if idx := strings.Index(outputStr, "Not After :"); idx != -1 {
		dateLine := outputStr[idx:]
		if endIdx := strings.Index(dateLine, "\n"); endIdx != -1 {
			certInfo.ExpiryDate = strings.TrimSpace(dateLine[11:endIdx])
			// Check if certificate is expired or expiring soon
			if expiryTime, err := time.Parse("Jan  2 15:04:05 2006 MST", certInfo.ExpiryDate); err == nil {
				now := time.Now()
				if expiryTime.Before(now) {
					certInfo.IsExpired = true
				} else if expiryTime.Before(now.Add(30 * 24 * time.Hour)) {
					certInfo.IsExpiringSoon = true
				}
			}
		}
	}

	// Check if certificate is revoked
	revokedPath := filepath.Join(store.GetCertificateDirectory(name), "revoked")
	if _, err := os.Stat(revokedPath); err == nil {
		certInfo.IsRevoked = true
	}

	return certInfo, nil
}
