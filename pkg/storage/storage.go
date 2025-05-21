package storage

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// Storage handles certificate storage
type Storage struct {
	basePath string
}

// NewStorage creates a new storage handler
func NewStorage(basePath string) (*Storage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Storage{
		basePath: basePath,
	}, nil
}

// GetBasePath returns the base path for storage
func (s *Storage) GetBasePath() string {
	return s.basePath
}

// GetCADirectory returns the path to the CA directory
func (s *Storage) GetCADirectory() string {
	return filepath.Join(s.basePath, "ca")
}

// GetCAPublicKeyPath returns the path to the CA certificate
func (s *Storage) GetCAPublicKeyPath() string {
	return filepath.Join(s.GetCADirectory(), "ca.pem")
}

// GetCAPrivateKeyPath returns the path to the CA private key
func (s *Storage) GetCAPrivateKeyPath() string {
	return filepath.Join(s.GetCADirectory(), "ca.key")
}

// GetCAEncryptedKeyPath returns the path to the encrypted CA private key
func (s *Storage) GetCAEncryptedKeyPath() string {
	return filepath.Join(s.GetCADirectory(), "ca.key.enc")
}

// GetCAPublicCopyPath returns the path to the public copy of the CA certificate
func (s *Storage) GetCAPublicCopyPath() string {
	return filepath.Join(s.basePath, "ca.pem")
}

// GetCertificateDirectory returns the path to a certificate directory
func (s *Storage) GetCertificateDirectory(name string) string {
	return filepath.Join(s.basePath, name)
}

// GetCertificatePath returns the path to a certificate
func (s *Storage) GetCertificatePath(name string) string {
	return filepath.Join(s.GetCertificateDirectory(name), name+".crt")
}

// GetCertificateKeyPath returns the path to a certificate's private key
func (s *Storage) GetCertificateKeyPath(name string) string {
	return filepath.Join(s.GetCertificateDirectory(name), name+".key")
}

// GetCertificateP12Path returns the path to a certificate's P12 file
func (s *Storage) GetCertificateP12Path(name string) string {
	return filepath.Join(s.GetCertificateDirectory(name), name+".p12")
}

// GetCertificatePasswordPath returns the path to a certificate's password file
func (s *Storage) GetCertificatePasswordPath(name string) string {
	return filepath.Join(s.GetCertificateDirectory(name), name+".pw")
}

// GetCertificateBundlePath returns the path to a certificate bundle
func (s *Storage) GetCertificateBundlePath(name string) string {
	return filepath.Join(s.GetCertificateDirectory(name), name+".bundle.crt")
}

// SaveCAInfo saves the CA information to files
func (s *Storage) SaveCAInfo(caName, caKey, organization, country string) error {
	caDir := s.GetCADirectory()
	if err := os.MkdirAll(caDir, 0755); err != nil {
		return fmt.Errorf("failed to create CA directory: %w", err)
	}

	// Save CA name
	if err := os.WriteFile(filepath.Join(caDir, "CA_NAME.txt"), []byte(caName), 0644); err != nil {
		return fmt.Errorf("failed to save CA name: %w", err)
	}

	// Save CA key password
	if err := os.WriteFile(filepath.Join(caDir, "CA_KEY.txt"), []byte(caKey), 0600); err != nil {
		return fmt.Errorf("failed to save CA key: %w", err)
	}

	// Save organization
	if err := os.WriteFile(filepath.Join(caDir, "O.txt"), []byte(organization), 0644); err != nil {
		return fmt.Errorf("failed to save organization: %w", err)
	}

	// Save country
	if err := os.WriteFile(filepath.Join(caDir, "C.txt"), []byte(country), 0644); err != nil {
		return fmt.Errorf("failed to save country: %w", err)
	}

	return nil
}

// GetCAInfo retrieves the CA information
func (s *Storage) GetCAInfo() (string, string, string, string, error) {
	caDir := s.GetCADirectory()

	// Read CA name
	caNameBytes, err := os.ReadFile(filepath.Join(caDir, "CA_NAME.txt"))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to read CA name: %w", err)
	}
	caName := string(caNameBytes)

	// Read CA key password
	caKeyBytes, err := os.ReadFile(filepath.Join(caDir, "CA_KEY.txt"))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to read CA key: %w", err)
	}
	caKey := string(caKeyBytes)

	// Read organization
	orgBytes, err := os.ReadFile(filepath.Join(caDir, "O.txt"))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to read organization: %w", err)
	}
	organization := string(orgBytes)

	// Read country
	countryBytes, err := os.ReadFile(filepath.Join(caDir, "C.txt"))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to read country: %w", err)
	}
	country := string(countryBytes)

	return caName, caKey, organization, country, nil
}

// ListCertificates returns a list of all certificates
func (s *Storage) ListCertificates() ([]string, error) {
	// List directories in base path, excluding "ca"
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var certificates []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "ca" && entry.Name() != "service" && entry.Name() != "settings" {
			// Check if this directory contains a certificate
			certPath := filepath.Join(s.basePath, entry.Name(), entry.Name()+".crt")
			if _, err := os.Stat(certPath); err == nil {
				certificates = append(certificates, entry.Name())
			}
		}
	}

	return certificates, nil
}

// DeleteCertificate deletes a certificate and its associated files
func (s *Storage) DeleteCertificate(name string) error {
	certDir := s.GetCertificateDirectory(name)
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		return fmt.Errorf("certificate does not exist: %s", name)
	}

	// Remove the entire certificate directory
	if err := os.RemoveAll(certDir); err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	return nil
}

