// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

	get := func(token string, path string) *http.Response {
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
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}

	// No token → unauthorized.
	if resp := get("", "/api/v1/profiles"); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no-token status = %d, want 401", resp.StatusCode)
	}

	// Viewer token → viewer-provisioned user can read profiles.
	if resp := get(viewerToken, "/api/v1/profiles"); resp.StatusCode != http.StatusOK {
		t.Errorf("viewer token status = %d, want 200", resp.StatusCode)
	}

	// Admin token → admin route accessible.
	if resp := get(adminToken, "/api/v1/version"); resp.StatusCode != http.StatusOK {
		t.Errorf("admin token status = %d, want 200", resp.StatusCode)
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
