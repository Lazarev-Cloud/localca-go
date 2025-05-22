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
	CAName         string
	CAKeyPassword  string
	Organization   string
	Country        string
	StoragePath    string
	EmailEnabled   bool
	SMTPServer     string
	SMTPPort       int
	SMTPUser       string
	SMTPPassword   string
	SMTPUseTLS     bool
	EmailFrom      string
	EmailTo        string
	TLSEnabled     bool
	DataDir        string
	ListenAddr     string
	AllowLocalhost bool
	// KeyDB cache configuration
	CacheEnabled  bool
	KeyDBHost     string
	KeyDBPort     int
	KeyDBPassword string
	KeyDBDB       int
	CacheTTL      int
	// PostgreSQL configuration
	DatabaseEnabled  bool
	DatabaseHost     string
	DatabasePort     int
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseSSLMode  string
	// S3/MinIO configuration
	S3Enabled    bool
	S3Endpoint   string
	S3AccessKey  string
	S3SecretKey  string
	S3BucketName string
	S3UseSSL     bool
	S3Region     string
	// Logging configuration
	LogLevel  string
	LogFormat string
	LogOutput string
}

// LoadConfig loads the configuration from environment variables or defaults
func LoadConfig() (*Config, error) {
	cfg := &Config{
		CAName:         getEnvOrDefault("CA_NAME", "LocalCA"),
		Organization:   getEnvOrDefault("ORGANIZATION", "LocalCA Organization"),
		Country:        getEnvOrDefault("COUNTRY", "US"),
		DataDir:        getEnvOrDefault("DATA_DIR", "./data"),
		ListenAddr:     getEnvOrDefault("LISTEN_ADDR", ":8080"),
		StoragePath:    getEnv("STORAGE_PATH", "/app/certs"),
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

	// Load KeyDB cache settings
	cacheEnabled := getEnv("CACHE_ENABLED", "false")
	cfg.CacheEnabled = strings.ToLower(cacheEnabled) == "true"

	if cfg.CacheEnabled {
		cfg.KeyDBHost = getEnv("KEYDB_HOST", "localhost")

		keydbPort := getEnv("KEYDB_PORT", "6379")
		port, err := strconv.Atoi(keydbPort)
		if err != nil {
			return nil, errors.New("invalid KEYDB_PORT value")
		}
		cfg.KeyDBPort = port

		cfg.KeyDBPassword = getEnv("KEYDB_PASSWORD", "")

		keydbDB := getEnv("KEYDB_DB", "0")
		db, err := strconv.Atoi(keydbDB)
		if err != nil {
			return nil, errors.New("invalid KEYDB_DB value")
		}
		cfg.KeyDBDB = db

		cacheTTL := getEnv("CACHE_TTL", "3600")
		ttl, err := strconv.Atoi(cacheTTL)
		if err != nil {
			return nil, errors.New("invalid CACHE_TTL value")
		}
		cfg.CacheTTL = ttl
	}

	// Load PostgreSQL database settings
	databaseEnabled := getEnv("DATABASE_ENABLED", "false")
	cfg.DatabaseEnabled = strings.ToLower(databaseEnabled) == "true"

	if cfg.DatabaseEnabled {
		cfg.DatabaseHost = getEnv("DATABASE_HOST", "localhost")

		databasePort := getEnv("DATABASE_PORT", "5432")
		port, err := strconv.Atoi(databasePort)
		if err != nil {
			return nil, errors.New("invalid DATABASE_PORT value")
		}
		cfg.DatabasePort = port

		cfg.DatabaseName = getEnv("DATABASE_NAME", "localca")
		cfg.DatabaseUser = getEnv("DATABASE_USER", "localca")
		cfg.DatabasePassword = getEnv("DATABASE_PASSWORD", "")
		cfg.DatabaseSSLMode = getEnv("DATABASE_SSL_MODE", "disable")

		if cfg.DatabasePassword == "" {
			return nil, errors.New("DATABASE_PASSWORD is required when DATABASE_ENABLED is true")
		}
	}

	// Load S3/MinIO settings
	s3Enabled := getEnv("S3_ENABLED", "false")
	cfg.S3Enabled = strings.ToLower(s3Enabled) == "true"

	if cfg.S3Enabled {
		cfg.S3Endpoint = getEnv("S3_ENDPOINT", "localhost:9000")
		cfg.S3AccessKey = getEnv("S3_ACCESS_KEY", "")
		cfg.S3SecretKey = getEnv("S3_SECRET_KEY", "")
		cfg.S3BucketName = getEnv("S3_BUCKET_NAME", "localca-certificates")
		cfg.S3Region = getEnv("S3_REGION", "us-east-1")

		s3UseSSL := getEnv("S3_USE_SSL", "false")
		cfg.S3UseSSL = strings.ToLower(s3UseSSL) == "true"

		if cfg.S3AccessKey == "" || cfg.S3SecretKey == "" {
			return nil, errors.New("S3_ACCESS_KEY and S3_SECRET_KEY are required when S3_ENABLED is true")
		}
	}

	// Load logging settings
	cfg.LogLevel = getEnv("LOG_LEVEL", "info")
	cfg.LogFormat = getEnv("LOG_FORMAT", "json")
	cfg.LogOutput = getEnv("LOG_OUTPUT", "stdout")

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
