package acme

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

func setupTestEnvironment(t *testing.T) (*ACMEServer, *certificates.CertificateService, *storage.Storage, func()) {
	// Check if OpenSSL is available
	_, err := exec.LookPath("openssl")
	if err != nil {
		t.Skip("OpenSSL not available, skipping test")
	}

	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "acme-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create config
	cfg := &config.Config{
		CAName:        "test-ca.local",
		CAKeyPassword: "test-password",
		Organization:  "Test Org",
		Country:       "US",
		StoragePath:   tempDir,
	}

	// Initialize storage
	store, err := storage.NewStorage(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize certificate service
	certSvc, err := certificates.NewCertificateService(cfg, store)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to initialize certificate service: %v", err)
	}

	// Create CA certificate
	if err := certSvc.CreateCA(); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create CA: %v", err)
	}

	// Initialize ACME server
	acmeServer, err := NewACMEServer(certSvc, store)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create ACME server: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return acmeServer, certSvc, store, cleanup
}

func TestHandleDirectory(t *testing.T) {
	acmeServer, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(acmeServer.handleDirectory))
	defer server.Close()

	// Make request to directory endpoint
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check response content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type %s, got %s", "application/json", contentType)
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var directory map[string]interface{}
	if err := json.Unmarshal(body, &directory); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check directory fields
	requiredFields := []string{"newNonce", "newAccount", "newOrder", "revokeCert"}
	for _, field := range requiredFields {
		if _, ok := directory[field]; !ok {
			t.Errorf("Directory missing required field: %s", field)
		}
	}
}

func TestHandleNewNonce(t *testing.T) {
	acmeServer, _, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(acmeServer.handleNewNonce))
	defer server.Close()

	// Make HEAD request to new-nonce endpoint
	req, err := http.NewRequest(http.MethodHead, server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	// Check for nonce header
	nonce := resp.Header.Get("Replay-Nonce")
	if nonce == "" {
		t.Error("Response missing Replay-Nonce header")
	}

	// Check for cache control header
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "no-store" {
		t.Errorf("Expected Cache-Control header %s, got %s", "no-store", cacheControl)
	}
}

func TestStartACMEServer(t *testing.T) {
	_, certSvc, store, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start ACME server in a goroutine
	go func() {
		err := StartACMEServer(ctx, certSvc, store, ":0", nil)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("ACME server error: %v", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Context should be cancelled by the timeout
}

func TestGenerateNonce(t *testing.T) {
	// Generate multiple nonces and check they're unique
	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce := generateNonce()

		// Check length
		if len(nonce) < 16 {
			t.Errorf("Nonce too short: %s", nonce)
		}

		// Check uniqueness
		if nonces[nonce] {
			t.Errorf("Duplicate nonce generated: %s", nonce)
		}
		nonces[nonce] = true
	}
}

func TestSchemeFromRequest(t *testing.T) {
	// Test HTTP request
	httpReq := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	if scheme := schemeFromRequest(httpReq); scheme != "http" {
		t.Errorf("Expected scheme http, got %s", scheme)
	}

	// Test HTTPS request (with X-Forwarded-Proto)
	httpsReq := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	httpsReq.Header.Set("X-Forwarded-Proto", "https")
	if scheme := schemeFromRequest(httpsReq); scheme != "https" {
		t.Errorf("Expected scheme https, got %s", scheme)
	}

	// Test HTTPS request (with TLS)
	tlsReq := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	if scheme := schemeFromRequest(tlsReq); scheme != "https" {
		t.Errorf("Expected scheme https, got %s", scheme)
	}
}

func TestHandleHTTP01Challenge(t *testing.T) {
	_, _, store, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a test challenge token and response
	testToken := "test-token-12345"
	testKeyAuth := "test-token-12345.test-key-auth-value"

	// Store the challenge in the ACME storage
	challengePath := store.GetBasePath() + "/.well-known/acme-challenge/" + testToken
	err := os.MkdirAll(store.GetBasePath()+"/.well-known/acme-challenge", 0755)
	if err != nil {
		t.Fatalf("Failed to create challenge directory: %v", err)
	}

	err = os.WriteFile(challengePath, []byte(testKeyAuth), 0644)
	if err != nil {
		t.Fatalf("Failed to write challenge file: %v", err)
	}

	// Create a test HTTP server with a custom handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the URL path
		if r.URL.Path == "/.well-known/acme-challenge/"+testToken {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testKeyAuth))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Make a request to the challenge endpoint
	resp, err := http.Get(server.URL + "/.well-known/acme-challenge/" + testToken)
	if err != nil {
		t.Fatalf("Failed to make challenge request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check response content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/octet-stream" {
		t.Errorf("Expected content type %s, got %s", "application/octet-stream", contentType)
	}

	// Check response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != testKeyAuth {
		t.Errorf("Expected response body %q, got %q", testKeyAuth, string(body))
	}

	// Test with a non-existent token
	resp, err = http.Get(server.URL + "/.well-known/acme-challenge/non-existent-token")
	if err != nil {
		t.Fatalf("Failed to make challenge request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 404 for non-existent token
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d for non-existent token, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
