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
	"path/filepath"
	"strings"
	"time"
)

// CreateClientCertificate creates a new client certificate and p12 file
func (c *CertificateService) CreateClientCertificate(commonName string, p12Password string) error {
	// Create directory for the certificate
	certDir := c.storage.GetCertificateDirectory(commonName)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Save p12 password
	if err := os.WriteFile(c.storage.GetCertificatePasswordPath(commonName), []byte(p12Password), 0600); err != nil {
		return fmt.Errorf("failed to save certificate password: %w", err)
	}

	// Generate client key pair
	clientPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate client private key: %w", err)
	}

	// Save client private key to file
	clientKeyPath := c.storage.GetCertificateKeyPath(commonName)
	clientKeyFile, err := os.Create(clientKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create client key file: %w", err)
	}
	defer clientKeyFile.Close()

	if err := pem.Encode(clientKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientPrivKey),
	}); err != nil {
		return fmt.Errorf("failed to encode client private key: %w", err)
	}

	// Create client certificate template
	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year validity
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageEmailProtection},
		DNSNames:    []string{commonName},
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

	// Create client certificate
	clientCertBytes, err := x509.CreateCertificate(rand.Reader, &clientTemplate, caCert, &clientPrivKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create client certificate: %w", err)
	}

	// Save client certificate to file
	clientCertPath := c.storage.GetCertificatePath(commonName)
	clientCertFile, err := os.Create(clientCertPath)
	if err != nil {
		return fmt.Errorf("failed to create client certificate file: %w", err)
	}
	defer clientCertFile.Close()

	if err := pem.Encode(clientCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: clientCertBytes}); err != nil {
		return fmt.Errorf("failed to encode client certificate: %w", err)
	}

	// Create PKCS#12 file
	p12Path := c.storage.GetCertificateP12Path(commonName)
	cmd := exec.Command(
		"openssl", "pkcs12",
		"-export",
		"-out", p12Path,
		"-inkey", clientKeyPath,
		"-in", clientCertPath,
		"-certfile", c.storage.GetCAPublicKeyPath(),
		"-passout", fmt.Sprintf("pass:%s", p12Password),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create PKCS#12 file: %w", err)
	}

	// Set permissions
	os.Chmod(clientKeyPath, 0600)
	os.Chmod(clientCertPath, 0644)
	os.Chmod(p12Path, 0644)

	return nil
}

// RenewClientCertificate renews an existing client certificate
func (c *CertificateService) RenewClientCertificate(commonName string) error {
	// Check if certificate exists
	certPath := c.storage.GetCertificatePath(commonName)
	keyPath := c.storage.GetCertificateKeyPath(commonName)
	p12Path := c.storage.GetCertificateP12Path(commonName)
	passwordPath := c.storage.GetCertificatePasswordPath(commonName)

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate does not exist: %s", commonName)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate key does not exist: %s", commonName)
	}
	if _, err := os.Stat(p12Path); os.IsNotExist(err) {
		return fmt.Errorf("p12 file does not exist: %s", commonName)
	}
	if _, err := os.Stat(passwordPath); os.IsNotExist(err) {
		return fmt.Errorf("password file does not exist: %s", commonName)
	}

	// Read p12 password
	passwordBytes, err := os.ReadFile(passwordPath)
	if err != nil {
		return fmt.Errorf("failed to read password file: %w", err)
	}
	p12Password := string(passwordBytes)

	// Generate CSR
	csrPath := c.storage.GetCertificateDirectory(commonName) + "/" + commonName + ".csr"
	cmd := exec.Command(
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
	sanConfig += "nsCertType = client, email\n"
	sanConfig += "nsComment = \"OpenSSL Generated Client Certificate\"\n"
	sanConfig += "subjectKeyIdentifier = hash\n"
	sanConfig += "authorityKeyIdentifier = keyid,issuer\n"
	sanConfig += "keyUsage = critical, nonRepudiation, digitalSignature, keyEncipherment\n"
	sanConfig += "extendedKeyUsage = clientAuth, emailProtection\n"
	sanConfig += "subjectAltName = @alt_names\n\n"
	sanConfig += "[alt_names]\n"
	sanConfig += fmt.Sprintf("DNS.1 = %s\n", commonName)

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

	// Create PKCS#12 file
	cmd = exec.Command(
		"openssl", "pkcs12",
		"-export",
		"-out", p12Path,
		"-inkey", keyPath,
		"-in", certPath,
		"-certfile", c.storage.GetCAPublicKeyPath(),
		"-passout", fmt.Sprintf("pass:%s", p12Password),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create PKCS#12 file: %w", err)
	}

	// Cleanup
	os.Remove(csrPath)
	os.Remove(sanConfigPath)

	return nil
}

