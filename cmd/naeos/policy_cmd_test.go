package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

func writePolicyFile(t *testing.T, p *policy.Policy) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal policy: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	return path
}

func TestPolicyValidateCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := writePolicyFile(t, &policy.Policy{
		ID:      "p1",
		Version: "1.0.0",
		Default: policy.DecisionAllow,
	})
	root := NewRootCommand()
	out, err := executeCommand(root, "policy", "validate", "--file", path)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !strings.Contains(out, "is valid") {
		t.Fatalf("expected valid message, got %q", out)
	}
}

func TestPolicyValidateCommandInvalid(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := writePolicyFile(t, &policy.Policy{ID: "p1", Version: "1.0.0", Default: policy.Decision("NOPE")})
	root := NewRootCommand()
	_, err := executeCommand(root, "policy", "validate", "--file", path)
	if err == nil {
		t.Fatal("expected error for invalid policy")
	}
}

func TestPolicyRegisterListGet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := writePolicyFile(t, &policy.Policy{
		ID:      "prod-deploy",
		Version: "1.0.0",
		Name:    "Production Deployment",
		Scope:   policy.Scope{Resource: "deploy", Action: "run", Environment: "production"},
		Default: policy.DecisionRequireApproval,
	})
	root := NewRootCommand()
	if _, err := executeCommand(root, "policy", "register", "--file", path); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	out, err := executeCommand(root, "policy", "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(out, "prod-deploy") {
		t.Fatalf("expected prod-deploy in list, got %q", out)
	}
	out, err = executeCommand(root, "policy", "get", "prod-deploy")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(out, "REQUIRE_APPROVAL") {
		t.Fatalf("expected REQUIRE_APPROVAL in get, got %q", out)
	}
}

func TestPolicyGetUnknown(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := NewRootCommand()
	_, err := executeCommand(root, "policy", "get", "missing")
	if err == nil {
		t.Fatal("expected error for unknown policy")
	}
}

func TestPolicySetActive(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := NewRootCommand()
	dir := t.TempDir()
	p1 := filepath.Join(dir, "v1.json")
	p2 := filepath.Join(dir, "v2.json")
	d1, _ := json.Marshal(&policy.Policy{ID: "p", Version: "1.0.0", Default: policy.DecisionAllow})
	d2, _ := json.Marshal(&policy.Policy{ID: "p", Version: "2.0.0", Default: policy.DecisionDeny})
	_ = os.WriteFile(p1, d1, 0o644)
	_ = os.WriteFile(p2, d2, 0o644)
	if _, err := executeCommand(root, "policy", "register", "--file", p1); err != nil {
		t.Fatalf("register v1: %v", err)
	}
	if _, err := executeCommand(root, "policy", "register", "--file", p2); err != nil {
		t.Fatalf("register v2: %v", err)
	}
	if _, err := executeCommand(root, "policy", "set-active", "p", "1.0.0"); err != nil {
		t.Fatalf("set-active: %v", err)
	}
	out, err := executeCommand(root, "policy", "get", "p")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !strings.Contains(out, `"Version": "1.0.0"`) {
		t.Fatalf("expected active version 1.0.0, got %q", out)
	}
}

func TestControlEvaluateNoPolicyDeny(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := NewRootCommand()
	out, err := executeCommand(root, "control", "evaluate", "--resource", "deploy", "--action", "run")
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if !strings.Contains(out, "DENY") {
		t.Fatalf("expected DENY (fail-closed), got %q", out)
	}
}

func TestControlEvaluateApprovalRequired(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := writePolicyFile(t, &policy.Policy{
		ID:      "prod-deploy",
		Version: "1.0.0",
		Scope:   policy.Scope{Resource: "deploy", Action: "run", Environment: "production"},
		Default: policy.DecisionRequireApproval,
	})
	root := NewRootCommand()
	if _, err := executeCommand(root, "policy", "register", "--file", path); err != nil {
		t.Fatalf("register: %v", err)
	}
	out, err := executeCommand(root, "control", "evaluate",
		"--resource", "deploy", "--action", "run", "--environment", "production")
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if !strings.Contains(out, "REQUIRE_APPROVAL") {
		t.Fatalf("expected REQUIRE_APPROVAL, got %q", out)
	}
}

func TestControlEvaluateJSON(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := NewRootCommand()
	out, err := executeCommand(root, "control", "evaluate",
		"--resource", "deploy", "--action", "run", "--output", "json", "--environment", "staging")
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if !strings.Contains(out, `"Decision": "DENY"`) {
		t.Fatalf("expected JSON DENY decision, got %q", out)
	}
}
