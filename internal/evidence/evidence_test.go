package evidence

import (
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

func TestAppendAndRetrieve(t *testing.T) {
	store := NewStore()
	rec, err := store.Append(EvidenceRecord{
		Actor:    "ci-bot",
		Resource: "deploy",
		Action:   "run",
		Decision: control.DecisionAllow,
		PolicyID: "p1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID == "" {
		t.Fatal("expected ID to be generated")
	}
	if rec.Hash == "" {
		t.Fatal("expected hash to be computed")
	}
	if rec.PreviousHash != "" {
		t.Fatal("expected first record to have empty previous_hash")
	}
	if store.Len() != 1 {
		t.Fatalf("expected 1 record, got %d", store.Len())
	}
}

func TestChainIntegrity(t *testing.T) {
	store := NewStore()
	for i := 0; i < 5; i++ {
		_, err := store.Append(EvidenceRecord{
			Actor:    "user",
			Resource: "deploy",
			Action:   "run",
			Decision: control.DecisionAllow,
			PolicyID: "p1",
		})
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
	}

	idx, err := store.Verify()
	if err != nil {
		t.Fatalf("chain verification failed at index %d: %v", idx, err)
	}
	if idx != -1 {
		t.Fatalf("expected chain intact (idx=-1), got %d", idx)
	}
}

func TestChainTamperDetection(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "a1", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a2", Decision: control.DecisionDeny, PolicyID: "p2"})

	// Tamper with the first record's hash.
	store.mu.Lock()
	store.records[0].Hash = "tampered"
	store.mu.Unlock()

	idx, err := store.Verify()
	if err == nil {
		t.Fatal("expected verification error after tampering")
	}
	if idx != 0 {
		t.Fatalf("expected tamper at index 0, got %d", idx)
	}
}

func TestChainPreviousHashLinking(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "a1", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a2", Decision: control.DecisionDeny, PolicyID: "p2"})

	// Break the chain by changing the second record's previous_hash.
	store.mu.Lock()
	store.records[1].PreviousHash = "broken"
	store.mu.Unlock()

	idx, err := store.Verify()
	if err == nil {
		t.Fatal("expected verification error after chain break")
	}
	if idx != 1 {
		t.Fatalf("expected break at index 1, got %d", idx)
	}
}

func TestQueryByActor(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "alice", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "bob", Decision: control.DecisionDeny, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "alice", Decision: control.DecisionAllow, PolicyID: "p1"})

	records := store.ByActor("alice")
	if len(records) != 2 {
		t.Fatalf("expected 2 records for alice, got %d", len(records))
	}
}

func TestQueryByResource(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "a", Resource: "deploy", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Resource: "build", Decision: control.DecisionDeny, PolicyID: "p1"})

	records := store.ByResource("deploy")
	if len(records) != 1 {
		t.Fatalf("expected 1 record for deploy, got %d", len(records))
	}
}

func TestQueryByDecision(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionDeny, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionRequireApproval, PolicyID: "p1"})

	if len(store.Denied()) != 1 {
		t.Fatalf("expected 1 denied, got %d", len(store.Denied()))
	}
}

func TestQueryComposite(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "alice", Resource: "deploy", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "alice", Resource: "build", Decision: control.DecisionDeny, PolicyID: "p2"})
	store.Append(EvidenceRecord{Actor: "bob", Resource: "deploy", Decision: control.DecisionAllow, PolicyID: "p1"})

	results := store.Query(EvidenceQuery{Actor: "alice", Resource: "deploy"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].PolicyID != "p1" {
		t.Fatalf("expected policy p1, got %s", results[0].PolicyID)
	}
}

func TestQueryTimeRange(t *testing.T) {
	store := NewStore()
	now := time.Now().UTC()
	store.Append(EvidenceRecord{Actor: "a", Timestamp: now.Add(-2 * time.Hour), Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Timestamp: now.Add(-1 * time.Hour), Decision: control.DecisionDeny, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Timestamp: now, Decision: control.DecisionAllow, PolicyID: "p1"})

	results := store.Query(EvidenceQuery{
		From: now.Add(-90 * time.Minute),
		To:   now.Add(-30 * time.Minute),
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result in time range, got %d", len(results))
	}
}

func TestQueryLimit(t *testing.T) {
	store := NewStore()
	for i := 0; i < 10; i++ {
		store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionAllow, PolicyID: "p1"})
	}

	results := store.Query(EvidenceQuery{Limit: 3})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestApprovals(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{
		Actor:    "alice",
		Decision: control.DecisionAllow,
		PolicyID: "p1",
		Approval: &ApprovalRecord{
			Approver:     "admin",
			ArtifactHash: "abc123",
			ArtifactName: "deploy.tar.gz",
			Timestamp:    time.Now().UTC(),
			Valid:        true,
		},
	})
	store.Append(EvidenceRecord{
		Actor:    "bob",
		Decision: control.DecisionDeny,
		PolicyID: "p1",
	})

	approvals := store.Approvals()
	if len(approvals) != 1 {
		t.Fatalf("expected 1 approval, got %d", len(approvals))
	}
	if approvals[0].Approval.Approver != "admin" {
		t.Fatalf("expected approver admin, got %s", approvals[0].Approval.Approver)
	}
}

