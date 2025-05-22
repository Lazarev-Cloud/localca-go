package security

import (
	"fmt"
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

	// Prevent multiple consecutive dots (could be used for traversal)
	for strings.Contains(safeName, "..") {
		safeName = strings.ReplaceAll(safeName, "..", ".")
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

// ValidatePassword validates and sanitizes password input to prevent injection attacks
func ValidatePassword(password string) string {
	if password == "" {
		return ""
	}

	// Remove shell metacharacters and control chars that could be used for injection
	// Allow printable ASCII except shell metacharacters
	reg := regexp.MustCompile(`[^a-zA-Z0-9!@#$%^&*()_+=\[\]{}|:";'<>?,./]`)
	sanitized := reg.ReplaceAllString(password, "")

	// Remove any characters that could be used for command injection
	dangerousChars := []string{"`", "$", "\\", "\"", "'", ";", "&", "|", "<", ">", "(", ")", "{", "}", "[", "]"}
	for _, char := range dangerousChars {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}

	// Limit length for security
	if len(sanitized) > 100 {
		sanitized = sanitized[:100]
	}

	return sanitized
}

// ValidateSubjectDN validates and sanitizes certificate subject DN components
func ValidateSubjectDN(component string) string {
	if component == "" {
		return ""
	}

	// Allow only alphanumeric, dots, hyphens, spaces for DN components
	// Remove characters that could be used for injection or breaking DN parsing
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-\.\s]`)
	sanitized := reg.ReplaceAllString(component, "")

	// Remove DN separator characters that could break parsing
	sanitized = strings.ReplaceAll(sanitized, "/", "")
	sanitized = strings.ReplaceAll(sanitized, "=", "")
	sanitized = strings.ReplaceAll(sanitized, ",", "")
	sanitized = strings.ReplaceAll(sanitized, "+", "")

	// Limit length
	if len(sanitized) > 64 {
		sanitized = sanitized[:64]
	}

	return strings.TrimSpace(sanitized)
}

// ValidateFilePath validates a file path to prevent path traversal attacks
func ValidateFilePath(path string) bool {
	if path == "" {
		return false
	}

	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return false
	}

	// Ensure the path doesn't contain null bytes
	if strings.Contains(cleanPath, "\x00") {
		return false
	}

	// For absolute paths, ensure they don't escape the expected base directory
	if filepath.IsAbs(cleanPath) {
		// Allow only paths under /app or the current working directory
		allowedPrefixes := []string{"/app", "/tmp"}
		isAllowed := false
		for _, prefix := range allowedPrefixes {
			if strings.HasPrefix(cleanPath, prefix) {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return false
		}
	}

	return true
}

// SecureJoinPath safely joins path components and validates the result
func SecureJoinPath(base string, components ...string) (string, error) {
	if base == "" {
		return "", fmt.Errorf("base path cannot be empty")
	}

	// Clean the base path
	cleanBase := filepath.Clean(base)

	// Validate each component
	for _, component := range components {
		if component == "" {
			continue
		}
		
		// Validate the component
		safeComponent := ValidateFileName(component)
		if safeComponent == "" {
			return "", fmt.Errorf("invalid path component: %s", component)
		}
		
		cleanBase = filepath.Join(cleanBase, safeComponent)
	}

	// Final validation of the complete path
	if !ValidateFilePath(cleanBase) {
		return "", fmt.Errorf("resulting path is not safe: %s", cleanBase)
	}

	return cleanBase, nil
}