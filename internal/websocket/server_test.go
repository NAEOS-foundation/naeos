package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Parallel()
	s := NewServer()
	if s == nil {
		t.Fatal("expected server to be created")
	}
}

func TestMessageSerialization(t *testing.T) {
	t.Parallel()
	msg := Message{
		Type:    "test",
		Payload: map[string]string{"key": "value"},
		Time:    time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if decoded.Type != "test" {
		t.Errorf("expected type 'test', got %s", decoded.Type)
	}
}

func TestBroadcast(t *testing.T) {
	t.Parallel()
	s := NewServer()
	go s.Run()

	s.Broadcast("test", map[string]string{"message": "hello"})

	if s.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", s.ClientCount())
	}
}

func TestEventBroadcaster(t *testing.T) {
	t.Parallel()
	s := NewServer()
	broadcaster := NewEventBroadcaster(s)
	go s.Run()

	broadcaster.PipelineStarted("pipeline-123")
	broadcaster.PipelineCompleted("pipeline-123", "10s")
	broadcaster.PipelineFailed("pipeline-123", "error")
	broadcaster.SpecValidated(true, []string{})
	broadcaster.ArtifactGenerated("main.go", "./cmd/main.go")
	broadcaster.LogMessage("info", "test message")
}

func TestGenerateID(t *testing.T) {
	t.Parallel()
	id1 := generateID()
	id2 := generateID()

	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}

func TestStop(t *testing.T) {
	t.Parallel()
	s := NewServer()
	go s.Run()
	s.Stop()
}

// --- test helpers ---

// startTestRunWithServer creates a Server, starts Run() in a goroutine,
// and sleeps briefly to let the select loop initialize.
func startTestRunWithServer() *Server {
	s := NewServer()
	go s.Run()
	time.Sleep(15 * time.Millisecond)
	return s
}

// makeTestClient creates a Client with the given send-channel capacity.
// The conn field is left nil; tests that only exercise register / unregister /
// broadcast paths never touch it.
func makeTestClient(s *Server, sendCap int) *Client {
	return &Client{
		server: s,
		send:   make(chan []byte, sendCap),
		id:     generateID(),
	}
}

// --- new tests ---

func TestServerRegisterAndUnregister(t *testing.T) {
	t.Parallel()
	s := startTestRunWithServer()

	c := makeTestClient(s, 256)
	s.register <- c
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Fatalf("expected 1 client after register, got %d", s.ClientCount())
	}

	s.unregister <- c
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after unregister, got %d", s.ClientCount())
	}
}

func TestServerBroadcastToClients(t *testing.T) {
	t.Parallel()
	s := startTestRunWithServer()

	c := makeTestClient(s, 256)
	s.register <- c
	time.Sleep(25 * time.Millisecond)

	// Drain the system message that was broadcast on registration.
	<-c.send

	// Broadcast a custom event.
	s.Broadcast("test.event", map[string]string{"data": "hello"})

	select {
	case msg := <-c.send:
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			t.Fatalf("failed to unmarshal broadcast message: %v", err)
		}
		if m.Type != "test.event" {
			t.Errorf("expected message type 'test.event', got %s", m.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for broadcast message")
	}
}

func TestServerBroadcastFullChannel(t *testing.T) {
	t.Parallel()
	s := startTestRunWithServer()

	// Create a client whose send channel can hold only 1 message.
	c := makeTestClient(s, 1)
	s.register <- c
	time.Sleep(25 * time.Millisecond)

	// The registration triggers sendSystemMessage which fills the
	// capacity-1 send channel.  Broadcast another message; the Run
	// loop should hit the default case and remove the client.
	s.Broadcast("overflow", "trigger")
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 0 {
		t.Errorf("expected 0 clients after full-channel broadcast, got %d",
			s.ClientCount())
	}
}

func TestStopWithNoClients(t *testing.T) {
	t.Parallel()
	s := NewServer()
	go s.Run()

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Stop returned promptly – good.
	case <-time.After(2 * time.Second):
		t.Fatal("Stop took too long with no clients")
	}
}

func TestGenerateIDUniqueness(t *testing.T) {
	t.Parallel()
	const count = 100
	ids := make(map[string]bool, count)
	for i := 0; i < count; i++ {
		id := generateID()
		if ids[id] {
			t.Fatalf("duplicate ID generated on iteration %d: %s", i, id)
		}
		ids[id] = true
		// Tiny sleep to guarantee distinct nanosecond timestamps.
		time.Sleep(time.Microsecond)
	}
}

func TestEventBroadcasterAllMethods(t *testing.T) {
	t.Parallel()
	s := startTestRunWithServer()
	b := NewEventBroadcaster(s)

	c := makeTestClient(s, 256)
	s.register <- c
	time.Sleep(25 * time.Millisecond)
	<-c.send // drain system message

	tests := []struct {
		name string
		fn   func()
		typ  string
	}{
		{
			name: "PipelineStarted",
			fn:   func() { b.PipelineStarted("p1") },
			typ:  "pipeline.started",
		},
		{
			name: "PipelineCompleted",
			fn:   func() { b.PipelineCompleted("p1", "10s") },
			typ:  "pipeline.completed",
		},
		{
			name: "PipelineFailed",
			fn:   func() { b.PipelineFailed("p1", "err") },
			typ:  "pipeline.failed",
		},
		{
			name: "SpecValidated",
			fn:   func() { b.SpecValidated(true, []string{"e1"}) },
			typ:  "spec.validated",
		},
		{
			name: "ArtifactGenerated",
			fn:   func() { b.ArtifactGenerated("main.go", "./cmd/main.go") },
			typ:  "artifact.generated",
		},
		{
			name: "LogMessage",
			fn:   func() { b.LogMessage("info", "hello") },
			typ:  "log",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.fn()
			select {
			case msg := <-c.send:
				var m Message
				if err := json.Unmarshal(msg, &m); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if m.Type != tc.typ {
					t.Errorf("expected type %q, got %q", tc.typ, m.Type)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("timed out waiting for %s", tc.name)
			}
		})
	}
}

func TestWSObserverAllMethods(t *testing.T) {
	t.Parallel()
	s := startTestRunWithServer()
	b := NewEventBroadcaster(s)
	o := NewWSObserver(b)

	c := makeTestClient(s, 256)
	s.register <- c
	time.Sleep(25 * time.Millisecond)
	<-c.send // drain system message

	tests := []struct {
		name string
		fn   func()
		typ  string
	}{
		{
			name: "OnPipelineStart",
			fn:   func() { o.OnPipelineStart("p1") },
			typ:  "pipeline.started",
		},
		{
			name: "OnPipelineComplete",
			fn:   func() { o.OnPipelineComplete("p1", 5, "1.5s") },
			typ:  "pipeline.completed",
		},
		{
			name: "OnPipelineFailed",
			fn:   func() { o.OnPipelineFailed("p1", "err") },
			typ:  "pipeline.failed",
		},
		{
			name: "OnArtifactGenerated",
			fn:   func() { o.OnArtifactGenerated("main.go", "cmd/main.go") },
			typ:  "artifact.generated",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.fn()
			select {
			case msg := <-c.send:
				var m Message
				if err := json.Unmarshal(msg, &m); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if m.Type != tc.typ {
					t.Errorf("expected type %q, got %q", tc.typ, m.Type)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("timed out waiting for %s", tc.name)
			}
		})
	}
}
