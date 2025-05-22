package storage

// StorageInterface defines the common storage operations
type StorageInterface interface {
	// Path methods
	GetBasePath() string
	GetCADirectory() string
	GetCAPublicKeyPath() string
	GetCAPrivateKeyPath() string
	GetCAEncryptedKeyPath() string
	GetCAPublicCopyPath() string
	GetCertificateDirectory(name string) string
	GetCertificatePath(name string) string
	GetCertificateKeyPath(name string) string
	GetCertificateP12Path(name string) string
	GetCertificatePasswordPath(name string) string
	GetCertificateBundlePath(name string) string

	// CA operations
	SaveCAInfo(caName, caKey, organization, country string) error
	GetCAInfo() (string, string, string, string, error)

	// Certificate operations
	ListCertificates() ([]string, error)
	DeleteCertificate(name string) error
	CreateCertificateDirectory(name string) error
	GetCertificateNameBySerial(serialNumber string) (string, error)
	SaveCertificateSerialMapping(serialNumber, certName string) error

	// Email settings
	SaveEmailSettings(server, port, username, password, from, to string, useTLS, useStartTLS bool) error
	GetEmailSettings() (string, string, string, string, string, string, bool, bool, error)
}

// Ensure both Storage and CachedStorage implement the interface
var _ StorageInterface = (*Storage)(nil)
var _ StorageInterface = (*CachedStorage)(nil)
