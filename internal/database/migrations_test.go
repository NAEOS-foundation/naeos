package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMigrations(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "000001_create_users.up.sql"), []byte("CREATE TABLE users (id INT)"), 0644)
	os.WriteFile(filepath.Join(dir, "000001_create_users.down.sql"), []byte("DROP TABLE users"), 0644)
	os.WriteFile(filepath.Join(dir, "000002_add_email.up.sql"), []byte("ALTER TABLE users ADD COLUMN email TEXT"), 0644)
	os.WriteFile(filepath.Join(dir, "000002_add_email.down.sql"), []byte("SELECT 1"), 0644)

	migrations, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}
	if migrations[0].Version != 1 {
		t.Errorf("expected version 1, got %d", migrations[0].Version)
	}
	if migrations[0].Name != "create_users" {
		t.Errorf("expected name 'create_users', got %s", migrations[0].Name)
	}
	if migrations[0].Up != "CREATE TABLE users (id INT)" {
		t.Errorf("unexpected up SQL: %s", migrations[0].Up)
	}
	if migrations[0].Down != "DROP TABLE users" {
		t.Errorf("unexpected down SQL: %s", migrations[0].Down)
	}
	if migrations[1].Version != 2 {
		t.Errorf("expected version 2, got %d", migrations[1].Version)
	}
}

func TestLoadMigrationsEmptyDir(t *testing.T) {
	dir := t.TempDir()
	migrations, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}

func TestLoadMigrationsInvalidDir(t *testing.T) {
	_, err := LoadMigrations("/nonexistent/dir")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestLoadMigrationsSkipsNonSQL(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "000001_create_users.up.sql"), []byte("CREATE TABLE users (id INT)"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Migrations"), 0644)
	os.WriteFile(filepath.Join(dir, "000001_create_users.txt"), []byte("not a migration"), 0644)

	migrations, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 1 {
		t.Errorf("expected 1 migration, got %d", len(migrations))
	}
}

func TestMigrationChecksum(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "init", Up: "CREATE TABLE t(id INT)", Down: "DROP TABLE t"},
		{Version: 2, Name: "add_col", Up: "ALTER TABLE t ADD COLUMN name TEXT", Down: "SELECT 1"},
	}
	checksum := MigrationChecksum(migrations)
	if checksum == "" {
		t.Error("expected non-empty checksum")
	}

	checksum2 := MigrationChecksum(migrations)
	if checksum != checksum2 {
		t.Error("expected same checksum for same migrations")
	}

	different := []Migration{
		{Version: 1, Name: "init", Up: "CREATE TABLE t(id INT)", Down: "DROP TABLE t"},
	}
	if MigrationChecksum(different) == checksum {
		t.Error("expected different checksum for different migrations")
	}
}
