package main

import (
	"strings"
	"testing"
)

func TestPerfCommandShowsHelp(t *testing.T) {
	root := newRootCommand()
	_, err := executeCommand(root, "perf")
	if err != nil {
		t.Fatalf("execute perf failed: %v", err)
	}
}

func TestPerfPoolCreate(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "perf", "pool-create", "--name", "db", "--min", "2", "--max", "10")
	if err != nil {
		t.Fatalf("perf pool-create failed: %v", err)
	}
	if !strings.Contains(output, "Created pool") {
		t.Fatalf("expected pool created message, got %q", output)
	}
	if !strings.Contains(output, "min=2") || !strings.Contains(output, "max=10") {
		t.Fatalf("expected pool config in output, got %q", output)
	}
}

func TestPerfPoolAcquireRelease(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "perf", "pool-acquire", "--name", "db")
	if err != nil {
		t.Fatalf("perf pool-acquire failed: %v", err)
	}
	if !strings.Contains(output, "Acquired connection") {
		t.Fatalf("expected acquire message, got %q", output)
	}
	if !strings.Contains(output, "Released connection") {
		t.Fatalf("expected release message, got %q", output)
	}
}

func TestPerfPoolStats(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "perf", "pool-stats", "--name", "db")
	if err != nil {
		t.Fatalf("perf pool-stats failed: %v", err)
	}
	if !strings.Contains(output, "Pool: db") {
		t.Fatalf("expected pool name in output, got %q", output)
	}
}

func TestPerfCacheSet(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "perf", "cache-set", "--key", "mykey", "--value", "myvalue", "--ttl", "60s")
	if err != nil {
		t.Fatalf("perf cache-set failed: %v", err)
	}
	if !strings.Contains(output, "Cached 'mykey'") {
		t.Fatalf("expected cache set message, got %q", output)
	}
}

func TestPerfCacheGetMiss(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "perf", "cache-get", "--key", "nonexistent")
	if err != nil {
		t.Fatalf("perf cache-get failed: %v", err)
	}
	if !strings.Contains(output, "Cache miss") {
		t.Fatalf("expected cache miss message, got %q", output)
	}
}

func TestPerfCacheStats(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "perf", "cache-stats")
	if err != nil {
		t.Fatalf("perf cache-stats failed: %v", err)
	}
	if !strings.Contains(output, "Cache: naeos") {
		t.Fatalf("expected cache stats, got %q", output)
	}
}
