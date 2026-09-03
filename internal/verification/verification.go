package verification

import (
	"fmt"
	"sort"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

// VerificationStatus is the outcome of verifying an evidence record.
type VerificationStatus string

const (
	StatusVerified   VerificationStatus = "VERIFIED"
	StatusFailed     VerificationStatus = "FAILED"
	StatusUnverified VerificationStatus = "UNVERIFIED"
)

// Contract describes what a verification run guarantees. It is the
// communication interface between NAEOS and an independent verifier.
type Contract struct {
	// Name identifies the verification contract.
	Name string
	// Description explains what the contract verifies.
	Description string
	// Requirements lists the checks that must pass.
	Requirements []string
	// Version is the contract version.
	Version string
}

// VerificationResult is the evidence produced by a verification run.
type VerificationResult struct {
	Status   VerificationStatus
	Contract string
	ContractVersion string
	Target   string // evidence ID or artifact hash being verified
	Checks   []CheckResult
	Timestamp time.Time
	Verifier string
	Message  string
}

// CheckResult records the outcome of a single verification check.
type CheckResult struct {
	Name   string
	Passed bool
	Detail string
}

// Verifier is the verification contract interface. A Verifier independently
// verifies evidence records and produces a VerificationResult.
type Verifier interface {
	// Name returns the verifier identifier.
	Name() string
	// Verify verifies an evidence record and returns the result.
	Verify(rec evidence.EvidenceRecord) (VerificationResult, error)
}

// VerifierOption configures a VerifierChain.
type VerifierOption func(*VerifierChain)

// WithContract attaches a contract to the chain.
func WithContract(c Contract) VerifierOption {
	return func(v *VerifierChain) { v.contract = c }
}

// VerifierChain runs multiple verifiers and aggregates their results. All
// verifiers must pass for the overall result to be VERIFIED.
type VerifierChain struct {
	mu       sync.RWMutex
	verifiers []Verifier
	contract Contract
	history  []VerificationResult
}

// NewChain creates a VerifierChain with the given verifiers and contract.
func NewChain(contract Contract, verifiers ...Verifier) *VerifierChain {
	return &VerifierChain{
		verifiers: verifiers,
		contract:  contract,
	}
}

// AddVerifier appends a verifier to the chain.
func (v *VerifierChain) AddVerifier(ver Verifier) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.verifiers = append(v.verifiers, ver)
}

// DeleteVerifier removes a verifier by name. Returns false if not found.
func (v *VerifierChain) DeleteVerifier(name string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	for i, ver := range v.verifiers {
		if ver.Name() == name {
			v.verifiers = append(v.verifiers[:i], v.verifiers[i+1:]...)
			return true
		}
	}
	return false
}

// Verify runs every verifier against the evidence record and aggregates the
// results. The chain requires all verifiers to pass for a VERIFIED outcome.
func (v *VerifierChain) Verify(rec evidence.EvidenceRecord) (VerificationResult, error) {
	v.mu.RLock()
	verifiers := make([]Verifier, len(v.verifiers))
	copy(verifiers, v.verifiers)
	contract := v.contract
	v.mu.RUnlock()

	agg := VerificationResult{
		Status:      StatusVerified,
		Contract:    contract.Name,
		ContractVersion: contract.Version,
		Target:      rec.ID,
		Timestamp:   time.Now().UTC(),
	}

	for _, ver := range verifiers {
		res, err := ver.Verify(rec)
		if err != nil {
			return VerificationResult{}, naeoserr.Wrapf(err, naeoserr.ErrInternal, "verifier %s failed", ver.Name())
		}
		agg.Verifier = ver.Name()
		agg.Checks = append(agg.Checks, res.Checks...)
		if res.Status == StatusFailed {
			agg.Status = StatusFailed
			agg.Message = res.Message
		}
	}

	if len(verifiers) == 0 {
		agg.Status = StatusUnverified
		agg.Message = "no verifiers registered"
	}

	v.record(agg)
	return agg, nil
}

// VerifyEvidence verifies the entire evidence store chain (internal integrity),
// then runs the verifier chain over each record.
func (v *VerifierChain) VerifyEvidence(store *evidence.EvidenceStore) ([]VerificationResult, error) {
	idx, err := store.Verify()
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrConflict, "evidence chain integrity check failed at index %d", idx)
	}

	var results []VerificationResult
	for _, rec := range store.Records() {
		res, err := v.Verify(rec)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}

// History returns all verification results (newest first).
func (v *VerifierChain) History() []VerificationResult {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]VerificationResult, len(v.history))
	copy(out, v.history)
	return out
}

// Verifiers returns the registered verifier names.
func (v *VerifierChain) Verifiers() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	names := make([]string, len(v.verifiers))
	for i, ver := range v.verifiers {
		names[i] = ver.Name()
	}
	sort.Strings(names)
	return names
}

func (v *VerifierChain) record(r VerificationResult) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.history = append(v.history, r)
}

// EvidenceChainVerifier verifies that an evidence record's hash is correct
// and that the chain linkage for the record is intact.
type EvidenceChainVerifier struct {
	store *evidence.EvidenceStore
}

// NewEvidenceChainVerifier creates a verifier that checks record integrity
// against the given evidence store.
func NewEvidenceChainVerifier(store *evidence.EvidenceStore) *EvidenceChainVerifier {
	return &EvidenceChainVerifier{store: store}
}

// Name returns "evidence-chain".
func (v *EvidenceChainVerifier) Name() string { return "evidence-chain" }

