package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
)

// FileInfo represents information about a certificate file
type FileInfo struct {
	Name     string
	Path     string
	Content  string
	IsP12    bool
	FileSize string
}

// CertificateDetails represents details of a certificate
type CertificateDetails struct {
	CommonName      string
	Issuer          string
	Serial          string
	NotBefore       string
	NotAfter        string
	Subject         string
	SubjectAltNames []string
	KeyUsage        string
	ExtKeyUsage     string
}

// filesHandler handles the certificate files page
func filesHandler(certSvc certificates.CertificateServiceInterface, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get certificate name
		name := c.Query("name")
		if name == "" {
			// Use a fixed, internal URL to prevent open redirect vulnerabilities
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		// Validate certificate name to prevent path traversal
		// Only allow alphanumeric characters, hyphens, and underscores
		if !isValidCertName(name) {
			c.HTML(http.StatusBadRequest, "files.html", gin.H{
				"Error": "Invalid certificate name",
			})
			return
		}

		// Get the safe name (base name only)
		safeName := filepath.Base(name)

		// Check if certificate exists
		certDir := store.GetCertificateDirectory(safeName)
		if _, err := os.Stat(certDir); os.IsNotExist(err) {
			c.HTML(http.StatusNotFound, "files.html", gin.H{
				"Error": "Certificate not found",
			})
			return
		}

		// Get certificate details
		certDetails, err := getCertificateDetails(store.GetCertificatePath(safeName))
		if err != nil {
			log.Printf("Failed to get certificate details: %v", err)
		}

		// Get certificate files
		files, err := os.ReadDir(certDir)
		if err != nil {
			log.Printf("Failed to read certificate directory: %v", err)
			c.HTML(http.StatusInternalServerError, "files.html", gin.H{
				"Error": fmt.Sprintf("Failed to read certificate directory: %v", err),
			})
			return
		}

		// Process files
		fileInfos := make([]FileInfo, 0, len(files))
		for _, file := range files {
			// Skip password file and revoked flag
			if strings.HasSuffix(file.Name(), ".pw") || file.Name() == "revoked" {
				continue
			}

			// Sanitize file name to prevent path traversal
			safeFileName := filepath.Base(file.Name())
			filePath := filepath.Join(certDir, safeFileName)

			info, err := file.Info()
			if err != nil {
				log.Printf("Failed to get file info: %v", err)
				continue
			}

			// Check file size
			fileSize := fmt.Sprintf("%.1f KB", float64(info.Size())/1024.0)

			// Check if p12 file
			isP12 := strings.HasSuffix(safeFileName, ".p12")

			// Read file content for non-p12 files
			content := ""
			if !isP12 {
				contentBytes, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Failed to read file: %v", err)
					content = fmt.Sprintf("Error reading file: %v", err)
				} else {
					content = string(contentBytes)
				}
			}

			fileInfos = append(fileInfos, FileInfo{
				Name:     safeFileName,
				Path:     filePath,
				Content:  content,
				IsP12:    isP12,
				FileSize: fileSize,
			})
		}

		// Render template
		c.HTML(http.StatusOK, "files.html", gin.H{
			"Name":               safeName,
			"Files":              fileInfos,
			"CertificateDetails": certDetails,
			"CSRFToken":          c.GetString("csrf_token"),
		})
	}
}

// isValidCertName checks if a certificate name is valid
// Only allows alphanumeric characters, hyphens, and underscores
func isValidCertName(name string) bool {
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.') {
			return false
		}
	}
	return true
}

