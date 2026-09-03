package evidence

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

// EvidenceRecord is the atomic unit of governance evidence. It links a
// policy decision to an execution outcome and captures the artifact
// integrity hash, forming a complete, tamper-evident audit trail for
// consequential AI engineering actions.
type EvidenceRecord struct {
	ID            string               `json:"id"`
	Timestamp     time.Time            `json:"timestamp"`
	Actor         string               `json:"actor"`
	Resource      string               `json:"resource"`
	Action        string               `json:"action"`
	Environment   string               `json:"environment,omitempty"`
	PolicyID      string               `json:"policy_id"`
	PolicyVersion string               `json:"policy_version"`
	RuleID        string               `json:"rule_id,omitempty"`
	Decision      control.Decision     `json:"decision"`
	DecisionReasons []string           `json:"decision_reasons,omitempty"`
	ArtifactName  string               `json:"artifact_name,omitempty"`
	ArtifactHash  string               `json:"artifact_hash,omitempty"`
	ArtifactSize  int                  `json:"artifact_size,omitempty"`
	ExecutionStatus string             `json:"execution_status,omitempty"`
	ExecutionOutput string             `json:"execution_output,omitempty"`
	ExecutionDurationMs int64          `json:"execution_duration_ms,omitempty"`
	Approval      *ApprovalRecord      `json:"approval,omitempty"`
	Metadata      map[string]any       `json:"metadata,omitempty"`
	PreviousHash  string               `json:"previous_hash"`
	Hash          string               `json:"hash"`
}

// ApprovalRecord captures an explicit human or system approval bound to
// a specific artifact version.
type ApprovalRecord struct {
	Approver      string    `json:"approver"`
	ArtifactHash  string    `json:"artifact_hash"`
	ArtifactName  string    `json:"artifact_name"`
	Timestamp     time.Time `json:"timestamp"`
	Reason        string    `json:"reason,omitempty"`
	Valid         bool      `json:"valid"`
}

// EvidenceStore is an append-only, tamper-evident store of evidence
// records. Each record's hash chains to the previous record, forming
// an integrity chain similar to a blockchain.
type EvidenceStore struct {
	mu      sync.RWMutex
	records []EvidenceRecord
	latest  string // hash of the most recent record
}

// NewStore creates an empty EvidenceStore.
func NewStore() *EvidenceStore {
	return &EvidenceStore{}
}

// Append adds an evidence record to the store. The record's hash is
// computed from its content plus the previous record's hash, forming
// a tamper-evident chain. If the record has no ID, one is generated.
func (s *EvidenceStore) Append(rec EvidenceRecord) (EvidenceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rec.Timestamp.IsZero() {
		rec.Timestamp = time.Now().UTC()
	}
	if rec.ID == "" {
		rec.ID = fmt.Sprintf("ev-%d", rec.Timestamp.UnixNano())
	}

	rec.PreviousHash = s.latest
	rec.Hash = s.computeHash(rec)
	s.records = append(s.records, rec)
	s.latest = rec.Hash
	return rec, nil
}

// Verify walks the entire chain and verifies that each record's hash
// matches its computed value and that the PreviousHash links correctly.
// Returns the index of the first broken record, or -1 if the chain is
// intact.
func (s *EvidenceStore) Verify() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var prevHash string
	for i, rec := range s.records {
		if rec.PreviousHash != prevHash {
			return i, naeoserr.New(naeoserr.ErrConflict,
				fmt.Sprintf("chain break at index %d: expected previous_hash %q, got %q", i, prevHash, rec.PreviousHash))
		}
		computed := s.computeHashUnlocked(rec)
		if rec.Hash != computed {
			return i, naeoserr.New(naeoserr.ErrConflict,
				fmt.Sprintf("hash mismatch at index %d: expected %s, got %s", i, computed, rec.Hash))
		}
		prevHash = rec.Hash
	}
	return -1, nil
}

// Len returns the number of evidence records.
func (s *EvidenceStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}

