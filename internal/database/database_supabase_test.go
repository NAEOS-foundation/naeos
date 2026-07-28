package database

import (
	"context"
	"testing"
)

func TestSupabaseNonContextMethods(t *testing.T) {
	t.Parallel()

	db := NewSupabase()
	db.Connect(&Config{Host: "localhost", Port: 5432})

	_, err := db.Exec("SELECT 1")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}

	_, err = db.Query("SELECT 1")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	_, err = db.QueryRow("SELECT 1")
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}

	_, err = db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	if err := db.Migrate([]Migration{{Version: 1, Name: "v1", Up: "CREATE TABLE t(id INT)"}}); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if err := db.Rollback(0); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if err := db.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestSupabaseContextMethods(t *testing.T) {
	t.Parallel()

	db := NewSupabase()
	db.Connect(&Config{})
	ctx := context.Background()

	result, err := db.ExecContext(ctx, "INSERT INTO t VALUES ($1)", 1)
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	rows, err := db.QueryContext(ctx, "SELECT * FROM t")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if rows == nil {
		t.Error("expected non-nil rows")
	}

	row, err := db.QueryRowContext(ctx, "SELECT * FROM t WHERE id = $1", 1)
	if err != nil {
		t.Fatalf("QueryRowContext: %v", err)
	}
	if row == nil {
		t.Error("expected non-nil row")
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO t VALUES ($1)", 2); err != nil {
		t.Fatalf("tx ExecContext: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("tx Commit: %v", err)
	}

	if err := db.MigrateContext(ctx, []Migration{{Version: 1, Name: "v1", Up: "CREATE TABLE x(id INT)"}}); err != nil {
		t.Fatalf("MigrateContext: %v", err)
	}
	if db.MigrationVersion() != 1 {
		t.Errorf("expected version 1, got %d", db.MigrationVersion())
	}

	if err := db.RollbackContext(ctx, 0); err != nil {
		t.Fatalf("RollbackContext: %v", err)
	}
	if db.MigrationVersion() != 0 {
		t.Errorf("expected version 0, got %d", db.MigrationVersion())
	}
}

func TestSupabaseHealthCheck(t *testing.T) {
	t.Parallel()

	db := NewSupabase()
	db.Connect(&Config{})
	if err := db.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestSupabaseNotConnected(t *testing.T) {
	t.Parallel()

	db := NewSupabase()

	_, err := db.Exec("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.ExecContext(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.Query("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.QueryContext(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.QueryRow("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.QueryRowContext(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.Begin()
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.BeginTx(context.Background())
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.Migrate([]Migration{{Version: 1, Name: "v1", Up: "CREATE TABLE t(id INT)"}})
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.MigrateContext(context.Background(), []Migration{{Version: 1, Name: "v1", Up: "CREATE TABLE t(id INT)"}})
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.Rollback(0)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.RollbackContext(context.Background(), 0)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.Ping()
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.HealthCheck()
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestSupabaseMigrationVersion(t *testing.T) {
	t.Parallel()

	db := NewSupabase()
	db.Connect(&Config{})
	if v := db.MigrationVersion(); v != 0 {
		t.Errorf("expected 0, got %d", v)
	}

	db.Migrate([]Migration{{Version: 1, Name: "v1", Up: "CREATE TABLE x(id INT)"}})
	if v := db.MigrationVersion(); v != 1 {
		t.Errorf("expected 1, got %d", v)
	}
}
