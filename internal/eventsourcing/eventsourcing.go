package eventsourcing

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Event struct {
	ID        string         `json:"id"`
	StreamID  string         `json:"stream_id"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data"`
	Version   int            `json:"version"`
	Timestamp time.Time      `json:"timestamp"`
}

type EventStore interface {
	Append(streamID string, events []Event) error
	Load(streamID string) ([]Event, error)
	LoadFrom(streamID string, fromVersion int) ([]Event, error)
}

type InMemoryStore struct {
	streams map[string][]Event
	mu      sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		streams: make(map[string][]Event),
	}
}

func (s *InMemoryStore) Append(streamID string, events []Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.streams[streamID]
	for i := range events {
		events[i].Version = len(existing) + i + 1
		events[i].StreamID = streamID
		if events[i].Timestamp.IsZero() {
			events[i].Timestamp = time.Now()
		}
	}
	s.streams[streamID] = append(existing, events...)
	return nil
}

func (s *InMemoryStore) Load(streamID string) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events, ok := s.streams[streamID]
	if !ok {
		return nil, nil // Stream does not exist — no events to return
	}
	out := make([]Event, len(events))
	copy(out, events)
	return out, nil
}

func (s *InMemoryStore) LoadFrom(streamID string, fromVersion int) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events, ok := s.streams[streamID]
	if !ok {
		return nil, nil // Stream does not exist — no events to return
	}

	var out []Event
	for _, e := range events {
		if e.Version >= fromVersion {
			out = append(out, e)
		}
	}
	return out, nil
}

func (s *InMemoryStore) StreamCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.streams)
}

func (s *InMemoryStore) EventCount(streamID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.streams[streamID])
}

type Aggregate struct {
	ID      string
	Version int
	Events  []Event
}

func (a *Aggregate) Apply(event Event) {
	event.Version = a.Version + 1
	a.Events = append(a.Events, event)
	a.Version = event.Version
}

func (a *Aggregate) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

type PipelineRunSnapshot struct {
	Aggregate
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	Artifacts  int            `json:"artifacts"`
	Error      string         `json:"error,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

func NewPipelineRun(id, name string) *PipelineRunSnapshot {
	return &PipelineRunSnapshot{
		Aggregate: Aggregate{ID: id},
		Name:      name,
		Status:    "started",
		Metadata:  make(map[string]any),
	}
}

func (s *PipelineRunSnapshot) Started() {
	s.Apply(Event{
		Type: "pipeline.started",
		Data: map[string]any{"name": s.Name},
	})
}

func (s *PipelineRunSnapshot) StageCompleted(stage string, artifacts int) {
	s.Status = "running"
	s.Artifacts = artifacts
	s.Apply(Event{
		Type: "pipeline.stage_completed",
		Data: map[string]any{"stage": stage, "artifacts": artifacts},
	})
}

func (s *PipelineRunSnapshot) Completed(artifacts int) {
	s.Status = "completed"
	s.Artifacts = artifacts
	s.Apply(Event{
		Type: "pipeline.completed",
		Data: map[string]any{"artifacts": artifacts},
	})
}

func (s *PipelineRunSnapshot) Failed(err error) {
	s.Status = "failed"
	if err != nil {
		s.Error = err.Error()
	}
	s.Apply(Event{
		Type: "pipeline.failed",
		Data: map[string]any{"error": s.Error},
	})
}

func (s *PipelineRunSnapshot) MarshalJSON() ([]byte, error) {
	type Alias PipelineRunSnapshot
	return json.Marshal(&struct {
		*Alias
		TotalEvents int `json:"total_events"`
	}{
		Alias:       (*Alias)(s),
		TotalEvents: len(s.Events),
	})
}

func RebuildFromEvents(id string, events []Event) *PipelineRunSnapshot {
	snap := &PipelineRunSnapshot{
		Aggregate: Aggregate{ID: id},
		Metadata:  make(map[string]any),
	}
	for _, e := range events {
		snap.Apply(e)
		switch e.Type {
		case "pipeline.started":
			if n, ok := e.Data["name"].(string); ok {
				snap.Name = n
			}
		case "pipeline.stage_completed":
			snap.Status = "running"
			if a, ok := e.Data["artifacts"].(float64); ok {
				snap.Artifacts = int(a)
			}
		case "pipeline.completed":
			snap.Status = "completed"
			if a, ok := e.Data["artifacts"].(float64); ok {
				snap.Artifacts = int(a)
			}
		case "pipeline.failed":
			snap.Status = "failed"
			if errStr, ok := e.Data["error"].(string); ok {
				snap.Error = errStr
			}
		}
	}
	return snap
}

func EventToJSON(e Event) ([]byte, error) {
	return json.Marshal(e)
}

func EventFromJSON(data []byte) (Event, error) {
	var e Event
	err := json.Unmarshal(data, &e)
	return e, err
}

func FormatEvent(e Event) string {
	return fmt.Sprintf("[%d] %s: %s", e.Version, e.Type, e.Timestamp.Format(time.RFC3339))
}

type FileStore struct {
	dir string
	mu  sync.RWMutex
}

func NewFileStore(dir string) *FileStore {
	return &FileStore{dir: dir}
}

func (s *FileStore) Append(streamID string, events []Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, _ := s.loadRaw(streamID)
	startVersion := len(existing) + 1
	for i := range events {
		events[i].Version = startVersion + i
		events[i].StreamID = streamID
		if events[i].Timestamp.IsZero() {
			events[i].Timestamp = time.Now()
		}
	}

	all := append(existing, events...)
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal events: %w", err)
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	path := s.dir + "/" + streamID + ".json"
	return os.WriteFile(path, data, 0o644)
}

func (s *FileStore) Load(streamID string) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events, err := s.loadRaw(streamID)
	if err != nil {
		return nil, err
	}
	if events == nil {
		return nil, nil // No persisted events for this stream
	}
	out := make([]Event, len(events))
	copy(out, events)
	return out, nil
}

func (s *FileStore) LoadFrom(streamID string, fromVersion int) ([]Event, error) {
	events, err := s.Load(streamID)
	if err != nil {
		return nil, err
	}
	var out []Event
	for _, e := range events {
		if e.Version >= fromVersion {
			out = append(out, e)
		}
	}
	return out, nil
}

func (s *FileStore) loadRaw(streamID string) ([]Event, error) {
	path := s.dir + "/" + streamID + ".json"
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No file on disk — stream has no events
		}
		return nil, err
	}
	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *FileStore) StreamIDs() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		name := e.Name()
		if len(name) > 5 && name[len(name)-5:] == ".json" {
			ids = append(ids, name[:len(name)-5])
		}
	}
	return ids, nil
}
