// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
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

	decisionID := controlPlaneID("PGDECISION")
	decisionEvent, _, err := store.Append(ctx, LedgerEvent{
		ID: controlPlaneID("PGDECISIONEVENT"), DecisionID: decisionID,
		AgentID: "integration-test", EventType: "AUTHORIZATION_DECISION", Decision: DecisionAllow,
	})
	if err != nil {
		t.Fatalf("append canonical decision event: %v", err)
	}
	canonicalDecision, err := store.Decision(ctx, decisionID)
	if err != nil || canonicalDecision.ID != decisionEvent.ID {
		t.Fatalf("Decision = (%+v, %v), want canonical decision event", canonicalDecision, err)
	}
	if exists, err := store.HasExecution(ctx, decisionID); err != nil || exists {
		t.Fatalf("HasExecution before execution intent = (%t, %v), want false", exists, err)
	}
	// Race two independent execution claims for the same decision. The
	// database ledger-head lock must make this a single-winner operation.
	const competingClaims = 2
	claimErrors := make([]error, competingClaims)
	var claimWG sync.WaitGroup
	for i := 0; i < competingClaims; i++ {
		claimWG.Add(1)
		go func(index int) {
			defer claimWG.Done()
			_, _, claimErrors[index] = store.Append(ctx, LedgerEvent{
				ID: controlPlaneID("PGEXECUTION"), DecisionID: decisionID,
				ExecutionID: controlPlaneID("PGEXECUTIONID"), AgentID: "integration-test",
				EventType: "EXECUTION_STARTED",
				Metadata:  map[string]string{"claim": fmt.Sprintf("%d", index)},
			})
		}(i)
	}
	claimWG.Wait()
	successes, conflicts := 0, 0
	for _, claimErr := range claimErrors {
		switch {
		case claimErr == nil:
			successes++
		case errors.Is(claimErr, ErrRepositoryConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent execution claim error: %v", claimErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent execution claims: successes=%d conflicts=%d, want exactly one each", successes, conflicts)
	}
	if exists, err := store.HasExecution(ctx, decisionID); err != nil || !exists {
		t.Fatalf("HasExecution after execution intent = (%t, %v), want true", exists, err)
	}
	if _, err := store.Decision(ctx, "missing-decision"); !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("missing Decision error = %v, want ErrRepositoryNotFound", err)
	}

	expectedLedgerEvents := count + 2
	events, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("load and verify ledger after concurrent appends: %v", err)
	}
	if len(events) != expectedLedgerEvents {
		t.Fatalf("loaded %d events; expected exactly %d in isolated database", len(events), expectedLedgerEvents)
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
	if len(restartedEvents) != expectedLedgerEvents {
		t.Fatalf("restart recovered %d events; expected %d", len(restartedEvents), expectedLedgerEvents)
	}

	// Verify that the chain is checked against the authoritative head, not
	// merely against its own event sequence and predecessor hashes.
	adminURL := os.Getenv("NAEOS_CONTROLPLANE_ADMIN_TEST_DATABASE_URL")
	if adminURL == "" {
		t.Skip("set NAEOS_CONTROLPLANE_ADMIN_TEST_DATABASE_URL to verify ledger-head tampering")
	}
	adminPool, err := pgxpool.New(ctx, adminURL)
	if err != nil {
		t.Fatalf("create isolated admin test pool: %v", err)
	}
	defer adminPool.Close()
	if err := adminPool.Ping(ctx); err != nil {
		t.Fatalf("ping isolated admin test database: %v", err)
	}

	var originalSequence int64
	var originalHash string
	if err := adminPool.QueryRow(ctx, `SELECT last_sequence, event_hash
		FROM naeos_control_plane.ledger_head WHERE singleton = TRUE`).Scan(&originalSequence, &originalHash); err != nil {
		t.Fatalf("read original ledger head for tamper test: %v", err)
	}
	if originalSequence != int64(expectedLedgerEvents) || originalHash != restartedEvents[len(restartedEvents)-1].EventHash {
		t.Fatalf("unexpected original ledger head: sequence=%d hash=%s", originalSequence, originalHash)
	}

	if _, err := adminPool.Exec(ctx, `UPDATE naeos_control_plane.ledger_head
		SET last_sequence = $1 WHERE singleton = TRUE`, originalSequence+1); err != nil {
		t.Fatalf("tamper isolated ledger head sequence: %v", err)
	}
	if _, err := restartedStore.Load(ctx); !errors.Is(err, ErrRepositoryIntegrity) {
		t.Fatalf("Load after ledger-head sequence tampering = %v, want ErrRepositoryIntegrity", err)
	}
	if _, err := adminPool.Exec(ctx, `UPDATE naeos_control_plane.ledger_head
		SET last_sequence = $1, event_hash = $2 WHERE singleton = TRUE`, originalSequence, originalHash); err != nil {
		t.Fatalf("restore isolated ledger head after sequence test: %v", err)
	}

	if _, err := adminPool.Exec(ctx, `UPDATE naeos_control_plane.ledger_head
		SET event_hash = 'tampered' WHERE singleton = TRUE`); err != nil {
		t.Fatalf("tamper isolated ledger head hash: %v", err)
	}
	if _, err := restartedStore.Load(ctx); !errors.Is(err, ErrRepositoryIntegrity) {
		t.Fatalf("Load after ledger-head hash tampering = %v, want ErrRepositoryIntegrity", err)
	}
	if _, err := adminPool.Exec(ctx, `UPDATE naeos_control_plane.ledger_head
		SET last_sequence = $1, event_hash = $2 WHERE singleton = TRUE`, originalSequence, originalHash); err != nil {
		t.Fatalf("restore isolated ledger head after hash test: %v", err)
	}
	if _, err := restartedStore.Load(ctx); err != nil {
		t.Fatalf("Load after restoring authoritative ledger head: %v", err)
	}
}
