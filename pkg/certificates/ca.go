package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"time"

	"github.com/yourusername/localca-go/pkg/config"
	"github.com/yourusername/localca-go/pkg/storage"
)

// CertificateService handles certificate operations
type CertificateService struct {
	config  *config.Config
	storage *storage.Storage
}

// NewCertificateService creates a new certificate service
func NewCertificateService(cfg *config.Config, store *storage.Storage) (*CertificateService, error) {
	return &CertificateService{
		config:  cfg,
		storage: store,
	}, nil
}

// CAExists checks if CA certificate exists
func (c *CertificateService) CAExists() (bool, error) {
	caCertPath := c.storage.GetCAPublicKeyPath()
	caKeyPath := c.storage.GetCAPrivateKeyPath()

	// Check if both files exist
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		return false, nil
	}
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

// CreateCA creates a new CA certificate
func (c *CertificateService) CreateCA() error {
	// Save CA info
	if err := c.storage.SaveCAInfo(c.config.CAName, c.config.CAKeyPassword, c.config.Organization, c.config.Country); err != nil {
		return fmt.Errorf("failed to save CA info: %w", err)
	}

	// Create directory for CA
	caDir := c.storage.GetCADirectory()
	if err := os.MkdirAll(caDir, 0755); err != nil {
		return fmt.Errorf("failed to create CA directory: %w", err)
	}

	// Generate CA key pair
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// Create CA certificate template
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   c.config.CAName,
			Organization: []string{c.config.Organization},
			Country:      []string{c.config.Country},
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
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Save CA certificate to file
	caCertFile, err := os.Create(c.storage.GetCAPublicKeyPath())
	if err != nil {
		return fmt.Errorf("failed to create CA certificate file: %w", err)
	}
	defer caCertFile.Close()

	// Write CA certificate in PEM format
	if err := pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return fmt.Errorf("failed to encode CA certificate: %w", err)
	}

	// Save CA private key to file
	caKeyFile, err := os.Create(c.storage.GetCAPrivateKeyPath())
	if err != nil {
		return fmt.Errorf("failed to create CA key file: %w", err)
	}
	defer caKeyFile.Close()

	// Write CA private key in PEM format
	if err := pem.Encode(caKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	}); err != nil {
		return fmt.Errorf("failed to encode CA private key: %w", err)
	}

	// Create an encrypted version of the private key using OpenSSL (for compatibility)
	cmd := exec.Command(
		"openssl", "rsa", 
		"-in", c.storage.GetCAPrivateKeyPath(),
		"-out", c.storage.GetCAEncryptedKeyPath(),
		"-aes256", 
		"-passout", fmt.Sprintf("pass:%s", c.config.CAKeyPassword),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to encrypt CA private key: %w", err)
	}

	// Copy CA certificate to public location
	cmd = exec.Command("cp", c.storage.GetCAPublicKeyPath(), c.storage.GetCAPublicCopyPath())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy CA certificate: %w", err)
	}

	// Set appropriate permissions
	os.Chmod(c.storage.GetCAPrivateKeyPath(), 0600)
	os.Chmod(c.storage.GetCAEncryptedKeyPath(), 0600)
	os.Chmod(c.storage.GetCAPublicKeyPath(), 0644)
	os.Chmod(c.storage.GetCAPublicCopyPath(), 0644)

	return nil
}

// RenewCA renews the CA certificate
func (c *CertificateService) RenewCA() error {
	// Create a CSR from the current CA certificate and key
	cmd := exec.Command(
		"openssl", "x509", 
		"-x509toreq",
		"-in", c.storage.GetCAPublicKeyPath(),
		"-signkey", c.storage.GetCAPrivateKeyPath(),
		"-out", c.storage.GetCADirectory()+"/ca.csr",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create CSR from CA: %w", err)
	}

	// Create a new CA certificate with extended validity
	cmd = exec.Command(
		"openssl", "x509",
		"-req",
		"-days", "3650", // 10 years
		"-in", c.storage.GetCADirectory()+"/ca.csr",
		"-signkey", c.storage.GetCAPrivateKeyPath(),
		"-out", c.storage.GetCADirectory()+"/ca-new.pem",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create new CA certificate: %w", err)
	}

	// Replace existing CA certificate
	cmd = exec.Command("mv", c.storage.GetCADirectory()+"/ca-new.pem", c.storage.GetCAPublicKeyPath())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to replace CA certificate: %w", err)
	}

	// Copy to public location
	cmd = exec.Command("cp", c.storage.GetCAPublicKeyPath(), c.storage.GetCAPublicCopyPath())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy new CA certificate: %w", err)
	}

	return nil
}

// CreateServiceCertificate creates a certificate for the LocalCA service itself
func (c *CertificateService) CreateServiceCertificate() error {
	// Implementation to create a certificate for the service
	// This would use the CA to sign a certificate for the web service
	// Similar to CreateServerCertificate but for the service itself
	return nil
}

// GetAllCertificates returns a list of all issued certificates
func (c *CertificateService) GetAllCertificates() ([]Certificate, error) {
	// Implementation to list all certificates
	return nil, nil
}

// GetCertificateInfo returns information about a specific certificate
func (c *CertificateService) GetCertificateInfo(name string) (*Certificate, error) {
	// Implementation to get certificate details
	return nil, nil
}