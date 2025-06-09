package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock certificate service for testing
type mockCertificateService struct {
	certificates map[string]*certificates.Certificate
	nextID       int
}

func newMockCertificateService() *mockCertificateService {
	return &mockCertificateService{
		certificates: make(map[string]*certificates.Certificate),
		nextID:       1,
	}
}

// Implement the CertificateServiceInterface
func (m *mockCertificateService) CAExists() (bool, error) {
	return true, nil
}

func (m *mockCertificateService) CreateCA() error {
	return nil
}

func (m *mockCertificateService) RenewCA() error {
	return nil
}

func (m *mockCertificateService) CreateServiceCertificate() error {
	return nil
}

func (m *mockCertificateService) CreateServerCertificate(commonName string, domains []string) error {
	cert := &certificates.Certificate{
		CommonName:   commonName,
		SerialNumber: fmt.Sprintf("serial_%d", m.nextID),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, 365),
		Issuer:       "LocalCA Test",
		IsClient:     false,
		Path:         fmt.Sprintf("/test/certs/%s", commonName),
	}
	m.certificates[fmt.Sprintf("cert_%d", m.nextID)] = cert
	m.nextID++
	return nil
}

func (m *mockCertificateService) CreateClientCertificate(commonName, password string) error {
	cert := &certificates.Certificate{
		CommonName:   commonName,
		SerialNumber: fmt.Sprintf("serial_%d", m.nextID),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, 365),
		Issuer:       "LocalCA Test",
		IsClient:     true,
		Path:         fmt.Sprintf("/test/certs/%s", commonName),
	}
	m.certificates[fmt.Sprintf("cert_%d", m.nextID)] = cert
	m.nextID++
	return nil
}

func (m *mockCertificateService) RevokeCertificate(name string) error {
	// Find certificate by name
	for _, cert := range m.certificates {
		if cert.CommonName == name {
			// In a real implementation, this would mark the certificate as revoked
			return nil
		}
	}
	return fmt.Errorf("certificate not found: %s", name)
}

func (m *mockCertificateService) RenewServerCertificate(name string) error {
	// Find certificate by name and renew it
	for _, cert := range m.certificates {
		if cert.CommonName == name && !cert.IsClient {
			// Create renewed certificate
			renewed := &certificates.Certificate{
				CommonName:   cert.CommonName,
				SerialNumber: fmt.Sprintf("serial_%d", m.nextID),
				NotBefore:    time.Now(),
				NotAfter:     time.Now().AddDate(0, 0, 365),
				Issuer:       cert.Issuer,
				IsClient:     cert.IsClient,
				Path:         cert.Path,
			}

			newID := fmt.Sprintf("cert_%d", m.nextID)
			m.certificates[newID] = renewed
			m.nextID++
			return nil
		}
	}
	return fmt.Errorf("server certificate not found: %s", name)
}

func (m *mockCertificateService) RenewClientCertificate(name string) error {
	// Find certificate by name and renew it
	for _, cert := range m.certificates {
		if cert.CommonName == name && cert.IsClient {
			// Create renewed certificate
			renewed := &certificates.Certificate{
				CommonName:   cert.CommonName,
				SerialNumber: fmt.Sprintf("serial_%d", m.nextID),
				NotBefore:    time.Now(),
				NotAfter:     time.Now().AddDate(0, 0, 365),
				Issuer:       cert.Issuer,
				IsClient:     cert.IsClient,
				Path:         cert.Path,
			}

			newID := fmt.Sprintf("cert_%d", m.nextID)
			m.certificates[newID] = renewed
			m.nextID++
			return nil
		}
	}
	return fmt.Errorf("client certificate not found: %s", name)
}

func (m *mockCertificateService) GetAllCertificates() ([]certificates.Certificate, error) {
	var certs []certificates.Certificate
	for _, cert := range m.certificates {
		certs = append(certs, *cert)
	}
	return certs, nil
}

func (m *mockCertificateService) GetCertificateInfo(name string) (*certificates.Certificate, error) {
	for _, cert := range m.certificates {
		if cert.CommonName == name {
			return cert, nil
		}
	}
	return nil, fmt.Errorf("certificate not found: %s", name)
}

func TestCertificateOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cert-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create storage and mock service
	store, err := storage.NewStorage(tempDir)
	require.NoError(t, err)

	mockSvc := newMockCertificateService()

	// Setup authenticated session
	authConfig := &AuthConfig{
		AdminUsername:     "admin",
		AdminPasswordHash: "$2a$10$CwTycUXWue0Thq9StjUM0uJ8/jFZntRxJb8A.1Nzeqy.gFw8qtqJO",
		SetupCompleted:    true,
	}
	err = saveAuthConfig(authConfig, store)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Note: We'll need to create a proper SetupAPIRoutes function or mock it
	// For now, let's create a simple test setup
	api := router.Group("/api")
	api.Use(authMiddleware(store))

	// Add certificate routes (simplified for testing)
	api.GET("/certificates", func(c *gin.Context) {
		certs, err := mockSvc.GetAllCertificates()
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"certificates": certs,
			},
		})
	})

	// Create session for authentication
	sessionToken := generateSessionToken()
	createTestSession(t, store, sessionToken, "admin")

	t.Run("Certificate Listing", func(t *testing.T) {
		// Create some test certificates first
		mockSvc.CreateServerCertificate("server1.local", []string{"server1.local"})
		mockSvc.CreateClientCertificate("client1@local", "password")

		req := httptest.NewRequest("GET", "/api/certificates", nil)
		req.Header.Set("Cookie", "session="+sessionToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		certificates := data["certificates"].([]interface{})
		assert.GreaterOrEqual(t, len(certificates), 2)
	})

	t.Run("Authentication Required", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/certificates", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCertificateValidation(t *testing.T) {
	t.Run("Certificate Field Validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			commonName  string
			isValid     bool
			description string
		}{
			{
				name:        "Valid Domain",
				commonName:  "test.local",
				isValid:     true,
				description: "Standard domain name",
			},
			{
				name:        "Valid Subdomain",
				commonName:  "api.test.local",
				isValid:     true,
				description: "Subdomain",
			},
			{
				name:        "Valid Email",
				commonName:  "user@test.local",
				isValid:     true,
				description: "Email format for client certificates",
			},
			{
				name:        "Invalid Double Dot",
				commonName:  "invalid..domain",
				isValid:     false,
				description: "Double dots are invalid",
			},
			{
				name:        "Empty Common Name",
				commonName:  "",
				isValid:     false,
				description: "Empty common name",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test common name validation logic
				isValid := validateCommonName(tc.commonName)
				assert.Equal(t, tc.isValid, isValid, tc.description)
			})
		}
	})
}

func TestCertificateServiceMethods(t *testing.T) {
	mockSvc := newMockCertificateService()

	t.Run("Create Server Certificate", func(t *testing.T) {
		err := mockSvc.CreateServerCertificate("test.local", []string{"test.local", "www.test.local"})
		require.NoError(t, err)
		assert.Equal(t, "test.local", mockSvc.certificates["cert_1"].CommonName)
		assert.False(t, mockSvc.certificates["cert_1"].IsClient)
		assert.NotEmpty(t, mockSvc.certificates["cert_1"].SerialNumber)
	})

	t.Run("Create Client Certificate", func(t *testing.T) {
		err := mockSvc.CreateClientCertificate("client@test.local", "password")
		require.NoError(t, err)
		assert.Equal(t, "client@test.local", mockSvc.certificates["cert_2"].CommonName)
		assert.True(t, mockSvc.certificates["cert_2"].IsClient)
		assert.NotEmpty(t, mockSvc.certificates["cert_2"].SerialNumber)
	})

	t.Run("List Certificates", func(t *testing.T) {
		certs, err := mockSvc.GetAllCertificates()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(certs), 2) // From previous tests
	})

	t.Run("Get Certificate Info", func(t *testing.T) {
		// Create a certificate first
		err := mockSvc.CreateServerCertificate("content.local", []string{"content.local"})
		require.NoError(t, err)

		// Get certificate info by name
		cert, err := mockSvc.GetCertificateInfo("content.local")
		require.NoError(t, err)
		assert.NotNil(t, cert)
		assert.Equal(t, "content.local", cert.CommonName)
		assert.False(t, cert.IsClient)

		// Test non-existent certificate
		_, err = mockSvc.GetCertificateInfo("nonexistent.local")
		assert.Error(t, err)
	})
}

// Helper function to validate common names (simplified version)
func validateCommonName(commonName string) bool {
	if commonName == "" {
		return false
	}
	if strings.Contains(commonName, "..") {
		return false
	}
	return true
}

// Helper function to create a test session
func createTestSession(t *testing.T, store *storage.Storage, sessionToken, username string) {
	sessionData := map[string]interface{}{
		"username":   username,
		"created_at": time.Now().Unix(),
		"expires_at": time.Now().Add(24 * time.Hour).Unix(),
	}

	sessionBytes, _ := json.Marshal(sessionData)

	// Use the same encoding as validateSession function
	sessionFileBase := base64.URLEncoding.EncodeToString([]byte(sessionToken))
	if len(sessionFileBase) > 100 {
		sessionFileBase = sessionFileBase[:100] // Limit filename length
	}
	sessionPath := filepath.Join(store.GetBasePath(), "sessions", sessionFileBase)
	err := os.MkdirAll(filepath.Dir(sessionPath), 0700)
	require.NoError(t, err)

	err = os.WriteFile(sessionPath, sessionBytes, 0600)
	require.NoError(t, err)
}
