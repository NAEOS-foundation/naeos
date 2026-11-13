// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"sync"
	"time"
)

// MemoryGrantRepository stores grants in process memory for tests and demo use.
// It is not a durable production implementation.
type MemoryGrantRepository struct {
	mu      sync.RWMutex
	grants  map[string]Grant
	byAgent map[string]string
}

// NewMemoryGrantRepository creates an empty in-memory grant repository.
func NewMemoryGrantRepository() *MemoryGrantRepository {
	return &MemoryGrantRepository{
		grants:  make(map[string]Grant),
		byAgent: make(map[string]string),
	}
}

func (r *MemoryGrantRepository) Create(ctx context.Context, grant Grant) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.grants == nil || r.byAgent == nil {
		return ErrRepositoryUnavailable
	}
	if grant.GrantID == "" || grant.AgentID == "" || grant.PolicyID == "" ||
		grant.PolicyVersion < 1 || len(grant.Capabilities) == 0 ||
		grant.Status != "active" || grant.Revoked {
		return ErrRepositoryConflict
	}
	if !grant.CreatedAt.IsZero() && !grant.ExpiresAt.IsZero() &&
		!grant.CreatedAt.Before(grant.ExpiresAt) {
		return ErrRepositoryConflict
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.grants[grant.GrantID]; exists {
		return ErrRepositoryConflict
	}
	if _, exists := r.byAgent[grant.AgentID]; exists {
		return ErrRepositoryConflict
	}
	grant = cloneGrant(grant)
	r.grants[grant.GrantID] = grant
	r.byAgent[grant.AgentID] = grant.GrantID
	return nil
}

func (r *MemoryGrantRepository) Get(ctx context.Context, grantID string) (Grant, error) {
	if err := ctx.Err(); err != nil {
		return Grant{}, err
	}
	if r == nil || r.grants == nil || r.byAgent == nil {
		return Grant{}, ErrRepositoryUnavailable
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	grant, exists := r.grants[grantID]
	if !exists {
		return Grant{}, ErrRepositoryNotFound
	}
	return cloneGrant(grant), nil
}

func (r *MemoryGrantRepository) GetByAgent(ctx context.Context, agentID string) (Grant, error) {
	if err := ctx.Err(); err != nil {
		return Grant{}, err
	}
	if r == nil || r.grants == nil || r.byAgent == nil {
		return Grant{}, ErrRepositoryUnavailable
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	grantID, exists := r.byAgent[agentID]
	if !exists {
		return Grant{}, ErrRepositoryNotFound
	}
	grant, exists := r.grants[grantID]
	if !exists {
		return Grant{}, ErrRepositoryIntegrity
	}
	return cloneGrant(grant), nil
}

// Revoke is an atomic, one-way state change. Repeated revocation returns conflict.
func (r *MemoryGrantRepository) Revoke(ctx context.Context, grantID, reason string, at time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.grants == nil || r.byAgent == nil {
		return ErrRepositoryUnavailable
	}
	if reason == "" {
		return ErrRepositoryConflict
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	grant, exists := r.grants[grantID]
	if !exists {
		return ErrRepositoryNotFound
	}
	if grant.Revoked || grant.Status != "active" {
		return ErrRepositoryConflict
	}
	revokedAt := at.UTC()
	grant.Revoked = true
	grant.Status = "revoked"
	grant.RevokeReason = reason
	grant.RevokedAt = &revokedAt
	r.grants[grantID] = grant
	return nil
}

func cloneGrant(grant Grant) Grant {
	grant.Capabilities = append([]Capability(nil), grant.Capabilities...)
	if grant.RevokedAt != nil {
		revokedAt := *grant.RevokedAt
		grant.RevokedAt = &revokedAt
	}
	return grant
}

var _ GrantRepository = (*MemoryGrantRepository)(nil)
