package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
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

// JWSHeader represents the protected header of a JWS
type JWSHeader struct {
	Alg   string `json:"alg"`
	Kid   string `json:"kid,omitempty"`
	Jwk   *JWK   `json:"jwk,omitempty"`
	Nonce string `json:"nonce"`
	URL   string `json:"url"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	N   string `json:"n,omitempty"` // RSA modulus
	E   string `json:"e,omitempty"` // RSA public exponent
	Crv string `json:"crv,omitempty"` // EC curve
	X   string `json:"x,omitempty"` // EC x coordinate
	Y   string `json:"y,omitempty"` // EC y coordinate
}

// ParseJWS parses a JWS from JSON
func ParseJWS(data []byte) (*JWS, error) {
	var jws JWS
	if err := json.Unmarshal(data, &jws); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWS: %w", err)
	}
	return &jws, nil
}

// ParseJWSHeader parses the JWS protected header
func (j *JWS) ParseJWSHeader() (*JWSHeader, error) {
	// Decode base64url-encoded protected header
	headerBytes, err := base64.RawURLEncoding.DecodeString(j.Protected)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWS protected header: %w", err)
	}

	// Parse header
	var header JWSHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWS header: %w", err)
	}

	return &header, nil
}

// Verify verifies the JWS signature
func (j *JWS) Verify() ([]byte, error) {
	// Parse header
	header, err := j.ParseJWSHeader()
	if err != nil {
		return nil, err
	}

	// Get signing algorithm
	var publicKey crypto.PublicKey
	var hashFunc crypto.Hash

	// Check if key ID is provided
	if header.Kid != "" {
		// Key ID is provided, need to look up the key
		// This would typically be handled by the ACME service
		return nil, fmt.Errorf("key ID lookup not implemented")
	}

	// Check if JWK is provided
	if header.Jwk == nil {
		return nil, fmt.Errorf("JWK not provided")
	}

	// Parse JWK based on key type
	switch header.Jwk.Kty {
	case "RSA":
		// Parse RSA key
		if header.Jwk.N == "" || header.Jwk.E == "" {
			return nil, fmt.Errorf("invalid RSA key parameters")
		}

		// Decode modulus and exponent
		n, err := base64.RawURLEncoding.DecodeString(header.Jwk.N)
		if err != nil {
			return nil, fmt.Errorf("failed to decode RSA modulus: %w", err)
		}

		e, err := base64.RawURLEncoding.DecodeString(header.Jwk.E)
		if err != nil {
			return nil, fmt.Errorf("failed to decode RSA exponent: %w", err)
		}

		// Convert to big integers
		modulus := new(big.Int).SetBytes(n)
		
		// Convert exponent bytes to int
		var exponent int
		for i := 0; i < len(e); i++ {
			exponent = exponent*256 + int(e[i])
		}

		// Create RSA public key
		publicKey = &rsa.PublicKey{
			N: modulus,
			E: exponent,
		}

	case "EC":
		// Parse EC key
		if header.Jwk.Crv == "" || header.Jwk.X == "" || header.Jwk.Y == "" {
			return nil, fmt.Errorf("invalid EC key parameters")
		}

		// Only P-256 curve is supported
		if header.Jwk.Crv != "P-256" {
			return nil, fmt.Errorf("unsupported EC curve: %s", header.Jwk.Crv)
		}

		// Decode coordinates
		x, err := base64.RawURLEncoding.DecodeString(header.Jwk.X)
		if err != nil {
			return nil, fmt.Errorf("failed to decode EC x coordinate: %w", err)
		}

		y, err := base64.RawURLEncoding.DecodeString(header.Jwk.Y)
		if err != nil {
			return nil, fmt.Errorf("failed to decode EC y coordinate: %w", err)
		}

		// Create EC public key
		xInt := new(big.Int).SetBytes(x)
		yInt := new(big.Int).SetBytes(y)

		// This is a simplified example - in a real implementation, we would use the proper EC curve
		publicKey = &ecdsa.PublicKey{
			// Set up proper ECDSA curve parameters
			X: xInt,
			Y: yInt,
		}

	default:
		return nil, fmt.Errorf("unsupported key type: %s", header.Jwk.Kty)
	}

	// Check algorithm
	switch header.Alg {
	case "RS256":
		hashFunc = crypto.SHA256
	case "ES256":
		hashFunc = crypto.SHA256
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(j.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(j.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Compute message to be verified
	message := j.Protected + "." + j.Payload

	// Hash message
	h := sha256.New()
	h.Write([]byte(message))
	hashed := h.Sum(nil)

	// Verify signature
	switch key := publicKey.(type) {
	case *rsa.PublicKey:
		if header.Alg != "RS256" {
			return nil, fmt.Errorf("algorithm mismatch for RSA key")
		}
		err = rsa.VerifyPKCS1v15(key, hashFunc, hashed, signature)
		if err != nil {
			return nil, fmt.Errorf("RSA signature verification failed: %w", err)
		}
	case *ecdsa.PublicKey:
		if header.Alg != "ES256" {
			return nil, fmt.Errorf("algorithm mismatch for EC key")
		}
		// ECDSA signature verification is more complex and would involve converting
		// the signature to R and S components and using ecdsa.Verify
		return nil, fmt.Errorf("ECDSA signature verification not implemented")
	default:
		return nil, fmt.Errorf("unsupported public key type")
	}

	return payload, nil
}

// GetJWK extracts the JWK from the JWS
func (j *JWS) GetJWK() (*JWK, error) {
	header, err := j.ParseJWSHeader()
	if err != nil {
		return nil, err
	}
	
	if header.Jwk == nil {
		return nil, fmt.Errorf("JWK not found in JWS header")
	}
	
	return header.Jwk, nil
}

// GetKID extracts the Key ID from the JWS
func (j *JWS) GetKID() (string, error) {
	header, err := j.ParseJWSHeader()
	if err != nil {
		return "", err
	}
	
	return header.Kid, nil
}

// GetURL extracts the URL from the JWS
func (j *JWS) GetURL() (string, error) {
	header, err := j.ParseJWSHeader()
	if err != nil {
		return "", err
	}
	
	return header.URL, nil
}

// GetNonce extracts the nonce from the JWS
func (j *JWS) GetNonce() (string, error) {
	header, err := j.ParseJWSHeader()
	if err != nil {
		return "", err
	}
	
	return header.Nonce, nil
}

// SerializeJWK serializes a public key to JWK format
func SerializeJWK(publicKey crypto.PublicKey) (*JWK, error) {
	switch key := publicKey.(type) {
	case *rsa.PublicKey:
		// RSA key
		// Convert modulus to bytes
		n := key.N.Bytes()
		
		// Convert exponent to bytes
		var e []byte
		temp := key.E
		for temp > 0 {
			e = append([]byte{byte(temp & 0xFF)}, e...)
			temp >>= 8
		}
		if len(e) == 0 {
			e = []byte{0}
		}
		
		return &JWK{
			Kty: "RSA",
			N:   base64.RawURLEncoding.EncodeToString(n),
			E:   base64.RawURLEncoding.EncodeToString(e),
		}, nil
		
	case *ecdsa.PublicKey:
		// EC key
		// This is a simplified example - in a real implementation, we would use the proper EC curve
		x := key.X.Bytes()
		y := key.Y.Bytes()
		
		// Determine curve
		var crv string
		switch {
		case key.Curve.Params().BitSize == 256:
			crv = "P-256"
		case key.Curve.Params().BitSize == 384:
			crv = "P-384"
		case key.Curve.Params().BitSize == 521:
			crv = "P-521"
		default:
			return nil, fmt.Errorf("unsupported EC curve size: %d", key.Curve.Params().BitSize)
		}
		
		return &JWK{
			Kty: "EC",
			Crv: crv,
			X:   base64.RawURLEncoding.EncodeToString(x),
			Y:   base64.RawURLEncoding.EncodeToString(y),
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported public key type")
	}
}

// ThumbprintJWK computes the JWK thumbprint as per RFC7638
func ThumbprintJWK(jwk *JWK) (string, error) {
	var jsonData []byte
	var err error
	
	switch jwk.Kty {
	case "RSA":
		// RSA keys
		jsonData, err = json.Marshal(struct {
			E   string `json:"e"`
			Kty string `json:"kty"`
			N   string `json:"n"`
		}{
			E:   jwk.E,
			Kty: jwk.Kty,
			N:   jwk.N,
		})
		
	case "EC":
		// EC keys
		jsonData, err = json.Marshal(struct {
			Crv string `json:"crv"`
			Kty string `json:"kty"`
			X   string `json:"x"`
			Y   string `json:"y"`
		}{
			Crv: jwk.Crv,
			Kty: jwk.Kty,
			X:   jwk.X,
			Y:   jwk.Y,
		})
		
	default:
		return "", fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}
	
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWK: %w", err)
	}
	
	// Compute SHA-256 hash
	h := sha256.Sum256(jsonData)
	
	// Base64url encode
	return base64.RawURLEncoding.EncodeToString(h[:]), nil
}