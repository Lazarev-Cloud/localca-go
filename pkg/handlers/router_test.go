package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCSRFMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router with the middleware
	router := gin.New()
	router.Use(csrfMiddleware())

	// Add test routes
	router.GET("/test-get", func(c *gin.Context) {
		token := c.GetString("csrf_token")
		if token == "" {
			t.Error("CSRF token not set in context")
		}
		c.String(http.StatusOK, "success")
	})

	router.POST("/test-post", func(c *gin.Context) {
		token := c.GetString("csrf_token")
		if token == "" {
			t.Error("CSRF token not set in context")
		}
		c.String(http.StatusOK, "success")
	})

	// Test GET request - should succeed and set a token
	req, _ := http.NewRequest("GET", "/test-get", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Get the token from the context (we need to simulate this since it's not in the response)
	var csrfToken string
	// Make another GET request to get a token
	req, _ = http.NewRequest("GET", "/test-get", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// For testing purposes, we'll grab a token from the store
	csrfStore.mutex.RLock()
	for token := range csrfStore.tokens {
		csrfToken = token
		break
	}
	csrfStore.mutex.RUnlock()

	if csrfToken == "" {
		t.Fatal("Failed to get CSRF token")
	}

	// Test POST request with valid token
	form := url.Values{}
	form.Add("csrf_token", csrfToken)
	req, _ = http.NewRequest("POST", "/test-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with valid token, got %d", w.Code)
	}

	// Test POST request with invalid token
	form = url.Values{}
	form.Add("csrf_token", "invalid-token")
	req, _ = http.NewRequest("POST", "/test-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 with invalid token, got %d", w.Code)
	}

	// Test POST request with missing token
	req, _ = http.NewRequest("POST", "/test-post", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 with missing token, got %d", w.Code)
	}
}

func TestGenerateCSRFToken(t *testing.T) {
	token1 := generateCSRFToken()
	token2 := generateCSRFToken()

	if token1 == "" {
		t.Error("Generated token is empty")
	}

	if token1 == token2 {
		t.Error("Generated tokens are not unique")
	}
}

func TestValidateCSRFToken(t *testing.T) {
	// Generate a token
	token := generateCSRFToken()

	// Store it
	csrfStore.mutex.Lock()
	csrfStore.tokens[token] = time.Now().Add(time.Hour)
	csrfStore.mutex.Unlock()

	// Test validation
	if !validateCSRFToken(token) {
		t.Error("Valid token not validated")
	}

	if validateCSRFToken("invalid-token") {
		t.Error("Invalid token incorrectly validated")
	}
}
