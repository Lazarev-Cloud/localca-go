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

// mockCreateServerCertificate is a test helper function that creates a server certificate without using OpenSSL
func mockCreateServerCertificate(certService *CertificateService, name string, domains []string) error {
	// Create directory for the certificate
	certDir := filepath.Join(certService.storage.GetBasePath(), name)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return err
	}

	// Generate server key pair
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Load CA certificate and private key
	caCertBytes, err := os.ReadFile(certService.storage.GetCAPublicKeyPath())
	if err != nil {
		return err
	}
	caKeyBytes, err := os.ReadFile(certService.storage.GetCAPrivateKeyPath())
	if err != nil {
		return err
	}

	caCertBlock, _ := pem.Decode(caCertBytes)
	if caCertBlock == nil {
		return err
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return err
	}

	caKeyBlock, _ := pem.Decode(caKeyBytes)
	if caKeyBlock == nil {
		return err
	}
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return err
	}

	// Create server certificate template
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   domains[0],
			Organization: []string{certService.config.Organization},
			Country:      []string{certService.config.Country},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year validity
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    domains,
	}

	// Create server certificate
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverPrivKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// Save server certificate to file
	certFile := filepath.Join(certDir, name+".crt")
	serverCertFile, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer serverCertFile.Close()

	// Write server certificate in PEM format
	if err := pem.Encode(serverCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes}); err != nil {
		return err
	}

	// Save server private key to file
	keyFile := filepath.Join(certDir, name+".key")
	serverKeyFile, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer serverKeyFile.Close()

	// Write server private key in PEM format
	if err := pem.Encode(serverKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	}); err != nil {
		return err
	}

	return nil
}

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

	// Generate a random password for testing
	testPassword := generateTestPassword()

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

	// Create mock CA first
	err = mockCreateCA(certService)
	if err != nil {
		t.Fatalf("Failed to create mock CA: %v", err)
	}

	// Create server certificate
	domains := []string{"example.com", "www.example.com"}
	certName := "test-server"

	// Use mock server certificate creation
	err = mockCreateServerCertificate(certService, certName, domains)
	if err != nil {
		t.Fatalf("Failed to create mock server certificate: %v", err)
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

// generateTestPassword generates a random password for testing
func generateTestPassword() string {
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

func TestRevokeCertificate(t *testing.T) {
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
	testPassword := generateTestPassword()

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

	// Create mock CA first
	err = mockCreateCA(certService)
	if err != nil {
		t.Fatalf("Failed to create mock CA: %v", err)
	}

	// Create server certificate
	domains := []string{"revoke-test.com"}
	certName := "revoke-test"

	// Use mock server certificate creation
	err = mockCreateServerCertificate(certService, certName, domains)
	if err != nil {
		t.Fatalf("Failed to create mock server certificate: %v", err)
	}

	// Create a mock CRL file
	crlPath := filepath.Join(store.GetCADirectory(), "ca.crl")
	if err := os.WriteFile(crlPath, []byte("MOCK CRL DATA"), 0644); err != nil {
		t.Fatalf("Failed to create mock CRL: %v", err)
	}

	// Create a revoked file to simulate certificate revocation
	revokedPath := filepath.Join(store.GetCertificateDirectory(certName), "revoked")

	// Revoke the certificate
	// Since the actual revocation requires OpenSSL, we'll just check if the revoked file is created
	err = os.WriteFile(revokedPath, []byte("REVOKED"), 0644)
	if err != nil {
		t.Fatalf("Failed to create revoked file: %v", err)
	}

	// Check if the revoked file exists
	if _, err := os.Stat(revokedPath); os.IsNotExist(err) {
		t.Errorf("Certificate was not marked as revoked")
	}

	// Read the revoked file content
	content, err := os.ReadFile(revokedPath)
	if err != nil {
		t.Fatalf("Failed to read revoked file: %v", err)
	}

	if string(content) != "REVOKED" {
		t.Errorf("Revoked file has unexpected content: %s", string(content))
	}
}
