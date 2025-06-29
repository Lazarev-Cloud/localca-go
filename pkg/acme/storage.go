package acme

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"crypto/x509"
)

// ACMEStorage handles storage of ACME data
type ACMEStorage struct {
	basePath   string
	mutex      sync.RWMutex
	accounts   map[string]*Account
	orders     map[string]*Order
	authzs     map[string]*Authorization
	challenges map[string]*Challenge
}

// NewACMEStorage creates a new ACME storage
func NewACMEStorage(basePath string) (*ACMEStorage, error) {
	// Create ACME directory if it doesn't exist
	acmeDir := filepath.Join(basePath, "acme")
	if err := os.MkdirAll(acmeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create ACME directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{"accounts", "orders", "authz", "challenges"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(acmeDir, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create ACME subdirectory %s: %w", dir, err)
		}
	}

	storage := &ACMEStorage{
		basePath:   acmeDir,
		accounts:   make(map[string]*Account),
		orders:     make(map[string]*Order),
		authzs:     make(map[string]*Authorization),
		challenges: make(map[string]*Challenge),
	}

	// Load existing data
	if err := storage.loadData(); err != nil {
		return nil, fmt.Errorf("failed to load ACME data: %w", err)
	}

	return storage, nil
}

// loadData loads existing ACME data from disk
func (s *ACMEStorage) loadData() error {
	// Load accounts
	accountsDir := filepath.Join(s.basePath, "accounts")
	files, err := os.ReadDir(accountsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read accounts directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Sanitize file name to prevent path traversal
		fileName := filepath.Base(file.Name())
		filePath := filepath.Join(accountsDir, fileName)

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var account Account
		if err := json.Unmarshal(data, &account); err != nil {
			continue
		}

		s.accounts[account.ID] = &account
	}

	// Load orders
	ordersDir := filepath.Join(s.basePath, "orders")
	files, err = os.ReadDir(ordersDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read orders directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Sanitize file name to prevent path traversal
		fileName := filepath.Base(file.Name())
		filePath := filepath.Join(ordersDir, fileName)

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var order Order
		if err := json.Unmarshal(data, &order); err != nil {
			continue
		}

		s.orders[order.ID] = &order
	}

	// Load authorizations
	authzDir := filepath.Join(s.basePath, "authz")
	files, err = os.ReadDir(authzDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read authz directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Sanitize file name to prevent path traversal
		fileName := filepath.Base(file.Name())
		filePath := filepath.Join(authzDir, fileName)

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var authz Authorization
		if err := json.Unmarshal(data, &authz); err != nil {
			continue
		}

		s.authzs[authz.ID] = &authz
	}

	// Load challenges
	challengesDir := filepath.Join(s.basePath, "challenges")
	files, err = os.ReadDir(challengesDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read challenges directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Sanitize file name to prevent path traversal
		fileName := filepath.Base(file.Name())
		filePath := filepath.Join(challengesDir, fileName)

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var challenge Challenge
		if err := json.Unmarshal(data, &challenge); err != nil {
			continue
		}

		s.challenges[challenge.ID] = &challenge
	}

	return nil
}

// SaveAccount saves an account to disk
func (s *ACMEStorage) SaveAccount(account *Account) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store in memory
	s.accounts[account.ID] = account

	// Store on disk
	data, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	filePath := filepath.Join(s.basePath, "accounts", account.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write account file: %w", err)
	}

	return nil
}

// GetAccount retrieves an account by ID
func (s *ACMEStorage) GetAccount(id string) (*Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	account, ok := s.accounts[id]
	if !ok {
		return nil, fmt.Errorf("account not found: %s", id)
	}

	return account, nil
}

