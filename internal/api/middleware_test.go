package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)

	if !rl.Allow("client1") {
		t.Error("expected first request allowed")
	}
	if !rl.Allow("client1") {
		t.Error("expected second request allowed")
	}
	if !rl.Allow("client1") {
		t.Error("expected third request allowed")
	}
	if rl.Allow("client1") {
		t.Error("expected fourth request denied")
	}

	if !rl.Allow("client2") {
		t.Error("expected different client allowed")
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	rl.Allow("client1")
	rl.Allow("client1")

	if rl.Allow("client1") {
		t.Error("should be rate limited")
	}

	rl.Reset()

	if !rl.Allow("client1") {
		t.Error("should be allowed after reset")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	rl := NewRateLimiter(2, 10*time.Millisecond)

	rl.Allow("client1")
	rl.Allow("client1")

	if rl.Allow("client1") {
		t.Error("should be rate limited")
	}

	time.Sleep(15 * time.Millisecond)

	if !rl.Allow("client1") {
		t.Error("should be allowed after refill")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w3.Code)
	}

	retryAfter := w3.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header")
	}
}

func TestRateLimiterXForwardedFor(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}
