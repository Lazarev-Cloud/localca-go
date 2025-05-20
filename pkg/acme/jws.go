package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
)

// JWS represents a JSON Web Signature
type JWS struct {
	Protected string `json:"protected"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

// JWSHeader represents the header of a JWS
type JWSHeader struct {
	Alg   string `json:"alg"`
	Kid   string `json:"kid,omitempty"`
	Nonce string `json:"nonce"`
	URL   string `json:"url"`
	Jwk   *JWK   `json:"jwk,omitempty"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv,omitempty"` // For EC keys
	X   string `json:"x,omitempty"`   // For EC keys
	Y   string `json:"y,omitempty"`   // For EC keys
	N   string `json:"n,omitempty"`   // For RSA keys
	E   string `json:"e,omitempty"`   // For RSA keys
}

// ParseJWS parses a JWS from JSON
func ParseJWS(data []byte) (*JWS, error) {
	var jws JWS
	if err := json.Unmarshal(data, &jws); err != nil {
		return nil, fmt.Errorf("failed to parse JWS: %w", err)
	}
	return &jws, nil
}

// VerifyJWS verifies a JWS signature
func VerifyJWS(jws *JWS, expectedNonce string, expectedURL string) ([]byte, crypto.PublicKey, error) {
	// Decode protected header
	headerJSON, err := base64.RawURLEncoding.DecodeString(jws.Protected)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode JWS protected header: %w", err)
	}

	// Parse header
	var header JWSHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWS header: %w", err)
	}

	// Verify nonce if expected
	if expectedNonce != "" && header.Nonce != expectedNonce {
		return nil, nil, fmt.Errorf("invalid nonce")
	}

	// Verify URL if expected
	if expectedURL != "" && header.URL != expectedURL {
		return nil, nil, fmt.Errorf("invalid URL")
	}

	// Get public key
	var pubKey crypto.PublicKey
	if header.Jwk != nil {
		// Key is in the JWK
		pubKey, err = jwkToPublicKey(header.Jwk)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse JWK: %w", err)
		}
	} else if header.Kid != "" {
		// Key is referenced by KID
		// This would typically involve looking up the key in a database
		return nil, nil, fmt.Errorf("KID-based key lookup not implemented")
	} else {
		return nil, nil, fmt.Errorf("no key provided in JWS")
	}

	// Verify signature
	signatureInput := jws.Protected + "." + jws.Payload
	signatureBytes, err := base64.RawURLEncoding.DecodeString(jws.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode JWS signature: %w", err)
	}

	// Hash the signature input
	hash := sha256.Sum256([]byte(signatureInput))

	// Verify signature based on algorithm
	switch header.Alg {
	case "RS256":
		rsaKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, nil, fmt.Errorf("key is not an RSA key")
		}
		if err := rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, hash[:], signatureBytes); err != nil {
			return nil, nil, fmt.Errorf("invalid signature: %w", err)
		}
	case "ES256":
		ecKey, ok := pubKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, nil, fmt.Errorf("key is not an ECDSA key")
		}
		// ECDSA signature is in the format r || s
		if len(signatureBytes) != 64 {
			return nil, nil, fmt.Errorf("invalid ECDSA signature length")
		}
		r := new(big.Int).SetBytes(signatureBytes[:32])
		s := new(big.Int).SetBytes(signatureBytes[32:])
		if !ecdsa.Verify(ecKey, hash[:], r, s) {
			return nil, nil, fmt.Errorf("invalid ECDSA signature")
		}
	default:
		return nil, nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(jws.Payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode JWS payload: %w", err)
	}

	return payload, pubKey, nil
}

// jwkToPublicKey converts a JWK to a public key
func jwkToPublicKey(jwk *JWK) (crypto.PublicKey, error) {
	switch jwk.Kty {
	case "RSA":
		// Parse RSA key
		if jwk.N == "" || jwk.E == "" {
			return nil, fmt.Errorf("incomplete RSA key")
		}
		
		// Decode modulus
		n, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			return nil, fmt.Errorf("failed to decode RSA modulus: %w", err)
		}
		
		// Decode exponent
		e, err := base64.RawURLEncoding.DecodeString(jwk.E)
		if err != nil {
			return nil, fmt.Errorf("failed to decode RSA exponent: %w", err)
		}
		
		// Convert exponent to int
		var exponent int
		for i := 0; i < len(e); i++ {
			exponent = exponent<<8 + int(e[i])
		}
		
		return &rsa.PublicKey{
			N: new(big.Int).SetBytes(n),
			E: exponent,
		}, nil
		
	case "EC":
		// Parse EC key
		if jwk.Crv == "" || jwk.X == "" || jwk.Y == "" {
			return nil, fmt.Errorf("incomplete EC key")
		}
		
		// Only P-256 is supported for now
		if jwk.Crv != "P-256" {
			return nil, fmt.Errorf("unsupported curve: %s", jwk.Crv)
		}
		
		// Decode X and Y coordinates
		x, err := base64.RawURLEncoding.DecodeString(jwk.X)
		if err != nil {
			return nil, fmt.Errorf("failed to decode EC X coordinate: %w", err)
		}
		
		y, err := base64.RawURLEncoding.DecodeString(jwk.Y)
		if err != nil {
			return nil, fmt.Errorf("failed to decode EC Y coordinate: %w", err)
		}
		
		return &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(x),
			Y:     new(big.Int).SetBytes(y),
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}
} 