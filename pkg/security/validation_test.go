package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid filename", "test.txt", "test.txt"},
		{"Empty input", "", ""},
		{"Path traversal attempt", "../../../etc/passwd", "passwd"},
		{"Complex path", "/var/www/html/../config.txt", "config.txt"},
		{"Dangerous characters", "test<script>alert()</script>.txt", "script.txt"},
		{"Hidden file", ".secret", "secret"},
		{"Special characters", "test@#$%^&*().txt", "test.txt"},
		{"Long filename", strings.Repeat("a", 150), strings.Repeat("a", 100)},
		{"Unicode characters", "тест.txt", "txt"},
		{"Windows path", "C:\\Windows\\System32\\evil.exe", "evil.exe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFileName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCommonName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid domain", "example.com", "example.com"},
		{"Valid subdomain", "api.example.com", "api.example.com"},
		{"Empty input", "", ""},
		{"Long domain", strings.Repeat("a", 100), strings.Repeat("a", 64)},
		{"Invalid characters", "test<script>", "testscript"},
		{"SQL injection attempt", "'; DROP TABLE users; --", "DROPTABLEusers--"},
		{"IP address", "192.168.1.1", "192.168.1.1"},
		{"Wildcard domain", "*.example.com", ".example.com"},
		{"Unicode domain", "测试.com", ".com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCommonName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSerialNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid hex", "1234567890ABCDEF", "1234567890ABCDEF"},
		{"Lower case hex", "abcdef123456", "abcdef123456"},
		{"Empty input", "", ""},
		{"Invalid characters", "123GHIJK", "123"},
		{"Mixed case", "1a2B3c4D", "1a2B3c4D"},
		{"Too long", strings.Repeat("A", 50), strings.Repeat("A", 40)},
		{"Special characters", "123-456-789", "123456789"},
		{"SQL injection", "'; DROP TABLE certs; --", "DABEce"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSerialNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid email", "user@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Empty input", "", false},
		{"Invalid format", "invalid-email", false},
		{"Missing @", "userexample.com", false},
		{"Missing domain", "user@", false},
		{"Missing user", "@example.com", false},
		{"Multiple @", "user@@example.com", false},
		{"Too long", strings.Repeat("a", 250) + "@example.com", false},
		{"Valid with numbers", "user123@example123.com", true},
		{"Valid with special chars", "user.test+tag@example.com", true},
		{"Invalid TLD", "user@example", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmailAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{"Simple password", "password123", "Should allow basic alphanumeric"},
		{"Empty input", "", "Should handle empty string"},
		{"Command injection", "; rm -rf /", "Should remove dangerous characters"},
		{"Shell metacharacters", "pass`word", "Should remove backticks"},
		{"Variable expansion", "pass$word", "Should remove dollar signs"},
		{"Quote injection", "pass'word\"", "Should remove quotes"},
		{"Pipe injection", "pass|word", "Should remove pipes"},
		{"Redirect injection", "pass>word", "Should remove redirects"},
		{"Subprocess injection", "pass(word)", "Should remove parentheses"},
		{"Too long password", strings.Repeat("a", 150), "Should limit length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(tt.input)

			// Check that dangerous characters are removed
			dangerousChars := []string{"`", "$", "\\", "\"", "'", ";", "&", "|", "<", ">", "(", ")", "{", "}", "[", "]"}
			for _, char := range dangerousChars {
				assert.NotContains(t, result, char, "Should not contain dangerous character: %s", char)
			}

			// Check length limit
			assert.LessOrEqual(t, len(result), 100, "Should not exceed 100 characters")
		})
	}
}

func TestValidateSubjectDN(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{"Valid DN component", "example.com", "Should allow valid domain"},
		{"Empty input", "", "Should handle empty string"},
		{"DN injection", "/CN=evil/", "Should remove DN separators"},
		{"Equals injection", "test=value", "Should remove equals"},
		{"Comma injection", "test,value", "Should remove commas"},
		{"Plus injection", "test+value", "Should remove plus signs"},
		{"Special characters", "test@#$%", "Should remove special chars"},
		{"Too long", strings.Repeat("a", 100), "Should limit length"},
		{"Unicode", "测试", "Should handle unicode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSubjectDN(tt.input)

			// Check that DN separator characters are removed
			dnSeparators := []string{"/", "=", ",", "+"}
			for _, char := range dnSeparators {
				assert.NotContains(t, result, char, "Should not contain DN separator: %s", char)
			}

			// Check length limit
			assert.LessOrEqual(t, len(result), 64, "Should not exceed 64 characters")
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{"Normal text", "Hello World", "Should preserve normal text"},
		{"Empty input", "", "Should handle empty string"},
		{"Control characters", "test\x00\x01\x02", "Should remove control chars"},
		{"CRLF injection", "test\r\ninjection", "Should remove CRLF"},
		{"Null bytes", "test\x00injection", "Should remove null bytes"},
		{"Too long input", strings.Repeat("a", 1500), "Should limit length"},
		{"Mixed dangerous", "test\r\n\x00evil", "Should clean all dangerous chars"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)

			// Check that control characters are removed
			assert.NotContains(t, result, "\r", "Should not contain carriage return")
			assert.NotContains(t, result, "\n", "Should not contain line feed")
			assert.NotContains(t, result, "\x00", "Should not contain null bytes")

			// Check length limit
			assert.LessOrEqual(t, len(result), 1000, "Should not exceed 1000 characters")
		})
	}
}

func BenchmarkValidateFileName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateFileName("../../../etc/passwd")
	}
}

func BenchmarkValidatePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidatePassword("pass`word$evil;injection")
	}
}

func BenchmarkSanitizeInput(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SanitizeInput("test\r\n\x00injection\x01\x02")
	}
}
