package acme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// ACMEStore handles ACME-specific storage operations
type ACMEStore struct {
	store      *storage.Storage
	baseDir    string
	accountsDir string
	ordersDir   string
	authzDir    string
	challengesDir string
	nonceDir     string
	mutex      sync.RWMutex
}

// NewACMEStore creates a new ACME storage handler
func NewACMEStore(store *storage.Storage) (*ACMEStore, error) {
	baseDir := filepath.Join(store.GetBasePath(), "acme")
	accountsDir := filepath.Join(baseDir, "accounts")
	ordersDir := filepath.Join(baseDir, "orders")
	authzDir := filepath.Join(baseDir, "authz")
	challengesDir := filepath.Join(baseDir, "challenges")
	nonceDir := filepath.Join(baseDir, "nonces")

	// Create directories
	for _, dir := range []string{baseDir, accountsDir, ordersDir, authzDir, challengesDir, nonceDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create ACME directory %s: %w", dir, err)
		}
	}

	return &ACMEStore{
		store:        store,
		baseDir:      baseDir,
		accountsDir:  accountsDir,
		ordersDir:    ordersDir,
		authzDir:     authzDir,
		challengesDir: challengesDir,
		nonceDir:     nonceDir,
	}, nil
}

// SaveAccount saves an ACME account
func (s *ACMEStore) SaveAccount(account *Account) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate account
	if account.ID == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	// Sanitize account ID for use in filename
	accountID, err := sanitizeID(account.ID)
	if err != nil {
		return err
	}

	// Marshal account to JSON
	data, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	// Save account to file
	accountPath := filepath.Join(s.accountsDir, accountID+".json")
	if err := os.WriteFile(accountPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write account file: %w", err)
	}

	return nil
}

// GetAccount retrieves an ACME account
func (s *ACMEStore) GetAccount(accountID string) (*Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Sanitize account ID
	sanitizedID, err := sanitizeID(accountID)
	if err != nil {
		return nil, err
	}

	// Read account file
	accountPath := filepath.Join(s.accountsDir, sanitizedID+".json")
	data, err := os.ReadFile(accountPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("account not found: %s", accountID)
		}
		return nil, fmt.Errorf("failed to read account file: %w", err)
	}

	// Unmarshal account
	var account Account
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

// FindAccountByKey finds an account by its JWK
func (s *ACMEStore) FindAccountByKey(jwk []byte) (*Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// List all account files
	files, err := os.ReadDir(s.accountsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts directory: %w", err)
	}

	// Check each account
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Read account file
		accountPath := filepath.Join(s.accountsDir, file.Name())
		data, err := os.ReadFile(accountPath)
		if err != nil {
			continue
		}

		// Unmarshal account
		var account Account
		if err := json.Unmarshal(data, &account); err != nil {
			continue
		}

		// Compare JWK
		if account.Key != nil && string(account.Key) == string(jwk) {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account not found for the given key")
}

// SaveOrder saves an ACME order
func (s *ACMEStore) SaveOrder(order *Order) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate order
	if order.ID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	// Sanitize order ID
	orderID, err := sanitizeID(order.ID)
	if err != nil {
		return err
	}

	// Marshal order to JSON
	data, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Save order to file
	orderPath := filepath.Join(s.ordersDir, orderID+".json")
	if err := os.WriteFile(orderPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write order file: %w", err)
	}

	return nil
}

// GetOrder retrieves an ACME order
func (s *ACMEStore) GetOrder(orderID string) (*Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Sanitize order ID
	sanitizedID, err := sanitizeID(orderID)
	if err != nil {
		return nil, err
	}

	// Read order file
	orderPath := filepath.Join(s.ordersDir, sanitizedID+".json")
	data, err := os.ReadFile(orderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("order not found: %s", orderID)
		}
		return nil, fmt.Errorf("failed to read order file: %w", err)
	}

	// Unmarshal order
	var order Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// ListOrdersByAccount lists all orders for an account
func (s *ACMEStore) ListOrdersByAccount(accountID string) ([]Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Sanitize account ID
	sanitizedID, err := sanitizeID(accountID)
	if err != nil {
		return nil, err
	}

	// List all order files
	files, err := os.ReadDir(s.ordersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read orders directory: %w", err)
	}

	// Filter orders by account ID
	var orders []Order
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Read order file
		orderPath := filepath.Join(s.ordersDir, file.Name())
		data, err := os.ReadFile(orderPath)
		if err != nil {
			continue
		}

		// Unmarshal order
		var order Order
		if err := json.Unmarshal(data, &order); err != nil {
			continue
		}

		// Check if order belongs to the account
		if order.AccountID == sanitizedID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// SaveAuthorization saves an ACME authorization
func (s *ACMEStore) SaveAuthorization(authz *Authorization) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate authorization
	if authz.ID == "" {
		return fmt.Errorf("authorization ID cannot be empty")
	}

	// Sanitize authorization ID
	authzID, err := sanitizeID(authz.ID)
	if err != nil {
		return err
	}

	// Marshal authorization to JSON
	data, err := json.MarshalIndent(authz, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal authorization: %w", err)
	}

	// Save authorization to file
	authzPath := filepath.Join(s.authzDir, authzID+".json")
	if err := os.WriteFile(authzPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write authorization file: %w", err)
	}

	return nil
}

