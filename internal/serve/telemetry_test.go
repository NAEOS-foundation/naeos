package serve

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/observability"
)

func TestTelemetryMiddlewareRecordsCorrelation(t *testing.T) {
	tel := &telemetry{
		tracer: observability.NewTracer("test"),
		corr:   observability.NewCorrelationStore(),
	}

	handler := tel.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("X-Request-ID", "req-trace-test")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	spans := tel.tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name != "http.request" {
		t.Errorf("expected http.request span, got %s", spans[0].Name)
	}

	entry, ok := tel.corr.ByRequestID("req-trace-test")
	if !ok {
		t.Fatal("expected correlation entry")
	}
	if entry.RequestID != "req-trace-test" || entry.TenantID != "tenant-1" {
		t.Errorf("unexpected correlation entry %+v", entry)
	}
	if entry.TraceID != spans[0].TraceID {
		t.Errorf("expected trace id correlation %s, got %s", spans[0].TraceID, entry.TraceID)
	}
}

func TestTelemetryMiddlewareNoHeaders(t *testing.T) {
	tel := &telemetry{
		tracer: observability.NewTracer("test"),
		corr:   observability.NewCorrelationStore(),
	}

	handler := tel.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	spans := tel.tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if tel.corr.Len() != 0 {
		t.Errorf("expected no correlation entries without request id, got %d", tel.corr.Len())
	}
}

func TestObservabilityConfigIsEnabled(t *testing.T) {
	var cfg Observability
	if cfg.IsEnabled() {
		t.Error("empty observability config should be disabled")
	}

	cfg = Observability{OTLPEndpoint: "http://localhost:4318"}
	if !cfg.IsEnabled() {
		t.Error("expected enabled with OTLP endpoint")
	}
}

func TestServerNewWithOTLP(t *testing.T) {
	addr := freeAddr(t)
	cfg := DefaultConfig()
	cfg.Listeners = []Listener{{Addr: addr, Name: "api", API: true}}
	cfg.Observability = Observability{OTLPEndpoint: "http://localhost:4318"}

	srv, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if srv.tel == nil {
		t.Fatal("expected telemetry to be initialized")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- srv.StartWithContext(ctx)
	}()

	// Poll until the server accepts connections.
	for start := time.Now(); time.Since(start) < 4*time.Second; {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	cancel()
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop after cancel")
	case err := <-done:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	}

	// A failed OTLP export must not crash shutdown.
	if len(srv.tel.tracer.GetSpans()) == 0 {
		t.Log("no request spans recorded")
	}
}
