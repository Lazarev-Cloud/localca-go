package acme

import (
	"crypto/rsa"
	"os"
	"testing"
	"time"
)

func TestNewACMEStorage(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "acme-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewACMEStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create ACME storage: %v", err)
	}

	// Check if directories were created
	dirs := []string{"acme", "acme/accounts", "acme/orders", "acme/authz", "acme/challenges"}
	for _, dir := range dirs {
		path := tempDir + "/" + dir
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", path)
		}
	}

	// Check if maps were initialized
	if storage.accounts == nil {
		t.Error("Accounts map was not initialized")
	}
	if storage.orders == nil {
		t.Error("Orders map was not initialized")
	}
	if storage.authzs == nil {
		t.Error("Authorizations map was not initialized")
	}
	if storage.challenges == nil {
		t.Error("Challenges map was not initialized")
	}
}

func TestSaveAndGetAccount(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "acme-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewACMEStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create ACME storage: %v", err)
	}

	// Create test account
	account := &Account{
		ID:        "test-account",
		Key:       &rsa.PublicKey{},
		Contact:   []string{"mailto:test@example.com"},
		Status:    AccountStatusValid,
		CreatedAt: time.Now(),
	}

	// Save account
	if err := storage.SaveAccount(account); err != nil {
		t.Fatalf("Failed to save account: %v", err)
	}

	// Get account
	retrievedAccount, err := storage.GetAccount("test-account")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	// Check account
	if retrievedAccount.ID != account.ID {
		t.Errorf("Expected account ID %s, got %s", account.ID, retrievedAccount.ID)
	}
	if retrievedAccount.Status != account.Status {
		t.Errorf("Expected account status %s, got %s", account.Status, retrievedAccount.Status)
	}
	if len(retrievedAccount.Contact) != len(account.Contact) {
		t.Errorf("Expected %d contacts, got %d", len(account.Contact), len(retrievedAccount.Contact))
	} else if retrievedAccount.Contact[0] != account.Contact[0] {
		t.Errorf("Expected contact %s, got %s", account.Contact[0], retrievedAccount.Contact[0])
	}
}

func TestSaveAndGetOrder(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "acme-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewACMEStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create ACME storage: %v", err)
	}

	// Create test order
	order := &Order{
		ID:        "test-order",
		AccountID: "test-account",
		Status:    OrderStatusPending,
		Expires:   time.Now().Add(24 * time.Hour),
		Identifiers: []Identifier{
			{Type: "dns", Value: "example.com"},
		},
		CreatedAt: time.Now(),
	}

	// Save order
	if err := storage.SaveOrder(order); err != nil {
		t.Fatalf("Failed to save order: %v", err)
	}

	// Get order
	retrievedOrder, err := storage.GetOrder("test-order")
	if err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}

	// Check order
	if retrievedOrder.ID != order.ID {
		t.Errorf("Expected order ID %s, got %s", order.ID, retrievedOrder.ID)
	}
	if retrievedOrder.AccountID != order.AccountID {
		t.Errorf("Expected account ID %s, got %s", order.AccountID, retrievedOrder.AccountID)
	}
	if retrievedOrder.Status != order.Status {
		t.Errorf("Expected order status %s, got %s", order.Status, retrievedOrder.Status)
	}
	if len(retrievedOrder.Identifiers) != len(order.Identifiers) {
		t.Errorf("Expected %d identifiers, got %d", len(order.Identifiers), len(retrievedOrder.Identifiers))
	} else {
		if retrievedOrder.Identifiers[0].Type != order.Identifiers[0].Type {
			t.Errorf("Expected identifier type %s, got %s", order.Identifiers[0].Type, retrievedOrder.Identifiers[0].Type)
		}
		if retrievedOrder.Identifiers[0].Value != order.Identifiers[0].Value {
			t.Errorf("Expected identifier value %s, got %s", order.Identifiers[0].Value, retrievedOrder.Identifiers[0].Value)
		}
	}
}

func TestGetOrdersByAccount(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "acme-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewACMEStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create ACME storage: %v", err)
	}

	// Create test orders
	orders := []*Order{
		{
			ID:        "test-order-1",
			AccountID: "test-account",
			Status:    OrderStatusPending,
			CreatedAt: time.Now(),
		},
		{
			ID:        "test-order-2",
			AccountID: "test-account",
			Status:    OrderStatusReady,
			CreatedAt: time.Now(),
		},
		{
			ID:        "test-order-3",
			AccountID: "other-account",
			Status:    OrderStatusPending,
			CreatedAt: time.Now(),
		},
	}

	// Save orders
	for _, order := range orders {
		if err := storage.SaveOrder(order); err != nil {
			t.Fatalf("Failed to save order: %v", err)
		}
	}

	// Get orders by account
	retrievedOrders, err := storage.GetOrdersByAccount("test-account")
	if err != nil {
		t.Fatalf("Failed to get orders by account: %v", err)
	}

	// Check orders
	if len(retrievedOrders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(retrievedOrders))
	}

	// Check that all retrieved orders belong to the account
	for _, order := range retrievedOrders {
		if order.AccountID != "test-account" {
			t.Errorf("Retrieved order %s belongs to account %s, not test-account", order.ID, order.AccountID)
		}
	}
}

