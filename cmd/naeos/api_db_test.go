// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/api"
)

// TestAttachDatabaseNoEnv verifies attachDatabase is a no-op without
// NAEOS_DB_DRIVER/NAEOS_DB_DATABASE set.
func TestAttachDatabaseNoEnv(t *testing.T) {
	t.Setenv("NAEOS_DB_DRIVER", "")
	t.Setenv("NAEOS_DB_DATABASE", "")

	server := api.NewServer(":0", &api.AuthConfig{})
	if err := attachDatabase(server); err != nil {
		t.Fatalf("attachDatabase without env should be a no-op, got %v", err)
	}
}

// TestAttachDatabaseSQLite verifies attachDatabase connects to a SQLite file
// and initializes the pipeline_runs schema.
func TestAttachDatabaseSQLite(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("NAEOS_DB_DRIVER", "sqlite")
	t.Setenv("NAEOS_DB_DATABASE", dbFile)

	server := api.NewServer(":0", &api.AuthConfig{})
	if err := attachDatabase(server); err != nil {
		t.Fatalf("attachDatabase(sqlite) failed: %v", err)
	}

	if _, err := os.Stat(dbFile); err != nil {
		t.Fatalf("expected sqlite file to be created, got error: %v", err)
	}
}

// TestAttachDatabaseUnknownDriver verifies attachDatabase rejects unknown drivers.
func TestAttachDatabaseUnknownDriver(t *testing.T) {
	t.Setenv("NAEOS_DB_DRIVER", "does-not-exist")
	t.Setenv("NAEOS_DB_DATABASE", ":memory:")

	server := api.NewServer(":0", &api.AuthConfig{})
	if err := attachDatabase(server); err == nil {
		t.Fatal("attachDatabase should return an error for unknown driver")
	}
}

func TestAPIPostgresDatabaseConfigFromEnv(t *testing.T) {
	t.Setenv("NAEOS_DB_DRIVER", "postgres")
	t.Setenv("NAEOS_DB_DATABASE", "naeos")
	t.Setenv("NAEOS_DB_HOST", "db.internal")
	t.Setenv("NAEOS_DB_PORT", "")
	t.Setenv("NAEOS_DB_USER", "naeos")
	t.Setenv("NAEOS_DB_PASSWORD", "test-password")
	t.Setenv("NAEOS_DB_PASSWORD_FILE", "")
	t.Setenv("NAEOS_DB_SSLMODE", "")

	driver, config, err := apiDatabaseConfigFromEnv()
	if err != nil {
		t.Fatalf("apiDatabaseConfigFromEnv() error: %v", err)
	}
	if driver != "postgres" {
		t.Fatalf("expected postgres driver, got %q", driver)
	}
	if config.Host != "db.internal" || config.Port != 5432 || config.User != "naeos" || config.Database != "naeos" || config.SSLMode != "verify-full" {
		t.Fatalf("unexpected postgres config: %+v", config)
	}
}

func TestAPIPostgresDatabaseConfigRejectsInvalidPort(t *testing.T) {
	t.Setenv("NAEOS_DB_DRIVER", "postgres")
	t.Setenv("NAEOS_DB_DATABASE", "naeos")
	t.Setenv("NAEOS_DB_HOST", "db.internal")
	t.Setenv("NAEOS_DB_PORT", "not-a-port")
	t.Setenv("NAEOS_DB_USER", "naeos")
	t.Setenv("NAEOS_DB_PASSWORD", "test-password")
	t.Setenv("NAEOS_DB_PASSWORD_FILE", "")
	t.Setenv("NAEOS_DB_SSLMODE", "verify-full")

	if _, _, err := apiDatabaseConfigFromEnv(); err == nil {
		t.Fatal("expected invalid database port to fail")
	}
}

func TestAPIPostgresDatabasePasswordFromSecretFile(t *testing.T) {
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte("file-password\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NAEOS_DB_DRIVER", "postgres")
	t.Setenv("NAEOS_DB_DATABASE", "naeos")
	t.Setenv("NAEOS_DB_HOST", "db.internal")
	t.Setenv("NAEOS_DB_PORT", "")
	t.Setenv("NAEOS_DB_USER", "naeos")
	t.Setenv("NAEOS_DB_PASSWORD", "")
	t.Setenv("NAEOS_DB_PASSWORD_FILE", passwordFile)
	t.Setenv("NAEOS_DB_SSLMODE", "verify-full")

	_, config, err := apiDatabaseConfigFromEnv()
	if err != nil {
		t.Fatalf("apiDatabaseConfigFromEnv() error: %v", err)
	}
	if config.Password != "file-password" {
		t.Fatal("expected DB password to be read from the secret file")
	}
}
