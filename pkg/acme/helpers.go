package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GetBaseURL returns the base URL for the ACME service
func (a *ACMEService) GetBaseURL() string {
	return a.baseURL
}

// GenerateNonce generates a new nonce
func (a *ACMEService) GenerateNonce() (string, error) {
	return a.nonceManager.GenerateNonce()
}

// ValidateNonce validates a nonce
func (a *ACMEService) ValidateNonce(nonce string) bool {
	return a.nonceManager.ValidateNonce(nonce)
}

// GetAccount retrieves an account from storage
func (a *ACMEService) GetAccount(accountID string) (*Account, error) {
	store := NewACMEStore(a.store)
	return store.GetAccount(accountID)
}

// SaveAccount saves an account to storage
func (a *ACMEService) SaveAccount(account *Account) error {
	store := NewACMEStore(a.store)
	return store.SaveAccount(account)
}

// GetOrder retrieves an order from storage
func (a *ACMEService) GetOrder(orderID string) (*Order, error) {
	store := NewACMEStore(a.store)
	return store.GetOrder(orderID)
}

// SaveOrder saves an order to storage
func (a *ACMEService) SaveOrder(order *Order) error {
	store := NewACMEStore(a.store)
	return store.SaveOrder(order)
}

// GetAuthorization retrieves an authorization from storage
func (a *ACMEService) GetAuthorization(authzID string) (*Authorization, error) {
	store := NewACMEStore(a.store)
	return store.GetAuthorization(authzID)
}

// SaveAuthorization saves an authorization to storage
func (a *ACMEService) SaveAuthorization(authz *Authorization) error {
	store := NewACMEStore(a.store)
	return store.SaveAuthorization(authz)
}

// GetChallenge retrieves a challenge from storage
func (a *ACMEService) GetChallenge(challengeID string) (*Challenge, error) {
	store := NewACMEStore(a.store)
	return store.GetChallenge(challengeID)
}

// SaveChallenge saves a challenge to storage
func (a *ACMEService) SaveChallenge(challenge *Challenge) error {
	store := NewACMEStore(a.store)
	return store.SaveChallenge(challenge)
}

// ComputeKeyAuthorization computes the key authorization for a challenge
func ComputeKeyAuthorization(token string, jwk *JWK) (string, error) {
	// Compute JWK thumbprint
	thumbprint, err := ThumbprintJWK(jwk)
	if err != nil {
		return "", fmt.Errorf("failed to compute JWK thumbprint: %w", err)
	}

	// Key authorization = token + "." + JWK thumbprint
	return token + "." + thumbprint, nil
}

// HTTP-01 challenge verification

// HTTP01ChallengeVerifier verifies an HTTP-01 challenge
type HTTP01ChallengeVerifier struct {
	client *http.Client
}

// NewHTTP01ChallengeVerifier creates a new HTTP-01 challenge verifier
func NewHTTP01ChallengeVerifier() *HTTP01ChallengeVerifier {
	return &HTTP01ChallengeVerifier{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Verify verifies an HTTP-01 challenge
func (v *HTTP01ChallengeVerifier) Verify(domain, token, keyAuth string) error {
	// Compute URL
	url := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s", domain, token)

	// Make request
	resp, err := v.client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP-01 challenge verification failed - connection error: %w", err)
	}
	defer resp.Body.Close()

	// Check response code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP-01 challenge verification failed - response status: %s", resp.Status)
	}

	// Read response body
	buf := make([]byte, len(keyAuth)+1)
	n, err := resp.Body.Read(buf)
	if err != nil && (err.Error() != "EOF" || n == 0) {
		return fmt.Errorf("HTTP-01 challenge verification failed - response read error: %w", err)
	}

	// Compare key authorization
	if string(buf[:n]) != keyAuth {
		return fmt.Errorf("HTTP-01 challenge verification failed - response does not match key authorization")
	}

	return nil
}

// DNS-01 challenge verification

// DNS01ChallengeVerifier verifies a DNS-01 challenge
type DNS01ChallengeVerifier struct{}

// NewDNS01ChallengeVerifier creates a new DNS-01 challenge verifier
func NewDNS01ChallengeVerifier() *DNS01ChallengeVerifier {
	return &DNS01ChallengeVerifier{}
}

