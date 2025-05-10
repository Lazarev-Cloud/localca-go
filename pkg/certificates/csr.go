package certificates

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CreateServerCertificateFromCSR creates a server certificate from a CSR
func (c *CertificateService) CreateServerCertificateFromCSR(commonName string, additionalDomains []string, csr *x509.CertificateRequest) error {
	// Validate commonName
	if !isValidName(commonName) {
		return fmt.Errorf("invalid common name: %s", commonName)
	}

	// Validate all additional domains
	for _, domain := range additionalDomains {
		if !isValidName(domain) {
			return fmt.Errorf("invalid additional domain: %s", domain)
		}
	}

	// Verify that all domains in the CSR match the domains requested
	// Create a map of all allowed domains
	allowedDomains := make(map[string]bool)
	allowedDomains[commonName] = true
	for _, domain := range additionalDomains {
		allowedDomains[domain] = true
	}

	// Check CSR Common Name
	if csr.Subject.CommonName != "" && !allowedDomains[csr.Subject.CommonName] {
		return fmt.Errorf("CSR Common Name doesn't match requested domains: %s", csr.Subject.CommonName)
	}

	// Check all DNS names in CSR
	for _, dnsName := range csr.DNSNames {
		if !allowedDomains[dnsName] {
			return fmt.Errorf("CSR contains unauthorized domain: %s", dnsName)
		}
	}

	// Create directory for the certificate
	certDir := c.storage.GetCertificateDirectory(commonName)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Write CSR to file
	csrPath := filepath.Join(certDir, commonName+".csr")
	csrOut, err := os.Create(csrPath)
	if err != nil {
		return fmt.Errorf("failed to create CSR file: %w", err)
	}
	if err := pem.Encode(csrOut, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr.Raw}); err != nil {
		csrOut.Close()
		return fmt.Errorf("failed to encode CSR: %w", err)
	}
	csrOut.Close()

	// Create list of all domains for SAN extension
	dnsNames := []string{commonName}
	for _, domain := range additionalDomains {
		dnsNames = append(dnsNames, domain)
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

	sanConfigPath := filepath.Join(certDir, commonName+".ext")
	if err := os.WriteFile(sanConfigPath, []byte(sanConfig), 0644); err != nil {
		return fmt.Errorf("failed to write SAN config: %w", err)
	}

	// Sign certificate with CA
	serverCertPath := c.storage.GetCertificatePath(commonName)
	cmd := exec.Command(
		"openssl", "x509",
		"-req",
		"-in", csrPath,
		"-CA", c.storage.GetCAPublicKeyPath(),
		"-CAkey", c.storage.GetCAPrivateKeyPath(),
		"-CAcreateserial",
		"-out", serverCertPath,
		"-days", "365",
		"-sha256",
		"-extfile", sanConfigPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to sign certificate: %w - %s", err, string(output))
	}

	// Create certificate bundle with CA
	bundlePath := c.storage.GetCertificateBundlePath(commonName)
	bundleFile, err := os.Create(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate bundle file: %w", err)
	}
	defer bundleFile.Close()

	// Write server certificate
	serverCertBytes, err := os.ReadFile(serverCertPath)
	if err != nil {
		return fmt.Errorf("failed to read server certificate: %w", err)
	}
	if _, err := bundleFile.Write(serverCertBytes); err != nil {
		return fmt.Errorf("failed to write server certificate to bundle: %w", err)
	}

	// Write CA certificate
	caCertBytes, err := os.ReadFile(c.storage.GetCAPublicKeyPath())
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}
	if _, err := bundleFile.Write(caCertBytes); err != nil {
		return fmt.Errorf("failed to write CA certificate to bundle: %w", err)
	}

	// Remove temporary files
	os.Remove(sanConfigPath)

	// Set permissions
	os.Chmod(serverCertPath, 0644)
	os.Chmod(bundlePath, 0644)

	return nil
}

// GetCertificatePath returns the path to a certificate
func (c *CertificateService) GetCertificatePath(name string) string {
	return c.storage.GetCertificatePath(name)
}

// GetCertificateBundlePath returns the path to a certificate bundle
func (c *CertificateService) GetCertificateBundlePath(name string) string {
	return c.storage.GetCertificateBundlePath(name)
}