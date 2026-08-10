package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileAuditorLog(t *testing.T) {
	dir := t.TempDir()
	auditor, err := NewFileAuditor(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = auditor.Log(AuditEvent{
		UserID:   "user1",
		Action:   "create",
		Resource: "project",
		Status:   "success",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".naeos", "audit.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty log file")
	}
}

func TestMemoryAuditorQuery(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{UserID: "u1", Action: "create", Resource: "project", Status: "success"})
	auditor.Log(AuditEvent{UserID: "u2", Action: "delete", Resource: "project", Status: "failed"})
	auditor.Log(AuditEvent{UserID: "u1", Action: "update", Resource: "config", Status: "success"})

	events := auditor.Query(Query{UserID: "u1"})
	if len(events) != 2 {
		t.Errorf("expected 2 events for u1, got %d", len(events))
	}

	events = auditor.Query(Query{Action: "delete"})
	if len(events) != 1 {
		t.Errorf("expected 1 delete event, got %d", len(events))
	}

	events = auditor.Query(Query{Resource: "project"})
	if len(events) != 2 {
		t.Errorf("expected 2 project events, got %d", len(events))
	}

	events = auditor.Query(Query{Status: "failed"})
	if len(events) != 1 {
		t.Errorf("expected 1 failed event, got %d", len(events))
	}
}

func TestMemoryAuditorQueryPagination(t *testing.T) {
	auditor := NewMemoryAuditor()

	for i := 0; i < 10; i++ {
		auditor.Log(AuditEvent{UserID: "u1", Action: "action"})
	}

	events := auditor.Query(Query{Limit: 3})
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	events = auditor.Query(Query{Offset: 8, Limit: 3})
	if len(events) != 2 {
		t.Errorf("expected 2 events (offset 8), got %d", len(events))
	}
}

func TestMemoryAuditorQueryTimeRange(t *testing.T) {
	auditor := NewMemoryAuditor()

	past := time.Now().Add(-1 * time.Hour)
	recent := time.Now().Add(-1 * time.Minute)

	auditor.Log(AuditEvent{UserID: "u1", Timestamp: past})
	auditor.Log(AuditEvent{UserID: "u2", Timestamp: recent})

	events := auditor.Query(Query{From: time.Now().Add(-30 * time.Minute)})
	if len(events) != 1 {
		t.Errorf("expected 1 recent event, got %d", len(events))
	}
}

func TestMemoryAuditorAggregate(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{UserID: "u1", Action: "create", Resource: "project", Status: "success"})
	auditor.Log(AuditEvent{UserID: "u1", Action: "create", Resource: "config", Status: "success"})
	auditor.Log(AuditEvent{UserID: "u2", Action: "delete", Resource: "project", Status: "failed"})

	agg := auditor.Aggregate()

	if agg.Total != 3 {
		t.Errorf("expected 3 total, got %d", agg.Total)
	}
	if agg.ByAction["create"] != 2 {
		t.Errorf("expected 2 create, got %d", agg.ByAction["create"])
	}
	if agg.ByUser["u1"] != 2 {
		t.Errorf("expected 2 u1, got %d", agg.ByUser["u1"])
	}
	if agg.ByResource["project"] != 2 {
		t.Errorf("expected 2 project, got %d", agg.ByResource["project"])
	}
	if agg.ByStatus["failed"] != 1 {
		t.Errorf("expected 1 failed, got %d", agg.ByStatus["failed"])
	}
}

func TestMemoryAuditorRetention(t *testing.T) {
	auditor := NewMemoryAuditor()

	for i := 0; i < 10; i++ {
		auditor.Log(AuditEvent{UserID: "u1", Action: "action"})
	}

	removed := auditor.ApplyRetention(RetentionPolicy{MaxCount: 5})
	if removed != 5 {
		t.Errorf("expected 5 removed, got %d", removed)
	}
	if auditor.Len() != 5 {
		t.Errorf("expected 5 remaining, got %d", auditor.Len())
	}
}

