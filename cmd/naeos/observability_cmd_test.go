package main

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *safeBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

func TestObservabilityCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "observability")
	if err != nil {
		t.Fatalf("execute observability failed: %v", err)
	}
}

func TestObsTrace(t *testing.T) {
	root := NewRootCommand()
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
	root := NewRootCommand()
	output, err := executeCommand(root, "observability", "log", "--level", "info", "--message", "Server started")
	if err != nil {
		t.Fatalf("observability log failed: %v", err)
	}
	if !strings.Contains(output, "[info]") {
		t.Fatalf("expected log level in output, got %q", output)
	}
}

func TestObsMetrics(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "observability", "metrics")
	if err != nil {
		t.Fatalf("observability metrics failed: %v", err)
	}
	if !strings.Contains(output, "NAME") {
		t.Fatalf("expected metrics table header, got %q", output)
	}
}

func TestObsStatus(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "observability", "status")
	if err != nil {
		t.Fatalf("observability status failed: %v", err)
	}
	if !strings.Contains(output, "Observability Stack") {
		t.Fatalf("expected status header, got %q", output)
	}
}

func TestObsDashboard(t *testing.T) {
	buf := &safeBuffer{}
	root := NewRootCommand()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"observability", "dashboard", "--port", "0"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	done := make(chan error, 1)
	go func() {
		_, err := root.ExecuteC()
		done <- err
	}()

	var output string
	select {
	case <-time.After(500 * time.Millisecond):
		output = buf.String()
	case err := <-done:
		if err != nil {
			t.Fatalf("observability dashboard failed: %v", err)
		}
		output = buf.String()
	}

	if !strings.Contains(output, "Starting observability dashboard") {
		t.Fatalf("expected dashboard start message, got %q", output)
	}
}
