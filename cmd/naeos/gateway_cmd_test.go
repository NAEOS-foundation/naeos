package main

import (
	"strings"
	"testing"
)

func TestGatewayCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "gateway")
	if err != nil {
		t.Fatalf("execute gateway failed: %v", err)
	}
}

func TestGatewayStatus(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "gateway", "status")
	if err != nil {
		t.Fatalf("gateway status failed: %v", err)
	}
	if !strings.Contains(output, "API Gateway Status") {
		t.Fatalf("expected gateway status header, got %q", output)
	}
}

func TestGatewayRateStatus(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "gateway", "rate-status")
	if err != nil {
		t.Fatalf("gateway rate-status failed: %v", err)
	}
	if !strings.Contains(output, "Rate Limiter") {
		t.Fatalf("expected rate limiter info, got %q", output)
	}
}

func TestGatewayCBStatus(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "gateway", "cb-status", "--name", "api")
	if err != nil {
		t.Fatalf("gateway cb-status failed: %v", err)
	}
	if !strings.Contains(output, "Circuit Breaker: api") {
		t.Fatalf("expected circuit breaker name, got %q", output)
	}
	if !strings.Contains(output, "State:") {
		t.Fatalf("expected state info, got %q", output)
	}
}

func TestGatewayLBListEmpty(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "gateway", "lb-list")
	if err != nil {
		t.Fatalf("gateway lb-list failed: %v", err)
	}
	if !strings.Contains(output, "No backends configured") {
		t.Fatalf("expected no backends message, got %q", output)
	}
}

func TestGatewayAddBackend(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "gateway", "add-backend", "--name", "backend1", "--url", "http://localhost:8080")
	if err != nil {
		t.Fatalf("gateway add-backend failed: %v", err)
	}
	if !strings.Contains(output, "Added backend") {
		t.Fatalf("expected add success message, got %q", output)
	}
}
