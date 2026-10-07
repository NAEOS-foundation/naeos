// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/database"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

// sharedDurableInvocationDB models a database whose state survives creation of
// new DatabaseInvocationStore instances. The mutex represents the database's
// atomic uniqueness constraint for the invocation_id primary key.
type sharedDurableInvocationDB struct {
	database.Database
	mu      sync.Mutex
	claimed map[string]struct{}
}

func newSharedDurableInvocationDB() *sharedDurableInvocationDB {
	return &sharedDurableInvocationDB{claimed: make(map[string]struct{})}
}

func (d *sharedDurableInvocationDB) Name() string { return "sqlite" }

func (d *sharedDurableInvocationDB) Exec(query string, args ...any) (database.Result, error) {
	if len(args) == 0 {
		return database.Result{}, nil
	}
	invocationID, ok := args[0].(string)
	if !ok || invocationID == "" {
		return database.Result{}, fmt.Errorf("invalid invocation id")
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.claimed[invocationID]; exists {
		return database.Result{}, nil
	}
	d.claimed[invocationID] = struct{}{}
	return database.Result{RowsAffected: 1}, nil
}

func TestDatabaseInvocationStoreSurvivesStoreRestart(t *testing.T) {
	db := newSharedDurableInvocationDB()

	firstProcessStore := NewDatabaseInvocationStore(db)
	claimed, err := firstProcessStore.Claim("restart-proof")
	if err != nil {
		t.Fatalf("first process claim failed: %v", err)
	}
	if !claimed {
		t.Fatal("first process should claim a fresh invocation")
	}

	// A new store represents a restarted process. The durable database remains
	// the source of truth, so the consumed invocation must stay unavailable.
	restartedProcessStore := NewDatabaseInvocationStore(db)
	claimed, err = restartedProcessStore.Claim("restart-proof")
	if err != nil {
		t.Fatalf("restarted process claim failed: %v", err)
	}
	if claimed {
		t.Fatal("restarted process must not reclaim a consumed invocation")
	}
}

func TestDatabaseInvocationStoreSharedAcrossGatewayInstances(t *testing.T) {
	db := newSharedDurableInvocationDB()
	storeA := NewDatabaseInvocationStore(db)
	storeB := NewDatabaseInvocationStore(db)

	cp := &stubControlPlane{decision: control.DecisionAllow, policyID: "p1"}
	sb := &countingSandbox{}

	gatewayA := New(cp, sb, WithReplayProtection(true), WithInvocationStore(storeA))
	gatewayB := New(cp, sb, WithReplayProtection(true), WithInvocationStore(storeB))

	const attempts = 32
	var wg sync.WaitGroup
	wg.Add(attempts)
	results := make(chan ExecutionResult, attempts)
	errors := make(chan error, attempts)

	for i := 0; i < attempts; i++ {
		go func() {
			defer wg.Done()
			result, err := gatewayA.Authorize(ToolRequest{
				InvocationID: "shared-replica-invocation",
				Tool:         "shell",
				Action:       "run",
			})
			if i%2 == 1 {
				result, err = gatewayB.Authorize(ToolRequest{
					InvocationID: "shared-replica-invocation",
					Tool:         "shell",
					Action:       "run",
				})
			}
			results <- result
			errors <- err
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	var completed int
	for err := range errors {
		if err != nil {
			t.Fatalf("unexpected gateway error: %v", err)
		}
	}
	for result := range results {
		if result.Status == "completed" {
			completed++
		}
	}
	if completed != 1 {
		t.Fatalf("expected exactly one completed execution, got %d", completed)
	}
	if got := sb.Count(); got != 1 {
		t.Fatalf("expected exactly one sandbox side effect, got %d", got)
	}
}

type countingSandbox struct {
	mu    sync.Mutex
	count int
}

func (s *countingSandbox) Execute(ToolRequest) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
	return "ok", nil
}

func (s *countingSandbox) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

func TestDatabaseInvocationStoreExpiryDoesNotReopenReplayWindow(t *testing.T) {
	db := newSharedDurableInvocationDB()
	store := NewDatabaseInvocationStore(db)

	claimed, err := store.Claim("retained-invocation")
	if err != nil {
		t.Fatalf("initial claim failed: %v", err)
	}
	if !claimed {
		t.Fatal("expected initial claim to succeed")
	}

	time.Sleep(time.Millisecond)

	claimed, err = store.Claim("retained-invocation")
	if err != nil {
		t.Fatalf("repeat claim failed: %v", err)
	}
	if claimed {
		t.Fatal("consumed invocation must remain rejected; retention must not silently reopen replay")
	}
}

var _ ControlPlane = (*stubControlPlane)(nil)
var _ = control.DecisionAllow
