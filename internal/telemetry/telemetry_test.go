package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type mockExporter struct {
	spans [][]Span
	mu    sync.Mutex
}

func (m *mockExporter) ExportSpans(spans []Span) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	batch := make([]Span, len(spans))
	copy(batch, spans)
	m.spans = append(m.spans, batch)
	return nil
}

func (m *mockExporter) Flush() error { return nil }

func (m *mockExporter) totalSpans() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, batch := range m.spans {
		total += len(batch)
	}
	return total
}

func TestServiceStartEndSpan(t *testing.T) {
	mock := &mockExporter{}
	svc := NewService(Config{BatchSize: 10}, mock)

	span := svc.StartSpan("test-span")
	if span.Name != "test-span" {
		t.Errorf("expected name 'test-span', got %s", span.Name)
	}
	if span.ID == "" {
		t.Error("expected non-empty ID")
	}

	svc.EndSpan(span)
	if svc.SpanCount() != 1 {
		t.Errorf("expected 1 buffered span, got %d", svc.SpanCount())
	}
}

func TestServiceAutoFlush(t *testing.T) {
	mock := &mockExporter{}
	svc := NewService(Config{BatchSize: 2}, mock)

	for i := 0; i < 3; i++ {
		span := svc.StartSpan("span")
		svc.EndSpan(span)
	}

	if mock.totalSpans() < 2 {
		t.Errorf("expected at least 2 exported spans, got %d", mock.totalSpans())
	}
}

func TestServiceFlush(t *testing.T) {
	mock := &mockExporter{}
	svc := NewService(Config{BatchSize: 100}, mock)

	span := svc.StartSpan("span")
	svc.EndSpan(span)

	if err := svc.Flush(); err != nil {
		t.Fatal(err)
	}
	if svc.SpanCount() != 0 {
		t.Errorf("expected 0 buffered spans after flush, got %d", svc.SpanCount())
	}
	if mock.totalSpans() != 1 {
		t.Errorf("expected 1 exported span, got %d", mock.totalSpans())
	}
}

func TestParentChildSpans(t *testing.T) {
	mock := &mockExporter{}
	svc := NewService(Config{BatchSize: 100}, mock)

	parent := svc.StartSpan("parent")
	child := svc.StartSpanWithParent("child", parent.ID)

	if child.ParentID != parent.ID {
		t.Errorf("expected child ParentID=%s, got %s", parent.ID, child.ParentID)
	}
	if child.ID == parent.ID {
		t.Error("child and parent should have different IDs")
	}
}

func TestSpanLabels(t *testing.T) {
	span := &Span{
		Name:   "labeled",
		Labels: map[string]string{"env": "test", "service": "api"},
	}
	if span.Labels["env"] != "test" {
		t.Errorf("expected env=test, got %s", span.Labels["env"])
	}
}

func TestHTTPExporterNew(t *testing.T) {
	exp := NewHTTPExporter("http://localhost:9999", 0)
	if exp.client == nil {
		t.Fatal("expected non-nil client")
	}
	if exp.client.Timeout != 5*time.Second {
		t.Errorf("expected default timeout 5s, got %v", exp.client.Timeout)
	}
	if exp.endpoint != "http://localhost:9999" {
		t.Errorf("expected endpoint 'http://localhost:9999', got %s", exp.endpoint)
	}
}

func TestHTTPExporterFlushEmpty(t *testing.T) {
	exp := NewHTTPExporter("http://localhost:9999", time.Second)
	if err := exp.Flush(); err != nil {
		t.Fatalf("expected nil error on empty flush, got %v", err)
	}
}

