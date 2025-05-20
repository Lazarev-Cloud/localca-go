package acme

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
)

func TestParseJWS(t *testing.T) {
	// Create a test JWS
	testJWS := JWS{
		Protected: "eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoiX3Rlc3RfbiIsImUiOiJBUUFCIn0sIm5vbmNlIjoidGVzdC1ub25jZSIsInVybCI6Imh0dHBzOi8vZXhhbXBsZS5jb20vYWNtZS9uZXctYWNjb3VudCJ9",
		Payload:   "eyJ0ZXJtc09mU2VydmljZUFncmVlZCI6dHJ1ZSwiY29udGFjdCI6WyJtYWlsdG86YWRtaW5AZXhhbXBsZS5jb20iXX0",
		Signature: "test-signature",
	}

	// Marshal to JSON
	jwsJSON, err := json.Marshal(testJWS)
	if err != nil {
		t.Fatalf("Failed to marshal JWS: %v", err)
	}

	// Parse JWS
	parsedJWS, err := ParseJWS(jwsJSON)
	if err != nil {
		t.Fatalf("Failed to parse JWS: %v", err)
	}

	// Check fields
	if parsedJWS.Protected != testJWS.Protected {
		t.Errorf("Expected protected %s, got %s", testJWS.Protected, parsedJWS.Protected)
	}
	if parsedJWS.Payload != testJWS.Payload {
		t.Errorf("Expected payload %s, got %s", testJWS.Payload, parsedJWS.Payload)
	}
	if parsedJWS.Signature != testJWS.Signature {
		t.Errorf("Expected signature %s, got %s", testJWS.Signature, parsedJWS.Signature)
	}
}

func TestJwkToPublicKey_RSA(t *testing.T) {
	// Create RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Convert modulus and exponent to base64url
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())

	// Create JWK
	jwk := &JWK{
		Kty: "RSA",
		N:   n,
		E:   e,
	}

	// Convert JWK to public key
	parsedKey, err := jwkToPublicKey(jwk)
	if err != nil {
		t.Fatalf("Failed to parse JWK: %v", err)
	}

	// Check key type
	parsedRSAKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("Parsed key is not an RSA key")
	}

	// Check modulus
	if parsedRSAKey.N.Cmp(publicKey.N) != 0 {
		t.Error("Parsed key modulus does not match original")
	}

	// Check exponent
	if parsedRSAKey.E != publicKey.E {
		t.Errorf("Parsed key exponent %d does not match original %d", parsedRSAKey.E, publicKey.E)
	}
}

func TestJwkToPublicKey_EC(t *testing.T) {
	// Create EC key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate EC key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Convert coordinates to base64url
	x := base64.RawURLEncoding.EncodeToString(publicKey.X.Bytes())
	y := base64.RawURLEncoding.EncodeToString(publicKey.Y.Bytes())

	// Create JWK
	jwk := &JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   x,
		Y:   y,
	}

	// Convert JWK to public key
	parsedKey, err := jwkToPublicKey(jwk)
	if err != nil {
		t.Fatalf("Failed to parse JWK: %v", err)
	}

	// Check key type
	parsedECKey, ok := parsedKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Parsed key is not an EC key")
	}

	// Check curve
	if parsedECKey.Curve != elliptic.P256() {
		t.Error("Parsed key curve does not match original")
	}

	// Check coordinates
	if parsedECKey.X.Cmp(publicKey.X) != 0 {
		t.Error("Parsed key X coordinate does not match original")
	}
	if parsedECKey.Y.Cmp(publicKey.Y) != 0 {
		t.Error("Parsed key Y coordinate does not match original")
	}
}

func TestJwkToPublicKey_UnsupportedType(t *testing.T) {
	// Create JWK with unsupported type
	jwk := &JWK{
		Kty: "UNSUPPORTED",
	}

	// Try to convert JWK to public key
	_, err := jwkToPublicKey(jwk)
	if err == nil {
		t.Fatal("Expected error for unsupported key type")
	}
}

func TestVerifyJWS_InvalidNonce(t *testing.T) {
	// Create a test JWS
	testJWS := &JWS{
		Protected: "eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoiX3Rlc3RfbiIsImUiOiJBUUFCIn0sIm5vbmNlIjoidGVzdC1ub25jZSIsInVybCI6Imh0dHBzOi8vZXhhbXBsZS5jb20vYWNtZS9uZXctYWNjb3VudCJ9",
		Payload:   "eyJ0ZXJtc09mU2VydmljZUFncmVlZCI6dHJ1ZSwiY29udGFjdCI6WyJtYWlsdG86YWRtaW5AZXhhbXBsZS5jb20iXX0",
		Signature: "test-signature",
	}

	// Verify JWS with wrong nonce
	_, _, err := VerifyJWS(testJWS, "wrong-nonce", "")
	if err == nil {
		t.Fatal("Expected error for invalid nonce")
	}
}

func TestVerifyJWS_InvalidURL(t *testing.T) {
	// Create a test JWS
	testJWS := &JWS{
		Protected: "eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoiX3Rlc3RfbiIsImUiOiJBUUFCIn0sIm5vbmNlIjoidGVzdC1ub25jZSIsInVybCI6Imh0dHBzOi8vZXhhbXBsZS5jb20vYWNtZS9uZXctYWNjb3VudCJ9",
		Payload:   "eyJ0ZXJtc09mU2VydmljZUFncmVlZCI6dHJ1ZSwiY29udGFjdCI6WyJtYWlsdG86YWRtaW5AZXhhbXBsZS5jb20iXX0",
		Signature: "test-signature",
	}

	// Verify JWS with wrong URL
	_, _, err := VerifyJWS(testJWS, "test-nonce", "wrong-url")
	if err == nil {
		t.Fatal("Expected error for invalid URL")
	}
}

func TestCreateAndVerifyRSASignature(t *testing.T) {
	// Create RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Create test data
	data := []byte("test data")
	hash := sha256.Sum256(data)

	// Sign data
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, sha256.New(), hash[:])
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(&privateKey.PublicKey, sha256.New(), hash[:], signature)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
}

func TestCreateAndVerifyECDSASignature(t *testing.T) {
	// Create ECDSA key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	// Create test data
	data := []byte("test data")
	hash := sha256.Sum256(data)

	// Sign data
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	// Verify signature
	if !ecdsa.Verify(&privateKey.PublicKey, hash[:], r, s) {
		t.Fatal("Failed to verify ECDSA signature")
	}
}
