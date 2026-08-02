package main

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/supabase"
)

func TestSupabaseAuthSignoutNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "signout")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseAuthUserNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "user")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseAuthAdminListNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "admin", "list")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseAuthAdminCreateNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "admin", "create", "--email", "a@b.c", "--password", "x")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseAuthAdminDeleteNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "admin", "delete", "--user-id", "u-1")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseStorageListBucketsNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "list-buckets")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseStorageCreateBucketNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "create-bucket", "--name", "bucket-a")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseStorageUploadNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "upload", "--bucket", "b", "--source", "x.txt", "--path", "x.txt")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseStorageDownloadNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "download", "--bucket", "b", "--path", "x.txt", "--dest", "x.txt")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseStorageDeleteNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "delete", "--bucket", "b", "--path", "x.txt")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSupabaseSQLNotConfigured(t *testing.T) {
	supabase.SetConfigDir(t.TempDir())
	defer supabase.SetConfigDir("")

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "sql", "SELECT 1")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}
