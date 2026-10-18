package interop

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

// This package is a deterministic contract harness. It is deliberately not wired
// into a production NAEOS enforcement path; passing tests validate this mock only.

type Candidate struct {
	Tool  string         `json:"tool"`
	Input map[string]any `json:"input"`
}

func (c Candidate) Digest() (string, error) {
	if c.Tool == "" {
		return "", errors.New("tool is required")
	}
	// encoding/json sorts map keys. The contract pins this representation for
	// the test harness; a cross-language adapter must implement the same spec or
	// negotiate a separately versioned canonicalization profile.
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

type Decision string

const (
	Proceed     Decision = "PROCEED"
	NeedsReview Decision = "NEEDS_REVIEW"
	Silence     Decision = "SILENCE"
)

type Authorization struct {
	ID              string
	Decision        Decision
	CandidateDigest string
	PolicyVersion   string
	Subject         string
	Scope           string
	ExpiresAt       time.Time
}

type Outcome string

const (
	Succeeded Outcome = "SUCCEEDED"
	Failed    Outcome = "FAILED"
	Unknown   Outcome = "UNKNOWN"
)

type Evidence struct {
	DecisionID             string
	InvocationID           string
	CandidateDigest        string
	PolicyVersion          string
	Outcome                Outcome
	RetryAllowed           bool
	ReconciliationRequired bool
	Reason                 string
}

type Gate struct {
	CurrentPolicy string
	Now           time.Time
	ProviderError bool
	Invocations   int
	SideEffects   map[string]string // action ID -> committed mock effect
}

func (g *Gate) Authorize(c Candidate, a Authorization, subject, scope string) (string, error) {
	if g.ProviderError {
		return "", errors.New("authorization provider unavailable")
	}
	digest, err := c.Digest()
	if err != nil {
		return "", err
	}
	if a.Decision != Proceed {
		return "", fmt.Errorf("decision is not PROCEED")
	}
	if a.CandidateDigest != digest {
		return "", errors.New("candidate digest mismatch")
	}
	if a.PolicyVersion == "" || a.PolicyVersion != g.CurrentPolicy {
		return "", errors.New("stale or missing policy version")
	}
	if !g.Now.Before(a.ExpiresAt) {
		return "", errors.New("authorization expired")
	}
	if a.Subject != subject || a.Scope != scope {
		return "", errors.New("subject or scope mismatch")
	}
	return digest, nil
}

func (g *Gate) Execute(c Candidate, a Authorization, subject, scope, invocationID string) Evidence {
	digest, err := g.Authorize(c, a, subject, scope)
	e := Evidence{
		DecisionID: a.ID, InvocationID: invocationID,
		CandidateDigest: a.CandidateDigest, PolicyVersion: a.PolicyVersion,
		Outcome: Failed, RetryAllowed: false,
	}
	if err != nil {
		e.Reason = err.Error()
		return e
	}
	g.Invocations++
	if g.SideEffects == nil {
		g.SideEffects = make(map[string]string)
	}
	g.SideEffects[invocationID] = digest
	e.CandidateDigest = digest
	e.Outcome = Succeeded
	e.RetryAllowed = true // only after a definitive result in this mock
	return e
}

func validSetup(t *testing.T) (*Gate, Candidate, Authorization) {
	t.Helper()
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	c := Candidate{Tool: "mock.transfer", Input: map[string]any{"amount": 10, "destination": "sandbox-account"}}
	digest, err := c.Digest()
	if err != nil {
		t.Fatal(err)
	}
	return &Gate{CurrentPolicy: "policy-v1", Now: now, SideEffects: make(map[string]string)}, c, Authorization{
		ID: "decision-001", Decision: Proceed, CandidateDigest: digest,
		PolicyVersion: "policy-v1", Subject: "agent-test", Scope: "mock.transfer",
		ExpiresAt: now.Add(5 * time.Minute),
	}
}

func TestT01FreshApprovalExecutesOnce(t *testing.T) {
	g, c, a := validSetup(t)
	e := g.Execute(c, a, "agent-test", "mock.transfer", "invoke-001")
	if e.Outcome != Succeeded || g.Invocations != 1 {
		t.Fatalf("expected one successful invocation; outcome=%s invocations=%d", e.Outcome, g.Invocations)
	}
	if e.DecisionID != a.ID || e.CandidateDigest != a.CandidateDigest || e.PolicyVersion != a.PolicyVersion {
		t.Fatalf("decision provenance not correlated: %+v", e)
	}
}

func TestT02CandidateMutationIsBlocked(t *testing.T) {
	g, c, a := validSetup(t)
	c.Input["amount"] = 999
	e := g.Execute(c, a, "agent-test", "mock.transfer", "invoke-002")
	if e.Outcome != Failed || e.Reason != "candidate digest mismatch" || g.Invocations != 0 {
		t.Fatalf("mutated candidate should be blocked: evidence=%+v invocations=%d", e, g.Invocations)
	}
}

func TestT03PolicyChangeInvalidatesApproval(t *testing.T) {
	g, c, a := validSetup(t)
	g.CurrentPolicy = "policy-v2"
	e := g.Execute(c, a, "agent-test", "mock.transfer", "invoke-003")
	if e.Outcome != Failed || e.Reason != "stale or missing policy version" || g.Invocations != 0 {
		t.Fatalf("stale approval should be blocked: evidence=%+v", e)
	}
}

func TestT04AuthorizationProviderFailureFailsClosed(t *testing.T) {
	g, c, a := validSetup(t)
	g.ProviderError = true
	e := g.Execute(c, a, "agent-test", "mock.transfer", "invoke-004")
	if e.Outcome != Failed || g.Invocations != 0 || e.Reason != "authorization provider unavailable" {
		t.Fatalf("provider failure must fail closed: evidence=%+v invocations=%d", e, g.Invocations)
	}
}

func TestT05UnknownSideEffectSuppressesAutomaticRetry(t *testing.T) {
	g, c, a := validSetup(t)
	digest, err := g.Authorize(c, a, "agent-test", "mock.transfer")
	if err != nil {
		t.Fatal(err)
	}
	actionID := "action-unknown-005"
	// Fault injection: side effect commits, but the response is lost.
	g.Invocations++
	g.SideEffects[actionID] = digest
	e := Evidence{
		DecisionID: a.ID, InvocationID: actionID, CandidateDigest: digest,
		PolicyVersion: a.PolicyVersion, Outcome: Unknown,
		RetryAllowed: false, ReconciliationRequired: true,
		Reason: "response lost after mock side effect commit",
	}
	if e.Outcome != Unknown || e.RetryAllowed || !e.ReconciliationRequired {
		t.Fatalf("UNKNOWN must require reconciliation and suppress retry: %+v", e)
	}
	// No retry is performed: UNKNOWN with reconciliation required must remain pending.
	if g.Invocations != 1 {
		t.Fatalf("unexpected retry: invocation count=%d", g.Invocations)
	}
}

func TestT06ReconciliationResolvesUnknownByActionID(t *testing.T) {
	g, c, a := validSetup(t)
	digest, err := g.Authorize(c, a, "agent-test", "mock.transfer")
	if err != nil {
		t.Fatal(err)
	}
	actionID := "action-reconcile-006"
	g.Invocations++
	g.SideEffects[actionID] = digest
	// The mock side-effect ledger is the authoritative reconciliation source.
	observed, ok := g.SideEffects[actionID]
	if !ok || observed != digest {
		t.Fatal("expected committed side effect to be discoverable by action ID")
	}
	resolved := Evidence{
		DecisionID: a.ID, InvocationID: actionID, CandidateDigest: observed,
		PolicyVersion: a.PolicyVersion, Outcome: Succeeded,
		RetryAllowed: false, ReconciliationRequired: false,
		Reason: "reconciled from mock side-effect ledger",
	}
	if resolved.Outcome != Succeeded || resolved.ReconciliationRequired {
		t.Fatalf("reconciliation did not resolve UNKNOWN: %+v", resolved)
	}
	if g.Invocations != 1 {
		t.Fatalf("reconciliation must not re-execute action; invocations=%d", g.Invocations)
	}
}

func TestT07SubjectOrScopeMismatchIsBlocked(t *testing.T) {
	g, c, a := validSetup(t)
	e := g.Execute(c, a, "different-agent", "mock.transfer", "invoke-007")
	if e.Outcome != Failed || e.Reason != "subject or scope mismatch" || g.Invocations != 0 {
		t.Fatalf("subject mismatch should be blocked: evidence=%+v", e)
	}
}

func TestCanonicalDigestIsStableAcrossMapInsertionOrder(t *testing.T) {
	a := Candidate{Tool: "mock.read", Input: map[string]any{"z": "last", "a": "first"}}
	b := Candidate{Tool: "mock.read", Input: map[string]any{"a": "first", "z": "last"}}
	da, err := a.Digest()
	if err != nil {
		t.Fatal(err)
	}
	db, err := b.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if da != db {
		t.Fatalf("expected deterministic digest; got %s and %s", da, db)
	}
}