func TestHTTPExporterExportSpans(t *testing.T) {
	var received []Span
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/traces" {
			t.Errorf("expected path /v1/traces, got %s", r.URL.Path)
		}
		var spans []Span
		if err := json.NewDecoder(r.Body).Decode(&spans); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		received = spans
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exp := NewHTTPExporter(srv.URL, 5*time.Second)
	spans := []Span{
		{Name: "s1", ID: "id1", Status: "ok"},
		{Name: "s2", ID: "id2", Status: "ok"},
	}
	if err := exp.ExportSpans(spans); err != nil {
		t.Fatalf("ExportSpans failed: %v", err)
	}
	if len(received) != 2 {
		t.Fatalf("expected 2 spans received by server, got %d", len(received))
	}
	if received[0].Name != "s1" || received[1].Name != "s2" {
		t.Errorf("unexpected span names: %v, %v", received[0].Name, received[1].Name)
	}
}

func TestHTTPExporterExportSpansError(t *testing.T) {
	exp := NewHTTPExporter("http://127.0.0.1:1", time.Second)
	spans := []Span{{Name: "s1", ID: "id1", Status: "ok"}}
	err := exp.ExportSpans(spans)
	if err == nil {
		t.Fatal("expected error for bad endpoint, got nil")
	}
}

func TestServiceDefaults(t *testing.T) {
	mock := &mockExporter{}
	svc := NewService(Config{}, mock)
	if svc.config.Timeout != 5*time.Second {
		t.Errorf("expected default timeout 5s, got %v", svc.config.Timeout)
	}
	if svc.config.BatchSize != 100 {
		t.Errorf("expected default batch size 100, got %d", svc.config.BatchSize)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	id3 := generateID()

	if id1 == id2 || id2 == id3 {
		t.Error("expected unique IDs from generateID")
	}
	// The counter is global and monotonically increasing.
	// Just verify the format and that they are ordered.
	if id1 >= id2 || id2 >= id3 {
		t.Errorf("expected strictly increasing IDs: %s < %s < %s", id1, id2, id3)
	}
}

func TestSpanCountAfterMultipleEndSpan(t *testing.T) {
	mock := &mockExporter{}
	svc := NewService(Config{BatchSize: 100}, mock)

	for i := 0; i < 5; i++ {
		span := svc.StartSpan("span")
		svc.EndSpan(span)
	}

	if svc.SpanCount() != 5 {
		t.Errorf("expected 5 buffered spans, got %d", svc.SpanCount())
	}
}

func TestSpanStatusString(t *testing.T) {
	if StatusOK.String() != "OK" {
		t.Errorf("expected OK, got %s", StatusOK.String())
	}
	if StatusError.String() != "ERROR" {
		t.Errorf("expected ERROR, got %s", StatusError.String())
	}
	if StatusUnset.String() != "UNSET" {
		t.Errorf("expected UNSET, got %s", StatusUnset.String())
	}
}

func TestSpanAttributes(t *testing.T) {
	t.Run("set and get string", func(t *testing.T) {
		attr := &SpanAttributes{}
		attr.Set("key", "value")
		v, ok := attr.Get("key")
		if !ok || v != "value" {
			t.Errorf("expected value, got %v, %v", v, ok)
		}
	})

	t.Run("set and get int64", func(t *testing.T) {
		attr := &SpanAttributes{}
		attr.Set("count", int64(42))
		v, ok := attr.Get("count")
		if !ok || v != int64(42) {
			t.Errorf("expected 42, got %v", v)
		}
	})

	t.Run("set and get float64", func(t *testing.T) {
		attr := &SpanAttributes{}
		attr.Set("rate", 3.14)
		v, ok := attr.Get("rate")
		if !ok || v != 3.14 {
			t.Errorf("expected 3.14, got %v", v)
		}
	})

	t.Run("set and get bool", func(t *testing.T) {
		attr := &SpanAttributes{}
		attr.Set("flag", true)
		v, ok := attr.Get("flag")
		if !ok || v != true {
			t.Errorf("expected true, got %v", v)
		}
	})

	t.Run("has returns true", func(t *testing.T) {
		attr := &SpanAttributes{}
		attr.Set("k", "v")
		if !attr.Has("k") {
			t.Error("expected Has to return true")
		}
	})

	t.Run("has returns false for missing", func(t *testing.T) {
		attr := &SpanAttributes{}
		if attr.Has("missing") {
			t.Error("expected Has to return false")
		}
	})

	t.Run("len returns attribute count", func(t *testing.T) {
		attr := &SpanAttributes{}
		attr.Set("a", "1")
		attr.Set("b", "2")
		if attr.Len() != 2 {
			t.Errorf("expected 2, got %d", attr.Len())
		}
	})

	t.Run("len zero for empty", func(t *testing.T) {
		attr := &SpanAttributes{}
		if attr.Len() != 0 {
			t.Errorf("expected 0, got %d", attr.Len())
		}
	})

	t.Run("get missing returns false", func(t *testing.T) {
		attr := &SpanAttributes{}
		_, ok := attr.Get("missing")
		if ok {
			t.Error("expected false for missing key")
		}
	})

	t.Run("marshal JSON", func(t *testing.T) {
		attr := &SpanAttributes{}
		attr.Set("name", "test")
		attr.Set("count", int64(3))
		data, err := attr.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty JSON")
		}
	})
}

