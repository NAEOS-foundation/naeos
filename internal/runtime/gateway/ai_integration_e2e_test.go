// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

func TestCodexGovernedExecutionEndToEnd(t *testing.T) {
	root := t.TempDir()

	reg := policy.NewRegistry()
	if err := reg.Register(&policy.Policy{
		ID:      "repository-write",
		Name:    "Repository Write",
		Version: "1.0.0",
		Scope: policy.Scope{
			Resource:    "repository",
			Action:      "write",
			Environment: "development",
		},
		Default: policy.DecisionAllow,
		Active:  true,
	}); err != nil {
		t.Fatalf("register policy v1: %v", err)
	}

	cp := control.New(reg)
	sb := NewDefaultSandbox(SandboxConfig{FilesystemRoot: root})
	gw := New(cp, sb)
	adapter := CodexToolAdapter{}
	gw.RegisterAdapter(adapter.Name(), adapter)

	raw := map[string]any{
		"type": "function_call",
		"name": "filesystem",
		"arguments": map[string]any{
			"capability":  "repository.write",
			"action":      "write",
			"resource":    "repository",
			"environment": "development",
			"actor":       "codex",
			"payload": map[string]any{
				"path":    "src/main.go",
				"content": "package main\n",
			},
			"context": map[string]any{
				"task_id": "e2e-codex-001",
			},
		},
	}

	result, err := gw.AuthorizeFromAdapter("codex", raw)
	if err != nil {
		t.Fatalf("authorized Codex request failed: %v", err)
	}
	if result.Decision != control.DecisionAllow {
		t.Fatalf("expected ALLOW, got %s", result.Decision)
	}
	if result.PolicyID != "repository-write" {
		t.Fatalf("expected repository-write policy, got %s", result.PolicyID)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed execution, got %s", result.Status)
	}

	target := filepath.Join(root, "src", "main.go")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected governed execution to create %s: %v", target, err)
	}
	if string(data) != "package main\n" {
		t.Fatalf("unexpected executed artifact: %q", string(data))
	}

	// A policy mutation must invalidate the previously valid authorization at
	// the execution boundary. The same agent request is therefore re-evaluated
	// against policy v2 and must not produce a second side effect.
	if err := reg.Register(&policy.Policy{
		ID:      "repository-write",
		Name:    "Repository Write",
		Version: "2.0.0",
		Scope: policy.Scope{
			Resource:    "repository",
			Action:      "write",
			Environment: "development",
		},
		Default: policy.DecisionDeny,
		Active:  true,
	}); err != nil {
		t.Fatalf("register policy v2: %v", err)
	}

	result, err = gw.AuthorizeFromAdapter("codex", raw)
	if err != nil {
		t.Fatalf("expected fail-closed denial without gateway error, got %v", err)
	}
	if result.Decision != control.DecisionDeny {
		t.Fatalf("expected DENY after policy rotation, got %s", result.Decision)
	}
	if result.Status != "denied" {
		t.Fatalf("expected denied status after policy rotation, got %s", result.Status)
	}
	if got := sb.ExecutedCount(); got != 2 {
		t.Fatalf("expected gateway to enter sandbox only for the two authorization attempts, got %d", got)
	}

	data, err = os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected original artifact to remain readable: %v", err)
	}
	if string(data) != "package main\n" {
		t.Fatalf("denied replay changed the artifact: %q", string(data))
	}

	history := gw.History()
	if len(history) != 2 {
		t.Fatalf("expected two durable gateway results, got %d", len(history))
	}
	if history[0].PolicyID != "repository-write" || history[0].Status != "completed" {
		t.Fatalf("unexpected first evidence: %+v", history[0])
	}
	if history[1].Decision != control.DecisionDeny || history[1].Status != "denied" {
		t.Fatalf("unexpected second evidence: %+v", history[1])
	}
}

func TestCodexGovernedExecutionRejectsCapabilityEscalationEndToEnd(t *testing.T) {
	reg := policy.NewRegistry()
	if err := reg.Register(&policy.Policy{
		ID:      "repository-access",
		Name:    "Repository Access",
		Version: "1.0.0",
		Scope: policy.Scope{
			Resource:    "repository",
			Environment: "development",
		},
		Default: policy.DecisionAllow,
		Active:  true,
	}); err != nil {
		t.Fatalf("register policy: %v", err)
	}

	cp := control.New(reg)
	issued, err := cp.Evaluate(control.Request{
		Capability:  "repository.read",
		Resource:    "repository",
		Action:      "read",
		Environment: "development",
		Actor:       "codex",
	})
	if err != nil {
		t.Fatalf("issue initial authorization: %v", err)
	}

	// A decision for read cannot be replayed as authority for write, even
	// though the same policy independently allows a fresh write request.
	current, err := cp.ValidateDecision(control.Request{
		Capability:  "repository.write",
		Resource:    "repository",
		Action:      "write",
		Environment: "development",
		Actor:       "codex",
	}, issued)
	if err == nil {
		t.Fatalf("expected capability escalation to be rejected, got current=%+v", current)
	}
}

func TestCodexGovernedExecutionJSONEnvelopeEndToEnd(t *testing.T) {
	reg := policy.NewRegistry()
	if err := reg.Register(&policy.Policy{
		ID:      "repository-read",
		Name:    "Repository Read",
		Version: "1.0.0",
		Scope: policy.Scope{
			Resource:    "repository",
			Action:      "read",
			Environment: "development",
		},
		Default: policy.DecisionAllow,
		Active:  true,
	}); err != nil {
		t.Fatalf("register policy: %v", err)
	}

	cp := control.New(reg)
	sb := NewDefaultSandbox(SandboxConfig{})
	gw := New(cp, sb)
	gw.RegisterAdapter("codex", CodexToolAdapter{})

	envelope := map[string]any{
		"type": "function_call",
		"name": "repository",
		"arguments": map[string]any{
			"capability":  "repository.read",
			"action":      "read",
			"resource":    "repository",
			"environment": "development",
			"actor":       "codex",
		},
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal Codex envelope: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode Codex envelope: %v", err)
	}

	result, err := gw.AuthorizeFromAdapter("codex", decoded)
	if err != nil {
		t.Fatalf("JSON-envelope request failed: %v", err)
	}
	if result.Decision != control.DecisionAllow || result.Status != "completed" {
		t.Fatalf("expected governed completion, got decision=%s status=%s", result.Decision, result.Status)
	}
}
