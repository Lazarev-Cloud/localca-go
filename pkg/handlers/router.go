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
	"golang.org/x/time/rate"
)

// ClientIP represents IP-based rate limiter
type ClientIP struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter for controlling request rates
type RateLimiter struct {
	ips    map[string]*ClientIP
	mu     sync.Mutex
	rate   rate.Limit
	bucket int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		ips:    make(map[string]*ClientIP),
		rate:   r,
		bucket: b,
	}
}

// GetLimiter returns rate limiter for a client IP
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.ips[ip]
	if !exists {
		client = &ClientIP{
			limiter:  rate.NewLimiter(rl.rate, rl.bucket),
			lastSeen: time.Now(),
		}
		rl.ips[ip] = client
	} else {
		client.lastSeen = time.Now()
	}

	return client.limiter
}

// CleanupStaleClients removes inactive client IPs
func (rl *RateLimiter) CleanupStaleClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-time.Hour)
	for ip, client := range rl.ips {
		if client.lastSeen.Before(threshold) {
			delete(rl.ips, ip)
		}
	}
}

// RateLimiterMiddleware controls request rate
func RateLimiterMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.GetLimiter(ip)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, APIResponse{
				Success: false,
				Message: "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SetupRoutes configures the routes for the application
func SetupRoutes(router *gin.Engine, certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) {
	// Add middleware
	router.Use(gin.Recovery())

	// Setup rate limiting (10 requests per second, burst of 20)
	limiter := NewRateLimiter(10, 20)
	router.Use(RateLimiterMiddleware(limiter))

	// Start cleanup goroutine
	go func() {
		for {
			time.Sleep(time.Hour)
			limiter.CleanupStaleClients()
		}
	}()

	// Configure secure session with HTTP-only cookies
	router.Use(secureSessionMiddleware())

	// Configure CSRF protection
	router.Use(csrfMiddleware())

	// Configure security headers
	router.Use(securityHeadersMiddleware())

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
}

// securityHeadersMiddleware adds security headers to responses
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' https://code.jquery.com https://cdn.jsdelivr.net; style-src 'self' https://cdn.jsdelivr.net https://maxcdn.bootstrapcdn.com; img-src 'self' data:")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
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
			c.Set("csrf_token", token)

			// Set the token in a cookie as well
			c.SetCookie("csrf_token", token, 3600, "/", "", false, false)
			c.Next()
			return
		}

		// For POST, PUT, DELETE requests, validate the token
		formToken := c.PostForm("csrf_token")
		headerToken := c.GetHeader("X-CSRF-Token")
		cookieToken, _ := c.Cookie("csrf_token")

		// Check if any token matches
		validToken := false
		if formToken != "" && formToken == cookieToken {
			validToken = true
		}
		if headerToken != "" && headerToken == cookieToken {
			validToken = true
		}

		if !validToken {
			c.JSON(http.StatusForbidden, APIResponse{
				Success: false,
				Message: "CSRF token invalid or missing",
			})
			c.Abort()
			return
		}

		// Generate a new token for the response
		newToken := generateCSRFToken()
		c.Set("csrf_token", newToken)
		c.SetCookie("csrf_token", newToken, 3600, "/", "", false, false)
		c.Next()
	}
}

// generateCSRFToken generates a random token for CSRF protection
func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// If we can't get random data, use timestamp as fallback
		b = []byte(time.Now().String())
	}
	return base64.StdEncoding.EncodeToString(b)
}

// secureSessionMiddleware adds secure session management
func secureSessionMiddleware() gin.HandlerFunc {
	// In a production app, you'd use a proper session library like gorilla/sessions
	return func(c *gin.Context) {
		// Generate session ID if not exist
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			sessionID = generateSessionID()
			c.SetCookie("session_id", sessionID, 3600, "/", "", false, true)
		}

		c.Set("session_id", sessionID)
		c.Next()
	}
}

// generateSessionID creates a secure session ID
func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// If we can't get random data, use timestamp as fallback
		b = []byte(time.Now().String())
	}
	return base64.StdEncoding.EncodeToString(b)
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
