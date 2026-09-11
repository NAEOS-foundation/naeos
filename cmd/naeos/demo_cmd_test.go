package main

import (
	"context"
	"strings"
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
