package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	CAName        string
	CAKeyPassword string
	Organization  string
	Country       string
	StoragePath   string
	EmailEnabled  bool
	SMTPServer    string
	SMTPPort      int
	SMTPUser      string
	SMTPPassword  string
	SMTPUseTLS    bool
	EmailFrom     string
	EmailTo       string
	TLSEnabled    bool
	HTTPPort      int
	HTTPSPort     int
	Hostname      string
	ACMEEnabled   bool
	ACMEBaseURL   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		StoragePath: getEnv("STORAGE_PATH", "/app/certs"),
	}

	// Load CA Name
	config.CAName = getEnv("CA_NAME", "")
	if config.CAName == "" {
		return nil, errors.New("CA_NAME environment variable is required")
	}

	// Load CA Key Password
	keyFile := getEnv("CA_KEY_FILE", "")
	keyEnv := getEnv("CA_KEY", "")

	if keyFile != "" {
		content, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, err
		}
		config.CAKeyPassword = strings.TrimSpace(string(content))
	} else if keyEnv != "" {
		config.CAKeyPassword = keyEnv
	} else {
		// Default to a secure password if none provided
		config.CAKeyPassword = "secure-default-password"
	}

	// Load Organization and Country
	config.Organization = getEnv("ORGANIZATION", "LocalCA")
	config.Country = getEnv("COUNTRY", "US")

	// Load Email settings
	emailEnabled := getEnv("EMAIL_NOTIFY", "false")
	config.EmailEnabled = strings.ToLower(emailEnabled) == "true"

	if config.EmailEnabled {
		config.SMTPServer = getEnv("SMTP_SERVER", "")
		if config.SMTPServer == "" {
			return nil, errors.New("SMTP_SERVER is required when EMAIL_NOTIFY is true")
		}

		smtpPort := getEnv("SMTP_PORT", "25")
		port, err := strconv.Atoi(smtpPort)
		if err != nil {
			return nil, errors.New("invalid SMTP_PORT value")
		}
		config.SMTPPort = port

		config.SMTPUser = getEnv("SMTP_USER", "")
		config.SMTPPassword = getEnv("SMTP_PASSWORD", "")

		smtpTLS := getEnv("SMTP_USE_TLS", "false")
		config.SMTPUseTLS = strings.ToLower(smtpTLS) == "true"

		config.EmailFrom = getEnv("EMAIL_FROM", "")
		config.EmailTo = getEnv("EMAIL_TO", "")
	}

	// Load server configuration
	config.Hostname = getEnv("HOSTNAME", "localhost")
	
	// Load HTTP port 
	httpPort := getEnv("HTTP_PORT", "8080")
	port, err := strconv.Atoi(httpPort)
	if err != nil {
		return nil, errors.New("invalid HTTP_PORT value")
	}
	config.HTTPPort = port

	// Load TLS settings
	tlsEnabled := getEnv("TLS_ENABLED", "false")
	config.TLSEnabled = strings.ToLower(tlsEnabled) == "true"

	// Load HTTPS port if TLS is enabled
	if config.TLSEnabled {
		httpsPort := getEnv("HTTPS_PORT", "8443")
		port, err := strconv.Atoi(httpsPort)
		if err != nil {
			return nil, errors.New("invalid HTTPS_PORT value")
		}
		config.HTTPSPort = port
	}

	// Load ACME settings
	acmeEnabled := getEnv("ACME_ENABLED", "false")
	config.ACMEEnabled = strings.ToLower(acmeEnabled) == "true"
	
	if config.ACMEEnabled {
		config.ACMEBaseURL = getEnv("ACME_BASE_URL", "")
		// Will be auto-set in main.go if not provided
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}