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
	nonces     map[string]bool
	accounts   map[string]*Account
	mutex      sync.RWMutex
	keyPair    *ecdsa.PrivateKey
}

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

	return &ACMEServer{
		certSvc:    certSvc,
		storage:    store,
		domains:    make(map[string]bool),
		challenges: make(map[string]string),
		nonces:     make(map[string]bool),
		accounts:   accounts,
		keyPair:    keyPair,
	}, nil
}

// SetupRoutes configures the ACME server routes
func (s *ACMEServer) SetupRoutes(router *http.ServeMux) {
	// Directory endpoint
	router.HandleFunc("/acme/directory", s.handleDirectory)

	// New nonce endpoint
	router.HandleFunc("/acme/new-nonce", s.handleNewNonce)

	// New account endpoint
	router.HandleFunc("/acme/new-account", s.handleNewAccount)

	// New order endpoint
	router.HandleFunc("/acme/new-order", s.handleNewOrder)

	// Account endpoint
	router.HandleFunc("/acme/account/", s.handleAccount)

	// Order endpoint
	router.HandleFunc("/acme/order/", s.handleOrder)

	// Authorization endpoint
	router.HandleFunc("/acme/authz/", s.handleAuthorization)

	// Challenge endpoint
	router.HandleFunc("/acme/challenge/", s.handleChallenge)

	// Certificate endpoint
	router.HandleFunc("/acme/certificate/", s.handleCertificate)

	// Revocation endpoint
	router.HandleFunc("/acme/revoke-cert", s.handleRevocation)
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
	s.nonces[nonce] = true
	s.mutex.Unlock()

	w.Header().Set("Replay-Nonce", nonce)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
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
