package telemetry

import (
	"sync"
	"testing"
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
