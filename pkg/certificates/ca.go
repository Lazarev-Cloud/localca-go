package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
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

	// Copy CA certificate to public location using safe Go file operations
	if err := safeCopyFile(c.storage.GetCAPublicKeyPath(), c.storage.GetCAPublicCopyPath()); err != nil {
		return fmt.Errorf("failed to copy CA certificate: %w", err)
	}

	// Set appropriate permissions with proper error handling
	const (
		privateFileMode = 0600 // Read/write for owner only
		publicFileMode  = 0644 // Read for everyone, write for owner
	)

	if err := os.Chmod(c.storage.GetCAPrivateKeyPath(), privateFileMode); err != nil {
		return fmt.Errorf("failed to set permissions on CA private key: %w", err)
	}
	if err := os.Chmod(c.storage.GetCAEncryptedKeyPath(), privateFileMode); err != nil {
		return fmt.Errorf("failed to set permissions on CA encrypted key: %w", err)
	}
	if err := os.Chmod(c.storage.GetCAPublicKeyPath(), publicFileMode); err != nil {
		return fmt.Errorf("failed to set permissions on CA public key: %w", err)
	}
	if err := os.Chmod(c.storage.GetCAPublicCopyPath(), publicFileMode); err != nil {
		return fmt.Errorf("failed to set permissions on CA public copy: %w", err)
	}

	return nil
}

// RenewCA renews the CA certificate
func (c *CertificateService) RenewCA() error {
	// Use full path to OpenSSL for security
	opensslPath, err := exec.LookPath("openssl")
	if err != nil {
		return fmt.Errorf("failed to find openssl executable: %w", err)
	}

	// Create a CSR from the current CA certificate and key
	cmd := exec.Command(
		opensslPath, "x509",
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
		opensslPath, "x509",
		"-req",
		"-days", "3650", // 10 years
		"-in", c.storage.GetCADirectory()+"/ca.csr",
		"-signkey", c.storage.GetCAPrivateKeyPath(),
		"-out", c.storage.GetCADirectory()+"/ca-new.pem",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create new CA certificate: %w", err)
	}

	// Replace existing CA certificate - use Go's file operations instead of exec
	srcFile, err := os.ReadFile(c.storage.GetCADirectory() + "/ca-new.pem")
	if err != nil {
		return fmt.Errorf("failed to read new CA certificate: %w", err)
	}

	if err := os.WriteFile(c.storage.GetCAPublicKeyPath(), srcFile, 0644); err != nil {
		return fmt.Errorf("failed to replace CA certificate: %w", err)
	}

	// Copy to public location - use Go's file operations instead of exec
	if err := os.WriteFile(c.storage.GetCAPublicCopyPath(), srcFile, 0644); err != nil {
		return fmt.Errorf("failed to copy new CA certificate: %w", err)
	}

	return nil
}

