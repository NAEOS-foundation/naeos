package controlplane

import (
	"testing"
	"time"
)

func TestSessionVerifier_VerifySessionPassesWhenLedgerIsClean(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                  "POLICY-ALLOWED",
		Version:             1,
		Status:              "active",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		AllowedCapabilities: []Capability{"repository.read"},
	}
	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}
	ledger := NewLedger()
	ledger.Append(LedgerEvent{
		AgentID:    "agent-1",
		Capability: "repository.read",
		EventType:  "EXECUTION_ALLOWED",
		Decision:   DecisionAllow,
		Reason:     ReasonAllowed,
	})
	verifier := NewSessionVerifier(ledger, NewEvaluator(store))
	result := verifier.VerifySession("agent-1")
	if result.Result != "PASS" {
		t.Fatalf("expected PASS, got %s", result.Result)
	}
	if !result.PolicyCompliant {
		t.Fatalf("expected policy compliance true")
	}
}

func TestSessionVerifier_RejectsDeniedLedgerEvents(t *testing.T) {
	store := NewPolicyStore()
	policy := &Policy{
		ID:                    "POLICY-SECURE",
		Version:               1,
		Status:                "active",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		AllowedCapabilities:   []Capability{"repository.read"},
		ProtectedCapabilities: []Capability{"credential.rotate"},
	}
	if err := store.Set(policy); err != nil {
		t.Fatalf("store.Set() error = %v", err)
	}
	ledger := NewLedger()
	ledger.Append(LedgerEvent{
		AgentID:    "agent-2",
		Capability: "credential.rotate",
		EventType:  "EXECUTION_BLOCKED",
		Decision:   DecisionDeny,
		Reason:     ReasonDeniedProtected,
	})
	verifier := NewSessionVerifier(ledger, NewEvaluator(store))
	result := verifier.VerifySession("agent-2")
	if result.Result != "PASS" {
		t.Fatalf("expected PASS for prevented attack, got %s", result.Result)
	}
	if result.BlockedAttempts == 0 {
		t.Fatal("expected blocked attempt count > 0")
	}
}
