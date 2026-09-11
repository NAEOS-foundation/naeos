package investordemo

import (
	"testing"
	"time"
)

func TestGrantStore_GetGrant(t *testing.T) {
	store := NewGrantStore()
	if _, err := store.GetGrant("missing"); err == nil {
		t.Error("expected error for missing grant")
	}

	if err := store.StoreGrant(&CapabilityGrant{
		GrantID: "GRANT-X", AgentID: "agent-x", PolicyID: "POLICY-017", PolicyVersion: 12,
		Capabilities: []Capability{"repository.read"}, Status: "active", ExpiresAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.StoreGrant(&CapabilityGrant{GrantID: "", AgentID: "a"}); err == nil {
		t.Error("expected error for empty grant ID")
	}
	if err := store.StoreGrant(&CapabilityGrant{GrantID: "G1", AgentID: ""}); err == nil {
		t.Error("expected error for empty agent ID")
	}

	got, err := store.GetGrant("GRANT-X")
	if err != nil {
		t.Fatal(err)
	}
	if got.GrantID != "GRANT-X" || got.AgentID != "agent-x" {
		t.Errorf("unexpected grant: %+v", got)
	}
}

func TestGrantStore_RevokeAndList(t *testing.T) {
	store := NewGrantStore()
	now := time.Now()

	g := &CapabilityGrant{
		GrantID: "GRANT-R", AgentID: "agent-r", PolicyID: "POLICY-017", PolicyVersion: 12,
		Capabilities: []Capability{"repository.read"}, Status: "active", ExpiresAt: now.Add(time.Hour),
	}
	if err := store.StoreGrant(g); err != nil {
		t.Fatal(err)
	}

	if err := store.RevokeGrant("GRANT-R", "policy updated", &now); err != nil {
		t.Fatal(err)
	}
	if !g.Revoked || g.Status != "revoked" || g.RevokeReason != "policy updated" {
		t.Errorf("unexpected revoked grant state: %+v", g)
	}
	if g.IsValid() {
		t.Error("revoked grant should not be valid")
	}

	if err := store.RevokeGrant("missing", "why", &now); err == nil {
		t.Error("expected error revoking missing grant")
	}

	list := store.ListGrants()
	if len(list) != 1 {
		t.Errorf("expected 1 grant, got %d", len(list))
	}

	byAgent, err := store.GetGrantByAgent("agent-r")
	if err != nil {
		t.Fatal(err)
	}
	if byAgent.GrantID != "GRANT-R" {
		t.Errorf("expected GRANT-R, got %s", byAgent.GrantID)
	}
	if _, err := store.GetGrantByAgent("nobody"); err == nil {
		t.Error("expected error for unknown agent")
	}
}

func TestGrantStore_ExpiredGrant(t *testing.T) {
	store := NewGrantStore()
	g := &CapabilityGrant{
		GrantID: "GRANT-E", AgentID: "agent-e", PolicyID: "POLICY-017", PolicyVersion: 12,
		Capabilities: []Capability{"repository.read"}, Status: "active",
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	if err := store.StoreGrant(g); err != nil {
		t.Fatal(err)
	}
	if g.IsValid() {
		t.Error("expired grant should not be valid")
	}
}

func TestAuditLedger_Filtering(t *testing.T) {
	ledger := NewAuditLedger()
	_ = ledger.RecordEvent(&AuditEvent{EventID: "A1", AgentID: "agent-a", EventType: "AUTHORIZATION_GRANTED", Timestamp: time.Now()})
	_ = ledger.RecordEvent(&AuditEvent{EventID: "A2", AgentID: "agent-a", EventType: "EXECUTION_BLOCKED", Timestamp: time.Now()})
	_ = ledger.RecordEvent(&AuditEvent{EventID: "A3", AgentID: "agent-b", EventType: "AUTHORIZATION_GRANTED", Timestamp: time.Now()})

	if got := ledger.GetEventsByAgent("agent-a"); len(got) != 2 {
		t.Errorf("expected 2 events for agent-a, got %d", len(got))
	}
	if got := ledger.GetEventsByType("AUTHORIZATION_GRANTED"); len(got) != 2 {
		t.Errorf("expected 2 granted events, got %d", len(got))
	}
	if got := ledger.CountEventsByType("EXECUTION_BLOCKED"); got != 1 {
		t.Errorf("expected 1 blocked event, got %d", got)
	}
	if got := ledger.CountEventsByType("NOTHING"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestPolicyEngine_APISurface(t *testing.T) {
	setup := SetupDemoEnvironment()

	ps := NewPolicyStore()
	_ = ps.StorePolicy(&Policy{
		PolicyID: "POLICY-APPR", Version: 3, Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
		RequiresExplicitAuth:  true,
		ApprovalRequired:      []Capability{"production.deploy"},
		AllowedCapabilities:   []Capability{"repository.read", "production.deploy"},
		ProtectedCapabilities: []Capability{"policy.modify"},
	})
	pe := NewPolicyEngine(ps, NewAuditLedger())

	if approved, err := pe.RequiresApproval("POLICY-APPR", 3, "repository.read"); err != nil || approved {
		t.Errorf("expected no approval needed for repository.read (approved=%v err=%v)", approved, err)
	}
	approved, err := pe.RequiresApproval("POLICY-APPR", 3, "production.deploy")
	if err != nil || !approved {
		t.Errorf("expected approval needed for production.deploy (approved=%v err=%v)", approved, err)
	}
	if _, err := pe.RequiresApproval("POLICY-APPR", 99, "x"); err == nil {
		t.Error("expected error for unknown policy version")
	}

	if !pe.EnforcesExplicitAuth("POLICY-APPR", 3) {
		t.Error("expected explicit auth enforced")
	}
	_ = ps.StorePolicy(&Policy{
		PolicyID: "POLICY-LAX", Version: 1, Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if pe.EnforcesExplicitAuth("POLICY-LAX", 1) {
		t.Error("expected no explicit auth enforcement for lax policy")
	}
	if pe.EnforcesExplicitAuth("POLICY-APPR", 99) {
		t.Error("expected false for missing version")
	}

	// Revocation via the capability authority surface.
	grant, err := setup.GrantStore.GetGrantByAgent("agent-payment-01")
	if err != nil {
		t.Fatal(err)
	}
	if err := setup.CapabilityAuthority.RevokeGrant(grant.GrantID, "audit remediation"); err != nil {
		t.Fatal(err)
	}
	if !grant.Revoked {
		t.Error("expected grant to be revoked via capability authority")
	}
	if err := setup.CapabilityAuthority.RevokeGrant("missing-grant", "why"); err == nil {
		t.Error("expected error for missing grant")
	}
}

func TestRunAllScenarios(t *testing.T) {
	setup := SetupDemoEnvironment()
	results := RunAllScenarios(setup)
	if len(results) != 6 {
		t.Fatalf("expected 6 scenario results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Passed {
			t.Errorf("scenario %s failed: actual=%s expected=%s details=%s",
				r.ScenarioName, r.ActualResult, r.ExpectedResult, r.Details)
		}
	}
}
