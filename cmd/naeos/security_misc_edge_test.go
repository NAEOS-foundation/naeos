package main

import (
	"strings"
	"testing"
)

func TestSecuritySetSecretMissingKey(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security", "set-secret", "--name", "x", "--value", "y")
	if err == nil {
		t.Fatal("expected error when --key missing")
	}
}

func TestSecurityGetSecretMissingKey(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security", "get-secret", "--name", "x")
	if err == nil {
		t.Fatal("expected error when --key missing")
	}
}

func TestSecurityGetSecretWithKeyNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security", "get-secret", "--name", "ghost", "--key", "k")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSecurityListSecretsEmpty(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "list-secrets")
	if err != nil {
		t.Fatalf("list-secrets: %v", err)
	}
	if !strings.Contains(output, "No secrets stored.") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSecuritySanitizeModes(t *testing.T) {
	root := NewRootCommand()
	for _, mode := range []string{"sql", "xss", "path"} {
		output, err := executeCommand(root, "security", "sanitize", "--input", "input; DROP TABLE users", "--mode", mode)
		if err != nil {
			t.Fatalf("sanitize %s: %v", mode, err)
		}
		if !strings.Contains(output, "Sanitized:") {
			t.Errorf("mode %s: unexpected output %q", mode, output)
		}
	}
}

func TestSecuritySanitizeMissingInput(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security", "sanitize")
	if err == nil {
		t.Fatal("expected error when --input missing")
	}
}

func TestSecurityHashPasswordMissing(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security", "hash-password")
	if err == nil {
		t.Fatal("expected error when --password missing")
	}
}

func TestSecurityValidateUnknownRule(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "validate", "--name", "unknown", "--value", "anything")
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !strings.Contains(output, "Validation failed") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSecurityValidateMissingValue(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security", "validate", "--name", "email")
	if err == nil {
		t.Fatal("expected error when --value missing")
	}
}

func TestSecurityAuditMultipleSeverities(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "server.go", "package main\n\nfunc main() {\n\t_ = \"0.0.0.0\"\n\thttp.ListenAndServe(\":8080\", nil)\n}\n")
	writeTestFile(t, dir, "debug.yaml", "debug: true\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "security", "audit", "--input", dir)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if !strings.Contains(output, "MEDIUM") {
		t.Errorf("expected medium findings, got %q", output)
	}
	if !strings.Contains(output, "LOW") {
		t.Errorf("expected low findings, got %q", output)
	}
	if !strings.Contains(output, "Summary:") {
		t.Errorf("expected summary line, got %q", output)
	}
}

func TestSecurityAuditJSONOutput(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", "package main\n\nvar password = \"secret123\"\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "security", "audit", "--input", dir, "--output", "json")
	if err != nil {
		t.Fatalf("audit json: %v", err)
	}
	if !strings.Contains(output, `"summary"`) || !strings.Contains(output, `"critical"`) {
		t.Errorf("expected json summary, got %q", output)
	}
}

func TestSecurityAuditYAMLOutput(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "security", "audit", "--input", dir, "--output", "yaml")
	if err != nil {
		t.Fatalf("audit yaml: %v", err)
	}
	if !strings.Contains(output, "directory:") {
		t.Errorf("expected yaml output, got %q", output)
	}
}

func TestSecurityAuditNonexistentDir(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "audit", "--input", "/nonexistent/dir")
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if !strings.Contains(output, "Scanned 0 files") {
		t.Errorf("expected zero files scanned, got %q", output)
	}
}

func TestMigrationStatusNoConnectionsEdge(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := NewRootCommand()
	output, err := executeCommand(root, "migration", "status")
	if err != nil {
		t.Fatalf("migration status: %v", err)
	}
	if !strings.Contains(output, "No database connections configured.") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestAuthCreateRoleWithDenyAndParents(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-role", "limited",
		"--permission", "spec:read", "--permission", "spec:write",
		"--deny", "spec:delete",
		"--parent", "viewer")
	if err != nil {
		t.Fatalf("create-role: %v", err)
	}
	if !strings.Contains(output, "Created role: limited") {
		t.Errorf("unexpected output: %q", output)
	}
}
