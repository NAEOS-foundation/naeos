// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"time"
)

// MemoryApprovalRepository adapts ApprovalStore for tests and demo composition.
// It is not a durable production implementation.
type MemoryApprovalRepository struct {
	store *ApprovalStore
}

// NewMemoryApprovalRepository returns a repository backed by the supplied store.
func NewMemoryApprovalRepository(store *ApprovalStore) *MemoryApprovalRepository {
	return &MemoryApprovalRepository{store: store}
}

func (r *MemoryApprovalRepository) Create(ctx context.Context, approval Approval) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.store == nil || r.store.approvals == nil {
		return ErrRepositoryUnavailable
	}
	if approval.ID == "" || approval.DecisionID == "" || approval.Approver == "" || approval.Status != "pending" {
		return ErrRepositoryConflict
	}
	if !approval.ExpiresAt.IsZero() && !approval.CreatedAt.IsZero() && !approval.CreatedAt.Before(approval.ExpiresAt) {
		return ErrRepositoryConflict
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.approvals[approval.ID]; exists {
		return ErrRepositoryConflict
	}
	r.store.approvals[approval.ID] = approval
	return nil
}

func (r *MemoryApprovalRepository) Get(ctx context.Context, id string) (Approval, error) {
	if err := ctx.Err(); err != nil {
		return Approval{}, err
	}
	if r == nil || r.store == nil || r.store.approvals == nil {
		return Approval{}, ErrRepositoryUnavailable
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	approval, exists := r.store.approvals[id]
	if !exists {
		return Approval{}, ErrRepositoryNotFound
	}
	return approval, nil
}

// Transition atomically compares the current status and updates the approval.
// Allowed transitions are pending->approved/rejected/expired and approved->consumed.
func (r *MemoryApprovalRepository) Transition(ctx context.Context, id, expectedStatus, nextStatus string, at time.Time) (Approval, error) {
	if err := ctx.Err(); err != nil {
		return Approval{}, err
	}
	if r == nil || r.store == nil || r.store.approvals == nil {
		return Approval{}, ErrRepositoryUnavailable
	}
	if !validApprovalTransition(expectedStatus, nextStatus) {
		return Approval{}, ErrRepositoryConflict
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	approval, exists := r.store.approvals[id]
	if !exists {
		return Approval{}, ErrRepositoryNotFound
	}
	if approval.Status != expectedStatus {
		return Approval{}, ErrRepositoryConflict
	}
	at = at.UTC()
	if !approval.ExpiresAt.IsZero() && !at.Before(approval.ExpiresAt) {
		if expectedStatus == "pending" && nextStatus != "expired" {
			approval.Status = "expired"
			r.store.approvals[id] = approval
			return Approval{}, ErrRepositoryConflict
		}
		if expectedStatus == "approved" && nextStatus == "consumed" {
			return Approval{}, ErrRepositoryConflict
		}
	}
	if nextStatus == "expired" && (approval.ExpiresAt.IsZero() || at.Before(approval.ExpiresAt)) {
		return Approval{}, ErrRepositoryConflict
	}
	if nextStatus == "approved" {
		approval.ApprovedAt = at
	}
	approval.Status = nextStatus
	r.store.approvals[id] = approval
	return approval, nil
}

func validApprovalTransition(expectedStatus, nextStatus string) bool {
	switch expectedStatus {
	case "pending":
		return nextStatus == "approved" || nextStatus == "rejected" || nextStatus == "expired"
	case "approved":
		return nextStatus == "consumed"
	default:
		return false
	}
}

var _ ApprovalRepository = (*MemoryApprovalRepository)(nil)
