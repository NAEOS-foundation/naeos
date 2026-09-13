// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"net/http"
	"testing"
	"time"
)

func TestRoomManager(t *testing.T) {
	rm := NewRoomManager()
	r := rm.GetOrCreate("test")
	if r.Name != "test" {
		t.Errorf("expected room name 'test', got %s", r.Name)
	}
	if rm.ClientCount("test") != 0 {
		t.Errorf("expected 0 clients, got %d", rm.ClientCount("test"))
	}
}

func TestRoomJoinLeave(t *testing.T) {
	r := &Room{Name: "r1", clients: make(map[*Client]bool)}
	c := makeTestClient(NewServer(), 10)
	r.Join(c)
	if r.ClientCount() != 1 {
		t.Errorf("expected 1, got %d", r.ClientCount())
	}
	r.Leave(c)
	if r.ClientCount() != 0 {
		t.Errorf("expected 0, got %d", r.ClientCount())
	}
}

func TestRoomBroadcast(t *testing.T) {
	r := &Room{Name: "r1", clients: make(map[*Client]bool)}
	c := makeTestClient(NewServer(), 10)
	r.Join(c)

	r.Broadcast("test", map[string]string{"msg": "hello"})
	select {
	case msg := <-c.send:
		if len(msg) == 0 {
			t.Error("expected non-empty message")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}
}

func TestRoomBroadcastExcept(t *testing.T) {
	r := &Room{Name: "r1", clients: make(map[*Client]bool)}
	c1 := makeTestClient(NewServer(), 10)
	c2 := makeTestClient(NewServer(), 10)
	r.Join(c1)
	r.Join(c2)

	r.BroadcastExcept("test", "data", c1)

	select {
	case <-c1.send:
		t.Error("c1 should not receive message")
	case <-c2.send:
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}
}

func TestRoomManagerList(t *testing.T) {
	rm := NewRoomManager()
	rm.GetOrCreate("a")
	rm.GetOrCreate("b")
	names := rm.List()
	if len(names) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(names))
	}
}

func TestRoomManagerDelete(t *testing.T) {
	rm := NewRoomManager()
	rm.GetOrCreate("x")
	rm.Delete("x")
	_, ok := rm.Get("x")
	if ok {
		t.Error("expected room to be deleted")
	}
}

func TestServerJoinLeaveRoom(t *testing.T) {
	s := NewServer()
	go s.Run()
	c := makeTestClient(s, 10)
	s.register <- c
	time.Sleep(25 * time.Millisecond)

	s.JoinRoom("lobby", c)
	if s.RoomManager().ClientCount("lobby") != 1 {
		t.Errorf("expected 1 in lobby, got %d", s.RoomManager().ClientCount("lobby"))
	}

	s.LeaveRoom("lobby", c)
	if s.RoomManager().ClientCount("lobby") != 0 {
		t.Errorf("expected 0 in lobby, got %d", s.RoomManager().ClientCount("lobby"))
	}
}

func TestServerBroadcastToRoom(t *testing.T) {
	s := NewServer()
	go s.Run()
	c := makeTestClient(s, 10)
	s.register <- c
	time.Sleep(25 * time.Millisecond)

	s.JoinRoom("news", c)
	<-c.send

	s.BroadcastToRoom("news", "update", "content")
	select {
	case <-c.send:
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}
}

func TestHistory(t *testing.T) {
	h := NewHistory(5)
	h.Add("event1", "payload1", "room1")
	h.Add("event2", "payload2", "room1")
	h.Add("event3", "payload3", "room2")

	if h.Len() != 3 {
		t.Errorf("expected 3, got %d", h.Len())
	}

	recent := h.Recent(2)
	if len(recent) != 2 {
		t.Errorf("expected 2, got %d", len(recent))
	}

	filtered := h.FilterByRoom("room1")
	if len(filtered) != 2 {
		t.Errorf("expected 2 room1 events, got %d", len(filtered))
	}

	typed := h.FilterByType("event2")
	if len(typed) != 1 {
		t.Errorf("expected 1 event2, got %d", len(typed))
	}
}

func TestHistoryMaxSize(t *testing.T) {
	h := NewHistory(3)
	for i := 0; i < 5; i++ {
		h.Add("event", i, "")
	}
	if h.Len() != 3 {
		t.Errorf("expected 3, got %d", h.Len())
	}
}

func TestHistoryClear(t *testing.T) {
	h := NewHistory(10)
	h.Add("event", nil, "")
	h.Clear()
	if h.Len() != 0 {
		t.Errorf("expected 0, got %d", h.Len())
	}
}

func TestHistorySince(t *testing.T) {
	h := NewHistory(10)
	h.Add("event", nil, "")
	time.Sleep(10 * time.Millisecond)
	cutoff := time.Now()
	h.Add("event", nil, "")
	since := h.Since(cutoff)
	if len(since) != 1 {
		t.Errorf("expected 1, got %d", len(since))
	}
}

