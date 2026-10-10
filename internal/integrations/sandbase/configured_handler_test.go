// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sandbase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

func writeRuntimeConfig(t *testing.T, path string, version int, grantVersion int, allowed bool) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	policy := controlplane.Policy{
		ID: "sandbase-policy", Version: version, Status: "active",
		CreatedAt: now, UpdatedAt: now,
		AllowedCapabilities: []controlplane.Capability{},
	}
	if allowed {
		policy.AllowedCapabilities = []controlplane.Capability{"tool.execute"}
	}
	config := RuntimeConfig{
		Policy: policy,
		Grant: controlplane.Grant{
			GrantID: "grant-sandbase", AgentID: "trusted-agent",
			PolicyID: policy.ID, PolicyVersion: grantVersion,
			Capabilities: []controlplane.Capability{"tool.execute"},
			CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
		},
	}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func configuredRequest(handler http.Handler) *httptest.ResponseRecorder {
	request := authenticatedRequest(http.MethodPost, validSandBaseRequestBody())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestConfiguredHTTPHandlerReloadsPolicyAndFailsClosedOnVersionReuse(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "sandbase-authz.json")
	ledgerPath := filepath.Join(t.TempDir(), "ledger.json")
	writeRuntimeConfig(t, configPath, 1, 1, true)

	handler, err := NewConfiguredHTTPHandler(configPath, testBearerToken, ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	first := configuredRequest(handler)
	if first.Code != http.StatusOK {
		t.Fatalf("initial authorization failed: status=%d body=%s", first.Code, first.Body.String())
	}
	var firstDecision SandBaseAuthorizationResponse
	if err := json.Unmarshal(first.Body.Bytes(), &firstDecision); err != nil {
		t.Fatal(err)
	}
	if firstDecision.Decision != "allow" {
		t.Fatalf("expected initial allow, got %+v", firstDecision)
	}

	// A new active policy version while the grant remains on v1 must veto.
	writeRuntimeConfig(t, configPath, 2, 1, true)
	second := configuredRequest(handler)
	if second.Code != http.StatusOK {
		t.Fatalf("expected a readable stale-grant decision: status=%d body=%s", second.Code, second.Body.String())
	}
	var secondDecision SandBaseAuthorizationResponse
	if err := json.Unmarshal(second.Body.Bytes(), &secondDecision); err != nil {
		t.Fatal(err)
	}
	if secondDecision.Decision != "reauthorize" || secondDecision.PolicyVersion != "2" {
		t.Fatalf("expected reauthorize for active policy v2 and grant v1, got %+v", secondDecision)
	}

	// Policy contents are immutable within a version; an unversioned edit fails closed.
	writeRuntimeConfig(t, configPath, 2, 2, false)
	third := configuredRequest(handler)
	if third.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected policy version reuse to fail closed, got status=%d body=%s", third.Code, third.Body.String())
	}
	var thirdDecision SandBaseAuthorizationResponse
	if err := json.Unmarshal(third.Body.Bytes(), &thirdDecision); err != nil {
		t.Fatal(err)
	}
	if thirdDecision.Decision != "deny" {
		t.Fatalf("expected fail-closed deny, got %+v", thirdDecision)
	}

	if _, err := os.Stat(ledgerPath); err != nil {
		t.Fatalf("expected authorization ledger snapshot to be persisted: %v", err)
	}
}