func TestMetricsCollector(t *testing.T) {
	t.Run("counter incr and get", func(t *testing.T) {
		mc := NewMetricsCollector()
		mc.IncrCounter("requests", 5)
		mc.IncrCounter("requests", 3)
		if v := mc.GetCounter("requests"); v != 8 {
			t.Errorf("expected 8, got %d", v)
		}
	})

	t.Run("get non-existent counter returns 0", func(t *testing.T) {
		mc := NewMetricsCollector()
		if v := mc.GetCounter("missing"); v != 0 {
			t.Errorf("expected 0, got %d", v)
		}
	})

	t.Run("gauge set and get", func(t *testing.T) {
		mc := NewMetricsCollector()
		mc.SetGauge("cpu", 0.75)
		if v := mc.GetGauge("cpu"); v != 0.75 {
			t.Errorf("expected 0.75, got %f", v)
		}
	})

	t.Run("get non-existent gauge returns 0", func(t *testing.T) {
		mc := NewMetricsCollector()
		if v := mc.GetGauge("missing"); v != 0.0 {
			t.Errorf("expected 0.0, got %f", v)
		}
	})

	t.Run("histogram record and get", func(t *testing.T) {
		mc := NewMetricsCollector()
		mc.RecordHistogram("latency", 1.0)
		mc.RecordHistogram("latency", 2.0)
		vals := mc.GetHistogram("latency")
		if len(vals) != 2 {
			t.Errorf("expected 2 values, got %d", len(vals))
		}
		if vals[0] != 1.0 || vals[1] != 2.0 {
			t.Errorf("unexpected values: %v", vals)
		}
	})

	t.Run("get non-existent histogram returns nil", func(t *testing.T) {
		mc := NewMetricsCollector()
		if v := mc.GetHistogram("missing"); v != nil {
			t.Errorf("expected nil, got %v", v)
		}
	})

	t.Run("export returns all metrics", func(t *testing.T) {
		mc := NewMetricsCollector()
		mc.IncrCounter("req", 10)
		mc.SetGauge("cpu", 0.5)
		mc.RecordHistogram("lat", 100.0)
		exported := mc.Export()
		if len(exported) != 3 {
			t.Errorf("expected 3 metrics, got %d", len(exported))
		}
	})
}

