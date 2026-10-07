// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

func TestReferenceExternalAdapterNormalizesAdoptionEnvelope(t *testing.T) {
	adapter := ReferenceExternalAdapter{}
	request, err := adapter.NormalizeTool(map[string]any{
		"contract":      "naeos.agent-adoption",
		"version":       "1.0",
		"request_id":    "req-reference-01",
		"invocation_id": "inv-reference-01",
		"actor":         "reference-agent",
		"capability":    "repository.write",
		"tool":          "file-edit",
		"action":        "write",
		"resource":      "src/app.go",
		"environment":   "development",
		"payload":       map[string]any{"content": "example"},
		"context":       map[string]any{"repository": "example/repo", "commit": "abc123"},
	})
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}

	if request.InvocationID != "inv-reference-01" ||
		request.Tool != "file-edit" ||
		request.Action != "write" ||
		request.Capability != "repository.write" {
		t.Fatalf("normalized request lost adoption identity: %+v", request)
	}
	if request.Context["request_id"] != "req-reference-01" {
		t.Fatalf("request identity was not preserved: %+v", request.Context)
	}
}

func TestReferenceExternalAdapterRejectsUnsupportedContractBeforeGateway(t *testing.T) {
	adapter := ReferenceExternalAdapter{}
	_, err := adapter.NormalizeTool(map[string]any{
		"contract":      "naeos.agent-adoption",
		"version":       "2.0",
		"request_id":    "req-reference-02",
		"invocation_id": "inv-reference-02",
		"tool":          "deploy",
		"action":        "execute",
	})
	if err == nil {
		t.Fatal("expected unsupported major version to fail closed")
	}
}

func TestReferenceExternalAdapterUsesSameGatewayBoundary(t *testing.T) {
	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "reference-policy"}
	sb := &countingSandbox{}
	adapter := ReferenceExternalAdapter{}
	gw := New(cp, sb, WithReplayProtection(true))
	gw.RegisterAdapter(adapter.Name(), adapter)
	if err := gw.GrantAdapterPolicy(adapter.Name(), AdapterPolicy{AllowedTools: []string{"*"}, AllowedActions: []string{"*"}}); err != nil {
		t.Fatal(err)
	}

	result, err := gw.AuthorizeFromAdapter(adapter.Name(), map[string]any{
		"contract":      "naeos.agent-adoption",
		"version":       "1.0",
		"request_id":    "req-reference-03",
		"invocation_id": "inv-reference-03",
		"actor":         "reference-agent",
		"capability":    "repository.write",
		"tool":          "file-edit",
		"action":        "write",
		"resource":      "src/app.go",
	})
	if err != nil {
		t.Fatalf("governed adapter execution failed: %v", err)
	}
	if result.Decision != control.DecisionAllow || result.Status != "completed" {
		t.Fatalf("unexpected governed result: %+v", result)
	}
	if sb.Count() != 1 {
		t.Fatalf("expected exactly one gateway-controlled side effect, got %d", sb.Count())
	}

	replay, err := gw.AuthorizeFromAdapter(adapter.Name(), map[string]any{
		"contract":      "naeos.agent-adoption",
		"version":       "1.0",
		"request_id":    "req-reference-04",
		"invocation_id": "inv-reference-03",
		"actor":         "reference-agent",
		"capability":    "repository.write",
		"tool":          "file-edit",
		"action":        "write",
	})
	if err != nil {
		t.Fatalf("replay returned unexpected transport error: %v", err)
	}
	if replay.Status != "denied" || sb.Count() != 1 {
		t.Fatalf("reference adapter bypassed replay boundary: result=%+v side_effects=%d", replay, sb.Count())
	}
}
