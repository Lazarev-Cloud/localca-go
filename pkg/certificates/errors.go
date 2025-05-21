package certificates

import "errors"

// Error definitions for certificate operations
var (
	// ErrCertificateNotFound is returned when a certificate is not found
	ErrCertificateNotFound = errors.New("certificate not found")

	// ErrInvalidCertificate is returned when a certificate is invalid
	ErrInvalidCertificate = errors.New("invalid certificate")

	// ErrCANotFound is returned when the CA certificate is not found
	ErrCANotFound = errors.New("CA certificate not found")

	// ErrCAKeyNotFound is returned when the CA key is not found
	ErrCAKeyNotFound = errors.New("CA key not found")

	// ErrCertificateAlreadyExists is returned when a certificate already exists
	ErrCertificateAlreadyExists = errors.New("certificate already exists")

	// ErrCertificateRevoked is returned when a certificate is already revoked
	ErrCertificateRevoked = errors.New("certificate is already revoked")
)
