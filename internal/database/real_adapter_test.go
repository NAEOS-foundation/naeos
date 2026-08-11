//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func openSQLite(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRealMySQLTxAllMethods(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	sqlTx, err := rawDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	tx := &RealMySQLTx{tx: sqlTx}

	result, err := tx.Exec("INSERT INTO t (val) VALUES (?)", "hello")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	result, err = tx.ExecContext(context.Background(), "INSERT INTO t (val) VALUES (?)", "world")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	rows, err := tx.Query("SELECT * FROM t")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	rows, err = tx.QueryContext(context.Background(), "SELECT * FROM t")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func TestRealMySQLTxRollback(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	sqlTx, err := rawDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	tx := &RealMySQLTx{tx: sqlTx}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

func TestRealPostgreSQLTxAllMethods(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	sqlTx, err := rawDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	tx := &RealPostgreSQLTx{tx: sqlTx}

	result, err := tx.Exec("INSERT INTO t (val) VALUES (?)", "hello")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	result, err = tx.ExecContext(context.Background(), "INSERT INTO t (val) VALUES (?)", "world")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	rows, err := tx.Query("SELECT * FROM t")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	rows, err = tx.QueryContext(context.Background(), "SELECT * FROM t")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func TestRealPostgreSQLTxRollback(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	sqlTx, err := rawDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	tx := &RealPostgreSQLTx{tx: sqlTx}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

func TestRealSupabaseTxAllMethods(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	sqlTx, err := rawDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	tx := &RealSupabaseTx{tx: sqlTx}

	result, err := tx.Exec("INSERT INTO t (val) VALUES (?)", "hello")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	result, err = tx.ExecContext(context.Background(), "INSERT INTO t (val) VALUES (?)", "world")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	rows, err := tx.Query("SELECT * FROM t")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	rows, err = tx.QueryContext(context.Background(), "SELECT * FROM t")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func TestRealSupabaseTxRollback(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	sqlTx, err := rawDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	tx := &RealSupabaseTx{tx: sqlTx}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

func TestRealMySQLConnectedMethods(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	m := NewRealMySQL()
	m.db = rawDB
	m.config = &Config{Timeout: time.Second}

	ctx := context.Background()

	result, err := m.ExecContext(ctx, "INSERT INTO test (val) VALUES (?)", "hello")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}
	if result.LastInsertID != 1 {
		t.Errorf("expected last insert id 1, got %d", result.LastInsertID)
	}

	rows, err := m.QueryContext(ctx, "SELECT * FROM test")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	row, err := m.QueryRowContext(ctx, "SELECT * FROM test WHERE id = 1")
	if err != nil {
		t.Fatalf("QueryRowContext: %v", err)
	}
	if row["val"] != "hello" {
		t.Errorf("expected val 'hello', got %v", row["val"])
	}

	row, err = m.QueryRowContext(ctx, "SELECT * FROM test WHERE id = 999")
	if err != nil {
		t.Fatalf("QueryRowContext not found: %v", err)
	}
	if len(row) != 0 {
		t.Errorf("expected empty row, got %d columns", len(row))
	}

	tx, err := m.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("tx.Rollback: %v", err)
	}

	if err := m.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	if err := m.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}

	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealMySQLMigrateAndRollback(t *testing.T) {
	rawDB := openSQLite(t)
	m := NewRealMySQL()
	m.db = rawDB
	m.config = &Config{Timeout: 5 * time.Second}
	ctx := context.Background()

	migrations := []Migration{
		{Version: 1, Name: "create_table", Up: "CREATE TABLE IF NOT EXISTS mig_test (id INTEGER)", Down: "DROP TABLE IF EXISTS mig_test"},
	}
	if err := m.MigrateContext(ctx, migrations); err != nil {
		t.Fatalf("MigrateContext: %v", err)
	}

	if err := m.MigrateContext(ctx, migrations); err != nil {
		t.Fatalf("MigrateContext duplicate (idempotent): %v", err)
	}

	if err := m.RollbackContext(ctx, 0); err != nil {
		t.Fatalf("RollbackContext: %v", err)
	}
}

func TestRealPostgreSQLConnectedMethods(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	p := NewRealPostgreSQL()
	p.db = rawDB
	p.config = &Config{Timeout: time.Second}

	ctx := context.Background()

	result, err := p.ExecContext(ctx, "INSERT INTO test (val) VALUES (?)", "hello")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1, got %d", result.RowsAffected)
	}

	rows, err := p.QueryContext(ctx, "SELECT * FROM test")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	row, err := p.QueryRowContext(ctx, "SELECT * FROM test WHERE id = 1")
	if err != nil {
		t.Fatalf("QueryRowContext: %v", err)
	}
	if row["val"] != "hello" {
		t.Errorf("expected val 'hello', got %v", row["val"])
	}

	row, err = p.QueryRowContext(ctx, "SELECT * FROM test WHERE id = 999")
	if err != nil {
		t.Fatalf("QueryRowContext not found: %v", err)
	}
	if len(row) != 0 {
		t.Errorf("expected empty row, got %d columns", len(row))
	}

	tx, err := p.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("tx.Rollback: %v", err)
	}

	if err := p.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	if err := p.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}

	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealMySQLQueryRowNotFoundNoRows(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE empty_t (id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}

	m := NewRealMySQL()
	m.db = rawDB
	row, err := m.QueryRow("SELECT * FROM empty_t WHERE id = 999")
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if len(row) != 0 {
		t.Errorf("expected empty row, got %d columns", len(row))
	}
}

func TestRealPostgreSQLQueryRowNotFoundNoRows(t *testing.T) {
	rawDB := openSQLite(t)
	_, err := rawDB.Exec("CREATE TABLE empty_t (id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}

	p := NewRealPostgreSQL()
	p.db = rawDB
	row, err := p.QueryRow("SELECT * FROM empty_t WHERE id = 999")
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if len(row) != 0 {
		t.Errorf("expected empty row, got %d columns", len(row))
	}
}
