package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// JWTClaims represents the claims in a JSON Web Token.
type JWTClaims struct {
	Sub       string `json:"sub"`
	Name      string `json:"name"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

// JWTValidator validates and generates HS256-signed JWTs.
type JWTValidator struct {
	secret []byte
}

// NewJWTValidator creates a validator that signs tokens with the given secret.
func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{
		secret: []byte(secret),
	}
}

// Generate creates a signed JWT from the given claims.
func (j *JWTValidator) Generate(claims *JWTClaims) (string, error) {
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = time.Now().Add(24 * time.Hour).Unix()
	}
	if claims.IssuedAt == 0 {
		claims.IssuedAt = time.Now().Unix()
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := fmt.Sprintf("%s.%s", headerB64, claimsB64)
	signature := j.sign(signingInput)

	return fmt.Sprintf("%s.%s", signingInput, signature), nil
}

// Validate verifies the signature and expiry of a JWT and returns its claims.
func (j *JWTValidator) Validate(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		slog.Error("jwt validation failed", "reason", "invalid token format")
		return nil, fmt.Errorf("invalid token format")
	}

	signingInput := fmt.Sprintf("%s.%s", parts[0], parts[1])
	expectedSig := j.sign(signingInput)

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		slog.Error("jwt validation failed", "reason", "invalid signature")
		return nil, fmt.Errorf("invalid signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		slog.Error("jwt validation failed", "reason", "invalid claims encoding", "error", err)
		return nil, fmt.Errorf("invalid claims encoding: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		slog.Error("jwt validation failed", "reason", "invalid claims format", "error", err)
		return nil, fmt.Errorf("invalid claims format: %w", err)
	}

	if claims.ExpiresAt < time.Now().Unix() {
		slog.Error("jwt validation failed", "reason", "token expired")
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func (j *JWTValidator) sign(input string) string {
	mac := hmac.New(sha256.New, j.secret)
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// OIDCDiscovery represents an OpenID Connect discovery document.
type OIDCDiscovery struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

// OIDCDiscoveryDocument returns the OIDC discovery metadata for the given issuer.
func (j *JWTValidator) OIDCDiscoveryDocument(issuer string) *OIDCDiscovery {
	return &OIDCDiscovery{
		Issuer:                           issuer,
		AuthorizationEndpoint:            issuer + "/authorize",
		TokenEndpoint:                    issuer + "/token",
		ResponseTypesSupported:           []string{"code"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"HS256"},
	}
}
