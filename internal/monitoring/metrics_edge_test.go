package monitoring

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func metricValue(reg *Registry, name string, labels map[string]string) (float64, bool) {
	family, ok := reg.GetFamily(name)
	if !ok {
		return 0, false
	}
	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			return m.Value, true
		}
	}
	return 0, false
}

func TestGaugeSetUnknownFamily(t *testing.T) {
	reg := NewRegistry()
	reg.GaugeSet("does_not_exist", 1, nil)
}

func TestGaugeSetExistingMetric(t *testing.T) {
	reg := NewRegistry()
	reg.Register("gauge_test", Gauge, "Test gauge")
	reg.GaugeSet("gauge_test", 5, map[string]string{"env": "prod"})
	reg.GaugeSet("gauge_test", 10, map[string]string{"env": "prod"})

	family, _ := reg.GetFamily("gauge_test")
	if len(family.Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(family.Metrics))
	}
	if family.Metrics[0].Value != 10 {
		t.Errorf("expected value 10, got %v", family.Metrics[0].Value)
	}
}

func TestGaugeSetNewMetric(t *testing.T) {
	reg := NewRegistry()
	reg.Register("gauge_test", Gauge, "Test gauge")
	reg.GaugeSet("gauge_test", 7, nil)

	family, _ := reg.GetFamily("gauge_test")
	if len(family.Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(family.Metrics))
	}
	if family.Metrics[0].Value != 7 {
		t.Errorf("expected value 7, got %v", family.Metrics[0].Value)
	}
	if family.Metrics[0].Type != Gauge {
		t.Errorf("expected type gauge, got %v", family.Metrics[0].Type)
	}
}

func TestGaugeSetCardinalityLimit(t *testing.T) {
	reg := NewRegistryWithOptions(1, 5*time.Minute)
	reg.Register("gauge_limited", Gauge, "Limited gauge")
	reg.GaugeSet("gauge_limited", 1, map[string]string{"a": "1"})
	reg.GaugeSet("gauge_limited", 2, map[string]string{"b": "2"})

	family, _ := reg.GetFamily("gauge_limited")
	if len(family.Metrics) != 1 {
		t.Errorf("expected cardinality-limited 1 metric, got %d", len(family.Metrics))
	}
}

func TestStatusResponseWriterWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := newStatusResponseWriter(rec)

	sw.WriteHeader(http.StatusTeapot)
	if sw.statusCode != http.StatusTeapot {
		t.Errorf("expected status 418, got %d", sw.statusCode)
	}
	if !sw.written {
		t.Error("expected written flag set")
	}

	// Second WriteHeader must not overwrite the captured code.
	rec.Code = http.StatusOK
	sw.WriteHeader(http.StatusNotFound)
	if sw.statusCode != http.StatusTeapot {
		t.Errorf("expected captured status 418, got %d", sw.statusCode)
	}
}

func TestStatusResponseWriterWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := newStatusResponseWriter(rec)

	n, err := sw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if !sw.written {
		t.Error("expected written flag set after Write")
	}
	if sw.statusCode != http.StatusOK {
		t.Errorf("expected default status 200, got %d", sw.statusCode)
	}
}

func TestStatusResponseWriterUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := newStatusResponseWriter(rec)
	if sw.Unwrap() != rec {
		t.Error("expected Unwrap to return the underlying ResponseWriter")
	}
}

func TestRecordRuntime(t *testing.T) {
	m := NewMetrics()
	m.RecordRuntime()

	got := 0
	for _, name := range []string{
		"naeos_go_goroutines",
		"naeos_go_memory_alloc_bytes",
		"naeos_go_memory_sys_bytes",
		"naeos_go_gc_pauses_total",
	} {
		if v, ok := metricValue(m.registry, name, nil); ok {
			if v >= 0 {
				got++
			}
		}
	}
	if got != 4 {
		t.Errorf("expected 4 runtime metrics recorded, got %d", got)
	}
}

func TestStartRuntimeCollector(t *testing.T) {
	m := NewMetrics()
	m.StartRuntimeCollector(20 * time.Millisecond)
	time.Sleep(60 * time.Millisecond)

	if v := m.registry.GaugeValue("naeos_go_goroutines", nil); v <= 0 {
		t.Errorf("expected goroutines metric after collector tick, got %v", v)
	}
}

func TestObservePipelineDuration(t *testing.T) {
	m := NewMetrics()
	m.ObservePipelineDuration(2.5)

	if v, ok := metricValue(m.registry, "naeos_pipeline_duration_seconds", nil); !ok {
		t.Error("expected pipeline duration histogram registered")
	} else if v != 2.5 {
		t.Errorf("expected observed value 2.5, got %v", v)
	}
}

func TestIncSpecValidations(t *testing.T) {
	m := NewMetrics()
	m.IncSpecValidations(true)
	m.IncSpecValidations(false)

	success, ok := metricValue(m.registry, "naeos_spec_validations_total", map[string]string{"status": "success"})
	if !ok || success != 1 {
		t.Errorf("expected 1 success validation, got %v (ok=%v)", success, ok)
	}
	failure, ok := metricValue(m.registry, "naeos_spec_validations_total", map[string]string{"status": "failure"})
	if !ok || failure != 1 {
		t.Errorf("expected 1 failure validation, got %v (ok=%v)", failure, ok)
	}
}

func TestIncArtifacts(t *testing.T) {
	m := NewMetrics()
	m.IncArtifacts()
	m.IncArtifacts()

	if v, ok := metricValue(m.registry, "naeos_artifacts_generated_total", nil); !ok || v != 2 {
		t.Errorf("expected 2 artifacts, got %v (ok=%v)", v, ok)
	}
}

func TestSetWebSocketConnections(t *testing.T) {
	m := NewMetrics()
	m.SetWebSocketConnections(3)

	if v, ok := metricValue(m.registry, "naeos_active_websocket_connections", nil); !ok || v != 3 {
		t.Errorf("expected 3 ws connections, got %v (ok=%v)", v, ok)
	}
}

func TestMetricsObserverLifecycle(t *testing.T) {
	m := NewMetrics()
	o := NewMetricsObserver(m)

	o.OnPipelineStart("p1")
	o.OnPipelineComplete("p1", 2, "1.5s")
	o.OnPipelineComplete("p2", 0, "not-a-duration")
	o.OnPipelineFailed("p3", "boom")
	o.OnArtifactGenerated("a.txt", "/tmp/a.txt")

	success, ok := metricValue(m.registry, "naeos_pipelines_total", map[string]string{"status": "success"})
	if !ok || success != 2 {
		t.Errorf("expected 2 successful pipelines, got %v (ok=%v)", success, ok)
	}
	failure, ok := metricValue(m.registry, "naeos_pipelines_total", map[string]string{"status": "failure"})
	if !ok || failure != 1 {
		t.Errorf("expected 1 failed pipeline, got %v (ok=%v)", failure, ok)
	}
	artifacts, ok := metricValue(m.registry, "naeos_artifacts_generated_total", nil)
	if !ok || artifacts != 3 {
		t.Errorf("expected 3 artifacts (2 complete + 1 direct), got %v (ok=%v)", artifacts, ok)
	}
}

func TestMetricsMiddlewareStatus(t *testing.T) {
	m := NewMetrics()
	mw := MetricsMiddleware(m)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 through middleware, got %d", rec.Code)
	}
	if v, ok := metricValue(m.registry, "naeos_requests_total", map[string]string{"method": "GET", "path": "/missing", "status": "404"}); !ok || v != 1 {
		t.Errorf("expected request recorded with 404, got %v (ok=%v)", v, ok)
	}
}
