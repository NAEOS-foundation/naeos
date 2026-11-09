// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/database"
	_ "modernc.org/sqlite"
)

// sqliteInvocationDB uses a real file-backed SQLite database while embedding
// the unused Database interface methods for the narrow replay-store contract.
type sqliteInvocationDB struct {
	database.Database
	db *sql.DB
}

func (d *sqliteInvocationDB) Name() string { return "sqlite" }

func (d *sqliteInvocationDB) Exec(query string, args ...any) (database.Result, error) {
	result, err := d.db.Exec(query, args...)
	if err != nil {
		return database.Result{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return database.Result{}, err
	}
	return database.Result{RowsAffected: affected}, nil
}

func openSQLiteInvocationDB(t *testing.T, path string) *sqliteInvocationDB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open SQLite replay database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite replay database: %v", err)
		}
	})
	return &sqliteInvocationDB{db: db}
}

func TestDatabaseInvocationStorePersistsAcrossSQLiteReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replay.db")
	firstDB := openSQLiteInvocationDB(t, path)

	claimed, err := NewDatabaseInvocationStore(firstDB).Claim("persisted-invocation")
	if err != nil {
		t.Fatalf("first claim failed: %v", err)
	}
	if !claimed {
		t.Fatal("expected initial claim to succeed")
	}

	if err := firstDB.db.Close(); err != nil {
		t.Fatalf("close first SQLite connection: %v", err)
	}
	reopenedDB := openSQLiteInvocationDB(t, path)

	claimed, err = NewDatabaseInvocationStore(reopenedDB).Claim("persisted-invocation")
	if err != nil {
		t.Fatalf("claim after SQLite reopen failed: %v", err)
	}
	if claimed {
		t.Fatal("consumed invocation was claimable after reopening the database")
	}
}

func TestDatabaseInvocationStoreAtomicClaimAcrossSQLiteConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "concurrent-replay.db")
	dbA := openSQLiteInvocationDB(t, path)
	dbB := openSQLiteInvocationDB(t, path)
	storeA := NewDatabaseInvocationStore(dbA)
	storeB := NewDatabaseInvocationStore(dbB)

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
			claimed, err := s.Claim("single-atomic-sqlite-claim")
			claims <- claimed
			errors <- err
		}(store)
	}

	wg.Wait()
	close(claims)
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("unexpected concurrent claim error: %v", err)
		}
	}
	successes := 0
	for claimed := range claims {
		if claimed {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one successful SQL claim, got %d", successes)
	}
}

type restModeInvocationDB struct {
	database.Database
}

func (d *restModeInvocationDB) Name() string     { return "supabase" }
func (d *restModeInvocationDB) IsRESTMode() bool { return true }
func (d *restModeInvocationDB) Exec(string, ...any) (database.Result, error) {
	return database.Result{}, fmt.Errorf("Exec must not be reached in REST mode")
}

func TestDatabaseInvocationStoreFailsClosedForSupabaseREST(t *testing.T) {
	store := NewDatabaseInvocationStore(&restModeInvocationDB{})
	claimed, err := store.Claim("rest-mode-invocation")
	if err == nil {
		t.Fatal("expected REST mode claim to fail closed")
	}
	if claimed {
		t.Fatal("REST mode claim must not be reported successful")
	}
}

type transientInitFailureDB struct {
	database.Database
	delegate *sharedDurableInvocationDB
	mu       sync.Mutex
	failed   bool
}

func (d *transientInitFailureDB) Name() string { return "sqlite" }

func (d *transientInitFailureDB) Exec(query string, args ...any) (database.Result, error) {
	d.mu.Lock()
	if !d.failed {
		d.failed = true
		d.mu.Unlock()
		return database.Result{}, fmt.Errorf("temporary database unavailable")
	}
	d.mu.Unlock()
	return d.delegate.Exec(query, args...)
}

func TestDatabaseInvocationStoreRetriesTransientTableInitialization(t *testing.T) {
	db := &transientInitFailureDB{delegate: newSharedDurableInvocationDB()}
	store := NewDatabaseInvocationStore(db)

	if claimed, err := store.Claim("retry-init-invocation"); err == nil || claimed {
		t.Fatal("expected first claim to fail on transient initialization error")
	}
	claimed, err := store.Claim("retry-init-invocation")
	if err != nil {
		t.Fatalf("claim after transient initialization failure: %v", err)
	}
	if !claimed {
		t.Fatal("expected claim to succeed after initialization retry")
	}
}