// CreateServiceCertificate creates a certificate for the LocalCA service itself
func (c *CertificateService) CreateServiceCertificate() error {
	// Get hostname for the service
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// Create directory for the service certificate
	certDir := filepath.Join(c.storage.GetBasePath(), "service")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create service certificate directory: %w", err)
	}

	// Create additional domains for the service certificate
	// Include common local names and the CA name itself
	additionalDomains := []string{
		"localhost",
		"localca",
		"localca.local",
		c.config.CAName,
	}

	// If hostname is not already in the list, add it
	hostnameExists := false
	for _, domain := range additionalDomains {
		if domain == hostname {
			hostnameExists = true
			break
		}
	}
	if !hostnameExists {
		additionalDomains = append(additionalDomains, hostname)
	}

	// Add IP addresses as well
	// Get all IP addresses of the host
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			// Check if it's an IP address and not a network prefix
			ipnet, ok := addr.(*net.IPNet)
			if ok && !ipnet.IP.IsLoopback() {
				// Only add IPv4 addresses for now
				if ipnet.IP.To4() != nil {
					additionalDomains = append(additionalDomains, ipnet.IP.String())
				}
			}
		}
		// Always add localhost IP
		additionalDomains = append(additionalDomains, "127.0.0.1")
	}

	// Generate server key pair
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate service private key: %w", err)
	}

	// Save server private key to file
	serverKeyPath := filepath.Join(c.storage.GetBasePath(), "service.key")
	serverKeyFile, err := os.Create(serverKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create service key file: %w", err)
	}
	defer serverKeyFile.Close()

	if err := pem.Encode(serverKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	}); err != nil {
		return fmt.Errorf("failed to encode service private key: %w", err)
	}

	// Create server certificate template
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: "localca-service",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(3, 0, 0), // 3 years validity
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    additionalDomains,
	}

	// Load CA certificate
	caCertBytes, err := os.ReadFile(c.storage.GetCAPublicKeyPath())
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertBlock, _ := pem.Decode(caCertBytes)
	if caCertBlock == nil {
		return fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load CA private key
	caKeyBytes, err := os.ReadFile(c.storage.GetCAPrivateKeyPath())
	if err != nil {
		return fmt.Errorf("failed to read CA private key: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyBytes)
	if caKeyBlock == nil {
		return fmt.Errorf("failed to decode CA private key PEM")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Create server certificate
	serverCertBytes, err := x509.CreateCertificate(
		rand.Reader,
		&serverTemplate,
		caCert,
		&serverPrivKey.PublicKey,
		caKey,
	)
	if err != nil {
		return fmt.Errorf("failed to create service certificate: %w", err)
	}

	// Save server certificate to file
	serverCertPath := filepath.Join(c.storage.GetBasePath(), "service.crt")
	serverCertFile, err := os.Create(serverCertPath)
	if err != nil {
		return fmt.Errorf("failed to create service certificate file: %w", err)
	}
	defer serverCertFile.Close()

	if err := pem.Encode(serverCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes}); err != nil {
		return fmt.Errorf("failed to encode service certificate: %w", err)
	}

	// Set proper permissions
	const (
		privateFileMode = 0600 // Read/write for owner only
		publicFileMode  = 0644 // Read for everyone, write for owner
	)

	if err := os.Chmod(serverKeyPath, privateFileMode); err != nil {
		return fmt.Errorf("failed to set permissions on service key: %w", err)
	}
	if err := os.Chmod(serverCertPath, publicFileMode); err != nil {
		return fmt.Errorf("failed to set permissions on service certificate: %w", err)
	}

	return nil
}

// GetAllCertificates returns a list of all issued certificates
func (c *CertificateService) GetAllCertificates() ([]Certificate, error) {
	// Get the list of certificate names
	certNames, err := c.storage.ListCertificates()
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %w", err)
	}

	// Get certificate details for each
	certificates := make([]Certificate, 0, len(certNames))
	for _, name := range certNames {
		cert, err := c.GetCertificateInfo(name)
		if err != nil {
			continue
		}
		certificates = append(certificates, *cert)
	}

	return certificates, nil
}

// GetCertificateInfo returns information about a specific certificate
func (c *CertificateService) GetCertificateInfo(name string) (*Certificate, error) {
	certPath := c.storage.GetCertificatePath(name)

	// Check if certificate exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate does not exist: %s", name)
	}

	// Use OpenSSL to get certificate details
	cmd := exec.Command(
		"openssl", "x509",
		"-in", certPath,
		"-noout",
		"-text",
		"-serial",
	)
	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate info: %w", err)
	}

	// Parse the output to create a Certificate struct
	// This is a simplified implementation and may need to be expanded
	cert := &Certificate{
		CommonName: name,
		Path:       certPath,
	}

	// Check if it's a client certificate
	p12Path := c.storage.GetCertificateP12Path(name)
	cert.IsClient = false
	if _, err := os.Stat(p12Path); err == nil {
		cert.IsClient = true
	}

	// Extract details from output
	// This would require parsing the OpenSSL output
	// Full implementation would need to extract SerialNumber, NotBefore, NotAfter, etc.

	return cert, nil
}

// safeCopyFile safely copies a file from src to dst using Go standard library
func safeCopyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
