package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
	assert.NotNil(t, logger.auditLogger)
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "invalid",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	// Should default to info level
	assert.Equal(t, logrus.InfoLevel, logger.Logger.Level)
}

func TestNewLogger_TextFormat(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "debug",
		LogFormat: "text",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.IsType(t, &logrus.TextFormatter{}, logger.Logger.Formatter)
}

func TestNewLogger_JSONFormat(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "debug",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	assert.IsType(t, &logrus.JSONFormatter{}, logger.Logger.Formatter)
}

func TestLogger_AuditInfo(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Capture audit logger output
	var buf bytes.Buffer
	logger.auditLogger.SetOutput(&buf)

	// Log audit event
	fields := logrus.Fields{"extra": "data"}
	logger.AuditInfo("create", "certificate", "test.com", "192.168.1.1", "test-agent", fields)

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// Verify audit fields
	assert.Equal(t, true, logEntry["audit"])
	assert.Equal(t, "create", logEntry["action"])
	assert.Equal(t, "certificate", logEntry["resource"])
	assert.Equal(t, "test.com", logEntry["resource_id"])
	assert.Equal(t, "192.168.1.1", logEntry["user_ip"])
	assert.Equal(t, "test-agent", logEntry["user_agent"])
	assert.Equal(t, "data", logEntry["extra"])
	assert.Equal(t, "Audit event", logEntry["msg"])
}

func TestLogger_AuditError(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Capture audit logger output
	var buf bytes.Buffer
	logger.auditLogger.SetOutput(&buf)

	// Log audit error
	logger.AuditError("delete", "certificate", "test.com", "192.168.1.1", "test-agent", "permission denied", nil)

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	// Verify audit fields
	assert.Equal(t, true, logEntry["audit"])
	assert.Equal(t, "delete", logEntry["action"])
	assert.Equal(t, "certificate", logEntry["resource"])
	assert.Equal(t, "test.com", logEntry["resource_id"])
	assert.Equal(t, "192.168.1.1", logEntry["user_ip"])
	assert.Equal(t, "test-agent", logEntry["user_agent"])
	assert.Equal(t, "permission denied", logEntry["error"])
	assert.Equal(t, "Audit event failed", logEntry["msg"])
	assert.Equal(t, "error", logEntry["level"])
}

func TestLogger_LogCertificateCreated(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Capture audit logger output
	var buf bytes.Buffer
	logger.auditLogger.SetOutput(&buf)

	logger.LogCertificateCreated("example.com", "192.168.1.1", "test-agent")

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "create", logEntry["action"])
	assert.Equal(t, "certificate", logEntry["resource"])
	assert.Equal(t, "example.com", logEntry["resource_id"])
	assert.Equal(t, "example.com", logEntry["certificate_name"])
}

func TestLogger_LogAuthSuccess(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Capture audit logger output
	var buf bytes.Buffer
	logger.auditLogger.SetOutput(&buf)

	logger.LogAuthSuccess("192.168.1.1", "test-agent")

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "authenticate", logEntry["action"])
	assert.Equal(t, "auth", logEntry["resource"])
	assert.Equal(t, "192.168.1.1", logEntry["user_ip"])
	assert.Equal(t, "test-agent", logEntry["user_agent"])
}

func TestLogger_WithFields(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Capture main logger output
	var buf bytes.Buffer
	logger.Logger.SetOutput(&buf)

	// Log with fields
	logger.WithFields(logrus.Fields{
		"component": "test",
		"operation": "validate",
	}).Info("Test message")

	// Check that output contains the fields
	output := buf.String()
	assert.Contains(t, output, "component")
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "operation")
	assert.Contains(t, output, "validate")
}

func TestLogger_LevelFiltering(t *testing.T) {
	cfg := &config.Config{
		LogLevel:  "warn",
		LogFormat: "json",
		LogOutput: "stdout",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Capture main logger output
	var buf bytes.Buffer
	logger.Logger.SetOutput(&buf)

	// Log at different levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	output := buf.String()

	// Debug and Info should be filtered out
	assert.NotContains(t, output, "Debug message")
	assert.NotContains(t, output, "Info message")

	// Warn and Error should be included
	assert.Contains(t, output, "Warning message")
	assert.Contains(t, output, "Error message")
}

func TestLogger_FileOutput(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.Config{
		LogLevel:  "info",
		LogFormat: "json",
		LogOutput: tmpFile.Name(),
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Log a message
	logger.Info("Test file output")

	// Read the file
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify content
	assert.Contains(t, string(content), "Test file output")
}
