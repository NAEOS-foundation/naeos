// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/multitenant"
	"github.com/golang-jwt/jwt/v5"
)

type testJWKS struct {
	Keys []testJWK `json:"keys"`
}

type testJWK struct {
	KeyType string `json:"kty"`
	Use     string `json:"use"`
	Alg     string `json:"alg"`
	KeyID   string `json:"kid"`
	N       string `json:"n,omitempty"`
	E       string `json:"e,omitempty"`
	Curve   string `json:"crv,omitempty"`
	X       string `json:"x,omitempty"`
	Y       string `json:"y,omitempty"`
}

func testRS256Key(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

func testJWKForKey(key *rsa.PrivateKey, kid string) testJWK {
	exponent := make([]byte, 0, 4)
	for value := key.PublicKey.E; value > 0; value >>= 8 {
		exponent = append([]byte{byte(value)}, exponent...)
	}
	return testJWK{
		KeyType: "RSA",
		Use:     "sig",
		Alg:     "RS256",
		KeyID:   kid,
		N:       base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		E:       base64.RawURLEncoding.EncodeToString(exponent),
	}
}

func testJWKForECDSAKey(key *ecdsa.PrivateKey, kid string) testJWK {
	return testJWK{
		KeyType: "EC",
		Use:     "sig",
		Alg:     "ES256",
		KeyID:   kid,
		Curve:   "P-256",
		X:       base64.RawURLEncoding.EncodeToString(key.PublicKey.X.FillBytes(make([]byte, 32))),
		Y:       base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.FillBytes(make([]byte, 32))),
	}
}

func newTestSupabaseVerifier(t *testing.T, handler http.Handler) (*SupabaseAuthVerifier, string) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	verifier, err := newSupabaseAuthVerifier(SupabaseAuthConfig{
		URL:      server.URL,
		Audience: defaultSupabaseAudience,
	}, server.Client())
	if err != nil {
		t.Fatalf("create Supabase verifier: %v", err)
	}
	t.Cleanup(verifier.Close)
	return verifier, server.URL + "/auth/v1"
}

func signSupabaseTestToken(t *testing.T, key *rsa.PrivateKey, kid string, issuer, audience, role string, issuedAt, expiresAt time.Time, appMetadata map[string]any) string {
	t.Helper()
	claims := map[string]any{
		"iss":          issuer,
		"sub":          "user-123",
		"aud":          audience,
		"role":         role,
		"exp":          expiresAt.Unix(),
		"app_metadata": appMetadata,
	}
	if !issuedAt.IsZero() {
		claims["iat"] = issuedAt.Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(claims))
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign Supabase test token: %v", err)
	}
	return signed
}

func TestSupabaseAuthVerifierValidateAccessToken(t *testing.T) {
	key := testRS256Key(t)
	jwks, err := json.Marshal(testJWKS{Keys: []testJWK{testJWKForKey(key, "key-1")}})
	if err != nil {
		t.Fatal(err)
	}
	verifier, issuer := newTestSupabaseVerifier(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(jwks)
	}))

	now := time.Now()
	token := signSupabaseTestToken(t, key, "key-1", issuer, defaultSupabaseAudience, "authenticated",
		now, now.Add(time.Hour), map[string]any{
			"naeos_roles":     []string{"developer", "admin", "not-a-role", "admin"},
			"naeos_tenant_id": "tenant-42",
		})
	user, err := verifier.Validate(token)
	if err != nil {
		t.Fatalf("validate access token: %v", err)
	}
	if user.ID != "user-123" || user.TenantID != "tenant-42" {
		t.Fatalf("unexpected identity claims: %+v", user)
	}
	if len(user.Roles) != 2 || user.Roles[0] != "developer" || user.Roles[1] != "admin" {
		t.Fatalf("unexpected allow-listed roles: %v", user.Roles)
	}
}

