package websocket

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckOriginAllowed(t *testing.T) {
	s := NewServer()
	s.SetAllowedOrigins([]string{"http://example.com", "*"})
	r1 := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	r1.Header.Set("Origin", "http://example.com")
	if !s.checkOrigin(r1) {
		t.Error("expected allowed origin to pass")
	}
}

func TestCheckOriginDenied(t *testing.T) {
	s := NewServer()
	s.SetAllowedOrigins([]string{"http://example.com"})
	r := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	r.Header.Set("Origin", "http://evil.com")
	if s.checkOrigin(r) {
		t.Error("expected disallowed origin to fail")
	}
}

func TestNewHistoryWithZero(t *testing.T) {
	h := NewHistory(0)
	if h.maxSize != 100 {
		t.Errorf("expected maxSize 100, got %d", h.maxSize)
	}
}

func TestAuthMiddlewareUnauthorized(t *testing.T) {
	a := NewAuthMiddleware(func(r *http.Request) (string, error) {
		return "", fmt.Errorf("no token")
	})
	h := a.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), "GET", "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMessageMarshalJSON(t *testing.T) {
	msg := &Message{Type: "test", Payload: "hello"}
	data, err := msg.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty json")
	}
}

func TestReplayToClientWithZero(t *testing.T) {
	h := NewHistory(10)
	h.Add("event", "data1", "")
	h.Add("event", "data2", "")
	s := NewServer()
	c := makeTestClient(s, 10)
	h.ReplayToClient(c, 0)
	select {
	case <-c.send:
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}
}

func TestCheckOriginNoOrigins(t *testing.T) {
	s := NewServer()
	r := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	r.Header.Set("Origin", "http://example.com")
	if !s.checkOrigin(r) {
		t.Error("expected true when no origins set")
	}
}

func TestGetOrCreateExisting(t *testing.T) {
	rm := NewRoomManager()
	rm.GetOrCreate("foo")
	r2 := rm.GetOrCreate("foo")
	if r2.Name != "foo" {
		t.Errorf("expected foo, got %s", r2.Name)
	}
}
