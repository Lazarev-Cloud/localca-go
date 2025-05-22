package config

import (
	"errors"
	"os"
	"path/filepath"
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
	DataDir       string
	ListenAddr    string
	AllowLocalhost bool
}

// LoadConfig loads the configuration from environment variables or defaults
func LoadConfig() (*Config, error) {
	cfg := &Config{
		CAName:        getEnvOrDefault("CA_NAME", "LocalCA"),
		Organization:  getEnvOrDefault("ORGANIZATION", "LocalCA Organization"),
		Country:       getEnvOrDefault("COUNTRY", "US"),
		DataDir:       getEnvOrDefault("DATA_DIR", "./data"),
		ListenAddr:    getEnvOrDefault("LISTEN_ADDR", ":8080"),
		StoragePath:   getEnv("STORAGE_PATH", "/app/certs"),
		AllowLocalhost: strings.ToLower(getEnv("ALLOW_LOCALHOST", "false")) == "true",
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, err
	}

	// Validate required fields
	if cfg.CAName == "" {
		return nil, errors.New("CA_NAME environment variable is required")
	}

	// Load CA Key Password - allow empty for fresh installs
	keyFile := getEnv("CA_KEY_FILE", "")
	keyEnv := getEnv("CA_KEY", "")

	if keyFile != "" {
		content, err := os.ReadFile(keyFile)
		if err != nil {
			// If the file doesn't exist during fresh install, that's OK
			if !os.IsNotExist(err) {
				return nil, err
			}
			cfg.CAKeyPassword = "" // Will be set during setup
		} else {
			cfg.CAKeyPassword = strings.TrimSpace(string(content))
		}
	} else if keyEnv != "" {
		cfg.CAKeyPassword = keyEnv
	} else {
		// During fresh install, no CA key is expected
		cfg.CAKeyPassword = ""
	}

	// Load Email settings
	emailEnabled := getEnv("EMAIL_NOTIFY", "false")
	cfg.EmailEnabled = strings.ToLower(emailEnabled) == "true"

	if cfg.EmailEnabled {
		cfg.SMTPServer = getEnv("SMTP_SERVER", "")
		if cfg.SMTPServer == "" {
			return nil, errors.New("SMTP_SERVER is required when EMAIL_NOTIFY is true")
		}

		smtpPort := getEnv("SMTP_PORT", "25")
		port, err := strconv.Atoi(smtpPort)
		if err != nil {
			return nil, errors.New("invalid SMTP_PORT value")
		}
		cfg.SMTPPort = port

		cfg.SMTPUser = getEnv("SMTP_USER", "")
		cfg.SMTPPassword = getEnv("SMTP_PASSWORD", "")

		smtpTLS := getEnv("SMTP_USE_TLS", "false")
		cfg.SMTPUseTLS = strings.ToLower(smtpTLS) == "true"

		cfg.EmailFrom = getEnv("EMAIL_FROM", "")
		cfg.EmailTo = getEnv("EMAIL_TO", "")
	}

	// Load TLS settings
	tlsEnabled := getEnv("TLS_ENABLED", "false")
	cfg.TLSEnabled = strings.ToLower(tlsEnabled) == "true"

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// GetCADirectory returns the path to the CA directory
func (c *Config) GetCADirectory() string {
	return filepath.Join(c.DataDir, "ca")
}

// GetCertificatesDirectory returns the path to the certificates directory
func (c *Config) GetCertificatesDirectory() string {
	return c.DataDir
}
