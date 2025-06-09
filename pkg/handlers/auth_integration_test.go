package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationFlow(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "auth-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create storage
	store, err := storage.NewStorage(tempDir)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add the API routes
	SetupAPIRoutes(router, nil, store)

	t.Run("Complete Authentication Flow", func(t *testing.T) {
		// Step 1: Check initial setup status
		t.Run("Initial Setup Status", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/setup", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response.Success)
			data := response.Data.(map[string]interface{})
			assert.False(t, data["setup_completed"].(bool))
			assert.True(t, data["setup_required"].(bool))
			assert.NotEmpty(t, data["setup_token"])
		})

		// Step 2: Complete setup with valid credentials
		t.Run("Complete Setup", func(t *testing.T) {
			// First get the setup token
			req := httptest.NewRequest("GET", "/api/setup", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var setupStatus APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &setupStatus)
			require.NoError(t, err)

			data := setupStatus.Data.(map[string]interface{})
			setupToken := data["setup_token"].(string)

			// Complete setup
			setupData := map[string]string{
				"username":    "admin",
				"password":    "testpassword123",
				"setup_token": setupToken,
			}

			setupJSON, _ := json.Marshal(setupData)
			req = httptest.NewRequest("POST", "/api/setup", bytes.NewBuffer(setupJSON))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response APIResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Setup completed successfully", response.Message)
		})

		// Step 3: Verify setup completion
		t.Run("Verify Setup Completion", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/setup", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			data := response.Data.(map[string]interface{})
			assert.True(t, data["setup_completed"].(bool))
			assert.False(t, data["setup_required"].(bool))
		})

		// Step 4: Test login with correct credentials
		t.Run("Login Success", func(t *testing.T) {
			loginData := map[string]string{
				"username": "admin",
				"password": "testpassword123",
			}

			loginJSON, _ := json.Marshal(loginData)
			req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(loginJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Login successful", response.Message)

			// Check that session cookie is set
			cookies := w.Result().Cookies()
			var sessionCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == "session" {
					sessionCookie = cookie
					break
				}
			}
			require.NotNil(t, sessionCookie)
			assert.NotEmpty(t, sessionCookie.Value)
		})

		// Step 5: Test login with wrong password
		t.Run("Login Wrong Password", func(t *testing.T) {
			loginData := map[string]string{
				"username": "admin",
				"password": "wrongpassword",
			}

			loginJSON, _ := json.Marshal(loginData)
			req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(loginJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.Equal(t, "Invalid credentials", response.Message)
		})

		// Step 6: Test login with wrong username
		t.Run("Login Wrong Username", func(t *testing.T) {
			loginData := map[string]string{
				"username": "wronguser",
				"password": "testpassword123",
			}

			loginJSON, _ := json.Marshal(loginData)
			req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(loginJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.Equal(t, "Invalid credentials", response.Message)
		})
	})
}

func TestSetupValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "setup-validation-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := storage.NewStorage(tempDir)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupAPIRoutes(router, nil, store)

	t.Run("Setup Validation Tests", func(t *testing.T) {
		// Get setup token first
		req := httptest.NewRequest("GET", "/api/setup", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var setupStatus APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &setupStatus)
		require.NoError(t, err)

		data := setupStatus.Data.(map[string]interface{})
		validToken := data["setup_token"].(string)

		testCases := []struct {
			name           string
			username       string
			password       string
			setupToken     string
			expectedStatus int
			expectedError  string
		}{
			{
				name:           "Missing Username",
				username:       "",
				password:       "password123",
				setupToken:     validToken,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "Username and password are required",
			},
			{
				name:           "Missing Password",
				username:       "admin",
				password:       "",
				setupToken:     validToken,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "Username and password are required",
			},
			{
				name:           "Invalid Setup Token",
				username:       "admin",
				password:       "password123",
				setupToken:     "invalid_token",
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "Invalid or expired setup token",
			},
			{
				name:           "Valid Setup",
				username:       "admin",
				password:       "password123",
				setupToken:     validToken,
				expectedStatus: http.StatusOK,
				expectedError:  "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				setupData := map[string]string{
					"username":    tc.username,
					"password":    tc.password,
					"setup_token": tc.setupToken,
				}

				setupJSON, _ := json.Marshal(setupData)
				req := httptest.NewRequest("POST", "/api/setup", bytes.NewBuffer(setupJSON))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tc.expectedStatus, w.Code)

				var response APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tc.expectedStatus == http.StatusOK {
					assert.True(t, response.Success)
				} else {
					assert.False(t, response.Success)
					assert.Equal(t, tc.expectedError, response.Message)
				}
			})
		}
	})
}

