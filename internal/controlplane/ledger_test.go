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