// RevokeCertificate revokes a certificate and updates the CRL
func (c *CertificateService) RevokeCertificate(commonName string) error {
    // Check if certificate exists
    certPath := c.storage.GetCertificatePath(commonName)
    if _, err := os.Stat(certPath); os.IsNotExist(err) {
        return fmt.Errorf("certificate not found: %s", commonName)
    }

    // Get serial number of the certificate
    cmd := exec.Command(
        "openssl", "x509",
        "-in", certPath,
        "-noout",
        "-serial",
    )
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to get certificate serial number: %w", err)
    }

    serialLine := string(output)
    serialParts := strings.SplitN(serialLine, "=", 2)
    if len(serialParts) != 2 {
        return fmt.Errorf("invalid serial number format: %s", serialLine)
    }

    serial := strings.TrimSpace(serialParts[1])

    // Initialize CRL directory if it doesn't exist
    crlDir := filepath.Join(c.storage.GetCADirectory(), "crl")
    if err := os.MkdirAll(crlDir, 0755); err != nil {
        return fmt.Errorf("failed to create CRL directory: %w", err)
    }

    // Create or update CRL index file
    indexPath := filepath.Join(crlDir, "index.txt")
    if _, err := os.Stat(indexPath); os.IsNotExist(err) {
        // Create empty index file if it doesn't exist
        if err := os.WriteFile(indexPath, []byte(""), 0644); err != nil {
            return fmt.Errorf("failed to create CRL index file: %w", err)
        }
    }

    // Create openssl.cnf for CRL generation if it doesn't exist
    opensslCnfPath := filepath.Join(crlDir, "openssl.cnf")
    if _, err := os.Stat(opensslCnfPath); os.IsNotExist(err) {
        // Create openssl.cnf with basic CRL configuration
        opensslCnf := `
[ ca ]
default_ca = CA_default

[ CA_default ]
database = ` + indexPath + `
serial = ` + filepath.Join(crlDir, "serial.txt") + `
default_md = sha256
default_crl_days = 30
`
        if err := os.WriteFile(opensslCnfPath, []byte(opensslCnf), 0644); err != nil {
            return fmt.Errorf("failed to create openssl.cnf: %w", err)
        }
    }

    // Create or update serial file
    serialPath := filepath.Join(crlDir, "serial.txt")
    if _, err := os.Stat(serialPath); os.IsNotExist(err) {
        // Create serial file with initial value if it doesn't exist
        if err := os.WriteFile(serialPath, []byte("01"), 0644); err != nil {
            return fmt.Errorf("failed to create serial file: %w", err)
        }
    }

    // Add certificate to index file with revocation status
    now := time.Now().UTC().Format("060102150405Z") // YYMMDDhhmmssZ format
    revocationLine := fmt.Sprintf("R\t%s\t%s\t%s\tunknown\t/CN=%s\n", 
        now, serial, now, commonName)

    indexFile, err := os.OpenFile(indexPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        return fmt.Errorf("failed to open index file: %w", err)
    }
    defer indexFile.Close()

    if _, err := indexFile.WriteString(revocationLine); err != nil {
        return fmt.Errorf("failed to write to index file: %w", err)
    }

    // Generate CRL
    crlPath := filepath.Join(c.storage.GetCADirectory(), "ca.crl")
    cmd = exec.Command(
        "openssl", "ca",
        "-config", opensslCnfPath,
        "-gencrl",
        "-keyfile", c.storage.GetCAPrivateKeyPath(),
        "-cert", c.storage.GetCAPublicKeyPath(),
        "-out", crlPath,
    )
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to generate CRL: %w", err)
    }

    // Make the CRL accessible
    crlPublicPath := filepath.Join(c.storage.GetBasePath(), "ca.crl")
    cmd = exec.Command("cp", crlPath, crlPublicPath)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to copy CRL to public location: %w", err)
    }

    // Mark the certificate as revoked in our system
    revokedFlagPath := filepath.Join(c.storage.GetCertificateDirectory(commonName), "revoked")
    if err := os.WriteFile(revokedFlagPath, []byte(now), 0644); err != nil {
        return fmt.Errorf("failed to mark certificate as revoked: %w", err)
    }

    return nil
}