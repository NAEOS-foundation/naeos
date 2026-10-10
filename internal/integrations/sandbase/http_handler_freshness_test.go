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

func TestHTTPHandlerRequestsReauthorizationAfterPolicyVersionChanges(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	v1 := &controlplane.Policy{
		ID: "sandbase-policy", Version: 1, Status: "active",
		CreatedAt: now, UpdatedAt: now,
		AllowedCapabilities: []controlplane.Capability{"tool.execute"},
	}
	store := controlplane.NewPolicyStore()
	if err := store.Set(v1); err != nil {
		t.Fatal(err)
	}
	grant := &controlplane.Grant{
		GrantID: "grant-sandbase", AgentID: "trusted-agent", PolicyID: v1.ID,
		PolicyVersion: 1, Capabilities: []controlplane.Capability{"tool.execute"},
		CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	adapter := &Adapter{
		Gateway: controlplane.NewDecisionGateway(controlplane.NewEvaluator(store), controlplane.NewLedger()),
		Policy:  v1, Grant: grant,
	}

	v2 := *v1
	v2.Version = 2
	v2.UpdatedAt = now.Add(time.Second)
	if err := store.Set(&v2); err != nil {
		t.Fatal(err)
	}

	handler := NewHTTPHandler(adapter, testBearerToken)
	request := httptest.NewRequest(http.MethodPost, "/authorize", strings.NewReader(validSandBaseRequestBody()))
	request.Header.Set("Authorization", "Bearer "+testBearerToken)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected a readable reauthorization decision, got HTTP %d: %s", response.Code, response.Body.String())
	}

	var decision SandBaseAuthorizationResponse
	if err := json.Unmarshal(response.Body.Bytes(), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "reauthorize" {
		t.Fatalf("stale grant must request reauthorization, got %+v", decision)
	}
	if decision.PolicyVersion != "2" {
		t.Fatalf("response should identify active policy v2, got %+v", decision)
	}
}
