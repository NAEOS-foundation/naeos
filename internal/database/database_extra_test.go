package database

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigValidationSSLMode(t *testing.T) {
	t.Parallel()

	modes := []struct {
		mode    string
		wantErr bool
	}{
		{"disable", false},
		{"require", false},
		{"verify-ca", false},
		{"verify-full", false},
		{"prefer", true},
		{"allow", true},
	}
	for _, tt := range modes {
		t.Run(tt.mode, func(t *testing.T) {
			t.Parallel()
			c := Config{Host: "h", Port: 1, User: "u", Database: "d", SSLMode: tt.mode}
			err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SSLMode=%s: Validate() error = %v, wantErr %v", tt.mode, err, tt.wantErr)
			}
		})
	}
}

func TestFactoryNewFromConfigMock(t *testing.T) {
	t.Parallel()

	db, err := NewFromConfig("mock-sqlite", &Config{Host: "localhost", Port: 1, User: "u", Database: "d"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.Name() != "sqlite" {
		t.Errorf("expected 'sqlite', got %q", db.Name())
	}
}

func TestConnectionStore(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.Add("mydb", "postgresql", &Config{Host: "localhost", Port: 5432})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := s.Get("mydb")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Driver != "postgresql" {
		t.Errorf("expected driver 'postgresql', got %q", got.Driver)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 connection, got %d", len(list))
	}

	if err := s.Remove("mydb"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	_, err = s.Get("mydb")
	if err == nil {
		t.Fatal("expected error after remove")
	}
}

func TestConnectionStoreRemoveNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent connection")
	}
}

func TestConnectionStoreGetNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	_, err := s.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent connection")
	}
}

func TestLoggingDatabaseAllMethods(t *testing.T) {
	t.Parallel()

	inner := NewSQLite()
	inner.Connect(&Config{Host: "h", Port: 1, User: "u", Database: ":memory:"})

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ldb := NewLoggingDatabase(inner, logger)

	if ldb.Name() != "sqlite" {
		t.Errorf("expected 'sqlite', got %q", ldb.Name())
	}

	if err := ldb.Ping(); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	_, err := ldb.Exec("SELECT 1")
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	_, err = ldb.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("ExecContext() error = %v", err)
	}

	_, err = ldb.Query("SELECT 1")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	_, err = ldb.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("QueryContext() error = %v", err)
	}

	_, err = ldb.QueryRow("SELECT 1")
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}

	_, err = ldb.QueryRowContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("QueryRowContext() error = %v", err)
	}

	_, err = ldb.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	_, err = ldb.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	migrations := []Migration{{Version: 1, Name: "test", Up: "SELECT 1"}}
	if err := ldb.Migrate(migrations); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	if err := ldb.MigrateContext(context.Background(), migrations); err != nil {
		t.Fatalf("MigrateContext() error = %v", err)
	}

	if err := ldb.Rollback(0); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	if err := ldb.RollbackContext(context.Background(), 0); err != nil {
		t.Fatalf("RollbackContext() error = %v", err)
	}

	if err := ldb.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}

	if err := ldb.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestLoggingDatabaseLongQuery(t *testing.T) {
	t.Parallel()

	inner := NewSQLite()
	inner.Connect(&Config{Host: "h", Port: 1, User: "u", Database: "d"})

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ldb := NewLoggingDatabase(inner, logger)

	longQuery := "SELECT " + string(make([]byte, 300))
	_, err := ldb.Query(longQuery)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
}

func TestLoadMigrationsMultipleVersions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "002_add_email.up.sql"), []byte("ALTER TABLE users ADD email TEXT"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "002_add_email.down.sql"), []byte("ALTER TABLE users DROP email"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "001_create_users.up.sql"), []byte("CREATE TABLE users"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "001_create_users.down.sql"), []byte("DROP TABLE users"), 0o644); err != nil {
		t.Fatal(err)
	}

	migrations, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("LoadMigrations() error = %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}
	if migrations[0].Version != 1 {
		t.Errorf("expected first migration version 1, got %d", migrations[0].Version)
	}
	if migrations[1].Version != 2 {
		t.Errorf("expected second migration version 2, got %d", migrations[1].Version)
	}
}

func TestLoadMigrationsUpOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "001_init.up.sql"), []byte("CREATE TABLE t"), 0o644)

	migrations, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("LoadMigrations() error = %v", err)
	}
	if len(migrations) != 1 {
		t.Fatalf("expected 1 migration, got %d", len(migrations))
	}
	if migrations[0].Up != "CREATE TABLE t" {
		t.Errorf("unexpected Up SQL: %q", migrations[0].Up)
	}
	if migrations[0].Down != "" {
		t.Errorf("expected empty Down SQL, got %q", migrations[0].Down)
	}
}

func TestPoolGetEmpty(t *testing.T) {
	t.Parallel()

	pool := NewPool(5, 2, time.Hour)
	got := pool.Get()
	if got != nil {
		t.Error("expected nil from empty pool")
	}
}

