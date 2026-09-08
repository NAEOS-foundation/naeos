package verification

import (
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

func mkRecord(t *testing.T, store *evidence.EvidenceStore, actor, resource string, decision control.Decision, artifactName, artifactHash string, approval *evidence.ApprovalRecord) evidence.EvidenceRecord {
	t.Helper()
	rec, err := store.Append(evidence.EvidenceRecord{
		Actor:        actor,
		Resource:     resource,
		Action:       "run",
		Environment:  "production",
		PolicyID:     "p1",
		Decision:     decision,
		ArtifactName: artifactName,
		ArtifactHash: artifactHash,
		Approval:     approval,
	})
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}
	return rec
}

func TestEvidenceChainVerifierPasses(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)

	verifier := NewEvidenceChainVerifier(store)
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusVerified {
		t.Fatalf("expected VERIFIED, got %s", res.Status)
	}
}

func TestEvidenceChainVerifierDetectsTampering(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)

	// Tamper with the hash.
	rec.Hash = "tampered"

	verifier := NewEvidenceChainVerifier(store)
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("expected FAILED for tampered record, got %s", res.Status)
	}
}

func TestArtifactHashVerifierMatch(t *testing.T) {
	store := evidence.NewStore()
	content := []byte("artifact-content")
	hash := evidence.ComputeArtifactHash(content)
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "app.tar.gz", hash, nil)

	verifier := NewArtifactHashVerifier(map[string]string{"app.tar.gz": hash})
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusVerified {
		t.Fatalf("expected VERIFIED, got %s", res.Status)
	}
}

func TestArtifactHashVerifierMismatch(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "app.tar.gz", "deadbeef", nil)

	verifier := NewArtifactHashVerifier(map[string]string{"app.tar.gz": "cafebabe"})
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("expected FAILED on mismatch, got %s", res.Status)
	}
}

func TestArtifactHashVerifierNoHash(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)

	verifier := NewArtifactHashVerifier(map[string]string{})
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusUnverified {
		t.Fatalf("expected UNVERIFIED, got %s", res.Status)
	}
}

type staticArtifactSource map[string][]byte

func (s staticArtifactSource) Content(name string) ([]byte, bool) {
	c, ok := s[name]
	return c, ok
}

func TestArtifactHashVerifierWithSource(t *testing.T) {
	store := evidence.NewStore()
	content := []byte("live-content")
	hash := evidence.ComputeArtifactHash(content)
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "live.txt", hash, nil)

	src := staticArtifactSource{"live.txt": content}
	verifier := NewArtifactHashVerifierWithSource(nil, src)
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusVerified {
		t.Fatalf("expected VERIFIED, got %s", res.Status)
	}
}

func TestApprovalBindingVerifier(t *testing.T) {
	store := evidence.NewStore()
	artifactHash := "abc123"
	approval := &evidence.ApprovalRecord{
		Approver:     "admin",
		ArtifactHash: artifactHash,
		ArtifactName: "app.tar.gz",
		Timestamp:    time.Now(),
		Valid:        true,
	}
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "app.tar.gz", artifactHash, approval)

	verifier := NewApprovalBindingVerifier()
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusVerified {
		t.Fatalf("expected VERIFIED, got %s", res.Status)
	}
}

func TestApprovalBindingVerifierMismatch(t *testing.T) {
	store := evidence.NewStore()
	approval := &evidence.ApprovalRecord{
		Approver:     "admin",
		ArtifactHash: "wrong-hash",
		ArtifactName: "app.tar.gz",
		Timestamp:    time.Now(),
		Valid:        true,
	}
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "app.tar.gz", "actual-hash", approval)

	verifier := NewApprovalBindingVerifier()
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("expected FAILED on binding mismatch, got %s", res.Status)
	}
}

func TestVerifierChainAllPass(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)

	chain := NewChain(Contract{Name: "basic", Version: "1.0"},
		NewEvidenceChainVerifier(store),
	)

	res, err := chain.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusVerified {
		t.Fatalf("expected VERIFIED, got %s", res.Status)
	}
	if len(res.Checks) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(res.Checks))
	}
}

func TestVerifierChainOneFails(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)
	rec.Hash = "tampered"

	chain := NewChain(Contract{Name: "basic", Version: "1.0"},
		NewEvidenceChainVerifier(store),
		NewApprovalBindingVerifier(),
	)

	res, err := chain.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("expected FAILED, got %s", res.Status)
	}
}

func TestVerifierChainEmpty(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)

	chain := NewChain(Contract{Name: "none", Version: "1.0"})
	res, err := chain.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusUnverified {
		t.Fatalf("expected UNVERIFIED for empty chain, got %s", res.Status)
	}
}

func TestVerifierChainAddRemove(t *testing.T) {
	store := evidence.NewStore()
	verifier := NewEvidenceChainVerifier(store)
	chain := NewChain(Contract{Name: "basic", Version: "1.0"}, verifier)

	if len(chain.Verifiers()) != 1 {
		t.Fatalf("expected 1 verifier, got %d", len(chain.Verifiers()))
	}
	if !chain.DeleteVerifier("evidence-chain") {
		t.Fatal("expected to delete verifier")
	}
	if len(chain.Verifiers()) != 0 {
		t.Fatalf("expected 0 verifiers, got %d", len(chain.Verifiers()))
	}
	if chain.DeleteVerifier("evidence-chain") {
		t.Fatal("expected delete to fail for missing verifier")
	}
}

func TestEventBusFanout(t *testing.T) {
	bus := NewEventBus()
	s1 := NewMemorySink()
	s2 := NewMemorySink()
	bus.RegisterSink("one", s1)
	bus.RegisterSink("two", s2)

	err := bus.Publish(Event{
		Type:   EventVerificationComplete,
		Target: "rec-1",
		Status: StatusVerified,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s1.Count() != 1 {
		t.Fatalf("expected 1 event in sink one, got %d", s1.Count())
	}
	if s2.Count() != 1 {
		t.Fatalf("expected 1 event in sink two, got %d", s2.Count())
	}
	if bus.SinkCount() != 2 {
		t.Fatalf("expected 2 sinks, got %d", bus.SinkCount())
	}
	if len(bus.History()) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(bus.History()))
	}
}

func TestEventBusUnregister(t *testing.T) {
	bus := NewEventBus()
	s := NewMemorySink()
	bus.RegisterSink("one", s)
	bus.UnregisterSink("one")
	if bus.SinkCount() != 0 {
		t.Fatalf("expected 0 sinks, got %d", bus.SinkCount())
	}
}

type failingSink struct{}

func (failingSink) Emit(Event) error { return &sinkError{} }

type sinkError struct{}

func (s *sinkError) Error() string { return "sink failed" }

func TestEventBusFailureIsolation(t *testing.T) {
	bus := NewEventBus()
	good := NewMemorySink()
	bus.RegisterSink("bad", failingSink{})
	bus.RegisterSink("good", good)

	err := bus.Publish(Event{Type: EventVerificationComplete, Target: "x", Status: StatusVerified})
	if err == nil {
		t.Fatal("expected error from failing sink")
	}
	// Good sink still received the event despite the bad one failing.
	if good.Count() != 1 {
		t.Fatalf("expected good sink to receive event, got %d", good.Count())
	}
}

func TestVerifierChainConcurrent(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)

	chain := NewChain(Contract{Name: "basic", Version: "1.0"},
		NewEvidenceChainVerifier(store),
	)

	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			chain.Verify(rec)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
	if len(chain.History()) != 20 {
		t.Fatalf("expected 20 history entries, got %d", len(chain.History()))
	}
}
