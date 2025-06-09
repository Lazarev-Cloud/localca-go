package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// SetupAPIOnlyRoutes configures API-only routes (no web UI)
func SetupAPIOnlyRoutes(router *gin.Engine, certSvc certificates.CertificateServiceInterface, store *storage.Storage) {
	// Add middleware
	router.Use(gin.Recovery())

	// Add security headers for API
	router.Use(apiSecurityHeadersMiddleware())

	// Add authentication middleware for API
	router.Use(apiAuthMiddleware(store))

	// Setup API routes
	SetupAPIRoutes(router, certSvc, store)

	// Health check endpoint (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "localca-api",
		})
	})

	// Version endpoint (public)
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": "1.0.0",
			"service": "localca-api",
		})
	})
}

// apiSecurityHeadersMiddleware adds security headers for API-only server
func apiSecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Essential security headers for API
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Server", "LocalCA-API")
		c.Header("X-Powered-By", "")

		// API-specific headers
		c.Header("Content-Type", "application/json")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// CORS headers
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
			if allowedOrigins == "" {
				allowedOrigins = "http://localhost:3000,https://localhost:3000"
			}

			if allowedOrigins == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				origins := strings.Split(allowedOrigins, ",")
				for _, allowedOrigin := range origins {
					if strings.TrimSpace(allowedOrigin) == origin {
						c.Header("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// apiAuthMiddleware handles authentication for API endpoints only
func apiAuthMiddleware(store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path
		path := c.Request.URL.Path

		// Skip authentication for public API paths
		if isPublicAPIPath(path) {
			c.Next()
			return
		}

		// Check if setup is completed
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			c.Abort()
			return
		}

		// If setup is not completed, only allow setup endpoints
		if !authConfig.SetupCompleted {
			if strings.HasPrefix(path, "/api/setup") {
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Setup required",
				Data: map[string]interface{}{
					"setup_required": true,
				},
			})
			c.Abort()
			return
		}

		// Check if user is authenticated
		session, err := c.Cookie("session")
		if err != nil || !validateSession(session, store) {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPublicAPIPath checks if the API path is publicly accessible
func isPublicAPIPath(path string) bool {
	publicPaths := []string{
		"/health",
		"/version",
		"/api/login",
		"/api/setup",
		"/api/auth/status",
		"/.well-known/acme-challenge/",
		"/api/download/ca",
		"/api/download/crl",
		"/acme/",
	}

	for _, prefix := range publicPaths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
