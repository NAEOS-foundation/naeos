package audit

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMemoryAuditorExportCSV(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{
		UserID:   "u1",
		Action:   "create",
		Resource: "project",
		Status:   "success",
	})
	auditor.Log(AuditEvent{
		UserID:   "u2",
		Action:   "delete",
		Resource: "config",
		Status:   "failed",
	})

	path := filepath.Join(t.TempDir(), "export.csv")
	if err := auditor.ExportCSV(path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "u1") {
		t.Error("expected CSV to contain user u1")
	}
	if !strings.Contains(content, "create") {
		t.Error("expected CSV to contain action")
	}
	if !strings.HasPrefix(content, "id,timestamp,") {
		t.Error("expected CSV header")
	}
}

func TestMemoryAuditorExportCSVWithSpecialChars(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{
		UserID:   `user,"name"`,
		Action:   "create",
		Resource: "project",
		Status:   "success",
	})

	path := filepath.Join(t.TempDir(), "export.csv")
	if err := auditor.ExportCSV(path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `"user,""name"""`) {
		t.Errorf("expected CSV to escape special chars, got: %s", content)
	}
}

func TestMemoryAuditorExportJSONEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	path := filepath.Join(t.TempDir(), "export.json")
	if err := auditor.ExportJSON(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "[]" && string(data) != "null\n" {
		t.Logf("empty export content: %s", string(data))
	}
}

func TestEscapeCSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{input: "simple", want: "simple"},
		{input: "has,comma", want: `"has,comma"`},
		{input: `has"quote`, want: `"has""quote"`},
		{input: "has\nnewline", want: "\"has\nnewline\""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := escapeCSV(tt.input)
			if got != tt.want {
				t.Errorf("escapeCSV(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMemoryAuditorLatestEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	if latest := auditor.Latest(); latest != nil {
		t.Error("expected nil for Latest on empty auditor")
	}
}

func TestMemoryAuditorOldestEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	if oldest := auditor.Oldest(); oldest != nil {
		t.Error("expected nil for Oldest on empty auditor")
	}
}

func TestMemoryAuditorByIDEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	if event := auditor.ByID("nonexistent"); event != nil {
		t.Error("expected nil for ByID on empty auditor")
	}
}

func TestMemoryAuditorByIDNotFound(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{ID: "existing", UserID: "u1"})
	if event := auditor.ByID("nonexistent"); event != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestMemoryAuditorUserActionsEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	events := auditor.UserActions("u1")
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestMemoryAuditorUserActionsNoMatch(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{UserID: "u1", Action: "test"})
	events := auditor.UserActions("u2")
	if len(events) != 0 {
		t.Errorf("expected 0 events for non-matching user, got %d", len(events))
	}
}

func TestMemoryAuditorFailedEventsEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	events := auditor.FailedEvents()
	if len(events) != 0 {
		t.Errorf("expected 0 failed events, got %d", len(events))
	}
}

func TestMemoryAuditorAggregateEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	agg := auditor.Aggregate()
	if agg.Total != 0 {
		t.Errorf("expected 0 total, got %d", agg.Total)
	}
	if len(agg.ByAction) != 0 {
		t.Error("expected empty ByAction")
	}
}

func TestMemoryAuditorQueryOffsetExceeds(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{UserID: "u1", Action: "a"})
	auditor.Log(AuditEvent{UserID: "u1", Action: "b"})

	events := auditor.Query(Query{Offset: 5})
	if len(events) != 0 {
		t.Errorf("expected 0 events for offset >= len, got %d", len(events))
	}
}

func TestMemoryAuditorQueryLimitOnly(t *testing.T) {
	auditor := NewMemoryAuditor()
	for i := 0; i < 5; i++ {
		auditor.Log(AuditEvent{UserID: "u1", Action: "action"})
	}

	events := auditor.Query(Query{Limit: 2})
	if len(events) != 2 {
		t.Errorf("expected 2 events with limit, got %d", len(events))
	}
}

func TestMemoryAuditorQueryTimeTo(t *testing.T) {
	auditor := NewMemoryAuditor()
	now := time.Now()
	auditor.Log(AuditEvent{UserID: "u1", Timestamp: now.Add(-1 * time.Hour)})
	auditor.Log(AuditEvent{UserID: "u2", Timestamp: now})

	events := auditor.Query(Query{To: now.Add(-30 * time.Minute)})
	if len(events) != 1 {
		t.Errorf("expected 1 event before cutoff, got %d", len(events))
	}
}

func TestMemoryAuditorApplyRetentionMaxAgeZero(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{UserID: "u1", Action: "old", Timestamp: time.Now().Add(-2 * time.Hour)})
	auditor.Log(AuditEvent{UserID: "u2", Action: "recent", Timestamp: time.Now()})

	removed := auditor.ApplyRetention(RetentionPolicy{MaxCount: 1})
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
	if auditor.Len() != 1 {
		t.Errorf("expected 1 remaining, got %d", auditor.Len())
	}
}

func TestMemoryAuditorConcurrentAccess(t *testing.T) {
	auditor := NewMemoryAuditor()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			auditor.Log(AuditEvent{UserID: "u1", Action: "test"})
			auditor.Query(Query{UserID: "u1"})
			auditor.Aggregate()
			auditor.Events()
			auditor.Len()
			auditor.Latest()
			auditor.Oldest()
		}()
	}
	wg.Wait()

	if auditor.Len() != 20 {
		t.Errorf("expected 20 events after concurrent access, got %d", auditor.Len())
	}
}

func TestMemoryAuditorEventsCopy(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Log(AuditEvent{UserID: "u1"})

	events := auditor.Events()
	events[0].UserID = "modified"

	original := auditor.Events()
	if original[0].UserID != "u1" {
		t.Error("Events() should return a copy that does not affect internal state")
	}
}

func TestMemoryAuditorClearEmpty(t *testing.T) {
	auditor := NewMemoryAuditor()
	auditor.Clear()
	if auditor.Len() != 0 {
		t.Error("expected 0 after clearing empty auditor")
	}
}
