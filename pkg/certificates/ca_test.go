package certificates

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

func TestCertificateService_CreateCA(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create config
	cfg := &config.Config{
		CAName:        "Test CA",
		CAKeyPassword: "testpassword",
		Organization:  "Test Org",
		Country:       "US",
		DataDir:       tempDir,
	}

	// Create certificate service
	certService, err := NewCertificateService(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create certificate service: %v", err)
	}

	// Create CA
	err = certService.CreateCA()
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Check if files were created
	certFile := store.GetCAPublicKeyPath()
	keyFile := store.GetCAPrivateKeyPath()

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Errorf("CA certificate file was not created at %s", certFile)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Errorf("CA key file was not created at %s", keyFile)
	}
}

func TestCertificateService_CAExists(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create config
	cfg := &config.Config{
		CAName:        "Test CA",
		CAKeyPassword: "testpassword",
		Organization:  "Test Org",
		Country:       "US",
		DataDir:       tempDir,
	}

	// Create certificate service
	certService, err := NewCertificateService(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create certificate service: %v", err)
	}

	// Check if CA exists (should not exist yet)
	exists, err := certService.CAExists()
	if err != nil {
		t.Fatalf("Failed to check if CA exists: %v", err)
	}
	if exists {
		t.Errorf("CA should not exist yet")
	}

	// Create CA
	err = certService.CreateCA()
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Check if CA exists (should exist now)
	exists, err = certService.CAExists()
	if err != nil {
		t.Fatalf("Failed to check if CA exists: %v", err)
	}
	if !exists {
		t.Errorf("CA should exist now")
	}
}

func TestCreateCA(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create CA
	ca, err := CreateCA("Test CA", "test@example.com", tempDir)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Verify CA properties
	if ca.CommonName != "Test CA" {
		t.Errorf("Expected CommonName 'Test CA', got '%s'", ca.CommonName)
	}

	if ca.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", ca.Email)
	}

	// Check if files were created
	certFile := filepath.Join(tempDir, "ca.crt")
	keyFile := filepath.Join(tempDir, "ca.key")

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Errorf("CA certificate file was not created at %s", certFile)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Errorf("CA key file was not created at %s", keyFile)
	}
}

func TestLoadCA(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create CA
	_, err = CreateCA("Test CA", "test@example.com", tempDir)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Load CA
	loadedCA, err := LoadCA(tempDir)
	if err != nil {
		t.Fatalf("Failed to load CA: %v", err)
	}

	// Verify loaded CA properties
	if loadedCA.CommonName != "Test CA" {
		t.Errorf("Expected CommonName 'Test CA', got '%s'", loadedCA.CommonName)
	}

	if loadedCA.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", loadedCA.Email)
	}
}

func TestCAExpiryTime(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create CA
	ca, err := CreateCA("Test CA", "test@example.com", tempDir)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Check expiry time (should be 10 years from now, with some tolerance)
	expectedExpiry := time.Now().AddDate(10, 0, 0)
	tolerance := 24 * time.Hour // 1 day tolerance

	diff := ca.ExpiryTime.Sub(expectedExpiry)
	if diff < -tolerance || diff > tolerance {
		t.Errorf("CA expiry time is not within expected range. Got %v, expected around %v",
			ca.ExpiryTime, expectedExpiry)
	}
}
