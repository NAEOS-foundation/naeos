package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("expected registry to be created")
	}
}

func TestRegistryRegister(t *testing.T) {
	reg := NewRegistry()
	reg.Register("test_counter", Counter, "A test counter")

	families := reg.GetFamilies()
	if len(families) != 1 {
		t.Errorf("expected 1 family, got %d", len(families))
	}
	if families[0].Name != "test_counter" {
		t.Errorf("expected name 'test_counter', got %s", families[0].Name)
	}
}

func TestRegistryCounterInc(t *testing.T) {
	reg := NewRegistry()
	reg.Register("requests", Counter, "Total requests")

	reg.CounterInc("requests", nil)
	reg.CounterInc("requests", nil)

	families := reg.GetFamilies()
	if len(families) != 1 {
		t.Fatalf("expected 1 family, got %d", len(families))
	}

	if families[0].Metrics[0].Value != 2 {
		t.Errorf("expected value 2, got %f", families[0].Metrics[0].Value)
	}
}

func TestRegistryCounterWithLabels(t *testing.T) {
	reg := NewRegistry()
	reg.Register("http_requests", Counter, "HTTP requests")

	labels1 := map[string]string{"method": "GET", "path": "/api"}
	labels2 := map[string]string{"method": "POST", "path": "/api"}

	reg.CounterInc("http_requests", labels1)
	reg.CounterInc("http_requests", labels1)
	reg.CounterInc("http_requests", labels2)

	families := reg.GetFamilies()
	if len(families[0].Metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(families[0].Metrics))
	}
}

func TestRegistryCounterAdd(t *testing.T) {
	reg := NewRegistry()
	reg.Register("bytes", Counter, "Bytes transferred")

	reg.CounterAdd("bytes", 100, map[string]string{"dir": "in"})
	reg.CounterAdd("bytes", 200, map[string]string{"dir": "in"})

	val := reg.CounterValue("bytes", map[string]string{"dir": "in"})
	if val != 300 {
		t.Errorf("expected 300, got %f", val)
	}
}

func TestRegistryGauge(t *testing.T) {
	reg := NewRegistry()
	reg.Register("connections", Gauge, "Active connections")

	reg.GaugeSet("connections", 10, nil)
	reg.GaugeInc("connections", nil)
	reg.GaugeDec("connections", nil)

	families := reg.GetFamilies()
	if families[0].Metrics[0].Value != 10 {
		t.Errorf("expected value 10, got %f", families[0].Metrics[0].Value)
	}
}

func TestRegistryHistogram(t *testing.T) {
	reg := NewRegistry()
	reg.Register("duration", Histogram, "Request duration")

	reg.HistogramObserve("duration", 0.1, nil)
	reg.HistogramObserve("duration", 0.2, nil)

	families := reg.GetFamilies()
	if families[0].Count != 2 {
		t.Errorf("expected count 2, got %d", families[0].Count)
	}
	if diff := families[0].Sum - 0.3; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("expected sum ~0.3, got %f", families[0].Sum)
	}
}

func TestHistogramBucketing(t *testing.T) {
	reg := NewRegistry()
	reg.Register("req_duration", Histogram, "Request duration")

	reg.HistogramObserveWithBuckets("req_duration", 0.01, nil)
	reg.HistogramObserveWithBuckets("req_duration", 0.5, nil)
	reg.HistogramObserveWithBuckets("req_duration", 5.0, nil)

	family, ok := reg.GetFamily("req_duration")
	if !ok {
		t.Fatal("expected to find req_duration")
	}

	if family.Count != 3 {
		t.Errorf("expected count 3, got %d", family.Count)
	}
	if family.Sum != 5.51 {
		t.Errorf("expected sum 5.51, got %f", family.Sum)
	}

	// 0.01 goes in bucket 0.01, 0.5 goes in bucket 0.5, 5.0 goes in bucket 5
	cumCount := uint64(0)
	bucketHit := false
	for i, bound := range family.Buckets {
		cumCount += family.Counts[i]
		if bound == 0.01 && cumCount < 1 {
			t.Errorf("expected at least 1 in bucket <=0.01")
		}
		if bound == 0.5 && cumCount < 2 {
			t.Errorf("expected at least 2 in bucket <=0.5")
		}
		if bound == 5.0 {
			bucketHit = true
			if cumCount < 3 {
				t.Errorf("expected at least 3 in bucket <=5.0")
			}
		}
	}
	if !bucketHit {
		t.Error("expected 5.0 bucket to exist")
	}
}

