package observability

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}

func TestGenerateIDConcurrency(t *testing.T) {
	var mu sync.Mutex
	ids := make(map[string]bool)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := generateID()
			mu.Lock()
			if ids[id] {
				t.Errorf("duplicate ID: %s", id)
			}
			ids[id] = true
			mu.Unlock()
		}()
	}
	wg.Wait()
}

func TestJSONExporter(t *testing.T) {
	exporter := NewJSONExporter()

	spans := []*Span{{Name: "test-span"}}
	exporter.ExportSpans(spans)

	metrics := []*Metric{{Name: "test-metric", Type: MetricCounter, Value: 1}}
	exporter.ExportMetrics(metrics)

	entries := []LogEntry{{Message: "test-log"}}
	exporter.ExportLogs(entries)

	data, err := exporter.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFileExporter(t *testing.T) {
	tmpFile := t.TempDir() + "/export.jsonl"
	exporter, err := NewFileExporter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileExporter failed: %v", err)
	}
	defer exporter.Close()

	spans := []*Span{{Name: "test-span"}}
	exporter.ExportSpans(spans)

	metrics := []*Metric{{Name: "test-metric", Type: MetricCounter, Value: 1}}
	exporter.ExportMetrics(metrics)

	entries := []LogEntry{{Message: "test-log"}}
	exporter.ExportLogs(entries)

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestMemoryCleanup(t *testing.T) {
	tracer := NewTracer("test")
	cleanup := NewMemoryCleanup(tracer, 100*time.Millisecond)

	span := tracer.StartSpan("old-span")
	span.EndTime = time.Now().Add(-200 * time.Millisecond)

	tracer.StartSpan("new-span")

	cleanup.cleanup()

	if tracer.SpanCount() != 1 {
		t.Errorf("expected 1 span after cleanup, got %d", tracer.SpanCount())
	}
}

func TestMemoryCleanupStartStop(t *testing.T) {
	tracer := NewTracer("test")
	cleanup := NewMemoryCleanup(tracer, time.Second)

	cleanup.Start()
	cleanup.Start()

	cleanup.Stop()
	cleanup.Stop()
}

func TestHistogramBuckets(t *testing.T) {
	mc := NewMetricsCollector("test")

	mc.Histogram("request_duration", 0.05, nil)
	mc.Histogram("request_duration", 0.5, nil)
	mc.Histogram("request_duration", 5.0, nil)

	h := mc.GetHistogram("request_duration")
	if h == nil {
		t.Fatal("expected histogram")
	}

	if h.Count != 3 {
		t.Errorf("expected count 3, got %d", h.Count)
	}

	if h.Min != 0.05 {
		t.Errorf("expected min 0.05, got %f", h.Min)
	}

	if h.Max != 5.0 {
		t.Errorf("expected max 5.0, got %f", h.Max)
	}
}

func TestMetricsCollectorReset(t *testing.T) {
	mc := NewMetricsCollector("test")
	mc.Counter("req", nil)
	mc.Gauge("cpu", 50.0, nil)

	mc.Reset()

	metrics := mc.GetMetrics()
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics after reset, got %d", len(metrics))
	}
}

func TestLoggerSetLevel(t *testing.T) {
	l := NewLogger("test", LogLevelInfo)
	l.Debug("debug", nil)
	l.Info("info", nil)

	if len(l.GetEntries()) != 1 {
		t.Errorf("expected 1 entry, got %d", len(l.GetEntries()))
	}

	l.SetLevel(LogLevelDebug)
	l.Debug("debug2", nil)

	if len(l.GetEntries()) != 2 {
		t.Errorf("expected 2 entries, got %d", len(l.GetEntries()))
	}
}

func TestLoggerClear(t *testing.T) {
	l := NewLogger("test", LogLevelDebug)
	l.Info("msg1", nil)
	l.Info("msg2", nil)

	l.Clear()

	if len(l.GetEntries()) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(l.GetEntries()))
	}
}

func TestTracerSpanCount(t *testing.T) {
	tracer := NewTracer("test")
	tracer.StartSpan("a")
	tracer.StartSpan("b")

	if tracer.SpanCount() != 2 {
		t.Errorf("expected 2 spans, got %d", tracer.SpanCount())
	}
}

func TestTracerClear(t *testing.T) {
	tracer := NewTracer("test")
	tracer.StartSpan("a")
	tracer.StartSpan("b")

	tracer.Clear()

	if tracer.SpanCount() != 0 {
		t.Errorf("expected 0 spans after clear, got %d", tracer.SpanCount())
	}
}

func BenchmarkTracerStartEnd(b *testing.B) {
	tracer := NewTracer("bench")
	for i := 0; i < b.N; i++ {
		span := tracer.StartSpan("op")
		tracer.EndSpan(span)
	}
}

func BenchmarkLoggerInfo(b *testing.B) {
	logger := NewLogger("bench", LogLevelDebug)
	for i := 0; i < b.N; i++ {
		logger.Info("message", map[string]any{"key": "value"})
	}
}

func BenchmarkMetricsCollector(b *testing.B) {
	mc := NewMetricsCollector("bench")
	for i := 0; i < b.N; i++ {
		mc.Counter("test_metric", map[string]string{"label": "val"})
	}
}

func BenchmarkJSONExporter(b *testing.B) {
	tracer := NewTracer("bench")
	for i := 0; i < 100; i++ {
		span := tracer.StartSpan("op")
		tracer.EndSpan(span)
	}
	exporter := NewJSONExporter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exporter.ExportSpans(tracer.GetSpans())
		exporter.Flush()
	}
}