func TestPoolPutFull(t *testing.T) {
	t.Parallel()

	pool := NewPool(1, 1, time.Hour)

	db := NewSQLite()
	db.Connect(&Config{Host: "h", Port: 1, User: "u", Database: "d"})

	pool.Put(db)
	if pool.Size() != 1 {
		t.Errorf("expected size 1, got %d", pool.Size())
	}

	pool.Put(db)
	if pool.Size() != 1 {
		t.Errorf("expected size 1 (overflow closed), got %d", pool.Size())
	}
}

func TestManagerConnectAllNotFound(t *testing.T) {
	t.Parallel()

	m := NewManager()
	configs := map[string]*Config{"missing": {Host: "h", Port: 1, User: "u", Database: "d"}}
	if err := m.ConnectAll(configs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerGetNotFound(t *testing.T) {
	t.Parallel()

	m := NewManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent database")
	}
}

func TestManagerRemoveNonExistent(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.Remove("nonexistent")
}

func TestTransactionContext(t *testing.T) {
	t.Parallel()

	db := NewPostgreSQL()
	db.Connect(&Config{})

	tx, err := db.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	_, err = tx.ExecContext(context.Background(), "INSERT INTO t VALUES (1)")
	if err != nil {
		t.Fatalf("ExecContext() error = %v", err)
	}

	_, err = tx.QueryContext(context.Background(), "SELECT * FROM t")
	if err != nil {
		t.Fatalf("QueryContext() error = %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
}

func TestSQLiteAllMethods(t *testing.T) {
	t.Parallel()

	db := NewSQLite()
	db.Connect(&Config{Host: "h", Port: 1, User: "u", Database: ":memory:"})

	ctx := context.Background()

	_, err := db.ExecContext(ctx, "CREATE TABLE t (id INTEGER)")
	if err != nil {
		t.Fatalf("ExecContext() error = %v", err)
	}

	_, err = db.QueryContext(ctx, "SELECT * FROM t")
	if err != nil {
		t.Fatalf("QueryContext() error = %v", err)
	}

	_, err = db.QueryRowContext(ctx, "SELECT * FROM t WHERE id = 1")
	if err != nil {
		t.Fatalf("QueryRowContext() error = %v", err)
	}

	migrations := []Migration{{Version: 1, Name: "m", Up: "SELECT 1", Down: "SELECT 1"}}
	if err := db.MigrateContext(ctx, migrations); err != nil {
		t.Fatalf("MigrateContext() error = %v", err)
	}

	if err := db.RollbackContext(ctx, 0); err != nil {
		t.Fatalf("RollbackContext() error = %v", err)
	}

	if v := db.MigrationVersion(); v != 0 {
		t.Errorf("expected version 0, got %d", v)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestMySQLAllMethods(t *testing.T) {
	t.Parallel()

	db := NewMySQL()
	db.Connect(&Config{Host: "h", Port: 1, User: "u", Database: "d"})

	ctx := context.Background()

	_, err := db.ExecContext(ctx, "INSERT INTO t VALUES (1)")
	if err != nil {
		t.Fatalf("ExecContext() error = %v", err)
	}

	_, err = db.QueryContext(ctx, "SELECT * FROM t")
	if err != nil {
		t.Fatalf("QueryContext() error = %v", err)
	}

	_, err = db.QueryRowContext(ctx, "SELECT * FROM t WHERE id = 1")
	if err != nil {
		t.Fatalf("QueryRowContext() error = %v", err)
	}

	migrations := []Migration{{Version: 1, Name: "m", Up: "SELECT 1", Down: "SELECT 1"}}
	if err := db.MigrateContext(ctx, migrations); err != nil {
		t.Fatalf("MigrateContext() error = %v", err)
	}

	if err := db.RollbackContext(ctx, 0); err != nil {
		t.Fatalf("RollbackContext() error = %v", err)
	}

	if v := db.MigrationVersion(); v != 0 {
		t.Errorf("expected version 0, got %d", v)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestBaseTransactionRollback(t *testing.T) {
	t.Parallel()

	db := NewPostgreSQL()
	db.Connect(&Config{})

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
}

func TestWithRetryDefaultParams(t *testing.T) {
	t.Parallel()

	calls := 0
	err := WithRetry(context.Background(), 0, 0, func(ctx context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetryNonTransient(t *testing.T) {
	t.Parallel()

	err := WithRetry(context.Background(), 3, time.Millisecond, func(ctx context.Context) error {
		return &os.PathError{Op: "open", Path: "/not/found", Err: os.ErrNotExist}
	})
	if err == nil {
		t.Fatal("expected error for non-transient failure")
	}
}

func TestWithRetryContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := WithRetry(ctx, 3, time.Millisecond, func(ctx context.Context) error {
		return context.Canceled
	})
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestIsTransientErrorNil(t *testing.T) {
	t.Parallel()
	if isTransientError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsTransientErrorDeadlineExceeded(t *testing.T) {
	t.Parallel()
	if !isTransientError(context.DeadlineExceeded) {
		t.Error("expected true for DeadlineExceeded")
	}
}
