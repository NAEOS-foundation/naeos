// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sandbase

import (
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

func TestAdapterEmitsBoundAuthorizationContract(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, time.UTC)
	policy := &controlplane.Policy{
		ID: "sandbase-policy", Version: 1, Status: "active",
		CreatedAt: now, UpdatedAt: now,
		AllowedCapabilities:  []controlplane.Capability{"repository.write"},
		RequiresExplicitAuth: true,
	}
	grant := &controlplane.Grant{
		GrantID: "grant-1", AgentID: "agent-1", PolicyID: policy.ID,
		PolicyVersion: 1, Capabilities: []controlplane.Capability{"repository.write"},
		CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	adapter := &Adapter{
		Gateway: controlplane.NewDecisionGateway(controlplane.NewEvaluator(), controlplane.NewLedger()),
		Policy:  policy, Grant: grant,
	}

	got, err := adapter.Authorize(AuthorizationRequest{
		SessionID: "session-1", RequestID: "req-1", AgentID: "agent-1",
		Capability: "repository.write", Target: "repo.write",
		ArtifactHash: "sha256:artifact", Context: map[string]string{"branch": "main"}, Timestamp: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Decision != "ALLOW" || got.PolicyVersion != 1 {
		t.Fatalf("unexpected decision: %+v", got)
	}
	if got.ContextDigest == "" || got.DecisionID == "" {
		t.Fatalf("missing binding fields: %+v", got)
	}
}

func TestAdapterContractDocumentsFreshnessBoundary(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, time.UTC)
	store := controlplane.NewPolicyStore()
	v1 := &controlplane.Policy{
		ID: "sandbase-policy", Version: 1, Status: "active",
		CreatedAt: now, UpdatedAt: now,
		AllowedCapabilities: []controlplane.Capability{"repository.write"},
		RequiresExplicitAuth: true,
	}
	v2 := *v1
	v2.Version = 2
	v2.UpdatedAt = now.Add(time.Minute)
	if err := store.Set(v1); err != nil {
		t.Fatal(err)
	}

	ledger := controlplane.NewLedger()
	gateway := controlplane.NewDecisionGateway(controlplane.NewEvaluator(store), ledger)
	grant := &controlplane.Grant{
		GrantID: "grant-1", AgentID: "agent-1", PolicyID: v1.ID,
		PolicyVersion: 1, Capabilities: []controlplane.Capability{"repository.write"},
		CreatedAt: now, ExpiresAt: now.Add(time.Hour), Status: "active",
	}
	adapter := &Adapter{Gateway: gateway, Policy: v1, Grant: grant}
	req := AuthorizationRequest{
		SessionID: "session-1", RequestID: "req-1", AgentID: "agent-1",
		Capability: "repository.write", Target: "repo.write",
		ArtifactHash: "sha256:artifact", Timestamp: now,
	}
	decision, err := adapter.Authorize(req)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "ALLOW" {
		t.Fatalf("expected initial ALLOW, got %+v", decision)
	}

	if err := store.Set(&v2); err != nil {
		t.Fatal(err)
	}
	_, event := gateway.ExecuteDecision(
		controlplane.AuthorizeRequest{
			RequestID: req.RequestID, AgentID: req.AgentID,
			Action: controlplane.Action{
				AgentID: req.AgentID, Capability: controlplane.Capability(req.Capability),
				ArtifactHash: req.ArtifactHash,
			},
			Grant: grant, Policy: v1, Timestamp: now.Add(2 * time.Minute),
		},
		controlplane.DecisionResult{
			DecisionID: decision.DecisionID, RequestID: decision.RequestID,
			AgentID: req.AgentID, Status: controlplane.DecisionAllow,
			Requested: controlplane.Capability(req.Capability), ArtifactHash: req.ArtifactHash,
		},
	)
	if event.EventType != "EXECUTION_BLOCKED" || event.Reason != controlplane.ReasonDeniedStalePolicy {
		t.Fatalf("expected stale authorization block, got %+v", event)
	}
}
