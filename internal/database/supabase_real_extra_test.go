//go:build !nosql

package database

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRealSupabaseRESTContextMethods(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]map[string]any{{"n": float64(1)}})
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "svc",
		SupabaseAccessToken:    "tok",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if _, err := s.ExecContext(context.Background(), "SELECT 1"); err != nil {
		t.Errorf("ExecContext: %v", err)
	}
	if _, err := s.QueryContext(context.Background(), "SELECT 1"); err != nil {
		t.Errorf("QueryContext: %v", err)
	}
	if _, err := s.QueryRowContext(context.Background(), "SELECT 1"); err != nil {
		t.Errorf("QueryRowContext: %v", err)
	}
	if _, err := s.BeginTx(context.Background()); err == nil {
		t.Error("expected BeginTx error in REST mode")
	}
	if err := s.MigrateContext(context.Background(), []Migration{{Version: 1, Name: "m", Up: "SELECT 1"}}); err != nil {
		t.Errorf("MigrateContext: %v", err)
	}
	if err := s.RollbackContext(context.Background(), 0); err != nil {
		t.Errorf("RollbackContext: %v", err)
	}
}

func TestRealSupabaseRESTQueryCache(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]map[string]any{{"n": float64(1)}})
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "svc",
		SupabaseAccessToken:    "tok",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if _, err := s.Query("SELECT cached"); err != nil {
		t.Fatalf("Query: %v", err)
	}
	if _, err := s.Query("SELECT cached"); err != nil {
		t.Fatalf("Query cached: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 HTTP call (cached), got %d", calls)
	}

	s.queryCache["SELECT cached"] = cachedResult{
		rows:      []Row{{"n": float64(42)}},
		expiresAt: time.Now().Add(-time.Second),
	}
	rows, err := s.Query("SELECT cached")
	if err != nil {
		t.Fatalf("Query after expiry: %v", err)
	}
	if len(rows) != 1 || rows[0]["n"] != float64(1) {
		t.Fatalf("expected fresh rows after expiry, got %v", rows)
	}

	s.cacheClear()
	if _, ok := s.cacheGet("SELECT cached"); ok {
		t.Error("expected cache cleared")
	}
}

func TestRealSupabaseRESTExecRowsAffected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]map[string]any{{"rows_affected": float64(7)}})
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "svc",
		SupabaseAccessToken:    "tok",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	res, err := s.Exec("INSERT INTO t VALUES (1)")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if res.RowsAffected != 7 {
		t.Errorf("expected 7 rows affected, got %d", res.RowsAffected)
	}
}

func TestRealSupabaseRESTQueryRowEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "svc",
		SupabaseAccessToken:    "tok",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	row, err := s.QueryRow("SELECT empty")
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if len(row) != 0 {
		t.Errorf("expected empty row, got %v", row)
	}
}

func TestRealSupabaseConnectNonRESTFailure(t *testing.T) {
	s := NewRealSupabase()
	cfg := &Config{
		User:     "test",
		Password: "test",
		Host:     "192.0.2.1",
		Port:     1,
		Database: "test",
		SSLMode:  "disable",
		Timeout:  1 * time.Second,
	}
	err := s.Connect(cfg)
	if err == nil {
		t.Fatal("expected error connecting to unreachable host")
	}
}

func TestRealSupabaseNotConnectedPaths(t *testing.T) {
	s := NewRealSupabase()
	if err := s.Ping(); err == nil {
		t.Error("expected Ping error when not connected")
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close nil: %v", err)
	}
	if _, err := s.ExecContext(context.Background(), "SELECT 1"); err == nil {
		t.Error("expected ExecContext error")
	}
	if _, err := s.QueryContext(context.Background(), "SELECT 1"); err == nil {
		t.Error("expected QueryContext error")
	}
	if _, err := s.QueryRowContext(context.Background(), "SELECT 1"); err == nil {
		t.Error("expected QueryRowContext error")
	}
	if _, err := s.BeginTx(context.Background()); err == nil {
		t.Error("expected BeginTx error")
	}
	if err := s.MigrateContext(context.Background(), nil); err == nil {
		t.Error("expected MigrateContext error")
	}
	if err := s.RollbackContext(context.Background(), 0); err == nil {
		t.Error("expected RollbackContext error")
	}
	if err := s.HealthCheck(); err == nil {
		t.Error("expected HealthCheck error")
	}
}

func TestRealSupabaseSQLiteBacked(t *testing.T) {
	s := NewRealSupabase()
	s.db = openSQLiteBacked(t)
	s.config = &Config{}

	res, err := s.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "erin", "erin@example.com")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}

	rows, err := s.Query("SELECT name FROM users")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row, err := s.QueryRow("SELECT email FROM users WHERE name = $1", "erin")
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if row["email"] != "erin@example.com" {
		t.Errorf("expected email, got %v", row["email"])
	}

	if err := s.Ping(); err != nil {
		t.Errorf("Ping: %v", err)
	}
	if err := s.HealthCheck(); err != nil {
		t.Errorf("HealthCheck: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestRealSupabaseSQLiteBackedTx(t *testing.T) {
	s := NewRealSupabase()
	s.db = openSQLiteBacked(t)
	s.config = &Config{}

	tx, err := s.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	res, err := tx.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "frank", "frank@example.com")
	if err != nil {
		t.Fatalf("tx Exec: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", res.RowsAffected)
	}
	rows, err := tx.Query("SELECT name FROM users WHERE name = $1", "frank")
	if err != nil {
		t.Fatalf("tx Query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	tx2, err := s.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx 2: %v", err)
	}
	if err := tx2.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

func TestRealSupabaseSQLiteBackedMigrateError(t *testing.T) {
	s := NewRealSupabase()
	s.db = openSQLiteBacked(t)
	s.config = &Config{}

	err := s.MigrateContext(context.Background(), []Migration{{Version: 1, Name: "m", Up: "SELECT 1"}})
	if err == nil {
		t.Fatal("expected error: Postgres DDL not supported by SQLite")
	}
	if err := s.Migrate([]Migration{{Version: 1, Name: "m"}}); err == nil {
		t.Error("expected Migrate error")
	}
}

func TestRealSupabaseRESTMigrateErrorPaths(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "relation does not exist"})
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "svc",
		SupabaseAccessToken:    "tok",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if err := s.Migrate([]Migration{{Version: 1, Name: "m", Up: "SELECT 1"}}); err == nil {
		t.Error("expected Migrate error")
	}
}