// SaveEmailSettings saves the email notification settings
func (s *Storage) SaveEmailSettings(server, port, username, password, from, to string, useTLS, useStartTLS bool) error {
	settingsDir := filepath.Join(s.basePath, "settings")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Save SMTP server
	if err := os.WriteFile(filepath.Join(settingsDir, "smtp_server.txt"), []byte(server), 0644); err != nil {
		return fmt.Errorf("failed to save SMTP server: %w", err)
	}

	// Save SMTP port
	if err := os.WriteFile(filepath.Join(settingsDir, "smtp_port.txt"), []byte(port), 0644); err != nil {
		return fmt.Errorf("failed to save SMTP port: %w", err)
	}

	// Save SMTP username
	if err := os.WriteFile(filepath.Join(settingsDir, "smtp_user.txt"), []byte(username), 0644); err != nil {
		return fmt.Errorf("failed to save SMTP username: %w", err)
	}

	// Save SMTP password
	if err := os.WriteFile(filepath.Join(settingsDir, "smtp_pass.txt"), []byte(password), 0600); err != nil {
		return fmt.Errorf("failed to save SMTP password: %w", err)
	}

	// Save email from
	if err := os.WriteFile(filepath.Join(settingsDir, "email_from.txt"), []byte(from), 0644); err != nil {
		return fmt.Errorf("failed to save email from: %w", err)
	}

	// Save email to
	if err := os.WriteFile(filepath.Join(settingsDir, "email_to.txt"), []byte(to), 0644); err != nil {
		return fmt.Errorf("failed to save email to: %w", err)
	}

	// Save TLS setting
	tlsValue := "0"
	if useTLS {
		tlsValue = "1"
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "use_tls.txt"), []byte(tlsValue), 0644); err != nil {
		return fmt.Errorf("failed to save TLS setting: %w", err)
	}

	// Save StartTLS setting
	startTLSValue := "0"
	if useStartTLS {
		startTLSValue = "1"
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "use_starttls.txt"), []byte(startTLSValue), 0644); err != nil {
		return fmt.Errorf("failed to save StartTLS setting: %w", err)
	}

	return nil
}

// GetEmailSettings retrieves the email notification settings
func (s *Storage) GetEmailSettings() (string, string, string, string, string, string, bool, bool, error) {
	settingsDir := filepath.Join(s.basePath, "settings")
	if _, err := os.Stat(settingsDir); os.IsNotExist(err) {
		return "", "", "", "", "", "", false, false, nil
	}

	// Default values
	server := ""
	port := "25"
	username := ""
	password := ""
	from := ""
	to := ""
	useTLS := false
	useStartTLS := false

	// Read SMTP server
	serverBytes, err := os.ReadFile(filepath.Join(settingsDir, "smtp_server.txt"))
	if err == nil {
		server = string(serverBytes)
	}

	// Read SMTP port
	portBytes, err := os.ReadFile(filepath.Join(settingsDir, "smtp_port.txt"))
	if err == nil {
		port = string(portBytes)
	}

	// Read SMTP username
	usernameBytes, err := os.ReadFile(filepath.Join(settingsDir, "smtp_user.txt"))
	if err == nil {
		username = string(usernameBytes)
	}

	// Read SMTP password
	passwordBytes, err := os.ReadFile(filepath.Join(settingsDir, "smtp_pass.txt"))
	if err == nil {
		password = string(passwordBytes)
	}

	// Read email from
	fromBytes, err := os.ReadFile(filepath.Join(settingsDir, "email_from.txt"))
	if err == nil {
		from = string(fromBytes)
	}

	// Read email to
	toBytes, err := os.ReadFile(filepath.Join(settingsDir, "email_to.txt"))
	if err == nil {
		to = string(toBytes)
	}

	// Read TLS setting
	tlsBytes, err := os.ReadFile(filepath.Join(settingsDir, "use_tls.txt"))
	if err == nil {
		useTLS = string(tlsBytes) == "1"
	}

	// Read StartTLS setting
	startTLSBytes, err := os.ReadFile(filepath.Join(settingsDir, "use_starttls.txt"))
	if err == nil {
		useStartTLS = string(startTLSBytes) == "1"
	}

	return server, port, username, password, from, to, useTLS, useStartTLS, nil
}

// GetCertificateNameBySerial finds a certificate by its serial number
func (s *Storage) GetCertificateNameBySerial(serialNumber string) (string, error) {
	// List all certificates
	certNames, err := s.ListCertificates()
	if err != nil {
		return "", fmt.Errorf("failed to list certificates: %w", err)
	}

	// Check each certificate for the matching serial number
	for _, name := range certNames {
		// Sanitize name to prevent path traversal
		safeName := filepath.Base(name)
		certPath := s.GetCertificatePath(safeName)

		// Read and parse the certificate
		certData, err := os.ReadFile(certPath)
		if err != nil {
			continue // Skip if can't read
		}

		block, _ := pem.Decode(certData)
		if block == nil {
			continue // Skip if not PEM format
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue // Skip if can't parse
		}

		// Check if serial number matches
		certSerial := fmt.Sprintf("%X", cert.SerialNumber)
		if certSerial == serialNumber {
			return safeName, nil
		}
	}

	return "", fmt.Errorf("certificate with serial number %s not found", serialNumber)
}

// CreateCertificateDirectory creates a directory for a certificate
func (s *Storage) CreateCertificateDirectory(name string) error {
	certDir := filepath.Join(s.basePath, name)
	return os.MkdirAll(certDir, 0755)
}

// SaveCertificateSerialMapping saves a mapping from serial number to certificate name
func (s *Storage) SaveCertificateSerialMapping(serialNumber, certName string) error {
	serialsDir := filepath.Join(s.basePath, "serials")
	if err := os.MkdirAll(serialsDir, 0755); err != nil {
		return err
	}

	serialFile := filepath.Join(serialsDir, serialNumber)
	return os.WriteFile(serialFile, []byte(certName), 0644)
}
