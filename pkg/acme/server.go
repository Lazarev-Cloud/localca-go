package acme

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/certificates"
	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// ACMEServer implements an ACME server for automated certificate issuance
type ACMEServer struct {
	certSvc    *certificates.CertificateService
	storage    *storage.Storage
	domains    map[string]bool
	challenges map[string]string
	nonces     map[string]time.Time // Changed to track nonce expiration time
	accounts   map[string]*Account
	mutex      sync.RWMutex
	keyPair    *ecdsa.PrivateKey
	// Rate limiting
	ipRateLimits      map[string]*RateLimit
	accountRateLimits map[string]*RateLimit
	rateLimitMutex    sync.RWMutex
}

// RateLimit represents rate limiting information
type RateLimit struct {
	Count      int
	ResetTime  time.Time
	LastAccess time.Time
}

// NonceExpiration is the time after which a nonce expires
const NonceExpiration = 1 * time.Hour

// RateLimitWindow is the time window for rate limiting
const RateLimitWindow = 1 * time.Hour

// RateLimitMax is the maximum number of requests per window
const RateLimitMax = 100

// RateLimitBurst is the maximum number of requests in a short time period
const RateLimitBurst = 20

// RateLimitBurstWindow is the time window for burst rate limiting
const RateLimitBurstWindow = 1 * time.Minute

// Account represents an ACME account
type Account struct {
	ID        string
	Key       crypto.PublicKey
	Contact   []string
	Status    string
	CreatedAt time.Time
}

// NewACMEServer creates a new ACME server
func NewACMEServer(certSvc *certificates.CertificateService, store *storage.Storage) (*ACMEServer, error) {
	// Create ACME directory if it doesn't exist
	acmeDir := filepath.Join(store.GetBasePath(), "acme")
	if err := os.MkdirAll(acmeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create ACME directory: %w", err)
	}

	// Generate or load server key
	keyPath := filepath.Join(acmeDir, "server.key")
	var keyPair *ecdsa.PrivateKey

	if _, statErr := os.Stat(keyPath); os.IsNotExist(statErr) {
		// Generate new key
		var err error
		keyPair, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ACME server key: %w", err)
		}

		// Save key to file
		keyBytes, err := x509.MarshalECPrivateKey(keyPair)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ACME server key: %w", err)
		}

		keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to create ACME server key file: %w", err)
		}
		defer keyFile.Close()

		if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
			return nil, fmt.Errorf("failed to write ACME server key: %w", err)
		}
	} else {
		// Load existing key
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read ACME server key: %w", err)
		}

		block, _ := pem.Decode(keyBytes)
		if block == nil {
			return nil, fmt.Errorf("failed to decode ACME server key")
		}

		keyPair, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ACME server key: %w", err)
		}
	}

	// Load accounts if they exist
	accounts := make(map[string]*Account)
	accountsPath := filepath.Join(acmeDir, "accounts.json")
	if _, err := os.Stat(accountsPath); err == nil {
		accountsData, err := os.ReadFile(accountsPath)
		if err == nil {
			// Try to unmarshal accounts
			var accountsRaw map[string]json.RawMessage
			if err := json.Unmarshal(accountsData, &accountsRaw); err == nil {
				for id, rawAccount := range accountsRaw {
					var account Account
					if err := json.Unmarshal(rawAccount, &account); err == nil {
						accounts[id] = &account
					}
				}
			}
		}
	}

	// Start a goroutine to clean up expired nonces periodically
	server := &ACMEServer{
		certSvc:           certSvc,
		storage:           store,
		domains:           make(map[string]bool),
		challenges:        make(map[string]string),
		nonces:            make(map[string]time.Time),
		accounts:          accounts,
		keyPair:           keyPair,
		ipRateLimits:      make(map[string]*RateLimit),
		accountRateLimits: make(map[string]*RateLimit),
	}

	// Start cleanup goroutines
	go server.cleanupExpiredNonces()
	go server.cleanupRateLimits()

	return server, nil
}

