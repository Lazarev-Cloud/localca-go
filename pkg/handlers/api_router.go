package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

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

	// Add rate limiting before auth
	router.Use(globalRateLimitMiddleware())
	// Add authentication middleware for API
	router.Use(apiAuthMiddleware(store))

	// Setup API routes
	SetupAPIRoutes(router, certSvc, store)

	// Metrics endpoint (basic placeholder)
	router.GET("/api/metrics", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; version=0.0.4")
		c.String(http.StatusOK, "localca_up 1\n")
	})

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
		// Remove server banner
		c.Writer.Header().Del("Server")
		c.Header("X-Powered-By", "")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

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

			if allowedOrigins == "*" && gin.Mode() != gin.ReleaseMode {
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

// Simple IP-based global rate limiter with X-RateLimit-* headers
func globalRateLimitMiddleware() gin.HandlerFunc {
	type clientState struct {
		remaining int
		reset     time.Time
	}
	var (
		mu      sync.Mutex
		buckets = make(map[string]*clientState)
		window  = time.Minute
		// Defaults; can be tuned by env later
		limitPerMin = 600
	)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		mu.Lock()
		st, ok := buckets[ip]
		if !ok || now.After(st.reset) {
			st = &clientState{remaining: limitPerMin, reset: now.Add(window)}
			buckets[ip] = st
		}
		if st.remaining <= 0 {
			resetSec := int(time.Until(st.reset).Seconds())
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limitPerMin))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetSec))
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Too many requests"})
			return
		}
		st.remaining--
		rem := st.remaining
		resetSec := int(time.Until(st.reset).Seconds())
		mu.Unlock()

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limitPerMin))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rem))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetSec))
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

		// Check if user is authenticated via JWT (Authorization: Bearer)
		if !validateJWTFromRequest(c) {
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
		"/api/health",
		"/version",
		"/api/login",
		"/api/auth/refresh",
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
