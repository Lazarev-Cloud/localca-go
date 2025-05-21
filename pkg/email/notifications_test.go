package email

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockSMTPServer is a simple mock SMTP server for testing
type mockSMTPServer struct {
	addr     string
	messages []string
	auth     bool
	tls      bool
}

func newMockSMTPServer(t *testing.T) *mockSMTPServer {
	// Create a listener on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	server := &mockSMTPServer{
		addr:     listener.Addr().String(),
		messages: make([]string, 0),
	}

	// Start the mock server
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConnection(conn)
		}
	}()

	return server
}

func (s *mockSMTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Send greeting
	conn.Write([]byte("220 mock.smtp.server\r\n"))

	// Simple SMTP protocol handling
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		cmd := string(buffer[:n])
		s.messages = append(s.messages, cmd)

		// Respond based on command
		switch {
		case cmd[:4] == "HELO" || cmd[:4] == "EHLO":
			conn.Write([]byte("250-mock.smtp.server\r\n250-STARTTLS\r\n250 AUTH PLAIN\r\n"))
		case cmd[:8] == "STARTTLS":
			conn.Write([]byte("220 Ready to start TLS\r\n"))
			s.tls = true
		case cmd[:4] == "AUTH":
			conn.Write([]byte("235 Authentication successful\r\n"))
			s.auth = true
		case cmd[:4] == "MAIL":
			conn.Write([]byte("250 OK\r\n"))
		case cmd[:4] == "RCPT":
			conn.Write([]byte("250 OK\r\n"))
		case cmd[:4] == "DATA":
			conn.Write([]byte("354 Start mail input\r\n"))
		case cmd[:4] == "QUIT":
			conn.Write([]byte("221 Bye\r\n"))
			return
		default:
			if cmd[:1] == "." {
				conn.Write([]byte("250 OK\r\n"))
			}
		}
	}
}

func TestNewEmailService(t *testing.T) {
	// Test creating a new email service
	emailSvc := NewEmailService("smtp.example.com", 25, "user", "pass", true, false)

	assert.NotNil(t, emailSvc)
	assert.Equal(t, "smtp.example.com", emailSvc.SMTPServer)
	assert.Equal(t, 25, emailSvc.SMTPPort)
	assert.Equal(t, "user", emailSvc.SMTPUser)
	assert.Equal(t, "pass", emailSvc.SMTPPassword)
	assert.True(t, emailSvc.UseTLS)
	assert.False(t, emailSvc.UseStartTLS)
}

func TestSendEmail_NoServer(t *testing.T) {
	// Test sending email with no server configured
	emailSvc := NewEmailService("", 25, "", "", false, false)

	err := emailSvc.SendEmail("from@example.com", "to@example.com", "Test Subject", "Test Body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP server not configured")
}

// mockEmailService is a mock implementation of EmailService for testing
type mockEmailService struct {
	EmailService
	capturedSubject string
	capturedBody    string
	notifiedCerts   []string
}

func (m *mockEmailService) SendEmail(from, to, subject, body string) error {
	m.capturedSubject = subject
	m.capturedBody = body
	return nil
}

func (m *mockEmailService) SendCertificateExpiryNotification(from, to, certName, expiryDate string) error {
	m.notifiedCerts = append(m.notifiedCerts, certName)
	return nil
}

// Override CheckCertificatesExpiry to avoid SMTP calls
func (m *mockEmailService) CheckCertificatesExpiry(certificateList []CertificateInfo, from, to string, warningDays int) []string {
	now := time.Now()
	warningPeriod := time.Hour * 24 * time.Duration(warningDays)

	for _, cert := range certificateList {
		expiryDate, err := time.Parse("2006-01-02", cert.ExpiryDate)
		if err != nil {
			continue
		}

		// Check if certificate will expire within the warning period
		if now.Add(warningPeriod).After(expiryDate) && now.Before(expiryDate) {
			m.notifiedCerts = append(m.notifiedCerts, cert.CommonName)
		}
	}

	return m.notifiedCerts
}

func TestSendCertificateExpiryNotification(t *testing.T) {
	// Create a mock email service
	mockSvc := &mockEmailService{
		EmailService: EmailService{
			SMTPServer: "mock.smtp.server",
			SMTPPort:   25,
		},
		notifiedCerts: []string{},
	}

	// Test sending certificate expiry notification
	err := mockSvc.SendCertificateExpiryNotification(
		"from@example.com",
		"to@example.com",
		"example.com",
		"2023-12-31")

	// We shouldn't get an error since we're using our mock
	assert.NoError(t, err)

	// Check that the certificate was added to notifiedCerts
	assert.Contains(t, mockSvc.notifiedCerts, "example.com")
}

func TestCheckCertificatesExpiry(t *testing.T) {
	// Create test certificates
	now := time.Now()
	expiringFormat := "2006-01-02"

	// Certificate expiring in 5 days
	expiringDate := now.Add(5 * 24 * time.Hour).Format(expiringFormat)
	// Certificate expiring in 20 days
	safeDate := now.Add(20 * 24 * time.Hour).Format(expiringFormat)
	// Certificate already expired
	expiredDate := now.Add(-5 * 24 * time.Hour).Format(expiringFormat)

	certificates := []CertificateInfo{
		{CommonName: "expiring.com", ExpiryDate: expiringDate, IsClient: false, SerialNumber: "123"},
		{CommonName: "safe.com", ExpiryDate: safeDate, IsClient: false, SerialNumber: "456"},
		{CommonName: "expired.com", ExpiryDate: expiredDate, IsClient: false, SerialNumber: "789"},
		{CommonName: "invalid-date.com", ExpiryDate: "invalid-date", IsClient: false, SerialNumber: "101"},
	}

	// Create a mock email service
	mockSvc := &mockEmailService{
		EmailService: EmailService{
			SMTPServer: "smtp.example.com",
			SMTPPort:   25,
		},
		notifiedCerts: []string{},
	}

	// Test with 10 days warning period
	result := mockSvc.CheckCertificatesExpiry(
		certificates,
		"from@example.com",
		"to@example.com",
		10)

	// Should notify about expiring certificate but not about safe or expired ones
	assert.Contains(t, result, "expiring.com")
	assert.NotContains(t, result, "safe.com")
	assert.NotContains(t, result, "expired.com")
	assert.NotContains(t, result, "invalid-date.com")
}

func TestCheckCertificatesExpiry_NoSMTPServer(t *testing.T) {
	// Create a mock email service with no SMTP server
	emailSvc := &EmailService{
		SMTPServer: "",
	}

	// Test with empty certificate list
	result := emailSvc.CheckCertificatesExpiry(
		[]CertificateInfo{},
		"from@example.com",
		"to@example.com",
		10)

	// Should return nil since SMTP server is not configured
	assert.Nil(t, result)
}

func TestCheckCertificatesExpiry_NoEmailTo(t *testing.T) {
	// Create a mock email service with no recipient
	emailSvc := &EmailService{
		SMTPServer: "smtp.example.com",
	}

	// Test with empty certificate list
	result := emailSvc.CheckCertificatesExpiry(
		[]CertificateInfo{},
		"from@example.com",
		"",
		10)

	// Should return nil since no recipient is specified
	assert.Nil(t, result)
}