// cleanupExpiredNonces periodically removes expired nonces
func (s *ACMEServer) cleanupExpiredNonces() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for nonce, expiry := range s.nonces {
			if now.After(expiry) {
				delete(s.nonces, nonce)
			}
		}
		s.mutex.Unlock()
	}
}

// cleanupRateLimits periodically removes expired rate limits
func (s *ACMEServer) cleanupRateLimits() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.rateLimitMutex.Lock()
		now := time.Now()

		// Clean up IP rate limits
		for ip, limit := range s.ipRateLimits {
			if now.After(limit.ResetTime) {
				delete(s.ipRateLimits, ip)
			}
		}

		// Clean up account rate limits
		for account, limit := range s.accountRateLimits {
			if now.After(limit.ResetTime) {
				delete(s.accountRateLimits, account)
			}
		}

		s.rateLimitMutex.Unlock()
	}
}

// checkRateLimit checks if a request is within rate limits
func (s *ACMEServer) checkRateLimit(r *http.Request, accountID string) bool {
	ip := getClientIP(r)
	now := time.Now()

	// Check IP-based rate limit
	s.rateLimitMutex.Lock()
	defer s.rateLimitMutex.Unlock()

	ipLimit, ok := s.ipRateLimits[ip]
	if !ok {
		ipLimit = &RateLimit{
			Count:      1,
			ResetTime:  now.Add(RateLimitWindow),
			LastAccess: now,
		}
		s.ipRateLimits[ip] = ipLimit
		return true
	}

	// Reset if window has expired
	if now.After(ipLimit.ResetTime) {
		ipLimit.Count = 1
		ipLimit.ResetTime = now.Add(RateLimitWindow)
		ipLimit.LastAccess = now
		return true
	}

	// Check if limit exceeded
	if ipLimit.Count >= RateLimitMax {
		return false
	}

	// Check for burst rate limiting - if too many requests in a short time
	if now.Sub(ipLimit.LastAccess) < RateLimitBurstWindow && ipLimit.Count > RateLimitBurst {
		return false
	}

	// Increment counter and update time
	ipLimit.Count++
	ipLimit.LastAccess = now

	// If we have an account ID, also check account-based rate limit
	if accountID != "" {
		acctLimit, ok := s.accountRateLimits[accountID]
		if !ok {
			acctLimit = &RateLimit{
				Count:      1,
				ResetTime:  now.Add(RateLimitWindow),
				LastAccess: now,
			}
			s.accountRateLimits[accountID] = acctLimit
			return true
		}

		// Reset if window has expired
		if now.After(acctLimit.ResetTime) {
			acctLimit.Count = 1
			acctLimit.ResetTime = now.Add(RateLimitWindow)
			acctLimit.LastAccess = now
			return true
		}

		// Check if limit exceeded
		if acctLimit.Count >= RateLimitMax {
			return false
		}

		// Check for burst rate limiting for account
		if now.Sub(acctLimit.LastAccess) < RateLimitBurstWindow && acctLimit.Count > RateLimitBurst {
			return false
		}

		// Increment counter and update time
		acctLimit.Count++
		acctLimit.LastAccess = now
	}

	return true
}

// getClientIP gets the client IP address from a request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return forwardedFor
	}

	// Otherwise use RemoteAddr
	return r.RemoteAddr
}

// SetupRoutes configures the ACME server routes
func (s *ACMEServer) SetupRoutes(router *http.ServeMux) {
	// Directory endpoint
	router.HandleFunc("/acme/directory", s.securityMiddleware(s.handleDirectory))

	// New nonce endpoint
	router.HandleFunc("/acme/new-nonce", s.securityMiddleware(s.handleNewNonce))

	// New account endpoint
	router.HandleFunc("/acme/new-account", s.securityMiddleware(s.handleNewAccount))

	// New order endpoint
	router.HandleFunc("/acme/new-order", s.securityMiddleware(s.handleNewOrder))

	// Account endpoint
	router.HandleFunc("/acme/account/", s.securityMiddleware(s.handleAccount))

	// Order endpoint
	router.HandleFunc("/acme/order/", s.securityMiddleware(s.handleOrder))

	// Authorization endpoint
	router.HandleFunc("/acme/authz/", s.securityMiddleware(s.handleAuthorization))

	// Challenge endpoint
	router.HandleFunc("/acme/challenge/", s.securityMiddleware(s.handleChallenge))

	// Certificate endpoint
	router.HandleFunc("/acme/certificate/", s.securityMiddleware(s.handleCertificate))

	// Revocation endpoint
	router.HandleFunc("/acme/revoke-cert", s.securityMiddleware(s.handleRevocation))
}

