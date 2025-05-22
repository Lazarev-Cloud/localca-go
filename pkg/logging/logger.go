package logging

import (
	"io"
	"os"
	"strings"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/sirupsen/logrus"
)

// Logger wraps logrus with additional functionality
type Logger struct {
	*logrus.Logger
	auditLogger *logrus.Logger
}

// NewLogger creates a new structured logger
func NewLogger(cfg *config.Config) (*Logger, error) {
	// Create main logger
	mainLogger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	mainLogger.SetLevel(level)

	// Set formatter
	switch strings.ToLower(cfg.LogFormat) {
	case "json":
		mainLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	case "text":
		mainLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		mainLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// Set output
	var output io.Writer
	switch strings.ToLower(cfg.LogOutput) {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.LogOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// Fallback to stdout if file can't be opened
			output = os.Stdout
		} else {
			output = file
		}
	}
	mainLogger.SetOutput(output)

	// Create audit logger (always JSON format for structured audit logs)
	auditLogger := logrus.New()
	auditLogger.SetLevel(logrus.InfoLevel)
	auditLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	// Audit logs can go to a separate file or same output
	auditOutput := output
	if cfg.LogOutput != "stdout" && cfg.LogOutput != "stderr" {
		// Create separate audit log file
		auditFile, err := os.OpenFile(cfg.LogOutput+".audit", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			auditOutput = auditFile
		}
	}
	auditLogger.SetOutput(auditOutput)

	return &Logger{
		Logger:      mainLogger,
		auditLogger: auditLogger,
	}, nil
}

// AuditInfo logs an audit event at info level
func (l *Logger) AuditInfo(action, resource, resourceID, userIP, userAgent string, fields logrus.Fields) {
	entry := l.auditLogger.WithFields(logrus.Fields{
		"audit":       true,
		"action":      action,
		"resource":    resource,
		"resource_id": resourceID,
		"user_ip":     userIP,
		"user_agent":  userAgent,
	})

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Info("Audit event")
}

// AuditError logs an audit event at error level
func (l *Logger) AuditError(action, resource, resourceID, userIP, userAgent, errorMsg string, fields logrus.Fields) {
	entry := l.auditLogger.WithFields(logrus.Fields{
		"audit":       true,
		"action":      action,
		"resource":    resource,
		"resource_id": resourceID,
		"user_ip":     userIP,
		"user_agent":  userAgent,
		"error":       errorMsg,
	})

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Error("Audit event failed")
}

// WithFields creates a new entry with the given fields
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithField creates a new entry with a single field
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithError creates a new entry with an error field
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// Certificate operation logging helpers
func (l *Logger) LogCertificateCreated(certName, userIP, userAgent string) {
	l.AuditInfo("create", "certificate", certName, userIP, userAgent, logrus.Fields{
		"certificate_name": certName,
	})
}

func (l *Logger) LogCertificateDeleted(certName, userIP, userAgent string) {
	l.AuditInfo("delete", "certificate", certName, userIP, userAgent, logrus.Fields{
		"certificate_name": certName,
	})
}

func (l *Logger) LogCertificateRevoked(certName, userIP, userAgent string) {
	l.AuditInfo("revoke", "certificate", certName, userIP, userAgent, logrus.Fields{
		"certificate_name": certName,
	})
}

func (l *Logger) LogCertificateDownloaded(certName, userIP, userAgent string) {
	l.AuditInfo("download", "certificate", certName, userIP, userAgent, logrus.Fields{
		"certificate_name": certName,
	})
}

// CA operation logging helpers
func (l *Logger) LogCACreated(caName, userIP, userAgent string) {
	l.AuditInfo("create", "ca", caName, userIP, userAgent, logrus.Fields{
		"ca_name": caName,
	})
}

func (l *Logger) LogCAAccessed(caName, userIP, userAgent string) {
	l.AuditInfo("access", "ca", caName, userIP, userAgent, logrus.Fields{
		"ca_name": caName,
	})
}

// Authentication logging helpers
func (l *Logger) LogAuthSuccess(userIP, userAgent string) {
	l.AuditInfo("authenticate", "auth", "", userIP, userAgent, nil)
}

func (l *Logger) LogAuthFailure(userIP, userAgent, reason string) {
	l.AuditError("authenticate", "auth", "", userIP, userAgent, reason, logrus.Fields{
		"failure_reason": reason,
	})
}

// Configuration logging helpers
func (l *Logger) LogConfigChanged(setting, userIP, userAgent string) {
	l.AuditInfo("update", "config", setting, userIP, userAgent, logrus.Fields{
		"setting": setting,
	})
}

// S3 operation logging helpers
func (l *Logger) LogS3Upload(objectName, userIP, userAgent string) {
	l.AuditInfo("upload", "s3", objectName, userIP, userAgent, logrus.Fields{
		"object_name": objectName,
	})
}

func (l *Logger) LogS3Download(objectName, userIP, userAgent string) {
	l.AuditInfo("download", "s3", objectName, userIP, userAgent, logrus.Fields{
		"object_name": objectName,
	})
}

func (l *Logger) LogS3Delete(objectName, userIP, userAgent string) {
	l.AuditInfo("delete", "s3", objectName, userIP, userAgent, logrus.Fields{
		"object_name": objectName,
	})
}