func TestSaveAndGetAuthorization(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "acme-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewACMEStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create ACME storage: %v", err)
	}

	// Create test authorization
	authz := &Authorization{
		ID:         "test-authz",
		Identifier: Identifier{Type: "dns", Value: "example.com"},
		Status:     AuthzStatusPending,
		Expires:    time.Now().Add(24 * time.Hour),
		OrderID:    "test-order",
		CreatedAt:  time.Now(),
	}

	// Save authorization
	if err := storage.SaveAuthorization(authz); err != nil {
		t.Fatalf("Failed to save authorization: %v", err)
	}

	// Get authorization
	retrievedAuthz, err := storage.GetAuthorization("test-authz")
	if err != nil {
		t.Fatalf("Failed to get authorization: %v", err)
	}

	// Check authorization
	if retrievedAuthz.ID != authz.ID {
		t.Errorf("Expected authorization ID %s, got %s", authz.ID, retrievedAuthz.ID)
	}
	if retrievedAuthz.Status != authz.Status {
		t.Errorf("Expected authorization status %s, got %s", authz.Status, retrievedAuthz.Status)
	}
	if retrievedAuthz.OrderID != authz.OrderID {
		t.Errorf("Expected order ID %s, got %s", authz.OrderID, retrievedAuthz.OrderID)
	}
	if retrievedAuthz.Identifier.Type != authz.Identifier.Type {
		t.Errorf("Expected identifier type %s, got %s", authz.Identifier.Type, retrievedAuthz.Identifier.Type)
	}
	if retrievedAuthz.Identifier.Value != authz.Identifier.Value {
		t.Errorf("Expected identifier value %s, got %s", authz.Identifier.Value, retrievedAuthz.Identifier.Value)
	}
}

func TestCleanupExpired(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "acme-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewACMEStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create ACME storage: %v", err)
	}

	// Create test orders
	orders := []*Order{
		{
			ID:        "expired-order",
			AccountID: "test-account",
			Status:    OrderStatusPending,
			Expires:   time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        "valid-order",
			AccountID: "test-account",
			Status:    OrderStatusPending,
			Expires:   time.Now().Add(24 * time.Hour), // Valid
			CreatedAt: time.Now(),
		},
	}

	// Save orders
	for _, order := range orders {
		if err := storage.SaveOrder(order); err != nil {
			t.Fatalf("Failed to save order: %v", err)
		}
	}

	// Create test authorizations
	authzs := []*Authorization{
		{
			ID:         "expired-authz",
			Identifier: Identifier{Type: "dns", Value: "example.com"},
			Status:     AuthzStatusPending,
			Expires:    time.Now().Add(-1 * time.Hour), // Expired
			OrderID:    "expired-order",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
		},
		{
			ID:         "valid-authz",
			Identifier: Identifier{Type: "dns", Value: "example.org"},
			Status:     AuthzStatusPending,
			Expires:    time.Now().Add(24 * time.Hour), // Valid
			OrderID:    "valid-order",
			CreatedAt:  time.Now(),
		},
	}

	// Save authorizations
	for _, authz := range authzs {
		if err := storage.SaveAuthorization(authz); err != nil {
			t.Fatalf("Failed to save authorization: %v", err)
		}
	}

	// Cleanup expired records
	if err := storage.CleanupExpired(); err != nil {
		t.Fatalf("Failed to cleanup expired records: %v", err)
	}

	// Check that expired order was removed
	_, err = storage.GetOrder("expired-order")
	if err == nil {
		t.Error("Expired order was not removed")
	}

	// Check that valid order still exists
	_, err = storage.GetOrder("valid-order")
	if err != nil {
		t.Errorf("Valid order was removed: %v", err)
	}

	// Check that expired authorization was removed
	_, err = storage.GetAuthorization("expired-authz")
	if err == nil {
		t.Error("Expired authorization was not removed")
	}

	// Check that valid authorization still exists
	_, err = storage.GetAuthorization("valid-authz")
	if err != nil {
		t.Errorf("Valid authorization was removed: %v", err)
	}
}
