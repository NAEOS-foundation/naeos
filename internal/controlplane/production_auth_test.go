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

	"github.com/NAEOS-foundation/naeos/internal/api"
)

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
	return &ProductionRuntime{
		identity: productionTokenValidatorStub{claims: claims, err: validateErr},
		gateway:  gateway,
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
	claims := &api.JWTClaims{Sub: "agent-gateway", Scope: ControlPlaneAuthorizeScope}
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

func TestProductionAuthorizationRequiresScopeAndPrincipalBinding(t *testing.T) {
	reqBody := AuthorizeRequest{AgentID: "agent-gateway", Action: Action{AgentID: "agent-gateway", Capability: "artifact.read"}}
	runtime, _ := productionAuthorizationHandlerFixture(t, &api.JWTClaims{Sub: "agent-gateway", Scope: "naeos:repository:read"}, nil)
	if got := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", reqBody).Code; got != http.StatusForbidden {
		t.Fatalf("wrong scope status = %d; want 403", got)
	}

	runtime.identity = productionTokenValidatorStub{claims: &api.JWTClaims{Sub: "other-agent", Scope: ControlPlaneAuthorizeScope}}
	if got := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", reqBody).Code; got != http.StatusForbidden {
		t.Fatalf("agent mismatch status = %d; want 403", got)
	}
}

func TestProductionAuthorizationUsesVerifiedPrincipalAndRepositoryGateway(t *testing.T) {
	claims := &api.JWTClaims{Sub: "agent-gateway", Scope: ControlPlaneAuthorizeScope}
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

func TestProductionAuthorizationCapabilityDeniedByCanonicalGrant(t *testing.T) {
	runtime, req := productionAuthorizationHandlerFixture(t, &api.JWTClaims{Sub: "agent-gateway"}, nil)
	req.Action.Capability = "admin"
	response := sendProductionAuthorizationRequest(runtime.ProductionHandler(), "valid", req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("ungranted capability status = %d; want 403; body=%s", response.Code, response.Body.String())
	}
}
