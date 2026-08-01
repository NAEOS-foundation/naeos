package main

import (
	"strings"
	"testing"
)

func TestAuthWhoamiNoKey(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "whoami")
	if err != nil {
		t.Fatalf("auth whoami failed: %v", err)
	}
	if !strings.Contains(output, "No API key provided") {
		t.Fatalf("expected 'No API key provided', got %q", output)
	}
}

func TestAuthWhoamiInvalidKey(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "whoami", "--api-key", "not-a-real-key")
	if err != nil {
		t.Fatalf("auth whoami failed: %v", err)
	}
	if !strings.Contains(output, "Authentication failed") {
		t.Fatalf("expected 'Authentication failed', got %q", output)
	}
}

func TestAuthCreateUser(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-user", "--name", "john", "--email", "john@example.com", "--role", "admin")
	if err != nil {
		t.Fatalf("auth create-user failed: %v", err)
	}
	if !strings.Contains(output, "Created user: john") {
		t.Fatalf("expected 'Created user: john', got %q", output)
	}
	if !strings.Contains(output, "Roles: admin") {
		t.Fatalf("expected 'Roles: admin', got %q", output)
	}
}

func TestAuthCreateUserMultipleRoles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-user", "--name", "jane", "--role", "admin", "--role", "viewer")
	if err != nil {
		t.Fatalf("auth create-user failed: %v", err)
	}
	if !strings.Contains(output, "Roles: admin, viewer") {
		t.Fatalf("expected 'Roles: admin, viewer', got %q", output)
	}
}

func TestAuthCreateUserMissingName(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "create-user", "--email", "x@example.com")
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestAuthListUsersEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "list-users")
	if err != nil {
		t.Fatalf("auth list-users failed: %v", err)
	}
	if !strings.Contains(output, "No users found") {
		t.Fatalf("expected 'No users found', got %q", output)
	}
}

func TestAuthListUsersAfterCreate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	if _, err := executeCommand(root, "auth", "create-user", "--name", "john", "--email", "john@example.com"); err != nil {
		t.Fatalf("auth create-user failed: %v", err)
	}

	output, err := executeCommand(root, "auth", "list-users")
	if err != nil {
		t.Fatalf("auth list-users failed: %v", err)
	}
	if !strings.Contains(output, "john") {
		t.Fatalf("expected listed user 'john', got %q", output)
	}
}

func TestAuthListRolesEmpty(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "list-roles")
	if err != nil {
		t.Fatalf("auth list-roles failed: %v", err)
	}
	if !strings.Contains(output, "No roles defined") {
		t.Fatalf("expected 'No roles defined', got %q", output)
	}
}

func TestAuthCreateKey(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-key", "--name", "my-key", "--user-id", "u1")
	if err != nil {
		t.Fatalf("auth create-key failed: %v", err)
	}
	if !strings.Contains(output, "Created API key:") {
		t.Fatalf("expected 'Created API key:', got %q", output)
	}
	if !strings.Contains(output, "Name: my-key | User: u1") {
		t.Fatalf("expected key name/user in output, got %q", output)
	}
}

func TestAuthCreateKeyMissingFlags(t *testing.T) {
	root := NewRootCommand()
	if _, err := executeCommand(root, "auth", "create-key"); err == nil {
		t.Fatal("expected error when --name and --user-id are missing")
	}
	_, err := executeCommand(root, "auth", "create-key", "--name", "k1")
	if err == nil {
		t.Fatal("expected error when --user-id is missing")
	}
}

func TestAuthLoginUnknownProvider(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "login", "--provider", "unknown")
	if err != nil {
		t.Fatalf("auth login failed: %v", err)
	}
	if !strings.Contains(output, "not registered") {
		t.Fatalf("expected 'not registered', got %q", output)
	}
}

func TestAuthLogout(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "logout")
	if err != nil {
		t.Fatalf("auth logout failed: %v", err)
	}
	if !strings.Contains(output, "Logged out successfully") {
		t.Fatalf("expected 'Logged out successfully', got %q", output)
	}
}

func TestAuthCreateRole(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-role", "deployer", "--permission", "pipeline:write")
	if err != nil {
		t.Fatalf("auth create-role failed: %v", err)
	}
	if !strings.Contains(output, "Created role: deployer") {
		t.Fatalf("expected 'Created role: deployer', got %q", output)
	}
}

func TestAuthCreateRoleInvalidPermission(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "create-role", "deployer", "--permission", "badformat")
	if err == nil {
		t.Fatal("expected error for invalid permission format")
	}
}

func TestAuthCreateRoleInvalidDeny(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "create-role", "deployer", "--deny", "badformat")
	if err == nil {
		t.Fatal("expected error for invalid deny format")
	}
}

func TestAuthDeleteRoleNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "delete-role", "nonexistent")
	if err == nil {
		t.Fatal("expected error for deleting a nonexistent role")
	}
}

func TestAuthAssignRoleUserNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "assign-role", "ghost-user", "viewer")
	if err == nil {
		t.Fatal("expected error for assigning role to unknown user")
	}
}

func TestAuthCreateRoleFromTemplate(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-role-from-template", "auditor")
	if err != nil {
		t.Fatalf("auth create-role-from-template failed: %v", err)
	}
	if !strings.Contains(output, `Created role "auditor" from template "auditor"`) {
		t.Fatalf("expected template role creation, got %q", output)
	}
}

func TestAuthCreateRoleFromTemplateCustomName(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "create-role-from-template", "soc2_auditor", "--role-name", "soc2-reader")
	if err != nil {
		t.Fatalf("auth create-role-from-template failed: %v", err)
	}
	if !strings.Contains(output, `Created role "soc2-reader" from template "soc2_auditor"`) {
		t.Fatalf("expected custom role name, got %q", output)
	}
}

func TestAuthListRoleTemplates(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "list-role-templates")
	if err != nil {
		t.Fatalf("auth list-role-templates failed: %v", err)
	}
	for _, want := range []string{"auditor", "soc2_auditor", "gdpr_admin", "hipaa_admin"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected template %q in output, got %q", want, output)
		}
	}
}
