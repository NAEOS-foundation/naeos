//go:build !nosql

package database

import (
	"context"
	"path/filepath"
	"testing"
	"time"
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

func TestRealSQLiteNotConnected(t *testing.T) {
	db := NewRealSQLite()

	_, err := db.Exec("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.Query("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.QueryRow("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.Begin()
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

	err = db.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRealSQLiteContextMethods(t *testing.T) {
	db := NewRealSQLite()
	err := db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	_, err = db.ExecContext(ctx, "CREATE TABLE ctx_test (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO ctx_test (val) VALUES (?)", "ctx-value")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	rows, err := db.QueryContext(ctx, "SELECT * FROM ctx_test")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	row, err := db.QueryRowContext(ctx, "SELECT val FROM ctx_test WHERE id = 1")
	if err != nil {
		t.Fatalf("QueryRowContext: %v", err)
	}
	if row["val"] != "ctx-value" {
		t.Errorf("expected 'ctx-value', got %v", row["val"])
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO ctx_test (val) VALUES (?)", "tx-ctx")
	if err != nil {
		t.Fatalf("tx ExecContext: %v", err)
	}

	_, err = tx.QueryContext(ctx, "SELECT * FROM ctx_test")
	if err != nil {
		t.Fatalf("tx QueryContext: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("tx Commit: %v", err)
	}

	migrations := []Migration{{Version: 1, Name: "ctx_mig", Up: "SELECT 1", Down: "SELECT 1"}}
	if err := db.MigrateContext(ctx, migrations); err != nil {
		t.Fatalf("MigrateContext: %v", err)
	}

	if err := db.RollbackContext(ctx, 0); err != nil {
		t.Fatalf("RollbackContext: %v", err)
	}
}

func TestRealSQLiteTransaction(t *testing.T) {
	db := NewRealSQLite()
	err := db.Connect(&Config{Database: ":memory:"})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE tx_test (id INTEGER PRIMARY KEY, val TEXT)")

	tx, err := db.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	_, err = tx.Exec("INSERT INTO tx_test (val) VALUES (?)", "in-tx")
	if err != nil {
		t.Fatalf("tx.Exec: %v", err)
	}

	_, err = tx.Query("SELECT * FROM tx_test")
	if err != nil {
		t.Fatalf("tx.Query: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("tx.Rollback: %v", err)
	}

	rows, _ := db.Query("SELECT * FROM tx_test")
	if len(rows) != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", len(rows))
	}
}

func TestRealSQLiteExecError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	_, err := db.Exec("INVALID SQL !!!")
	if err == nil {
		t.Error("expected error for invalid SQL")
	}
}

func TestRealSQLiteMigrateMultiple(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Name: "create_users", Up: "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)", Down: "DROP TABLE IF EXISTS users"},
		{Version: 2, Name: "add_email", Up: "ALTER TABLE users ADD COLUMN email TEXT DEFAULT ''", Down: ""},
	}

	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("re-Migrate (idempotent): %v", err)
	}

	if err := db.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

func TestRealSQLiteMigrateInvalidSQL(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	err := db.Migrate([]Migration{{Version: 1, Name: "bad", Up: "NOT VALID SQL !!!"}})
	if err == nil {
		t.Fatal("expected error for invalid SQL")
	}
}

func TestRealSQLiteHealthCheck(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	if err := db.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestRealSQLiteName(t *testing.T) {
	db := NewRealSQLite()
	if db.Name() != "sqlite" {
		t.Errorf("expected 'sqlite', got %q", db.Name())
	}
}

func TestRealSQLiteQueryRowNotFound(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	db.Exec("CREATE TABLE empty_tbl (id INTEGER PRIMARY KEY)")

	row, err := db.QueryRow("SELECT * FROM empty_tbl WHERE id = 999")
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if len(row) != 0 {
		t.Errorf("expected empty row for no results, got %d columns", len(row))
	}
}

func TestRealSQLiteTimeout(t *testing.T) {
	db := NewRealSQLite()
	err := db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping with timeout: %v", err)
	}
}

func TestRealSQLiteDefaultEmptyDatabase(t *testing.T) {
	db := NewRealSQLite()
	err := db.Connect(&Config{MaxOpenConns: 1})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestRealSQLiteRollbackEmpty(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	db.Migrate([]Migration{{Version: 1, Name: "tmp", Up: "CREATE TABLE IF NOT EXISTS _tmp(id INT)", Down: "DROP TABLE IF EXISTS _tmp"}})

	if err := db.Rollback(1); err != nil {
		t.Fatalf("Rollback to same version (noop): %v", err)
	}
}

func TestRealSQLiteRollbackWithDownSQL(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Name: "test", Up: "CREATE TABLE IF NOT EXISTS test_rollback (id INTEGER)", Down: "DROP TABLE IF EXISTS test_rollback"},
	}
	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if err := db.Rollback(0); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

func TestRealSQLiteExecQueryRowContext(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	ctx := context.Background()

	_, err := db.ExecContext(ctx, "CREATE TABLE ctx2 (id INTEGER PRIMARY KEY, v INTEGER)")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO ctx2 (v) VALUES (?)", 42)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	rows, err := db.QueryContext(ctx, "SELECT * FROM ctx2")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	_, err = db.QueryRowContext(ctx, "SELECT * FROM ctx2 WHERE id = 999")
	if err != nil {
		t.Fatalf("QueryRowContext: %v", err)
	}
}

func TestRealSQLiteBeginTxContext(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	ctx := context.Background()

	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	_, err = tx.ExecContext(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("tx.ExecContext: %v", err)
	}

	_, err = tx.QueryContext(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("tx.QueryContext: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit: %v", err)
	}
}

func TestRealSQLiteRollbackWithInvalidDownSQL(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:", MaxOpenConns: 1})
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Name: "bad_rollback", Up: "CREATE TABLE IF NOT EXISTS bad_rollback (id INTEGER)", Down: "NOT VALID SQL !!!"},
	}
	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	err := db.Rollback(0)
	if err == nil {
		t.Fatal("expected error for invalid down SQL")
	}
}

func TestRealSQLiteConnectAllConfigOptions(t *testing.T) {
	db := NewRealSQLite()
	err := db.Connect(&Config{
		Database:        ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		Timeout:         5 * time.Second,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestRealSQLiteConnectWithFileDB(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.db")
	db := NewRealSQLite()
	err := db.Connect(&Config{Database: tmpFile})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestRealSQLiteDefaultContextWithTimeout(t *testing.T) {
	db := NewRealSQLite()
	db.config = &Config{Timeout: 1 * time.Second}
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSQLiteDefaultContextWithoutTimeout(t *testing.T) {
	db := NewRealSQLite()
	db.config = &Config{}
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSQLiteDefaultContextNilConfig(t *testing.T) {
	db := NewRealSQLite()
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSQLiteQueryContextError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	_, err := db.QueryContext(context.Background(), "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("expected error for query on nonexistent table")
	}
}

func TestRealSQLiteQueryRowContextError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	_, err := db.QueryRowContext(context.Background(), "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("expected error for query row on nonexistent table")
	}
}

func TestRealSQLiteMigrateContextCheckError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Name: "test", Up: "CREATE TABLE IF NOT EXISTS t1(id INT)"},
	}

	err := db.MigrateContext(context.Background(), migrations)
	if err != nil {
		t.Fatalf("MigrateContext: %v", err)
	}

	migrations = []Migration{
		{Version: 1, Name: "test", Up: "CREATE TABLE IF NOT EXISTS t1(id INT)"},
		{Version: 2, Name: "test2", Up: "CREATE TABLE IF NOT EXISTS t2(id INT)"},
	}

	err = db.MigrateContext(context.Background(), migrations)
	if err != nil {
		t.Fatalf("MigrateContext second run: %v", err)
	}
}

func TestRealSQLiteRollbackEmptyDown(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Name: "no_down", Up: "CREATE TABLE IF NOT EXISTS no_down (id INTEGER)"},
	}
	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if err := db.Rollback(0); err != nil {
		t.Fatalf("Rollback with empty Down: %v", err)
	}
}

func TestRealSQLiteRollbackNoMigrations(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	db.MigrateContext(context.Background(), []Migration{})

	if err := db.Rollback(0); err != nil {
		t.Fatalf("Rollback with no migrations: %v", err)
	}
}

func TestRealSQLiteExecContextError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	_, err := db.ExecContext(context.Background(), "INVALID SQL SYNTAX!!!")
	if err == nil {
		t.Error("expected error for invalid SQL")
	}
}