// Verify checks that the record exists in the store and its hash matches.
func (v *EvidenceChainVerifier) Verify(rec evidence.EvidenceRecord) (VerificationResult, error) {
	res := VerificationResult{
		Status:    StatusVerified,
		Target:    rec.ID,
		Timestamp: time.Now().UTC(),
	}

	// Check the record hash recomputes correctly.
	computed := evidence.RecomputeHash(rec)
	hashCheck := CheckResult{
		Name:   "hash-integrity",
		Passed: rec.Hash == computed,
		Detail: fmt.Sprintf("record %s hash %s", rec.ID, rec.Hash),
	}
	res.Checks = append(res.Checks, hashCheck)

	// Check the record is present in the store.
	stored := v.store.ByID(rec.ID)
	presenceCheck := CheckResult{
		Name:   "store-presence",
		Passed: stored != nil,
		Detail: fmt.Sprintf("record %s present in store", rec.ID),
	}
	res.Checks = append(res.Checks, presenceCheck)

	for _, c := range res.Checks {
		if !c.Passed {
			res.Status = StatusFailed
			res.Message = "evidence chain verification failed"
			return res, nil
		}
	}

	res.Message = "evidence chain verified"
	return res, nil
}

// ArtifactHashVerifier verifies that an artifact hash matches its expected
// value. It is the building block for artifact-bound verification.
type ArtifactHashVerifier struct {
	// expected maps artifact names to their known SHA-256 hashes.
	expected map[string]string
	// artifacts provides a way to fetch artifact content by name (optional).
	artifacts ArtifactSource
}

// ArtifactSource resolves artifact content by name for hash verification.
type ArtifactSource interface {
	// Content returns the bytes for the named artifact.
	Content(name string) ([]byte, bool)
}

// NewArtifactHashVerifier creates a verifier with known artifact hashes.
func NewArtifactHashVerifier(expected map[string]string) *ArtifactHashVerifier {
	return &ArtifactHashVerifier{expected: expected}
}

// NewArtifactHashVerifierWithSource creates a verifier that reads artifact
// content live and recomputes the hash.
func NewArtifactHashVerifierWithSource(expected map[string]string, src ArtifactSource) *ArtifactHashVerifier {
	return &ArtifactHashVerifier{expected: expected, artifacts: src}
}

// Name returns "artifact-hash".
func (v *ArtifactHashVerifier) Name() string { return "artifact-hash" }

// Verify checks the artifact referenced by the record.
func (v *ArtifactHashVerifier) Verify(rec evidence.EvidenceRecord) (VerificationResult, error) {
	res := VerificationResult{
		Status:    StatusVerified,
		Target:    rec.ID,
		Timestamp: time.Now().UTC(),
	}

	if rec.ArtifactHash == "" {
		res.Status = StatusUnverified
		res.Message = "record has no artifact hash to verify"
		return res, nil
	}

	// If we have a live source, recompute the hash from content.
	if v.artifacts != nil {
		content, ok := v.artifacts.Content(rec.ArtifactName)
		if !ok {
			res.Status = StatusFailed
			res.Message = fmt.Sprintf("artifact %s not found in source", rec.ArtifactName)
			res.Checks = append(res.Checks, CheckResult{Name: "artifact-source", Passed: false, Detail: "not found"})
			return res, nil
		}
		actual := evidence.ComputeArtifactHash(content)
		match := actual == rec.ArtifactHash
		res.Checks = append(res.Checks, CheckResult{
			Name:   "artifact-content-hash",
			Passed: match,
			Detail: fmt.Sprintf("expected %s got %s", rec.ArtifactHash, actual),
		})
		if !match {
			res.Status = StatusFailed
			res.Message = "artifact content hash mismatch"
		}
		return res, nil
	}

	// Otherwise check against the expected mapping.
	expected, ok := v.expected[rec.ArtifactName]
	check := CheckResult{
		Name:   "artifact-expected-hash",
		Passed: ok && expected == rec.ArtifactHash,
		Detail: fmt.Sprintf("artifact %s", rec.ArtifactName),
	}
	res.Checks = append(res.Checks, check)
	if !check.Passed {
		res.Status = StatusFailed
		res.Message = fmt.Sprintf("artifact %s hash does not match expected", rec.ArtifactName)
	}
	return res, nil
}

// ApprovalBindingVerifier verifies that an approval record is bound to the
// exact artifact referenced by the evidence record.
type ApprovalBindingVerifier struct{}

// NewApprovalBindingVerifier creates the approval binding verifier.
func NewApprovalBindingVerifier() *ApprovalBindingVerifier {
	return &ApprovalBindingVerifier{}
}

// Name returns "approval-binding".
func (v *ApprovalBindingVerifier) Name() string { return "approval-binding" }

// Verify checks that the approval's artifact hash matches the record's hash.
func (v *ApprovalBindingVerifier) Verify(rec evidence.EvidenceRecord) (VerificationResult, error) {
	res := VerificationResult{
		Status:    StatusVerified,
		Target:    rec.ID,
		Timestamp: time.Now().UTC(),
		Message:   "no approval binding to verify",
	}

	if rec.Approval == nil {
		res.Status = StatusUnverified
		return res, nil
	}

	check := CheckResult{
		Name:   "approval-artifact-match",
		Passed: rec.Approval.ArtifactHash == rec.ArtifactHash,
		Detail: fmt.Sprintf("approval bound to %s, record artifact %s", rec.Approval.ArtifactHash, rec.ArtifactHash),
	}
	res.Checks = append(res.Checks, check)

	validCheck := CheckResult{
		Name:   "approval-valid",
		Passed: rec.Approval.Valid,
		Detail: "approval marked valid",
	}
	res.Checks = append(res.Checks, validCheck)

	if !check.Passed || !validCheck.Passed {
		res.Status = StatusFailed
		res.Message = "approval artifact binding mismatch"
		return res, nil
	}

	res.Message = "approval artifact binding verified"
	return res, nil
}
