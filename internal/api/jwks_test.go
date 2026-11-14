// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testServiceClaims struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	ClientID  string   `json:"client_id,omitempty"`
	Audience  []string `json:"aud"`
	Expires   int64    `json:"exp"`
	NotBefore int64    `json:"nbf,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Scope     string   `json:"scope,omitempty"`
}

func newTestJWKSValidator(t *testing.T) (*JWKSValidator, *rsa.PrivateKey) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	validator, err := NewJWKSValidator("https://issuer.example", "naeos-core", "https://issuer.example/jwks")
	if err != nil {
		t.Fatalf("create JWKS validator: %v", err)
	}
	validator.keys = map[string]crypto.PublicKey{"test-key": &key.PublicKey}
	validator.fetched = time.Now()
	return validator, key
}

func makeTestServiceToken(t *testing.T, key *rsa.PrivateKey, claims testServiceClaims) string {
	t.Helper()

	header, err := json.Marshal(serviceJWTHeader{Alg: "RS256", Kid: "test-key"})
	if err != nil {
		t.Fatalf("marshal service JWT header: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal service JWT claims: %v", err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign service JWT: %v", err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func defaultTestServiceClaims() testServiceClaims {
	now := time.Now().Unix()
	return testServiceClaims{
		Issuer:   "https://issuer.example",
		Subject:  "naeos-intelligence",
		Audience: []string{"naeos-core"},
		Expires:  now + 300,
		IssuedAt: now,
		Scope:    "naeos:repository:read",
	}
}

func TestJWKSValidatorValidatesServiceJWT(t *testing.T) {
	validator, key := newTestJWKSValidator(t)
	token := makeTestServiceToken(t, key, defaultTestServiceClaims())

	claims, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Sub != "naeos-intelligence" || !hasJWTClaimScope(claims, "naeos:repository:read") {
		t.Fatalf("unexpected service identity or scope: %+v", claims)
	}
}

func TestJWKSValidatorRejectsWrongIssuerAudienceAndExpiredToken(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*testServiceClaims)
	}{
		{name: "wrong issuer", mutate: func(c *testServiceClaims) { c.Issuer = "https://untrusted.example" }},
		{name: "wrong audience", mutate: func(c *testServiceClaims) { c.Audience = []string{"other-api"} }},
		{name: "expired", mutate: func(c *testServiceClaims) { c.Expires = time.Now().Add(-time.Minute).Unix() }},
	} {
		t.Run(test.name, func(t *testing.T) {
			validator, key := newTestJWKSValidator(t)
			claims := defaultTestServiceClaims()
			test.mutate(&claims)
			token := makeTestServiceToken(t, key, claims)
			if _, err := validator.Validate(context.Background(), token); err == nil {
				t.Fatal("Validate accepted a token with invalid required claims")
			}
		})
	}
}

func TestServiceJWTIsRestrictedToAllowlistedReadOnlyRepositoryRoute(t *testing.T) {
	validator, key := newTestJWKSValidator(t)
	server := NewServer(":0", &AuthConfig{Enabled: false, ServiceJWKS: validator})
	t.Setenv("NAEOS_GITHUB_ALLOWED_REPOSITORIES", "")

	publicRequest := httptest.NewRequest(http.MethodGet, "/api/v1/specs", nil)
	publicResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(publicResponse, publicRequest)
	if publicResponse.Code != http.StatusOK {
		t.Fatalf("unrelated route status = %d, want %d when user auth is disabled", publicResponse.Code, http.StatusOK)
	}

	tests := []struct {
		name   string
		method string
		path   string
		scope  string
		want   string
	}{
		{name: "wrong scope", method: http.MethodGet, path: "/api/v1/repositories/owner/repo", scope: "naeos:repository:write", want: "service token lacks permission"},
		{name: "wrong method", method: http.MethodPost, path: "/api/v1/repositories/owner/repo", scope: "naeos:repository:read", want: "service token lacks permission"},
		{name: "allowlist enforced", method: http.MethodGet, path: "/api/v1/repositories/owner/repo", scope: "naeos:repository:read", want: "repository is not enabled"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := defaultTestServiceClaims()
			claims.Scope = test.scope
			req := httptest.NewRequest(test.method, test.path, nil)
			req.Header.Set("Authorization", "Bearer "+makeTestServiceToken(t, key, claims))
			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusForbidden, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), test.want) {
				t.Fatalf("response %q does not contain %q", w.Body.String(), test.want)
			}
		})
	}

	userProtectedServer := NewServer(":0", &AuthConfig{Enabled: true, JWTSecret: "test-user-secret", ServiceJWKS: validator})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/specs", nil)
	req.Header.Set("Authorization", "Bearer "+makeTestServiceToken(t, key, defaultTestServiceClaims()))
	w := httptest.NewRecorder()
	userProtectedServer.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "service token lacks permission") {
		t.Fatalf("service JWT reached a user route: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestJWKSValidatorDoesNotFollowRedirects(t *testing.T) {
	validator, err := NewJWKSValidator("https://issuer.example", "naeos-core", "https://issuer.example/jwks")
	if err != nil {
		t.Fatalf("create JWKS validator: %v", err)
	}
	if validator.client.CheckRedirect == nil {
		t.Fatal("JWKS client must define an explicit redirect policy")
	}
	request := httptest.NewRequest(http.MethodGet, "https://attacker.example/jwks", nil)
	if err := validator.client.CheckRedirect(request, nil); err != http.ErrUseLastResponse {
		t.Fatalf("redirect policy error = %v, want %v", err, http.ErrUseLastResponse)
	}
}

func TestJWKSValidatorWarmupAcceptsCachedTrustedKeyset(t *testing.T) {
	validator, _ := newTestJWKSValidator(t)
	if err := validator.Warmup(context.Background()); err != nil {
		t.Fatalf("Warmup with an initialized keyset: %v", err)
	}
}

func TestJWKSValidatorDoesNotRefreshForUnknownKidWithinCacheTTL(t *testing.T) {
	validator, _ := newTestJWKSValidator(t)
	requests := 0
	validator.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return nil, errors.New("unexpected network fetch")
	})
	for i := 0; i < 2; i++ {
		if _, err := validator.key(context.Background(), "attacker-controlled-kid"); err == nil {
			t.Fatal("unknown kid was trusted")
		}
	}
	if requests != 0 {
		t.Fatalf("JWKS network refreshes = %d; want 0 while cache is fresh", requests)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestJWKSValidatorValidatesES256SupabaseStyleToken(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ES256 key: %v", err)
	}
	validator, err := NewJWKSValidator("https://issuer.example", "naeos-control-plane", "https://issuer.example/.well-known/jwks.json")
	if err != nil {
		t.Fatalf("create validator: %v", err)
	}
	validator.keys = map[string]crypto.PublicKey{"ec-key": &key.PublicKey}
	validator.fetched = time.Now()

	header, err := json.Marshal(serviceJWTHeader{Alg: "ES256", Kid: "ec-key"})
	if err != nil {
		t.Fatal(err)
	}
	claims := defaultTestServiceClaims()
	claims.Audience = []string{"naeos-control-plane"}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(unsigned))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		t.Fatalf("sign ES256 token: %v", err)
	}
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	token := unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)

	verified, err := validator.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate ES256: %v", err)
	}
	if verified.Sub != claims.Subject {
		t.Fatalf("verified subject = %q; want %q", verified.Sub, claims.Subject)
	}
}

func TestJWKSValidatorWarmupLoadsES256JWKFromEndpoint(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ES256 key: %v", err)
	}
	x := make([]byte, 32)
	y := make([]byte, 32)
	key.X.FillBytes(x)
	key.Y.FillBytes(y)
	document, err := json.Marshal(serviceJWKS{Keys: []serviceJWK{{
		Kty: "EC", Use: "sig", Alg: "ES256", Kid: "ec-key", Crv: "P-256",
		X: base64.RawURLEncoding.EncodeToString(x), Y: base64.RawURLEncoding.EncodeToString(y),
	}}})
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}
	validator, err := NewJWKSValidator("https://issuer.example", "naeos-control-plane", "https://issuer.example/.well-known/jwks.json")
	if err != nil {
		t.Fatalf("create validator: %v", err)
	}
	validator.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(document))),
			Request:    request,
		}, nil
	})
	if err := validator.Warmup(context.Background()); err != nil {
		t.Fatalf("Warmup: %v", err)
	}
	if _, ok := validator.keys["ec-key"].(*ecdsa.PublicKey); !ok {
		t.Fatalf("loaded JWKS key type = %T; want *ecdsa.PublicKey", validator.keys["ec-key"])
	}
}
