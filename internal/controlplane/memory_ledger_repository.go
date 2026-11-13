// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"fmt"
	"sync"
)

// MemoryLedgerRepository adapts the in-memory Ledger to the context-aware
// repository contract. It is intended for tests and the investor demo only.
// Its sequence values are process-local and provide no cross-process durability.
type MemoryLedgerRepository struct {
	ledger *Ledger
	mu     sync.Mutex
}

func NewMemoryLedgerRepository(ledger *Ledger) *MemoryLedgerRepository {
	return &MemoryLedgerRepository{ledger: ledger}
}

func (r *MemoryLedgerRepository) Append(ctx context.Context, event LedgerEvent) (LedgerEvent, int64, error) {
	if err := ctx.Err(); err != nil {
		return LedgerEvent{}, 0, err
	}
	if r == nil || r.ledger == nil {
		return LedgerEvent{}, 0, ErrRepositoryUnavailable
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	stored := r.ledger.Append(event)
	if stored.ID == "" {
		if err := r.ledger.PersistenceError(); err != nil {
			return LedgerEvent{}, 0, fmt.Errorf("%w: append event: %w", ErrRepositoryUnavailable, err)
		}
		return LedgerEvent{}, 0, ErrRepositoryConflict
	}
	return stored, int64(len(r.ledger.Events())), nil
}

func (r *MemoryLedgerRepository) LoadVerified(ctx context.Context) ([]LedgerEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r == nil || r.ledger == nil {
		return nil, ErrRepositoryUnavailable
	}
	events := r.ledger.Events()
	var previousHash string
	for index, event := range events {
		if event.ID == "" || event.Timestamp.IsZero() || event.EventHash == "" {
			return nil, fmt.Errorf("%w: invalid event at sequence %d", ErrRepositoryIntegrity, index+1)
		}
		if event.PreviousHash != previousHash || event.EventHash != hashLedgerEvent(event) {
			return nil, fmt.Errorf("%w: hash chain mismatch at sequence %d", ErrRepositoryIntegrity, index+1)
		}
		previousHash = event.EventHash
	}
	return events, nil
}

func (r *MemoryLedgerRepository) Decision(ctx context.Context, decisionID string) (LedgerEvent, error) {
	if err := ctx.Err(); err != nil {
		return LedgerEvent{}, err
	}
	if r == nil || r.ledger == nil {
		return LedgerEvent{}, ErrRepositoryUnavailable
	}
	event, ok := r.ledger.Decision(decisionID)
	if !ok {
		return LedgerEvent{}, ErrRepositoryNotFound
	}
	return event, nil
}

func (r *MemoryLedgerRepository) HasExecution(ctx context.Context, decisionID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if r == nil || r.ledger == nil {
		return false, ErrRepositoryUnavailable
	}
	return r.ledger.HasExecution(decisionID), nil
}

var _ LedgerRepository = (*MemoryLedgerRepository)(nil)
