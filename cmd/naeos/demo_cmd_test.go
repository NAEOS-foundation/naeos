package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDemoCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "demo", "--help")
	if err != nil {
		t.Fatalf("demo --help failed: %v", err)
	}
	if !strings.Contains(output, "investor demo") {
		t.Fatalf("expected demo help text, got %q", output)
	}
	if !strings.Contains(output, "--siem-endpoint") || !strings.Contains(output, "--otlp-endpoint") {
		t.Fatalf("expected observability flags in help, got %q", output)
	}
}

func TestDemoCommandObservability(t *testing.T) {
	var posted atomic.Int64
	siem := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posted.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer siem.Close()

	buf := &safeBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := NewRootCommand()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetContext(ctx)
	root.SetArgs([]string{"demo", "--addr", ":0",
		"--siem-endpoint", siem.URL,
		"--otlp-endpoint", "http://127.0.0.1:4318",
		"--tenant-id", "acme"})
	root.SilenceErrors = true
	root.SilenceUsage = true

	done := make(chan error, 1)
	go func() {
		_, err := root.ExecuteC()
		done <- err
	}()

	ok := waitString(t, buf, "SIEM forwarding", 2*time.Second)
	if !ok {
		t.Fatalf("expected SIEM forwarding startup line, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "acme") {
		t.Errorf("expected tenant in startup line, got %q", buf.String())
	}

	cancel()
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("demo command did not stop after cancel")
	case <-done:
	}
}

func waitString(t *testing.T, w *safeBuffer, substr string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(w.String(), substr) {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestDemoCommandServesAPIs(t *testing.T) {
	buf := &safeBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := NewRootCommand()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetContext(ctx)
	root.SetArgs([]string{"demo", "--addr", ":0"})
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
		cancel()
		if err != nil {
			t.Fatalf("demo command failed: %v", err)
		}
	}

	if !strings.Contains(output, "investor demo listening") {
		t.Fatalf("expected demo listening message, got %q", output)
	}

	cancel()
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("demo command did not stop after cancel")
	case <-done:
	}
}