func TestCustomBuckets(t *testing.T) {
	reg := NewRegistry()
	buckets := []float64{1, 5, 10}
	reg.RegisterWithBuckets("custom", "Custom histogram", buckets)

	family, ok := reg.GetFamily("custom")
	if !ok {
		t.Fatal("expected custom histogram")
	}
	if len(family.Buckets) != 3 {
		t.Errorf("expected 3 buckets, got %d", len(family.Buckets))
	}
	if len(family.Counts) != 4 {
		t.Errorf("expected 4 counts (buckets+1), got %d", len(family.Counts))
	}
}

func TestFormatPrometheus(t *testing.T) {
	reg := NewRegistry()
	reg.Register("test_metric", Counter, "A test metric")
	reg.CounterInc("test_metric", nil)

	output := reg.FormatPrometheus()
	if !strings.Contains(output, "# HELP test_metric") {
		t.Error("expected HELP line")
	}
	if !strings.Contains(output, "# TYPE test_metric counter") {
		t.Error("expected TYPE line")
	}
	if !strings.Contains(output, "test_metric") {
		t.Error("expected metric name")
	}
}

func TestFormatLabels(t *testing.T) {
	labels := map[string]string{"method": "GET", "status": "200"}
	result := formatLabels(labels)

	if !strings.Contains(result, "method=\"GET\"") {
		t.Error("expected method label")
	}
	if !strings.Contains(result, "status=\"200\"") {
		t.Error("expected status label")
	}
}

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()
	if metrics == nil {
		t.Fatal("expected metrics to be created")
	}

	families := metrics.Registry().GetFamilies()
	if len(families) < 5 {
		t.Errorf("expected at least 5 default metrics, got %d", len(families))
	}
}

func TestMetricsIncRequests(t *testing.T) {
	metrics := NewMetrics()
	metrics.IncRequests("GET", "/api/health", "200")

	families := metrics.Registry().GetFamilies()
	for _, f := range families {
		if f.Name == "naeos_requests_total" {
			if len(f.Metrics) == 0 {
				t.Error("expected at least 1 metric")
			}
			return
		}
	}
	t.Error("expected naeos_requests_total metric")
}

func TestMetricsIncPipelines(t *testing.T) {
	metrics := NewMetrics()
	metrics.IncPipelines("success")
	metrics.IncPipelines("failure")

	families := metrics.Registry().GetFamilies()
	for _, f := range families {
		if f.Name == "naeos_pipelines_total" {
			if len(f.Metrics) != 2 {
				t.Errorf("expected 2 metrics, got %d", len(f.Metrics))
			}
			return
		}
	}
	t.Error("expected naeos_pipelines_total metric")
}

func TestHealthHandler(t *testing.T) {
	handler := HealthHandler()

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/health", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "healthy") {
		t.Error("expected 'healthy' in response")
	}
}

func TestReadyHandler(t *testing.T) {
	handler := ReadyHandler()

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/ready", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "ready") {
		t.Error("expected 'ready' in response")
	}
}

func TestPrometheusHandler(t *testing.T) {
	reg := NewRegistry()
	reg.Register("test_metric", Counter, "Test")
	reg.CounterInc("test_metric", nil)

	handler := PrometheusHandler(reg)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "test_metric") {
		t.Error("expected metric in response")
	}
}

