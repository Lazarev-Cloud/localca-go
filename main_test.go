package main

import (
	"os"
	"testing"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// TestMainCompilation is a simple test to ensure the main package compiles correctly
func TestMainCompilation(t *testing.T) {
	// This test doesn't actually test any functionality
	// It's just to ensure that the main package compiles correctly
	// The fact that this test runs means the main package compiled successfully
}

// TestConfigLoading tests that the configuration can be loaded
func TestConfigLoading(t *testing.T) {
	// Save original environment variables
	origEnvVars := map[string]string{
		"CA_NAME":      os.Getenv("CA_NAME"),
		"CA_KEY":       os.Getenv("CA_KEY"),
		"ORGANIZATION": os.Getenv("ORGANIZATION"),
		"COUNTRY":      os.Getenv("COUNTRY"),
		"DATA_DIR":     os.Getenv("DATA_DIR"),
		"LISTEN_ADDR":  os.Getenv("LISTEN_ADDR"),
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

	// Set test environment variables
	os.Setenv("CA_NAME", "Test CA")
	os.Setenv("CA_KEY", "testpassword")
	os.Setenv("ORGANIZATION", "Test Org")
	os.Setenv("COUNTRY", "US")
	os.Setenv("DATA_DIR", "./testdata")
	os.Setenv("LISTEN_ADDR", ":9090")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify configuration
	if cfg.CAName != "Test CA" {
		t.Errorf("Expected CAName 'Test CA', got '%s'", cfg.CAName)
	}
	if cfg.CAKeyPassword != "testpassword" {
		t.Errorf("Expected CAKeyPassword 'testpassword', got '%s'", cfg.CAKeyPassword)
	}
	if cfg.Organization != "Test Org" {
		t.Errorf("Expected Organization 'Test Org', got '%s'", cfg.Organization)
	}
	if cfg.Country != "US" {
		t.Errorf("Expected Country 'US', got '%s'", cfg.Country)
	}
	if cfg.DataDir != "./testdata" {
		t.Errorf("Expected DataDir './testdata', got '%s'", cfg.DataDir)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("Expected ListenAddr ':9090', got '%s'", cfg.ListenAddr)
	}
}

// TestStorageInitialization tests that storage can be initialized
func TestStorageInitialization(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Verify storage
	if store.GetBasePath() != tempDir {
		t.Errorf("Expected BasePath '%s', got '%s'", tempDir, store.GetBasePath())
	}

	// Verify the base directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Base directory '%s' was not created", tempDir)
	}
}
