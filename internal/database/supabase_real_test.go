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

func TestRealSupabaseRESTConnect(t *testing.T) {
	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "test-svc-key",
		SupabaseURL:            "https://test-proj.supabase.co",
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect in REST mode should not error: %v", err)
	}
	if !s.restMode {
		t.Fatal("expected restMode to be true")
	}
	if err := s.Ping(); err != nil {
		t.Fatalf("Ping in REST mode should not error: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close in REST mode should not error: %v", err)
	}
}

func TestRealSupabaseRESTExecQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-access-token" {
			t.Errorf("expected Bearer test-access-token, got %q", r.Header.Get("Authorization"))
		}

		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}

		switch body.Query {
		case "SELECT 1":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]map[string]any{{"1": float64(1)}})
		case "CREATE TABLE test (id INT)":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]map[string]any{})
		default:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]map[string]any{})
		}
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "test-svc-key",
		SupabaseAccessToken:    "test-access-token",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	result, err := s.Query("SELECT 1")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if result[0]["1"] != float64(1) {
		t.Fatalf("expected 1, got %v", result[0]["1"])
	}

	_, err = s.Exec("CREATE TABLE test (id INT)")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	row, err := s.QueryRow("SELECT 1")
	if err != nil {
		t.Fatalf("QueryRow failed: %v", err)
	}
	if row["1"] != float64(1) {
		t.Fatalf("expected 1, got %v", row["1"])
	}

	_, err = s.Begin()
	if err == nil {
		t.Fatal("expected error for Begin in REST mode")
	}

	if err := s.HealthCheck(); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestRealSupabaseRESTMigrate(t *testing.T) {
	var calls []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		calls = append(calls, body.Query)

		// Return SELECT COUNT(*) result when checking migration version
		if body.Query == "SELECT COUNT(*) as cnt FROM _migrations WHERE version = 1" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]map[string]any{{"cnt": float64(0)}})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "test-svc-key",
		SupabaseAccessToken:    "test-access-token",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	migrations := []Migration{
		{Version: 1, Name: "initial", Up: "CREATE TABLE users (id INT)", Down: "DROP TABLE users"},
	}
	if err := s.Migrate(migrations); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	foundCreate := false
	foundInsert := false
	for _, c := range calls {
		if c == "CREATE TABLE IF NOT EXISTS _migrations (version INTEGER PRIMARY KEY, name TEXT, down_sql TEXT, applied_at TIMESTAMPTZ DEFAULT NOW())" {
			foundCreate = true
		}
		if c == "INSERT INTO _migrations (version, name, down_sql) VALUES (1, 'initial', 'DROP TABLE users')" {
			foundInsert = true
		}
	}
	if !foundCreate {
		t.Fatal("expected CREATE TABLE _migrations call")
	}
	if !foundInsert {
		t.Fatal("expected INSERT INTO _migrations call")
	}
}

func TestRealSupabaseCacheSetGet(t *testing.T) {
	t.Parallel()

	s := NewRealSupabase()
	rows := []Row{{"id": int64(1), "name": "test"}}
	s.cacheSet("SELECT 1", rows)

	got, ok := s.cacheGet("SELECT 1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if got[0]["id"] != int64(1) {
		t.Errorf("expected id 1, got %v", got[0]["id"])
	}
}

