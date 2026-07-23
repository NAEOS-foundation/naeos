package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
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
