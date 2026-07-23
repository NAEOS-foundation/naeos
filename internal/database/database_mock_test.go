//go:build !nosql

package database

import (
	"context"
	"testing"
)

func TestPostgreSQLMigrationVersion(t *testing.T) {
	t.Parallel()
	db := NewPostgreSQL()
	db.Connect(&Config{})
	if v := db.MigrationVersion(); v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
	db.Migrate([]Migration{{Version: 1, Name: "v1", Up: "CREATE TABLE x(id INT)"}})
	if v := db.MigrationVersion(); v != 1 {
		t.Errorf("expected 1, got %d", v)
	}
}

func TestPostgreSQLNonContextMethods(t *testing.T) {
	t.Parallel()
	db := NewPostgreSQL()
	db.Connect(&Config{})

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

func TestMySQLNonContextMethods(t *testing.T) {
	t.Parallel()
	db := NewMySQL()
	db.Connect(&Config{})

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

func TestSQLiteNonContextMethods(t *testing.T) {
	t.Parallel()
	db := NewSQLite()
	db.Connect(&Config{Database: ":memory:"})

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

func TestFactoryNewAllDrivers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		driver string
		want   string
	}{
		{"postgresql", "postgresql"},
		{"postgres", "postgresql"},
		{"mock-postgresql", "postgresql"},
		{"mock-mysql", "mysql"},
		{"mock-sqlite", "sqlite"},
		{"unknown-driver", ""},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			t.Parallel()
			db := New(tt.driver)
			if tt.want == "" {
				if db != nil {
					t.Errorf("expected nil for unknown driver")
				}
			} else {
				if db == nil {
					t.Fatalf("expected non-nil for driver %s", tt.driver)
				}
				if db.Name() != tt.want {
					t.Errorf("expected name %q, got %q", tt.want, db.Name())
				}
			}
		})
	}
}

func TestPostgreSQLNotConnected(t *testing.T) {
	t.Parallel()
	db := NewPostgreSQL()

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

func TestMySQLNotConnected(t *testing.T) {
	t.Parallel()
	db := NewMySQL()

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

func TestSQLiteNotConnected(t *testing.T) {
	t.Parallel()
	db := NewSQLite()

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
