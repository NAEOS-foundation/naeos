//go:build !nosql

package database

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestRealSupabaseRESTHealthCheckError(t *testing.T) {
	s := NewRealSupabase()
	err := s.HealthCheck()
	if err == nil {
		t.Fatal("expected error for HealthCheck without Connect")
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
