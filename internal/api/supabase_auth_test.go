// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"context"
	"crypto"
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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/auth"
)

// supabaseTestIdentity holds the key and JWKS fixture used to mint signed
// Supabase-style access tokens in tests.
type supabaseTestIdentity struct {
	key      *rsa.PrivateKey
	kid      string
	issuer   string
	audience string
	jwksURL  string
}

func newSupabaseTestIdentity(t *testing.T) *supabaseTestIdentity {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	identity := &supabaseTestIdentity{
		key:      key,
		kid:      "test-key",
		issuer:   "https://ref.supabase.co/auth/v1",
		audience: "authenticated",
	}
	pub := key.PublicKey
	jwksDoc := map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"use": "sig",
			"alg": "RS256",
			"kid": identity.kid,
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}),
		}},
	}
	docBytes, err := json.Marshal(jwksDoc)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(docBytes)
	}))
	t.Cleanup(jwksServer.Close)
	identity.jwksURL = jwksServer.URL
	return identity
}

func (id *supabaseTestIdentity) sign(claims map[string]any) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"alg":"RS256","kid":%q,"typ":"JWT"}`, id.kid)))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, id.key, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (id *supabaseTestIdentity) validator() *JWKSValidator {
	return &JWKSValidator{
		issuer:   id.issuer,
		audience: id.audience,
		url:      id.jwksURL,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (id *supabaseTestIdentity) auth() *SupabaseAuth {
	return &SupabaseAuth{validator: id.validator()}
}

func supabaseAccessClaims(id *supabaseTestIdentity, sub string, roles []string) map[string]any {
	claims := map[string]any{
		"iss":  id.issuer,
		"aud":  "authenticated",
		"sub":  sub,
		"role": "authenticated",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	if roles != nil {
		claims["app_metadata"] = map[string]any{"naeos_roles": roles}
	}
	return claims
}

func TestParseSupabaseClaims(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{
		"sub":"11111111-1111-1111-1111-111111111111",
		"email":"dev@example.com",
		"role":"authenticated",
		"user_metadata":{"name":"Dev User"},
		"app_metadata":{"naeos_roles":["developer", "  ", 42]}
	}`))
	token := "header." + payload + ".signature"
	claims, err := ParseSupabaseClaims(token)
	if err != nil {
		t.Fatalf("ParseSupabaseClaims: %v", err)
	}
	if claims.Sub != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("sub = %q", claims.Sub)
	}
	if claims.Email != "dev@example.com" {
		t.Errorf("email = %q", claims.Email)
	}
	if claims.Name != "Dev User" {
		t.Errorf("name = %q", claims.Name)
	}
	if claims.Role != "authenticated" {
		t.Errorf("role = %q", claims.Role)
	}
	if len(claims.AppRoles) != 1 || claims.AppRoles[0] != "developer" {
		t.Errorf("app roles = %v", claims.AppRoles)
	}
}

func TestParseSupabaseClaimsDefaultsNameToEmail(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"s","email":"a@b.c","role":"anon"}`))
	claims, err := ParseSupabaseClaims("h." + payload + ".s")
	if err != nil {
		t.Fatalf("ParseSupabaseClaims: %v", err)
	}
	if claims.Name != "a@b.c" {
		t.Errorf("name = %q, want email fallback", claims.Name)
	}
}

func TestParseSupabaseClaimsRejectsMalformedTokens(t *testing.T) {
	if _, err := ParseSupabaseClaims("not-a-token"); err == nil {
		t.Fatal("expected error for malformed token")
	}
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"s"}`))
	if _, err := ParseSupabaseClaims("h.@@@invalid@@@.s"); err == nil {
		t.Fatal("expected error for invalid base64 payload")
	}
	_ = payload
}

func TestMapSupabaseRoles(t *testing.T) {
	rbac := auth.NewRBAC()
	auth.SetupDefaultRoles(rbac)
	cases := []struct {
		name  string
		roles []string
		want  []string
	}{
		{"no mapping falls back to viewer", nil, []string{"viewer"}},
		{"valid roles pass through", []string{"developer", "admin"}, []string{"developer", "admin"}},
		{"unknown roles are dropped", []string{"superuser", "root"}, []string{"viewer"}},
		{"mixed known and unknown", []string{"admin", "superuser"}, []string{"admin"}},
		{"duplicates removed", []string{"admin", "admin"}, []string{"admin"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MapSupabaseRoles(&SupabaseClaims{AppRoles: tc.roles}, rbac)
			if len(got) != len(tc.want) {
				t.Fatalf("roles = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("roles = %v, want %v", got, tc.want)
				}
			}
		})
	}
	if got := MapSupabaseRoles(&SupabaseClaims{AppRoles: []string{"admin"}}, nil); len(got) != 1 || got[0] != "viewer" {
		t.Errorf("nil rbac roles = %v, want [viewer]", got)
	}
}

