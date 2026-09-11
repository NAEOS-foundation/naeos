package controlplane

import (
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