// GetAuthorization retrieves an ACME authorization
func (s *ACMEStore) GetAuthorization(authzID string) (*Authorization, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Sanitize authorization ID
	sanitizedID, err := sanitizeID(authzID)
	if err != nil {
		return nil, err
	}

	// Read authorization file
	authzPath := filepath.Join(s.authzDir, sanitizedID+".json")
	data, err := os.ReadFile(authzPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("authorization not found: %s", authzID)
		}
		return nil, fmt.Errorf("failed to read authorization file: %w", err)
	}

	// Unmarshal authorization
	var authz Authorization
	if err := json.Unmarshal(data, &authz); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization: %w", err)
	}

	return &authz, nil
}

// SaveChallenge saves an ACME challenge
func (s *ACMEStore) SaveChallenge(challenge *Challenge) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate challenge
	if challenge.ID == "" {
		return fmt.Errorf("challenge ID cannot be empty")
	}

	// Sanitize challenge ID
	challengeID, err := sanitizeID(challenge.ID)
	if err != nil {
		return err
	}

	// Marshal challenge to JSON
	data, err := json.MarshalIndent(challenge, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal challenge: %w", err)
	}

	// Save challenge to file
	challengePath := filepath.Join(s.challengesDir, challengeID+".json")
	if err := os.WriteFile(challengePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write challenge file: %w", err)
	}

	return nil
}

// GetChallenge retrieves an ACME challenge
func (s *ACMEStore) GetChallenge(challengeID string) (*Challenge, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Sanitize challenge ID
	sanitizedID, err := sanitizeID(challengeID)
	if err != nil {
		return nil, err
	}

	// Read challenge file
	challengePath := filepath.Join(s.challengesDir, sanitizedID+".json")
	data, err := os.ReadFile(challengePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("challenge not found: %s", challengeID)
		}
		return nil, fmt.Errorf("failed to read challenge file: %w", err)
	}

	// Unmarshal challenge
	var challenge Challenge
	if err := json.Unmarshal(data, &challenge); err != nil {
		return nil, fmt.Errorf("failed to unmarshal challenge: %w", err)
	}

	return &challenge, nil
}

// SaveNonce saves an ACME nonce
func (s *ACMEStore) SaveNonce(nonce string, expiry time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Sanitize nonce
	sanitizedNonce, err := sanitizeID(nonce)
	if err != nil {
		return err
	}

	// Create nonce data
	nonceData := struct {
		Nonce  string    `json:"nonce"`
		Expiry time.Time `json:"expiry"`
	}{
		Nonce:  nonce,
		Expiry: expiry,
	}

	// Marshal nonce to JSON
	data, err := json.Marshal(nonceData)
	if err != nil {
		return fmt.Errorf("failed to marshal nonce: %w", err)
	}

	// Save nonce to file
	noncePath := filepath.Join(s.nonceDir, sanitizedNonce+".json")
	if err := os.WriteFile(noncePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write nonce file: %w", err)
	}

	return nil
}

// ValidateNonce checks if a nonce is valid and removes it
func (s *ACMEStore) ValidateNonce(nonce string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Sanitize nonce
	sanitizedNonce, err := sanitizeID(nonce)
	if err != nil {
		return false, err
	}

	// Read nonce file
	noncePath := filepath.Join(s.nonceDir, sanitizedNonce+".json")
	data, err := os.ReadFile(noncePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read nonce file: %w", err)
	}

	// Unmarshal nonce
	var nonceData struct {
		Nonce  string    `json:"nonce"`
		Expiry time.Time `json:"expiry"`
	}
	if err := json.Unmarshal(data, &nonceData); err != nil {
		return false, fmt.Errorf("failed to unmarshal nonce: %w", err)
	}

	// Check if nonce is expired
	if time.Now().After(nonceData.Expiry) {
		// Remove expired nonce
		os.Remove(noncePath)
		return false, nil
	}

	// Remove used nonce
	os.Remove(noncePath)

	return true, nil
}

// CleanupExpiredNonces removes expired nonces
func (s *ACMEStore) CleanupExpiredNonces() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// List all nonce files
	files, err := os.ReadDir(s.nonceDir)
	if err != nil {
		return fmt.Errorf("failed to read nonces directory: %w", err)
	}

	// Check each nonce
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Read nonce file
		noncePath := filepath.Join(s.nonceDir, file.Name())
		data, err := os.ReadFile(noncePath)
		if err != nil {
			continue
		}

		// Unmarshal nonce
		var nonceData struct {
			Nonce  string    `json:"nonce"`
			Expiry time.Time `json:"expiry"`
		}
		if err := json.Unmarshal(data, &nonceData); err != nil {
			continue
		}

		// Check if nonce is expired
		if time.Now().After(nonceData.Expiry) {
			// Remove expired nonce
			os.Remove(noncePath)
		}
	}

	return nil
}

// sanitizeID sanitizes an ID for use in a filename
func sanitizeID(id string) (string, error) {
	// Replace any non-alphanumeric characters with underscores
	id = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, id)

	// Check if ID is still valid
	if id == "" {
		return "", fmt.Errorf("invalid ID after sanitization")
	}

	return id, nil
}