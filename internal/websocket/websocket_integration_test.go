//go:build integration

package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// startHTTPTestServer wires a Server into an httptest.Server and returns both.
func startHTTPTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	s := NewServer()
	go s.Run()
	time.Sleep(15 * time.Millisecond)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.HandleWebSocket(w, r)
	}))
	return s, ts
}

func dialWS(t *testing.T, ts *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + ts.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial WebSocket: %v", err)
	}
	return conn
}

func TestHandleWebSocketUpgrade(t *testing.T) {
	s, ts := startHTTPTestServer(t)
	defer ts.Close()

	conn := dialWS(t, ts)
	defer conn.Close()

	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Errorf("expected 1 client after upgrade, got %d", s.ClientCount())
	}
}

func TestReadWritePumpPingMessage(t *testing.T) {
	_, ts := startHTTPTestServer(t)
	defer ts.Close()

	conn := dialWS(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	// Drain the system message sent on registration.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage()

	// Send a ping message; readPump should reply with pong.
	pingMsg, _ := json.Marshal(Message{Type: "ping"})
	if err := conn.WriteMessage(websocket.TextMessage, pingMsg); err != nil {
		t.Fatalf("send ping: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, pongData, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read pong: %v", err)
	}

	var pong Message
	if err := json.Unmarshal(pongData, &pong); err != nil {
		t.Fatalf("unmarshal pong: %v", err)
	}
	if pong.Type != "pong" {
		t.Errorf("expected type 'pong', got %s", pong.Type)
	}
}

func TestReadWritePumpBroadcast(t *testing.T) {
	s, ts := startHTTPTestServer(t)
	defer ts.Close()

	conn := dialWS(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	// Drain the system message sent on registration.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatalf("drain system message: %v", err)
	}

	// Broadcast an event and read it from the WebSocket client.
	s.Broadcast("test.event", map[string]string{"data": "hello"})

	_, eventData, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	var event Message
	if err := json.Unmarshal(eventData, &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if event.Type != "test.event" {
		t.Errorf("expected type 'test.event', got %s", event.Type)
	}
}

func TestStopWithConnectedClients(t *testing.T) {
	s, ts := startHTTPTestServer(t)
	defer ts.Close()

	conn := dialWS(t, ts)
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", s.ClientCount())
	}

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop took too long with connected client")
	}

	conn.Close()
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 0 {
		t.Errorf("expected 0 clients after stop, got %d", s.ClientCount())
	}
}

func TestClientReadPumpInvalidJSON(t *testing.T) {
	s, ts := startHTTPTestServer(t)
	defer ts.Close()

	conn := dialWS(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	// Drain system message.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage()

	// Send invalid JSON; readPump silently ignores it.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("not json")); err != nil {
		t.Fatalf("send invalid JSON: %v", err)
	}
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Errorf("expected 1 client (invalid JSON ignored), got %d", s.ClientCount())
	}
}

func TestClientReadPumpNonPingMessage(t *testing.T) {
	s, ts := startHTTPTestServer(t)
	defer ts.Close()

	conn := dialWS(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	// Drain system message.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage()

	// Send a valid JSON message with an unknown type (not "ping").
	unknownMsg, _ := json.Marshal(Message{Type: "unknown", Payload: "data"})
	if err := conn.WriteMessage(websocket.TextMessage, unknownMsg); err != nil {
		t.Fatalf("send unknown type: %v", err)
	}
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Errorf("expected 1 client (non-ping ignored), got %d", s.ClientCount())
	}
}
