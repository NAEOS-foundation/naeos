package controlplane

import (
	"fmt"
	"sync"
	"time"
)

// Approval records an explicit human or service approval for a pending decision.
type Approval struct {
	ID           string    `json:"id"`
	DecisionID   string    `json:"decision_id"`
	Approver     string    `json:"approver"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	ApprovedAt   time.Time `json:"approved_at,omitempty"`
	ArtifactHash string    `json:"artifact_hash,omitempty"`
	Reason       string    `json:"reason,omitempty"`
}

// IsValid reports whether an approval can authorize execution now.
func (a Approval) IsValid(now time.Time) bool {
	return a.Status == "approved" && a.DecisionID != "" && (a.ExpiresAt.IsZero() || now.Before(a.ExpiresAt))
}

// ApprovalStore manages the approval lifecycle for pending decisions.
type ApprovalStore struct {
	mu        sync.RWMutex
	approvals map[string]Approval
}

func NewApprovalStore() *ApprovalStore {
	return &ApprovalStore{approvals: make(map[string]Approval)}
}

func (s *ApprovalStore) Create(decisionID, approver string, expiresAt time.Time) (Approval, error) {
	if s == nil {
		return Approval{}, fmt.Errorf("approval store unavailable")
	}
	if decisionID == "" || approver == "" {
		return Approval{}, fmt.Errorf("decision ID and approver are required")
	}
	approval := Approval{
		ID:         controlPlaneID("APR"),
		DecisionID: decisionID,
		Approver:   approver,
		Status:     "pending",
		CreatedAt:  time.Now().UTC(),
		ExpiresAt:  expiresAt,
	}
	s.mu.Lock()
	s.approvals[approval.ID] = approval
	s.mu.Unlock()
	return approval, nil
}

func (s *ApprovalStore) Approve(id, reason, artifactHash string, now time.Time) (Approval, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	approval, ok := s.approvals[id]
	if !ok {
		return Approval{}, fmt.Errorf("approval %s not found", id)
	}
	if approval.Status != "pending" {
		return Approval{}, fmt.Errorf("approval %s is already %s", id, approval.Status)
	}
	if !approval.ExpiresAt.IsZero() && !now.Before(approval.ExpiresAt) {
		approval.Status = "expired"
		s.approvals[id] = approval
		return Approval{}, fmt.Errorf("approval %s is expired", id)
	}
	approval.Status = "approved"
	approval.ApprovedAt = now.UTC()
	approval.Reason = reason
	approval.ArtifactHash = artifactHash
	s.approvals[id] = approval
	return approval, nil
}

func (s *ApprovalStore) Get(id string) (Approval, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	approval, ok := s.approvals[id]
	if !ok {
		return Approval{}, fmt.Errorf("approval %s not found", id)
	}
	return approval, nil
}

// Consume marks an approved approval as used by a single execution.
func (s *ApprovalStore) Consume(id string, now time.Time) (Approval, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	approval, ok := s.approvals[id]
	if !ok {
		return Approval{}, fmt.Errorf("approval %s not found", id)
	}
	if !approval.IsValid(now) {
		return Approval{}, fmt.Errorf("approval %s is not valid", id)
	}
	approval.Status = "consumed"
	s.approvals[id] = approval
	return approval, nil
}
