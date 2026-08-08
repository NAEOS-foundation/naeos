package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// startHTTPTestServerNonTagged wires a Server into an httptest.Server. It
// mirrors the integration-tagged helpers so the HTTP upgrade / pump paths
// are covered by the default test run.
func startHTTPTestServerNonTagged(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	s := NewServer()
	go s.Run()
	time.Sleep(15 * time.Millisecond)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.HandleWebSocket(w, r)
	}))
	return s, ts
}

func dialWSNonTagged(t *testing.T, ts *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + ts.URL[len("http"):]
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial WebSocket: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
	return conn
}

func TestHTTPUpgradeSuccess(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	conn := dialWSNonTagged(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Errorf("expected 1 client after upgrade, got %d", s.ClientCount())
	}
	if len(s.ClientIDs()) != 1 {
		t.Errorf("expected 1 client ID, got %d", len(s.ClientIDs()))
	}
}

func TestHTTPUpgradeFailure(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("plain GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSwitchingProtocols {
		t.Error("expected non-upgraded response for plain GET")
	}
	time.Sleep(25 * time.Millisecond)
	if s.ClientCount() != 0 {
		t.Errorf("expected 0 clients after failed upgrade, got %d", s.ClientCount())
	}
}

func TestHTTPReadPumpPingPong(t *testing.T) {
	_, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	conn := dialWSNonTagged(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage() // drain system message

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

func TestHTTPReadPumpInvalidJSON(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	conn := dialWSNonTagged(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage() // drain system message

	if err := conn.WriteMessage(websocket.TextMessage, []byte("not json")); err != nil {
		t.Fatalf("send invalid JSON: %v", err)
	}
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Errorf("expected 1 client (invalid JSON ignored), got %d", s.ClientCount())
	}
}

func TestHTTPReadPumpUnknownType(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	conn := dialWSNonTagged(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage() // drain system message

	unknownMsg, _ := json.Marshal(Message{Type: "unknown", Payload: "data"})
	if err := conn.WriteMessage(websocket.TextMessage, unknownMsg); err != nil {
		t.Fatalf("send unknown type: %v", err)
	}
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Errorf("expected 1 client (non-ping ignored), got %d", s.ClientCount())
	}
}

func TestHTTPReadPumpInterceptorBlock(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	blocked := false
	s.AddInterceptor(func(clientID string, msg *Message) bool {
		return false
	})

	conn := dialWSNonTagged(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage() // drain system message

	pingMsg, _ := json.Marshal(Message{Type: "ping"})
	if err := conn.WriteMessage(websocket.TextMessage, pingMsg); err != nil {
		t.Fatalf("send ping: %v", err)
	}

	// The blocking interceptor must prevent the pong from being sent.
	conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	conn.SetReadLimit(0)
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Error("expected message blocked by interceptor")
	}
	if blocked {
		t.Error("unexpected interceptor call state")
	}
	if s.ClientCount() != 1 {
		t.Errorf("expected 1 client after blocked message, got %d", s.ClientCount())
	}
}

func TestHTTPWritePumpBroadcast(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	conn := dialWSNonTagged(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage() // drain system message

	s.Broadcast("test.event", map[string]string{"data": "hello"})

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
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

func TestHTTPStopConnectedClient(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	conn := dialWSNonTagged(t, ts)
	defer conn.Close()
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", s.ClientCount())
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn.ReadMessage() // drain system message

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

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Error("expected close frame after Stop")
	}

	time.Sleep(25 * time.Millisecond)
	if s.ClientCount() != 0 {
		t.Errorf("expected 0 clients after stop, got %d", s.ClientCount())
	}
}

func TestHTTPReadPumpDisconnectUnregisters(t *testing.T) {
	s, ts := startHTTPTestServerNonTagged(t)
	defer ts.Close()

	conn := dialWSNonTagged(t, ts)
	time.Sleep(25 * time.Millisecond)

	if s.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", s.ClientCount())
	}

	conn.Close()
	time.Sleep(50 * time.Millisecond)

	if s.ClientCount() != 0 {
		t.Errorf("expected 0 clients after disconnect, got %d", s.ClientCount())
	}
}
