// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// ClaimExecution consumes the exact approved record and appends EXECUTION_STARTED
// in the same PostgreSQL transaction. A crash can therefore commit both effects
// or neither; it cannot leave an unpaired consumed approval or execution intent.
func (r *PostgresLedgerRepository) ClaimExecution(ctx context.Context, event LedgerEvent, approval Approval, at time.Time) (LedgerEvent, int64, error) {
	if err := ctx.Err(); err != nil {
		return LedgerEvent{}, 0, err
	}
	if r == nil || r.ledger == nil || r.ledger.pool == nil {
		return LedgerEvent{}, 0, ErrRepositoryUnavailable
	}
	if approval.ID == "" || approval.DecisionID == "" || event.EventType != "EXECUTION_STARTED" ||
		event.DecisionID != approval.DecisionID || event.AgentID != approval.AgentID ||
		event.Capability != approval.Capability || event.ArtifactHash != approval.ArtifactHash {
		return LedgerEvent{}, 0, ErrRepositoryConflict
	}
	at = at.UTC()
	if event.ID == "" {
		event.ID = controlPlaneID("EVT")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = at
	} else {
		event.Timestamp = event.Timestamp.UTC()
	}

	const maxPredecessorRetries = 64
	for attempt := 0; attempt < maxPredecessorRetries; attempt++ {
		tx, err := r.ledger.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return LedgerEvent{}, 0, fmt.Errorf("%w: begin atomic execution claim: %w", ErrRepositoryUnavailable, err)
		}

		var consumedID string
		err = tx.QueryRow(ctx, `UPDATE naeos_control_plane.approvals
			SET status = 'consumed', consumed_at = $8
			WHERE approval_id = $1 AND decision_id = $2 AND agent_id = $3
			  AND capability = $4 AND policy_id = $5 AND policy_version = $6
			  AND artifact_hash = $7 AND status = 'approved'
			  AND (expires_at IS NULL OR expires_at > $8)
			RETURNING approval_id`,
			approval.ID, approval.DecisionID, approval.AgentID, string(approval.Capability),
			approval.PolicyID, approval.PolicyVersion, approval.ArtifactHash, at).Scan(&consumedID)
		if errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			return LedgerEvent{}, 0, ErrRepositoryConflict
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return LedgerEvent{}, 0, postgresRepositoryError(err)
		}

		var previousHash string
		if err := tx.QueryRow(ctx, `SELECT event_hash
			FROM naeos_control_plane.read_ledger_head()`).Scan(&previousHash); err != nil {
			_ = tx.Rollback(ctx)
			return LedgerEvent{}, 0, postgresRepositoryError(err)
		}
		candidate := cloneLedgerEvent(event)
		candidate.PreviousHash = previousHash
		candidate.EventHash = hashLedgerEvent(candidate)
		payload, err := json.Marshal(candidate)
		if err != nil {
			_ = tx.Rollback(ctx)
			return LedgerEvent{}, 0, fmt.Errorf("%w: encode atomic execution intent: %w", ErrRepositoryIntegrity, err)
		}

		var sequence int64
		err = tx.QueryRow(ctx, `SELECT naeos_control_plane.append_ledger_event($1::jsonb, $2, $3)`,
			payload, previousHash, candidate.EventHash).Scan(&sequence)
		if err != nil {
			_ = tx.Rollback(ctx)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "P0001" &&
				pgErr.Message == "ledger predecessor mismatch" && attempt+1 < maxPredecessorRetries {
				continue
			}
			if errors.As(err, &pgErr) && pgErr.Code == "P0001" &&
				pgErr.Message == "execution already claimed for decision" {
				return LedgerEvent{}, 0, ErrRepositoryConflict
			}
			return LedgerEvent{}, 0, postgresRepositoryError(err)
		}
		if err := tx.Commit(ctx); err != nil {
			return LedgerEvent{}, 0, fmt.Errorf("%w: commit atomic execution claim: %w", ErrRepositoryUnavailable, err)
		}
		if sequence < 1 {
			return LedgerEvent{}, 0, fmt.Errorf("%w: invalid committed ledger sequence %d", ErrRepositoryIntegrity, sequence)
		}
		return candidate, sequence, nil
	}
	return LedgerEvent{}, 0, fmt.Errorf("%w: ledger head changed too frequently during execution claim", ErrRepositoryUnavailable)
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

var (
	_ LedgerRepository                 = (*PostgresLedgerRepository)(nil)
	_ AtomicExecutionClaimRepository   = (*PostgresLedgerRepository)(nil)
)
