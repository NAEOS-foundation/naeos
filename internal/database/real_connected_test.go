//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	_ "modernc.org/sqlite"
)

func openSQLiteBacked(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)"); err != nil {
		t.Fatalf("create users: %v", err)
	}
	return db
}

func TestRealPostgreSQLConnectedExecQuery(t *testing.T) {
	db := NewRealPostgreSQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	res, err := db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "alice", "alice@example.com")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}

	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "alice" {
		t.Errorf("expected name alice, got %v", rows[0]["name"])
	}

	row, err := db.QueryRow("SELECT email FROM users WHERE id = $1", 1)
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if row["email"] != "alice@example.com" {
		t.Errorf("expected email, got %v", row["email"])
	}

	empty, err := db.QueryRow("SELECT email FROM users WHERE id = $1", 999)
	if err != nil {
		t.Fatalf("QueryRow empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty row, got %v", empty)
	}
}

func TestRealPostgreSQLConnectedErrors(t *testing.T) {
	db := NewRealPostgreSQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	if _, err := db.Exec("INSERT INTO nope VALUES (1)"); err == nil {
		t.Error("expected Exec error for missing table")
	}
	if _, err := db.Query("SELECT * FROM nope"); err == nil {
		t.Error("expected Query error for missing table")
	}
	if _, err := db.QueryRow("SELECT * FROM nope"); err == nil {
		t.Error("expected QueryRow error for missing table")
	}
	if err := db.Ping(); err != nil {
		t.Errorf("Ping: %v", err)
	}
	if err := db.HealthCheck(); err != nil {
		t.Errorf("HealthCheck: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestRealPostgreSQLConnectedBeginTx(t *testing.T) {
	db := NewRealPostgreSQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	res, err := tx.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "bob", "bob@example.com")
	if err != nil {
		t.Fatalf("tx Exec: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}
	rows, err := tx.Query("SELECT name FROM users WHERE name = $1", "bob")
	if err != nil {
		t.Fatalf("tx Query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	tx2, err := db.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := tx2.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	tx3, err := db.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if _, err := tx3.Exec("INSERT INTO nope VALUES (1)"); err == nil {
		t.Error("expected tx Exec error")
	}
	_ = tx3.Rollback()
}

func TestRealPostgreSQLMigrateContextErrorPath(t *testing.T) {
	db := NewRealPostgreSQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	err := db.MigrateContext(context.Background(), []Migration{{Version: 1, Name: "m1", Up: "SELECT 1"}})
	if err == nil {
		t.Fatal("expected error: Postgres DDL not supported by SQLite backend")
	}
	if err := db.Migrate([]Migration{{Version: 1, Name: "m1"}}); err == nil {
		t.Error("expected Migrate error")
	}
}

func TestRealPostgreSQLRollbackContextErrorPath(t *testing.T) {
	db := NewRealPostgreSQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	err := db.RollbackContext(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error: _migrations table missing")
	}
	if err := db.Rollback(0); err == nil {
		t.Error("expected Rollback error")
	}
}

func TestRealPostgreSQLMigrateContextSuccess(t *testing.T) {
	db := NewRealPostgreSQL()
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db.db = raw
	db.config = &Config{}
	defer raw.Close()

	migrations := []Migration{
		{Version: 1, Name: "create_items", Up: "CREATE TABLE items (id INTEGER PRIMARY KEY)", Down: "DROP TABLE items"},
		{Version: 2, Name: "add_qty", Up: "ALTER TABLE items ADD COLUMN qty INTEGER", Down: "ALTER TABLE items DROP COLUMN qty"},
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS _migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE items").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO _migrations").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT COUNT").WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec("ALTER TABLE items").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO _migrations").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations (pass 1): %v", err)
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS _migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT COUNT").WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate (idempotent): %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations (idempotent): %v", err)
	}
}

func TestRealPostgreSQLMigrateContextFailures(t *testing.T) {
	db := NewRealPostgreSQL()
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db.db = raw
	db.config = &Config{}
	defer raw.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS _migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO nope").WillReturnError(errors.New("syntax error"))

	err = db.Migrate([]Migration{{Version: 1, Name: "bad", Up: "INSERT INTO nope VALUES (1)"}})
	if err == nil {
		t.Fatal("expected error for failing migration")
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS _migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO _migrations").WillReturnError(errors.New("constraint failed"))

	err = db.Migrate([]Migration{{Version: 1, Name: "dup", Up: "SELECT 1"}})
	if err == nil {
		t.Fatal("expected error when recording migration fails")
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS _migrations").WillReturnError(errors.New("permission denied"))
	if err := db.Migrate([]Migration{{Version: 1, Name: "m", Up: "SELECT 1"}}); err == nil {
		t.Error("expected error when creating table fails")
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS _migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT").WithArgs(1).WillReturnError(errors.New("table locked"))
	if err := db.Migrate([]Migration{{Version: 1, Name: "m", Up: "SELECT 1"}}); err == nil {
		t.Error("expected error when checking migration fails")
	}
}

func TestRealPostgreSQLRollbackContextSuccess(t *testing.T) {
	db := NewRealPostgreSQL()
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db.db = raw
	db.config = &Config{}
	defer raw.Close()

	rows := sqlmock.NewRows([]string{"version", "name", "down_sql"})
	rows.AddRow(2, "v2", "ALTER TABLE items DROP COLUMN qty")
	rows.AddRow(1, "v1", "DROP TABLE items")

	mock.ExpectQuery("SELECT version, name, down_sql").WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec("ALTER TABLE items").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM _migrations").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("DROP TABLE items").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM _migrations").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := db.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestRealPostgreSQLRollbackContextFailures(t *testing.T) {
	db := NewRealPostgreSQL()
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db.db = raw
	db.config = &Config{}
	defer raw.Close()

	mock.ExpectQuery("SELECT version, name, down_sql").WillReturnError(errors.New("no such table"))
	if err := db.RollbackContext(context.Background(), 0); err == nil {
		t.Error("expected error when querying migrations fails")
	}

	rows := sqlmock.NewRows([]string{"version", "name", "down_sql"})
	rows.AddRow(2, "v2", "DROP TABLE items")
	mock.ExpectQuery("SELECT version, name, down_sql").WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec("DROP TABLE items").WillReturnError(errors.New("syntax error"))
	if err := db.RollbackContext(context.Background(), 0); err == nil {
		t.Error("expected error when down migration fails")
	}

	rows2 := sqlmock.NewRows([]string{"version", "name", "down_sql"})
	rows2.AddRow(2, "v2", "")
	mock.ExpectQuery("SELECT version, name, down_sql").WillReturnRows(rows2)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM _migrations").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := db.RollbackContext(context.Background(), 0); err != nil {
		t.Errorf("rollback without down sql: %v", err)
	}
}

func TestRealPostgreSQLConnectedPingHealth(t *testing.T) {
	db := NewRealPostgreSQL()
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db.db = raw
	db.config = &Config{}
	defer raw.Close()

	mock.ExpectPing()
	mock.ExpectPing()
	if err := db.Ping(); err != nil {
		t.Errorf("Ping: %v", err)
	}
	if err := db.HealthCheck(); err != nil {
		t.Errorf("HealthCheck: %v", err)
	}
}

func TestRealMySQLConnectedExecQuery(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	res, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "carol", "carol@example.com")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}

	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "carol" {
		t.Errorf("expected name carol, got %v", rows[0]["name"])
	}

	row, err := db.QueryRow("SELECT email FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if row["email"] != "carol@example.com" {
		t.Errorf("expected email, got %v", row["email"])
	}

	empty, err := db.QueryRow("SELECT email FROM users WHERE id = ?", 999)
	if err != nil {
		t.Fatalf("QueryRow empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty row, got %v", empty)
	}
}

func TestRealMySQLConnectedErrors(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	if _, err := db.Exec("INSERT INTO nope VALUES (1)"); err == nil {
		t.Error("expected Exec error")
	}
	if _, err := db.Query("SELECT * FROM nope"); err == nil {
		t.Error("expected Query error")
	}
	if _, err := db.QueryRow("SELECT * FROM nope"); err == nil {
		t.Error("expected QueryRow error")
	}
	if err := db.Ping(); err != nil {
		t.Errorf("Ping: %v", err)
	}
	if err := db.HealthCheck(); err != nil {
		t.Errorf("HealthCheck: %v", err)
	}
}

func TestRealMySQLConnectedBeginTx(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	res, err := tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "dave", "dave@example.com")
	if err != nil {
		t.Fatalf("tx Exec: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}
	rows, err := tx.Query("SELECT name FROM users WHERE name = ?", "dave")
	if err != nil {
		t.Fatalf("tx Query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	tx2, err := db.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := tx2.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	tx3, err := db.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if _, err := tx3.Exec("INSERT INTO nope VALUES (1)"); err == nil {
		t.Error("expected tx Exec error")
	}
	if _, err := tx3.Query("SELECT * FROM nope"); err == nil {
		t.Error("expected tx Query error")
	}
	_ = tx3.Rollback()
}

func TestRealMySQLMigrateContextSuccess(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	migrations := []Migration{
		{Version: 1, Name: "create_users", Up: "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)", Down: "DROP TABLE items"},
		{Version: 2, Name: "add_qty", Up: "ALTER TABLE items ADD COLUMN qty INTEGER DEFAULT 0", Down: "ALTER TABLE items DROP COLUMN qty"},
	}
	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate (idempotent): %v", err)
	}

	var count int
	if err := db.db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 recorded migrations, got %d", count)
	}
}

func TestRealMySQLMigrateContextFailures(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	err := db.Migrate([]Migration{{Version: 1, Name: "bad", Up: "INSERT INTO nope VALUES (1)"}})
	if err == nil {
		t.Error("expected error for failing migration")
	}

	rows, err := db.db.Query("SELECT COUNT(*) FROM _migrations")
	if err != nil {
		t.Fatalf("query _migrations: %v", err)
	}
	rows.Close()

	if err := db.Migrate([]Migration{{Version: 1, Name: "dup", Up: "SELECT 1", Down: "bad down"}}); err != nil {
		t.Fatalf("Migrate setup: %v", err)
	}
	err = db.Migrate([]Migration{{Version: 1, Name: "dup", Up: "SELECT 1"}})
	if err != nil {
		t.Fatalf("Migrate duplicate should be skipped: %v", err)
	}

	err = db.Migrate([]Migration{{Version: 2, Name: "record-fail", Up: "INSERT INTO users (name, email) VALUES ('x', 'y')"}})
	if err != nil {
		t.Fatalf("Migrate record-fail setup: %v", err)
	}
}

func TestRealMySQLRollbackContextSuccess(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	migrations := []Migration{
		{Version: 1, Name: "v1", Up: "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)", Down: "DROP TABLE items"},
		{Version: 2, Name: "v2", Up: "ALTER TABLE items ADD COLUMN qty INTEGER DEFAULT 0", Down: "ALTER TABLE items DROP COLUMN qty"},
	}
	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if err := db.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	var count int
	if err := db.db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 migration remaining, got %d", count)
	}
}

func TestRealMySQLRollbackContextFailures(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	db.config = &Config{}

	err := db.Rollback(0)
	if err == nil {
		t.Fatal("expected error: no _migrations table")
	}

	db.db.Exec("CREATE TABLE _migrations (version INT PRIMARY KEY, name VARCHAR(255), down_sql TEXT)")
	_, err = db.db.Exec("INSERT INTO _migrations (version, name, down_sql) VALUES (1, 'bad', 'INSERT INTO nope VALUES (1)')")
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	err = db.Rollback(0)
	if err == nil {
		t.Error("expected error when down migration fails")
	}

	_, err = db.db.Exec("INSERT INTO _migrations (version, name, down_sql) VALUES (2, 'ok', 'SELECT 1')")
	if err != nil {
		t.Fatalf("seed 2: %v", err)
	}
	if err := db.Rollback(1); err != nil {
		t.Fatalf("Rollback 2 should succeed: %v", err)
	}
}

func TestRealMySQLCloseConnected(t *testing.T) {
	db := NewRealMySQL()
	db.db = openSQLiteBacked(t)
	if err := db.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestRealPostgreSQLCloseConnected(t *testing.T) {
	db := NewRealPostgreSQL()
	db.db = openSQLiteBacked(t)
	if err := db.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
