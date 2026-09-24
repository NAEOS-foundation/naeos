// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package messagequeue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type OutboxStore interface {
	Enqueue(ctx context.Context, msg *Message) error
	Claim(ctx context.Context, workerID string, limit int, lease time.Duration) ([]*Message, error)
	Ack(ctx context.Context, messageID, workerID string) error
	Fail(ctx context.Context, messageID, workerID string, retryAt time.Time, reason string) error
	Migrate(ctx context.Context) error
}

type PostgresOutboxStore struct{ db *sql.DB }

func NewPostgresOutboxStore(db *sql.DB) (*PostgresOutboxStore, error) {
	if db == nil {
		return nil, fmt.Errorf("messagequeue: nil database")
	}
	return &PostgresOutboxStore{db: db}, nil
}

func (s *PostgresOutboxStore) Migrate(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS naeos_message_outbox (
	id TEXT PRIMARY KEY,
	idempotency_key TEXT NOT NULL DEFAULT '',
	topic TEXT NOT NULL,
	payload TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	retries INTEGER NOT NULL DEFAULT 0,
	max_retries INTEGER NOT NULL DEFAULT 3,
	status TEXT NOT NULL DEFAULT 'pending',
	available_at TIMESTAMPTZ NOT NULL,
	locked_by TEXT NOT NULL DEFAULT '',
	locked_until TIMESTAMPTZ NULL,
	processed_at TIMESTAMPTZ NULL,
	last_error TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_naeos_message_outbox_claim
	ON naeos_message_outbox (status, available_at, locked_until);
CREATE UNIQUE INDEX IF NOT EXISTS idx_naeos_message_outbox_idempotency
	ON naeos_message_outbox (idempotency_key)
	WHERE idempotency_key <> '';
`
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("messagequeue: migrate outbox: %w", err)
	}
	return nil
}

func (s *PostgresOutboxStore) Enqueue(ctx context.Context, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("messagequeue: nil message")
	}
	if msg.ID == "" {
		msg.ID = generateID()
	}
	if msg.Topic == "" {
		return fmt.Errorf("messagequeue: message topic is required")
	}
	if msg.MaxRetries <= 0 {
		msg.MaxRetries = 3
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("messagequeue: marshal payload: %w", err)
	}
	const query = `
INSERT INTO naeos_message_outbox
	(id, idempotency_key, topic, payload, created_at, retries, max_retries, available_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $5)
`
	if _, err := s.db.ExecContext(ctx, query, msg.ID, msg.IdempotencyKey, msg.Topic, string(payload), msg.Timestamp, msg.Retries, msg.MaxRetries); err != nil {
		return fmt.Errorf("messagequeue: enqueue %s: %w", msg.ID, err)
	}
	return nil
}

func (s *PostgresOutboxStore) Claim(ctx context.Context, workerID string, limit int, lease time.Duration) ([]*Message, error) {
	if workerID == "" {
		return nil, fmt.Errorf("messagequeue: worker id is required")
	}
	if limit <= 0 {
		limit = 1
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("messagequeue: begin claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
SELECT id, idempotency_key, topic, payload, created_at, retries, max_retries
FROM naeos_message_outbox
WHERE (status = 'pending' AND available_at <= NOW())
   OR (status = 'processing' AND locked_until < NOW())
ORDER BY created_at, id
FOR UPDATE SKIP LOCKED
LIMIT $1
`
	rows, err := tx.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("messagequeue: select claimable messages: %w", err)
	}
	defer rows.Close()

	type claimed struct {
		id, key, topic, payload string
		created                 time.Time
		retries, maxRetries     int
		decoded                 any
	}
	var messages []claimed
	for rows.Next() {
		var m claimed
		if err := rows.Scan(&m.id, &m.key, &m.topic, &m.payload, &m.created, &m.retries, &m.maxRetries); err != nil {
			return nil, fmt.Errorf("messagequeue: scan claimable message: %w", err)
		}
		if err := json.Unmarshal([]byte(m.payload), &m.decoded); err != nil {
			return nil, fmt.Errorf("messagequeue: decode payload %s: %w", m.id, err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("messagequeue: iterate claimable messages: %w", err)
	}

	lockedUntil := time.Now().UTC().Add(lease)
	for _, m := range messages {
		if _, err := tx.ExecContext(ctx, `
UPDATE naeos_message_outbox
SET status = 'processing', locked_by = $1, locked_until = $2
WHERE id = $3`, workerID, lockedUntil, m.id); err != nil {
			return nil, fmt.Errorf("messagequeue: lease %s: %w", m.id, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("messagequeue: commit claim: %w", err)
	}

	out := make([]*Message, 0, len(messages))
	for _, m := range messages {
		out = append(out, &Message{ID: m.id, IdempotencyKey: m.key, Topic: m.topic, Payload: m.decoded, Timestamp: m.created, Retries: m.retries, MaxRetries: m.maxRetries})
	}
	return out, nil
}

func (s *PostgresOutboxStore) Ack(ctx context.Context, messageID, workerID string) error {
	if messageID == "" {
		return fmt.Errorf("messagequeue: message id is required")
	}
	if workerID == "" {
		return fmt.Errorf("messagequeue: worker id is required")
	}
	const query = `
UPDATE naeos_message_outbox
SET status = 'done', processed_at = NOW(), locked_by = '', locked_until = NULL, last_error = ''
WHERE id = $1 AND status = 'processing' AND locked_by = $2 AND locked_until > NOW()
`
	result, err := s.db.ExecContext(ctx, query, messageID, workerID)
	if err != nil {
		return fmt.Errorf("messagequeue: ack %s: %w", messageID, err)
	}
	if n, err := result.RowsAffected(); err == nil && n == 0 {
		return fmt.Errorf("messagequeue: ack %s: lease not owned or expired", messageID)
	}
	return nil
}

func (s *PostgresOutboxStore) Fail(ctx context.Context, messageID, workerID string, retryAt time.Time, reason string) error {
	if messageID == "" {
		return fmt.Errorf("messagequeue: message id is required")
	}
	if workerID == "" {
		return fmt.Errorf("messagequeue: worker id is required")
	}
	const query = `
UPDATE naeos_message_outbox
SET status = CASE WHEN retries + 1 >= max_retries THEN 'dead' ELSE 'pending' END,
    available_at = CASE WHEN retries + 1 >= max_retries THEN available_at ELSE $1 END,
    retries = retries + 1, locked_by = '', locked_until = NULL, last_error = $2
WHERE id = $3 AND status = 'processing' AND locked_by = $4 AND locked_until > NOW()
`
	result, err := s.db.ExecContext(ctx, query, retryAt, reason, messageID, workerID)
	if err != nil {
		return fmt.Errorf("messagequeue: fail %s: %w", messageID, err)
	}
	if n, err := result.RowsAffected(); err == nil && n == 0 {
		return fmt.Errorf("messagequeue: fail %s: lease not owned or expired", messageID)
	}
	return nil
}