func TestSupabaseAuthVerifierAcceptsES256(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ECDSA key: %v", err)
	}
	jwks, err := json.Marshal(testJWKS{Keys: []testJWK{testJWKForECDSAKey(key, "ec-key")}})
	if err != nil {
		t.Fatal(err)
	}
	verifier, issuer := newTestSupabaseVerifier(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(jwks)
	}))
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss":  issuer,
		"sub":  "user-ec",
		"aud":  defaultSupabaseAudience,
		"role": "authenticated",
		"iat":  now.Unix(),
		"exp":  now.Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "ec-key"
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign ES256 test token: %v", err)
	}
	user, err := verifier.Validate(signed)
	if err != nil {
		t.Fatalf("validate ES256 access token: %v", err)
	}
	if user.ID != "user-ec" {
		t.Fatalf("subject = %q, want user-ec", user.ID)
	}
}

func TestSupabaseAuthVerifierRejectsInvalidClaims(t *testing.T) {
	key := testRS256Key(t)
	jwks, err := json.Marshal(testJWKS{Keys: []testJWK{testJWKForKey(key, "key-1")}})
	if err != nil {
		t.Fatal(err)
	}
	verifier, issuer := newTestSupabaseVerifier(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(jwks)
	}))
	now := time.Now()

	tests := []struct {
		name     string
		issuer   string
		audience string
		role     string
		issuedAt time.Time
		expiry   time.Time
	}{
		{name: "wrong issuer", issuer: issuer + "/wrong", audience: defaultSupabaseAudience, role: "authenticated", issuedAt: now, expiry: now.Add(time.Hour)},
		{name: "wrong audience", issuer: issuer, audience: "other", role: "authenticated", issuedAt: now, expiry: now.Add(time.Hour)},
		{name: "expired", issuer: issuer, audience: defaultSupabaseAudience, role: "authenticated", issuedAt: now.Add(-time.Hour), expiry: now.Add(-time.Minute)},
		{name: "anonymous role", issuer: issuer, audience: defaultSupabaseAudience, role: "anon", issuedAt: now, expiry: now.Add(time.Hour)},
		{name: "missing issued-at", issuer: issuer, audience: defaultSupabaseAudience, role: "authenticated", expiry: now.Add(time.Hour)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := signSupabaseTestToken(t, key, "key-1", tt.issuer, tt.audience, tt.role, tt.issuedAt, tt.expiry, nil)
			if _, err := verifier.Validate(token); err == nil {
				t.Fatal("expected invalid access token to be rejected")
			}
		})
	}
}

func TestSupabaseAuthVerifierDefaultsToViewerRole(t *testing.T) {
	key := testRS256Key(t)
	jwks, err := json.Marshal(testJWKS{Keys: []testJWK{testJWKForKey(key, "key-1")}})
	if err != nil {
		t.Fatal(err)
	}
	verifier, issuer := newTestSupabaseVerifier(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(jwks)
	}))
	now := time.Now()
	token := signSupabaseTestToken(t, key, "key-1", issuer, defaultSupabaseAudience, "authenticated",
		now, now.Add(time.Hour), map[string]any{"naeos_roles": []string{"unknown"}})
	user, err := verifier.Validate(token)
	if err != nil {
		t.Fatalf("validate access token: %v", err)
	}
	if len(user.Roles) != 1 || user.Roles[0] != "viewer" {
		t.Fatalf("expected least-privilege viewer role, got %v", user.Roles)
	}
}

func TestSupabaseAuthVerifierRefreshesUnknownKeyID(t *testing.T) {
	firstKey := testRS256Key(t)
	secondKey := testRS256Key(t)
	var mu sync.RWMutex
	currentJWKS := testJWKS{Keys: []testJWK{testJWKForKey(firstKey, "first")}}
	verifier, issuer := newTestSupabaseVerifier(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.RLock()
		defer mu.RUnlock()
		_ = json.NewEncoder(w).Encode(currentJWKS)
	}))
	mu.Lock()
	currentJWKS = testJWKS{Keys: []testJWK{testJWKForKey(secondKey, "second")}}
	mu.Unlock()

	now := time.Now()
	token := signSupabaseTestToken(t, secondKey, "second", issuer, defaultSupabaseAudience, "authenticated",
		now, now.Add(time.Hour), nil)
	if _, err := verifier.Validate(token); err != nil {
		t.Fatalf("validate token after key rotation: %v", err)
	}
}

