// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"fmt"
)

// PostgresLedgerRepository adapts the PostgreSQL ledger's storage API to the
// context-aware LedgerRepository contract used by RepositoryDecisionGateway.
// Reads currently verify the full append-only chain before resolving records;
// indexed, query-specific reads can be introduced without changing the gateway.
type PostgresLedgerRepository struct {
	ledger *PostgresLedger
}

// NewPostgresLedgerRepository wraps an initialized PostgreSQL ledger.
func NewPostgresLedgerRepository(ledger *PostgresLedger) (*PostgresLedgerRepository, error) {
	if ledger == nil || ledger.pool == nil {
		return nil, ErrRepositoryUnavailable
	}
	return &PostgresLedgerRepository{ledger: ledger}, nil
}

// Append commits an event through PostgresLedger and returns the signed BIGINT sequence required by LedgerRepository.
func (r *PostgresLedgerRepository) Append(ctx context.Context, event LedgerEvent) (LedgerEvent, int64, error) {
	if err := ctx.Err(); err != nil {
		return LedgerEvent{}, 0, err
	}
	if r == nil || r.ledger == nil {
		return LedgerEvent{}, 0, ErrRepositoryUnavailable
	}
	committed, sequence, err := r.ledger.Append(ctx, event)
	if err != nil {
		return LedgerEvent{}, 0, err
	}
	if sequence < 1 {
		return LedgerEvent{}, 0, fmt.Errorf("%w: invalid committed ledger sequence %d", ErrRepositoryIntegrity, sequence)
	}
	return committed, sequence, nil
}

// LoadVerified returns the full chain only after PostgresLedger validates its
// sequence continuity, predecessor links, and event hashes.
func (r *PostgresLedgerRepository) LoadVerified(ctx context.Context) ([]LedgerEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r == nil || r.ledger == nil {
		return nil, ErrRepositoryUnavailable
	}
	return r.ledger.Load(ctx)
}

// Decision resolves the canonical authorization decision from verified durable
// evidence. A missing record is distinct from infrastructure/integrity errors.
func (r *PostgresLedgerRepository) Decision(ctx context.Context, decisionID string) (LedgerEvent, error) {
	if decisionID == "" {
		return LedgerEvent{}, ErrRepositoryIntegrity
	}
	events, err := r.LoadVerified(ctx)
	if err != nil {
		return LedgerEvent{}, err
	}
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if event.EventType == "AUTHORIZATION_DECISION" && event.DecisionID == decisionID {
			return cloneLedgerEvent(event), nil
		}
	}
	return LedgerEvent{}, ErrRepositoryNotFound
}

// HasExecution reports whether durable evidence already started or recorded an
// execution for the decision. UNKNOWN remains blocking until reconciliation.
func (r *PostgresLedgerRepository) HasExecution(ctx context.Context, decisionID string) (bool, error) {
	if decisionID == "" {
		return false, ErrRepositoryIntegrity
	}
	events, err := r.LoadVerified(ctx)
	if err != nil {
		return false, err
	}
	for _, event := range events {
		if event.DecisionID == decisionID &&
			(event.EventType == "EXECUTION_STARTED" ||
				event.EventType == "EXECUTION_ALLOWED" ||
				event.EventType == "EXECUTION_UNKNOWN") {
			return true, nil
		}
	}
	return false, nil
}

var _ LedgerRepository = (*PostgresLedgerRepository)(nil)
