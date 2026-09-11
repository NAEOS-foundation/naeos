package observability

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOTLPExportSpans(t *testing.T) {
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/traces" {
			t.Errorf("expected /v1/traces, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		captured = body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exp := NewOTLPHTTPExporter(srv.URL, WithOTLPHeaders(map[string]string{"Authorization": "Bearer test"}))

	tr := NewTracer("test")
	span := tr.StartSpan("http.request")
	span.Attributes["method"] = "GET"
	span.Attributes["status_code"] = 200
	tr.AddEvent(span, "cache.miss", map[string]any{"key": "user/123"})
	tr.SetStatus(span, SpanStatusOK, "ok")
	tr.EndSpan(span)

	if err := exp.ExportSpans(tr.GetSpans()); err != nil {
		t.Fatalf("export error: %v", err)
	}

	if len(captured) == 0 {
		t.Fatal("expected request body")
	}

	var req otlpExportRequest
	if err := json.Unmarshal(captured, &req); err != nil {
		t.Fatalf("unmarshal otlp request: %v", err)
	}

	if len(req.ResourceSpans) != 1 {
		t.Fatalf("expected 1 resource span, got %d", len(req.ResourceSpans))
	}
	spans := req.ResourceSpans[0].ScopeSpans[0].Spans
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	s := spans[0]
	if s.Name != "http.request" {
		t.Errorf("expected name http.request, got %s", s.Name)
	}
	if s.TraceID != span.TraceID {
		t.Errorf("trace id mismatch")
	}
	if s.Status.Code != 1 {
		t.Errorf("expected status OK (1), got %d", s.Status.Code)
	}
	if len(s.Attributes) < 2 {
		t.Errorf("expected 2 attributes, got %d", len(s.Attributes))
	}
	if len(s.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(s.Events))
	}
}

func TestOTLPExportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("collector error"))
	}))
	defer srv.Close()

	exp := NewOTLPHTTPExporter(srv.URL)
	tr := NewTracer("test")
	span := tr.StartSpan("operation")
	tr.EndSpan(span)

	err := exp.ExportSpans(tr.GetSpans())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestOTLPExportNetworkError(t *testing.T) {
	exp := NewOTLPHTTPExporter("http://localhost:1")
	tr := NewTracer("test")
	span := tr.StartSpan("operation")
	tr.EndSpan(span)

	err := exp.ExportSpans(tr.GetSpans())
	if err == nil {
		t.Fatal("expected error for unreachable endpoint")
	}
}

func TestOTLPConvertSpansEmpty(t *testing.T) {
	exp := NewOTLPHTTPExporter("http://localhost:1")
	result := exp.convertSpans(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestOTLPChildSpans(t *testing.T) {
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exp := NewOTLPHTTPExporter(srv.URL)
	tr := NewTracer("test")

	parent := tr.StartSpan("parent")
	child := tr.StartSpanWithParent("child", parent.SpanID)
	child.Attributes["kind"] = "internal"
	tr.EndSpan(child)
	tr.EndSpan(parent)

	_ = exp.ExportSpans(tr.GetSpans())

	var req otlpExportRequest
	_ = json.Unmarshal(captured, &req)

	spans := req.ResourceSpans[0].ScopeSpans[0].Spans
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}

	// one span should have ParentSpanID
	found := false
	for _, s := range spans {
		if s.ParentSpanID != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one child span with parent")
	}
}

func TestOTLPExporterImplementsInterface(t *testing.T) {
	var _ Exporter = (*OTLPHTTPExporter)(nil)
}

func TestOTLPDefaultTimeout(t *testing.T) {
	exp := NewOTLPHTTPExporter("http://localhost:1")
	if exp.client.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", exp.client.Timeout)
	}
}
