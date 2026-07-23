//go:build !nosql

package database

import (
	"context"
	"testing"
	"time"
)

func TestRealMySQLName(t *testing.T) {
	db := NewRealMySQL()
	if db.Name() != "mysql" {
		t.Errorf("expected name 'mysql', got %s", db.Name())
	}
}

func TestRealMySQLConnectFailure(t *testing.T) {
	db := NewRealMySQL()
	err := db.Connect(&Config{
		Host:     "192.0.2.1",
		Port:     1,
		User:     "test",
		Password: "test",
		Database: "test",
		Timeout:  1 * time.Second,
	})
	if err == nil {
		t.Error("expected error when connecting to unreachable host")
	}
}

func TestRealMySQLNotConnected(t *testing.T) {
	db := NewRealMySQL()

	err := db.Ping()
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.Exec("SELECT 1")
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

	err = db.Migrate(nil)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.MigrateContext(context.Background(), nil)
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

	err = db.HealthCheck()
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestRealMySQLHealthCheck(t *testing.T) {
	db := NewRealMySQL()
	if err := db.HealthCheck(); err == nil {
		t.Error("expected error when not connected")
	}
}

func TestRealMySQLConnectAllConfigOptions(t *testing.T) {
	db := NewRealMySQL()
	err := db.Connect(&Config{
		Host:            "192.0.2.1",
		Port:            1,
		User:            "test",
		Password:        "test",
		Database:        "test",
		Timeout:         1 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	if err == nil {
		t.Error("expected error when connecting to unreachable host")
	}
}

func TestRealMySQLConnectNoOptionalConfig(t *testing.T) {
	db := NewRealMySQL()
	err := db.Connect(&Config{
		Host:     "192.0.2.1",
		Port:     1,
		User:     "test",
		Password: "test",
		Database: "test",
	})
	if err == nil {
		t.Error("expected error when connecting to unreachable host")
	}
}

func TestRealMySQLDefaultContextWithTimeout(t *testing.T) {
	db := NewRealMySQL()
	db.config = &Config{Timeout: 5 * time.Second}
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealMySQLDefaultContextWithoutTimeout(t *testing.T) {
	db := NewRealMySQL()
	db.config = &Config{}
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealMySQLDefaultContextNilConfig(t *testing.T) {
	db := NewRealMySQL()
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealMySQLCloseNil(t *testing.T) {
	db := NewRealMySQL()
	if err := db.Close(); err != nil {
		t.Fatalf("Close nil: %v", err)
	}
}
