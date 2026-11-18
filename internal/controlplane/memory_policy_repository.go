// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "context"

// MemoryPolicyRepository adapts the in-process PolicyStore to the context-aware
// repository contract. It is intended for tests and demo composition only;
// production deployments must use a durable implementation.
type MemoryPolicyRepository struct {
	store *PolicyStore
}

// NewMemoryPolicyRepository returns a repository backed by the supplied store.
func NewMemoryPolicyRepository(store *PolicyStore) *MemoryPolicyRepository {
	return &MemoryPolicyRepository{store: store}
}

func (r *MemoryPolicyRepository) PutVersion(ctx context.Context, policy *Policy) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.store == nil {
		return ErrRepositoryUnavailable
	}
	return r.store.Set(policy)
}

func (r *MemoryPolicyRepository) Active(ctx context.Context, policyID string) (*Policy, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r == nil || r.store == nil {
		return nil, ErrRepositoryUnavailable
	}
	return r.store.Active(policyID)
}

func (r *MemoryPolicyRepository) Version(ctx context.Context, policyID string, version int) (*Policy, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r == nil || r.store == nil {
		return nil, ErrRepositoryUnavailable
	}
	return r.store.Version(policyID, version)
}

func (r *MemoryPolicyRepository) CheckActive(ctx context.Context, policyID string, version int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.store == nil {
		return ErrRepositoryUnavailable
	}
	active, err := r.store.Active(policyID)
	if err != nil {
		return err
	}
	if active.Version != version {
		return ErrRepositoryConflict
	}
	return nil
}

var _ PolicyRepository = (*MemoryPolicyRepository)(nil)
