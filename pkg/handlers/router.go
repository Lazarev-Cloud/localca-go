package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the routes for the application
func SetupRoutes(router *gin.Engine, certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) {
	// Add middleware
	router.Use(gin.Recovery())

	// Configure CSRF protection
	router.Use(csrfMiddleware())

	// Configure session
	router.Use(sessionMiddleware())

	// Home page
	router.GET("/", indexHandler(certSvc, store, cfg))
	router.POST("/", createCertificateHandler(certSvc, store))

	// Certificate file view
	router.GET("/files", filesHandler(certSvc, store))

	// Operations
	router.POST("/renew", renewCertificateHandler(certSvc, store))
	router.POST("/delete", deleteCertificateHandler(certSvc, store))
	router.POST("/renew-ca", renewCAHandler(certSvc, store))
	router.POST("/revoke", revokeCertificateHandler(certSvc, store))

	// Settings
	router.GET("/settings", settingsHandler(certSvc, store, cfg))
	router.POST("/settings", saveSettingsHandler(certSvc, store, cfg))
	router.POST("/test-email", testEmailHandler(certSvc, store, cfg))

	// Certificate download
	router.GET("/download/ca", downloadCAHandler(certSvc, store))
	router.GET("/download/crl", downloadCRLHandler(certSvc, store))
	router.GET("/download/:name/:type", downloadCertificateHandler(certSvc, store))

	// Setup API routes for the Next.js frontend
	SetupAPIRoutes(router, certSvc, store)
}

// csrfMiddleware adds CSRF protection
func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF check for GET requests
		if c.Request.Method == "GET" {
			// Generate a new token for the response
			token := generateCSRFToken()
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// For POST, PUT, DELETE requests, validate the token
		token := c.PostForm("csrf_token")
		if token == "" {
			token = c.GetHeader("X-CSRF-Token")
		}

		// In a real implementation, we would validate the token against a session
		// For now, just ensure a token was provided
		if token == "" {
			c.JSON(http.StatusForbidden, APIResponse{
				Success: false,
				Message: "CSRF token missing",
			})
			c.Abort()
			return
		}

		// Generate a new token for the response
		newToken := generateCSRFToken()
		c.Set("csrf_token", newToken)
		c.Next()
	}
}

// generateCSRFToken generates a random token for CSRF protection
func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// sessionMiddleware adds session management
func sessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, this would initialize and manage user sessions
		// For now, just pass through
		c.Next()
	}
}

// APIResponse is the standard response format for API calls
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// CertificateInfo represents certificate information for display
type CertificateInfo struct {
	CommonName     string `json:"common_name"`
	ExpiryDate     string `json:"expiry_date"`
	IsClient       bool   `json:"is_client"`
	SerialNumber   string `json:"serial_number"`
	IsExpired      bool   `json:"is_expired"`
	IsExpiringSoon bool   `json:"is_expiring_soon"`
	IsRevoked      bool   `json:"is_revoked"`
}

// CAInfo represents CA information for display
type CAInfo struct {
	CommonName   string `json:"common_name"`
	Organization string `json:"organization"`
	Country      string `json:"country"`
	ExpiryDate   string `json:"expiry_date"`
	IsExpired    bool   `json:"is_expired"`
}
