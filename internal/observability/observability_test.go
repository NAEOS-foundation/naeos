package observability

import (
	"testing"
	"time"
)

func TestTracer(t *testing.T) {
	tr := NewTracer("test")

	span := tr.StartSpan("operation")
	if span == nil {
		t.Fatal("expected span")
	}
	if span.Name != "operation" {
		t.Errorf("expected name 'operation', got %s", span.Name)
	}

	tr.EndSpan(span)
	if span.EndTime.IsZero() {
		t.Error("expected end time to be set")
	}
}

func TestTracerWithParent(t *testing.T) {
	tr := NewTracer("test")

	parent := tr.StartSpan("parent")
	child := tr.StartSpanWithParent("child", parent.SpanID)

	if child.ParentID != parent.SpanID {
		t.Errorf("expected parent ID %s, got %s", parent.SpanID, child.ParentID)
	}
}

func TestTracerGetSpans(t *testing.T) {
	tr := NewTracer("test")

	tr.StartSpan("span1")
	tr.StartSpan("span2")
	tr.StartSpan("span3")

	spans := tr.GetSpans()
	if len(spans) != 3 {
		t.Errorf("expected 3 spans, got %d", len(spans))
	}
}

func TestTracerGetSpansByTrace(t *testing.T) {
	tr := NewTracer("test")

	span1 := tr.StartSpan("span1")
	tr.StartSpanWithParent("child1", span1.SpanID)

	spans := tr.GetSpansByTrace(span1.TraceID)
	if len(spans) != 2 {
		t.Errorf("expected 2 spans, got %d", len(spans))
	}
}

func TestTracerAddEvent(t *testing.T) {
	tr := NewTracer("test")

	span := tr.StartSpan("operation")
	tr.AddEvent(span, "event1", map[string]any{"key": "value"})

	if len(span.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(span.Events))
	}
}

func TestTracerSetStatus(t *testing.T) {
	tr := NewTracer("test")

	span := tr.StartSpan("operation")
	tr.SetStatus(span, SpanStatusOK, "success")

	if span.Status.Code != SpanStatusOK {
		t.Errorf("expected OK status, got %d", span.Status.Code)
	}
}

func TestLogger(t *testing.T) {
	l := NewLogger("test", LogLevelInfo)

	l.Info("test message", map[string]any{"key": "value"})

	entries := l.GetEntries()
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != "test message" {
		t.Errorf("expected 'test message', got %s", entries[0].Message)
	}
}

func TestLoggerLevel(t *testing.T) {
	l := NewLogger("test", LogLevelWarn)

	l.Debug("debug", nil)
	l.Info("info", nil)
	l.Warn("warn", nil)
	l.Error("error", nil)

	entries := l.GetEntries()
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLoggerGetEntriesByLevel(t *testing.T) {
	l := NewLogger("test", LogLevelDebug)

	l.Info("info1", nil)
	l.Warn("warn1", nil)
	l.Error("error1", nil)

	warns := l.GetEntriesByLevel(LogLevelWarn)
	if len(warns) != 1 {
		t.Errorf("expected 1 warn, got %d", len(warns))
	}
}

func TestMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector("test")

	mc.Counter("requests", map[string]string{"method": "GET"})
	mc.Counter("requests", map[string]string{"method": "GET"})
	mc.Gauge("connections", 10, nil)

	metrics := mc.GetMetrics()
	if len(metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(metrics))
	}
}

func TestMetricsCollectorCounter(t *testing.T) {
	mc := NewMetricsCollector("test")

	mc.Counter("requests", nil)
	mc.Counter("requests", nil)

	metrics := mc.GetMetrics()
	for _, m := range metrics {
		if m.Name == "requests" && m.Value != 2 {
			t.Errorf("expected value 2, got %f", m.Value)
		}
	}
}

func TestMetricsCollectorHistogram(t *testing.T) {
	mc := NewMetricsCollector("test")

	mc.Histogram("duration", 0.1, nil)
	mc.Histogram("duration", 0.2, nil)

	metrics := mc.GetMetrics()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}
}

func TestObservabilityStack(t *testing.T) {
	stack := NewStack("test")

	if stack.Tracer == nil {
		t.Error("expected tracer")
	}
	if stack.Logger == nil {
		t.Error("expected logger")
	}
	if stack.Metrics == nil {
		t.Error("expected metrics")
	}

	stack.Logger.Info("test", nil)
	stack.Metrics.Counter("test", nil)

	span := stack.Tracer.StartSpan("test")
	stack.Tracer.EndSpan(span)

	if len(stack.Tracer.GetSpans()) != 1 {
		t.Error("expected 1 span")
	}

	if len(stack.Logger.GetEntries()) != 1 {
		t.Error("expected 1 log entry")
	}

	if len(stack.Metrics.GetMetrics()) != 1 {
		t.Error("expected 1 metric")
	}
}

func TestSpanAttributes(t *testing.T) {
	tr := NewTracer("test")

	span := tr.StartSpan("operation")
	span.Attributes["key"] = "value"
	span.Attributes["number"] = 42

	if span.Attributes["key"] != "value" {
		t.Error("expected attribute")
	}
}

func TestSpanDuration(t *testing.T) {
	tr := NewTracer("test")

	span := tr.StartSpan("operation")
	time.Sleep(10 * time.Millisecond)
	tr.EndSpan(span)

	duration := span.EndTime.Sub(span.StartTime)
	if duration < 10*time.Millisecond {
		t.Error("expected duration >= 10ms")
	}
}
