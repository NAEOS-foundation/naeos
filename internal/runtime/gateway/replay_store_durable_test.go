// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"fmt"
	"sync"
	"testing"

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

func (d *sharedDurableInvocationDB) Exec(_ string, args ...any) (database.Result, error) {
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

func TestDatabaseInvocationStoreConcurrentSharedClaim(t *testing.T) {
	db := newSharedDurableInvocationDB()
	storeA := NewDatabaseInvocationStore(db)
	storeB := NewDatabaseInvocationStore(db)

	const attempts = 32
	var wg sync.WaitGroup
	wg.Add(attempts)

	claims := make(chan bool, attempts)
	errors := make(chan error, attempts)
	for i := 0; i < attempts; i++ {
		store := storeA
		if i%2 == 1 {
			store = storeB
		}
		go func(s InvocationStore) {
			defer wg.Done()
			claimed, err := s.Claim("shared-replica-invocation")
			claims <- claimed
			errors <- err
		}(store)
	}

	wg.Wait()
	close(claims)
	close(errors)

	successes := 0
	for err := range errors {
		if err != nil {
			t.Fatalf("unexpected claim error: %v", err)
		}
	}
	for claimed := range claims {
		if claimed {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one durable claim, got %d", successes)
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

	first, err := gatewayA.Authorize(ToolRequest{
		InvocationID: "gateway-shared-invocation",
		Tool:         "shell",
		Action:       "run",
	})
	if err != nil {
		t.Fatalf("first gateway execution failed: %v", err)
	}
	if first.Status != "completed" {
		t.Fatalf("expected first gateway execution to complete, got %s", first.Status)
	}

	second, err := gatewayB.Authorize(ToolRequest{
		InvocationID: "gateway-shared-invocation",
		Tool:         "shell",
		Action:       "run",
	})
	if err != nil {
		t.Fatalf("second gateway execution returned unexpected error: %v", err)
	}
	if second.Status != "denied" {
		t.Fatalf("expected second gateway execution to be denied, got %s", second.Status)
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
