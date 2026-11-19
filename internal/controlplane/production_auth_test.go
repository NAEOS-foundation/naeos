// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/api"
)

type productionBindingResolverStub struct{}

func (productionBindingResolverStub) ResolveAgent(_ context.Context, issuer, kind, principalID, requestedAgentID string) (string, error) {
	if issuer != "https://issuer.example" || (kind != "human" && kind != "service") {
		return "", ErrRepositoryNotFound
	}
	if principalID == "user:multi" && requestedAgentID == "" {
		return "", ErrRepositoryConflict
	}
	allowed := map[string]string{
		"agent-gateway": "agent-gateway",
		"user:123":      "agent-gateway",
		" user:123 ":    " agent-gateway ",
		"user:multi":    "agent-a",
	}
	agentID, ok := allowed[principalID]
	if !ok || (requestedAgentID != "" && requestedAgentID != agentID) {
		return "", ErrRepositoryNotFound
	}
	return agentID, nil
}

type productionTokenValidatorStub struct {
	claims *api.JWTClaims
	err    error
}

func (v productionTokenValidatorStub) Validate(context.Context, string) (*api.JWTClaims, error) {
	return v.claims, v.err
}

func productionAuthorizationHandlerFixture(t *testing.T, claims *api.JWTClaims, validateErr error) (*ProductionRuntime, AuthorizeRequest) {
	t.Helper()
	gateway, req := repositoryGatewayFixture(t)
	if claims != nil && claims.Scope == "" && len(claims.Scopes) == 0 {
		claims.Scope = decisionWriteScope
	}
	return &ProductionRuntime{
		identity:      productionTokenValidatorStub{claims: claims, err: validateErr},
		gateway:       gateway,
		bindings:      productionBindingResolverStub{},
		principalKind: "service",
	}, req
}

func sendProductionAuthorizationRequest(handler http.Handler, token string, request AuthorizeRequest) *httptest.ResponseRecorder {
	body, _ := json.Marshal(request)
	httpRequest := httptest.NewRequest(http.MethodPost, "/api/control-plane/decision", strings.NewReader(string(body)))
	if token != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httpRequest)
	return response
}

func TestProductionAuthorizationRequiresBearerAndValidIdentity(t *testing.T) {
	claims := &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway"}
	runtime, req := productionAuthorizationHandlerFixture(t, claims, nil)
	handler := runtime.ProductionHandler()

	if got := sendProductionAuthorizationRequest(handler, "", req).Code; got != http.StatusUnauthorized {
		t.Fatalf("missing bearer status = %d; want 401", got)
	}
	runtime.identity = productionTokenValidatorStub{err: errors.New("signature invalid")}
	if got := sendProductionAuthorizationRequest(handler, "invalid", req).Code; got != http.StatusUnauthorized {
		t.Fatalf("invalid token status = %d; want 401", got)
	}
}

