package observability

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/audit"
)

func sampleAuditEvent() audit.AuditEvent {
	return audit.AuditEvent{
		ID:           "ae-001",
		Timestamp:    time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		UserID:       "agent-payment-01",
		Action:       "execute",
		Resource:     "production.deploy",
		ResourceID:   "svc/payment",
		IP:           "10.0.0.4",
		Status:       "blocked",
		Details:      "CAPABILITY_NOT_GRANTED",
		Hash:         "abc123",
		PreviousHash: "def456",
		Metadata:     map[string]string{"decision": "deny"},
	}
}

func TestFormatCEF(t *testing.T) {
	line := FormatCEF(sampleAuditEvent(), "naeos")

	if !strings.HasPrefix(line, "CEF:0|NAEOS|naeos|1.0|") {
		t.Fatalf("unexpected CEF header: %s", line)
	}
	if !strings.Contains(line, "execute|production.deploy|8|") {
		t.Fatalf("expected action|resource|severity 8 for blocked action: %s", line)
	}
	if !strings.Contains(line, "suid=agent-payment-01") {
		t.Fatalf("expected suid extension: %s", line)
	}
	if !strings.Contains(line, "outcome=blocked") {
		t.Fatalf("expected outcome extension: %s", line)
	}
	if !strings.Contains(line, "src=10.0.0.4") {
		t.Fatalf("expected src extension: %s", line)
	}
	if !strings.Contains(line, "decision=deny") {
		t.Fatalf("expected metadata extension: %s", line)
	}
}

func TestFormatCEFEscapesPipes(t *testing.T) {
	event := sampleAuditEvent()
	event.Details = "value|with|pipes"
	line := FormatCEF(event, "naeos")

	if strings.Count(line, "value|with|pipes") > 0 {
		t.Fatalf("unescaped pipe in CEF line: %s", line)
	}
	if !strings.Contains(line, `value\|with\|pipes`) {
		t.Fatalf("expected escaped pipes: %s", line)
	}
}

func TestSIEMExportCEF(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		if r.Header.Get("X-Tenant-ID") != "tenant-a" {
			t.Errorf("expected tenant header")
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	exp := NewSIEMExporter(srv.URL, WithSIEMTenant("tenant-a"))
	if err := exp.ExportEvent(sampleAuditEvent()); err != nil {
		t.Fatalf("export error: %v", err)
	}
	if !strings.HasPrefix(received, "CEF:0|") {
		t.Fatalf("expected CEF payload, got %s", received)
	}
}

func TestSIEMExportJSON(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exp := NewSIEMExporter(srv.URL, WithSIEMFormat(SIEMJson))
	if err := exp.ExportEvent(sampleAuditEvent()); err != nil {
		t.Fatalf("export error: %v", err)
	}
	if !strings.HasPrefix(received, `{"id":"ae-001"`) {
		t.Fatalf("expected JSON payload, got %s", received)
	}
}

func TestSIEMExportMultiple(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exp := NewSIEMExporter(srv.URL)
	events := []audit.AuditEvent{sampleAuditEvent(), sampleAuditEvent()}
	if err := exp.ExportEvents(events); err != nil {
		t.Fatalf("export error: %v", err)
	}
	if strings.Count(received, "CEF:0|") != 2 {
		t.Fatalf("expected 2 CEF lines, got %s", received)
	}
}

func TestSIEMExportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	exp := NewSIEMExporter(srv.URL)
	if err := exp.ExportEvent(sampleAuditEvent()); err == nil {
		t.Fatal("expected error for 502 response")
	}
}

func TestSIEMExporterImplements(t *testing.T) {
	var _ interface{ ExportEvent(audit.AuditEvent) error } = (*SIEMExporter)(nil)
}

func TestSiementTimestamps(t *testing.T) {
	event := sampleAuditEvent()
	event.Timestamp = time.Time{}
	line := FormatCEF(event, "naeos")
	if !strings.Contains(line, "rt=0001-01-01T00:00:00Z") {
		t.Fatalf("expected zero timestamp serialized as UTC: %s", line)
	}
}