func TestSamplers(t *testing.T) {
	t.Run("always sample", func(t *testing.T) {
		s := AlwaysSample{}
		if !s.Sample("anything") {
			t.Error("AlwaysSample should return true")
		}
	})

	t.Run("never sample", func(t *testing.T) {
		s := NeverSample{}
		if s.Sample("anything") {
			t.Error("NeverSample should return false")
		}
	})

	t.Run("probabilistic sampler 100%", func(t *testing.T) {
		s := NewProbabilisticSampler(1.0)
		if !s.Sample("x") {
			t.Error("100% sampler should always return true")
		}
	})

	t.Run("probabilistic sampler 0%", func(t *testing.T) {
		s := NewProbabilisticSampler(0.0)
		if s.Sample("x") {
			t.Error("0% sampler should always return false")
		}
	})

	t.Run("probabilistic sampler clamps rate", func(t *testing.T) {
		s1 := NewProbabilisticSampler(-0.5)
		if s1.Rate != 0.0 {
			t.Errorf("expected 0, got %f", s1.Rate)
		}
		s2 := NewProbabilisticSampler(2.0)
		if s2.Rate != 1.0 {
			t.Errorf("expected 1, got %f", s2.Rate)
		}
	})

	t.Run("rate limiter sample", func(t *testing.T) {
		s := NewRateLimiterSampler(5)
		allowed := 0
		for i := 0; i < 100; i++ {
			if s.Sample("x") {
				allowed++
			}
		}
		if allowed > 10 {
			t.Errorf("expected ~5 samples per second, got %d", allowed)
		}
	})

	t.Run("rate limiter clamps zero", func(t *testing.T) {
		s := NewRateLimiterSampler(0)
		if s.maxPerSecond != 1 {
			t.Errorf("expected 1, got %d", s.maxPerSecond)
		}
	})
}

func TestBatchProcessor(t *testing.T) {
	t.Run("add span triggers flush at batch size", func(t *testing.T) {
		mock := &mockExporter{}
		bp := NewBatchProcessor(mock, 2, 0, 10)
		bp.AddSpan(Span{Name: "s1"})
		if bp.QueueSize() != 1 {
			t.Errorf("expected 1 queued, got %d", bp.QueueSize())
		}
		bp.AddSpan(Span{Name: "s2"})
		if bp.QueueSize() != 0 {
			t.Errorf("expected 0 after flush, got %d", bp.QueueSize())
		}
		if mock.totalSpans() != 2 {
			t.Errorf("expected 2 exported, got %d", mock.totalSpans())
		}
		bp.Stop()
	})

	t.Run("explicit flush", func(t *testing.T) {
		mock := &mockExporter{}
		bp := NewBatchProcessor(mock, 10, 0, 100)
		bp.AddSpan(Span{Name: "s1"})
		bp.Flush()
		if mock.totalSpans() != 1 {
			t.Errorf("expected 1 exported, got %d", mock.totalSpans())
		}
		bp.Stop()
	})

	t.Run("queue full error", func(t *testing.T) {
		bp := NewBatchProcessor(&mockExporter{}, 10, 0, 2)
		bp.AddSpan(Span{Name: "s1"})
		bp.AddSpan(Span{Name: "s2"})
		err := bp.AddSpan(Span{Name: "s3"})
		if err == nil {
			t.Error("expected queue full error")
		}
		bp.Stop()
	})

	t.Run("stop closes done channel", func(t *testing.T) {
		bp := NewBatchProcessor(&mockExporter{}, 10, time.Hour, 100)
		bp.Stop()
		_, ok := <-bp.done
		if ok {
			t.Error("expected done channel to be closed")
		}
	})

	t.Run("default batch size and max queue", func(t *testing.T) {
		bp := NewBatchProcessor(&mockExporter{}, 0, 0, 0)
		if bp.batchSize != 10 {
			t.Errorf("expected default batch size 10, got %d", bp.batchSize)
		}
		if bp.maxQueue != 1000 {
			t.Errorf("expected default max queue 1000, got %d", bp.maxQueue)
		}
	})
}