func TestRealSQLiteQueryRowNotFoundContext(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	db.Exec("CREATE TABLE empty_t (id INTEGER)")

	row, err := db.QueryRowContext(context.Background(), "SELECT * FROM empty_t WHERE id = 999")
	if err != nil {
		t.Fatalf("QueryRowContext: %v", err)
	}
	if len(row) != 0 {
		t.Errorf("expected empty row, got %d columns", len(row))
	}
}

func TestRealSQLiteMigrateContextApplyError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Name: "good", Up: "CREATE TABLE IF NOT EXISTS t1(id INT)"},
	}
	if err := db.MigrateContext(context.Background(), migrations); err != nil {
		t.Fatalf("MigrateContext: %v", err)
	}

	badMigrations := []Migration{
		{Version: 2, Name: "bad", Up: "INVALID SQL!!!"},
	}
	err := db.MigrateContext(context.Background(), badMigrations)
	if err == nil {
		t.Fatal("expected error for bad migration")
	}
}

func TestRealSQLiteMigrateContextRecordError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Name: "good", Up: "CREATE TABLE IF NOT EXISTS t1(id INT)"},
	}
	if err := db.MigrateContext(context.Background(), migrations); err != nil {
		t.Fatalf("MigrateContext: %v", err)
	}

	// Duplicate migration should be skipped
	err := db.MigrateContext(context.Background(), migrations)
	if err != nil {
		t.Fatalf("MigrateContext duplicate: %v", err)
	}
}

func TestRealSQLiteTxExecError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	db.Exec("CREATE TABLE tx_err (id INTEGER)")
	tx, _ := db.BeginTx(context.Background())
	defer tx.Rollback()

	_, err := tx.Exec("INVALID SQL!!!")
	if err == nil {
		t.Error("expected error for invalid SQL in tx")
	}
}

func TestRealSQLiteTxQueryError(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	tx, _ := db.BeginTx(context.Background())
	defer tx.Rollback()

	_, err := tx.Query("SELECT * FROM nonexistent")
	if err == nil {
		t.Error("expected error for query in tx on nonexistent table")
	}
}

func TestRealSQLiteTxQueryRowContext(t *testing.T) {
	db := NewRealSQLite()
	db.Connect(&Config{Database: ":memory:"})
	defer db.Close()

	db.Exec("CREATE TABLE tx_qr (id INTEGER PRIMARY KEY, v TEXT)")
	db.Exec("INSERT INTO tx_qr (v) VALUES (?)", "hello")

	tx, _ := db.BeginTx(context.Background())
	defer tx.Rollback()

	_, err := tx.QueryContext(context.Background(), "SELECT * FROM tx_qr")
	if err != nil {
		t.Fatalf("tx.QueryContext: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit: %v", err)
	}
}
