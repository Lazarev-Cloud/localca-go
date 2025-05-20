package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables
	origEnvVars := map[string]string{
		"LOCALCA_DATA_DIR":    os.Getenv("LOCALCA_DATA_DIR"),
		"LOCALCA_HOST":        os.Getenv("LOCALCA_HOST"),
		"LOCALCA_PORT":        os.Getenv("LOCALCA_PORT"),
		"LOCALCA_HTTPS_PORT":  os.Getenv("LOCALCA_HTTPS_PORT"),
		"LOCALCA_ACME_PORT":   os.Getenv("LOCALCA_ACME_PORT"),
		"LOCALCA_CA_NAME":     os.Getenv("LOCALCA_CA_NAME"),
		"LOCALCA_CA_EMAIL":    os.Getenv("LOCALCA_CA_EMAIL"),
		"LOCALCA_SMTP_HOST":   os.Getenv("LOCALCA_SMTP_HOST"),
		"LOCALCA_SMTP_PORT":   os.Getenv("LOCALCA_SMTP_PORT"),
		"LOCALCA_SMTP_USER":   os.Getenv("LOCALCA_SMTP_USER"),
		"LOCALCA_SMTP_PASS":   os.Getenv("LOCALCA_SMTP_PASS"),
		"LOCALCA_SMTP_FROM":   os.Getenv("LOCALCA_SMTP_FROM"),
		"LOCALCA_NOTIFY_DAYS": os.Getenv("LOCALCA_NOTIFY_DAYS"),
	}

	// Restore environment variables when test completes
	defer func() {
		for key, val := range origEnvVars {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Test default configuration
	os.Unsetenv("LOCALCA_DATA_DIR")
	os.Unsetenv("LOCALCA_HOST")
	os.Unsetenv("LOCALCA_PORT")
	os.Unsetenv("LOCALCA_HTTPS_PORT")
	os.Unsetenv("LOCALCA_ACME_PORT")
	os.Unsetenv("LOCALCA_CA_NAME")
	os.Unsetenv("LOCALCA_CA_EMAIL")
	os.Unsetenv("LOCALCA_SMTP_HOST")
	os.Unsetenv("LOCALCA_SMTP_PORT")
	os.Unsetenv("LOCALCA_SMTP_USER")
	os.Unsetenv("LOCALCA_SMTP_PASS")
	os.Unsetenv("LOCALCA_SMTP_FROM")
	os.Unsetenv("LOCALCA_NOTIFY_DAYS")

	cfg := LoadConfig()

	// Check default values
	if cfg.DataDir != "./data" {
		t.Errorf("Expected default DataDir './data', got '%s'", cfg.DataDir)
	}
	if cfg.Host != "localhost" {
		t.Errorf("Expected default Host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected default Port 8080, got %d", cfg.Port)
	}
	if cfg.HTTPSPort != 8443 {
		t.Errorf("Expected default HTTPSPort 8443, got %d", cfg.HTTPSPort)
	}
	if cfg.ACMEPort != 8555 {
		t.Errorf("Expected default ACMEPort 8555, got %d", cfg.ACMEPort)
	}
	if cfg.CAName != "Local CA" {
		t.Errorf("Expected default CAName 'Local CA', got '%s'", cfg.CAName)
	}
	if cfg.CAEmail != "admin@localhost" {
		t.Errorf("Expected default CAEmail 'admin@localhost', got '%s'", cfg.CAEmail)
	}
	if cfg.NotifyDays != 30 {
		t.Errorf("Expected default NotifyDays 30, got %d", cfg.NotifyDays)
	}

	// Test custom configuration
	os.Setenv("LOCALCA_DATA_DIR", "/custom/data")
	os.Setenv("LOCALCA_HOST", "example.com")
	os.Setenv("LOCALCA_PORT", "9090")
	os.Setenv("LOCALCA_HTTPS_PORT", "9443")
	os.Setenv("LOCALCA_ACME_PORT", "9555")
	os.Setenv("LOCALCA_CA_NAME", "Custom CA")
	os.Setenv("LOCALCA_CA_EMAIL", "ca@example.com")
	os.Setenv("LOCALCA_SMTP_HOST", "smtp.example.com")
	os.Setenv("LOCALCA_SMTP_PORT", "587")
	os.Setenv("LOCALCA_SMTP_USER", "user")
	os.Setenv("LOCALCA_SMTP_PASS", "pass")
	os.Setenv("LOCALCA_SMTP_FROM", "ca@example.com")
	os.Setenv("LOCALCA_NOTIFY_DAYS", "15")

	cfg = LoadConfig()

	// Check custom values
	if cfg.DataDir != "/custom/data" {
		t.Errorf("Expected custom DataDir '/custom/data', got '%s'", cfg.DataDir)
	}
	if cfg.Host != "example.com" {
		t.Errorf("Expected custom Host 'example.com', got '%s'", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("Expected custom Port 9090, got %d", cfg.Port)
	}
	if cfg.HTTPSPort != 9443 {
		t.Errorf("Expected custom HTTPSPort 9443, got %d", cfg.HTTPSPort)
	}
	if cfg.ACMEPort != 9555 {
		t.Errorf("Expected custom ACMEPort 9555, got %d", cfg.ACMEPort)
	}
	if cfg.CAName != "Custom CA" {
		t.Errorf("Expected custom CAName 'Custom CA', got '%s'", cfg.CAName)
	}
	if cfg.CAEmail != "ca@example.com" {
		t.Errorf("Expected custom CAEmail 'ca@example.com', got '%s'", cfg.CAEmail)
	}
	if cfg.SMTPHost != "smtp.example.com" {
		t.Errorf("Expected custom SMTPHost 'smtp.example.com', got '%s'", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("Expected custom SMTPPort 587, got %d", cfg.SMTPPort)
	}
	if cfg.SMTPUser != "user" {
		t.Errorf("Expected custom SMTPUser 'user', got '%s'", cfg.SMTPUser)
	}
	if cfg.SMTPPass != "pass" {
		t.Errorf("Expected custom SMTPPass 'pass', got '%s'", cfg.SMTPPass)
	}
	if cfg.SMTPFrom != "ca@example.com" {
		t.Errorf("Expected custom SMTPFrom 'ca@example.com', got '%s'", cfg.SMTPFrom)
	}
	if cfg.NotifyDays != 15 {
		t.Errorf("Expected custom NotifyDays 15, got %d", cfg.NotifyDays)
	}
}
