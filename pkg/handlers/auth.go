package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Default values
const (
	DefaultAdminUsername = "admin"
)

// AuthConfig stores authentication configuration
type AuthConfig struct {
	AdminUsername     string    `json:"admin_username"`
	AdminPasswordHash string    `json:"admin_password_hash"`
	SetupCompleted    bool      `json:"setup_completed"`
	SetupToken        string    `json:"setup_token,omitempty"`
	SetupTokenExpiry  time.Time `json:"setup_token_expiry,omitempty"`
}

// authMiddleware handles authentication
func authMiddleware(store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path
		path := c.Request.URL.Path

		// Skip authentication for public paths
		if isPublicPath(path) {
			c.Next()
			return
		}

		// Check if setup is completed
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// If setup is not completed, redirect to setup page
		if !authConfig.SetupCompleted {
			// Redirect to setup page
			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusUnauthorized, APIResponse{
					Success: false,
					Message: "Setup required",
					Data: map[string]interface{}{
						"setup_required": true,
					},
				})
			} else {
				c.Redirect(http.StatusSeeOther, "/setup")
			}
			c.Abort()
			return
		}

		// Check if user is authenticated
		session, err := c.Cookie("session")
		if err != nil || !validateSession(session, store) {
			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusUnauthorized, APIResponse{
					Success: false,
					Message: "Authentication required",
				})
			} else {
				c.Redirect(http.StatusSeeOther, "/login")
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPublicPath checks if the path is publicly accessible
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/static/",
		"/login",
		"/api/login",
		"/.well-known/acme-challenge/",
		"/download/ca",
		"/download/crl",
		"/acme/",
	}

	for _, prefix := range publicPaths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// LoadAuthConfig loads authentication configuration
func LoadAuthConfig(store *storage.Storage) (*AuthConfig, error) {
	configPath := filepath.Join(store.GetBasePath(), "auth.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := &AuthConfig{
			AdminUsername:    DefaultAdminUsername,
			SetupCompleted:   false,
			SetupToken:       generateSetupToken(),
			SetupTokenExpiry: time.Now().Add(24 * time.Hour),
		}

		// Save config
		if err := saveAuthConfig(config, store); err != nil {
			return nil, err
		}

		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse config
	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveAuthConfig saves authentication configuration
func saveAuthConfig(config *AuthConfig, store *storage.Storage) error {
	configPath := filepath.Join(store.GetBasePath(), "auth.json")

	// Marshal config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return err
	}

	return nil
}

// generateSetupToken generates a random token for initial setup
func generateSetupToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp if random generation fails
		return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// validateSetupToken validates the setup token
func validateSetupToken(config *AuthConfig, token string) bool {
	// Check if token is valid and not expired
	return config.SetupToken != "" &&
		subtle.ConstantTimeCompare([]byte(config.SetupToken), []byte(token)) == 1 &&
		time.Now().Before(config.SetupTokenExpiry)
}

// validateSession validates a session token
func validateSession(sessionToken string, store *storage.Storage) bool {
	// TODO: Implement proper session validation
	// For now, just check if the token exists
	return sessionToken != ""
}

// generateSessionToken generates a new session token
func generateSessionToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp if random generation fails
		return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// checkPasswordHash checks if a password matches a hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// completeSetup completes the initial setup
func completeSetup(username, password string, store *storage.Storage) error {
	// Hash password
	passwordHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	// Load config
	config, err := LoadAuthConfig(store)
	if err != nil {
		return err
	}

	// Update config
	config.AdminUsername = username
	config.AdminPasswordHash = passwordHash
	config.SetupCompleted = true
	config.SetupToken = "" // Clear setup token

	// Save config
	if err := saveAuthConfig(config, store); err != nil {
		return err
	}

	return nil
} 