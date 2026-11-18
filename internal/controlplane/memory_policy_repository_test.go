// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryPolicyRepositoryImplementsContract(t *testing.T) {
	store := NewPolicyStore()
	repo := NewMemoryPolicyRepository(store)
	ctx := context.Background()
	policy := &Policy{ID: "policy-contract", Version: 1, Status: "active", AllowedCapabilities: []Capability{"repo.read"}}
	if err := repo.PutVersion(ctx, policy); err != nil {
		t.Fatalf("PutVersion: %v", err)
	}
	active, err := repo.Active(ctx, policy.ID)
	if err != nil || active.Version != 1 {
		t.Fatalf("Active = (%+v, %v), want version 1", active, err)
	}
	historical, err := repo.Version(ctx, policy.ID, 1)
	if err != nil || historical.ID != policy.ID {
		t.Fatalf("Version = (%+v, %v), want policy", historical, err)
	}
	if err := repo.CheckActive(ctx, policy.ID, 1); err != nil {
		t.Fatalf("CheckActive current version: %v", err)
	}
	if err := repo.CheckActive(ctx, policy.ID, 2); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("CheckActive stale version error = %v, want conflict", err)
	}
}

func TestMemoryPolicyRepositoryPropagatesContextCancellation(t *testing.T) {
	repo := NewMemoryPolicyRepository(NewPolicyStore())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := repo.PutVersion(ctx, &Policy{ID: "p", Version: 1, Status: "active"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("PutVersion error = %v, want context.Canceled", err)
	}
	if _, err := repo.Active(ctx, "p"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Active error = %v, want context.Canceled", err)
	}
	if _, err := repo.Version(ctx, "p", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("Version error = %v, want context.Canceled", err)
	}
	if err := repo.CheckActive(ctx, "p", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("CheckActive error = %v, want context.Canceled", err)
	}
}

func TestMemoryPolicyRepositoryUnavailable(t *testing.T) {
	var repo *MemoryPolicyRepository
	ctx := context.Background()
	if err := repo.PutVersion(ctx, &Policy{ID: "p", Version: 1, Status: "active"}); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("PutVersion error = %v, want unavailable", err)
	}
	if _, err := repo.Active(ctx, "p"); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("Active error = %v, want unavailable", err)
	}
}
