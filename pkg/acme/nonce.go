package acme

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// NonceManager manages nonces for ACME
type NonceManager struct {
	nonces map[string]time.Time
	mutex  sync.Mutex
}

// NewNonceManager creates a new nonce manager
func NewNonceManager() *NonceManager {
	return &NonceManager{
		nonces: make(map[string]time.Time),
	}
}

// GenerateNonce generates a new nonce
func (n *NonceManager) GenerateNonce() (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to base64url
	nonce := base64.RawURLEncoding.EncodeToString(randomBytes)

	// Store nonce with expiry
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.nonces[nonce] = time.Now().Add(1 * time.Hour) // Nonces expire in 1 hour

	return nonce, nil
}

// ValidateNonce validates a nonce and removes it if valid
func (n *NonceManager) ValidateNonce(nonce string) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Check if nonce exists and is not expired
	expiry, exists := n.nonces[nonce]
	if !exists {
		return false
	}

	// Check if nonce is expired
	if time.Now().After(expiry) {
		delete(n.nonces, nonce)
		return false
	}

	// Remove nonce after use
	delete(n.nonces, nonce)
	return true
}

// CleanupExpiredNonces removes expired nonces
func (n *NonceManager) CleanupExpiredNonces() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	now := time.Now()
	for nonce, expiry := range n.nonces {
		if now.After(expiry) {
			delete(n.nonces, nonce)
		}
	}
}