func TestCollector(t *testing.T) {
	reg := NewRegistry()
	reg.Register("test", Counter, "Test")
	reg.CounterInc("test", nil)

	collector := NewCollector(reg)
	snapshot := collector.Collect()

	if snapshot == nil {
		t.Fatal("expected snapshot")
	}

	if len(snapshot.Families) != 1 {
		t.Errorf("expected 1 family, got %d", len(snapshot.Families))
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("expected timestamp")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	metrics := NewMetrics()
	middleware := MetricsMiddleware(metrics)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(inner)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMetricsMiddlewareCapturesStatus(t *testing.T) {
	metrics := NewMetrics()
	middleware := MetricsMiddleware(metrics)

	tests := []struct {
		name       string
		statusCode int
	}{
		{"404", http.StatusNotFound},
		{"500", http.StatusInternalServerError},
		{"201", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})
			handler := middleware(inner)
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}

			families := metrics.Registry().GetFamilies()
			for _, f := range families {
				if f.Name == "naeos_requests_total" {
					found := false
					for _, m := range f.Metrics {
						if m.Labels["status"] == tt.name {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected status label %q in metrics", tt.name)
					}
					return
				}
			}
			t.Error("expected naeos_requests_total metric")
		})
	}
}

func TestMetricsMiddlewareCapturesDuration(t *testing.T) {
	metrics := NewMetrics()
	middleware := MetricsMiddleware(metrics)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(inner)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/slow", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	families := metrics.Registry().GetFamilies()
	for _, f := range families {
		if f.Name == "naeos_request_duration_seconds" {
			if f.Count != 1 {
				t.Errorf("expected 1 observation, got %d", f.Count)
			}
			if f.Sum < 0.01 {
				t.Errorf("expected duration >= 0.01, got %f", f.Sum)
			}
			return
		}
	}
	t.Error("expected naeos_request_duration_seconds metric")
}

func TestUptime(t *testing.T) {
	metrics := NewMetrics()
	start := time.Now()
	uptime := time.Since(start).Seconds()
	metrics.SetUptime(uptime)

	families := metrics.Registry().GetFamilies()
	for _, f := range families {
		if f.Name == "naeos_uptime_seconds" {
			if f.Metrics[0].Value <= 0 {
				t.Error("expected positive uptime")
			}
			return
		}
	}
	t.Error("expected naeos_uptime_seconds metric")
}

func TestRegistryUnregister(t *testing.T) {
	reg := NewRegistry()
	reg.Register("temp", Counter, "Temporary")
	reg.CounterInc("temp", nil)

	if len(reg.GetFamilies()) != 1 {
		t.Fatal("expected 1 family before unregister")
	}

	reg.Unregister("temp")

	if len(reg.GetFamilies()) != 0 {
		t.Errorf("expected 0 families after unregister, got %d", len(reg.GetFamilies()))
	}
}

func TestRegistryGetFamily(t *testing.T) {
	reg := NewRegistry()
	reg.Register("my_counter", Counter, "My counter")

	family, ok := reg.GetFamily("my_counter")
	if !ok {
		t.Fatal("expected to find my_counter")
	}
	if family.Name != "my_counter" {
		t.Errorf("expected name 'my_counter', got %s", family.Name)
	}

	_, ok = reg.GetFamily("nonexistent")
	if ok {
		t.Error("expected nonexistent family to not be found")
	}
}

func TestRegistryCounterValue(t *testing.T) {
	reg := NewRegistry()
	reg.Register("req", Counter, "Requests")
	reg.CounterInc("req", map[string]string{"path": "/a"})
	reg.CounterInc("req", map[string]string{"path": "/a"})

	val := reg.CounterValue("req", map[string]string{"path": "/a"})
	if val != 2 {
		t.Errorf("expected 2, got %f", val)
	}

	val = reg.CounterValue("req", map[string]string{"path": "/b"})
	if val != 0 {
		t.Errorf("expected 0 for missing label, got %f", val)
	}

	val = reg.CounterValue("nonexistent", nil)
	if val != 0 {
		t.Errorf("expected 0 for nonexistent metric, got %f", val)
	}
}

func TestGaugeValue(t *testing.T) {
	reg := NewRegistry()
	reg.Register("g", Gauge, "Gauge")
	reg.GaugeSet("g", 42, nil)

	val := reg.GaugeValue("g", nil)
	if val != 42 {
		t.Errorf("expected 42, got %f", val)
	}
}

func TestRegistryOptions(t *testing.T) {
	reg := NewRegistryWithOptions(500, 10*time.Minute)
	if reg == nil {
		t.Fatal("expected registry")
	}
	if reg.maxCardinality != 500 {
		t.Errorf("expected cardinality 500, got %d", reg.maxCardinality)
	}
}

