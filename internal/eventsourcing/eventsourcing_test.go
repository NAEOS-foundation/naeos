package eventsourcing

import (
	"fmt"
	"os"
	"testing"
)

func TestInMemoryStoreAppendLoad(t *testing.T) {
	store := NewInMemoryStore()

	events := []Event{
		{Type: "created", Data: map[string]any{"name": "test"}},
		{Type: "updated", Data: map[string]any{"key": "value"}},
	}

	if err := store.Append("stream-1", events); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.Load("stream-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(loaded))
	}
	if loaded[0].Version != 1 {
		t.Errorf("expected version 1, got %d", loaded[0].Version)
	}
	if loaded[1].Version != 2 {
		t.Errorf("expected version 2, got %d", loaded[1].Version)
	}
	if loaded[0].StreamID != "stream-1" {
		t.Errorf("expected stream ID 'stream-1', got %q", loaded[0].StreamID)
	}
}

func TestInMemoryStoreLoadFrom(t *testing.T) {
	store := NewInMemoryStore()
	events := []Event{
		{Type: "e1"}, {Type: "e2"}, {Type: "e3"},
	}
	store.Append("s1", events)

	from, err := store.LoadFrom("s1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(from) != 2 {
		t.Fatalf("expected 2 events from version 2, got %d", len(from))
	}
	if from[0].Type != "e2" {
		t.Errorf("expected 'e2', got %q", from[0].Type)
	}
}

func TestInMemoryStoreEmpty(t *testing.T) {
	store := NewInMemoryStore()
	events, err := store.Load("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if events != nil {
		t.Errorf("expected nil, got %v", events)
	}
}

func TestAggregateApply(t *testing.T) {
	agg := &Aggregate{ID: "a1"}
	agg.Apply(Event{Type: "e1", Data: map[string]any{}})
	agg.Apply(Event{Type: "e2", Data: map[string]any{}})

	if agg.Version != 2 {
		t.Errorf("expected version 2, got %d", agg.Version)
	}
	if len(agg.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(agg.Events))
	}
}

func TestPipelineRunSnapshot(t *testing.T) {
	snap := NewPipelineRun("run-1", "test-pipeline")
	snap.Started()
	snap.StageCompleted("parse", 1)
	snap.StageCompleted("generate", 5)
	snap.Completed(5)

	if snap.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", snap.Status)
	}
	if snap.Artifacts != 5 {
		t.Errorf("expected 5 artifacts, got %d", snap.Artifacts)
	}
	if len(snap.Events) != 4 {
		t.Errorf("expected 4 events, got %d", len(snap.Events))
	}
}

func TestPipelineRunSnapshotFailed(t *testing.T) {
	snap := NewPipelineRun("run-2", "test")
	snap.Started()
	snap.Failed(fmt.Errorf("something broke"))

	if snap.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", snap.Status)
	}
	if snap.Error != "something broke" {
		t.Errorf("expected error message, got %q", snap.Error)
	}
}

func TestRebuildFromEvents(t *testing.T) {
	events := []Event{
		{Type: "pipeline.started", Data: map[string]any{"name": "myapp"}},
		{Type: "pipeline.stage_completed", Data: map[string]any{"stage": "parse", "artifacts": float64(3)}},
		{Type: "pipeline.completed", Data: map[string]any{"artifacts": float64(3)}},
	}

	snap := RebuildFromEvents("run-x", events)
	if snap.Name != "myapp" {
		t.Errorf("expected name 'myapp', got %q", snap.Name)
	}
	if snap.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", snap.Status)
	}
	if snap.Artifacts != 3 {
		t.Errorf("expected 3 artifacts, got %d", snap.Artifacts)
	}
	if snap.Version != 3 {
		t.Errorf("expected version 3, got %d", snap.Version)
	}
}

func TestStoreCounts(t *testing.T) {
	store := NewInMemoryStore()
	store.Append("s1", []Event{{Type: "a"}, {Type: "b"}})
	store.Append("s2", []Event{{Type: "c"}})

	if store.StreamCount() != 2 {
		t.Errorf("expected 2 streams, got %d", store.StreamCount())
	}
	if store.EventCount("s1") != 2 {
		t.Errorf("expected 2 events in s1, got %d", store.EventCount("s1"))
	}
	if store.EventCount("s2") != 1 {
		t.Errorf("expected 1 event in s2, got %d", store.EventCount("s2"))
	}
}

func TestFileStoreAppendLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	events := []Event{
		{Type: "created", Data: map[string]any{"name": "test"}},
		{Type: "updated", Data: map[string]any{"key": "value"}},
	}

	if err := store.Append("stream-1", events); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.Load("stream-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(loaded))
	}
	if loaded[0].Version != 1 {
		t.Errorf("expected version 1, got %d", loaded[0].Version)
	}
	if loaded[1].Version != 2 {
		t.Errorf("expected version 2, got %d", loaded[1].Version)
	}
}

func TestFileStorePersistence(t *testing.T) {
	dir := t.TempDir()
	store1 := NewFileStore(dir)
	store1.Append("s1", []Event{{Type: "e1"}, {Type: "e2"}})

	store2 := NewFileStore(dir)
	loaded, err := store2.Load("s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 persisted events, got %d", len(loaded))
	}
}

func TestFileStoreLoadFrom(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)
	store.Append("s1", []Event{{Type: "e1"}, {Type: "e2"}, {Type: "e3"}})

	from, err := store.LoadFrom("s1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(from) != 2 {
		t.Fatalf("expected 2 events from version 2, got %d", len(from))
	}
}

func TestFileStoreStreamIDs(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)
	store.Append("stream-a", []Event{{Type: "e1"}})
	store.Append("stream-b", []Event{{Type: "e2"}})

	ids, err := store.StreamIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 stream IDs, got %d", len(ids))
	}
}

func TestFileStoreEmpty(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	events, err := store.Load("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if events != nil {
		t.Errorf("expected nil, got %v", events)
	}
}

func TestFileStoreAppendIncremental(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	store.Append("s1", []Event{{Type: "e1"}})
	store.Append("s1", []Event{{Type: "e2"}})
	store.Append("s1", []Event{{Type: "e3"}})

	loaded, _ := store.Load("s1")
	if len(loaded) != 3 {
		t.Fatalf("expected 3 events, got %d", len(loaded))
	}
	for i, e := range loaded {
		if e.Version != i+1 {
			t.Errorf("expected version %d, got %d", i+1, e.Version)
		}
	}
}

func TestFileStoreJSONFile(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)
	store.Append("test-stream", []Event{{Type: "test", Data: map[string]any{"key": "value"}}})

	path := dir + "/test-stream.json"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON file")
	}
}
