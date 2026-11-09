// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"fmt"
	"sync"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/database"
)

// InvocationStore is the durable execution-boundary primitive for replay
// protection. Claim must be atomic: exactly one concurrent caller may claim a
// previously unseen invocation ID.
type InvocationStore interface {
	Claim(invocationID string) (bool, error)
}

// InMemoryInvocationStore is suitable for tests and explicitly local
// deployments. It does not provide restart or cross-replica protection.
type InMemoryInvocationStore struct {
	mu      sync.Mutex
	claimed map[string]struct{}
}

func NewInMemoryInvocationStore() *InMemoryInvocationStore {
	return &InMemoryInvocationStore{claimed: make(map[string]struct{})}
}

func (s *InMemoryInvocationStore) Claim(invocationID string) (bool, error) {
	if invocationID == "" {
		return false, fmt.Errorf("invocation id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.claimed[invocationID]; exists {
		return false, nil
	}
	s.claimed[invocationID] = struct{}{}
	return true, nil
}

// DatabaseInvocationStore persists consumed invocation IDs in the configured
// NAEOS database. The unique key is the concurrency boundary shared by
// replicas. Consumed IDs are retained until explicitly purged; this avoids
// silently reopening a replay window after process restart.
type DatabaseInvocationStore struct {
	db database.Database

	initMu       sync.Mutex
	tableReady   bool
}

func NewDatabaseInvocationStore(db database.Database) *DatabaseInvocationStore {
	return &DatabaseInvocationStore{db: db}
}

func (s *DatabaseInvocationStore) ensureTable() error {
	s.initMu.Lock()
	defer s.initMu.Unlock()

	if s.tableReady {
		return nil
	}
	if s.db == nil {
		return fmt.Errorf("invocation store database is nil")
	}
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS naeos_invocation_replay (
		invocation_id VARCHAR(255) PRIMARY KEY,
		consumed_at BIGINT NOT NULL
	)`); err != nil {
		// Do not cache transient initialization failures. A later claim may
		// succeed after the database becomes available.
		return err
	}
	s.tableReady = true
	return nil
}

func (s *DatabaseInvocationStore) Claim(invocationID string) (bool, error) {
	if invocationID == "" {
		return false, fmt.Errorf("invocation id is required")
	}
	if s.db == nil {
		return false, fmt.Errorf("invocation store database is nil")
	}

	// RealSupabase REST mode ignores Exec arguments and therefore cannot safely
	// execute the parameterized replay claim. Fail closed rather than silently
	// dropping the invocation ID or interpolating it into SQL.
	if restModeDB, ok := s.db.(interface{ IsRESTMode() bool }); ok && restModeDB.IsRESTMode() {
		return false, fmt.Errorf("durable invocation replay is unsupported in Supabase REST mode; use a direct database connection")
	}

	if err := s.ensureTable(); err != nil {
		return false, err
	}

	var query string
	switch s.db.Name() {
	case "mysql":
		query = "INSERT IGNORE INTO naeos_invocation_replay (invocation_id, consumed_at) VALUES (?, ?)"
	case "postgresql", "sqlite", "supabase":
		query = "INSERT INTO naeos_invocation_replay (invocation_id, consumed_at) VALUES (?, ?) ON CONFLICT (invocation_id) DO NOTHING"
	default:
		return false, fmt.Errorf("unsupported database for durable invocation replay: %s", s.db.Name())
	}

	result, err := s.db.Exec(query, invocationID, time.Now().Unix())
	if err != nil {
		return false, err
	}
	return result.RowsAffected == 1, nil
}
