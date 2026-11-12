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
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresLedger persists Control Plane ledger events in the private
// naeos_control_plane schema. It deliberately does not replace the in-memory
// Ledger until the production composition root is wired to this repository.
type PostgresLedger struct {
	pool *pgxpool.Pool
}

// NewPostgresLedger creates a repository over an already configured pool.
// Pool construction, TLS policy, and credential sourcing belong to the
// production composition root.
func NewPostgresLedger(pool *pgxpool.Pool) (*PostgresLedger, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres ledger pool is required")
	}
	return &PostgresLedger{pool: pool}, nil
}

// Append uses the SECURITY DEFINER head-reader and append functions so the
// runtime role does not need direct privileges on ledger_head. Concurrent
// writers can observe the same predecessor; the append function locks and
// validates the head, and this method retries only that stale-predecessor case.
func (l *PostgresLedger) Append(ctx context.Context, event LedgerEvent) (LedgerEvent, int64, error) {
	if l == nil || l.pool == nil {
		return LedgerEvent{}, 0, fmt.Errorf("postgres ledger unavailable")
	}
	if event.EventType == "" {
		return LedgerEvent{}, 0, fmt.Errorf("ledger event type is required")
	}
	if event.EventType == "EXECUTION_RECONCILED" {
		return LedgerEvent{}, 0, fmt.Errorf("reconciliation events must use the reconciliation workflow")
	}
	if event.ID == "" {
		event.ID = controlPlaneID("EVT")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	} else {
		event.Timestamp = event.Timestamp.UTC()
	}

	const maxPredecessorRetries = 64
	for attempt := 0; attempt < maxPredecessorRetries; attempt++ {
		tx, err := l.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return LedgerEvent{}, 0, fmt.Errorf("begin ledger append: %w", err)
		}

		var previousHash string
		if err := tx.QueryRow(ctx, `SELECT event_hash
			FROM naeos_control_plane.read_ledger_head()`).Scan(&previousHash); err != nil {
			_ = tx.Rollback(ctx)
			return LedgerEvent{}, 0, fmt.Errorf("read ledger head: %w", err)
		}

		candidate := event
		candidate.PreviousHash = previousHash
		candidate.EventHash = hashLedgerEvent(candidate)
		payload, err := json.Marshal(candidate)
		if err != nil {
			_ = tx.Rollback(ctx)
			return LedgerEvent{}, 0, fmt.Errorf("marshal ledger event: %w", err)
		}

		var sequence int64
		err = tx.QueryRow(ctx, `SELECT naeos_control_plane.append_ledger_event($1::jsonb, $2, $3)`,
			payload, previousHash, candidate.EventHash).Scan(&sequence)
		if err != nil {
			_ = tx.Rollback(ctx)
			var pgErr *pgconn.PgError
			if attempt+1 < maxPredecessorRetries && errors.As(err, &pgErr) &&
				pgErr.Code == "P0001" && pgErr.Message == "ledger predecessor mismatch" {
				continue
			}
			return LedgerEvent{}, 0, fmt.Errorf("append ledger event: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return LedgerEvent{}, 0, fmt.Errorf("commit ledger append: %w", err)
		}
		return cloneLedgerEvent(candidate), sequence, nil
	}
	return LedgerEvent{}, 0, fmt.Errorf("append ledger event: predecessor changed too frequently")
}

// Load verifies and returns the committed ledger in sequence order. Any
// sequence gap, broken predecessor, or event-hash mismatch fails closed.
func (l *PostgresLedger) Load(ctx context.Context) ([]LedgerEvent, error) {
	if l == nil || l.pool == nil {
		return nil, fmt.Errorf("postgres ledger unavailable")
	}
	rows, err := l.pool.Query(ctx, `SELECT sequence, event_payload
		FROM naeos_control_plane.ledger_events
		ORDER BY sequence ASC`)
	if err != nil {
		return nil, fmt.Errorf("query ledger events: %w", err)
	}
	defer rows.Close()

	events := make([]LedgerEvent, 0)
	var expectedSequence int64 = 1
	var previousHash string
	for rows.Next() {
		var sequence int64
		var payload []byte
		if err := rows.Scan(&sequence, &payload); err != nil {
			return nil, fmt.Errorf("scan ledger event: %w", err)
		}
		if sequence != expectedSequence {
			return nil, fmt.Errorf("ledger sequence gap: got %d, want %d", sequence, expectedSequence)
		}
		var event LedgerEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, fmt.Errorf("decode ledger event at sequence %d: %w", sequence, err)
		}
		if event.ID == "" || event.Timestamp.IsZero() || event.EventHash == "" {
			return nil, fmt.Errorf("invalid ledger event at sequence %d", sequence)
		}
		if event.PreviousHash != previousHash {
			return nil, fmt.Errorf("ledger predecessor mismatch at sequence %d", sequence)
		}
		if event.EventHash != hashLedgerEvent(event) {
			return nil, fmt.Errorf("ledger event hash mismatch at sequence %d", sequence)
		}
		events = append(events, cloneLedgerEvent(event))
		previousHash = event.EventHash
		expectedSequence++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ledger events: %w", err)
	}
	return events, nil
}
