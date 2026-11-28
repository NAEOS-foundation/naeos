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
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
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
			"iss":"https://ref.supabase.co/auth/v1",
			"sub":"11111111-1111-1111-1111-111111111111",
			"client_id":"web-client-123",
			"aud":["authenticated"],
			"email":"dev@example.com",
		"role":"authenticated",
			"user_metadata":{"name":"Dev User","naeos_roles":["admin"]},
		"app_metadata":{"naeos_roles":["developer", "  ", 42],"naeos_tenant_id":"tenant-a"}
	}`))
	token := "header." + payload + ".signature"
	claims, err := ParseSupabaseClaims(token)
	if err != nil {
		t.Fatalf("ParseSupabaseClaims: %v", err)
	}
	if claims.Sub != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("sub = %q", claims.Sub)
	}
	if claims.Issuer != "https://ref.supabase.co/auth/v1" {
		t.Errorf("iss = %q", claims.Issuer)
	}
	if claims.ClientID != "web-client-123" {
		t.Errorf("client_id = %q", claims.ClientID)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "authenticated" {
		t.Errorf("aud = %v", claims.Audience)
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
	if claims.TenantID != "tenant-a" {
		t.Errorf("tenant ID = %q, want tenant-a", claims.TenantID)
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

	validClaims := supabaseAccessClaims(id, "user-1", []string{"developer"})
	validClaims["client_id"] = "web-client-1"
	token, err := id.sign(validClaims)
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
	if baseClaims.ClientID != "web-client-1" || claims.ClientID != "web-client-1" {
		t.Errorf("client ID mismatch: %q / %q", baseClaims.ClientID, claims.ClientID)
	}
	if claims.Issuer != id.issuer || len(claims.Audience) != 1 || claims.Audience[0] != id.audience {
		t.Errorf("issuer/audience mismatch: %q / %v", claims.Issuer, claims.Audience)
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

func TestSupabaseAuthRejectsMissingSubjectAndNonUserTokens(t *testing.T) {
	id := newSupabaseTestIdentity(t)
	sa := id.auth()
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "client ID cannot replace missing subject",
			mutate: func(claims map[string]any) {
				delete(claims, "sub")
				claims["client_id"] = "client-only"
			},
		},
		{
			name: "whitespace-only subject",
			mutate: func(claims map[string]any) {
				claims["sub"] = " \t\n"
				claims["client_id"] = "client-only"
			},
		},
		{
			name: "service role token",
			mutate: func(claims map[string]any) {
				claims["role"] = "service_role"
				claims["client_id"] = "service-client"
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := supabaseAccessClaims(id, "user-1", nil)
			test.mutate(claims)
			token, err := id.sign(claims)
			if err != nil {
				t.Fatalf("sign token: %v", err)
			}
			if _, _, err := sa.Validate(context.Background(), token); err == nil {
				t.Fatal("Validate accepted a token without an authenticated user subject")
			}
		})
	}
}

func TestSupabaseAuthRejectsIssuerAndAudienceMismatch(t *testing.T) {
	id := newSupabaseTestIdentity(t)
	sa := id.auth()
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "wrong issuer", mutate: func(claims map[string]any) { claims["iss"] = "https://other.supabase.co/auth/v1" }},
		{name: "wrong audience", mutate: func(claims map[string]any) { claims["aud"] = "other-project" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			claims := supabaseAccessClaims(id, "user-1", nil)
			test.mutate(claims)
			token, err := id.sign(claims)
			if err != nil {
				t.Fatalf("sign token: %v", err)
			}
			if _, _, err := sa.Validate(context.Background(), token); err == nil {
				t.Fatal("Validate accepted mismatched issuer or audience")
			}
		})
	}
}

func TestSupabaseAuthPreservesOpaqueSubjectAndClientID(t *testing.T) {
	id := newSupabaseTestIdentity(t)
	claims := supabaseAccessClaims(id, "  opaque-user-subject \t", nil)
	claims["client_id"] = "  browser-client  "
	token, err := id.sign(claims)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	baseClaims, supabaseClaims, err := id.auth().Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if baseClaims.Sub != "  opaque-user-subject \t" || supabaseClaims.Sub != baseClaims.Sub {
		t.Fatalf("subject was not preserved: %q / %q", baseClaims.Sub, supabaseClaims.Sub)
	}
	if baseClaims.ClientID != "  browser-client  " || supabaseClaims.ClientID != baseClaims.ClientID {
		t.Fatalf("client ID was not preserved separately: %q / %q", baseClaims.ClientID, supabaseClaims.ClientID)
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
		"iss": issuer, "aud": audience, "sub": "supa-es256", "client_id": "browser-client", "role": "authenticated",
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
	if _, ok := server.authManager.GetUser("browser-client"); ok {
		t.Fatal("Supabase client_id was incorrectly provisioned as a human user")
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
	if status := getStatus(viewerToken, "/api/v1/artifacts"); status != http.StatusForbidden {
		t.Errorf("Supabase artifact read status = %d, want 403 until tenant scoping is supported", status)
	}
	tenantClaims := supabaseAccessClaims(id, "tenant-viewer", []string{"viewer"})
	tenantClaims["app_metadata"] = map[string]any{"naeos_tenant_id": "tenant-a"}
	tenantToken, err := id.sign(tenantClaims)
	if err != nil {
		t.Fatalf("sign tenant token: %v", err)
	}
	if status := getStatus(tenantToken, "/api/v1/artifacts"); status != http.StatusForbidden {
		t.Errorf("tenant-scoped Supabase artifact read status = %d, want 403", status)
	}
	artifactBody := strings.NewReader(`{"path":"viewer.txt","content":"not allowed"}`)
	artifactReq, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/artifacts", artifactBody)
	if err != nil {
		t.Fatalf("new artifact request: %v", err)
	}
	artifactReq.Header.Set("Authorization", "Bearer "+viewerToken)
	artifactReq.Header.Set("Content-Type", "application/json")
	artifactResp, err := ts.Client().Do(artifactReq)
	if err != nil {
		t.Fatalf("post artifact: %v", err)
	}
	artifactResp.Body.Close()
	if artifactResp.StatusCode != http.StatusForbidden {
		t.Errorf("viewer artifact POST status = %d, want 403", artifactResp.StatusCode)
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

func TestServerSupabaseTenantClaimScopesPipelineStatus(t *testing.T) {
	id := newSupabaseTestIdentity(t)
	server := NewServer(":0", &AuthConfig{Enabled: true, Supabase: id.auth()})
	server.pipelines = []pipelineRun{
		{ID: "tenant-a-run", TenantID: "tenant-a", Status: "completed"},
		{ID: "tenant-b-run", TenantID: "tenant-b", Status: "completed"},
	}
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(ts.Close)

	claims := supabaseAccessClaims(id, "tenant-user", []string{"viewer"})
	claims["app_metadata"] = map[string]any{
		"naeos_roles":     []string{"viewer"},
		"naeos_tenant_id": "tenant-a",
	}
	token, err := id.sign(claims)
	if err != nil {
		t.Fatalf("sign tenant token: %v", err)
	}

	request := func(tenantHeader string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/pipeline/status", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if tenantHeader != "" {
			req.Header.Set("X-Tenant-ID", tenantHeader)
		}
		resp, err := ts.Client().Do(req)
		if err != nil {
			t.Fatalf("do request: %v", err)
		}
		return resp
	}

	resp := request("")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("trusted tenant request status = %d, want 200", resp.StatusCode)
	}
	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode pipeline status: %v", err)
	}
	data, err := json.Marshal(result.Data)
	if err != nil {
		t.Fatalf("marshal pipeline status data: %v", err)
	}
	if strings.Contains(string(data), "tenant-b-run") {
		t.Fatalf("pipeline status leaked another tenant's run: %s", data)
	}
	if !strings.Contains(string(data), "tenant-a-run") {
		t.Fatalf("pipeline status omitted the caller's run: %s", data)
	}

	mismatch := request("tenant-b")
	defer mismatch.Body.Close()
	if mismatch.StatusCode != http.StatusForbidden {
		t.Fatalf("mismatched tenant header status = %d, want 403", mismatch.StatusCode)
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
