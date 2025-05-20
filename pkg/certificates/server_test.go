package certificates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

func TestCreateServerCertificate(t *testing.T) {
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

	// Create CA first
	err = certService.CreateCA()
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Create server certificate
	domains := []string{"example.com", "www.example.com"}
	certName := "test-server"
	err = certService.CreateServerCertificate(certName, domains)
	if err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}

	// Check if files were created
	certFile := filepath.Join(tempDir, certName, certName+".crt")
	keyFile := filepath.Join(tempDir, certName, certName+".key")

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Errorf("Server certificate file was not created at %s", certFile)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Errorf("Server key file was not created at %s", keyFile)
	}
}