func TestSupabaseAuthVerifierRejectsUntrustedURLs(t *testing.T) {
	for _, config := range []SupabaseAuthConfig{
		{URL: "http://supabase.example"},
		{URL: "https://supabase.example/other"},
		{URL: "https://user:password@supabase.example"},
		{URL: "https://supabase.example", JWKSURL: "http://keys.example/jwks"},
	} {
		if _, err := NewSupabaseAuthVerifier(config); err == nil {
			t.Errorf("expected URL configuration %q to be rejected", config.URL)
		}
	}
}

func TestSupabaseMiddlewareUsesRolesAndDeniesUnmappedRoutes(t *testing.T) {
	key := testRS256Key(t)
	jwks, err := json.Marshal(testJWKS{Keys: []testJWK{testJWKForKey(key, "key-1")}})
	if err != nil {
		t.Fatal(err)
	}
	verifier, issuer := newTestSupabaseVerifier(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(jwks)
	}))
	server := NewServer(":0", &AuthConfig{Enabled: true, SupabaseAuth: verifier})
	handler := server.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	now := time.Now()
	token := signSupabaseTestToken(t, key, "key-1", issuer, defaultSupabaseAudience, "authenticated",
		now, now.Add(time.Hour), map[string]any{"naeos_roles": []string{"viewer"}})

	tests := []struct {
		path   string
		method string
		want   int
	}{
		{path: "/api/v1/specs", method: http.MethodGet, want: http.StatusOK},
		{path: "/api/v1/specs/validate", method: http.MethodPost, want: http.StatusForbidden},
		{path: "/api/v1/unmapped", method: http.MethodGet, want: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.path+strconv.Itoa(tt.want), func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(nil))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}

func TestSupabaseMiddlewareRequiresTrustedTenantClaim(t *testing.T) {
	key := testRS256Key(t)
	jwks, err := json.Marshal(testJWKS{Keys: []testJWK{testJWKForKey(key, "key-1")}})
	if err != nil {
		t.Fatal(err)
	}
	verifier, issuer := newTestSupabaseVerifier(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(jwks)
	}))
	server := NewServer(":0", &AuthConfig{Enabled: true, SupabaseAuth: verifier})
	workspace := multitenant.New()
	if _, err := workspace.CreateTenant("tenant-42", "Test tenant", "free"); err != nil {
		t.Fatal(err)
	}
	server.SetWorkspace(workspace)
	handler := server.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	now := time.Now()
	tokenWithoutTenant := signSupabaseTestToken(t, key, "key-1", issuer, defaultSupabaseAudience, "authenticated",
		now, now.Add(time.Hour), nil)
	tokenWithTenant := signSupabaseTestToken(t, key, "key-1", issuer, defaultSupabaseAudience, "authenticated",
		now, now.Add(time.Hour), map[string]any{"naeos_tenant_id": "tenant-42"})

	for _, tt := range []struct {
		name  string
		token string
		want  int
	}{
		{name: "untrusted requested tenant", token: tokenWithoutTenant, want: http.StatusForbidden},
		{name: "trusted tenant claim", token: tokenWithTenant, want: http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/specs", nil)
			request.Header.Set("Authorization", "Bearer "+tt.token)
			request.Header.Set("X-Tenant-ID", "tenant-42")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}

func TestBearerTokenRejectsMalformedAndOversizedHeaders(t *testing.T) {
	for _, header := range []string{
		"",
		"Basic token",
		"Bearer",
		"Bearer one two",
		"Bearer " + strings.Repeat("a", (16<<10)+1),
	} {
		if _, ok := bearerToken(header); ok {
			t.Errorf("bearerToken(%q) unexpectedly accepted the header", header[:min(len(header), 24)])
		}
	}
	if token, ok := bearerToken("Bearer valid-token"); !ok || token != "valid-token" {
		t.Fatalf("valid bearer token parsed as %q, ok=%v", token, ok)
	}
}
