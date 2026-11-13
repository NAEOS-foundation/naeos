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
		return nil, fmt.Errorf("%w: postgres ledger pool is missing", ErrRepositoryUnavailable)
	}

	// Read events and the authoritative head from one repeatable-read snapshot.
	// Without a shared snapshot, a concurrent append between the two reads could
	// make a healthy ledger appear inconsistent (or hide a stale head).
	tx, err := l.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: begin ledger verification: %w", ErrRepositoryUnavailable, err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	rows, err := tx.Query(ctx, `SELECT sequence, event_payload
		FROM naeos_control_plane.ledger_events
		ORDER BY sequence ASC`)
	if err != nil {
		return nil, fmt.Errorf("%w: query ledger events: %w", ErrRepositoryUnavailable, err)
	}

	events := make([]LedgerEvent, 0)
	var expectedSequence int64 = 1
	var previousHash string
	for rows.Next() {
		var sequence int64
		var payload []byte
		if err := rows.Scan(&sequence, &payload); err != nil {
			rows.Close()
			return nil, fmt.Errorf("%w: scan ledger event: %w", ErrRepositoryUnavailable, err)
		}
		if sequence != expectedSequence {
			rows.Close()
			return nil, fmt.Errorf("%w: ledger sequence gap: got %d, want %d", ErrRepositoryIntegrity, sequence, expectedSequence)
		}
		var event LedgerEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			rows.Close()
			return nil, fmt.Errorf("%w: decode ledger event at sequence %d: %w", ErrRepositoryIntegrity, sequence, err)
		}
		if event.ID == "" || event.Timestamp.IsZero() || event.EventHash == "" {
			rows.Close()
			return nil, fmt.Errorf("%w: invalid ledger event at sequence %d", ErrRepositoryIntegrity, sequence)
		}
		if event.PreviousHash != previousHash {
			rows.Close()
			return nil, fmt.Errorf("%w: ledger predecessor mismatch at sequence %d", ErrRepositoryIntegrity, sequence)
		}
		if event.EventHash != hashLedgerEvent(event) {
			rows.Close()
			return nil, fmt.Errorf("%w: ledger event hash mismatch at sequence %d", ErrRepositoryIntegrity, sequence)
		}
		events = append(events, cloneLedgerEvent(event))
		previousHash = event.EventHash
		expectedSequence++
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return nil, fmt.Errorf("%w: iterate ledger events: %w", ErrRepositoryUnavailable, rowsErr)
	}

	var headSequence int64
	var headHash string
	if err := tx.QueryRow(ctx, `SELECT last_sequence, event_hash
		FROM naeos_control_plane.read_ledger_head()`).Scan(&headSequence, &headHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: authoritative ledger head is missing", ErrRepositoryIntegrity)
		}
		return nil, fmt.Errorf("%w: read authoritative ledger head: %w", ErrRepositoryUnavailable, err)
	}
	if headSequence != int64(len(events)) || headHash != previousHash {
		return nil, fmt.Errorf("%w: ledger head mismatch: head=(%d,%s), events=(%d,%s)",
			ErrRepositoryIntegrity, headSequence, headHash, len(events), previousHash)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: commit ledger verification snapshot: %w", ErrRepositoryUnavailable, err)
	}
	return events, nil
}
