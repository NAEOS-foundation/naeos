package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/supabase"
)

func dbTestName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestDBCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db")
	if err != nil {
		t.Fatalf("execute db failed: %v", err)
	}
}

func TestDBConnectSQLite(t *testing.T) {
	name := dbTestName("testdb")
	root := NewRootCommand()
	output, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", "test.db", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}
	if !strings.Contains(output, "Connected to") {
		t.Fatalf("expected connection success, got %q", output)
	}
	executeCommand(root, "db", "disconnect", "--name", name)
}

func TestDBDisconnect(t *testing.T) {
	name := dbTestName("disconndb")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", "test.db", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}

	output, err := executeCommand(root, "db", "disconnect", "--name", name)
	if err != nil {
		t.Fatalf("db disconnect failed: %v", err)
	}
	if !strings.Contains(output, "Disconnected") {
		t.Fatalf("expected disconnect message, got %q", output)
	}
}

func TestDBStatus(t *testing.T) {
	name := dbTestName("statusdb")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", ":memory:", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)

	output, err := executeCommand(root, "db", "status", "--name", name)
	if err != nil {
		t.Fatalf("db status failed: %v", err)
	}
	if !strings.Contains(output, "Connection:") {
		t.Fatalf("expected connection info, got %q", output)
	}
}

func TestDBConnectInvalidType(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "invalid", "--name", "faildb")
	if err == nil {
		t.Fatal("expected error for invalid database type")
	}
}

func TestDBMigrateNoSavedConnection(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "migrate", "--name", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent connection")
	}
}

func TestDBList(t *testing.T) {
	name := dbTestName("listdb")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", "list.db", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)

	output, err := executeCommand(root, "db", "list")
	if err != nil {
		t.Fatalf("db list failed: %v", err)
	}
	if !strings.Contains(output, name) {
		t.Fatalf("expected connection name in list, got %q", output)
	}
}

func TestDBConnectSupabaseFlagNoConfig(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--name", "supadb", "--supabase")
	if err == nil {
		t.Fatal("expected error for --supabase without config")
	}
	if !strings.Contains(err.Error(), "supabase not configured") {
		t.Fatalf("expected 'supabase not configured' error, got %v", err)
	}
}

func TestDBMigrateWithDir(t *testing.T) {
	name := dbTestName("migratedir")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", ":memory:", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)

	output, err := executeCommand(root, "db", "migrate", "--name", name)
	if err != nil {
		t.Fatalf("db migrate failed: %v", err)
	}
	if !strings.Contains(output, "Applied") {
		t.Fatalf("expected 'Applied' in output, got %q", output)
	}
}

func TestDBMigrateInvalidDriver(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "migrate", "--name", "nonexistent")
	if err == nil {
		t.Fatal("expected error for invalid driver")
	}
}

func TestDBStatusJSON(t *testing.T) {
	name := dbTestName("statusjson")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", ":memory:", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)

	output, err := executeCommand(root, "db", "status", "--name", name, "--output", "json")
	if err != nil {
		t.Fatalf("db status --output json failed: %v", err)
	}
	if !strings.Contains(output, "HEALTHY") {
		t.Fatalf("expected 'HEALTHY' in json output, got %q", output)
	}
}

func TestDBStatusNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "status", "--name", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent connection")
	}
}

func TestDBConnectInvalidHost(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", "failhost")
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

func TestDBDisconnectNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "disconnect", "--name", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent connection")
	}
}

func TestDBListEmpty(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "list")
	if err != nil {
		t.Fatalf("db list failed: %v", err)
	}
}

func TestDBListJSON(t *testing.T) {
	name := dbTestName("listjson")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", "list.json.db", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)

	output, err := executeCommand(root, "db", "list", "--output", "json")
	if err != nil {
		t.Fatalf("db list --output json failed: %v", err)
	}
	if !strings.Contains(output, name) {
		t.Fatalf("expected connection name in json output, got %q", output)
	}
}

func TestDBConnectSupabaseWithConfig(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	cfg := &supabase.Config{
		ProjectRef:     "test-proj",
		URL:            "https://test-proj.supabase.co",
		AnonKey:        "test-anon",
		ServiceRoleKey: "test-svc",
	}
	if err := supabase.SaveConfig(cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	name := dbTestName("supaconn")
	root := NewRootCommand()
	output, err := executeCommand(root, "db", "connect", "--name", name, "--supabase")
	if err != nil {
		t.Fatalf("db connect --supabase failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)
	if !strings.Contains(output, "Connected to supabase database") {
		t.Fatalf("expected 'Connected to supabase database', got %q", output)
	}
}

func TestDBStatusYAML(t *testing.T) {
	name := dbTestName("statusyaml")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", ":memory:", "--user", "testuser")
	if err != nil {
		t.Fatalf("db connect failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)

	output, err := executeCommand(root, "db", "status", "--name", name, "--output", "yaml")
	if err != nil {
		t.Fatalf("db status --output yaml failed: %v", err)
	}
	if !strings.Contains(output, "HEALTHY") {
		t.Fatalf("expected 'HEALTHY' in yaml output, got %q", output)
	}
}

func TestDBConnectDuplicateName(t *testing.T) {
	name := dbTestName("dupdb")
	root := NewRootCommand()
	_, err := executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", "dup1.db", "--user", "testuser")
	if err != nil {
		t.Fatalf("first db connect failed: %v", err)
	}
	defer executeCommand(root, "db", "disconnect", "--name", name)

	_, err = executeCommand(root, "db", "connect", "--type", "sqlite", "--name", name, "--database", "dup2.db", "--user", "testuser")
	if err == nil {
		t.Fatal("expected error for duplicate connection name")
	}
}
