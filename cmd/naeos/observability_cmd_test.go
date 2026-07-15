package main

import (
	"strings"
	"testing"
)

func TestObservabilityCommandShowsHelp(t *testing.T) {
	root := newRootCommand()
	_, err := executeCommand(root, "observability")
	if err != nil {
		t.Fatalf("execute observability failed: %v", err)
	}
}

func TestObsTrace(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "observability", "trace", "--name", "http-request")
	if err != nil {
		t.Fatalf("observability trace failed: %v", err)
	}
	if !strings.Contains(output, "Trace:") {
		t.Fatalf("expected trace output, got %q", output)
	}
	if !strings.Contains(output, "Span:") {
		t.Fatalf("expected span info, got %q", output)
	}
}

func TestObsLog(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "observability", "log", "--level", "info", "--message", "Server started")
	if err != nil {
		t.Fatalf("observability log failed: %v", err)
	}
	if !strings.Contains(output, "[info]") {
		t.Fatalf("expected log level in output, got %q", output)
	}
}

func TestObsMetrics(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "observability", "metrics")
	if err != nil {
		t.Fatalf("observability metrics failed: %v", err)
	}
	if !strings.Contains(output, "NAME") {
		t.Fatalf("expected metrics table header, got %q", output)
	}
}

func TestObsStatus(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "observability", "status")
	if err != nil {
		t.Fatalf("observability status failed: %v", err)
	}
	if !strings.Contains(output, "Observability Stack") {
		t.Fatalf("expected status header, got %q", output)
	}
}

func TestObsDashboard(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "observability", "dashboard", "--port", "8080")
	if err != nil {
		t.Fatalf("observability dashboard failed: %v", err)
	}
	if !strings.Contains(output, "Starting observability dashboard") {
		t.Fatalf("expected dashboard start message, got %q", output)
	}
	if !strings.Contains(output, "8080") {
		t.Fatalf("expected port in output, got %q", output)
	}
}