func TestMemoryAuditorRetentionMaxAge(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{UserID: "u1", Timestamp: time.Now().Add(-2 * time.Hour)})
	auditor.Log(AuditEvent{UserID: "u2", Timestamp: time.Now()})

	removed := auditor.ApplyRetention(RetentionPolicy{MaxAge: 1 * time.Hour})
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
}

func TestMemoryAuditorExportJSON(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{UserID: "u1", Action: "create"})

	path := filepath.Join(t.TempDir(), "export.json")
	if err := auditor.ExportJSON(path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty export")
	}
}

func TestMemoryAuditorLatestOldest(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{UserID: "u1", Action: "first"})
	auditor.Log(AuditEvent{UserID: "u2", Action: "second"})

	if auditor.Oldest().Action != "first" {
		t.Errorf("expected oldest 'first', got %q", auditor.Oldest().Action)
	}
	if auditor.Latest().Action != "second" {
		t.Errorf("expected latest 'second', got %q", auditor.Latest().Action)
	}
}

func TestMemoryAuditorByID(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{ID: "custom-id", UserID: "u1"})

	event := auditor.ByID("custom-id")
	if event == nil {
		t.Fatal("expected to find event by ID")
	}
	if event.UserID != "u1" {
		t.Errorf("expected UserID u1, got %q", event.UserID)
	}

	if auditor.ByID("nonexistent") != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestMemoryAuditorUserActions(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{UserID: "u1", Action: "a"})
	auditor.Log(AuditEvent{UserID: "u2", Action: "b"})
	auditor.Log(AuditEvent{UserID: "u1", Action: "c"})

	events := auditor.UserActions("u1")
	if len(events) != 2 {
		t.Errorf("expected 2 events for u1, got %d", len(events))
	}
}

func TestMemoryAuditorFailedEvents(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{Status: "success"})
	auditor.Log(AuditEvent{Status: "failed"})
	auditor.Log(AuditEvent{Status: "error"})
	auditor.Log(AuditEvent{Status: "success"})

	failed := auditor.FailedEvents()
	if len(failed) != 2 {
		t.Errorf("expected 2 failed events, got %d", len(failed))
	}
}

func TestMemoryAuditorClear(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{UserID: "u1"})
	auditor.Clear()

	if auditor.Len() != 0 {
		t.Errorf("expected 0 after clear, got %d", auditor.Len())
	}
}

func TestMemoryAuditorEmptyQuery(t *testing.T) {
	auditor := NewMemoryAuditor()

	events := auditor.Query(Query{})
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestAuditEventMetadata(t *testing.T) {
	auditor := NewMemoryAuditor()

	auditor.Log(AuditEvent{
		UserID:   "u1",
		Action:   "create",
		Metadata: map[string]string{"key": "value"},
	})

	events := auditor.Events()
	if events[0].Metadata["key"] != "value" {
		t.Errorf("expected metadata key=value, got %v", events[0].Metadata)
	}
}

func TestMemoryAuditorExportCSV(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{UserID: "u1", Action: "create", Resource: "project", Status: "success"})
	auditor.Log(AuditEvent{UserID: "u2", Action: "delete", Resource: "project", Status: "failed"})

	path := filepath.Join(t.TempDir(), "audit.csv")
	err := auditor.ExportCSV(path)
	if err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "u1") || !strings.Contains(content, "create") {
		t.Errorf("unexpected csv content: %s", content)
	}
	if !strings.HasPrefix(content, "id,timestamp,") {
		t.Errorf("expected csv header, got: %s", content[:20])
	}
}

func TestEscapeCSV(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"has,comma", `"has,comma"`},
		{"has\"quote", `"has""quote"`},
		{"has\nnewline", "\"has\nnewline\""},
		{"", ""},
	}
	for _, tt := range tests {
		got := escapeCSV(tt.input)
		if got != tt.want {
			t.Errorf("escapeCSV(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
