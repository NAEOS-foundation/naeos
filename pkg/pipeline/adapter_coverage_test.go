package pipeline

import (
	"context"
	"errors"
	"testing"

	pm "github.com/NAEOS-foundation/naeos/internal/pipelinemiddleware"
)

type failMW struct{ name string }

func (m *failMW) Name() string { return m.name }

func (m *failMW) Wrap(stage string, next pm.StageFunc) pm.StageFunc {
	return func(ctx context.Context, in *pm.StageInput) (*pm.StageOutput, error) {
		return nil, errors.New("middleware failed")
	}
}

func TestAdapterRunWithMiddlewareFailure(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewAdapter(p)
	adapter.UseMiddleware("pre-process", &failMW{name: "fail-mw"})
	_, err = adapter.RunWithMiddleware(context.Background(), "project: test")
	if err == nil {
		t.Error("expected error from middleware failure")
	}
}

func TestAdapterRunWithPipelineFailure(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewAdapter(p)
	_, err = adapter.RunWithMiddleware(context.Background(), "")
	if err == nil {
		t.Error("expected error from empty input")
	}
}

func TestAdapterEventCount(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewAdapter(p)
	adapter.RunWithMiddleware(context.Background(), "project: test")
	if count := adapter.EventCount(); count == 0 {
		t.Error("expected non-zero event count")
	}
}

func TestAdapterRunIDAfterRun(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewAdapter(p)
	adapter.RunWithMiddleware(context.Background(), "project: test")
	if id := adapter.RunID(); id == "" {
		t.Error("expected non-empty run ID")
	}
}

func TestAdapterRunSnapshot(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewAdapter(p)
	adapter.RunWithMiddleware(context.Background(), "project: test")
	snapshot := adapter.RunSnapshot()
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snapshot.Status != "completed" {
		t.Errorf("expected 'completed', got %q", snapshot.Status)
	}
}
