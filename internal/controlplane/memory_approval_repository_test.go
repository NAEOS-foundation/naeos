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

func approvalRepositoryFixture() Approval {
	now := time.Now().UTC()
	return Approval{
		ID: "APR-contract-1", DecisionID: "DEC-contract-1", AgentID: "agent-1",
		Capability: "repo.read", PolicyID: "policy-1", PolicyVersion: 1,
		Approver: "reviewer-1", Status: "pending", CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}
}

func TestMemoryApprovalRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryApprovalRepository(NewApprovalStore())
	approval := approvalRepositoryFixture()
	if err := repo.Create(ctx, approval); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := repo.Get(ctx, approval.ID)
	if err != nil || got.Status != "pending" {
		t.Fatalf("Get = (%+v, %v), want pending", got, err)
	}
	at := time.Now().UTC()
	got, err = repo.Transition(ctx, approval.ID, "pending", "approved", at)
	if err != nil || got.Status != "approved" || !got.ApprovedAt.Equal(at) {
		t.Fatalf("approve transition = (%+v, %v)", got, err)
	}
	got, err = repo.Transition(ctx, approval.ID, "approved", "consumed", at.Add(time.Minute))
	if err != nil || got.Status != "consumed" {
		t.Fatalf("consume transition = (%+v, %v)", got, err)
	}
	if _, err := repo.Transition(ctx, approval.ID, "pending", "approved", at); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("stale transition error = %v, want conflict", err)
	}
}

func TestMemoryApprovalRepositoryCreateValidationAndNotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryApprovalRepository(NewApprovalStore())
	approval := approvalRepositoryFixture()
	if err := repo.Create(ctx, approval); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, approval); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("duplicate Create error = %v, want conflict", err)
	}
	invalid := approvalRepositoryFixture()
	invalid.ID = ""
	if err := repo.Create(ctx, invalid); !errors.Is(err, ErrRepositoryConflict) {
		t.Fatalf("invalid Create error = %v, want conflict", err)
	}
	if _, err := repo.Get(ctx, "missing"); !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("missing Get error = %v, want not found", err)
	}
}

func TestMemoryApprovalRepositoryTransitionCompareAndSetConcurrent(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryApprovalRepository(NewApprovalStore())
	approval := approvalRepositoryFixture()
	if err := repo.Create(ctx, approval); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, next := range []string{"approved", "rejected"} {
		wg.Add(1)
		go func(next string) {
			defer wg.Done()
			<-start
			_, err := repo.Transition(ctx, approval.ID, "pending", next, time.Now())
			results <- err
		}(next)
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
			t.Fatalf("unexpected transition error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d, want exactly one winner", successes, conflicts)
	}
}

func TestMemoryApprovalRepositoryContextAndUnavailable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := NewMemoryApprovalRepository(NewApprovalStore())
	if err := repo.Create(ctx, approvalRepositoryFixture()); !errors.Is(err, context.Canceled) {
		t.Fatalf("Create error = %v, want canceled", err)
	}
	if _, err := repo.Get(ctx, "APR"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get error = %v, want canceled", err)
	}
	if _, err := repo.Transition(ctx, "APR", "pending", "approved", time.Now()); !errors.Is(err, context.Canceled) {
		t.Fatalf("Transition error = %v, want canceled", err)
	}

	var unavailable *MemoryApprovalRepository
	background := context.Background()
	if err := unavailable.Create(background, approvalRepositoryFixture()); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable Create error = %v", err)
	}
	if _, err := unavailable.Get(background, "APR"); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable Get error = %v", err)
	}
}

func TestMemoryApprovalRepositoryConcurrentConsumeSingleUse(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryApprovalRepository(NewApprovalStore())
	approval := approvalRepositoryFixture()
	if err := repo.Create(ctx, approval); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if _, err := repo.Transition(ctx, approval.ID, "pending", "approved", now); err != nil {
		t.Fatalf("approve: %v", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.Transition(ctx, approval.ID, "approved", "consumed", now.Add(time.Second))
			results <- err
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
			t.Fatalf("unexpected consume error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("consume winners=%d conflicts=%d; want exactly one winner", successes, conflicts)
	}
	final, err := repo.Get(ctx, approval.ID)
	if err != nil || final.Status != "consumed" {
		t.Fatalf("final approval = (%+v, %v), want consumed", final, err)
	}
}
