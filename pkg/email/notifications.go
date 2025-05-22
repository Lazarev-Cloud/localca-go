package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// SanitizeInput sanitizes input to prevent injection attacks
func SanitizeInput(input string) string {
	// Replace CRLF characters to prevent header injection
	return strings.ReplaceAll(strings.ReplaceAll(input, "\r", ""), "\n", "")
}

// EmailService handles email notifications
type EmailService struct {
	SMTPServer   string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	UseTLS       bool
	UseStartTLS  bool
}

// CertificateInfo holds information about a certificate for email notifications
type CertificateInfo struct {
	CommonName   string
	ExpiryDate   string
	IsClient     bool
	SerialNumber string
}

// NewEmailService creates a new email service
func NewEmailService(smtpServer string, smtpPort int, smtpUser, smtpPassword string, useTLS, useStartTLS bool) *EmailService {
	return &EmailService{
		SMTPServer:   smtpServer,
		SMTPPort:     smtpPort,
		SMTPUser:     smtpUser,
		SMTPPassword: smtpPassword,
		UseTLS:       useTLS,
		UseStartTLS:  useStartTLS,
	}
}

// SendEmail sends an email message
func (e *EmailService) SendEmail(from, to, subject, body string) error {
	if e.SMTPServer == "" {
		return fmt.Errorf("SMTP server not configured")
	}

	// Format the message
	// Sanitize all inputs to prevent injection attacks
	safeFrom := SanitizeInput(from)
	safeTo := SanitizeInput(to)
	safeSubject := SanitizeInput(subject)
	safeBody := SanitizeInput(body)

	message := []byte(
		fmt.Sprintf("To: %s\r\n"+
			"From: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s", safeTo, safeFrom, safeSubject, safeBody))

	// Set authentication
	auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPassword, e.SMTPServer)

	// Set connection address
	addr := fmt.Sprintf("%s:%d", e.SMTPServer, e.SMTPPort)

	// Send email based on TLS settings
	if e.UseTLS {
		// Configure TLS connection
		tlsConfig := &tls.Config{
			ServerName: e.SMTPServer,
		}

		// Connect to the server with TLS
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		// Create client
		client, err := smtp.NewClient(conn, e.SMTPServer)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		// Set up authentication
		if e.SMTPUser != "" {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}

		// Set the sender and recipient
		if err := client.Mail(from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		// Send the message body
		wc, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to start data transaction: %w", err)
		}
		defer wc.Close()

		_, err = wc.Write(message)
		if err != nil {
			return fmt.Errorf("failed to send message body: %w", err)
		}

		return nil
	} else if e.UseStartTLS {
		// Connect to the server without TLS first
		client, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer client.Close()

		// Start STARTTLS
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: e.SMTPServer}
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %w", err)
			}
		} else {
			return fmt.Errorf("STARTTLS not supported by server")
		}

		// Set up authentication
		if e.SMTPUser != "" {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}

		// Set the sender and recipient
		if err := client.Mail(from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		// Send the message body
		wc, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to start data transaction: %w", err)
		}
		defer wc.Close()

		_, err = wc.Write(message)
		if err != nil {
			return fmt.Errorf("failed to send message body: %w", err)
		}

		return nil
	} else {
		// Simple SMTP without encryption
		return smtp.SendMail(addr, auth, from, []string{to}, message)
	}
}

// SendCertificateExpiryNotification sends a notification for certificates that will expire soon
func (e *EmailService) SendCertificateExpiryNotification(from, to, certName, expiryDate string) error {
	subject := fmt.Sprintf("Certificate Expiry Warning: %s", certName)
	body := fmt.Sprintf(
		"Certificate Expiry Warning\n\n"+
			"The following certificate will expire soon:\n\n"+
			"Certificate Name: %s\n"+
			"Expiry Date: %s\n\n"+
			"Please renew this certificate before it expires to avoid service disruption.\n\n"+
			"This is an automated message from LocalCA.\n"+
			"Time: %s",
		certName, expiryDate, time.Now().Format("2006-01-02 15:04:05"))

	return e.SendEmail(from, to, subject, body)
}

// CheckCertificatesExpiry checks for certificates that will expire soon and sends notifications
func (e *EmailService) CheckCertificatesExpiry(certificateList []CertificateInfo, from, to string, warningDays int) []string {
	if e.SMTPServer == "" || to == "" {
		return nil
	}

	var notifiedCerts []string
	now := time.Now()
	warningPeriod := time.Hour * 24 * time.Duration(warningDays)

	for _, cert := range certificateList {
		expiryDate, err := time.Parse("2006-01-02", cert.ExpiryDate)
		if err != nil {
			continue
		}

		// Check if certificate will expire within the warning period
		if now.Add(warningPeriod).After(expiryDate) && now.Before(expiryDate) {
			// Send notification
			if err := e.SendCertificateExpiryNotification(from, to, cert.CommonName, cert.ExpiryDate); err == nil {
				notifiedCerts = append(notifiedCerts, cert.CommonName)
			}
		}
	}

	return notifiedCerts
}
