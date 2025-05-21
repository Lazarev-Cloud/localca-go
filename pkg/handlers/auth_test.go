package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCertificateService is a mock implementation for testing
type MockCertificateService struct {
	mock.Mock
}

// CAExists mocks CAExists method
func (m *MockCertificateService) CAExists() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// CreateCA mocks CreateCA method
func (m *MockCertificateService) CreateCA() error {
	args := m.Called()
	return args.Error(0)
}

// CreateServiceCertificate mocks CreateServiceCertificate method
func (m *MockCertificateService) CreateServiceCertificate() error {
	args := m.Called()
	return args.Error(0)
}

// RevokeCertificate mocks RevokeCertificate method
func (m *MockCertificateService) RevokeCertificate(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// CreateServerCertificate mocks CreateServerCertificate method
func (m *MockCertificateService) CreateServerCertificate(commonName string, domains []string) error {
	args := m.Called(commonName, domains)
	return args.Error(0)
}

// CreateClientCertificate mocks CreateClientCertificate method
func (m *MockCertificateService) CreateClientCertificate(commonName string, password string) error {
	args := m.Called(commonName, password)
	return args.Error(0)
}

// RenewServerCertificate mocks RenewServerCertificate method
func (m *MockCertificateService) RenewServerCertificate(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// RenewClientCertificate mocks RenewClientCertificate method
func (m *MockCertificateService) RenewClientCertificate(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// RenewCA mocks RenewCA method
func (m *MockCertificateService) RenewCA() error {
	args := m.Called()
	return args.Error(0)
}

// wrapCertificateService wraps the mock as a CertificateService
func wrapCertificateService(m *MockCertificateService) *certificates.CertificateService {
	// This is a hack to make the type system happy
	// In real usage, we only care about the interface methods
	return &certificates.CertificateService{}
}

func TestLoadAuthConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "auth-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test loading auth config (should create default config)
	config, err := LoadAuthConfig(store)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, DefaultAdminUsername, config.AdminUsername)
	assert.False(t, config.SetupCompleted)
	assert.NotEmpty(t, config.SetupToken)

	// Verify config file was created
	configPath := filepath.Join(tempDir, "auth.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Test loading existing config
	secondConfig, err := LoadAuthConfig(store)
	assert.NoError(t, err)
	assert.Equal(t, config.SetupToken, secondConfig.SetupToken)
}

func TestSaveAuthConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "auth-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create test config
	config := &AuthConfig{
		AdminUsername:     "testadmin",
		AdminPasswordHash: "testhash",
		SetupCompleted:    true,
		SetupToken:        "testtoken",
		SetupTokenExpiry:  time.Now().Add(time.Hour),
	}

	// Save config
	err = saveAuthConfig(config, store)
	assert.NoError(t, err)

	// Verify config file was created
	configPath := filepath.Join(tempDir, "auth.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Read and parse the config file
	data, err := os.ReadFile(configPath)
	assert.NoError(t, err)

	var savedConfig AuthConfig
	err = json.Unmarshal(data, &savedConfig)
	assert.NoError(t, err)

	// Verify saved values
	assert.Equal(t, config.AdminUsername, savedConfig.AdminUsername)
	assert.Equal(t, config.AdminPasswordHash, savedConfig.AdminPasswordHash)
	assert.Equal(t, config.SetupCompleted, savedConfig.SetupCompleted)
	assert.Equal(t, config.SetupToken, savedConfig.SetupToken)
}

func TestGenerateSetupToken(t *testing.T) {
	token := generateSetupToken()
	assert.NotEmpty(t, token)

	// Generate another token to verify randomness
	token2 := generateSetupToken()
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token, token2)
}

func TestValidateSetupToken(t *testing.T) {
	// Valid token
	validConfig := &AuthConfig{
		SetupToken:       "validtoken",
		SetupTokenExpiry: time.Now().Add(time.Hour),
	}
	assert.True(t, validateSetupToken(validConfig, "validtoken"))

	// Invalid token
	assert.False(t, validateSetupToken(validConfig, "invalidtoken"))

	// Expired token
	expiredConfig := &AuthConfig{
		SetupToken:       "expiredtoken",
		SetupTokenExpiry: time.Now().Add(-time.Hour),
	}
	assert.False(t, validateSetupToken(expiredConfig, "expiredtoken"))

	// Empty token
	emptyConfig := &AuthConfig{
		SetupToken: "",
	}
	assert.False(t, validateSetupToken(emptyConfig, "anytoken"))
}

func TestGenerateSessionToken(t *testing.T) {
	token := generateSessionToken()
	assert.NotEmpty(t, token)

	// Generate another token to verify randomness
	token2 := generateSessionToken()
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token, token2)
}

func TestHashAndCheckPassword(t *testing.T) {
	password := "testpassword"

	// Hash password
	hash, err := hashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Check valid password
	assert.True(t, checkPasswordHash(password, hash))

	// Check invalid password
	assert.False(t, checkPasswordHash("wrongpassword", hash))
}

func TestCompleteSetup(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "auth-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test complete setup
	username := "testadmin"
	password := "testpassword"

	// Complete setup
	err = completeSetup(username, password, store)
	assert.NoError(t, err)

	// Verify config
	config, err := LoadAuthConfig(store)
	assert.NoError(t, err)
	assert.Equal(t, username, config.AdminUsername)
	assert.True(t, config.SetupCompleted)
	assert.Empty(t, config.SetupToken)
	assert.True(t, checkPasswordHash(password, config.AdminPasswordHash))
}

func TestIsPublicPath(t *testing.T) {
	// Test public paths
	publicPaths := []string{
		"/static/css/style.css",
		"/login",
		"/api/login",
		"/.well-known/acme-challenge/token",
		"/download/ca",
		"/download/crl",
		"/acme/directory",
	}

	for _, path := range publicPaths {
		assert.True(t, isPublicPath(path), "Path should be public: %s", path)
	}

	// Test non-public paths
	nonPublicPaths := []string{
		"/",
		"/settings",
		"/api/certificates",
		"/files",
	}

	for _, path := range nonPublicPaths {
		assert.False(t, isPublicPath(path), "Path should not be public: %s", path)
	}
}

func TestAuthMiddleware(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "auth-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authMiddleware(store))

	// Add test routes
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// Test 1: Public path should pass through
	req := httptest.NewRequest("GET", "/static/test.css", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code) // 404 because the route doesn't exist, but middleware should pass

	// Test 2: Protected path should redirect to setup page when setup not completed
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/setup", w.Header().Get("Location"))

	// Test 3: Setup path should be accessible when setup not completed
	router = gin.New()
	router.Use(authMiddleware(store))
	router.GET("/setup", func(c *gin.Context) {
		c.String(http.StatusOK, "setup page")
	})

	req = httptest.NewRequest("GET", "/setup", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 4: Complete setup and verify protected path requires session
	err = completeSetup("admin", "password", store)
	assert.NoError(t, err)

	router = gin.New()
	router.Use(authMiddleware(store))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}
