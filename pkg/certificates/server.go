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
	"strings"
	"time"
)

// Certificate represents a certificate
type Certificate struct {
	CommonName   string
	SerialNumber string
	NotBefore    time.Time
	NotAfter     time.Time
	Issuer       string
	IsClient     bool
	Path         string
}

// CreateServerCertificate creates a new server certificate
func (c *CertificateService) CreateServerCertificate(commonName string, additionalDomains []string) error {
	// Create directory for the certificate
	certDir := c.storage.GetCertificateDirectory(commonName)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Generate server key pair
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate server private key: %w", err)
	}

	// Save server private key to file
	serverKeyPath := c.storage.GetCertificateKeyPath(commonName)
	serverKeyFile, err := os.Create(serverKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create server key file: %w", err)
	}
	defer serverKeyFile.Close()

	if err := pem.Encode(serverKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	}); err != nil {
		return fmt.Errorf("failed to encode server private key: %w", err)
	}

	// Create server certificate template
	dnsNames := append([]string{commonName}, additionalDomains...)
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year validity
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames,
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
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverPrivKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Save server certificate to file
	serverCertPath := c.storage.GetCertificatePath(commonName)
	serverCertFile, err := os.Create(serverCertPath)
	if err != nil {
		return fmt.Errorf("failed to create server certificate file: %w", err)
	}
	defer serverCertFile.Close()

	if err := pem.Encode(serverCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes}); err != nil {
		return fmt.Errorf("failed to encode server certificate: %w", err)
	}

	// Create certificate bundle with CA
	bundlePath := c.storage.GetCertificateBundlePath(commonName)
	bundleFile, err := os.Create(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate bundle file: %w", err)
	}
	defer bundleFile.Close()

	// Write server certificate
	if err := pem.Encode(bundleFile, &pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes}); err != nil {
		return fmt.Errorf("failed to encode server certificate in bundle: %w", err)
	}

	// Write CA certificate
	if _, err := bundleFile.Write(caCertBytes); err != nil {
		return fmt.Errorf("failed to write CA certificate to bundle: %w", err)
	}

	// Set permissions
	os.Chmod(serverKeyPath, 0600)
	os.Chmod(serverCertPath, 0644)
	os.Chmod(bundlePath, 0644)

	return nil
}

// RenewServerCertificate renews an existing server certificate
func (c *CertificateService) RenewServerCertificate(commonName string) error {
	// Check if certificate exists
	certPath := c.storage.GetCertificatePath(commonName)
	keyPath := c.storage.GetCertificateKeyPath(commonName)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate does not exist: %s", commonName)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate key does not exist: %s", commonName)
	}

	// Get DNS names from existing certificate
	cmd := exec.Command(
		"openssl", "x509",
		"-in", certPath,
		"-noout",
		"-text",
	)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get certificate info: %w", err)
	}

	// Parse DNS names
	outputStr := string(output)
	dnsNames := []string{commonName}
	sanIndex := strings.Index(outputStr, "DNS:")
	if sanIndex != -1 {
		sanPart := outputStr[sanIndex:]
		endIndex := strings.Index(sanPart, "\n")
		if endIndex != -1 {
			sanList := sanPart[:endIndex]
			parts := strings.Split(sanList, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "DNS:") {
					name := strings.TrimPrefix(part, "DNS:")
					if name != commonName {
						dnsNames = append(dnsNames, name)
					}
				}
			}
		}
	}

	// Generate CSR
	csrPath := c.storage.GetCertificateDirectory(commonName) + "/" + commonName + ".csr"
	cmd = exec.Command(
		"openssl", "req",
		"-new",
		"-key", keyPath,
		"-out", csrPath,
		"-subj", "/CN="+commonName,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	// Create SAN extension config
	sanConfig := "basicConstraints = CA:FALSE\n"
	sanConfig += "nsCertType = server\n"
	sanConfig += "nsComment = \"OpenSSL Generated Server Certificate\"\n"
	sanConfig += "subjectKeyIdentifier = hash\n"
	sanConfig += "authorityKeyIdentifier = keyid,issuer:always\n"
	sanConfig += "keyUsage = critical, digitalSignature, keyEncipherment\n"
	sanConfig += "extendedKeyUsage = serverAuth\n"
	sanConfig += "subjectAltName = @alt_names\n\n"
	sanConfig += "[alt_names]\n"

	for i, name := range dnsNames {
		sanConfig += fmt.Sprintf("DNS.%d = %s\n", i+1, name)
	}

	sanConfigPath := c.storage.GetCertificateDirectory(commonName) + "/" + commonName + ".ext"
	if err := os.WriteFile(sanConfigPath, []byte(sanConfig), 0644); err != nil {
		return fmt.Errorf("failed to write SAN config: %w", err)
	}

	// Sign certificate with CA
	cmd = exec.Command(
		"openssl", "x509",
		"-req",
		"-in", csrPath,
		"-CA", c.storage.GetCAPublicKeyPath(),
		"-CAkey", c.storage.GetCAPrivateKeyPath(),
		"-CAcreateserial",
		"-out", certPath,
		"-days", "365",
		"-sha256",
		"-extfile", sanConfigPath,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to sign certificate: %w", err)
	}

	// Update bundle
	cmd = exec.Command(
		"bash", "-c",
		fmt.Sprintf("cat %s %s > %s",
			certPath,
			c.storage.GetCAPublicKeyPath(),
			c.storage.GetCertificateBundlePath(commonName),
		),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create certificate bundle: %w", err)
	}

	// Cleanup
	os.Remove(csrPath)
	os.Remove(sanConfigPath)

	return nil
}