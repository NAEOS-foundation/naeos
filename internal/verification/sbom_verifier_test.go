package verification

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

func TestSBOMVerifierMatch(t *testing.T) {
	store := evidence.NewStore()
	content := []byte(`{"bomFormat":"CycloneDX","specVersion":"1.5","version":1}`)
	hash := evidence.ComputeArtifactHash(content)
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "sbom.json", hash, nil)

	verifier := NewSBOMVerifier(map[string]string{"sbom.json": hash})
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusVerified {
		t.Fatalf("expected VERIFIED, got %s", res.Status)
	}
}

func TestSBOMVerifierMismatch(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "sbom.json", "deadbeef", nil)

	verifier := NewSBOMVerifier(map[string]string{"sbom.json": "cafebabe"})
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("expected FAILED on mismatch, got %s", res.Status)
	}
}

func TestSBOMVerifierNoHash(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "", "", nil)

	verifier := NewSBOMVerifier(map[string]string{})
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusUnverified {
		t.Fatalf("expected UNVERIFIED, got %s", res.Status)
	}
}

func TestSBOMVerifierWithSource(t *testing.T) {
	store := evidence.NewStore()
	content := []byte(`{"bomFormat":"CycloneDX"}`)
	hash := evidence.ComputeArtifactHash(content)
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "sbom.json", hash, nil)

	src := staticArtifactSource{"sbom.json": content}
	verifier := NewSBOMVerifierWithSource(nil, src)
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusVerified {
		t.Fatalf("expected VERIFIED, got %s", res.Status)
	}
}

func TestSBOMVerifierWithSourceNotFound(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "missing.json", "abc", nil)

	src := staticArtifactSource{}
	verifier := NewSBOMVerifierWithSource(nil, src)
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("expected FAILED when SBOM not found, got %s", res.Status)
	}
}

func TestSBOMVerifierWithSourceMismatch(t *testing.T) {
	store := evidence.NewStore()
	rec := mkRecord(t, store, "alice", "deploy", control.DecisionAllow, "sbom.json", "deadbeef", nil)

	src := staticArtifactSource{"sbom.json": []byte("content")}
	verifier := NewSBOMVerifierWithSource(nil, src)
	res, err := verifier.Verify(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("expected FAILED on source mismatch, got %s", res.Status)
	}
}

func TestSBOMVerifierName(t *testing.T) {
	verifier := NewSBOMVerifier(nil)
	if verifier.Name() != "sbom-integrity" {
		t.Errorf("expected name sbom-integrity, got %s", verifier.Name())
	}
}
