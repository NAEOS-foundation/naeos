package verification

import (
	"sync"
	"time"
)

// EventType classifies a verification event.
type EventType string

const (
	EventVerificationRequested EventType = "verification.requested"
	EventVerificationComplete  EventType = "verification.complete"
	EventVerificationFailed    EventType = "verification.failed"
	EventVerifierRegistered    EventType = "verifier.registered"
	EventVerifierRemoved       EventType = "verifier.removed"
)

// Event is a verification lifecycle event that can be emitted to external
// sinks.
type Event struct {
	Type      EventType
	Timestamp time.Time
	Target    string // evidence ID or artifact being verified
	Verifier  string
	Status    VerificationStatus
	Message   string
	Metadata  map[string]any
}

// EventSink receives verification events.
type EventSink interface {
	// Emit delivers a verification event.
	Emit(e Event) error
}

// EventBus fans out verification events to registered sinks. It is
// thread-safe and failure-isolated (one sink's error does not stop others).
type EventBus struct {
	mu    sync.RWMutex
	sinks map[string]EventSink
	// history retains emitted events for auditability.
	history []Event
}

// NewEventBus creates a verification event bus.
func NewEventBus() *EventBus {
	return &EventBus{sinks: make(map[string]EventSink)}
}

// RegisterSink registers a named event sink.
func (b *EventBus) RegisterSink(name string, s EventSink) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sinks[name] = s
}

// UnregisterSink removes a sink by name.
func (b *EventBus) UnregisterSink(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.sinks, name)
}

// SinkNames returns the registered sink names (sorted).
func (b *EventBus) SinkNames() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, 0, len(b.sinks))
	for name := range b.sinks {
		out = append(out, name)
	}
	sortStrings(out)
	return out
}

// Publish emits an event to all registered sinks and records it in history.
// Sink errors are collected and returned as an aggregate; all sinks are
// still attempted regardless of individual failures.
func (b *EventBus) Publish(e Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}

	b.mu.RLock()
	sinks := make([]EventSink, 0, len(b.sinks))
	for _, s := range b.sinks {
		sinks = append(sinks, s)
	}
	b.history = append(b.history, e)
	b.mu.RUnlock()

	var firstErr error
	for _, s := range sinks {
		if err := s.Emit(e); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// History returns all events published (newest first).
func (b *EventBus) History() []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]Event, len(b.history))
	copy(out, b.history)
	reverse(out)
	return out
}

// SinkCount returns the number of registered sinks.
func (b *EventBus) SinkCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.sinks)
}

// MemorySink is an in-memory event sink useful for tests and local runs.
type MemorySink struct {
	mu     sync.Mutex
	events []Event
}

// NewMemorySink creates an in-memory event sink.
func NewMemorySink() *MemorySink {
	return &MemorySink{}
}

// Emit records the event.
func (m *MemorySink) Emit(e Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return nil
}

// Events returns all received events.
func (m *MemorySink) Events() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Event, len(m.events))
	copy(out, m.events)
	return out
}

// Count returns the number of received events.
func (m *MemorySink) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func reverse(s []Event) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
