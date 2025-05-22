package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/database"
	"github.com/Lazarev-Cloud/localca-go/pkg/logging"
	"github.com/Lazarev-Cloud/localca-go/pkg/s3storage"
)

// EnhancedStorage combines file storage, database, and S3 storage
type EnhancedStorage struct {
	fileStorage *Storage
	database    *database.Database
	s3Client    *s3storage.S3Client
	logger      *logging.Logger
	config      *config.Config
}

// NewEnhancedStorage creates a new enhanced storage instance
func NewEnhancedStorage(cfg *config.Config, logger *logging.Logger) (*EnhancedStorage, error) {
	// Initialize file storage (always needed as fallback)
	fileStorage, err := NewStorage(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize file storage: %w", err)
	}

	enhanced := &EnhancedStorage{
		fileStorage: fileStorage,
		logger:      logger,
		config:      cfg,
	}

	// Initialize database if enabled
	if cfg.DatabaseEnabled {
		db, err := database.NewDatabase(cfg)
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize database, falling back to file storage only")
		} else {
			// Run migrations
			if err := db.Migrate(); err != nil {
				logger.WithError(err).Warn("Failed to run database migrations")
			} else {
				enhanced.database = db
				logger.Info("Database storage initialized successfully")
			}
		}
	}

	// Initialize S3 client if enabled
	if cfg.S3Enabled {
		s3Client, err := s3storage.NewS3Client(cfg)
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize S3 storage, falling back to file storage only")
		} else {
			enhanced.s3Client = s3Client
			logger.Info("S3 storage initialized successfully")
		}
	}

	return enhanced, nil
}