// securityMiddleware adds security headers and rate limiting
func (s *ACMEServer) securityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Check rate limit
		if !s.checkRateLimit(r, "") {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// handleDirectory handles the ACME directory endpoint
func (s *ACMEServer) handleDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	baseURL := fmt.Sprintf("%s://%s", schemeFromRequest(r), r.Host)

	directory := map[string]interface{}{
		"newNonce":   baseURL + "/acme/new-nonce",
		"newAccount": baseURL + "/acme/new-account",
		"newOrder":   baseURL + "/acme/new-order",
		"revokeCert": baseURL + "/acme/revoke-cert",
		"keyChange":  baseURL + "/acme/key-change",
		"meta": map[string]interface{}{
			"termsOfService": baseURL + "/acme/terms",
			"website":        baseURL,
			"caaIdentities":  []string{"localca.local"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(directory)
}

// handleNewNonce handles the ACME new-nonce endpoint
func (s *ACMEServer) handleNewNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodHead && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nonce := generateNonce()
	s.mutex.Lock()
	s.nonces[nonce] = time.Now().Add(NonceExpiration) // Store with expiration time
	s.mutex.Unlock()

	w.Header().Set("Replay-Nonce", nonce)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

// validateNonce validates a nonce and removes it if valid
func (s *ACMEServer) validateNonce(nonce string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	expiry, exists := s.nonces[nonce]
	if !exists {
		return false
	}

	// Check if nonce has expired
	if time.Now().After(expiry) {
		delete(s.nonces, nonce)
		return false
	}

	// Remove the nonce to prevent reuse
	delete(s.nonces, nonce)
	return true
}

// handleNewAccount handles the ACME new-account endpoint
func (s *ACMEServer) handleNewAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement account creation
	// This will require:
	// 1. Validating the JWS signature
	// 2. Creating a new account
	// 3. Storing the account information

	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "valid",
		"orders": fmt.Sprintf("%s://%s/acme/orders", schemeFromRequest(r), r.Host),
	})
}

// handleNewOrder handles the ACME new-order endpoint
func (s *ACMEServer) handleNewOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement order creation
	// This will require:
	// 1. Validating the JWS signature
	// 2. Creating a new order
	// 3. Creating authorizations for each domain

	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "pending",
		"expires": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"identifiers": []map[string]string{
			{"type": "dns", "value": "example.com"},
		},
		"authorizations": []string{
			fmt.Sprintf("%s://%s/acme/authz/example", schemeFromRequest(r), r.Host),
		},
		"finalize": fmt.Sprintf("%s://%s/acme/finalize/example", schemeFromRequest(r), r.Host),
	})
}

// handleAccount handles the ACME account endpoint
func (s *ACMEServer) handleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement account management
	// This will require:
	// 1. Validating the JWS signature
	// 2. Retrieving the account
	// 3. Updating the account if necessary

	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "valid",
		"contact": []string{"mailto:admin@example.com"},
		"orders":  fmt.Sprintf("%s://%s/acme/orders", schemeFromRequest(r), r.Host),
	})
}

// handleOrder handles the ACME order endpoint
func (s *ACMEServer) handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement order management
	// This will require:
	// 1. Validating the JWS signature
	// 2. Retrieving the order
	// 3. Updating the order if necessary

	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "pending",
		"expires": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"identifiers": []map[string]string{
			{"type": "dns", "value": "example.com"},
		},
		"authorizations": []string{
			fmt.Sprintf("%s://%s/acme/authz/example", schemeFromRequest(r), r.Host),
		},
		"finalize": fmt.Sprintf("%s://%s/acme/finalize/example", schemeFromRequest(r), r.Host),
	})
}

