package controlplane

import (
	"testing"
	"time"
)

func TestApprovalStoreLifecycle(t *testing.T) {
	store := NewApprovalStore()
	created, err := store.Create("DEC-1", "reviewer-1", time.Now().Add(time.Hour))
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
