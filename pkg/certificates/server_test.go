package certificates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateServerCertificate(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "localca-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create CA first
	ca, err := CreateCA("Test CA", "test@example.com", tempDir)
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Create server certificate
	domains := []string{"example.com", "www.example.com"}
	cert, err := CreateServerCertificate("test-server", domains, ca, tempDir)
	if err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}

	// Verify certificate properties
	if cert.CommonName != "example.com" {
		t.Errorf("Expected CommonName 'example.com', got '%s'", cert.CommonName)
	}

	// Check if files were created
	certFile := filepath.Join(tempDir, "test-server.crt")
	keyFile := filepath.Join(tempDir, "test-server.key")

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Errorf("Server certificate file was not created at %s", certFile)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Errorf("Server key file was not created at %s", keyFile)
	}

	// Verify SANs
	for _, domain := range domains {
		found := false
		for _, san := range cert.SANs {
			if san == domain {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected SAN '%s' not found in certificate", domain)
		}
	}
}
