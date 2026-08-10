package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMCPHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "mcp", "--help")
	if err != nil {
		t.Fatalf("mcp --help failed: %v", err)
	}
	if !strings.Contains(output, "Model Context Protocol") {
		t.Fatalf("expected mcp help text, got %q", output)
	}
}

func TestMCPInvalidPortFlag(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "mcp", "--port", "not-a-port")
	if err == nil {
		t.Fatal("expected error for invalid port flag value")
	}
}

func TestMCPListenError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := NewRootCommand()
	root.SetContext(ctx)
	_, err := executeCommand(root, "mcp", "--port", "-1")
	if err == nil {
		t.Fatal("expected error when ListenAndServe fails")
	}
	cancel()
}

func TestMCPStartAndStop(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := NewRootCommand()
	root.SetContext(ctx)

	type result struct {
		output string
		err    error
	}
	resultCh := make(chan result, 1)
	go func() {
		output, err := executeCommand(root, "mcp", "--port", strconv.Itoa(port))
		resultCh <- result{output: output, err: err}
	}()

	started := false
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 300*time.Millisecond)
		if dialErr == nil {
			conn.Close()
			started = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !started {
		cancel()
		t.Fatal("MCP server did not start within timeout")
	}

	cancel()

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("mcp server should stop cleanly after cancel: %v", res.err)
		}
		if !strings.Contains(res.output, "NAEOS MCP server starting") {
			t.Fatalf("expected startup message, got %q", res.output)
		}
		if !strings.Contains(res.output, "MCP server stopped.") {
			t.Fatalf("expected stop message, got %q", res.output)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("MCP server did not stop after context cancel")
	}
}