func TestLoginFormats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "login-format-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := storage.NewStorage(tempDir)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupAPIRoutes(router, nil, store)

	// Complete setup first with consistent password
	req := httptest.NewRequest("GET", "/api/setup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var setupStatus APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &setupStatus)
	require.NoError(t, err)

	data := setupStatus.Data.(map[string]interface{})
	setupToken := data["setup_token"].(string)

	// Use consistent password for setup and login
	const testPassword = "testpass123"
	setupData := map[string]string{
		"username":    "admin",
		"password":    testPassword,
		"setup_token": setupToken,
	}

	setupJSON, _ := json.Marshal(setupData)
	req = httptest.NewRequest("POST", "/api/setup", bytes.NewBuffer(setupJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	t.Run("Login Format Tests", func(t *testing.T) {
		testCases := []struct {
			name        string
			contentType string
			body        string
			expected    int
		}{
			{
				name:        "JSON Format",
				contentType: "application/json",
				body:        fmt.Sprintf(`{"username":"admin","password":"%s"}`, testPassword),
				expected:    http.StatusOK,
			},
			{
				name:        "Form URL Encoded",
				contentType: "application/x-www-form-urlencoded",
				body:        fmt.Sprintf("username=admin&password=%s", testPassword),
				expected:    http.StatusOK,
			},
			{
				name:        "Invalid JSON",
				contentType: "application/json",
				body:        `{"username":"admin","password":}`,
				expected:    http.StatusBadRequest,
			},
			{
				name:        "Empty Body",
				contentType: "application/json",
				body:        "",
				expected:    http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/api/login", bytes.NewBufferString(tc.body))
				req.Header.Set("Content-Type", tc.contentType)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tc.expected, w.Code)
			})
		}
	})
}

func TestSessionManagement(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "session-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := storage.NewStorage(tempDir)
	require.NoError(t, err)

	t.Run("Session Validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			token    string
			expected bool
		}{
			{
				name:     "Empty Token",
				token:    "",
				expected: false,
			},
			{
				name:     "Short Token",
				token:    "short",
				expected: false,
			},
			{
				name:     "Long Token",
				token:    "very_long_token_that_exceeds_the_maximum_allowed_length_for_security_reasons",
				expected: false,
			},
			{
				name:     "Non-existent Token",
				token:    "dGVzdF90b2tlbl90aGF0X2RvZXNfbm90X2V4aXN0",
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := validateSession(tc.token, store)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("Session Creation and Validation", func(t *testing.T) {
		// Create a valid session
		sessionToken := generateSessionToken()
		require.NotEmpty(t, sessionToken)

		// Create session file using the same encoding as validateSession
		sessionFileBase := base64.URLEncoding.EncodeToString([]byte(sessionToken))
		if len(sessionFileBase) > 100 {
			sessionFileBase = sessionFileBase[:100] // Limit filename length
		}
		sessionPath := filepath.Join(store.GetBasePath(), "sessions", sessionFileBase)
		err := os.MkdirAll(filepath.Dir(sessionPath), 0700)
		require.NoError(t, err)

		sessionData := map[string]interface{}{
			"username":   "admin",
			"created_at": time.Now().Unix(),
			"expires_at": time.Now().Add(24 * time.Hour).Unix(),
		}

		sessionBytes, _ := json.Marshal(sessionData)
		err = os.WriteFile(sessionPath, sessionBytes, 0600)
		require.NoError(t, err)

		// Validate session
		assert.True(t, validateSession(sessionToken, store))

		// Test expired session
		oldTime := time.Now().Add(-10 * time.Hour)
		err = os.Chtimes(sessionPath, oldTime, oldTime)
		require.NoError(t, err)

		assert.False(t, validateSession(sessionToken, store))
	})
}

func TestPasswordSecurity(t *testing.T) {
	t.Run("Password Hashing", func(t *testing.T) {
		passwords := []string{
			"simple",
			"complex_password_123!",
			"unicode_测试_password",
			"very_long_password_with_many_characters_and_symbols_!@#$%^&*()",
		}

		for _, password := range passwords {
			t.Run("Password: "+password, func(t *testing.T) {
				// Hash password
				hash, err := hashPassword(password)
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, password, hash)

				// Verify password
				assert.True(t, checkPasswordHash(password, hash))
				assert.False(t, checkPasswordHash("wrong_password", hash))
			})
		}
	})

	t.Run("Token Generation", func(t *testing.T) {
		tokens := make(map[string]bool)

		// Generate multiple tokens and ensure uniqueness
		for i := 0; i < 100; i++ {
			token := generateSessionToken()
			require.NotEmpty(t, token)
			assert.False(t, tokens[token], "Token should be unique")
			tokens[token] = true
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "concurrent-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := storage.NewStorage(tempDir)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupAPIRoutes(router, nil, store)

	t.Run("Concurrent Login Attempts", func(t *testing.T) {
		// Complete setup first
		req := httptest.NewRequest("GET", "/api/setup", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var setupStatus APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &setupStatus)
		require.NoError(t, err)

		data := setupStatus.Data.(map[string]interface{})
		setupToken := data["setup_token"].(string)

		const testPassword = "testpass123"
		setupData := map[string]string{
			"username":    "admin",
			"password":    testPassword,
			"setup_token": setupToken,
		}

		setupJSON, _ := json.Marshal(setupData)
		req = httptest.NewRequest("POST", "/api/setup", bytes.NewBuffer(setupJSON))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Test concurrent login attempts
		const numGoroutines = 10
		results := make(chan int, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				loginData := map[string]string{
					"username": "admin",
					"password": testPassword,
				}

				loginJSON, _ := json.Marshal(loginData)
				req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(loginJSON))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				results <- w.Code
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < numGoroutines; i++ {
			code := <-results
			if code == http.StatusOK {
				successCount++
			}
		}

		// All concurrent requests should succeed
		assert.Equal(t, numGoroutines, successCount)
	})
}
