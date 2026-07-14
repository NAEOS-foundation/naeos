//go:build !nosql

package database

import (
	"context"
	"testing"
	"time"
)

func TestRealPostgreSQLName(t *testing.T) {
	db := NewRealPostgreSQL()
	if db.Name() != "postgresql" {
		t.Errorf("expected name 'postgresql', got %s", db.Name())
	}
}

func TestRealPostgreSQLConnectFailure(t *testing.T) {
	db := NewRealPostgreSQL()
	err := db.Connect(&Config{
		Host:     "192.0.2.1",
		Port:     1,
		User:     "test",
		Password: "test",
		Database: "test",
		SSLMode:  "disable",
		Timeout:  1 * time.Second,
	})
	if err == nil {
		t.Error("expected error when connecting to unreachable host")
	}
}

func TestRealPostgreSQLNotConnected(t *testing.T) {
	db := NewRealPostgreSQL()

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
