package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/email"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// settingsHandler handles the settings page
func settingsHandler(certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get email settings
		smtpServer, smtpPort, smtpUser, smtpPass, emailFrom, emailTo, useTLS, useStartTLS, err := store.GetEmailSettings()
		if err != nil {
			log.Printf("Failed to get email settings: %v", err)
		}

		// Render template
		c.HTML(http.StatusOK, "settings.html", gin.H{
			"SMTPServer":  smtpServer,
			"SMTPPort":    smtpPort,
			"SMTPUser":    smtpUser,
			"SMTPPass":    smtpPass,
			"EmailFrom":   emailFrom,
			"EmailTo":     emailTo,
			"UseTLS":      useTLS,
			"UseStartTLS": useStartTLS,
			"CSRFToken":   c.GetString("csrf_token"),
		})
	}
}

// saveSettingsHandler handles saving email settings
func saveSettingsHandler(certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get form data
		smtpServer := c.PostForm("smtp_server")
		smtpPort := c.PostForm("smtp_port")
		smtpUser := c.PostForm("smtp_user")
		smtpPass := c.PostForm("smtp_pass")
		emailFrom := c.PostForm("email_from")
		emailTo := c.PostForm("email_to")
		useTLS := c.PostForm("use_tls") == "on"
		useStartTLS := c.PostForm("use_starttls") == "on"

		// Validate input
		if smtpServer == "" {
			c.HTML(http.StatusBadRequest, "settings.html", gin.H{
				"Error":       "SMTP Server is required",
				"SMTPServer":  smtpServer,
				"SMTPPort":    smtpPort,
				"SMTPUser":    smtpUser,
				"SMTPPass":    smtpPass,
				"EmailFrom":   emailFrom,
				"EmailTo":     emailTo,
				"UseTLS":      useTLS,
				"UseStartTLS": useStartTLS,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Default port
		if smtpPort == "" {
			smtpPort = "25"
		}

		// Save settings
		if err := store.SaveEmailSettings(smtpServer, smtpPort, smtpUser, smtpPass, emailFrom, emailTo, useTLS, useStartTLS); err != nil {
			log.Printf("Failed to save email settings: %v", err)
			c.HTML(http.StatusInternalServerError, "settings.html", gin.H{
				"Error":       fmt.Sprintf("Failed to save settings: %v", err),
				"SMTPServer":  smtpServer,
				"SMTPPort":    smtpPort,
				"SMTPUser":    smtpUser,
				"SMTPPass":    smtpPass,
				"EmailFrom":   emailFrom,
				"EmailTo":     emailTo,
				"UseTLS":      useTLS,
				"UseStartTLS": useStartTLS,
				"CSRFToken":   c.GetString("csrf_token"),
			})
			return
		}

		// Redirect to settings page
		c.Redirect(http.StatusSeeOther, "/settings")
	}
}

// testEmailHandler tests email settings
func testEmailHandler(certSvc *certificates.CertificateService, store *storage.Storage, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get test email address
		testEmail := c.PostForm("test_email")
		if testEmail == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Test email address is required",
			})
			return
		}

		// Get email settings
		smtpServer, smtpPortStr, smtpUser, smtpPass, emailFrom, _, useTLS, useStartTLS, err := store.GetEmailSettings()
		if err != nil {
			log.Printf("Failed to get email settings: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to get email settings: %v", err),
			})
			return
		}

		// Convert port to int
		smtpPort, err := strconv.Atoi(smtpPortStr)
		if err != nil {
			log.Printf("Invalid SMTP port: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Invalid SMTP port: %v", err),
			})
			return
		}

		// Create email service
		emailSvc := email.NewEmailService(smtpServer, smtpPort, smtpUser, smtpPass, useTLS, useStartTLS)

		// Send test email
		subject := "LocalCA Test Email"
		body := "This is a test email from LocalCA."
		if err := emailSvc.SendEmail(emailFrom, testEmail, subject, body); err != nil {
			log.Printf("Failed to send test email: %v", err)
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send test email: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Test email sent successfully",
		})
	}
}