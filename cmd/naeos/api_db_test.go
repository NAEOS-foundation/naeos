// Copyright 2024-2026 NAEOS Foundation
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
