// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"crypto"
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
)

func generateServiceTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

func serviceTestJWKS(t *testing.T, key *rsa.PrivateKey, kid string) []byte {
	t.Helper()
	document := serviceJWKS{Keys: []serviceJWK{{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(bigIntBytes(key.E)),
	}}}
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}
	return body
}

func bigIntBytes(value int) []byte {
	return new(big.Int).SetInt64(int64(value)).Bytes()
}

func signedServiceTestToken(t *testing.T, key *rsa.PrivateKey, kid string, claims serviceJWTClaims) string {
	t.Helper()
	header, err := json.Marshal(serviceJWTHeader{Alg: "RS256", Kid: kid})
	if err != nil {
		t.Fatalf("marshal JWT header: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal JWT claims: %v", err)
	}
	input := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(input))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign service JWT: %v", err)
	}
	return input + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func validServiceTestClaims(issuer, audience, scope string) serviceJWTClaims {
	now := time.Now().Unix()
	return serviceJWTClaims{
		Issuer:    issuer,
		Subject:   "service-subject",
		ClientID:  "crm-intelligence-client",
		Audience:  json.RawMessage(fmt.Sprintf("%q", audience)),
		Expires:   now + 300,
		NotBefore: now - 5,
		IssuedAt:  now,
		Scope:     scope,
	}
}

func TestNewJWKSValidatorRequiresHTTPS(t *testing.T) {
	if _, err := NewJWKSValidator("https://issuer.example", "naeos-core", "http://keys.example/jwks"); err == nil {
		t.Fatal("expected HTTP JWKS URL to be rejected")
	}
	if _, err := NewJWKSValidator("https://issuer.example", "naeos-core", "https://keys.example/jwks"); err != nil {
		t.Fatalf("NewJWKSValidator() error = %v", err)
	}
}

func TestJWKSValidatorAcceptsScopedRS256ServiceToken(t *testing.T) {
	key := generateServiceTestKey(t)
	const issuer, audience, kid = "https://issuer.example", "naeos-core", "signing-key-1"
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(serviceTestJWKS(t, key, kid))
	}))
	defer jwksServer.Close()

	validator := &JWKSValidator{
		issuer:   issuer,
		audience: audience,
		url:      jwksServer.URL,
		client:   jwksServer.Client(),
	}
	token := signedServiceTestToken(t, key, kid, validServiceTestClaims(issuer, audience, "naeos:repository:read"))
	claims, err := validator.Validate(t.Context(), token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.Sub != "crm-intelligence-client" {
		t.Fatalf("client identity = %q, want crm-intelligence-client", claims.Sub)
	}
	if !hasJWTClaimScope(claims, "naeos:repository:read") {
		t.Fatal("expected repository-read scope")
	}
}

func TestJWKSValidatorRejectsInvalidClaimsAndHeader(t *testing.T) {
	key := generateServiceTestKey(t)
	const issuer, audience, kid = "https://issuer.example", "naeos-core", "signing-key-1"
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(serviceTestJWKS(t, key, kid))
	}))
	defer jwksServer.Close()
	validator := &JWKSValidator{issuer: issuer, audience: audience, url: jwksServer.URL, client: jwksServer.Client()}

	tests := []struct {
		name   string
		mutate func(*serviceJWTClaims)
	}{
		{name: "wrong issuer", mutate: func(c *serviceJWTClaims) { c.Issuer = "https://other.example" }},
		{name: "wrong audience", mutate: func(c *serviceJWTClaims) { c.Audience = json.RawMessage(`"other-api"`) }},
		{name: "expired", mutate: func(c *serviceJWTClaims) { c.Expires = time.Now().Unix() - 1 }},
		{name: "not yet valid", mutate: func(c *serviceJWTClaims) { c.NotBefore = time.Now().Unix() + 60 }},
		{name: "future issued at", mutate: func(c *serviceJWTClaims) { c.IssuedAt = time.Now().Unix() + 120 }},
		{name: "missing identity", mutate: func(c *serviceJWTClaims) { c.ClientID = ""; c.Subject = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := validServiceTestClaims(issuer, audience, "naeos:repository:read")
			tt.mutate(&claims)
			token := signedServiceTestToken(t, key, kid, claims)
			if _, err := validator.Validate(t.Context(), token); err == nil {
				t.Fatal("Validate() expected error")
			}
		})
	}

	header, _ := json.Marshal(serviceJWTHeader{Alg: "HS256", Kid: kid})
	payload, _ := json.Marshal(validServiceTestClaims(issuer, audience, "naeos:repository:read"))
	token := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload) + ".invalid"
	if _, err := validator.Validate(t.Context(), token); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("Validate() error = %v, want unsupported header", err)
	}
}