// Latest returns the most recent evidence record, or nil if empty.
func (s *EvidenceStore) Latest() *EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.records) == 0 {
		return nil
	}
	rec := s.records[len(s.records)-1]
	return &rec
}

// ByID returns the evidence record with the given ID.
func (s *EvidenceStore) ByID(id string) *EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.records {
		if s.records[i].ID == id {
			return &s.records[i]
		}
	}
	return nil
}

// ByActor returns all records for the given actor.
func (s *EvidenceStore) ByActor(actor string) []EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []EvidenceRecord
	for _, r := range s.records {
		if r.Actor == actor {
			out = append(out, r)
		}
	}
	return out
}

// ByResource returns all records for the given resource.
func (s *EvidenceStore) ByResource(resource string) []EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []EvidenceRecord
	for _, r := range s.records {
		if r.Resource == resource {
			out = append(out, r)
		}
	}
	return out
}

// ByPolicy returns all records for the given policy ID.
func (s *EvidenceStore) ByPolicy(policyID string) []EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []EvidenceRecord
	for _, r := range s.records {
		if r.PolicyID == policyID {
			out = append(out, r)
		}
	}
	return out
}

// ByDecision returns all records with the given decision.
func (s *EvidenceStore) ByDecision(dec control.Decision) []EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []EvidenceRecord
	for _, r := range s.records {
		if r.Decision == dec {
			out = append(out, r)
		}
	}
	return out
}

// Denied returns all records where the decision was DENY.
func (s *EvidenceStore) Denied() []EvidenceRecord {
	return s.ByDecision(control.DecisionDeny)
}

// Approvals returns all records that carry an approval.
func (s *EvidenceStore) Approvals() []EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []EvidenceRecord
	for _, r := range s.records {
		if r.Approval != nil && r.Approval.Valid {
			out = append(out, r)
		}
	}
	return out
}

// Query filters records by the given criteria. Empty fields are
// wildcards. Results are returned newest-first.
type EvidenceQuery struct {
	Actor      string
	Resource   string
	PolicyID   string
	Decision   control.Decision
	From       time.Time
	To         time.Time
	Limit      int
}

func (s *EvidenceStore) Query(q EvidenceQuery) []EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []EvidenceRecord
	for _, r := range s.records {
		if q.Actor != "" && r.Actor != q.Actor {
			continue
		}
		if q.Resource != "" && r.Resource != q.Resource {
			continue
		}
		if q.PolicyID != "" && r.PolicyID != q.PolicyID {
			continue
		}
		if q.Decision != "" && r.Decision != q.Decision {
			continue
		}
		if !q.From.IsZero() && r.Timestamp.Before(q.From) {
			continue
		}
		if !q.To.IsZero() && r.Timestamp.After(q.To) {
			continue
		}
		out = append(out, r)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp.After(out[j].Timestamp)
	})

	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out
}

// Records returns a copy of all evidence records (newest-first).
func (s *EvidenceStore) Records() []EvidenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EvidenceRecord, len(s.records))
	copy(out, s.records)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp.After(out[j].Timestamp)
	})
	return out
}

// Summary returns aggregate statistics about the evidence store.
type EvidenceSummary struct {
	TotalRecords    int                        `json:"total_records"`
	ByDecision      map[control.Decision]int   `json:"by_decision"`
	ByActor         map[string]int             `json:"by_actor"`
	ByPolicy        map[string]int             `json:"by_policy"`
	ByEnvironment   map[string]int             `json:"by_environment"`
	ApprovedCount   int                        `json:"approved_count"`
	DeniedCount     int                        `json:"denied_count"`
	ApprovalRequiredCount int                  `json:"approval_required_count"`
	WithArtifacts   int                        `json:"with_artifacts"`
	ChainIntact     bool                       `json:"chain_intact"`
}

