package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"encoding/base64"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// loginHandler handles the login page
func loginHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if setup is completed
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": "Internal server error",
			})
			return
		}

		// If setup is not completed, redirect to setup page
		if !authConfig.SetupCompleted {
			c.Redirect(http.StatusSeeOther, "/setup")
			return
		}

		// Render login page
		c.HTML(http.StatusOK, "login.html", gin.H{
			"CSRFToken": c.GetString("csrf_token"),
		})
	}
}

// loginPostHandler handles login form submission
func loginPostHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get form data
		username := c.PostForm("username")
		password := c.PostForm("password")

		// Validate input
		if username == "" || password == "" {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"Error":     "Username and password are required",
				"CSRFToken": c.GetString("csrf_token"),
			})
			return
		}

		// Load auth config
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": "Internal server error",
			})
			return
		}

		// Check credentials
		if username != authConfig.AdminUsername || !checkPasswordHash(password, authConfig.AdminPasswordHash) {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"Error":     "Invalid username or password",
				"CSRFToken": c.GetString("csrf_token"),
			})
			return
		}

		// Generate session token
		sessionToken := generateSessionToken()

		// Set session cookie
		c.SetCookie(
			"session",
			sessionToken,
			int(time.Hour.Seconds()*24), // 24 hours
			"/",
			"",
			c.Request.TLS != nil, // secure if using HTTPS
			true,                 // httpOnly
		)

		// Create session file
		sessionsDir := filepath.Join(store.GetBasePath(), "sessions")
		if err := os.MkdirAll(sessionsDir, 0700); err != nil {
			log.Printf("Failed to create sessions directory: %v", err)
		} else {
			sessionFile := filepath.Join(sessionsDir, base64.URLEncoding.EncodeToString([]byte(sessionToken)))
			if err := os.WriteFile(sessionFile, []byte(authConfig.AdminUsername), 0600); err != nil {
				log.Printf("Failed to create session file: %v", err)
			}
		}

		// Redirect to home page
		c.Redirect(http.StatusSeeOther, "/")
	}
}

// apiLoginHandler handles API login
func apiLoginHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get JSON data
		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request format",
			})
			return
		}

		// Validate input
		if loginData.Username == "" || loginData.Password == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Username and password are required",
			})
			return
		}

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

		// Check credentials
		if loginData.Username != authConfig.AdminUsername || !checkPasswordHash(loginData.Password, authConfig.AdminPasswordHash) {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Invalid username or password",
			})
			return
		}

		// Generate session token
		sessionToken := generateSessionToken()

		// Set session cookie
		c.SetCookie(
			"session",
			sessionToken,
			int(time.Hour.Seconds()*24), // 24 hours
			"/",
			"",
			c.Request.TLS != nil, // secure if using HTTPS
			true,                 // httpOnly
		)

		// Create session file
		sessionsDir := filepath.Join(store.GetBasePath(), "sessions")
		if err := os.MkdirAll(sessionsDir, 0700); err != nil {
			log.Printf("Failed to create sessions directory: %v", err)
		} else {
			sessionFile := filepath.Join(sessionsDir, base64.URLEncoding.EncodeToString([]byte(sessionToken)))
			if err := os.WriteFile(sessionFile, []byte(authConfig.AdminUsername), 0600); err != nil {
				log.Printf("Failed to create session file: %v", err)
			}
		}

		// Return success with token
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Login successful",
			Data: map[string]interface{}{
				"token": sessionToken,
			},
		})
	}
}

// logoutHandler handles logout
func logoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear session cookie
		c.SetCookie(
			"session",
			"",
			-1, // expire immediately
			"/",
			"",
			false, // secure
			true,  // httpOnly
		)

		// Redirect to login page
		c.Redirect(http.StatusSeeOther, "/login")
	}
}

// setupHandler handles the initial setup page
func setupHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if setup is already completed
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": "Internal server error",
			})
			return
		}

		// If setup is already completed, redirect to login page
		if authConfig.SetupCompleted {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		// Render setup page
		c.HTML(http.StatusOK, "setup.html", gin.H{
			"CSRFToken":  c.GetString("csrf_token"),
			"SetupToken": authConfig.SetupToken,
		})
	}
}

// setupPostHandler handles setup form submission
func setupPostHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get form data
		username := c.PostForm("username")
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")
		setupToken := c.PostForm("setup_token")

		// Validate input
		if username == "" || password == "" || confirmPassword == "" || setupToken == "" {
			c.HTML(http.StatusBadRequest, "setup.html", gin.H{
				"Error":     "All fields are required",
				"CSRFToken": c.GetString("csrf_token"),
			})
			return
		}

		// Check if passwords match
		if password != confirmPassword {
			c.HTML(http.StatusBadRequest, "setup.html", gin.H{
				"Error":     "Passwords do not match",
				"CSRFToken": c.GetString("csrf_token"),
			})
			return
		}

		// Load auth config
		authConfig, err := LoadAuthConfig(store)
		if err != nil {
			log.Printf("Failed to load auth config: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": "Internal server error",
			})
			return
		}

		// Check if setup is already completed
		if authConfig.SetupCompleted {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		// Validate setup token
		if !validateSetupToken(authConfig, setupToken) {
			c.HTML(http.StatusBadRequest, "setup.html", gin.H{
				"Error":     "Invalid or expired setup token",
				"CSRFToken": c.GetString("csrf_token"),
			})
			return
		}

		// Complete setup
		if err := completeSetup(username, password, store); err != nil {
			log.Printf("Failed to complete setup: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": "Failed to complete setup",
			})
			return
		}

		// Redirect to login page
		c.Redirect(http.StatusSeeOther, "/login")
	}
}

// apiSetupHandler handles API setup
func apiSetupHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET request returns setup info
		if c.Request.Method == "GET" {
			// Check if setup is already completed
			authConfig, err := LoadAuthConfig(store)
			if err != nil {
				log.Printf("Failed to load auth config: %v", err)
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Message: "Internal server error",
				})
				return
			}

			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"setup_completed": authConfig.SetupCompleted,
					"setup_token":     authConfig.SetupToken,
				},
			})
			return
		}

		// POST request completes setup
		var setupData struct {
			Username        string `json:"username"`
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirm_password"`
			SetupToken      string `json:"setup_token"`
		}
		if err := c.BindJSON(&setupData); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request format",
			})
			return
		}

		// Validate input
		if setupData.Username == "" || setupData.Password == "" || setupData.ConfirmPassword == "" || setupData.SetupToken == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "All fields are required",
			})
			return
		}

		// Check if passwords match
		if setupData.Password != setupData.ConfirmPassword {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Passwords do not match",
			})
			return
		}

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

		// Check if setup is already completed
		if authConfig.SetupCompleted {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Setup is already completed",
			})
			return
		}

		// Validate setup token
		if !validateSetupToken(authConfig, setupData.SetupToken) {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid or expired setup token",
			})
			return
		}

		// Complete setup
		if err := completeSetup(setupData.Username, setupData.Password, store); err != nil {
			log.Printf("Failed to complete setup: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to complete setup",
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Setup completed successfully",
		})
	}
}