// getCertificateDetails gets detailed information about a certificate
func getCertificateDetails(certPath string) (CertificateDetails, error) {
	details := CertificateDetails{}

	// Validate certificate path to prevent command injection
	if !filepath.IsAbs(certPath) {
		return details, fmt.Errorf("certificate path must be absolute")
	}

	// Check if the file exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return details, fmt.Errorf("certificate file does not exist: %w", err)
	}

	// Find OpenSSL path
	opensslPath, err := exec.LookPath("openssl")
	if err != nil {
		return details, fmt.Errorf("failed to find openssl executable: %w", err)
	}

	// Use OpenSSL to get certificate details
	cmd := exec.Command(
		opensslPath, "x509",
		"-in", certPath,
		"-noout",
		"-text",
	)
	output, err := cmd.Output()
	if err != nil {
		return details, fmt.Errorf("failed to get certificate details: %w", err)
	}

	// Parse output
	outputStr := string(output)

	// Extract serial number
	if idx := strings.Index(outputStr, "Serial Number:"); idx != -1 {
		serialPart := outputStr[idx:]
		if endIdx := strings.Index(serialPart, "\n"); endIdx != -1 {
			details.Serial = strings.TrimSpace(serialPart[15:endIdx])
		}
	}

	// Extract issuer
	if idx := strings.Index(outputStr, "Issuer:"); idx != -1 {
		issuerPart := outputStr[idx:]
		if endIdx := strings.Index(issuerPart, "\n"); endIdx != -1 {
			details.Issuer = strings.TrimSpace(issuerPart[7:endIdx])
		}
	}

	// Extract subject
	if idx := strings.Index(outputStr, "Subject:"); idx != -1 {
		subjectPart := outputStr[idx:]
		if endIdx := strings.Index(subjectPart, "\n"); endIdx != -1 {
			details.Subject = strings.TrimSpace(subjectPart[8:endIdx])
		}
	}

	// Extract validity
	if idx := strings.Index(outputStr, "Not Before:"); idx != -1 {
		validPart := outputStr[idx:]
		if endIdx := strings.Index(validPart, "\n"); endIdx != -1 {
			notBeforeStr := strings.TrimSpace(validPart[11:endIdx])
			if t, err := time.Parse("Jan 2 15:04:05 2006 MST", notBeforeStr); err == nil {
				details.NotBefore = t.Format("2006-01-02 15:04:05")
			} else {
				details.NotBefore = notBeforeStr
			}
		}
	}

	if idx := strings.Index(outputStr, "Not After :"); idx != -1 {
		validPart := outputStr[idx:]
		if endIdx := strings.Index(validPart, "\n"); endIdx != -1 {
			notAfterStr := strings.TrimSpace(validPart[11:endIdx])
			if t, err := time.Parse("Jan 2 15:04:05 2006 MST", notAfterStr); err == nil {
				details.NotAfter = t.Format("2006-01-02 15:04:05")
			} else {
				details.NotAfter = notAfterStr
			}
		}
	}

	// Extract common name
	if details.Subject != "" {
		if idx := strings.Index(details.Subject, "CN="); idx != -1 {
			cnPart := details.Subject[idx+3:]
			if endIdx := strings.Index(cnPart, ","); endIdx != -1 {
				details.CommonName = cnPart[:endIdx]
			} else {
				details.CommonName = cnPart
			}
		}
	}

	// Extract subject alternative names
	if idx := strings.Index(outputStr, "X509v3 Subject Alternative Name:"); idx != -1 {
		sanPart := outputStr[idx:]
		if endIdx := strings.Index(sanPart, "\n\n"); endIdx != -1 {
			sanLine := sanPart[:endIdx]
			if valueIdx := strings.Index(sanLine, "DNS:"); valueIdx != -1 {
				sanValues := sanLine[valueIdx:]
				for _, san := range strings.Split(sanValues, ", ") {
					if strings.HasPrefix(san, "DNS:") {
						details.SubjectAltNames = append(details.SubjectAltNames, strings.TrimPrefix(san, "DNS:"))
					}
				}
			}
		}
	}

	// Extract key usage
	if idx := strings.Index(outputStr, "X509v3 Key Usage:"); idx != -1 {
		kuPart := outputStr[idx:]
		if valueIdx := strings.Index(kuPart, "\n"); valueIdx != -1 {
			kuLine := kuPart[:valueIdx]
			if nextLineIdx := strings.Index(kuPart[valueIdx+1:], "\n"); nextLineIdx != -1 {
				details.KeyUsage = strings.TrimSpace(kuPart[valueIdx+1 : valueIdx+1+nextLineIdx]) // Existing logic
			}
			details.KeyUsage = strings.TrimSpace(kuLine) // Use kuLine here
		}
	}

	// Extract extended key usage
	if idx := strings.Index(outputStr, "X509v3 Extended Key Usage:"); idx != -1 {
		ekuPart := outputStr[idx:]
		if valueIdx := strings.Index(ekuPart, "\n"); valueIdx != -1 {
			ekuLine := ekuPart[:valueIdx]
			if nextLineIdx := strings.Index(ekuPart[valueIdx+1:], "\n"); nextLineIdx != -1 {
				details.ExtKeyUsage = strings.TrimSpace(ekuPart[valueIdx+1 : valueIdx+1+nextLineIdx]) // Existing logic
			}
			details.ExtKeyUsage = strings.TrimSpace(ekuLine) // Use ekuLine here
		}
	}

	return details, nil
}
