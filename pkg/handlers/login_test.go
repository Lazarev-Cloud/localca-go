package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Note: MockCertificateService is already declared in auth_test.go

// setupTestTemplates creates a temporary directory with minimal template files for testing
func setupTestTemplates(t *testing.T) string {
	// Create a temporary templates directory
	tempDir, err := os.MkdirTemp("", "templates-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Create basic template files
	templates := map[string]string{
		"login.html":     "<html><body>Login to LocalCA</body></html>",
		"setup.html":     "<html><body>LocalCA Initial Setup</body></html>",
		"error.html":     "<html><body>Error: {{.Error}}</body></html>",
		"dashboard.html": "<html><body>Dashboard</body></html>",
	}

	for name, content := range templates {
		filePath := filepath.Join(tempDir, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create template file %s: %v", name, err)
		}
	}

	return tempDir
}

// initTestRouter initializes a test Gin router with templates
func initTestRouter(templatesDir string) *gin.Engine {
	r := gin.New()
	r.SetFuncMap(template.FuncMap{})
	r.LoadHTMLFiles(
		filepath.Join(templatesDir, "login.html"),
		filepath.Join(templatesDir, "setup.html"),
		filepath.Join(templatesDir, "error.html"),
		filepath.Join(templatesDir, "dashboard.html"),
	)
	return r
}

func TestLoginHandler(t *testing.T) {
	// Setup test templates
	templatesDir := setupTestTemplates(t)
	defer os.RemoveAll(templatesDir)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "login-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create mock certificate service
	mockCertSvc := new(MockCertificateService)
	mockCertSvc.On("CAExists").Return(true, nil)

	// Create typed certificate service
	certSvc := wrapCertificateService(mockCertSvc)

	// Setup Gin router with templates
	gin.SetMode(gin.TestMode)
	router := initTestRouter(templatesDir)

	// Add test routes
	router.GET("/login", loginHandler(certSvc, store))

	// Test 1: Setup not completed, should redirect to setup
	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/setup", w.Header().Get("Location"))

	// Test 2: Setup completed, should show login page
	err = completeSetup("admin", "password", store)
	assert.NoError(t, err)

	req = httptest.NewRequest("GET", "/login", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Login to LocalCA")
}

func TestLoginPostHandler(t *testing.T) {
	// Setup test templates
	templatesDir := setupTestTemplates(t)
	defer os.RemoveAll(templatesDir)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "login-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create mock certificate service
	mockCertSvc := new(MockCertificateService)
	mockCertSvc.On("CAExists").Return(true, nil)

	// Create typed certificate service
	certSvc := wrapCertificateService(mockCertSvc)

	// Setup a test user
	username := "testadmin"
	password := "testpassword"
	err = completeSetup(username, password, store)
	assert.NoError(t, err)

	// Setup Gin router with templates
	gin.SetMode(gin.TestMode)
	router := initTestRouter(templatesDir)

	// Add test routes
	router.POST("/login", loginPostHandler(certSvc, store))

	// Test 1: Invalid credentials - missing username/password
	form := url.Values{}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Just check status code since our test template doesn't match the real error message

	// Test 2: Invalid credentials - wrong password
	form = url.Values{}
	form.Add("username", username)
	form.Add("password", "wrongpassword")
	req = httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	// Just check status code since our test template doesn't match the real error message

	// Test 3: Valid credentials
	form = url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	req = httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))

	// Verify session cookie was set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	assert.NotEmpty(t, sessionCookie.Value)
}