// Close closes all storage connections
func (e *EnhancedStorage) Close() error {
	var errors []error

	if e.database != nil {
		if err := e.database.Close(); err != nil {
			errors = append(errors, fmt.Errorf("database close error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("storage close errors: %v", errors)
	}

	return nil
}

// Health checks the health of all storage backends
func (e *EnhancedStorage) Health() map[string]error {
	health := make(map[string]error)

	// Check file storage (always available)
	health["file"] = nil

	// Check database
	if e.database != nil {
		health["database"] = e.database.Health()
	} else {
		health["database"] = fmt.Errorf("database not enabled")
	}

	// Check S3
	if e.s3Client != nil && e.s3Client.IsEnabled() {
		// Try to list objects to test connectivity
		_, err := e.s3Client.ListFiles("")
		health["s3"] = err
	} else {
		health["s3"] = fmt.Errorf("S3 not enabled")
	}

	return health
}

// Implement StorageInterface methods by delegating to file storage
func (e *EnhancedStorage) GetBasePath() string {
	return e.fileStorage.GetBasePath()
}

func (e *EnhancedStorage) GetCADirectory() string {
	return e.fileStorage.GetCADirectory()
}

func (e *EnhancedStorage) GetCAPublicKeyPath() string {
	return e.fileStorage.GetCAPublicKeyPath()
}

func (e *EnhancedStorage) GetCAPrivateKeyPath() string {
	return e.fileStorage.GetCAPrivateKeyPath()
}

func (e *EnhancedStorage) GetCAEncryptedKeyPath() string {
	return e.fileStorage.GetCAEncryptedKeyPath()
}

func (e *EnhancedStorage) GetCAPublicCopyPath() string {
	return e.fileStorage.GetCAPublicCopyPath()
}

func (e *EnhancedStorage) GetCertificateDirectory(name string) string {
	return e.fileStorage.GetCertificateDirectory(name)
}

func (e *EnhancedStorage) GetCertificatePath(name string) string {
	return e.fileStorage.GetCertificatePath(name)
}

func (e *EnhancedStorage) GetCertificateKeyPath(name string) string {
	return e.fileStorage.GetCertificateKeyPath(name)
}

func (e *EnhancedStorage) GetCertificateP12Path(name string) string {
	return e.fileStorage.GetCertificateP12Path(name)
}

func (e *EnhancedStorage) GetCertificatePasswordPath(name string) string {
	return e.fileStorage.GetCertificatePasswordPath(name)
}

func (e *EnhancedStorage) GetCertificateBundlePath(name string) string {
	return e.fileStorage.GetCertificateBundlePath(name)
}

// Enhanced CA operations with database and S3 support
func (e *EnhancedStorage) SaveCAInfo(caName, caKey, organization, country string) error {
	// Save to file storage first (primary)
	if err := e.fileStorage.SaveCAInfo(caName, caKey, organization, country); err != nil {
		return err
	}

	// Save to database if available
	if e.database != nil {
		keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(caKey)))
		caInfo := database.CAInfo{
			Name:         caName,
			Organization: organization,
			Country:      country,
			KeyHash:      keyHash,
		}

		if err := e.database.DB.Create(&caInfo).Error; err != nil {
			e.logger.WithError(err).Warn("Failed to save CA info to database")
		}
	}

	// Upload CA files to S3 if available
	if e.s3Client != nil && e.s3Client.IsEnabled() {
		e.uploadCAFilesToS3(caName)
	}

	return nil
}

func (e *EnhancedStorage) GetCAInfo() (string, string, string, string, error) {
	// Try database first if available
	if e.database != nil {
		var caInfo database.CAInfo
		if err := e.database.DB.First(&caInfo).Error; err == nil {
			// Get key from file storage
			_, key, _, _, err := e.fileStorage.GetCAInfo()
			if err != nil {
				return "", "", "", "", err
			}
			return caInfo.Name, key, caInfo.Organization, caInfo.Country, nil
		}
	}

	// Fallback to file storage
	return e.fileStorage.GetCAInfo()
}

// Enhanced certificate operations
func (e *EnhancedStorage) ListCertificates() ([]string, error) {
	// Try database first if available
	if e.database != nil {
		var certs []database.Certificate
		if err := e.database.DB.Find(&certs).Error; err == nil {
			names := make([]string, len(certs))
			for i, cert := range certs {
				names[i] = cert.Name
			}
			return names, nil
		}
	}

	// Fallback to file storage
	return e.fileStorage.ListCertificates()
}

func (e *EnhancedStorage) DeleteCertificate(name string) error {
	// Delete from file storage
	if err := e.fileStorage.DeleteCertificate(name); err != nil {
		return err
	}

	// Delete from database if available
	if e.database != nil {
		if err := e.database.DB.Where("name = ?", name).Delete(&database.Certificate{}).Error; err != nil {
			e.logger.WithError(err).Warn("Failed to delete certificate from database")
		}
	}

	// Delete from S3 if available
	if e.s3Client != nil && e.s3Client.IsEnabled() {
		if err := e.s3Client.DeleteCertificateFiles(name); err != nil {
			e.logger.WithError(err).Warn("Failed to delete certificate files from S3")
		}
	}

	return nil
}

func (e *EnhancedStorage) CreateCertificateDirectory(name string) error {
	return e.fileStorage.CreateCertificateDirectory(name)
}

func (e *EnhancedStorage) GetCertificateNameBySerial(serialNumber string) (string, error) {
	// Try database first if available
	if e.database != nil {
		var mapping database.SerialMapping
		if err := e.database.DB.Where("serial_number = ?", serialNumber).First(&mapping).Error; err == nil {
			return mapping.CertName, nil
		}
	}

	// Fallback to file storage
	return e.fileStorage.GetCertificateNameBySerial(serialNumber)
}

func (e *EnhancedStorage) SaveCertificateSerialMapping(serialNumber, certName string) error {
	// Save to file storage
	if err := e.fileStorage.SaveCertificateSerialMapping(serialNumber, certName); err != nil {
		return err
	}

	// Save to database if available
	if e.database != nil {
		mapping := database.SerialMapping{
			SerialNumber: serialNumber,
			CertName:     certName,
		}

		if err := e.database.DB.Create(&mapping).Error; err != nil {
			e.logger.WithError(err).Warn("Failed to save serial mapping to database")
		}
	}

	return nil
}

// Enhanced email settings operations
func (e *EnhancedStorage) SaveEmailSettings(server, port, username, password, from, to string, useTLS, useStartTLS bool) error {
	// Save to file storage
	if err := e.fileStorage.SaveEmailSettings(server, port, username, password, from, to, useTLS, useStartTLS); err != nil {
		return err
	}

	// Save to database if available
	if e.database != nil {
		// Delete existing settings
		e.database.DB.Where("1 = 1").Delete(&database.EmailSettings{})

		settings := database.EmailSettings{
			SMTPServer:  server,
			SMTPPort:    port,
			Username:    username,
			Password:    password, // TODO: Encrypt this
			FromEmail:   from,
			ToEmail:     to,
			UseTLS:      useTLS,
			UseStartTLS: useStartTLS,
		}

		if err := e.database.DB.Create(&settings).Error; err != nil {
			e.logger.WithError(err).Warn("Failed to save email settings to database")
		}
	}

	return nil
}

func (e *EnhancedStorage) GetEmailSettings() (string, string, string, string, string, string, bool, bool, error) {
	// Try database first if available
	if e.database != nil {
		var settings database.EmailSettings
		if err := e.database.DB.First(&settings).Error; err == nil {
			return settings.SMTPServer, settings.SMTPPort, settings.Username, settings.Password,
				settings.FromEmail, settings.ToEmail, settings.UseTLS, settings.UseStartTLS, nil
		}
	}

	// Fallback to file storage
	return e.fileStorage.GetEmailSettings()
}

// Additional methods for enhanced functionality

// SaveCertificateToDatabase saves certificate metadata to database
func (e *EnhancedStorage) SaveCertificateToDatabase(name, serialNumber, subject, issuer string, notBefore, notAfter time.Time) error {
	if e.database == nil {
		return nil // Database not available
	}

	cert := database.Certificate{
		Name:         name,
		SerialNumber: serialNumber,
		Subject:      subject,
		Issuer:       issuer,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		IsRevoked:    false,
	}

	return e.database.DB.Create(&cert).Error
}

// UploadCertificateToS3 uploads certificate files to S3
func (e *EnhancedStorage) UploadCertificateToS3(certName string) error {
	if e.s3Client == nil || !e.s3Client.IsEnabled() {
		return nil // S3 not available
	}

	// Collect all certificate files
	files := make(map[string][]byte)

	// Certificate file
	if data, err := os.ReadFile(e.GetCertificatePath(certName)); err == nil {
		files["cert.pem"] = data
	}

	// Private key file
	if data, err := os.ReadFile(e.GetCertificateKeyPath(certName)); err == nil {
		files["key.pem"] = data
	}

	// P12 file
	if data, err := os.ReadFile(e.GetCertificateP12Path(certName)); err == nil {
		files["cert.p12"] = data
	}

	// Password file
	if data, err := os.ReadFile(e.GetCertificatePasswordPath(certName)); err == nil {
		files["password.txt"] = data
	}

	// Bundle file
	if data, err := os.ReadFile(e.GetCertificateBundlePath(certName)); err == nil {
		files["bundle.pem"] = data
	}

	return e.s3Client.UploadCertificateFiles(certName, files)
}

// LogAudit logs an audit event to database and logger
func (e *EnhancedStorage) LogAudit(action, resource, resourceID, userIP, userAgent, details string, success bool, errorMsg string) {
	// Log to structured logger
	if success {
		e.logger.AuditInfo(action, resource, resourceID, userIP, userAgent, nil)
	} else {
		e.logger.AuditError(action, resource, resourceID, userIP, userAgent, errorMsg, nil)
	}

	// Log to database if available
	if e.database != nil {
		if err := e.database.LogAudit(action, resource, resourceID, userIP, userAgent, details, success, errorMsg); err != nil {
			e.logger.WithError(err).Warn("Failed to save audit log to database")
		}
	}
}

// uploadCAFilesToS3 uploads CA files to S3
func (e *EnhancedStorage) uploadCAFilesToS3(caName string) {
	if e.s3Client == nil || !e.s3Client.IsEnabled() {
		return
	}

	files := make(map[string][]byte)

	// CA certificate
	if data, err := os.ReadFile(e.GetCAPublicKeyPath()); err == nil {
		files["ca.pem"] = data
	}

	// CA private key (encrypted)
	if data, err := os.ReadFile(e.GetCAEncryptedKeyPath()); err == nil {
		files["ca-encrypted.key"] = data
	}

	// Upload files
	for filename, data := range files {
		objectName := fmt.Sprintf("ca/%s", filename)
		contentType := getContentTypeForFile(filename)
		if err := e.s3Client.UploadFile(objectName, data, contentType); err != nil {
			e.logger.WithError(err).Warn("Failed to upload CA file to S3")
		}
	}
}

// getContentTypeForFile returns the appropriate content type for a file
func getContentTypeForFile(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".crt", ".pem":
		return "application/x-pem-file"
	case ".key":
		return "application/x-pem-file"
	case ".p12", ".pfx":
		return "application/x-pkcs12"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// Ensure EnhancedStorage implements StorageInterface
var _ StorageInterface = (*EnhancedStorage)(nil)