func TestSupabaseAuthValidate(t *testing.T) {
	id := newSupabaseTestIdentity(t)
	sa := id.auth()

	token, err := id.sign(supabaseAccessClaims(id, "user-1", []string{"developer"}))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	baseClaims, claims, err := sa.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if baseClaims.Sub != "user-1" || claims.Sub != "user-1" {
		t.Errorf("subject mismatch: %q / %q", baseClaims.Sub, claims.Sub)
	}
	if claims.Email != "" {
		t.Errorf("unexpected email %q", claims.Email)
	}

	// Anonymous tokens must be rejected.
	anon := supabaseAccessClaims(id, "user-2", nil)
	anon["role"] = "anon"
	anon["aud"] = "anon"
	anonToken, err := id.sign(anon)
	if err != nil {
		t.Fatalf("sign anon token: %v", err)
	}
	if _, _, err := sa.Validate(context.Background(), anonToken); err == nil {
		t.Fatal("expected anon token rejection")
	}

	// Expired tokens must be rejected.
	expired := supabaseAccessClaims(id, "user-1", nil)
	expired["exp"] = time.Now().Add(-time.Hour).Unix()
	expiredToken, err := id.sign(expired)
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	if _, _, err := sa.Validate(context.Background(), expiredToken); err == nil {
		t.Fatal("expected expired token rejection")
	}
}

func TestServerSupabaseES256AuthFlow(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ECDSA key: %v", err)
	}
	const issuer = "https://ref.supabase.co/auth/v1"
	const audience = "authenticated"
	const kid = "es256-test-key"
	coordinate := func(value *big.Int) string {
		encoded := make([]byte, 32)
		value.FillBytes(encoded)
		return base64.RawURLEncoding.EncodeToString(encoded)
	}
	jwksDoc, err := json.Marshal(map[string]any{
		"keys": []map[string]any{{
			"kty": "EC", "use": "sig", "alg": "ES256", "kid": kid, "crv": "P-256",
			"x": coordinate(key.X), "y": coordinate(key.Y),
		}},
	})
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksDoc)
	}))
	defer jwksServer.Close()
	validator := &JWKSValidator{
		issuer: issuer, audience: audience, url: jwksServer.URL,
		client: &http.Client{Timeout: 5 * time.Second},
	}
	supabaseAuth := &SupabaseAuth{validator: validator}

	header, err := json.Marshal(map[string]string{"alg": "ES256", "kid": kid, "typ": "JWT"})
	if err != nil {
		t.Fatalf("marshal JWT header: %v", err)
	}
	payload, err := json.Marshal(map[string]any{
		"iss": issuer, "aud": audience, "sub": "supa-es256", "role": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("marshal JWT claims: %v", err)
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(header)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := encodedHeader + "." + encodedPayload
	digest := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		t.Fatalf("sign ES256 JWT: %v", err)
	}
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	token := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)

	server := NewServer(":0", &AuthConfig{Enabled: true, Supabase: supabaseAuth})
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/profiles", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ES256 Supabase user status = %d, want 200", resp.StatusCode)
	}
	if user, ok := server.authManager.GetUser("supa-es256"); !ok || user.ID != "supa-es256" {
		t.Fatalf("ES256 Supabase identity was not provisioned: user=%+v exists=%v", user, ok)
	}
}

func TestServerSupabaseAuthFlow(t *testing.T) {
	id := newSupabaseTestIdentity(t)
	server := NewServer(":0", &AuthConfig{Enabled: true, Supabase: id.auth()})
	_ = server

	// Viewer-default user reads a read-permitted route.
	viewerToken, err := id.sign(supabaseAccessClaims(id, "supa-viewer", nil))
	if err != nil {
		t.Fatalf("sign viewer token: %v", err)
	}
	// Admin user reaches an admin route.
	adminToken, err := id.sign(supabaseAccessClaims(id, "supa-admin", []string{"admin"}))
	if err != nil {
		t.Fatalf("sign admin token: %v", err)
	}

	ts := httptest.NewServer(server.Handler())
	t.Cleanup(ts.Close)

	getStatus := func(token string, path string) int {
		req, err := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := ts.Client().Do(req)
		if err != nil {
			t.Fatalf("do request: %v", err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	// No token → unauthorized.
	if status := getStatus("", "/api/v1/profiles"); status != http.StatusUnauthorized {
		t.Errorf("no-token status = %d, want 401", status)
	}

	// Viewer token → viewer-provisioned user can read profiles.
	if status := getStatus(viewerToken, "/api/v1/profiles"); status != http.StatusOK {
		t.Errorf("viewer token status = %d, want 200", status)
	}

	// Admin token → admin route accessible.
	if status := getStatus(adminToken, "/api/v1/version"); status != http.StatusOK {
		t.Errorf("admin token status = %d, want 200", status)
	}

	// Provisioned identities exist in the auth manager with mapped roles.
	adminUser, ok := server.authManager.GetUser("supa-admin")
	if !ok {
		t.Fatal("supa-admin not provisioned")
	}
	if len(adminUser.Roles) != 1 || adminUser.Roles[0] != "admin" {
		t.Errorf("admin roles = %v, want [admin]", adminUser.Roles)
	}
	viewerUser, ok := server.authManager.GetUser("supa-viewer")
	if !ok {
		t.Fatal("supa-viewer not provisioned")
	}
	if len(viewerUser.Roles) != 1 || viewerUser.Roles[0] != "viewer" {
		t.Errorf("viewer roles = %v, want [viewer]", viewerUser.Roles)
	}
}

func TestServerSupabaseAuthDisabled(t *testing.T) {
	id := newSupabaseTestIdentity(t)
	// Auth enabled but Supabase not configured: RS256 tokens must not be
	// accepted on non-repository routes.
	server := NewServer(":0", &AuthConfig{Enabled: true})
	token, err := id.sign(supabaseAccessClaims(id, "user-1", nil))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(ts.Close)
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/profiles", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}
