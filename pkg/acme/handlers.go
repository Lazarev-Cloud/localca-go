package acme

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// handleNewAccount handles ACME account creation
func (s *ACMEServer) handleNewAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := readRequestBody(r)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse JWS
	jws, err := ParseJWS(body)
	if err != nil {
		http.Error(w, "Invalid JWS", http.StatusBadRequest)
		return
	}

	// Get nonce from header
	nonce := r.Header.Get("Replay-Nonce")
	if nonce == "" {
		http.Error(w, "Missing nonce", http.StatusBadRequest)
		return
	}

	// Verify nonce
	if !s.validateNonce(nonce) {
		http.Error(w, "Invalid nonce", http.StatusBadRequest)
		return
	}

	// Verify JWS
	payload, pubKey, err := VerifyJWS(jws, nonce, r.URL.String())
	if err != nil {
		http.Error(w, "Invalid JWS signature", http.StatusBadRequest)
		return
	}

	// Parse account request
	var accountReq struct {
		Contact              []string `json:"contact"`
		TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
		OnlyReturnExisting   bool     `json:"onlyReturnExisting"`
	}

	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &accountReq); err != nil {
			http.Error(w, "Invalid account request", http.StatusBadRequest)
			return
		}
	}

	// Check if account already exists
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Printf("Failed to marshal public key: %v", err)
		http.Error(w, "Invalid public key", http.StatusBadRequest)
		return
	}

	existingAccount, err := s.acmeStorage.FindAccountByKey(pubKeyBytes)
	if err == nil && existingAccount != nil {
		// Account exists, return it
		baseURL := fmt.Sprintf("%s://%s", schemeFromRequest(r), r.Host)
		accountURL := fmt.Sprintf("%s/acme/account/%s", baseURL, existingAccount.ID)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", accountURL)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  existingAccount.Status,
			"contact": existingAccount.Contact,
			"orders":  fmt.Sprintf("%s/acme/orders/%s", baseURL, existingAccount.ID),
		})
		return
	}

	// If onlyReturnExisting is true and account doesn't exist, return error
	if accountReq.OnlyReturnExisting {
		http.Error(w, "Account does not exist", http.StatusBadRequest)
		return
	}

	// Create new account
	account := &Account{
		ID:        generateID(),
		Key:       pubKey,
		Contact:   accountReq.Contact,
		Status:    AccountStatusValid,
		CreatedAt: time.Now(),
	}

	// Save account
	if err := s.acmeStorage.SaveAccount(account); err != nil {
		log.Printf("Failed to save account: %v", err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	// Return account
	baseURL := fmt.Sprintf("%s://%s", schemeFromRequest(r), r.Host)
	accountURL := fmt.Sprintf("%s/acme/account/%s", baseURL, account.ID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", accountURL)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  account.Status,
		"contact": account.Contact,
		"orders":  fmt.Sprintf("%s/acme/orders/%s", baseURL, account.ID),
	})
}