func TestTracer(t *testing.T) {
	t.Run("start span", func(t *testing.T) {
		exp := NewInMemoryExporter()
		tr := NewTracer("test-service", exp)
		span := tr.StartSpan("op1")
		if span.Name != "op1" {
			t.Errorf("expected op1, got %s", span.Name)
		}
		if span.Labels["service"] != "test-service" {
			t.Errorf("expected test-service, got %s", span.Labels["service"])
		}
		if span.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("start span with parent", func(t *testing.T) {
		exp := NewInMemoryExporter()
		tr := NewTracer("svc", exp)
		child := tr.StartSpanWithParent("child", "parent-id", "trace-123")
		if child.ParentID != "parent-id" {
			t.Errorf("expected parent-id, got %s", child.ParentID)
		}
		if child.TraceID != "trace-123" {
			t.Errorf("expected trace-123, got %s", child.TraceID)
		}
	})

	t.Run("end span exports span", func(t *testing.T) {
		exp := NewInMemoryExporter()
		tr := NewTracer("svc", exp)
		span := tr.StartSpan("op")
		err := tr.EndSpan(span)
		if err != nil {
			t.Fatal(err)
		}
		if exp.TotalSpanCount() != 1 {
			t.Errorf("expected 1 span, got %d", exp.TotalSpanCount())
		}
	})

	t.Run("flush", func(t *testing.T) {
		exp := NewInMemoryExporter()
		tr := NewTracer("svc", exp)
		if err := tr.Flush(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestMultiExporter(t *testing.T) {
	t.Run("fan-out export spans", func(t *testing.T) {
		e1 := NewInMemoryExporter()
		e2 := NewInMemoryExporter()
		m := NewMultiExporter(e1, e2)
		spans := []Span{{Name: "s1"}}
		err := m.ExportSpans(spans)
		if err != nil {
			t.Fatal(err)
		}
		if e1.TotalSpanCount() != 1 || e2.TotalSpanCount() != 1 {
			t.Errorf("expected 1 span in each exporter, got %d, %d", e1.TotalSpanCount(), e2.TotalSpanCount())
		}
	})

	t.Run("fan-out flush", func(t *testing.T) {
		e1 := NewInMemoryExporter()
		e2 := NewInMemoryExporter()
		m := NewMultiExporter(e1, e2)
		if err := m.Flush(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestConsoleExporter(t *testing.T) {
	t.Run("export spans writes JSON lines", func(t *testing.T) {
		var buf bytes.Buffer
		c := NewConsoleExporter(&buf)
		spans := []Span{{Name: "s1"}, {Name: "s2"}}
		err := c.ExportSpans(spans)
		if err != nil {
			t.Fatal(err)
		}
		lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
		if len(lines) != 2 {
			t.Errorf("expected 2 lines, got %d", len(lines))
		}
	})

	t.Run("flush returns nil for plain writer", func(t *testing.T) {
		var buf bytes.Buffer
		c := NewConsoleExporter(&buf)
		if err := c.Flush(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestInMemoryExporter(t *testing.T) {
	t.Run("all spans after export", func(t *testing.T) {
		e := NewInMemoryExporter()
		e.ExportSpans([]Span{{Name: "s1"}, {Name: "s2"}})
		e.ExportSpans([]Span{{Name: "s3"}})
		if e.TotalSpanCount() != 3 {
			t.Errorf("expected 3 total, got %d", e.TotalSpanCount())
		}
		if e.BatchCount() != 2 {
			t.Errorf("expected 2 batches, got %d", e.BatchCount())
		}
		all := e.AllSpans()
		if len(all) != 3 {
			t.Errorf("expected 3 spans, got %d", len(all))
		}
	})

	t.Run("reset clears all", func(t *testing.T) {
		e := NewInMemoryExporter()
		e.ExportSpans([]Span{{Name: "s1"}})
		e.Reset()
		if e.TotalSpanCount() != 0 {
			t.Errorf("expected 0 after reset, got %d", e.TotalSpanCount())
		}
		if e.BatchCount() != 0 {
			t.Errorf("expected 0 batches after reset, got %d", e.BatchCount())
		}
	})

	t.Run("flush returns nil", func(t *testing.T) {
		e := NewInMemoryExporter()
		if err := e.Flush(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestHTTPExporterNon200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	exp := NewHTTPExporter(srv.URL, 5*time.Second)
	err := exp.ExportSpans([]Span{{Name: "s1"}})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
