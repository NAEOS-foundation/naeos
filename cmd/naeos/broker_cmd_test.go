package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrokerListEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "broker", "list")
	if err != nil {
		t.Fatalf("broker list failed: %v", err)
	}
	if !strings.Contains(output, "No broker connections.") {
		t.Fatalf("expected 'No broker connections.', got %q", output)
	}
}

func TestBrokerListWithSavedConnection(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	writeTestFile(t, filepath.Join(dir, ".config", "naeos"), "brokers.json",
		`[{"name":"myredis","driver":"redis","config":{"Host":"localhost","Port":6379,"Password":"","DB":0,"Timeout":0}}]`)

	root := NewRootCommand()
	output, err := executeCommand(root, "broker", "list")
	if err != nil {
		t.Fatalf("broker list failed: %v", err)
	}
	if !strings.Contains(output, "myredis") || !strings.Contains(output, "redis") {
		t.Fatalf("expected saved connection in output, got %q", output)
	}
}

func TestBrokerConnectMemory(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "broker", "connect", "--type", "memory", "--name", "mymem")
	if err != nil {
		t.Fatalf("broker connect memory failed: %v", err)
	}
	if !strings.Contains(output, "Connected to memory broker 'mymem'") {
		t.Fatalf("expected 'Connected to memory broker', got %q", output)
	}
}

func TestBrokerConnectUnsupportedType(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "broker", "connect", "--type", "weird", "--name", "b1")
	if err == nil {
		t.Fatal("expected error for unsupported broker type")
	}
}

func TestBrokerConnectMissingName(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "broker", "connect", "--type", "memory")
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestBrokerPublishNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "broker", "publish", "--name", "ghost", "--channel", "events", "--message", "hello")
	if err == nil {
		t.Fatal("expected error for unknown broker")
	}
}

func TestBrokerPublishMissingFlags(t *testing.T) {
	root := NewRootCommand()
	if _, err := executeCommand(root, "broker", "publish"); err == nil {
		t.Fatal("expected error when required flags are missing")
	}
	_, err := executeCommand(root, "broker", "publish", "--name", "b1", "--channel", "c")
	if err == nil {
		t.Fatal("expected error when --message is missing")
	}
}

func TestBrokerDisconnectHappy(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	if _, err := executeCommand(root, "broker", "connect", "--type", "memory", "--name", "mymem"); err != nil {
		t.Fatalf("broker connect failed: %v", err)
	}

	output, err := executeCommand(root, "broker", "disconnect", "--name", "mymem")
	if err != nil {
		t.Fatalf("broker disconnect failed: %v", err)
	}
	if !strings.Contains(output, "Disconnected from 'mymem'.") {
		t.Fatalf("expected 'Disconnected from', got %q", output)
	}
}

func TestBrokerDisconnectNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "broker", "disconnect", "--name", "ghost")
	if err == nil {
		t.Fatal("expected error when disconnecting an unknown broker")
	}
}

func TestBrokerDisconnectMissingName(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "broker", "disconnect")
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestBrokerPersistenceAcrossCommands(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	root := NewRootCommand()
	if _, err := executeCommand(root, "broker", "connect", "--type", "memory", "--name", "persisted"); err != nil {
		t.Fatalf("broker connect failed: %v", err)
	}

	root = NewRootCommand()
	output, err := executeCommand(root, "broker", "list")
	if err != nil {
		t.Fatalf("broker list failed: %v", err)
	}
	if !strings.Contains(output, "persisted") {
		t.Fatalf("expected persisted connection listed, got %q", output)
	}

	if _, err := os.Stat(filepath.Join(dir, ".config", "naeos", "brokers.json")); err != nil {
		t.Fatalf("expected brokers.json to be persisted: %v", err)
	}
}
