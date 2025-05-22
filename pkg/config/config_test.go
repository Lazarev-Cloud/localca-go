package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables
	origEnvVars := map[string]string{
		"CA_NAME":       os.Getenv("CA_NAME"),
		"CA_KEY_FILE":   os.Getenv("CA_KEY_FILE"),
		"CA_KEY":        os.Getenv("CA_KEY"),
		"ORGANIZATION":  os.Getenv("ORGANIZATION"),
		"COUNTRY":       os.Getenv("COUNTRY"),
		"DATA_DIR":      os.Getenv("DATA_DIR"),
		"LISTEN_ADDR":   os.Getenv("LISTEN_ADDR"),
		"EMAIL_NOTIFY":  os.Getenv("EMAIL_NOTIFY"),
		"SMTP_SERVER":   os.Getenv("SMTP_SERVER"),
		"SMTP_PORT":     os.Getenv("SMTP_PORT"),
		"SMTP_USER":     os.Getenv("SMTP_USER"),
		"SMTP_PASSWORD": os.Getenv("SMTP_PASSWORD"),
		"SMTP_USE_TLS":  os.Getenv("SMTP_USE_TLS"),
		"EMAIL_FROM":    os.Getenv("EMAIL_FROM"),
		"EMAIL_TO":      os.Getenv("EMAIL_TO"),
		"TLS_ENABLED":   os.Getenv("TLS_ENABLED"),
		"ALLOW_LOCALHOST": os.Getenv("ALLOW_LOCALHOST"),
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
	os.Unsetenv("CA_NAME")
	os.Unsetenv("CA_KEY_FILE")
	os.Unsetenv("CA_KEY")
	os.Unsetenv("ORGANIZATION")
	os.Unsetenv("COUNTRY")
	os.Unsetenv("DATA_DIR")
	os.Unsetenv("LISTEN_ADDR")
	os.Unsetenv("EMAIL_NOTIFY")
	os.Unsetenv("SMTP_SERVER")
	os.Unsetenv("SMTP_PORT")
	os.Unsetenv("SMTP_USER")
	os.Unsetenv("SMTP_PASSWORD")
	os.Unsetenv("SMTP_USE_TLS")
	os.Unsetenv("EMAIL_FROM")
	os.Unsetenv("EMAIL_TO")
	os.Unsetenv("TLS_ENABLED")
	os.Unsetenv("ALLOW_LOCALHOST")

	// Set required CA_KEY to avoid error
	os.Setenv("CA_KEY", "testpassword")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check default values
	if cfg.CAName != "LocalCA" {
		t.Errorf("Expected default CAName 'LocalCA', got '%s'", cfg.CAName)
	}
	if cfg.Organization != "LocalCA Organization" {
		t.Errorf("Expected default Organization 'LocalCA Organization', got '%s'", cfg.Organization)
	}
	if cfg.Country != "US" {
		t.Errorf("Expected default Country 'US', got '%s'", cfg.Country)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("Expected default DataDir './data', got '%s'", cfg.DataDir)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("Expected default ListenAddr ':8080', got '%s'", cfg.ListenAddr)
	}
	if cfg.CAKeyPassword != "testpassword" {
		t.Errorf("Expected CAKeyPassword 'testpassword', got '%s'", cfg.CAKeyPassword)
	}
	if cfg.AllowLocalhost {
		t.Errorf("Expected default AllowLocalhost to be false")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	customDataDir := filepath.Join(tempDir, "custom-data")

	// Test custom configuration
	os.Setenv("CA_NAME", "Custom CA")
	os.Setenv("CA_KEY", "custompassword")
	os.Setenv("ORGANIZATION", "Custom Organization")
	os.Setenv("COUNTRY", "DE")
	os.Setenv("DATA_DIR", customDataDir)
	os.Setenv("LISTEN_ADDR", ":9090")
	os.Setenv("EMAIL_NOTIFY", "true")
	os.Setenv("SMTP_SERVER", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USER", "user")
	os.Setenv("SMTP_PASSWORD", "pass")
	os.Setenv("SMTP_USE_TLS", "true")
	os.Setenv("EMAIL_FROM", "ca@example.com")
	os.Setenv("EMAIL_TO", "admin@example.com")
	os.Setenv("TLS_ENABLED", "true")
	os.Setenv("ALLOW_LOCALHOST", "true")

	cfg, err = LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check custom values
	if cfg.CAName != "Custom CA" {
		t.Errorf("Expected custom CAName 'Custom CA', got '%s'", cfg.CAName)
	}
	if cfg.Organization != "Custom Organization" {
		t.Errorf("Expected custom Organization 'Custom Organization', got '%s'", cfg.Organization)
	}
	if cfg.Country != "DE" {
		t.Errorf("Expected custom Country 'DE', got '%s'", cfg.Country)
	}
	if cfg.DataDir != customDataDir {
		t.Errorf("Expected custom DataDir '%s', got '%s'", customDataDir, cfg.DataDir)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("Expected custom ListenAddr ':9090', got '%s'", cfg.ListenAddr)
	}
	if cfg.CAKeyPassword != "custompassword" {
		t.Errorf("Expected custom CAKeyPassword 'custompassword', got '%s'", cfg.CAKeyPassword)
	}
	if !cfg.EmailEnabled {
		t.Errorf("Expected EmailEnabled to be true")
	}
	if cfg.SMTPServer != "smtp.example.com" {
		t.Errorf("Expected custom SMTPServer 'smtp.example.com', got '%s'", cfg.SMTPServer)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("Expected custom SMTPPort 587, got %d", cfg.SMTPPort)
	}
	if cfg.SMTPUser != "user" {
		t.Errorf("Expected custom SMTPUser 'user', got '%s'", cfg.SMTPUser)
	}
	if cfg.SMTPPassword != "pass" {
		t.Errorf("Expected custom SMTPPassword 'pass', got '%s'", cfg.SMTPPassword)
	}
	if !cfg.SMTPUseTLS {
		t.Errorf("Expected SMTPUseTLS to be true")
	}
	if cfg.EmailFrom != "ca@example.com" {
		t.Errorf("Expected custom EmailFrom 'ca@example.com', got '%s'", cfg.EmailFrom)
	}
	if cfg.EmailTo != "admin@example.com" {
		t.Errorf("Expected custom EmailTo 'admin@example.com', got '%s'", cfg.EmailTo)
	}
	if !cfg.TLSEnabled {
		t.Errorf("Expected TLSEnabled to be true")
	}
	if !cfg.AllowLocalhost {
		t.Errorf("Expected AllowLocalhost to be true")
	}
}
