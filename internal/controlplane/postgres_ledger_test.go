// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPostgresLedgerRequiresPool(t *testing.T) {
	if _, err := NewPostgresLedger(nil); err == nil {
		t.Fatal("expected nil pool to be rejected")
	}
}

// This test is opt-in and must point to a disposable isolated PostgreSQL
// database where the production_control_plane migration has been applied.
// It intentionally appends records; it never truncates or rewrites the ledger.
func TestPostgresLedgerIntegrationConcurrentAppendAndReload(t *testing.T) {
	databaseURL := os.Getenv("NAEOS_CONTROLPLANE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set NAEOS_CONTROLPLANE_TEST_DATABASE_URL to an isolated migrated PostgreSQL test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test database pool: %v", err)
	}
	defer func() {
		if pool != nil {
			pool.Close()
		}
	}()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	store, err := NewPostgresLedger(pool)
	if err != nil {
		t.Fatal(err)
	}

	var canReadHead, canUpdateHead, canInsertLedger, canUpdateLedger, canDeleteLedger bool
	err = pool.QueryRow(ctx, `SELECT
		has_table_privilege(current_user, 'naeos_control_plane.ledger_head', 'SELECT'),
		has_table_privilege(current_user, 'naeos_control_plane.ledger_head', 'UPDATE'),
		has_table_privilege(current_user, 'naeos_control_plane.ledger_events', 'INSERT'),
		has_table_privilege(current_user, 'naeos_control_plane.ledger_events', 'UPDATE'),
		has_table_privilege(current_user, 'naeos_control_plane.ledger_events', 'DELETE')`).Scan(
		&canReadHead, &canUpdateHead, &canInsertLedger, &canUpdateLedger, &canDeleteLedger)
	if err != nil {
		t.Fatalf("inspect runtime table privileges: %v", err)
	}
	if canReadHead || canUpdateHead {
		t.Fatalf("runtime must not access ledger_head directly: SELECT=%t UPDATE=%t", canReadHead, canUpdateHead)
	}
	if canInsertLedger || canUpdateLedger || canDeleteLedger {
		t.Fatalf("runtime must not mutate ledger_events directly: INSERT=%t UPDATE=%t DELETE=%t",
			canInsertLedger, canUpdateLedger, canDeleteLedger)
	}

	const count = 12
	sequences := make([]int64, count)
	errs := make([]error, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			event, sequence, appendErr := store.Append(ctx, LedgerEvent{
				ID:        controlPlaneID("PGTEST"),
				AgentID:   "integration-test",
				EventType: "POSTGRES_LEDGER_INTEGRATION_TEST",
				Metadata:  map[string]string{"index": fmt.Sprintf("%d", index)},
			})
			if appendErr == nil && (event.ID == "" || event.EventHash == "") {
				appendErr = fmt.Errorf("append returned event without ID/hash")
			}
			sequences[index] = sequence
			errs[index] = appendErr
		}(i)
	}
	wg.Wait()
	for i, appendErr := range errs {
		if appendErr != nil {
			t.Fatalf("append %d failed: %v", i, appendErr)
		}
	}
	sorted := append([]int64(nil), sequences...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	for i, sequence := range sorted {
		if sequence != int64(i+1) {
			t.Fatalf("concurrent append sequence is not contiguous from 1: %v", sorted)
		}
	}

	events, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("load and verify ledger after concurrent appends: %v", err)
	}
	if len(events) != count {
		t.Fatalf("loaded %d events; expected exactly %d in isolated database", len(events), count)
	}

	// Simulate a process restart by closing the original pool and loading
	// the committed ledger through a fresh connection pool.
	pool.Close()
	pool = nil
	restartedPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create restarted database pool: %v", err)
	}
	defer restartedPool.Close()
	restartedStore, err := NewPostgresLedger(restartedPool)
	if err != nil {
		t.Fatal(err)
	}
	restartedEvents, err := restartedStore.Load(ctx)
	if err != nil {
		t.Fatalf("reload ledger after pool restart: %v", err)
	}
	if len(restartedEvents) != count {
		t.Fatalf("restart recovered %d events; expected %d", len(restartedEvents), count)
	}
}
