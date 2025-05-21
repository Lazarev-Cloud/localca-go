package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// CSRFTokenStore stores valid CSRF tokens with expiration
type CSRFTokenStore struct {
	tokens map[string]time.Time
	mutex  sync.RWMutex
}

// Global CSRF token store
var csrfStore = &CSRFTokenStore{
	tokens: make(map[string]time.Time),
}

// Initialize CSRF token cleaner
func init() {
	go cleanupCSRFTokens()
}

// cleanupCSRFTokens periodically removes expired CSRF tokens
func cleanupCSRFTokens() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		csrfStore.mutex.Lock()
		now := time.Now()
		for token, expiry := range csrfStore.tokens {
			if now.After(expiry) {
				delete(csrfStore.tokens, token)
			}
		}
		csrfStore.mutex.Unlock()
	}
}

// SetupRoutes configures all routes for the application
func SetupRoutes(router *gin.Engine, certSvc certificates.CertificateServiceInterface, store *storage.Storage, cfg *config.Config) {
	// Add middleware
	router.Use(gin.Recovery())

	// Add security headers
	router.Use(securityHeadersMiddleware())

	// Configure CSRF protection
	router.Use(csrfMiddleware())

	// Configure session
	router.Use(sessionMiddleware())

	// Add authentication middleware
	router.Use(authMiddleware(store))

	// Authentication routes
	router.GET("/login", loginHandler(certSvc, store))
	router.POST("/login", loginPostHandler(certSvc, store))
	router.GET("/logout", logoutHandler())
	router.GET("/setup", setupHandler(certSvc, store, cfg))
	router.POST("/setup", setupPostHandler(certSvc, store))

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

// securityHeadersMiddleware adds security headers to prevent XSS and other attacks
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Add strict transport security header if using HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// csrfMiddleware adds CSRF protection
func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF check for GET requests
		if c.Request.Method == "GET" {
			// Generate a new token for the response
			token := generateCSRFToken()

			// Store token with expiration (24 hours)
			csrfStore.mutex.Lock()
			csrfStore.tokens[token] = time.Now().Add(24 * time.Hour)
			csrfStore.mutex.Unlock()

			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// For POST, PUT, DELETE requests, validate the token
		token := c.PostForm("csrf_token")
		if token == "" {
			token = c.GetHeader("X-CSRF-Token")
		}

		// Check if token exists and is valid
		if token == "" || !validateCSRFToken(token) {
			c.JSON(http.StatusForbidden, APIResponse{
				Success: false,
				Message: "Invalid or missing CSRF token",
			})
			c.Abort()
			return
		}

		// Generate a new token for the response
		newToken := generateCSRFToken()

		// Store new token with expiration
		csrfStore.mutex.Lock()
		csrfStore.tokens[newToken] = time.Now().Add(24 * time.Hour)
		csrfStore.mutex.Unlock()

		c.Set("csrf_token", newToken)
		c.Next()
	}
}

// generateCSRFToken generates a random token for CSRF protection
func generateCSRFToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// If we can't generate random bytes, use timestamp as fallback
		// This is not ideal but better than nothing
		return base64.StdEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// validateCSRFToken validates a CSRF token
func validateCSRFToken(token string) bool {
	csrfStore.mutex.RLock()
	defer csrfStore.mutex.RUnlock()

	expiry, exists := csrfStore.tokens[token]
	if !exists {
		return false
	}

	// Check if token has expired
	if time.Now().After(expiry) {
		return false
	}

	return true
}

// sessionMiddleware adds session management
func sessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, this would initialize and manage user sessions
		// For now, just pass through
		c.Next()
	}
}
