package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lazarev-Cloud/localca-go/pkg/acme"
	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// ACMEStats represents ACME usage statistics
type ACMEStats struct {
	Accounts      int `json:"accounts"`
	Orders        int `json:"orders"`
	PendingOrders int `json:"pending_orders"`
	IssuedCerts   int `json:"issued_certs"`
}

// setupACMEUIRoutes sets up the UI routes for ACME management
func setupACMEUIRoutes(router *gin.Engine, certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) {
	router.GET("/acme", handleACMESettings(certSvc, store, cfg))
	router.POST("/acme/enable", handleACMEEnable(certSvc, store, cfg))
	router.POST("/acme/disable", handleACMEDisable(certSvc, store, cfg))
	router.GET("/acme/logs", handleACMELogs(certSvc, store, cfg))
}

// handleACMESettings handles the ACME settings page
func handleACMESettings(certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine base URL
		baseURL := cfg.ACMEBaseURL
		if baseURL == "" {
			// Default to the hostname
			if cfg.TLSEnabled {
				baseURL = "https://" + cfg.Hostname
				if cfg.HTTPSPort != 443 {
					baseURL = fmt.Sprintf("%s:%d", baseURL, cfg.HTTPSPort)
				}
			} else {
				baseURL = "http://" + cfg.Hostname
				if cfg.HTTPPort != 80 {
					baseURL = fmt.Sprintf("%s:%d", baseURL, cfg.HTTPPort)
				}
			}
		}

		// Get ACME stats
		stats := getACMEStats(store)

		c.HTML(http.StatusOK, "acme.html", gin.H{
			"ACMEEnabled": cfg.ACMEEnabled,
			"ACMEBaseURL": baseURL + "/acme/directory",
			"BaseURL":     baseURL,
			"Stats":       stats,
			"CSRFToken":   c.GetString("csrf_token"),
		})
	}
}

// handleACMEEnable handles enabling ACME
func handleACMEEnable(certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get form data
		baseURL := c.PostForm("base_url")

		// Validate base URL
		if baseURL == "" {
			c.HTML(http.StatusBadRequest, "acme.html", gin.H{
				"Error":       "Base URL is required",
				"ACMEEnabled": cfg.ACMEEnabled,
				"BaseURL":     baseURL,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Save ACME settings
		settingsDir := filepath.Join(store.GetBasePath(), "settings")
		if err := os.MkdirAll(settingsDir, 0755); err != nil {
			log.Printf("Failed to create settings directory: %v", err)
			c.HTML(http.StatusInternalServerError, "acme.html", gin.H{
				"Error":       fmt.Sprintf("Failed to create settings directory: %v", err),
				"ACMEEnabled": cfg.ACMEEnabled,
				"BaseURL":     baseURL,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Save ACME enabled flag
		if err := os.WriteFile(filepath.Join(settingsDir, "acme_enabled.txt"), []byte("true"), 0644); err != nil {
			log.Printf("Failed to save ACME enabled flag: %v", err)
			c.HTML(http.StatusInternalServerError, "acme.html", gin.H{
				"Error":       fmt.Sprintf("Failed to save ACME settings: %v", err),
				"ACMEEnabled": cfg.ACMEEnabled,
				"BaseURL":     baseURL,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Save ACME base URL
		if err := os.WriteFile(filepath.Join(settingsDir, "acme_base_url.txt"), []byte(baseURL), 0644); err != nil {
			log.Printf("Failed to save ACME base URL: %v", err)
			c.HTML(http.StatusInternalServerError, "acme.html", gin.H{
				"Error":       fmt.Sprintf("Failed to save ACME settings: %v", err),
				"ACMEEnabled": cfg.ACMEEnabled,
				"BaseURL":     baseURL,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Update config
		cfg.ACMEEnabled = true
		cfg.ACMEBaseURL = baseURL

		// Redirect to ACME settings page
		c.Redirect(http.StatusSeeOther, "/acme")
	}
}

// handleACMEDisable handles disabling ACME
func handleACMEDisable(certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Save ACME settings
		settingsDir := filepath.Join(store.GetBasePath(), "settings")
		if err := os.MkdirAll(settingsDir, 0755); err != nil {
			log.Printf("Failed to create settings directory: %v", err)
			c.HTML(http.StatusInternalServerError, "acme.html", gin.H{
				"Error":       fmt.Sprintf("Failed to create settings directory: %v", err),
				"ACMEEnabled": cfg.ACMEEnabled,
				"BaseURL":     cfg.ACMEBaseURL,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Save ACME enabled flag
		if err := os.WriteFile(filepath.Join(settingsDir, "acme_enabled.txt"), []byte("false"), 0644); err != nil {
			log.Printf("Failed to save ACME enabled flag: %v", err)
			c.HTML(http.StatusInternalServerError, "acme.html", gin.H{
				"Error":       fmt.Sprintf("Failed to save ACME settings: %v", err),
				"ACMEEnabled": cfg.ACMEEnabled,
				"BaseURL":     cfg.ACMEBaseURL,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Update config
		cfg.ACMEEnabled = false

		// Redirect to ACME settings page
		c.Redirect(http.StatusSeeOther, "/acme")
	}
}

// handleACMELogs handles viewing ACME logs
func handleACMELogs(certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would be a placeholder for a real log viewer
		// In a production system, you'd want to implement proper logging
		c.String(http.StatusOK, "ACME logs are not implemented yet.")
	}
}

// getACMEStats retrieves ACME usage statistics
func getACMEStats(store *storage.Storage) ACMEStats {
	// Initialize the ACME store
	acmeStore, err := acme.NewACMEStore(store)
	if err != nil {
		return ACMEStats{}
	}

	stats := ACMEStats{}

	// Count accounts
	accountsDir := filepath.Join(store.GetBasePath(), "acme", "accounts")
	if entries, err := os.ReadDir(accountsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				stats.Accounts++
			}
		}
	}

	// Count orders
	ordersDir := filepath.Join(store.GetBasePath(), "acme", "orders")
	if entries, err := os.ReadDir(ordersDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				stats.Orders++

				// Read the order to check its status
				orderPath := filepath.Join(ordersDir, entry.Name())
				orderData, err := os.ReadFile(orderPath)
				if err == nil {
					// Check if the order contains "status":"pending" or "status":"processing"
					orderStr := string(orderData)
					if strings.Contains(orderStr, `"status":"pending"`) || 
					   strings.Contains(orderStr, `"status":"processing"`) {
						stats.PendingOrders++
					}

					// Check if the order contains "status":"valid"
					if strings.Contains(orderStr, `"status":"valid"`) {
						stats.IssuedCerts++
					}
				}
			}
		}
	}

	return stats
}