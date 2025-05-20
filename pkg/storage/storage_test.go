package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorage(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create new storage
	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Check if storage was initialized correctly
	if storage.GetBasePath() != tempDir {
		t.Errorf("Expected BasePath '%s', got '%s'", tempDir, storage.GetBasePath())
	}

	// Check if directories were created
	if err := os.MkdirAll(filepath.Join(tempDir, "certs"), 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "private"), 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "crl"), 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "metadata"), 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Verify directories exist
	dirs := []string{
		filepath.Join(tempDir, "certs"),
		filepath.Join(tempDir, "private"),
		filepath.Join(tempDir, "crl"),
		filepath.Join(tempDir, "metadata"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory '%s' was not created", dir)
		}
	}
}

func TestSaveAndGetCAInfo(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create new storage
	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Save CA info
	caName := "Test CA"
	caKey := "testpassword"
	organization := "Test Org"
	country := "US"

	err = storage.SaveCAInfo(caName, caKey, organization, country)
	if err != nil {
		t.Fatalf("Failed to save CA info: %v", err)
	}

	// Get CA info
	retrievedName, retrievedKey, retrievedOrg, retrievedCountry, err := storage.GetCAInfo()
	if err != nil {
		t.Fatalf("Failed to get CA info: %v", err)
	}

	// Verify CA info
	if retrievedName != caName {
		t.Errorf("Expected CA name '%s', got '%s'", caName, retrievedName)
	}
	if retrievedKey != caKey {
		t.Errorf("Expected CA key '%s', got '%s'", caKey, retrievedKey)
	}
	if retrievedOrg != organization {
		t.Errorf("Expected organization '%s', got '%s'", organization, retrievedOrg)
	}
	if retrievedCountry != country {
		t.Errorf("Expected country '%s', got '%s'", country, retrievedCountry)
	}
}

func TestListCertificates(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create new storage
	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create test certificates
	for i := 1; i <= 3; i++ {
		certName := fmt.Sprintf("test-cert-%d", i)
		certDir := filepath.Join(tempDir, certName)
		if err := os.MkdirAll(certDir, 0755); err != nil {
			t.Fatalf("Failed to create certificate directory: %v", err)
		}

		// Create empty certificate file
		certFile := filepath.Join(certDir, certName+".crt")
		if err := os.WriteFile(certFile, []byte("test certificate"), 0644); err != nil {
			t.Fatalf("Failed to create certificate file: %v", err)
		}
	}

	// List certificates
	certs, err := storage.ListCertificates()
	if err != nil {
		t.Fatalf("Failed to list certificates: %v", err)
	}

	// Verify we have 3 certificates
	if len(certs) != 3 {
		t.Errorf("Expected 3 certificates, got %d", len(certs))
	}

	// Verify we can find each certificate
	for i := 1; i <= 3; i++ {
		certName := fmt.Sprintf("test-cert-%d", i)
		found := false
		for _, name := range certs {
			if name == certName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Certificate with name '%s' not found in list", certName)
		}
	}
}
