// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/database"
)

func TestLoadReplayStoreDefaultsToInMemory(t *testing.T) {
	store, closeStore, err := loadReplayStore("")
	if err != nil {
		t.Fatalf("load in-memory replay store: %v", err)
	}
	defer closeStore()

	if store == nil {
		t.Fatal("expected replay store")
	}
}

func TestLoadReplayStoreUsesConfiguredDurableDatabase(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dbPath := filepath.Join(home, "replay.db")
	configStore := database.NewConnectionStore()
	if err := configStore.Add("replay-test", "sqlite", &database.Config{
		Database: dbPath,
	}); err != nil {
		t.Fatalf("save database connection: %v", err)
	}

	store, closeStore, err := loadReplayStore("replay-test")
	if err != nil {
		t.Fatalf("load durable replay store: %v", err)
	}
	defer closeStore()

	claimed, err := store.Claim("inv-cli-durable")
	if err != nil {
		t.Fatalf("claim first invocation: %v", err)
	}
	if !claimed {
		t.Fatal("expected first invocation claim to succeed")
	}

	// Reopen through the same persisted connection configuration to prove the
	// invocation identity survives process-local store recreation.
	store2, closeStore2, err := loadReplayStore("replay-test")
	if err != nil {
		t.Fatalf("reopen durable replay store: %v", err)
	}
	defer closeStore2()

	claimed, err = store2.Claim("inv-cli-durable")
	if err != nil {
		t.Fatalf("claim replayed invocation: %v", err)
	}
	if claimed {
		t.Fatal("expected replayed invocation to be denied")
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected durable database file: %v", err)
	}
}

func TestAgentRequestCommandExposesReferenceAdapterAndDurableReplay(t *testing.T) {
	cmd := newAgentRequestCommand()

	adapter, err := cmd.Flags().GetString("adapter")
	if err != nil {
		t.Fatalf("read adapter flag: %v", err)
	}
	if adapter != "json" {
		t.Fatalf("expected default adapter json, got %q", adapter)
	}

	if cmd.Flags().Lookup("replay-db") == nil {
		t.Fatal("expected --replay-db flag")
	}
	if cmd.Flags().Lookup("invocation-id") == nil {
		t.Fatal("expected --invocation-id flag")
	}
}
