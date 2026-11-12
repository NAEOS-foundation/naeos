// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func grantRepositoryFixture() Grant {
	now := time.Now().UTC()
	return Grant{
		GrantID: "GRANT-contract-1", AgentID: "agent-contract-1",
		PolicyID: "policy-contract-1", PolicyVersion: 1,
		Capabilities: []Capability{"repo.read"}, CreatedAt: now,
		ExpiresAt: now.Add(time.Hour), Status: "active",
	}
}

func TestMemoryGrantRepositoryCreateAndLookup(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGrantRepository()
	grant := grantRepositoryFixture()
	if err := repo.Create(ctx, grant); err != nil {
		t.Fatalf("Create: %v", err)
	}
	byID, err := repo.Get(ctx, grant.GrantID)
	if err != nil || byID.AgentID != grant.AgentID || !byID.IsValid(time.Now()) {
		t.Fatalf("Get = (%+v, %v), expected active valid grant", byID, err)
	}
	byAgent, err := repo.GetByAgent(ctx, grant.AgentID)
	if err != nil || byAgent.GrantID != grant.GrantID {
		t.Fatalf("GetByAgent = (%+v, %v)", byAgent, err)
	}
	if err := repo.Create(ctx, grant); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("duplicate ID error = %v, want conflict", err)
	}
	other := grantRepositoryFixture()
	other.GrantID = "GRANT-contract-2"
	if err := repo.Create(ctx, other); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("duplicate agent error = %v, want conflict", err)
	}
}

func TestMemoryGrantRepositoryRevocation(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGrantRepository()
	grant := grantRepositoryFixture()
	if err := repo.Create(ctx, grant); err != nil {
		t.Fatal(err)
	}
	at := time.Now().UTC()
	if err := repo.Revoke(ctx, grant.GrantID, "policy changed", at); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	got, err := repo.Get(ctx, grant.GrantID)
	if err != nil || !got.Revoked || got.Status != "revoked" ||
		got.RevokeReason != "policy changed" || got.RevokedAt == nil ||
		!got.RevokedAt.Equal(at) || got.IsValid(at) {
		t.Fatalf("revoked grant = (%+v, %v)", got, err)
	}
	if err := repo.Revoke(ctx, grant.GrantID, "second reason", at.Add(time.Minute)); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("repeat Revoke error = %v, want conflict", err)
	}
}

func TestMemoryGrantRepositoryConcurrentRevocationAndLookup(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGrantRepository()
	grant := grantRepositoryFixture()
	if err := repo.Create(ctx, grant); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- repo.Revoke(ctx, grant.GrantID, "revoked", time.Now())
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrRepositoryConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected Revoke error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d, want one atomic winner", successes, conflicts)
	}
	if _, err := repo.GetByAgent(ctx, grant.AgentID); err != nil {
		t.Fatalf("GetByAgent after revoke: %v", err)
	}
}

func TestMemoryGrantRepositoryValidationErrorsAndContext(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGrantRepository()
	invalid := grantRepositoryFixture()
	invalid.Status = "revoked"
	if err := repo.Create(ctx, invalid); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("invalid Create error = %v, want conflict", err)
	}
	if _, err := repo.Get(ctx, "missing"); !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("missing Get error = %v, want not found", err)
	}
	if _, err := repo.GetByAgent(ctx, "missing-agent"); !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("missing GetByAgent error = %v, want not found", err)
	}
	if err := repo.Revoke(ctx, "missing", "reason", time.Now()); !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("missing Revoke error = %v, want not found", err)
	}
	if err := repo.Revoke(ctx, "missing", "", time.Now()); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("empty reason error = %v, want conflict", err)
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := repo.Create(canceled, grantRepositoryFixture()); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Create error = %v", err)
	}
	if _, err := repo.Get(canceled, "grant"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Get error = %v", err)
	}
	if _, err := repo.GetByAgent(canceled, "agent"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled GetByAgent error = %v", err)
	}
	if err := repo.Revoke(canceled, "grant", "reason", time.Now()); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Revoke error = %v", err)
	}

	var unavailable *MemoryGrantRepository
	if err := unavailable.Create(ctx, grantRepositoryFixture()); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable Create error = %v", err)
	}
	if _, err := unavailable.Get(ctx, "grant"); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable Get error = %v", err)
	}
	if _, err := unavailable.GetByAgent(ctx, "agent"); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable GetByAgent error = %v", err)
	}
	if err := unavailable.Revoke(ctx, "grant", "reason", time.Now()); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable Revoke error = %v", err)
	}
}

func TestMemoryGrantRepositoryReturnsDefensiveCopies(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryGrantRepository()
	grant := grantRepositoryFixture()
	if err := repo.Create(ctx, grant); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, grant.GrantID)
	if err != nil {
		t.Fatal(err)
	}
	got.Capabilities[0] = "repo.write"
	got.AgentID = "mutated-agent"
	got2, err := repo.Get(ctx, grant.GrantID)
	if err != nil {
		t.Fatal(err)
	}
	if got2.Capabilities[0] != "repo.read" || got2.AgentID != grant.AgentID {
		t.Fatalf("stored grant changed through returned copy: %+v", got2)
	}
}