// handleNewOrder handles ACME order creation
func (s *ACMEServer) handleNewOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := readRequestBody(r)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse JWS
	jws, err := ParseJWS(body)
	if err != nil {
		http.Error(w, "Invalid JWS", http.StatusBadRequest)
		return
	}

	// Get nonce from header
	nonce := r.Header.Get("Replay-Nonce")
	if nonce == "" {
		http.Error(w, "Missing nonce", http.StatusBadRequest)
		return
	}

	// Verify nonce
	if !s.validateNonce(nonce) {
		http.Error(w, "Invalid nonce", http.StatusBadRequest)
		return
	}

	// Verify JWS and get account
	payload, pubKey, err := VerifyJWS(jws, nonce, r.URL.String())
	if err != nil {
		http.Error(w, "Invalid JWS signature", http.StatusBadRequest)
		return
	}

	// Find account
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Printf("Failed to marshal public key: %v", err)
		http.Error(w, "Invalid public key", http.StatusBadRequest)
		return
	}

	account, err := s.acmeStorage.FindAccountByKey(pubKeyBytes)
	if err != nil || account == nil {
		http.Error(w, "Account not found", http.StatusBadRequest)
		return
	}

	// Parse order request
	var orderReq struct {
		Identifiers []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"identifiers"`
		NotBefore *time.Time `json:"notBefore,omitempty"`
		NotAfter  *time.Time `json:"notAfter,omitempty"`
	}

	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &orderReq); err != nil {
			http.Error(w, "Invalid order request", http.StatusBadRequest)
			return
		}
	}

	// Create order
	order := &Order{
		ID:          generateID(),
		AccountID:   account.ID,
		Status:      OrderStatusPending,
		Identifiers: make([]Identifier, len(orderReq.Identifiers)),
		Expires:     time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	// Copy identifiers
	for i, id := range orderReq.Identifiers {
		order.Identifiers[i] = Identifier{
			Type:  id.Type,
			Value: id.Value,
		}
	}

	// Create authorizations for each identifier
	baseURL := fmt.Sprintf("%s://%s", schemeFromRequest(r), r.Host)
	for _, identifier := range order.Identifiers {
		authz := &Authorization{
			ID:         generateID(),
			OrderID:    order.ID,
			Identifier: identifier,
			Status:     AuthzStatusPending,
			Expires:    time.Now().Add(24 * time.Hour),
			CreatedAt:  time.Now(),
		}

		// Create HTTP-01 challenge
		challenge := &Challenge{
			ID:              generateID(),
			AuthorizationID: authz.ID,
			Type:            ChallengeTypeHTTP01,
			Status:          ChallengeStatusPending,
			Token:           generateToken(),
			URL:             fmt.Sprintf("%s/acme/challenge/%s", baseURL, generateID()),
			CreatedAt:       time.Now(),
		}

		authz.Challenges = []*Challenge{challenge}
		authz.Status = AuthzStatusPending
		order.Authorizations = append(order.Authorizations, fmt.Sprintf("%s/acme/authz/%s", baseURL, authz.ID))

		// Save authorization and challenges
		if err := s.acmeStorage.SaveAuthorization(authz); err != nil {
			log.Printf("Failed to save authorization: %v", err)
			http.Error(w, "Failed to create authorization", http.StatusInternalServerError)
			return
		}

		for _, challenge := range authz.Challenges {
			if err := s.acmeStorage.SaveChallenge(challenge); err != nil {
				log.Printf("Failed to save challenge: %v", err)
				http.Error(w, "Failed to create challenge", http.StatusInternalServerError)
				return
			}
		}
	}

	// Set finalize URL
	order.FinalizeURL = fmt.Sprintf("%s/acme/finalize/%s", baseURL, order.ID)

	// Save order
	if err := s.acmeStorage.SaveOrder(order); err != nil {
		log.Printf("Failed to save order: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Return order
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("%s/acme/order/%s", baseURL, order.ID))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         order.Status,
		"expires":        order.Expires.Format(time.RFC3339),
		"identifiers":    order.Identifiers,
		"authorizations": order.Authorizations,
		"finalize":       order.FinalizeURL,
	})
}

// handleChallenge handles ACME challenge validation
func (s *ACMEServer) handleChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract challenge ID from URL
	challengeID := strings.TrimPrefix(r.URL.Path, "/acme/challenge/")

	// Get challenge
	challenge, err := s.acmeStorage.GetChallenge(challengeID)
	if err != nil {
		http.Error(w, "Challenge not found", http.StatusNotFound)
		return
	}

	// Read request body
	body, err := readRequestBody(r)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse JWS
	jws, err := ParseJWS(body)
	if err != nil {
		http.Error(w, "Invalid JWS", http.StatusBadRequest)
		return
	}

	// Get nonce from header
	nonce := r.Header.Get("Replay-Nonce")
	if nonce == "" {
		http.Error(w, "Missing nonce", http.StatusBadRequest)
		return
	}

	// Verify nonce
	if !s.validateNonce(nonce) {
		http.Error(w, "Invalid nonce", http.StatusBadRequest)
		return
	}

	// Verify JWS
	_, pubKey, err := VerifyJWS(jws, nonce, r.URL.String())
	if err != nil {
		http.Error(w, "Invalid JWS signature", http.StatusBadRequest)
		return
	}

	// Find account
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Printf("Failed to marshal public key: %v", err)
		http.Error(w, "Invalid public key", http.StatusBadRequest)
		return
	}

	account, err := s.acmeStorage.FindAccountByKey(pubKeyBytes)
	if err != nil || account == nil {
		http.Error(w, "Account not found", http.StatusBadRequest)
		return
	}

	// Update challenge status to processing
	challenge.Status = ChallengeStatusProcessing
	if err := s.acmeStorage.SaveChallenge(challenge); err != nil {
		log.Printf("Failed to update challenge: %v", err)
		http.Error(w, "Failed to update challenge", http.StatusInternalServerError)
		return
	}

	// Validate challenge
	if s.validateHTTP01Challenge(challenge) {
		challenge.Status = ChallengeStatusValid
		challenge.Validated = time.Now()
	} else {
		challenge.Status = ChallengeStatusInvalid
		challenge.Error = &ProblemDetails{
			Type:   "urn:ietf:params:acme:error:unauthorized",
			Detail: "Challenge validation failed",
		}
	}

	if err := s.acmeStorage.SaveChallenge(challenge); err != nil {
		log.Printf("Failed to update challenge: %v", err)
		http.Error(w, "Failed to update challenge", http.StatusInternalServerError)
		return
	}

	// Return challenge
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(challenge)
}

// validateHTTP01Challenge validates an HTTP-01 challenge
func (s *ACMEServer) validateHTTP01Challenge(challenge *Challenge) bool {
	// Get authorization
	authz, err := s.acmeStorage.GetAuthorization(challenge.AuthorizationID)
	if err != nil {
		log.Printf("Failed to get authorization: %v", err)
		return false
	}

	// For now, just return true for testing
	// In a real implementation, you would make an HTTP request to validate
	log.Printf("Validating HTTP-01 challenge for %s", authz.Identifier.Value)
	return true
}

// handleFinalize handles ACME order finalization
func (s *ACMEServer) handleFinalize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract order ID from URL
	certID := strings.TrimPrefix(r.URL.Path, "/acme/finalize/")

	// Get order
	order, err := s.acmeStorage.GetOrder(certID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Read request body
	body, err := readRequestBody(r)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse JWS
	jws, err := ParseJWS(body)
	if err != nil {
		http.Error(w, "Invalid JWS", http.StatusBadRequest)
		return
	}

	// Get nonce from header
	nonce := r.Header.Get("Replay-Nonce")
	if nonce == "" {
		http.Error(w, "Missing nonce", http.StatusBadRequest)
		return
	}

	// Verify nonce
	if !s.validateNonce(nonce) {
		http.Error(w, "Invalid nonce", http.StatusBadRequest)
		return
	}

	// Verify JWS
	_, pubKey, err := VerifyJWS(jws, nonce, r.URL.String())
	if err != nil {
		http.Error(w, "Invalid JWS signature", http.StatusBadRequest)
		return
	}

	// Find account
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Printf("Failed to marshal public key: %v", err)
		http.Error(w, "Invalid public key", http.StatusBadRequest)
		return
	}

	account, err := s.acmeStorage.FindAccountByKey(pubKeyBytes)
	if err != nil || account == nil {
		http.Error(w, "Account not found", http.StatusBadRequest)
		return
	}

	// Create certificate using the certificate service
	certName := fmt.Sprintf("acme-%s", order.ID)
	domains := make([]string, len(order.Identifiers))
	for i, identifier := range order.Identifiers {
		domains[i] = identifier.Value
	}

	// Issue certificate
	err = s.certSvc.CreateServerCertificate(certName, domains)
	if err != nil {
		log.Printf("Failed to issue certificate: %v", err)
		http.Error(w, "Failed to issue certificate", http.StatusInternalServerError)
		return
	}

	// Read certificate file
	certPath := s.storage.GetCertificatePath(certName)
	certData, err := os.ReadFile(certPath)
	if err != nil {
		log.Printf("Failed to read certificate: %v", err)
		http.Error(w, "Failed to read certificate", http.StatusInternalServerError)
		return
	}

	// Update order status
	order.Status = OrderStatusValid
	order.CertificateURL = fmt.Sprintf("%s://%s/acme/certificate/%s", schemeFromRequest(r), r.Host, order.ID)
	if err := s.acmeStorage.SaveOrder(order); err != nil {
		log.Printf("Failed to update order: %v", err)
	}

	// Return certificate
	w.Header().Set("Content-Type", "application/pem-certificate-chain")
	w.WriteHeader(http.StatusOK)
	w.Write(certData)
}

// readRequestBody reads the request body
func readRequestBody(r *http.Request) ([]byte, error) {
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	return body, err
}