// FindAccountByKey finds an account by public key
func (s *ACMEStorage) FindAccountByKey(key []byte) (*Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Hash the key for comparison
	keyHash := sha256.Sum256(key)
	keyHashStr := base64.RawURLEncoding.EncodeToString(keyHash[:])

	for _, account := range s.accounts {
		if account.Key != nil {
			// Convert the account key to bytes for comparison
			var keyBytes []byte
			var err error

			switch k := account.Key.(type) {
			case *rsa.PublicKey:
				keyBytes, err = x509.MarshalPKIXPublicKey(k)
			case *ecdsa.PublicKey:
				keyBytes, err = x509.MarshalPKIXPublicKey(k)
			default:
				continue // Skip unknown key types
			}

			if err != nil {
				continue // Skip if we can't marshal the key
			}

			// Hash the account key
			accountKeyHash := sha256.Sum256(keyBytes)
			accountKeyHashStr := base64.RawURLEncoding.EncodeToString(accountKeyHash[:])

			// Compare the key hashes
			if keyHashStr == accountKeyHashStr {
				return account, nil
			}
		}
	}

	return nil, fmt.Errorf("account not found for key")
}

// SaveOrder saves an order to disk
func (s *ACMEStorage) SaveOrder(order *Order) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store in memory
	s.orders[order.ID] = order

	// Store on disk
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	filePath := filepath.Join(s.basePath, "orders", order.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write order file: %w", err)
	}

	return nil
}

// GetOrder retrieves an order by ID
func (s *ACMEStorage) GetOrder(id string) (*Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	order, ok := s.orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	return order, nil
}

// GetOrdersByAccount retrieves all orders for an account
func (s *ACMEStorage) GetOrdersByAccount(accountID string) ([]*Order, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var orders []*Order
	for _, order := range s.orders {
		if order.AccountID == accountID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// SaveAuthorization saves an authorization to disk
func (s *ACMEStorage) SaveAuthorization(authz *Authorization) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store in memory
	s.authzs[authz.ID] = authz

	// Store on disk
	data, err := json.Marshal(authz)
	if err != nil {
		return fmt.Errorf("failed to marshal authorization: %w", err)
	}

	filePath := filepath.Join(s.basePath, "authz", authz.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write authorization file: %w", err)
	}

	return nil
}

// GetAuthorization retrieves an authorization by ID
func (s *ACMEStorage) GetAuthorization(id string) (*Authorization, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	authz, ok := s.authzs[id]
	if !ok {
		return nil, fmt.Errorf("authorization not found: %s", id)
	}

	return authz, nil
}

// GetAuthorizationsByOrder retrieves all authorizations for an order
func (s *ACMEStorage) GetAuthorizationsByOrder(orderID string) ([]*Authorization, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var authzs []*Authorization
	for _, authz := range s.authzs {
		if authz.OrderID == orderID {
			authzs = append(authzs, authz)
		}
	}

	return authzs, nil
}

// SaveChallenge saves a challenge to disk
func (s *ACMEStorage) SaveChallenge(challenge *Challenge) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Store in memory
	s.challenges[challenge.ID] = challenge

	// Store on disk
	data, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("failed to marshal challenge: %w", err)
	}

	filePath := filepath.Join(s.basePath, "challenges", challenge.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write challenge file: %w", err)
	}

	return nil
}

// GetChallenge retrieves a challenge by ID
func (s *ACMEStorage) GetChallenge(id string) (*Challenge, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	challenge, ok := s.challenges[id]
	if !ok {
		return nil, fmt.Errorf("challenge not found: %s", id)
	}

	return challenge, nil
}

// GetChallengesByAuthorization retrieves all challenges for an authorization
func (s *ACMEStorage) GetChallengesByAuthorization(authzID string) ([]*Challenge, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var challenges []*Challenge
	for _, challenge := range s.challenges {
		if challenge.AuthorizationID == authzID {
			challenges = append(challenges, challenge)
		}
	}

	return challenges, nil
}

// CleanupExpired removes expired orders, authorizations, and challenges
func (s *ACMEStorage) CleanupExpired() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()

	// Clean up expired orders
	for id, order := range s.orders {
		if order.Expires.Before(now) {
			// Remove from memory
			delete(s.orders, id)

			// Remove from disk
			filePath := filepath.Join(s.basePath, "orders", id+".json")
			os.Remove(filePath)
		}
	}

	// Clean up expired authorizations
	for id, authz := range s.authzs {
		if authz.Expires.Before(now) {
			// Remove from memory
			delete(s.authzs, id)

			// Remove from disk
			filePath := filepath.Join(s.basePath, "authz", id+".json")
			os.Remove(filePath)
		}
	}

	return nil
}
