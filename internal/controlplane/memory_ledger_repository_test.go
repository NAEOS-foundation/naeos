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

func TestMemoryLedgerRepositoryContract(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryLedgerRepository(NewLedger())

	decision, sequence, err := repo.Append(ctx, LedgerEvent{
		ID: "decision-event", DecisionID: "decision-1", AgentID: "agent-1",
		EventType: "AUTHORIZATION_DECISION", Decision: DecisionAllow,
	})
	if err != nil || sequence != 1 || decision.ID != "decision-event" {
		t.Fatalf("Append decision = (%+v, %d, %v)", decision, sequence, err)
	}
	if got, err := repo.Decision(ctx, "decision-1"); err != nil || got.ID != decision.ID {
		t.Fatalf("Decision = (%+v, %v), expected canonical event", got, err)
	}
	if exists, err := repo.HasExecution(ctx, "decision-1"); err != nil || exists {
		t.Fatalf("HasExecution before intent = (%t, %v), want false", exists, err)
	}
	if _, _, err := repo.Append(ctx, LedgerEvent{
		ID: "execution-start", DecisionID: "decision-1", ExecutionID: "execution-1",
		AgentID: "agent-1", EventType: "EXECUTION_STARTED",
	}); err != nil {
		t.Fatalf("Append execution intent: %v", err)
	}
	if exists, err := repo.HasExecution(ctx, "decision-1"); err != nil || !exists {
		t.Fatalf("HasExecution after intent = (%t, %v), want true", exists, err)
	}
	loaded, err := repo.LoadVerified(ctx)
	if err != nil || len(loaded) != 2 {
		t.Fatalf("LoadVerified = (%d events, %v), want two events", len(loaded), err)
	}
	if _, err := repo.Decision(ctx, "missing"); !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("missing Decision error = %v, want not found", err)
	}
}

func TestMemoryLedgerRepositoryConcurrentAppendSequences(t *testing.T) {
	repo := NewMemoryLedgerRepository(NewLedger())
	const count = 32
	start := make(chan struct{})
	sequences := make([]int64, count)
	errs := make([]error, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			_, sequences[index], errs[index] = repo.Append(context.Background(), LedgerEvent{
				ID: "event-" + time.Now().UTC().Format("150405.000000000") + string(rune('a'+index)),
				AgentID: "agent", EventType: "TEST_EVENT",
			})
		}(i)
	}
	close(start)
	wg.Wait()
	seen := make(map[int64]bool, count)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
		if sequences[i] < 1 || sequences[i] > count || seen[sequences[i]] {
			t.Fatalf("duplicate/out-of-range sequence %d; all=%v", sequences[i], sequences)
		}
		seen[sequences[i]] = true
	}
}

func TestMemoryLedgerRepositoryUnavailableAndCanceled(t *testing.T) {
	ctx := context.Background()
	var unavailable *MemoryLedgerRepository
	if _, _, err := unavailable.Append(ctx, LedgerEvent{}); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable Append error = %v", err)
	}
	if _, err := unavailable.LoadVerified(ctx); !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("unavailable LoadVerified error = %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	repo := NewMemoryLedgerRepository(NewLedger())
	if _, _, err := repo.Append(canceled, LedgerEvent{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Append error = %v", err)
	}
	if _, err := repo.LoadVerified(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled LoadVerified error = %v", err)
	}
	if _, err := repo.Decision(canceled, "decision"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Decision error = %v", err)
	}
	if _, err := repo.HasExecution(canceled, "decision"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled HasExecution error = %v", err)
	}
}