func TestRegistryOptionsDefaults(t *testing.T) {
	reg := NewRegistryWithOptions(-1, 0)
	if reg == nil {
		t.Fatal("expected registry")
	}
	if reg.maxCardinality != 1000 {
		t.Errorf("expected default cardinality 1000, got %d", reg.maxCardinality)
	}
}

func TestCardinalityLimit(t *testing.T) {
	reg := NewRegistryWithOptions(3, 5*time.Minute)
	reg.Register("limited", Counter, "Limited counter")

	for i := 0; i < 10; i++ {
		reg.CounterInc("limited", map[string]string{"i": string(rune('a' + i))})
	}

	family, ok := reg.GetFamily("limited")
	if !ok {
		t.Fatal("expected limited metric")
	}
	if len(family.Metrics) != 3 {
		t.Errorf("expected 3 metrics (cardinality limit), got %d", len(family.Metrics))
	}
}

func TestCleanupStale(t *testing.T) {
	reg := NewRegistry()
	reg.Register("stale_test", Counter, "Test")

	reg.CounterInc("stale_test", map[string]string{"k": "v1"})
	reg.CounterInc("stale_test", map[string]string{"k": "v2"})

	// Manually set old timestamps
	reg.mu.Lock()
	family := reg.metrics["stale_test"]
	for _, m := range family.Metrics {
		m.CreatedAt = time.Now().Add(-2 * time.Hour)
	}
	reg.mu.Unlock()

	reg.CleanupStale(1 * time.Hour)

	family, ok := reg.GetFamily("stale_test")
	if ok && len(family.Metrics) > 0 {
		t.Errorf("expected stale metrics to be cleaned, got %d", len(family.Metrics))
	}
}

func TestStartStopCleanup(t *testing.T) {
	reg := NewRegistryWithOptions(1000, 100*time.Millisecond)
	reg.StartCleanup(1 * time.Second)
	time.Sleep(50 * time.Millisecond)
	reg.StopCleanup()
}

func TestHistogramPrometheusFormat(t *testing.T) {
	reg := NewRegistry()
	reg.Register("hist_test", Histogram, "Test histogram")

	reg.HistogramObserveWithBuckets("hist_test", 0.1, nil)
	reg.HistogramObserveWithBuckets("hist_test", 1.0, nil)

	output := reg.FormatPrometheus()
	if !strings.Contains(output, "# TYPE hist_test histogram") {
		t.Error("expected histogram TYPE line")
	}
	if !strings.Contains(output, "hist_test_bucket{le=") {
		t.Error("expected bucket lines in output")
	}
	if !strings.Contains(output, "hist_test_sum") {
		t.Error("expected sum line")
	}
	if !strings.Contains(output, "hist_test_count") {
		t.Error("expected count line")
	}
}

func BenchmarkCounterInc(b *testing.B) {
	reg := NewRegistry()
	reg.Register("bench_counter", Counter, "bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.CounterInc("bench_counter", map[string]string{"label": "value"})
	}
}

func BenchmarkGaugeSet(b *testing.B) {
	reg := NewRegistry()
	reg.Register("bench_gauge", Gauge, "bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.GaugeSet("bench_gauge", float64(i), nil)
	}
}

func BenchmarkHistogramObserve(b *testing.B) {
	reg := NewRegistry()
	reg.RegisterWithBuckets("bench_hist", "bench", []float64{0.1, 0.5, 1.0})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.HistogramObserveWithBuckets("bench_hist", float64(i%100)/100, nil)
	}
}

func BenchmarkFormatPrometheus(b *testing.B) {
	reg := NewRegistry()
	reg.Register("bench_counter", Counter, "bench")
	reg.Register("bench_gauge", Gauge, "bench")
	reg.RegisterWithBuckets("bench_hist", "bench", []float64{0.1, 0.5, 1.0})
	for i := 0; i < 100; i++ {
		reg.CounterInc("bench_counter", map[string]string{"label": fmt.Sprintf("val-%d", i)})
		reg.GaugeSet("bench_gauge", float64(i), map[string]string{"label": fmt.Sprintf("val-%d", i)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.FormatPrometheus()
	}
}
