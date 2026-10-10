// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sandbase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

const testBearerToken = "test-sandbase-authz-token"

func newHTTPTestAdapter(t *testing.T, allowed bool) *Adapter {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	policy := &controlplane.Policy{
		ID: "sandbase-policy", Version: 1, Status: "active",
		CreatedAt: now, UpdatedAt: now,
		RequiresExplicitAuth: true,
	}
	if allowed {
		policy.AllowedCapabilities = []controlplane.Capability{"tool.execute"}
	} else {
		policy.DeniedCapabilities = []controlplane.Capability{"tool.execute"}
	}
	store := controlplane.NewPolicyStore()
	if err := store.Set(policy); err != nil {
		t.Fatal(err)
	}
	grant := &controlplane.Grant{
		GrantID: "grant-sandbase", AgentID: "trusted-agent", PolicyID: policy.ID,
		PolicyVersion: 1, Capabilities: []controlplane.Capability{"tool.execute"},
		CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	return &Adapter{
		Gateway: controlplane.NewDecisionGateway(controlplane.NewEvaluator(store), controlplane.NewLedger()),
		Policy:  policy, Grant: grant,
	}
}

func authenticatedRequest(method string, body string) *http.Request {
	request := httptest.NewRequest(method, "/authorize", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+testBearerToken)
	return request
}

func validSandBaseRequestBody() string {
	return `{
		"schema":"sandbase.authz/v1",
		"session_id":"session-42",
		"invocation_id":"call-17",
		"capability":"tool.execute",
		"target":"repository.write",
		"arguments_digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"policy_context_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"digest_schema":"sandbase.digest/v1"
	}`
}

func TestHTTPHandlerAcceptsSandBaseAuthzV1(t *testing.T) {
	handler := NewHTTPHandler(newHTTPTestAdapter(t, true), testBearerToken)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, validSandBaseRequestBody()))

	if response.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", response.Code, response.Body.String())
	}
	var decision SandBaseAuthorizationResponse
	if err := json.Unmarshal(response.Body.Bytes(), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "allow" {
		t.Fatalf("expected lowercase allow for NAEOS ALLOW, got %+v", decision)
	}
	if decision.PolicyVersion != "1" || decision.DecisionID == "" {
		t.Fatalf("missing decision binding fields: %+v", decision)
	}
	if decision.ContextDigest != strings.Repeat("b", 64) {
		t.Fatalf("policy context digest was not preserved: %+v", decision)
	}
}

func TestHTTPHandlerFailsClosedOnNAEOSDeny(t *testing.T) {
	handler := NewHTTPHandler(newHTTPTestAdapter(t, false), testBearerToken)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, validSandBaseRequestBody()))

	if response.Code != http.StatusOK {
		t.Fatalf("expected a readable veto decision, got HTTP %d", response.Code)
	}
	var decision SandBaseAuthorizationResponse
	if err := json.Unmarshal(response.Body.Bytes(), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "deny" {
		t.Fatalf("NAEOS denial must never be translated to allow: %+v", decision)
	}
}

func TestHTTPHandlerRejectsInvalidDigestAndUnknownSchema(t *testing.T) {
	handler := NewHTTPHandler(newHTTPTestAdapter(t, true), testBearerToken)
	body := strings.Replace(validSandBaseRequestBody(), strings.Repeat("a", 64), "not-a-digest", 1)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, body))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected malformed digest to fail closed with HTTP 400, got %d", response.Code)
	}

	body = strings.Replace(validSandBaseRequestBody(), "sandbase.authz/v1", "sandbase.authz/v2", 1)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, body))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected unknown schema to be rejected, got %d", response.Code)
	}
}

func TestHTTPHandlerRejectsMissingOrIncorrectBearerToken(t *testing.T) {
	handler := NewHTTPHandler(newHTTPTestAdapter(t, true), testBearerToken)
	request := httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(validSandBaseRequestBody()))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing token to be rejected, got %d", response.Code)
	}

	request = authenticatedRequest(http.MethodPost, validSandBaseRequestBody())
	request.Header.Set("Authorization", "Bearer wrong-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected incorrect token to be rejected, got %d", response.Code)
	}
}

func TestHTTPHandlerRejectsEmptyConfiguredTokenAndNonPOST(t *testing.T) {
	handler := NewHTTPHandler(newHTTPTestAdapter(t, true), "")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodPost, validSandBaseRequestBody()))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected missing server token to fail closed, got %d", response.Code)
	}

	handler = NewHTTPHandler(newHTTPTestAdapter(t, true), testBearerToken)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, authenticatedRequest(http.MethodGet, ""))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected HTTP 405, got %d", response.Code)
	}
}
