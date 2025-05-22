package security

import (
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateFileName validates and sanitizes a filename to prevent path traversal attacks
func ValidateFileName(filename string) string {
	if filename == "" {
		return ""
	}

	// Get only the base name (removes any path components)
	safeName := filepath.Base(filename)

	// Remove any potentially dangerous characters
	// Allow only alphanumeric, hyphens, underscores, and dots
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)
	safeName = reg.ReplaceAllString(safeName, "")

	// Prevent files starting with dots (hidden files)
	if strings.HasPrefix(safeName, ".") {
		safeName = strings.TrimPrefix(safeName, ".")
	}

	// Limit length to prevent issues
	if len(safeName) > 100 {
		safeName = safeName[:100]
	}

	return safeName
}

// ValidateCommonName validates a certificate common name
func ValidateCommonName(cn string) string {
	if cn == "" {
		return ""
	}

	// Allow alphanumeric, dots, hyphens for domain names
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-\.]`)
	safeCN := reg.ReplaceAllString(cn, "")

	// Limit length
	if len(safeCN) > 64 {
		safeCN = safeCN[:64]
	}

	return safeCN
}

// ValidateSerialNumber validates a certificate serial number
func ValidateSerialNumber(serial string) string {
	if serial == "" {
		return ""
	}

	// Allow only hex characters for serial numbers
	reg := regexp.MustCompile(`[^a-fA-F0-9]`)
	safeSerial := reg.ReplaceAllString(serial, "")

	// Limit length
	if len(safeSerial) > 40 {
		safeSerial = safeSerial[:40]
	}

	return safeSerial
}

// ValidateEmailAddress performs basic email validation
func ValidateEmailAddress(email string) bool {
	if email == "" {
		return false
	}

	// Basic email regex validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email) && len(email) <= 254
}

// SanitizeInput removes potentially dangerous characters from general input
func SanitizeInput(input string) string {
	if input == "" {
		return ""
	}

	// Remove null bytes and control characters
	reg := regexp.MustCompile(`[\x00-\x1f\x7f]`)
	sanitized := reg.ReplaceAllString(input, "")

	// Remove CRLF to prevent injection
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\n", "")

	// Limit length
	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000]
	}

	return strings.TrimSpace(sanitized)
}