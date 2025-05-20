package acme

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// Order represents an ACME order
type Order struct {
	ID             string
	AccountID      string
	Status         string
	Expires        time.Time
	Identifiers    []Identifier
	Authorizations []string
	FinalizeURL    string
	CertificateURL string
	CSR            []byte
	NotBefore      time.Time
	NotAfter       time.Time
	Error          *ProblemDetails
	CreatedAt      time.Time
}

// Identifier represents an ACME identifier
type Identifier struct {
	Type  string
	Value string
}

// Authorization represents an ACME authorization
type Authorization struct {
	ID         string
	Identifier Identifier
	Status     string
	Expires    time.Time
	Challenges []*Challenge
	Wildcard   bool
	OrderID    string
	Error      *ProblemDetails
	CreatedAt  time.Time
}

// Challenge represents an ACME challenge
type Challenge struct {
	ID               string
	Type             string
	Status           string
	URL              string
	Token            string
	Validated        time.Time
	Error            *ProblemDetails
	AuthorizationID  string
	KeyAuthorization string
	CreatedAt        time.Time
}

// ProblemDetails represents an ACME error
type ProblemDetails struct {
	Type        string           `json:"type"`
	Detail      string           `json:"detail"`
	Status      int              `json:"status,omitempty"`
	Instance    string           `json:"instance,omitempty"`
	Subproblems []ProblemDetails `json:"subproblems,omitempty"`
}

// OrderStatus constants
const (
	OrderStatusPending    = "pending"
	OrderStatusReady      = "ready"
	OrderStatusProcessing = "processing"
	OrderStatusValid      = "valid"
	OrderStatusInvalid    = "invalid"
)

// AuthorizationStatus constants
const (
	AuthzStatusPending     = "pending"
	AuthzStatusValid       = "valid"
	AuthzStatusInvalid     = "invalid"
	AuthzStatusDeactivated = "deactivated"
	AuthzStatusExpired     = "expired"
	AuthzStatusRevoked     = "revoked"
)

// ChallengeStatus constants
const (
	ChallengeStatusPending    = "pending"
	ChallengeStatusProcessing = "processing"
	ChallengeStatusValid      = "valid"
	ChallengeStatusInvalid    = "invalid"
)

// ChallengeType constants
const (
	ChallengeTypeHTTP01 = "http-01"
	ChallengeTypeDNS01  = "dns-01"
)

// AccountStatus constants
const (
	AccountStatusValid       = "valid"
	AccountStatusDeactivated = "deactivated"
	AccountStatusRevoked     = "revoked"
)

// NewOrder creates a new ACME order
func NewOrder(accountID string, identifiers []Identifier, notBefore, notAfter time.Time) *Order {
	now := time.Now()
	return &Order{
		ID:          generateID(),
		AccountID:   accountID,
		Status:      OrderStatusPending,
		Expires:     now.Add(7 * 24 * time.Hour), // 7 days
		Identifiers: identifiers,
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		CreatedAt:   now,
	}
}

// NewAuthorization creates a new ACME authorization
func NewAuthorization(orderID string, identifier Identifier, wildcard bool) *Authorization {
	now := time.Now()
	return &Authorization{
		ID:         generateID(),
		Identifier: identifier,
		Status:     AuthzStatusPending,
		Expires:    now.Add(7 * 24 * time.Hour), // 7 days
		Wildcard:   wildcard,
		OrderID:    orderID,
		CreatedAt:  now,
	}
}

// NewChallenge creates a new ACME challenge
func NewChallenge(authzID string, challengeType string) *Challenge {
	now := time.Now()
	return &Challenge{
		ID:              generateID(),
		Type:            challengeType,
		Status:          ChallengeStatusPending,
		Token:           generateToken(),
		AuthorizationID: authzID,
		CreatedAt:       now,
	}
}

// generateID generates a random ID
func generateID() string {
	return generateToken()[:16]
}

// generateToken generates a random token
func generateToken() string {
	// Generate a random 32-byte token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		// Fall back to a fixed token in case of error
		return "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	}
	return base64URLEncode(tokenBytes)
}

// base64URLEncode encodes bytes using base64URL encoding without padding
func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
