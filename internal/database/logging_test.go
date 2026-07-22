package database

import (
	"context"
	"log/slog"
	"testing"
)

func TestLoggingDatabase(t *testing.T) {
	inner := NewPostgreSQL()
	inner.Connect(&Config{})

	logger := slog.Default()
	db := NewLoggingDatabase(inner, logger)

	if db.Name() != "postgresql" {
		t.Errorf("expected name 'postgresql', got %s", db.Name())
	}

	err := db.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (name) VALUES ($1)", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = db.QueryRow("SELECT * FROM test WHERE id = $1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tx.Exec("INSERT INTO test (name) VALUES ($1)", "tx")
	tx.Commit()

	err = db.Migrate([]Migration{{Version: 1, Name: "init", Up: "CREATE TABLE IF NOT EXISTS _m(id INT)"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.Rollback(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.HealthCheck()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoggingDatabaseNilLogger(t *testing.T) {
	inner := NewPostgreSQL()
	inner.Connect(&Config{})

	db := NewLoggingDatabase(inner, nil)
	if db == nil {
		t.Fatal("expected non-nil database")
	}

	err := db.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoggingDatabaseContextMethods(t *testing.T) {
	inner := NewPostgreSQL()
	inner.Connect(&Config{})

	db := NewLoggingDatabase(inner, nil)

	_, err := db.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = db.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = db.QueryRowContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tx, err := db.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tx.Commit()

	err = db.MigrateContext(context.Background(), []Migration{{Version: 1, Name: "init", Up: "CREATE TABLE IF NOT EXISTS _m(id INT)"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.RollbackContext(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSlowQueryLogging(t *testing.T) {
	inner := NewPostgreSQL()
	inner.Connect(&Config{})

	logger := slog.Default()
	db := NewLoggingDatabase(inner, logger)

	_, _ = db.ExecContext(context.Background(), "SELECT 1")
}