func TestProductionAuthorizationRequiresDecisionWriteScope(t *testing.T) {
	request := AuthorizeRequest{AgentID: "agent-gateway", Action: Action{AgentID: "agent-gateway", Capability: "artifact.read"}}
	tests := []struct {
		name   string
		claims *api.JWTClaims
		want   int
	}{
		{name: "missing scope", claims: &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway", Scope: " "}, want: http.StatusForbidden},
		{name: "wrong scope", claims: &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway", Scope: "naeos:control-plane:runs:read"}, want: http.StatusForbidden},
		{name: "substring is not a scope", claims: &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway", Scope: "naeos:control-plane:decision:write-extra"}, want: http.StatusForbidden},
		{name: "padded scopes array value is not exact", claims: &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway", Scopes: []string{" naeos:control-plane:decision:write "}}, want: http.StatusForbidden},
		{name: "exact scope token", claims: &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway", Scope: "openid profile naeos:control-plane:decision:write"}, want: http.StatusOK},
		{name: "scopes array", claims: &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway", Scopes: []string{"profile", "naeos:control-plane:decision:write"}}, want: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime, _ := productionAuthorizationHandlerFixture(t, test.claims, nil)
			response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", request)
			if response.Code != test.want {
				t.Fatalf("status = %d; want %d; body=%s", response.Code, test.want, response.Body.String())
			}
		})
	}
}

func TestProductionAuthorizationRejectsMissingScopeBeforeReadingBody(t *testing.T) {
	runtime, _ := productionAuthorizationHandlerFixture(t, &api.JWTClaims{
		Issuer: "https://issuer.example", Sub: "agent-gateway", Scope: "naeos:control-plane:runs:read",
	}, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/control-plane/decision", strings.NewReader("{not-json"))
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	runtime.ProductionHandler().ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want 403 before JSON parsing; body=%s", response.Code, response.Body.String())
	}
}

func TestProductionAuthorizationRequiresPrincipalBinding(t *testing.T) {
	reqBody := AuthorizeRequest{AgentID: "agent-gateway", Action: Action{AgentID: "agent-gateway", Capability: "artifact.read"}}
	runtime, _ := productionAuthorizationHandlerFixture(t, &api.JWTClaims{Issuer: "https://issuer.example", Sub: "other-agent"}, nil)
	if got := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", reqBody).Code; got != http.StatusForbidden {
		t.Fatalf("agent mismatch status = %d; want 403", got)
	}
}

func TestProductionAuthorizationResolvesSubjectWithoutUsingClientIDAsAgent(t *testing.T) {
	claims := &api.JWTClaims{
		Issuer:   "https://issuer.example",
		Sub:      "user:123",
		ClientID: "untrusted-agent-override",
	}
	runtime, req := productionAuthorizationHandlerFixture(t, claims, nil)
	runtime.principalKind = "human"
	req.AgentID = ""
	req.Action.AgentID = ""
	response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", req)
	if response.Code != http.StatusOK {
		t.Fatalf("authorization status = %d; want 200; body=%s", response.Code, response.Body.String())
	}
	var result DecisionResult
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.AgentID != "agent-gateway" {
		t.Fatalf("result agent = %q; want persisted binding agent-gateway, not client_id %q", result.AgentID, claims.ClientID)
	}
}

func TestProductionAuthorizationPreservesOpaqueSubjectAndRequestedAgentID(t *testing.T) {
	claims := &api.JWTClaims{
		Issuer:   "https://issuer.example",
		Sub:      " user:123 ",
		ClientID: "different-client",
	}
	runtime, req := productionAuthorizationHandlerFixture(t, claims, nil)
	runtime.principalKind = "human"
	req.AgentID = " agent-gateway "
	req.Action.AgentID = " agent-gateway "

	response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", req)
	if response.Code != http.StatusOK {
		t.Fatalf("authorization status = %d; want 200; body=%s", response.Code, response.Body.String())
	}
	var result DecisionResult
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.AgentID != " agent-gateway " {
		t.Fatalf("agent ID = %q; want exact bound identifier", result.AgentID)
	}
}

func TestProductionAuthorizationRejectsAmbiguousPrincipalBinding(t *testing.T) {
	claims := &api.JWTClaims{Issuer: "https://issuer.example", Sub: "user:multi"}
	runtime, req := productionAuthorizationHandlerFixture(t, claims, nil)
	runtime.principalKind = "human"
	req.AgentID = ""
	req.Action.AgentID = ""
	response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("ambiguous binding status = %d; want 403; body=%s", response.Code, response.Body.String())
	}
}

func TestProductionAuthorizationUsesVerifiedPrincipalAndRepositoryGateway(t *testing.T) {
	claims := &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway"}
	runtime, req := productionAuthorizationHandlerFixture(t, claims, nil)
	// Caller-supplied grant/policy values are explicitly not authoritative.
	req.Grant = &Grant{GrantID: "forged", AgentID: "attacker", Status: "active", Capabilities: []Capability{"admin"}}
	req.Policy = &Policy{ID: "forged", Version: 999, Status: "active", AllowedCapabilities: []Capability{"admin"}}
	response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", req)
	if response.Code != http.StatusOK {
		t.Fatalf("authorization status = %d; want 200; body=%s", response.Code, response.Body.String())
	}
	var result DecisionResult
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Status != DecisionAllow || result.AgentID != claims.Sub || result.DecisionID == "" {
		t.Fatalf("result = %+v; want repository-backed ALLOW bound to token principal", result)
	}
}

func TestProductionHandlerReadinessRequiresIdentityConfiguration(t *testing.T) {
	var runtime *ProductionRuntime
	handler := runtime.ProductionHandler()
	ready := httptest.NewRecorder()
	handler.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil runtime readiness = %d; want 503", ready.Code)
	}
	live := httptest.NewRecorder()
	handler.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("liveness = %d; want 200", live.Code)
	}
}

func TestProductionHandlerDoesNotExposeDemoRoutes(t *testing.T) {
	runtime, _ := productionAuthorizationHandlerFixture(t, nil, nil)
	handler := runtime.ProductionHandler()
	for _, path := range []string{"/api/reset", "/api/scenarios", "/api/investor-demo"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("route %q returned %d; want 404", path, response.Code)
		}
	}
}

func TestProductionAuthorizationUsesServerControlledTimestamp(t *testing.T) {
	runtime, req := productionAuthorizationHandlerFixture(t, &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway"}, nil)
	req.Timestamp = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", req)
	if response.Code != http.StatusOK {
		t.Fatalf("authorization status = %d; want 200; body=%s", response.Code, response.Body.String())
	}
	ledger, ok := runtime.gateway.Ledger.(*repositoryGatewayTestLedger)
	if !ok {
		t.Fatalf("ledger type = %T; want repositoryGatewayTestLedger", runtime.gateway.Ledger)
	}
	events, err := ledger.LoadVerified(context.Background())
	if err != nil {
		t.Fatalf("LoadVerified: %v", err)
	}
	if len(events) != 1 || events[0].Timestamp.Before(time.Now().UTC().Add(-5*time.Second)) {
		t.Fatalf("ledger event timestamp = %+v; caller-supplied past timestamp must be ignored", events)
	}
}

func TestProductionAuthorizationCapabilityDeniedByCanonicalGrant(t *testing.T) {
	runtime, req := productionAuthorizationHandlerFixture(t, &api.JWTClaims{Issuer: "https://issuer.example", Sub: "agent-gateway"}, nil)
	req.Action.Capability = "admin"
	response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("ungranted capability status = %d; want 403; body=%s", response.Code, response.Body.String())
	}
}
