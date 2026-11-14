// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	serviceJWKSCacheTTL = 5 * time.Minute
	serviceJWKSMaxBytes = 1 << 20
)

// JWKSValidator accepts only RS256 service JWTs signed by the configured
// issuer. It validates audience, expiry, not-before, and service scopes.
type JWKSValidator struct {
	issuer   string
	audience string
	url      string
	client   *http.Client
	mu       sync.Mutex
	keys     map[string]*rsa.PublicKey
	fetched  time.Time
}

type serviceJWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type serviceJWKS struct {
	Keys []serviceJWK `json:"keys"`
}

type serviceJWTHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

func isRS256Token(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	encoded, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var header serviceJWTHeader
	return json.Unmarshal(encoded, &header) == nil && header.Alg == "RS256"
}

func hasJWTClaimScope(claims *JWTClaims, expected string) bool {
	if claims == nil {
		return false
	}
	for _, scope := range strings.Fields(claims.Scope) {
		if scope == expected {
			return true
		}
	}
	for _, scope := range claims.Scopes {
		if strings.TrimSpace(scope) == expected {
			return true
		}
	}
	return false
}

type serviceJWTClaims struct {
	Issuer    string          `json:"iss"`
	Subject   string          `json:"sub"`
	ClientID  string          `json:"client_id"`
	Audience  json.RawMessage `json:"aud"`
	Expires   int64           `json:"exp"`
	NotBefore int64           `json:"nbf"`
	IssuedAt  int64           `json:"iat"`
	Scope     string          `json:"scope"`
	Scopes    []string        `json:"scopes"`
}

func NewJWKSValidator(issuer, audience, endpoint string) (*JWKSValidator, error) {
	issuer = strings.TrimSpace(issuer)
	audience = strings.TrimSpace(audience)
	endpoint = strings.TrimSpace(endpoint)
	issuerURL, issuerErr := url.Parse(issuer)
	endpointURL, endpointErr := url.Parse(endpoint)
	if issuer == "" || audience == "" || endpoint == "" || issuerErr != nil || endpointErr != nil ||
		issuerURL.Scheme != "https" || issuerURL.Host == "" || issuerURL.User != nil || issuerURL.Fragment != "" ||
		endpointURL.Scheme != "https" || endpointURL.Host == "" || endpointURL.User != nil || endpointURL.Fragment != "" {
		return nil, errors.New("service JWT issuer, audience, and HTTPS JWKS URL are required")
	}
	return &JWKSValidator{
		issuer:   issuer,
		audience: audience,
		url:      endpoint,
		client: &http.Client{
			Timeout: 5 * time.Second,
			// The configured JWKS URL is a trust anchor. Do not follow redirects
			// to a different endpoint when fetching signing keys.
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (v *JWKSValidator) Validate(ctx context.Context, token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed service JWT")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("malformed service JWT header")
	}
	var header serviceJWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Alg != "RS256" || header.Kid == "" {
		return nil, errors.New("unsupported service JWT header")
	}
	key, err := v.key(ctx, header.Kid)
	if err != nil {
		return nil, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("malformed service JWT signature")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return nil, errors.New("invalid service JWT signature")
	}
	claimBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("malformed service JWT claims")
	}
	var claims serviceJWTClaims
	if err := json.Unmarshal(claimBytes, &claims); err != nil {
		return nil, errors.New("malformed service JWT claims")
	}
	now := time.Now().Unix()
	if claims.Issuer != v.issuer || !serviceAudienceContains(claims.Audience, v.audience) || claims.Expires <= now ||
		(claims.NotBefore != 0 && claims.NotBefore > now) || (claims.IssuedAt != 0 && claims.IssuedAt > now+60) {
		return nil, errors.New("service JWT claims failed validation")
	}
	subject := strings.TrimSpace(claims.ClientID)
	if subject == "" {
		subject = strings.TrimSpace(claims.Subject)
	}
	if subject == "" {
		return nil, errors.New("service JWT has no client identity")
	}
	return &JWTClaims{
		Sub:       subject,
		ClientID:  claims.ClientID,
		ExpiresAt: claims.Expires,
		IssuedAt:  claims.IssuedAt,
		NotBefore: claims.NotBefore,
		Scope:     claims.Scope,
		Scopes:    claims.Scopes,
	}, nil
}

func serviceAudienceContains(raw json.RawMessage, expected string) bool {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return one == expected
	}
	var many []string
	if json.Unmarshal(raw, &many) == nil {
		for _, value := range many {
			if value == expected {
				return true
			}
		}
	}
	return false
}

func (v *JWKSValidator) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if time.Since(v.fetched) < serviceJWKSCacheTTL && v.keys != nil {
		if key := v.keys[kid]; key != nil {
			return key, nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.url, nil)
	if err != nil {
		return nil, errors.New("could not create JWKS request")
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, errors.New("JWKS endpoint unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("JWKS endpoint returned unsuccessful status")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, serviceJWKSMaxBytes+1))
	if err != nil || len(body) > serviceJWKSMaxBytes {
		return nil, errors.New("JWKS response invalid or too large")
	}
	var document serviceJWKS
	if err := json.Unmarshal(body, &document); err != nil {
		return nil, errors.New("JWKS response invalid")
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, item := range document.Keys {
		if item.Kty != "RSA" || (item.Use != "" && item.Use != "sig") || (item.Alg != "" && item.Alg != "RS256") || item.Kid == "" {
			continue
		}
		modulus, errN := base64.RawURLEncoding.DecodeString(item.N)
		exponent, errE := base64.RawURLEncoding.DecodeString(item.E)
		if errN != nil || errE != nil || len(modulus) < 256 || len(exponent) == 0 || len(exponent) > 4 {
			continue
		}
		modulusInt := new(big.Int).SetBytes(modulus)
		if modulusInt.BitLen() < 2048 {
			continue
		}
		e := 0
		for _, b := range exponent {
			e = e<<8 | int(b)
		}
		if e < 3 || e%2 == 0 {
			continue
		}
		keys[item.Kid] = &rsa.PublicKey{N: modulusInt, E: e}
	}
	if len(keys) == 0 {
		return nil, errors.New("JWKS contains no acceptable signing keys")
	}
	v.keys, v.fetched = keys, time.Now()
	if key := keys[kid]; key != nil {
		return key, nil
	}
	return nil, ErrJWKSKeyNotTrusted
}
\n// Warmup fetches and validates the configured JWKS before a production server becomes ready.\n// The sentinel for an intentionally unknown probe kid means the keyset itself was loaded.\nfunc (v *JWKSValidator) Warmup(ctx context.Context) error {\n\t_, err := v.key(ctx, "NAEOS-JWKS-WARMUP-UNKNOWN-KID")\n\tif errors.Is(err, ErrJWKSKeyNotTrusted) {\n\t\treturn nil\n\t}\n\treturn err\n}\n