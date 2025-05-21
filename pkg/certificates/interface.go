package certificates

// CertificateServiceInterface defines the interface for certificate operations
type CertificateServiceInterface interface {
	// CA operations
	CAExists() (bool, error)
	CreateCA() error
	RenewCA() error
	CreateServiceCertificate() error

	// Certificate operations
	CreateServerCertificate(commonName string, domains []string) error
	CreateClientCertificate(commonName, password string) error
	RevokeCertificate(name string) error
	RenewServerCertificate(name string) error
	RenewClientCertificate(name string) error
	GetAllCertificates() ([]Certificate, error)
	GetCertificateInfo(name string) (*Certificate, error)
}

// Ensure CertificateService implements CertificateServiceInterface
var _ CertificateServiceInterface = (*CertificateService)(nil)
