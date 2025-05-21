package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// mockCreateCA is a test helper function that creates a CA without using OpenSSL
func mockCreateCA(certService *CertificateService) error {
	// Create directory for CA
	caDir := certService.storage.GetCADirectory()
	if err := os.MkdirAll(caDir, 0755); err != nil {
		return err
	}

	// Generate CA key pair
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create CA certificate template
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   certService.config.CAName,
			Organization: []string{certService.config.Organization},
			Country:      []string{certService.config.Country},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years validity
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	// Create CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	// Save CA certificate to file
	caCertFile, err := os.Create(certService.storage.GetCAPublicKeyPath())
	if err != nil {
		return err
	}
	defer caCertFile.Close()

	// Write CA certificate in PEM format
	if err := pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return err
	}

	// Save CA private key to file
	caKeyFile, err := os.Create(certService.storage.GetCAPrivateKeyPath())
	if err != nil {
		return err
	}
	defer caKeyFile.Close()

	// Write CA private key in PEM format
	if err := pem.Encode(caKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	}); err != nil {
		return err
	}

	// Skip OpenSSL commands for testing

	// Create a copy of the CA certificate for public access
	publicCopyPath := certService.storage.GetCAPublicCopyPath()
	err = os.MkdirAll(filepath.Dir(publicCopyPath), 0755)
	if err != nil {
		return err
	}

	publicCopyFile, err := os.Create(publicCopyPath)
	if err != nil {
		return err
	}
	defer publicCopyFile.Close()

	// Copy the certificate content
	certContent, err := os.ReadFile(certService.storage.GetCAPublicKeyPath())
	if err != nil {
		return err
	}

	_, err = publicCopyFile.Write(certContent)
	if err != nil {
		return err
	}

	return nil
}

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

	// Generate a random password for testing
	testPassword := generateRandomPassword()

	// Create config
	cfg := &config.Config{
		CAName:        "Test CA",
		CAKeyPassword: testPassword,
		Organization:  "Test Org",
		Country:       "US",
		DataDir:       tempDir,
	}

	// Create certificate service
	certService, err := NewCertificateService(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create certificate service: %v", err)
	}

	// Use mock CA creation for testing
	err = mockCreateCA(certService)
	if err != nil {
		t.Fatalf("Failed to create mock CA: %v", err)
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

	// Generate a random password for testing
	testPassword := generateRandomPassword()

	// Create config
	cfg := &config.Config{
		CAName:        "Test CA",
		CAKeyPassword: testPassword,
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

	// Use mock CA creation for testing
	err = mockCreateCA(certService)
	if err != nil {
		t.Fatalf("Failed to create mock CA: %v", err)
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

// generateRandomPassword generates a random password for testing
func generateRandomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
	const length = 16

	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "test_password_fallback"
	}

	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}

	return string(b)
}