func TestServiceJWTMiddlewareRestrictsReadOnlyRepositoryEndpoint(t *testing.T) {
	key := generateServiceTestKey(t)
	const issuer, audience, kid = "https://issuer.example", "naeos-core", "signing-key-1"
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(serviceTestJWKS(t, key, kid))
	}))
	defer jwksServer.Close()

	githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/acme/project" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		if r.Header.Get("Authorization") != "Bearer github-test-token" {
			http.Error(w, "missing provider token", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 12345,
			"name": "project",
			"full_name": "acme/project",
			"owner": {"login": "acme"},
			"description": "test repository",
			"private": false,
			"archived": false,
			"default_branch": "main",
			"open_issues_count": 7,
			"stargazers_count": 11,
			"forks_count": 2,
			"size": 42,
			"created_at": "2026-01-01T00:00:00Z",
			"updated_at": "2026-01-02T00:00:00Z",
			"html_url": "https://github.com/acme/project"
		}`))
	}))
	defer githubServer.Close()

	t.Setenv("NAEOS_GITHUB_ALLOWED_REPOSITORIES", "acme/project")
	t.Setenv("GITHUB_TOKEN", "github-test-token")
	validator := &JWKSValidator{issuer: issuer, audience: audience, url: jwksServer.URL, client: jwksServer.Client()}
	server := &Server{
		Auth:                 &AuthConfig{Enabled: true, ServiceJWKS: validator},
		Router:               http.NewServeMux(),
		Limiter:              NewRateLimiter(100, time.Minute),
		repositoryAPIBase:    githubServer.URL,
		repositoryHTTPClient: githubServer.Client(),
	}
	server.Router.HandleFunc("/api/v1/repositories/", server.handleRepositoryMetadata)
	server.Router.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) { server.writeJSON(w, http.StatusOK, nil) })
	t.Cleanup(server.Limiter.Stop)
	handler := server.handlerWithMiddleware(server.Router.ServeHTTP)

	validToken := signedServiceTestToken(t, key, kid, validServiceTestClaims(issuer, audience, "naeos:repository:read"))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/acme/project", nil)
	request.Header.Set("Authorization", "Bearer "+validToken)
	response := httptest.NewRecorder()
	handler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("repository metadata status = %d, body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Success bool               `json:"success"`
		Data    repositoryMetadata `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode repository response: %v", err)
	}
	if !envelope.Success || envelope.Data.ID != "12345" || envelope.Data.FullName != "acme/project" {
		t.Fatalf("unexpected repository response: %+v", envelope)
	}

	otherRoute := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	otherRoute.Header.Set("Authorization", "Bearer "+validToken)
	otherResponse := httptest.NewRecorder()
	handler(otherResponse, otherRoute)
	if otherResponse.Code != http.StatusForbidden {
		t.Fatalf("service token other-route status = %d, want 403", otherResponse.Code)
	}

	writeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/repositories/acme/project", strings.NewReader("{}"))
	writeRequest.Header.Set("Authorization", "Bearer "+validToken)
	writeResponse := httptest.NewRecorder()
	handler(writeResponse, writeRequest)
	if writeResponse.Code != http.StatusForbidden {
		t.Fatalf("service token write status = %d, want 403", writeResponse.Code)
	}

	noScope := validServiceTestClaims(issuer, audience, "naeos:repository:write")
	noScopeToken := signedServiceTestToken(t, key, kid, noScope)
	noScopeRequest := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/acme/project", nil)
	noScopeRequest.Header.Set("Authorization", "Bearer "+noScopeToken)
	noScopeResponse := httptest.NewRecorder()
	handler(noScopeResponse, noScopeRequest)
	if noScopeResponse.Code != http.StatusForbidden {
		t.Fatalf("missing-scope status = %d, want 403", noScopeResponse.Code)
	}

	notAllowed := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/acme/private-project", nil)
	notAllowed.Header.Set("Authorization", "Bearer "+validToken)
	notAllowedResponse := httptest.NewRecorder()
	handler(notAllowedResponse, notAllowed)
	if notAllowedResponse.Code != http.StatusForbidden {
		t.Fatalf("non-allowlisted repository status = %d, want 403", notAllowedResponse.Code)
	}
}

func TestParseAllowedRepositoriesRejectsMalformedEntries(t *testing.T) {
	if _, err := parseAllowedRepositories("acme/project,invalid"); err == nil {
		t.Fatal("expected malformed allowlist entry to be rejected")
	}
	if _, err := parseAllowedRepositories(""); err == nil {
		t.Fatal("expected empty allowlist to be rejected")
	}
	allowed, err := parseAllowedRepositories(" Acme/Project,acme/project ")
	if err != nil {
		t.Fatalf("parseAllowedRepositories() error = %v", err)
	}
	if len(allowed) != 1 {
		t.Fatalf("allowlist entries = %d, want 1 normalized entry", len(allowed))
	}
}