func TestComputeArtifactHash(t *testing.T) {
	hash1 := ComputeArtifactHash([]byte("hello"))
	hash2 := ComputeArtifactHash([]byte("hello"))
	hash3 := ComputeArtifactHash([]byte("world"))

	if hash1 != hash2 {
		t.Fatal("expected same hash for same input")
	}
	if hash1 == hash3 {
		t.Fatal("expected different hash for different input")
	}
	if len(hash1) != 64 {
		t.Fatalf("expected 64-char hex hash, got %d", len(hash1))
	}
}

func TestSummary(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "alice", Decision: control.DecisionAllow, PolicyID: "p1", Environment: "prod"})
	store.Append(EvidenceRecord{Actor: "bob", Decision: control.DecisionDeny, PolicyID: "p1", Environment: "prod"})
	store.Append(EvidenceRecord{Actor: "alice", Decision: control.DecisionRequireApproval, PolicyID: "p2", Environment: "staging"})
	store.Append(EvidenceRecord{
		Actor:        "alice",
		Decision:     control.DecisionAllow,
		PolicyID:     "p1",
		ArtifactHash: "abc123",
		ArtifactName: "app.tar.gz",
		Approval:     &ApprovalRecord{Approver: "admin", Valid: true},
	})

	summary := store.Summary()
	if summary.TotalRecords != 4 {
		t.Fatalf("expected 4 total, got %d", summary.TotalRecords)
	}
	if summary.ByDecision[control.DecisionAllow] != 2 {
		t.Fatalf("expected 2 allows, got %d", summary.ByDecision[control.DecisionAllow])
	}
	if summary.ByDecision[control.DecisionDeny] != 1 {
		t.Fatalf("expected 1 deny, got %d", summary.ByDecision[control.DecisionDeny])
	}
	if summary.DeniedCount != 1 {
		t.Fatalf("expected 1 denied, got %d", summary.DeniedCount)
	}
	if summary.ApprovalRequiredCount != 1 {
		t.Fatalf("expected 1 approval-required, got %d", summary.ApprovalRequiredCount)
	}
	if summary.ApprovedCount != 1 {
		t.Fatalf("expected 1 approved, got %d", summary.ApprovedCount)
	}
	if summary.WithArtifacts != 1 {
		t.Fatalf("expected 1 with artifacts, got %d", summary.WithArtifacts)
	}
	if !summary.ChainIntact {
		t.Fatal("expected chain to be intact")
	}
	if summary.ByActor["alice"] != 3 {
		t.Fatalf("expected 3 for alice, got %d", summary.ByActor["alice"])
	}
	if summary.ByEnvironment["prod"] != 2 {
		t.Fatalf("expected 2 for prod, got %d", summary.ByEnvironment["prod"])
	}
}

func TestLatest(t *testing.T) {
	store := NewStore()
	if store.Latest() != nil {
		t.Fatal("expected nil for empty store")
	}

	store.Append(EvidenceRecord{Actor: "a1", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a2", Decision: control.DecisionDeny, PolicyID: "p2"})

	latest := store.Latest()
	if latest == nil {
		t.Fatal("expected non-nil latest")
	}
	if latest.Actor != "a2" {
		t.Fatalf("expected latest actor a2, got %s", latest.Actor)
	}
}

func TestByID(t *testing.T) {
	store := NewStore()
	rec, _ := store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionAllow, PolicyID: "p1"})

	found := store.ByID(rec.ID)
	if found == nil {
		t.Fatal("expected to find record by ID")
	}
	if found.Actor != "a" {
		t.Fatalf("expected actor a, got %s", found.Actor)
	}

	if store.ByID("nonexistent") != nil {
		t.Fatal("expected nil for nonexistent ID")
	}
}

func TestByPolicy(t *testing.T) {
	store := NewStore()
	store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionDeny, PolicyID: "p2"})
	store.Append(EvidenceRecord{Actor: "a", Decision: control.DecisionAllow, PolicyID: "p1"})

	results := store.ByPolicy("p1")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for p1, got %d", len(results))
	}
}

func TestRecordsSortedNewestFirst(t *testing.T) {
	store := NewStore()
	now := time.Now().UTC()
	store.Append(EvidenceRecord{Actor: "a", Timestamp: now.Add(-2 * time.Hour), Decision: control.DecisionAllow, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Timestamp: now.Add(-1 * time.Hour), Decision: control.DecisionDeny, PolicyID: "p1"})
	store.Append(EvidenceRecord{Actor: "a", Timestamp: now, Decision: control.DecisionAllow, PolicyID: "p1"})

	records := store.Records()
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	if records[0].Timestamp.Before(records[1].Timestamp) {
		t.Fatal("expected newest first")
	}
}

func TestEmptyStoreVerify(t *testing.T) {
	store := NewStore()
	idx, err := store.Verify()
	if err != nil {
		t.Fatalf("expected no error on empty store, got %v", err)
	}
	if idx != -1 {
		t.Fatalf("expected -1 for empty store, got %d", idx)
	}
}

func TestConcurrentAppend(t *testing.T) {
	store := NewStore()
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func(n int) {
			store.Append(EvidenceRecord{
				Actor:    "user",
				Resource: "deploy",
				Decision: control.DecisionAllow,
				PolicyID: "p1",
			})
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 20; i++ {
		<-done
	}
	if store.Len() != 20 {
		t.Fatalf("expected 20 records, got %d", store.Len())
	}
}