func TestSetupHandler(t *testing.T) {
	// Setup test templates
	templatesDir := setupTestTemplates(t)
	defer os.RemoveAll(templatesDir)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "setup-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create mock certificate service
	mockCertSvc := new(MockCertificateService)
	mockCertSvc.On("CAExists").Return(false, nil)
	mockCertSvc.On("CreateCA").Return(nil)

	// Create typed certificate service
	certSvc := wrapCertificateService(mockCertSvc)

	// Create config
	cfg := &config.Config{}

	// Setup Gin router with templates
	gin.SetMode(gin.TestMode)
	router := initTestRouter(templatesDir)

	// Add test routes
	router.GET("/setup", setupHandler(certSvc, store, cfg))

	// Test 1: Setup not completed, should show setup page
	req := httptest.NewRequest("GET", "/setup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "LocalCA Initial Setup")

	// Test 2: Setup completed, should redirect to login
	err = completeSetup("admin", "password", store)
	assert.NoError(t, err)

	req = httptest.NewRequest("GET", "/setup", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}

func TestSetupPostHandler(t *testing.T) {
	// Setup test templates
	templatesDir := setupTestTemplates(t)
	defer os.RemoveAll(templatesDir)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "setup-post-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create mock certificate service
	mockCertSvc := new(MockCertificateService)
	mockCertSvc.On("CAExists").Return(false, nil)
	mockCertSvc.On("CreateCA").Return(nil)
	mockCertSvc.On("CreateServiceCertificate").Return(nil)

	// Create typed certificate service
	certSvc := wrapCertificateService(mockCertSvc)

	// Setup Gin router with templates
	gin.SetMode(gin.TestMode)
	router := initTestRouter(templatesDir)

	// Add test routes
	router.POST("/setup", setupPostHandler(certSvc, store))

	// Get the setup token
	config, err := LoadAuthConfig(store)
	assert.NoError(t, err)
	setupToken := config.SetupToken

	// Test 1: Missing required fields
	form := url.Values{}
	req := httptest.NewRequest("POST", "/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Just check status code since our test template doesn't match the real error message

	// Test 2: Passwords don't match
	form = url.Values{}
	form.Add("username", "admin")
	form.Add("password", "password1")
	form.Add("confirm_password", "password2")
	form.Add("setup_token", setupToken)
	req = httptest.NewRequest("POST", "/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Just check status code since our test template doesn't match the real error message

	// Test 3: Invalid setup token
	form = url.Values{}
	form.Add("username", "admin")
	form.Add("password", "password")
	form.Add("confirm_password", "password")
	form.Add("setup_token", "invalidtoken")
	req = httptest.NewRequest("POST", "/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Just check status code since our test template doesn't match the real error message

	// Test 4: Valid setup
	form = url.Values{}
	form.Add("username", "admin")
	form.Add("password", "password")
	form.Add("confirm_password", "password")
	form.Add("setup_token", setupToken)
	req = httptest.NewRequest("POST", "/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))

	// Verify setup was completed
	config, err = LoadAuthConfig(store)
	assert.NoError(t, err)
	assert.True(t, config.SetupCompleted)
	assert.Equal(t, "admin", config.AdminUsername)
	assert.Empty(t, config.SetupToken)
}

func TestAPILoginHandler(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "api-login-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create mock certificate service
	mockCertSvc := new(MockCertificateService)
	mockCertSvc.On("CAExists").Return(true, nil)

	// Create typed certificate service
	certSvc := wrapCertificateService(mockCertSvc)

	// Setup a test user
	username := "testadmin"
	password := "testpassword"
	err = completeSetup(username, password, store)
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add test routes
	router.POST("/api/login", apiLoginHandler(certSvc, store))

	// Test 1: Invalid JSON
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid JSON format")

	// Test 2: Missing credentials
	loginData := map[string]string{}
	jsonData, _ := json.Marshal(loginData)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Username and password are required")

	// Test 3: Invalid credentials
	loginData = map[string]string{
		"username": username,
		"password": "wrongpassword",
	}
	jsonData, _ = json.Marshal(loginData)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid username or password")

	// Test 4: Valid credentials
	loginData = map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ = json.Marshal(loginData)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "Login successful")
	assert.NotNil(t, response.Data)
	assert.NotEmpty(t, response.Data.(map[string]interface{})["token"])

	// Verify session cookie was set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	assert.NotEmpty(t, sessionCookie.Value)
}

func TestAPISetupHandler(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "api-setup-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create mock certificate service
	mockCertSvc := new(MockCertificateService)
	mockCertSvc.On("CAExists").Return(false, nil)
	mockCertSvc.On("CreateCA").Return(nil)
	mockCertSvc.On("CreateServiceCertificate").Return(nil)

	// Create typed certificate service
	certSvc := wrapCertificateService(mockCertSvc)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add test routes
	router.GET("/api/setup", apiSetupHandler(certSvc, store))
	router.POST("/api/setup", apiSetupHandler(certSvc, store))

	// Test 1: GET setup info
	req := httptest.NewRequest("GET", "/api/setup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Get the setup token
	dataMap := response.Data.(map[string]interface{})
	setupToken := dataMap["setup_token"].(string)
	assert.NotEmpty(t, setupToken)
	assert.False(t, dataMap["setup_completed"].(bool))

	// Test 2: POST with invalid JSON
	req = httptest.NewRequest("POST", "/api/setup", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)

	// Test 3: POST with missing fields
	setupData := map[string]string{}
	jsonData, _ := json.Marshal(setupData)
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "All fields are required")

	// Test 4: POST with mismatched passwords
	setupData = map[string]string{
		"username":         "admin",
		"password":         "password1",
		"confirm_password": "password2",
		"setup_token":      setupToken,
	}
	jsonData, _ = json.Marshal(setupData)
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Passwords do not match")

	// Test 5: POST with invalid token
	setupData = map[string]string{
		"username":         "admin",
		"password":         "password",
		"confirm_password": "password",
		"setup_token":      "invalidtoken",
	}
	jsonData, _ = json.Marshal(setupData)
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid or expired setup token")

	// Test 6: Valid setup
	setupData = map[string]string{
		"username":         "admin",
		"password":         "password",
		"confirm_password": "password",
		"setup_token":      setupToken,
	}
	jsonData, _ = json.Marshal(setupData)
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "Setup completed successfully")

	// Verify setup was completed
	config, err := LoadAuthConfig(store)
	assert.NoError(t, err)
	assert.True(t, config.SetupCompleted)
	assert.Equal(t, "admin", config.AdminUsername)
	assert.Empty(t, config.SetupToken)

	// Test 7: Setup already completed
	setupData = map[string]string{
		"username":         "admin2",
		"password":         "password",
		"confirm_password": "password",
		"setup_token":      setupToken,
	}
	jsonData, _ = json.Marshal(setupData)
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Setup is already completed")
}

func TestLogoutHandler(t *testing.T) {
	// Setup test storage
	tempDir := t.TempDir()
	store, err := storage.NewStorage(tempDir)
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add test routes
	router.GET("/logout", logoutHandler(store))

	// Test logout
	req := httptest.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))

	// Verify session cookie was cleared
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	assert.Equal(t, "", sessionCookie.Value)
	assert.True(t, sessionCookie.MaxAge < 0)
}
