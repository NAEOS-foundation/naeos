package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCorrelationStoreRecordAndLookup(t *testing.T) {
	cs := NewCorrelationStore()

	cs.Record("req-001", "trace-001", "tenant-a")

	entry, ok := cs.ByRequestID("req-001")
	if !ok {
		t.Fatal("expected entry for req-001")
	}
	if entry.TraceID != "trace-001" {
		t.Errorf("expected trace-001, got %s", entry.TraceID)
	}
	if entry.TenantID != "tenant-a" {
		t.Errorf("expected tenant-a, got %s", entry.TenantID)
	}

	tid, ok := cs.TraceIDForRequest("req-001")
	if !ok || tid != "trace-001" {
		t.Fatalf("expected trace-001 via TraceIDForRequest, got %q ok=%v", tid, ok)
	}
}

func TestCorrelationStoreByTenant(t *testing.T) {
	cs := NewCorrelationStore()
	cs.Record("req-001", "trace-001", "tenant-a")
	cs.Record("req-002", "trace-002", "tenant-a")
	cs.Record("req-003", "trace-003", "tenant-b")

	traces := cs.TraceIDsByTenant("tenant-a")
	if len(traces) != 2 {
		t.Errorf("expected 2 traces for tenant-a, got %d", len(traces))
	}

	if cs.Len() != 3 {
		t.Errorf("expected store length 3, got %d", cs.Len())
	}
}

func TestCorrelationStoreMissing(t *testing.T) {
	cs := NewCorrelationStore()
	if _, ok := cs.ByRequestID("missing"); ok {
		t.Error("expected miss for missing request")
	}
	if _, ok := cs.TraceIDForRequest("missing"); ok {
		t.Error("expected miss for missing request id")
	}
	if len(cs.TraceIDsByTenant("nobody")) != 0 {
		t.Error("expected no traces for unknown tenant")
	}
}

func TestCorrelationMiddleware(t *testing.T) {
	var gotReqID, gotTenantID string
	handler := CorrelationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReqID = RequestIDFromContext(r)
		gotTenantID = TenantIDFromContext(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("X-Request-ID", "req-999")
	req.Header.Set("X-Tenant-ID", "tenant-zz")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotReqID != "req-999" {
		t.Errorf("expected req-999, got %q", gotReqID)
	}
	if gotTenantID != "tenant-zz" {
		t.Errorf("expected tenant-zz, got %q", gotTenantID)
	}
}

func TestCorrelationMiddlewareMissingHeaders(t *testing.T) {
	handler := CorrelationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r) != "" {
			t.Error("expected empty request id")
		}
		if TenantIDFromContext(r) != "" {
			t.Error("expected empty tenant id")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestLinkFromCorrelation(t *testing.T) {
	cs := NewCorrelationStore()
	carrier := LinkFromCorrelation(cs, "req-abc", "tenant-1", "trace-abc")

	if carrier.RequestID != "req-abc" {
		t.Errorf("unexpected request id %s", carrier.RequestID)
	}
	if carrier.TraceID != "trace-abc" {
		t.Errorf("unexpected trace id %s", carrier.TraceID)
	}

	if entry, ok := cs.ByRequestID("req-abc"); !ok || entry.TraceID != "trace-abc" {
		t.Error("expected correlation recorded")
	}
}
