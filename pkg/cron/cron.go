// pkg/cron/cron.go

package cron

import (
	"fmt"
	"log"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/email"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// CronService handles scheduled tasks
type CronService struct {
	certSvc      *certificates.CertificateService
	store        *storage.Storage
	runningTasks map[string]bool
}

// NewCronService creates a new cron service
func NewCronService(certSvc *certificates.CertificateService, store *storage.Storage) *CronService {
	return &CronService{
		certSvc:      certSvc,
		store:        store,
		runningTasks: make(map[string]bool),
	}
}

// StartExpiryChecker starts a task to check for expiring certificates
func (c *CronService) StartExpiryChecker() {
	// Only start if not already running
	if c.runningTasks["expiry_checker"] {
		return
	}
	c.runningTasks["expiry_checker"] = true

	go func() {
		// Initial delay to ensure all services are ready
		time.Sleep(5 * time.Minute)

		// Run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			// Check for expiring certificates
			c.checkExpiringCertificates()

			// Wait for next interval
			<-ticker.C
		}
	}()

	log.Println("Certificate expiry checker started")
}

// checkExpiringCertificates checks for certificates that will expire soon
func (c *CronService) checkExpiringCertificates() {
	// Get all certificates
	certs, err := c.certSvc.GetAllCertificates()
	if err != nil {
		log.Printf("Failed to get certificates: %v", err)
		return
	}

	// Get email settings
	smtpServer, smtpPortStr, smtpUser, smtpPass, emailFrom, emailTo, useTLS, useStartTLS, err := c.store.GetEmailSettings()
	if err != nil || smtpServer == "" || emailTo == "" {
		log.Printf("Email settings not configured or error: %v", err)
		return
	}

	// Convert SMTP port to int
	var smtpPort int
	if smtpPort, err = parseInt(smtpPortStr, 25); err != nil {
		log.Printf("Invalid SMTP port, using default 25: %v", err)
		smtpPort = 25
	}

	// Create email service
	emailSvc := email.NewEmailService(smtpServer, smtpPort, smtpUser, smtpPass, useTLS, useStartTLS)

	// Check each certificate
	var expiringCerts []email.CertificateInfo
	for _, cert := range certs {
		// Skip already expired certificates
		if cert.NotAfter.Before(time.Now()) {
			continue
		}

		// Check if expiring within 30 days
		if cert.NotAfter.Before(time.Now().Add(30 * 24 * time.Hour)) {
			expiringCerts = append(expiringCerts, email.CertificateInfo{
				CommonName:   cert.CommonName,
				ExpiryDate:   cert.NotAfter.Format("2006-01-02"),
				IsClient:     cert.IsClient,
				SerialNumber: cert.SerialNumber,
			})
		}
	}

	// Send notifications
	if len(expiringCerts) > 0 {
		notified := emailSvc.CheckCertificatesExpiry(expiringCerts, emailFrom, emailTo, 30)
		if len(notified) > 0 {
			log.Printf("Sent expiry notifications for %d certificates", len(notified))
		}
	}
}

// parseInt safely converts a string to an integer
func parseInt(s string, defaultValue int) (int, error) {
	if s == "" {
		return defaultValue, nil
	}

	// Try to parse as integer
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue, err
	}
	return result, nil
}
