// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"encoding/json"
	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadEvidenceStoreRejectsPersistedHashTampering(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store := evidence.NewStore()
	rec, err := store.Append(evidence.EvidenceRecord{
		RequestID: "manus-p26-001",
		Timestamp: time.Now().UTC(),
		Actor:     "test",
		Action:    "read",
	})
	if err != nil {
		t.Fatalf("append evidence: %v", err)
	}

	rec.RequestID = "manus-p26-tampered"
	path := filepath.Join(home, ".config", "naeos", "evidence.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create evidence directory: %v", err)
	}
	data, err := json.Marshal([]evidence.EvidenceRecord{rec})
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write evidence: %v", err)
	}

	_, err = loadEvidenceStore()
	if err == nil {
		t.Fatal("expected tampered evidence to be rejected")
	}
	if !strings.Contains(err.Error(), "evidence hash mismatch") {
		t.Fatalf("expected hash mismatch error, got %v", err)
	}
}
