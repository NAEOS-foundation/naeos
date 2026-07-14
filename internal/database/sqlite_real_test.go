//go:build !nosql

package database

import (
	"testing"
)

func TestRealSQLiteInMemory(t *testing.T) {
	db := NewRealSQLite()
	err := db.Connect(&Config{
		Database:     ":memory:",
		MaxOpenConns: 1,
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}

	result, err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	t.Logf("rows affected: %d", result.RowsAffected)

	_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "hello")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	row, err := db.QueryRow("SELECT name FROM test WHERE id = 1")
	if err != nil {
		t.Fatalf("queryrow: %v", err)
	}
	if row == nil {
		t.Error("expected row")
	}

	if err := db.Migrate([]Migration{
		{Version: 1, Name: "init", Up: "CREATE TABLE IF NOT EXISTS _m(id INT)", Down: "DROP TABLE IF EXISTS _m"},
	}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if err := db.Rollback(0); err != nil {
		t.Fatalf("rollback: %v", err)
	}
}

func TestRealSQLiteWALMode(t *testing.T) {
	db := NewRealSQLite()
	err := db.Connect(&Config{Database: ":memory:"})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("PRAGMA journal_mode")
	if err != nil {
		t.Fatalf("query pragma: %v", err)
	}
	if len(rows) > 0 {
		mode, ok := rows[0]["journal_mode"]
		if ok {
			t.Logf("journal_mode: %v", mode)
		}
	}
}