// ComputeDNS01KeyAuthorizationDigest computes the DNS-01 key authorization digest
func ComputeDNS01KeyAuthorizationDigest(keyAuth string) string {
	h := sha256.Sum256([]byte(keyAuth))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// Verify verifies a DNS-01 challenge
func (v *DNS01ChallengeVerifier) Verify(domain, token, keyAuth string) error {
	// Compute digest
	digest := ComputeDNS01KeyAuthorizationDigest(keyAuth)

	// Create domain name for TXT record
	recordName := fmt.Sprintf("_acme-challenge.%s", domain)

	// Look up TXT records
	txtRecords, err := net.LookupTXT(recordName)
	if err != nil {
		return fmt.Errorf("DNS-01 challenge verification failed - DNS lookup error: %w", err)
	}

	// Check if digest is in TXT records
	for _, record := range txtRecords {
		if record == digest {
			return nil
		}
	}

	return fmt.Errorf("DNS-01 challenge verification failed - TXT record not found or does not match")
}

// TLS-ALPN-01 challenge verification (simplified)

// TLSALPN01ChallengeVerifier verifies a TLS-ALPN-01 challenge
type TLSALPN01ChallengeVerifier struct{}

// NewTLSALPN01ChallengeVerifier creates a new TLS-ALPN-01 challenge verifier
func NewTLSALPN01ChallengeVerifier() *TLSALPN01ChallengeVerifier {
	return &TLSALPN01ChallengeVerifier{}
}

// Verify verifies a TLS-ALPN-01 challenge
// Note: This is simplified - a real implementation would verify the ALPN certificate
func (v *TLSALPN01ChallengeVerifier) Verify(domain, token, keyAuth string) error {
	// For a real implementation, we would:
	// 1. Open a TLS connection to the domain with ALPN protocol "acme-tls/1"
	// 2. Verify that the returned certificate contains the acmeValidation extension
	// 3. Verify that the extension value matches the SHA-256 digest of the key authorization
	
	// We'll mark this as unsupported in our implementation
	return fmt.Errorf("TLS-ALPN-01 verification is not supported in this implementation")
}

// verifyHTTP01Challenge verifies an HTTP-01 challenge
func (a *ACMEService) verifyHTTP01Challenge(challenge *Challenge, account *Account) error {
	// Get authorization
	authz, err := a.GetAuthorization(challenge.AuthorizationID)
	if err != nil {
		return fmt.Errorf("failed to get authorization: %w", err)
	}

	// Extract domain from authorization
	domain := authz.Identifier.Value

	// Compute key authorization for this challenge
	var jwk JWK
	if err := json.Unmarshal(account.Key, &jwk); err != nil {
		return fmt.Errorf("failed to unmarshal account JWK: %w", err)
	}

	keyAuth, err := ComputeKeyAuthorization(challenge.Token, &jwk)
	if err != nil {
		return fmt.Errorf("failed to compute key authorization: %w", err)
	}

	// Store key authorization
	challenge.KeyAuthorization = keyAuth

	// Verify HTTP-01 challenge
	verifier := NewHTTP01ChallengeVerifier()
	return verifier.Verify(domain, challenge.Token, keyAuth)
}

// verifyDNS01Challenge verifies a DNS-01 challenge
func (a *ACMEService) verifyDNS01Challenge(challenge *Challenge, account *Account) error {
	// Get authorization
	authz, err := a.GetAuthorization(challenge.AuthorizationID)
	if err != nil {
		return fmt.Errorf("failed to get authorization: %w", err)
	}

	// Extract domain from authorization
	domain := authz.Identifier.Value

	// Compute key authorization for this challenge
	var jwk JWK
	if err := json.Unmarshal(account.Key, &jwk); err != nil {
		return fmt.Errorf("failed to unmarshal account JWK: %w", err)
	}

	keyAuth, err := ComputeKeyAuthorization(challenge.Token, &jwk)
	if err != nil {
		return fmt.Errorf("failed to compute key authorization: %w", err)
	}

	// Store key authorization
	challenge.KeyAuthorization = keyAuth

	// Verify DNS-01 challenge
	verifier := NewDNS01ChallengeVerifier()
	return verifier.Verify(domain, challenge.Token, keyAuth)
}

// verifyTLSALPN01Challenge verifies a TLS-ALPN-01 challenge
func (a *ACMEService) verifyTLSALPN01Challenge(challenge *Challenge, account *Account) error {
	// Get authorization
	authz, err := a.GetAuthorization(challenge.AuthorizationID)
	if err != nil {
		return fmt.Errorf("failed to get authorization: %w", err)
	}

	// Extract domain from authorization
	domain := authz.Identifier.Value

	// Compute key authorization for this challenge
	var jwk JWK
	if err := json.Unmarshal(account.Key, &jwk); err != nil {
		return fmt.Errorf("failed to unmarshal account JWK: %w", err)
	}

	keyAuth, err := ComputeKeyAuthorization(challenge.Token, &jwk)
	if err != nil {
		return fmt.Errorf("failed to compute key authorization: %w", err)
	}

	// Store key authorization
	challenge.KeyAuthorization = keyAuth

	// Verify TLS-ALPN-01 challenge
	verifier := NewTLSALPN01ChallengeVerifier()
	return verifier.Verify(domain, challenge.Token, keyAuth)
}

// IsValidURL checks if a URL is valid
func IsValidURL(u string) bool {
	_, err := url.Parse(u)
	return err == nil
}

// IsValidIP checks if an IP address is valid
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidEmail checks if an email address is valid
// This is a simplified check, a real implementation would be more robust
func IsValidEmail(email string) bool {
	if !strings.HasPrefix(email, "mailto:") {
		return false
	}
	email = email[7:]
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}