func TestHistoryReplayToClient(t *testing.T) {
	h := NewHistory(10)
	h.Add("event", "data1", "")
	h.Add("event", "data2", "")

	s := NewServer()
	c := makeTestClient(s, 10)
	h.ReplayToClient(c, 2)

	count := 0
	for i := 0; i < 2; i++ {
		select {
		case <-c.send:
			count++
		case <-time.After(time.Second):
			break
		}
	}
	if count != 2 {
		t.Errorf("expected 2 replayed, got %d", count)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)
	if !rl.Allow("client1") {
		t.Error("expected first request allowed")
	}
	if !rl.Allow("client1") {
		t.Error("expected second request allowed")
	}
	if rl.Allow("client1") {
		t.Error("expected third request blocked")
	}
	if !rl.Allow("client2") {
		t.Error("expected other client allowed")
	}
}

func TestRateLimiterRemaining(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)
	if rl.Remaining("c1") != 3 {
		t.Errorf("expected 3 remaining, got %d", rl.Remaining("c1"))
	}
	rl.Allow("c1")
	if rl.Remaining("c1") != 2 {
		t.Errorf("expected 2 remaining, got %d", rl.Remaining("c1"))
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	rl := NewRateLimiter(1, time.Millisecond)
	rl.Allow("c1")
	time.Sleep(5 * time.Millisecond)
	rl.Cleanup()
	if _, ok := rl.limits["c1"]; ok {
		t.Error("expected cleanup to remove expired entries")
	}
}

func TestMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()
	mc.IncrSent(5)
	mc.IncrReceived(3)
	mc.IncrConnected(1)
	mc.IncrErrors(2)

	snap := mc.Snapshot()
	if snap["messages_sent"] != 5 {
		t.Errorf("expected 5 sent, got %d", snap["messages_sent"])
	}
	if snap["messages_received"] != 3 {
		t.Errorf("expected 3 received, got %d", snap["messages_received"])
	}
	if snap["clients_connected"] != 1 {
		t.Errorf("expected 1 connected, got %d", snap["clients_connected"])
	}
	if snap["errors_count"] != 2 {
		t.Errorf("expected 2 errors, got %d", snap["errors_count"])
	}
}

func TestServerWithRateLimit(t *testing.T) {
	s := NewServer().WithRateLimit(10, time.Second)
	if s.rateLimiter == nil {
		t.Error("expected rate limiter to be set")
	}
}

func TestServerWithAuth(t *testing.T) {
	s := NewServer().WithAuth(func(r *http.Request) (string, error) {
		return "client", nil
	})
	if s.authFunc == nil {
		t.Error("expected auth func to be set")
	}
}

func TestServerClientIDs(t *testing.T) {
	s := NewServer()
	go s.Run()
	c1 := makeTestClient(s, 10)
	c2 := makeTestClient(s, 10)
	s.register <- c1
	s.register <- c2
	time.Sleep(25 * time.Millisecond)

	ids := s.ClientIDs()
	if len(ids) != 2 {
		t.Errorf("expected 2 IDs, got %d", len(ids))
	}
}

func TestServerHistoryAndMetrics(t *testing.T) {
	s := NewServer()
	if s.History() == nil {
		t.Error("expected history to be non-nil")
	}
	if s.Metrics() == nil {
		t.Error("expected metrics to be non-nil")
	}
}

func TestServerAddInterceptor(t *testing.T) {
	s := NewServer()
	s.AddInterceptor(func(clientID string, msg *Message) bool {
		return msg.Type != "blocked"
	})
	if len(s.interceptors) != 1 {
		t.Errorf("expected 1 interceptor, got %d", len(s.interceptors))
	}
}

func TestClientRateInfo(t *testing.T) {
	s := NewServer()
	info := s.ClientRateInfo("c1")
	if info.Remaining != -1 {
		t.Errorf("expected -1 when no limiter, got %d", info.Remaining)
	}

	s.WithRateLimit(5, time.Second)
	info = s.ClientRateInfo("c1")
	if info.Remaining != 5 {
		t.Errorf("expected 5 remaining, got %d", info.Remaining)
	}
}

func TestHistoryAddReturnsEntry(t *testing.T) {
	h := NewHistory(10)
	entry := h.Add("evt", "pl", "room")
	if entry.Type != "evt" {
		t.Errorf("expected type 'evt', got %s", entry.Type)
	}
	if entry.Room != "room" {
		t.Errorf("expected room 'room', got %s", entry.Room)
	}
	if entry.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestHistoryRecentEmpty(t *testing.T) {
	h := NewHistory(10)
	recent := h.Recent(5)
	if len(recent) != 0 {
		t.Errorf("expected 0, got %d", len(recent))
	}
}

func TestRoomClients(t *testing.T) {
	r := &Room{Name: "r1", clients: make(map[*Client]bool)}
	c1 := makeTestClient(NewServer(), 10)
	c2 := makeTestClient(NewServer(), 10)
	r.Join(c1)
	r.Join(c2)
	clients := r.Clients()
	if len(clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(clients))
	}
}

func TestHistoryReplayRoomToClient(t *testing.T) {
	h := NewHistory(10)
	h.Add("evt", "d1", "roomA")
	h.Add("evt", "d2", "roomB")
	h.Add("evt", "d3", "roomA")

	s := NewServer()
	c := makeTestClient(s, 10)
	h.ReplayRoomToClient(c, "roomA", 10)

	count := 0
	for i := 0; i < 2; i++ {
		select {
		case <-c.send:
			count++
		case <-time.After(time.Second):
			break
		}
	}
	if count != 2 {
		t.Errorf("expected 2 replayed, got %d", count)
	}
}
