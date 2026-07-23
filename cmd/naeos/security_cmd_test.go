package main

import (
	"strings"
	"testing"
)

func TestSecurityCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security")
	if err != nil {
		t.Fatalf("execute security failed: %v", err)
	}
}

func TestSecuritySetSecret(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "set-secret", "--name", "test-secret", "--value", "my-secret-value", "--key", "test-key-123")
	if err != nil {
		t.Fatalf("set-secret failed: %v", err)
	}
	if !strings.Contains(output, "stored successfully") {
		t.Fatalf("expected success message, got %q", output)
	}
}

func TestSecurityGetSecretNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "security", "get-secret", "--name", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent secret")
	}
}

func TestSecuritySanitize(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "sanitize", "--input", "<script>alert('xss')</script>")
	if err != nil {
		t.Fatalf("sanitize failed: %v", err)
	}
	if !strings.Contains(output, "Sanitized:") {
		t.Fatalf("expected sanitized output, got %q", output)
	}
}

func TestSecuritySanitizeHTML(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "sanitize", "--input", "<b>bold</b>", "--mode", "html")
	if err != nil {
		t.Fatalf("sanitize html failed: %v", err)
	}
	if !strings.Contains(output, "Sanitized:") {
		t.Fatalf("expected sanitized output, got %q", output)
	}
}

func TestSecurityHashPassword(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "hash-password", "--password", "mypassword123")
	if err != nil {
		t.Fatalf("hash-password failed: %v", err)
	}
	if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("expected hash output")
	}
}

func TestSecurityValidateEmail(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "validate", "--name", "email", "--value", "test@example.com")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !strings.Contains(output, "Validation passed") {
		t.Fatalf("expected validation passed, got %q", output)
	}
}

func TestSecurityValidateName(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "security", "validate", "--name", "name", "--value", "ab")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !strings.Contains(output, "Validation failed") {
		t.Fatalf("expected validation failed for short name, got %q", output)
	}
}

func TestSecurityAudit(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "security", "audit", "--input", dir)
	if err != nil {
		t.Fatalf("security audit failed: %v", err)
	}
	if !strings.Contains(output, "Security Audit") {
		t.Fatalf("expected audit header, got %q", output)
	}
}

func TestSecurityAuditWithSecrets(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "config.go", "package main\n\nvar password = \"secret123\"\n")

	root := NewRootCommand()
	output, err := executeCommand(root, "security", "audit", "--input", dir)
	if err != nil {
		t.Fatalf("security audit failed: %v", err)
	}
	if !strings.Contains(output, "CRITICAL") {
		t.Fatalf("expected critical finding for hardcoded secret, got %q", output)
	}
}
