package database

import (
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDatabase_Disabled(t *testing.T) {
	cfg := &config.Config{
		DatabaseEnabled: false,
	}

	db, err := NewDatabase(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "database is not enabled")
}

func TestNewDatabase_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		DatabaseEnabled:  true,
		DatabaseHost:     "invalid-host",
		DatabaseUser:     "test",
		DatabasePassword: "test",
		DatabaseName:     "test",
		DatabasePort:     5432,
		DatabaseSSLMode:  "disable",
		LogLevel:         "info",
	}

	// This should fail to connect to invalid host
	db, err := NewDatabase(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestDatabase_LogAudit(t *testing.T) {
	// Skip if no database connection available
	t.Skip("Skipping database test - requires PostgreSQL connection")
}

func TestDatabase_GetAuditLogs(t *testing.T) {
	// Skip if no database connection available
	t.Skip("Skipping database test - requires PostgreSQL connection")
}

func TestDatabase_Migrate(t *testing.T) {
	// Skip if no database connection available
	t.Skip("Skipping database test - requires PostgreSQL connection")
}

func TestDatabase_Health(t *testing.T) {
	// Skip if no database connection available
	t.Skip("Skipping database test - requires PostgreSQL connection")
}

func TestCAInfo_TableName(t *testing.T) {
	ca := CAInfo{}
	assert.Equal(t, "ca_info", ca.TableName())
}

func TestCertificate_TableName(t *testing.T) {
	cert := Certificate{}
	assert.Equal(t, "certificates", cert.TableName())
}

func TestEmailSettings_TableName(t *testing.T) {
	email := EmailSettings{}
	assert.Equal(t, "email_settings", email.TableName())
}

func TestAuditLog_TableName(t *testing.T) {
	audit := AuditLog{}
	assert.Equal(t, "audit_logs", audit.TableName())
}

func TestSerialMapping_TableName(t *testing.T) {
	serial := SerialMapping{}
	assert.Equal(t, "serial_mappings", serial.TableName())
}

func TestCAInfo_Structure(t *testing.T) {
	ca := CAInfo{
		Name:         "Test CA",
		Organization: "Test Org",
		Country:      "US",
		KeyHash:      "abcdef123456",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	assert.Equal(t, "Test CA", ca.Name)
	assert.Equal(t, "Test Org", ca.Organization)
	assert.Equal(t, "US", ca.Country)
	assert.Equal(t, "abcdef123456", ca.KeyHash)
	assert.NotZero(t, ca.CreatedAt)
	assert.NotZero(t, ca.UpdatedAt)
}

func TestCertificate_Structure(t *testing.T) {
	cert := Certificate{
		Name:         "test.example.com",
		SerialNumber: "123456789ABCDEF",
		Subject:      "CN=test.example.com",
		Issuer:       "CN=Test CA",
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		IsRevoked:    false,
		S3Path:       "/certificates/test.example.com",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	assert.Equal(t, "test.example.com", cert.Name)
	assert.Equal(t, "123456789ABCDEF", cert.SerialNumber)
	assert.Equal(t, "CN=test.example.com", cert.Subject)
	assert.Equal(t, "CN=Test CA", cert.Issuer)
	assert.False(t, cert.IsRevoked)
	assert.Nil(t, cert.RevokedAt)
	assert.Equal(t, "/certificates/test.example.com", cert.S3Path)
	assert.NotZero(t, cert.CreatedAt)
	assert.NotZero(t, cert.UpdatedAt)
}

func TestEmailSettings_Structure(t *testing.T) {
	email := EmailSettings{
		SMTPServer:  "smtp.example.com",
		SMTPPort:    "587",
		Username:    "user@example.com",
		Password:    "encrypted_password",
		FromEmail:   "ca@example.com",
		ToEmail:     "admin@example.com",
		UseTLS:      true,
		UseStartTLS: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, "smtp.example.com", email.SMTPServer)
	assert.Equal(t, "587", email.SMTPPort)
	assert.Equal(t, "user@example.com", email.Username)
	assert.Equal(t, "encrypted_password", email.Password)
	assert.Equal(t, "ca@example.com", email.FromEmail)
	assert.Equal(t, "admin@example.com", email.ToEmail)
	assert.True(t, email.UseTLS)
	assert.False(t, email.UseStartTLS)
	assert.NotZero(t, email.CreatedAt)
	assert.NotZero(t, email.UpdatedAt)
}

func TestAuditLog_Structure(t *testing.T) {
	audit := AuditLog{
		Action:     "create",
		Resource:   "certificate",
		ResourceID: "test.example.com",
		UserIP:     "192.168.1.1",
		UserAgent:  "test-agent",
		Details:    "Certificate created successfully",
		Success:    true,
		Error:      "",
		CreatedAt:  time.Now(),
	}

	assert.Equal(t, "create", audit.Action)
	assert.Equal(t, "certificate", audit.Resource)
	assert.Equal(t, "test.example.com", audit.ResourceID)
	assert.Equal(t, "192.168.1.1", audit.UserIP)
	assert.Equal(t, "test-agent", audit.UserAgent)
	assert.Equal(t, "Certificate created successfully", audit.Details)
	assert.True(t, audit.Success)
	assert.Empty(t, audit.Error)
	assert.NotZero(t, audit.CreatedAt)
}

func TestSerialMapping_Structure(t *testing.T) {
	serial := SerialMapping{
		SerialNumber: "123456789ABCDEF",
		CertName:     "test.example.com",
		CreatedAt:    time.Now(),
	}

	assert.Equal(t, "123456789ABCDEF", serial.SerialNumber)
	assert.Equal(t, "test.example.com", serial.CertName)
	assert.NotZero(t, serial.CreatedAt)
}

func TestCertificate_RevokedState(t *testing.T) {
	// Test non-revoked certificate
	cert := Certificate{
		IsRevoked: false,
		RevokedAt: nil,
	}
	assert.False(t, cert.IsRevoked)
	assert.Nil(t, cert.RevokedAt)

	// Test revoked certificate
	revokedTime := time.Now()
	revokedCert := Certificate{
		IsRevoked: true,
		RevokedAt: &revokedTime,
	}
	assert.True(t, revokedCert.IsRevoked)
	assert.NotNil(t, revokedCert.RevokedAt)
	assert.Equal(t, revokedTime, *revokedCert.RevokedAt)
}

func TestEmailSettings_TLSOptions(t *testing.T) {
	// Test TLS enabled
	tlsEmail := EmailSettings{
		UseTLS:      true,
		UseStartTLS: false,
	}
	assert.True(t, tlsEmail.UseTLS)
	assert.False(t, tlsEmail.UseStartTLS)

	// Test StartTLS enabled
	startTLSEmail := EmailSettings{
		UseTLS:      false,
		UseStartTLS: true,
	}
	assert.False(t, startTLSEmail.UseTLS)
	assert.True(t, startTLSEmail.UseStartTLS)

	// Test both disabled
	plainEmail := EmailSettings{
		UseTLS:      false,
		UseStartTLS: false,
	}
	assert.False(t, plainEmail.UseTLS)
	assert.False(t, plainEmail.UseStartTLS)
}

func TestAuditLog_SuccessAndError(t *testing.T) {
	// Test successful audit log
	successAudit := AuditLog{
		Success: true,
		Error:   "",
	}
	assert.True(t, successAudit.Success)
	assert.Empty(t, successAudit.Error)

	// Test failed audit log
	failedAudit := AuditLog{
		Success: false,
		Error:   "Operation failed",
	}
	assert.False(t, failedAudit.Success)
	assert.Equal(t, "Operation failed", failedAudit.Error)
}