func (s *EvidenceStore) Summary() EvidenceSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := EvidenceSummary{
		TotalRecords: len(s.records),
		ByDecision:   make(map[control.Decision]int),
		ByActor:      make(map[string]int),
		ByPolicy:     make(map[string]int),
		ByEnvironment: make(map[string]int),
	}

	for _, r := range s.records {
		summary.ByDecision[r.Decision]++
		summary.ByActor[r.Actor]++
		summary.ByPolicy[r.PolicyID]++
		summary.ByEnvironment[r.Environment]++
		if r.ArtifactHash != "" {
			summary.WithArtifacts++
		}
		if r.Approval != nil && r.Approval.Valid {
			summary.ApprovedCount++
		}
	}

	summary.DeniedCount = summary.ByDecision[control.DecisionDeny]
	summary.ApprovalRequiredCount = summary.ByDecision[control.DecisionRequireApproval]

	// Quick chain integrity check (non-locking, read-side only).
	var prevHash string
	chainOK := true
	for _, rec := range s.records {
		if rec.PreviousHash != prevHash {
			chainOK = false
			break
		}
		computed := computeHashStatic(rec)
		if rec.Hash != computed {
			chainOK = false
			break
		}
		prevHash = rec.Hash
	}
	summary.ChainIntact = chainOK

	return summary
}

// ComputeArtifactHash computes the SHA-256 hex digest of the given bytes.
func ComputeArtifactHash(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// RecomputeHash recomputes the canonical hash of the given evidence record,
// independent of the stored Hash field. It returns the expected SHA-256 hex
// digest that the record's Hash field should equal.
func RecomputeHash(rec EvidenceRecord) string {
	return computeHashStatic(rec)
}

func (s *EvidenceStore) computeHash(rec EvidenceRecord) string {
	return computeHashStatic(rec)
}

func (s *EvidenceStore) computeHashUnlocked(rec EvidenceRecord) string {
	return computeHashStatic(rec)
}

func computeHashStatic(rec EvidenceRecord) string {
	// Hash the canonical content fields, excluding the Hash field itself.
	type hashable struct {
		ID              string          `json:"id"`
		Timestamp       time.Time       `json:"timestamp"`
		Actor           string          `json:"actor"`
		Resource        string          `json:"resource"`
		Action          string          `json:"action"`
		Environment     string          `json:"environment"`
		PolicyID        string          `json:"policy_id"`
		PolicyVersion   string          `json:"policy_version"`
		RuleID          string          `json:"rule_id"`
		Decision        string          `json:"decision"`
		DecisionReasons []string        `json:"decision_reasons"`
		ArtifactName    string          `json:"artifact_name"`
		ArtifactHash    string          `json:"artifact_hash"`
		ArtifactSize    int             `json:"artifact_size"`
		ExecutionStatus string          `json:"execution_status"`
		ExecutionOutput string          `json:"execution_output"`
		ExecutionDurationMs int64       `json:"execution_duration_ms"`
		Approval        *ApprovalRecord `json:"approval"`
		Metadata        map[string]any  `json:"metadata"`
		PreviousHash    string          `json:"previous_hash"`
	}

	h := hashable{
		ID:              rec.ID,
		Timestamp:       rec.Timestamp,
		Actor:           rec.Actor,
		Resource:        rec.Resource,
		Action:          rec.Action,
		Environment:     rec.Environment,
		PolicyID:        rec.PolicyID,
		PolicyVersion:   rec.PolicyVersion,
		RuleID:          rec.RuleID,
		Decision:        string(rec.Decision),
		DecisionReasons: rec.DecisionReasons,
		ArtifactName:    rec.ArtifactName,
		ArtifactHash:    rec.ArtifactHash,
		ArtifactSize:    rec.ArtifactSize,
		ExecutionStatus: rec.ExecutionStatus,
		ExecutionOutput: rec.ExecutionOutput,
		ExecutionDurationMs: rec.ExecutionDurationMs,
		Approval:        rec.Approval,
		Metadata:        rec.Metadata,
		PreviousHash:    rec.PreviousHash,
	}

	data, _ := json.Marshal(h)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
