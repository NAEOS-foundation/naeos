// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLedgerPersistence(t *testing.T) {
	ledger := NewLedger()
	ledger.Append(LedgerEvent{
		ID: "EVT-1", Timestamp: time.Now().UTC(), AgentID: "agent-1",
		EventType: "AUTHORIZATION_DECISION", DecisionID: "DEC-1",
		Decision: DecisionAllow,
	})
	path := filepath.Join(t.TempDir(), "evidence", "ledger.json")
	if err := ledger.Save(path); err != nil {
		t.Fatal(err)
	}
	restored, err := LoadLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.Events()) != 1 || restored.Events()[0].DecisionID != "DEC-1" {
		t.Fatalf("unexpected restored ledger: %+v", restored.Events())
	}
}

func TestLedgerAppendDoesNotCommitWhenPersistenceFails(t *testing.T) {
	ledger := NewLedger()
	root := t.TempDir()
	blocker := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("block"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(blocker, "ledger.json")
	ledger.SetPersistencePath(path)
	if event := ledger.Append(LedgerEvent{AgentID: "agent-1", EventType: "TEST"}); event.ID != "" {
		t.Fatalf("expected failed append to return no committed event, got %+v", event)
	}
	if got := ledger.Events(); len(got) != 0 {
		t.Fatalf("expected failed append to remain invisible, got %+v", got)
	}
	if ledger.PersistenceError() == nil {
		t.Fatal("expected persistence error to be recorded")
	}
}

func TestLedgerPersistenceRejectsTampering(t *testing.T) {
	ledger := NewLedger()
	ledger.Append(LedgerEvent{Timestamp: time.Now().UTC(), AgentID: "agent-1", EventType: "TEST"})
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := ledger.Save(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-2] ^= 1
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadLedger(path); err == nil {
		t.Fatal("expected tampered ledger to be rejected")
	}
}

func TestLedgerMetadataIsDefensivelyCopied(t *testing.T) {
	metadata := map[string]string{"policy_id": "policy-a", "policy_version": "1"}
	ledger := NewLedger()
	returned := ledger.Append(LedgerEvent{
		Timestamp: time.Now().UTC(), AgentID: "agent-1",
		EventType: "AUTHORIZATION_DECISION", DecisionID: "DEC-1",
		Decision: DecisionAllow, Metadata: metadata,
	})

	metadata["policy_id"] = "policy-mutated-input"
	returned.Metadata["policy_version"] = "999"

	canonical, ok := ledger.Decision("DEC-1")
	if !ok {
		t.Fatal("expected canonical decision")
	}
	if canonical.Metadata["policy_id"] != "policy-a" || canonical.Metadata["policy_version"] != "1" {
		t.Fatalf("ledger metadata changed through input or Append result: %#v", canonical.Metadata)
	}

	events := ledger.Events()
	events[0].Metadata["policy_id"] = "mutated-events"
	agentEvents := ledger.EventsForAgent("agent-1")
	agentEvents[0].Metadata["policy_id"] = "mutated-agent-events"
	queryEvents := ledger.Query(map[string]string{"decision_id": "DEC-1"})
	queryEvents[0].Metadata["policy_id"] = "mutated-query"

	canonical, ok = ledger.Decision("DEC-1")
	if !ok || canonical.Metadata["policy_id"] != "policy-a" {
		t.Fatalf("ledger metadata changed through snapshot accessors: %#v", canonical.Metadata)
	}
}
