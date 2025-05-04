package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/localca-go/pkg/certificates"
	"github.com/yourusername/localca-go/pkg/config"
	"github.com/yourusername/localca-go/pkg/storage"
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
	router.GET("/download/:name/:type", downloadCertificateHandler(certSvc, store))
}

// csrfMiddleware adds CSRF protection
func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for CSRF protection
		// Typically involves generating tokens, validating them, etc.
		c.Next()
	}
}

// sessionMiddleware adds session management
func sessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for session management
		// Could involve cookies, storage, etc.
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
	CommonName    string `json:"common_name"`
	ExpiryDate    string `json:"expiry_date"`
	IsClient      bool   `json:"is_client"`
	SerialNumber  string `json:"serial_number"`
	IsExpired     bool   `json:"is_expired"`
	IsExpiringSoon bool  `json:"is_expiring_soon"`
}

// CAInfo represents CA information for display
type CAInfo struct {
	CommonName   string `json:"common_name"`
	Organization string `json:"organization"`
	Country      string `json:"country"`
	ExpiryDate   string `json:"expiry_date"`
	IsExpired    bool   `json:"is_expired"`
}