// handleAuthorization handles the ACME authorization endpoint
func (s *ACMEServer) handleAuthorization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement authorization management
	// This will require:
	// 1. Validating the JWS signature
	// 2. Retrieving the authorization
	// 3. Updating the authorization if necessary

	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "pending",
		"expires": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"identifier": map[string]string{
			"type":  "dns",
			"value": "example.com",
		},
		"challenges": []map[string]interface{}{
			{
				"type":  "http-01",
				"url":   fmt.Sprintf("%s://%s/acme/challenge/http01/example", schemeFromRequest(r), r.Host),
				"token": "token",
			},
			{
				"type":  "dns-01",
				"url":   fmt.Sprintf("%s://%s/acme/challenge/dns01/example", schemeFromRequest(r), r.Host),
				"token": "token",
			},
		},
	})
}

// handleChallenge handles the ACME challenge endpoint
func (s *ACMEServer) handleChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement challenge validation
	// This will require:
	// 1. Validating the JWS signature
	// 2. Retrieving the challenge
	// 3. Validating the challenge
	// 4. Updating the challenge status

	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "valid",
		"type":      "http-01",
		"url":       fmt.Sprintf("%s://%s/acme/challenge/http01/example", schemeFromRequest(r), r.Host),
		"token":     "token",
		"validated": time.Now().Format(time.RFC3339),
	})
}

// handleCertificate handles the ACME certificate endpoint
func (s *ACMEServer) handleCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement certificate issuance
	// This will require:
	// 1. Validating the JWS signature
	// 2. Retrieving the order
	// 3. Verifying that all authorizations are valid
	// 4. Issuing the certificate

	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/pem-certificate-chain")
	w.Write([]byte("-----BEGIN CERTIFICATE-----\nMIIDazCCAlOgAwIBAgIUJlK7RCseiIHMJvTQRFNSGr11lPwwDQYJKoZIhvcNAQEL\n-----END CERTIFICATE-----"))
}

// handleRevocation handles the ACME certificate revocation endpoint
func (s *ACMEServer) handleRevocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement certificate revocation
	// This will require:
	// 1. Validating the JWS signature
	// 2. Retrieving the certificate
	// 3. Verifying that the requester is authorized to revoke the certificate
	// 4. Revoking the certificate

	// For now, return a success response
	w.WriteHeader(http.StatusOK)
}

// Helper functions

// generateNonce generates a random nonce
func generateNonce() string {
	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	return base64.RawURLEncoding.EncodeToString(nonceBytes)
}

// schemeFromRequest determines the scheme (http/https) from the request
func schemeFromRequest(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return "https"
	}
	return "http"
}

// StartACMEServer starts the ACME server
func StartACMEServer(ctx context.Context, certSvc *certificates.CertificateService, store *storage.Storage, addr string, tlsConfig *tls.Config) error {
	acmeServer, err := NewACMEServer(certSvc, store)
	if err != nil {
		return fmt.Errorf("failed to create ACME server: %w", err)
	}

	mux := http.NewServeMux()
	acmeServer.SetupRoutes(mux)

	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
		// Set timeouts to prevent slow client attacks
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("Starting ACME server on %s", addr)
	var listenErr error
	if tlsConfig != nil {
		// Get certificate paths
		certPath := filepath.Join(store.GetBasePath(), "service.crt")
		keyPath := filepath.Join(store.GetBasePath(), "service.key")
		listenErr = server.ListenAndServeTLS(certPath, keyPath)
	} else {
		listenErr = server.ListenAndServe()
	}

	if listenErr != nil && listenErr != http.ErrServerClosed {
		return fmt.Errorf("ACME server error: %w", listenErr)
	}

	return nil
}
