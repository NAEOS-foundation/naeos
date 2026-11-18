// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"fmt"
)

// LoadVerified implements the repository contract name for the PostgreSQL
// ledger's full-chain verification operation.
func (l *PostgresLedger) LoadVerified(ctx context.Context) ([]LedgerEvent, error) {
	return l.Load(ctx)
}

// Decision returns the canonical authorization decision for decisionID from
// the verified ledger. Missing records are distinguished from storage errors.
func (l *PostgresLedger) Decision(ctx context.Context, decisionID string) (LedgerEvent, error) {
	if decisionID == "" {
		return LedgerEvent{}, ErrRepositoryNotFound
	}
	events, err := l.LoadVerified(ctx)
	if err != nil {
		return LedgerEvent{}, fmt.Errorf("verify ledger for decision lookup: %w", err)
	}
	for index := len(events) - 1; index >= 0; index-- {
		event := events[index]
		if event.EventType == "AUTHORIZATION_DECISION" && event.DecisionID == decisionID {
			return event, nil
		}
	}
	return LedgerEvent{}, ErrRepositoryNotFound
}

// HasExecution reports whether a verified ledger contains an execution intent
// or durable execution outcome for the decision.
func (l *PostgresLedger) HasExecution(ctx context.Context, decisionID string) (bool, error) {
	if decisionID == "" {
		return false, nil
	}
	events, err := l.LoadVerified(ctx)
	if err != nil {
		return false, fmt.Errorf("verify ledger for execution lookup: %w", err)
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

var _ LedgerRepository = (*PostgresLedger)(nil)
