package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMySQLContextMethods(t *testing.T) {
	db := NewMySQL()
	db.Connect(&Config{})
	ctx := context.Background()

	result, err := db.ExecContext(ctx, "INSERT INTO t VALUES (?)", 1)
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

	row, err := db.QueryRowContext(ctx, "SELECT * FROM t WHERE id = ?", 1)
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
	if _, err := tx.ExecContext(ctx, "INSERT INTO t VALUES (?)", 2); err != nil {
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

func TestSQLiteContextMethods(t *testing.T) {
	db := NewSQLite()
	db.Connect(&Config{Database: ":memory:"})
	ctx := context.Background()

	result, err := db.ExecContext(ctx, "INSERT INTO t VALUES (?)", 1)
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

	row, err := db.QueryRowContext(ctx, "SELECT * FROM t WHERE id = ?", 1)
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
	if _, err := tx.ExecContext(ctx, "INSERT INTO t VALUES (?)", 2); err != nil {
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

func TestMySQLHealthCheck(t *testing.T) {
	db := NewMySQL()
	db.Connect(&Config{})
	if err := db.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestSQLiteHealthCheck(t *testing.T) {
	db := NewSQLite()
	db.Connect(&Config{Database: ":memory:"})
	if err := db.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestNewFromConfigMockPostgreSQL(t *testing.T) {
	db, err := NewFromConfig("mock-postgresql", &Config{Host: "localhost", Port: 5432, User: "u", Database: "d"})
	if err != nil {
		t.Fatalf("NewFromConfig: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	if db.Name() != "postgresql" {
		t.Errorf("expected postgresql, got %s", db.Name())
	}
	db.Close()
}

func TestNewFromConfigValidationError(t *testing.T) {
	_, err := NewFromConfig("mock-postgresql", &Config{})
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestNewFromConfigUnknownDriver(t *testing.T) {
	_, err := NewFromConfig("unknown", &Config{Host: "localhost", Port: 5432, User: "u", Database: "d"})
	if err == nil {
		t.Error("expected error for unknown driver")
	}
}

func TestLoggingDatabaseConnect(t *testing.T) {
	inner := NewPostgreSQL()
	logged := NewLoggingDatabase(inner, nil)
	if err := logged.Connect(&Config{Host: "localhost", Port: 5432}); err != nil {
		t.Fatalf("Connect: %v", err)
	}
}

func TestConfigValidateConnMaxLifetimeNegative(t *testing.T) {
	c := Config{Host: "localhost", Port: 5432, User: "u", Database: "d", ConnMaxLifetime: -1}
	if err := c.Validate(); err == nil {
		t.Error("expected error for negative ConnMaxLifetime")
	}
}

func TestConfigValidateConnMaxIdleTimeNegative(t *testing.T) {
	c := Config{Host: "localhost", Port: 5432, User: "u", Database: "d", ConnMaxIdleTime: -1}
	if err := c.Validate(); err == nil {
		t.Error("expected error for negative ConnMaxIdleTime")
	}
}

type timeoutErr struct{}

func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }
func (timeoutErr) Error() string   { return "timeout" }

type netError struct{}

func (netError) Timeout() bool   { return false }
func (netError) Temporary() bool { return true }
func (netError) Error() string   { return "network error" }

func TestIsTransientErrorNetError(t *testing.T) {
	if !isTransientError(netError{}) {
		t.Error("expected net.Error to be transient")
	}
}

func TestIsTransientErrorTimeout(t *testing.T) {
	if !isTransientError(timeoutErr{}) {
		t.Error("expected timeout to be transient")
	}
}

func TestIsTransientErrorContextCanceled(t *testing.T) {
	if isTransientError(context.Canceled) {
		t.Error("expected context.Canceled not to be transient")
	}
}

func TestManagerConnectAllError(t *testing.T) {
	m := NewManager()
	pg := NewPostgreSQL()
	pg.Connect(&Config{})
	m.Register("pg", pg)
	err := m.ConnectAll(map[string]*Config{
		"pg": {Host: "localhost", Port: 5432, User: "u", Database: "d"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type brokenDB struct {
	Database
}

func (b brokenDB) Close() error {
	return os.ErrPermission
}

func TestManagerCloseAllWithError(t *testing.T) {
	m := NewManager()
	m.Register("pg", NewPostgreSQL())
	m.Register("broken", brokenDB{})
	if err := m.CloseAll(); err == nil {
		t.Error("expected error from broken db close")
	}
}

func TestLoadMigrationsReadError(t *testing.T) {
	_, err := LoadMigrations("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent dir")
	}
}

func TestLoadMigrationsWithSubdir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "001_init.up.sql"), []byte("CREATE TABLE t(id INT)"), 0o644); err != nil {
		t.Fatal(err)
	}
	migs, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if len(migs) != 0 {
		t.Errorf("expected 0 migrations from subdirectory, got %d", len(migs))
	}
}

func TestBaseTransactionQueryContext(t *testing.T) {
	db := &BaseDatabase{tables: make(map[string][]Row)}
	tx := &BaseTransaction{db: db}
	rows, err := tx.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if rows == nil {
		t.Error("expected non-nil rows")
	}
}

func TestConnectionStoreAddDuplicate(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	if err := s.Add("test", "postgresql", &Config{Host: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Add("test", "postgresql", &Config{Host: "b"}); err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestConnectionStoreLoadUnmarshalError(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.filePath(), []byte("{invalid json}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Add("test", "pg", &Config{}); err == nil {
		t.Error("expected unmarshal error from corrupted file")
	}
}
