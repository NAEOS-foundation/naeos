package controlplane

import (
	"testing"
	"time"
)

func TestApprovalStoreLifecycle(t *testing.T) {
	store := NewApprovalStore()
	created, err := store.Create(DecisionResult{
		DecisionID: "DEC-1",
		Status:     DecisionPending,
	}, "reviewer-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	if created.Status != "pending" {
		t.Fatalf("expected pending approval, got %s", created.Status)
	}
	approved, err := store.Approve(created.ID, "reviewed", "sha256:abc", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !approved.IsValid(time.Now()) {
		t.Fatal("expected approval to be valid")
	}
}

func TestApprovalStoreReject(t *testing.T) {
	store := NewApprovalStore()
	created, err := store.Create(DecisionResult{DecisionID: "DEC-REJECT", Status: DecisionPending}, "reviewer-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	rejected, err := store.Reject(created.ID, "risk too high", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if rejected.Status != "rejected" || rejected.Reason != "risk too high" {
		t.Fatalf("unexpected rejection: %+v", rejected)
	}
	if _, err := store.Consume(created.ID, time.Now()); err == nil {
		t.Fatal("expected rejected approval to be unusable")
	}
}
