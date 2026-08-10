package main

import (
	"strings"
	"testing"
)

func TestMonitorHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "monitor", "--help")
	if err != nil {
		t.Fatalf("monitor --help failed: %v", err)
	}
	if !strings.Contains(output, "Prometheus metrics") {
		t.Fatalf("expected monitor help text, got %q", output)
	}
}

func TestMonitorInvalidPort(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "monitor", "--port", "127.0.0.1:99999")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if !strings.Contains(err.Error(), "listen") && !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected listen error mentioning port, got %v", err)
	}
}
