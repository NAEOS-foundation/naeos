package main

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/supabase"
)

func TestSupabaseInitProjectRef(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "init",
		"--project-ref", "test-proj",
		"--url", "https://test-proj.supabase.co",
		"--anon-key", "test-anon-key",
		"--service-role-key", "test-svc-key",
	)
	if err != nil {
		t.Fatalf("supabase init failed: %v", err)
	}
	if !strings.Contains(output, "test-proj") {
		t.Fatalf("expected 'test-proj' in output, got %q", output)
	}
}

func TestSupabaseInitFromEnv(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	t.Setenv("SUPABASE_PROJECT_REF", "env-proj")
	t.Setenv("SUPABASE_URL", "https://env-proj.supabase.co")
	t.Setenv("SUPABASE_PUBLISHABLE_KEY", "env-anon-key")
	t.Setenv("SUPABASE_SECRET_KEY", "env-svc-key")
	t.Setenv("SUPABASE_JWKS_URL", "https://env-proj.supabase.co/auth/v1/.well-known/jwks.json")

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "init")
	if err != nil {
		t.Fatalf("supabase init from env failed: %v", err)
	}
	if !strings.Contains(output, "env-proj") {
		t.Fatalf("expected 'env-proj' in output, got %q", output)
	}

	cfg, err := supabase.LoadConfig()
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if cfg.ProjectRef != "env-proj" {
		t.Fatalf("expected project ref 'env-proj', got %q", cfg.ProjectRef)
	}
	if cfg.AnonKey != "env-anon-key" {
		t.Fatalf("expected anon key 'env-anon-key', got %q", cfg.AnonKey)
	}
	if cfg.ServiceRoleKey != "env-svc-key" {
		t.Fatalf("expected service role key 'env-svc-key', got %q", cfg.ServiceRoleKey)
	}
}

func TestSupabaseInitMissingProjectRef(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "init")
	if err == nil {
		t.Fatal("expected error for missing --project-ref")
	}
	if !strings.Contains(err.Error(), "project-ref is required") {
		t.Fatalf("expected 'project-ref is required' error, got %v", err)
	}
}

func TestSupabaseSQLNoConfig(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "sql", "SELECT 1")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseSQLNoArgs(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "sql")
	if err == nil {
		t.Fatal("expected error when no SQL arg")
	}
}

func TestSupabaseStatusNotConfigured(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "status")
	if err != nil {
		t.Fatalf("supabase status should not error when not configured: %v", err)
	}
	if !strings.Contains(output, "Not configured") {
		t.Fatalf("expected 'Not configured' in output, got %q", output)
	}
}

func TestSupabaseStatusConfigured(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	cfg := &supabase.Config{
		ProjectRef:     "test-proj",
		URL:            "https://test-proj.supabase.co",
		AnonKey:        "test-anon-key",
		ServiceRoleKey: "test-svc-key",
	}
	if err := supabase.SaveConfig(cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "status")
	if err != nil {
		t.Fatalf("supabase status failed: %v", err)
	}
	if !strings.Contains(output, "test-proj") {
		t.Fatalf("expected 'test-proj' in output, got %q", output)
	}
	if !strings.Contains(output, "MaskKey") && !strings.Contains(output, "test-anon") && !strings.Contains(output, "Anon key") {
		t.Fatalf("expected anon key info in output, got %q", output)
	}
}

func TestSupabaseStatusPartialConfig(t *testing.T) {
	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	cfg := &supabase.Config{
		ProjectRef:     "test-proj",
		URL:            "https://test-proj.supabase.co",
		ServiceRoleKey: "test-svc-key",
	}
	if err := supabase.SaveConfig(cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "status")
	if err != nil {
		t.Fatalf("supabase status failed: %v", err)
	}
	if !strings.Contains(output, "PARTIAL") {
		t.Fatalf("expected 'PARTIAL' status, got %q", output)
	}
}

func TestSupabaseAuthHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth")
	if err != nil {
		t.Fatalf("supabase auth failed: %v", err)
	}
	if !strings.Contains(output, "Manage authentication") && !strings.Contains(output, "signup") {
		t.Fatalf("expected auth help output, got %q", output)
	}
}

func TestSupabaseStorageHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage")
	if err != nil {
		t.Fatalf("supabase storage failed: %v", err)
	}
	if !strings.Contains(output, "Manage storage") && !strings.Contains(output, "list-buckets") {
		t.Fatalf("expected storage help output, got %q", output)
	}
}

func TestSupabaseAuthSignupMissingFlags(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "signup")
	if err == nil {
		t.Fatal("expected error for missing --email and --password")
	}
}

func TestSupabaseStorageCreateBucketMissingName(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "create-bucket")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestSupabaseStorageUploadMissingSource(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "upload")
	if err == nil {
		t.Fatal("expected error for missing --source")
	}
}

func TestSupabaseStorageDownloadMissingSource(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "download")
	if err == nil {
		t.Fatal("expected error for missing --source")
	}
}

func TestSupabaseStorageDeleteMissingPath(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "delete")
	if err == nil {
		t.Fatal("expected error for missing --path")
	}

	supabase.SetConfigDir(".naeos/supabase")
}
