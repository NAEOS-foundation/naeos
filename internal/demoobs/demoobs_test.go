package demoobs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/investordemo"
	"github.com/NAEOS-foundation/naeos/internal/observability"
)

// waitFor polls fn until it returns true or times out.
func waitFor(t *testing.T, timeout time.Duration, fn func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestToAuditEvent(t *testing.T) {
	e := investigoEvent(t, "agent-a", "AUTHORIZATION_DENIED", "credential.rotate", "BLOCK", "CAPABILITY_NOT_GRANTED")
	got := toAuditEvent(e)

	if got.ID != e.EventID {
		t.Errorf("ID: got %q, want %q", got.ID, e.EventID)
	}
	if got.UserID != "agent-a" {
		t.Errorf("UserID: got %q, want agent-a", got.UserID)
	}
	if got.Action != "AUTHORIZATION_DENIED" {
		t.Errorf("Action: got %q", got.Action)
	}
	if got.Resource != "credential.rotate" {
		t.Errorf("Resource: got %q", got.Resource)
	}
	if got.Status != "BLOCK" {
		t.Errorf("Status: got %q, want BLOCK", got.Status)
	}
	if got.Details != "CAPABILITY_NOT_GRANTED" {
		t.Errorf("Details: got %q", got.Details)
	}
	if got.Metadata["agent_id"] != "agent-a" || got.Metadata["policy_id"] != "POLICY-017" {
		t.Errorf("Metadata: got %v", got.Metadata)
	}
	if got.Metadata["detail_target"] != "prod" {
		t.Errorf("detail_* metadata: want detail_target=prod, got %v", got.Metadata)
	}
}

func TestSIEMForwarder_CEF(t *testing.T) {
	var (
		mu     sync.Mutex
		got    string
		tenant string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		mu.Lock()
		got = buf.String()
		tenant = r.Header.Get("X-Tenant-ID")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	f := NewSIEMForwarder(srv.URL, observability.SIEMCEF, WithTenant("acme"))
	defer f.Close()

	e := investigoEvent(t, "agent-b", "AUTHORIZATION_GRANTED", "repository.write", "ALLOW", "ok")
	f.OnRecorded(e)

	if !waitFor(t, 2*time.Second, func() bool { mu.Lock(); defer mu.Unlock(); return got != "" }) {
		t.Fatal("no event exported within timeout")
	}
	mu.Lock()
	defer mu.Unlock()
	if !strings.HasPrefix(got, "CEF:0|NAEOS|naeos-investor-demo|1.0|") {
		t.Errorf("unexpected CEF frame: %s", got)
	}
	if !strings.Contains(got, "AUTHORIZATION_GRANTED") {
		t.Errorf("CEF frame missing action: %s", got)
	}
	if !strings.Contains(got, e.EventID) {
		t.Errorf("CEF frame missing event id %q: %s", e.EventID, got)
	}
	if tenant != "acme" {
		t.Errorf("tenant header: got %q, want acme", tenant)
	}
}

func TestSIEMForwarder_JSON(t *testing.T) {
	var (
		mu  sync.Mutex
		got string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		mu.Lock()
		got = buf.String()
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	f := NewSIEMForwarder(srv.URL, observability.SIEMJson)
	defer f.Close()

	e := investigoEvent(t, "agent-c", "AUTHORIZATION_DENIED", "policy.modify", "BLOCK", "POLICY_MUTATION_DENIED")
	f.OnRecorded(e)

	ok := waitFor(t, 2*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return strings.Contains(got, `"action":"AUTHORIZATION_DENIED"`)
	})
	mu.Lock()
	defer mu.Unlock()
	if !ok {
		t.Fatalf("expected NDJSON event, got: %s", got)
	}
}

func TestTracingMiddleware(t *testing.T) {
	tracer := NewDemoTracer()
	exp := &fakeExporter{}
	mw := TracingMiddleware(tracer, exp, nil)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("no"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/scenario", strings.NewReader("{}"))
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("X-Tenant-ID", "t-9")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want 403", rec.Code)
	}

	spans := tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	s := spans[0]
	if s.Name != "POST /api/scenario" {
		t.Errorf("span name: got %q", s.Name)
	}
	if s.Attributes["http.request_id"] != "req-123" {
		t.Errorf("request_id attribute: got %v", s.Attributes["http.request_id"])
	}
	if s.Attributes["tenant_id"] != "t-9" {
		t.Errorf("tenant attribute: got %v", s.Attributes["tenant_id"])
	}
	if s.Status.Code != observability.SpanStatusError {
		t.Errorf("status: got %v, want error", s.Status.Code)
	}

	if !waitFor(t, 2*time.Second, func() bool { return exp.Count() == 1 }) {
		t.Errorf("expected 1 exported span, got %d", exp.Count())
	}
}

// investigoEvent builds a control-plane audit event for tests.
func investigoEvent(t *testing.T, agent, eventType, capability, decision, reason string) *investordemo.AuditEvent {
	t.Helper()
	return &investordemo.AuditEvent{
		EventID:             "AUD-00042",
		Timestamp:           time.Now(),
		EventType:           eventType,
		AgentID:             agent,
		RequestedCapability: investordemo.Capability(capability),
		PolicyID:            "POLICY-017",
		PolicyVersion:       17,
		Decision:            decision,
		Reason:              reason,
		Details:             map[string]any{"target": "prod"},
	}
}

type fakeExporter struct {
	mu    sync.Mutex
	spans int
}

func (e *fakeExporter) ExportSpans(spans []*observability.Span) error {
	e.mu.Lock()
	e.spans += len(spans)
	e.mu.Unlock()
	return nil
}
func (e *fakeExporter) ExportMetrics([]*observability.Metric) error { return nil }
func (e *fakeExporter) ExportLogs([]observability.LogEntry) error   { return nil }

func (e *fakeExporter) Count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.spans
}