func TestRealSupabaseCacheGetMiss(t *testing.T) {
	t.Parallel()

	s := NewRealSupabase()
	_, ok := s.cacheGet("SELECT nonexistent")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestRealSupabaseCacheGetExpired(t *testing.T) {
	t.Parallel()

	s := NewRealSupabase()
	s.queryCache["test"] = cachedResult{rows: []Row{{"a": 1}}, expiresAt: time.Now().Add(-time.Hour)}
	_, ok := s.cacheGet("test")
	if ok {
		t.Error("expected cache miss for expired entry")
	}
	_, ok = s.queryCache["test"]
	if ok {
		t.Error("expired entry should be deleted from map")
	}
}

func TestRealSupabaseCacheClear(t *testing.T) {
	t.Parallel()

	s := NewRealSupabase()
	s.cacheSet("SELECT 1", []Row{{"id": int64(1)}})
	s.cacheClear()
	_, ok := s.cacheGet("SELECT 1")
	if ok {
		t.Error("expected cache miss after clear")
	}
}

func TestRealSupabaseCacheKey(t *testing.T) {
	t.Parallel()

	s := NewRealSupabase()
	key := s.cacheKey("  SELECT 1  ")
	if key != "SELECT 1" {
		t.Errorf("expected 'SELECT 1', got %q", key)
	}
}

func TestRealSupabaseIsWriteQuery(t *testing.T) {
	t.Parallel()

	s := NewRealSupabase()
	tests := []struct {
		query string
		want  bool
	}{
		{"SELECT * FROM users", false},
		{"INSERT INTO users VALUES (1)", true},
		{"UPDATE users SET name = 'x'", true},
		{"DELETE FROM users", true},
		{"CREATE TABLE t (id INT)", true},
		{"DROP TABLE t", true},
		{"ALTER TABLE t ADD COLUMN x INT", true},
		{"TRUNCATE TABLE t", true},
		{"REPLACE INTO t VALUES (1)", true},
		{"  select * from users", false},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := s.isWriteQuery(tt.query)
			if got != tt.want {
				t.Errorf("isWriteQuery(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestRealSupabaseDefaultContextWithTimeout(t *testing.T) {
	s := NewRealSupabase()
	s.config = &Config{Timeout: 5 * time.Second}
	ctx, cancel := s.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSupabaseDefaultContextWithoutTimeout(t *testing.T) {
	s := NewRealSupabase()
	s.config = &Config{}
	ctx, cancel := s.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSupabaseDefaultContextNilConfig(t *testing.T) {
	s := NewRealSupabase()
	ctx, cancel := s.defaultContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}
}

func TestRealSupabaseNotConnected(t *testing.T) {
	s := NewRealSupabase()

	err := s.Ping()
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.Exec("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.ExecContext(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.Query("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.QueryContext(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.QueryRow("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.QueryRowContext(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.Begin()
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = s.BeginTx(context.Background())
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = s.Migrate(nil)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = s.MigrateContext(context.Background(), nil)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = s.Rollback(0)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = s.RollbackContext(context.Background(), 0)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = s.HealthCheck()
	if err == nil {
		t.Error("expected error when not connected")
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}
}

func TestRealSupabaseHealthCheckError(t *testing.T) {
	s := NewRealSupabase()
	err := s.HealthCheck()
	if err == nil {
		t.Fatal("expected error for HealthCheck without Connect")
	}
}

func TestRealSupabaseConnectFailure(t *testing.T) {
	s := NewRealSupabase()
	err := s.Connect(&Config{
		Host:     "192.0.2.1",
		Port:     1,
		User:     "u",
		Password: "p",
		Database: "d",
		Timeout:  time.Second,
	})
	if err == nil {
		t.Error("expected error connecting to unreachable host")
	}
}

func TestRealSupabaseConnectFailureWithSSL(t *testing.T) {
	s := NewRealSupabase()
	err := s.Connect(&Config{
		Host:     "192.0.2.1",
		Port:     1,
		User:     "u",
		Password: "p",
		Database: "d",
		SSLMode:  "disable",
		Timeout:  time.Second,
	})
	if err == nil {
		t.Error("expected error connecting to unreachable host")
	}
}

func TestRealSupabaseRestClientDefaultURL(t *testing.T) {
	s := NewRealSupabase()
	s.config = &Config{
		SupabaseProjectRef:     "my-project",
		SupabaseServiceRoleKey: "svc-key",
	}
	client := s.restClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestRealSupabaseBeginTxRESTMode(t *testing.T) {
	s := NewRealSupabase()
	s.config = &Config{SupabaseProjectRef: "test"}
	s.restMode = true

	_, err := s.Begin()
	if err == nil {
		t.Error("expected error for Begin in REST mode")
	}

	_, err = s.BeginTx(context.Background())
	if err == nil {
		t.Error("expected error for BeginTx in REST mode")
	}
}

func TestRealSupabasePingRestModeNoConfig(t *testing.T) {
	s := NewRealSupabase()
	s.restMode = true
	err := s.Ping()
	if err == nil {
		t.Error("expected error for Ping in restMode without config")
	}
}

func TestRealSupabaseHealthCheckRestModeNoConfig(t *testing.T) {
	s := NewRealSupabase()
	s.restMode = true
	err := s.HealthCheck()
	if err == nil {
		t.Error("expected error for HealthCheck in restMode without config")
	}
}

func TestRealSupabaseRESTExecError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "test-svc-key",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	_, err := s.Query("SELECT 1")
	if err == nil {
		t.Error("expected error from Query with server error")
	}

	_, err = s.Exec("INSERT INTO t VALUES (1)")
	if err == nil {
		t.Error("expected error from Exec with server error")
	}
}

func TestRealSupabaseRESTRollback(t *testing.T) {
	var calls []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		calls = append(calls, body.Query)

		if body.Query == "SELECT version, name, down_sql FROM _migrations WHERE version > 0 ORDER BY version DESC" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode([]map[string]any{
				{"version": float64(2), "name": "add_column", "down_sql": "ALTER TABLE users DROP COLUMN status"},
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]map[string]any{})
	}))
	defer ts.Close()

	s := NewRealSupabase()
	cfg := &Config{
		SupabaseProjectRef:     "test-proj",
		SupabaseServiceRoleKey: "test-svc-key",
		SupabaseAccessToken:    "test-access-token",
		SupabaseURL:            ts.URL,
		SupabaseManagementURL:  ts.URL,
	}
	if err := s.Connect(cfg); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if err := s.Rollback(0); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	foundDown := false
	foundDelete := false
	for _, c := range calls {
		if c == "ALTER TABLE users DROP COLUMN status" {
			foundDown = true
		}
		if c == "DELETE FROM _migrations WHERE version = 2" {
			foundDelete = true
		}
	}
	if !foundDown {
		t.Fatal("expected down migration call")
	}
	if !foundDelete {
		t.Fatal("expected DELETE FROM _migrations call")
	}
}
