package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestDashboardHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "dashboard", "--help")
	if err != nil {
		t.Fatalf("dashboard help failed: %v", err)
	}
	if !strings.Contains(output, "Start the NAEOS web dashboard") {
		t.Fatalf("expected dashboard help text, got %q", output)
	}
	if !strings.Contains(output, "--port") {
		t.Fatalf("expected --port flag in help, got %q", output)
	}
}

func TestDashboardPortConflict(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	root := NewRootCommand()
	_, err = executeCommand(root, "dashboard", "--port", fmt.Sprintf("127.0.0.1:%d", port))
	if err == nil {
		t.Fatal("expected dashboard to fail when the port is already in use")
	}
}
