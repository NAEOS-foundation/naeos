//go:build !nosql

package database

import (
	"context"
	"testing"
	"time"
)

func TestRealSupabaseName(t *testing.T) {
	db := NewRealSupabase()
	if db.Name() != "supabase" {
		t.Errorf("expected name 'supabase', got %s", db.Name())
	}
}

func TestRealSupabaseConnectFailure(t *testing.T) {
	db := NewRealSupabase()
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

func TestRealSupabaseNotConnected(t *testing.T) {
	db := NewRealSupabase()

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

func TestRealSupabaseConnectWithDefaultSSL(t *testing.T) {
	db := NewRealSupabase()
	err := db.Connect(&Config{
		Host:     "192.0.2.1",
		Port:     5432,
		User:     "test",
		Password: "test",
		Database: "test",
		Timeout:  1 * time.Second,
	})
	if err == nil {
		t.Error("expected error when connecting to unreachable host")
	}
}

func TestRealSupabaseDefaultContextWithTimeout(t *testing.T) {
	db := NewRealSupabase()
	db.config = &Config{Timeout: 5 * time.Second}
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSupabaseDefaultContextWithoutTimeout(t *testing.T) {
	db := NewRealSupabase()
	db.config = &Config{}
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSupabaseDefaultContextNilConfig(t *testing.T) {
	db := NewRealSupabase()
	ctx, cancel := db.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSupabaseCloseNil(t *testing.T) {
	db := NewRealSupabase()
	if err := db.Close(); err != nil {
		t.Fatalf("Close nil: %v", err)
	}
}
