package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAPIKeyRateLimitAllowed(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.RegisterAPIKey("test-key-123", 10)

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAPIKeyRateLimitExceeded(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.RegisterAPIKey("limited-key", 2)

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
		req.Header.Set("X-API-Key", "limited-key")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if i < 2 && w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, w.Code)
		}
		if i == 2 && w.Code != http.StatusTooManyRequests {
			t.Errorf("request 2: expected 429, got %d", w.Code)
		}
	}
}

func TestFallbackToIPBasedLimiter(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUnknownAPIKeyFallsBackToIP(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.RegisterAPIKey("known-key", 5)

	handler := s.handlerWithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/v1/health", nil)
	req.Header.Set("X-API-Key", "unknown-key-value")
	req.RemoteAddr = "10.0.0.1:9999"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for unknown API key fallback, got %d", w.Code)
	}

	_ = time.Now()
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	rl.Allow("client1")
	rl.Allow("client2")
	rl.Reset()

	// After reset, client should have full allowance again
	if !rl.Allow("client1") {
		t.Error("expected client1 to be allowed after reset")
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	// Stop should not panic
	rl.Stop()
}

func TestRateLimiter_MiddlewareAllows(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiter_MiddlewareBlocks(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request should pass
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Second should be blocked
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}

func TestRateLimiter_MiddlewareForwardedFor(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Same X-Forwarded-For should be blocked
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}

func TestStatusRecorder(t *testing.T) {
	w := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	rec.WriteHeader(http.StatusNotFound)

	if rec.status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.status)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected response status 404, got %d", w.Code)
	}
}

func TestMaxBytesBody_Read(t *testing.T) {
	exceeded := false
	body := &maxBytesBody{
		ReadCloser: io.NopCloser(strings.NewReader("hello")),
		exceeded:   &exceeded,
	}

	p := make([]byte, 10)
	n, err := body.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("unexpected error, got %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if exceeded {
		t.Error("expected exceeded to be false")
	}
}

func TestMaxBytesBody_ReadExceeded(t *testing.T) {
	exceeded := false
	body := &maxBytesBody{
		ReadCloser: io.NopCloser(&exceededReader{}),
		exceeded:   &exceeded,
	}

	p := make([]byte, 10)
	_, _ = body.Read(p)
	if !exceeded {
		t.Error("expected exceeded to be true after MaxBytesError")
	}
}

type exceededReader struct{}

func (e *exceededReader) Read(p []byte) (int, error) {
	return 0, &http.MaxBytesError{Limit: 10}
}

func (e *exceededReader) Close() error {
	return nil
}

func TestMaxBytesResponseWriter_WriteHeader(t *testing.T) {
	exceeded := false
	w := httptest.NewRecorder()
	mw := &maxBytesResponseWriter{ResponseWriter: w, exceeded: &exceeded}

	mw.WriteHeader(http.StatusOK)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	exceeded = true
	mw2 := &maxBytesResponseWriter{ResponseWriter: httptest.NewRecorder(), exceeded: &exceeded}
	mw2.WriteHeader(http.StatusOK)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from original recorder, got %d", w.Code)
	}
}

func TestMaxBytesResponseWriter_Unwrap(t *testing.T) {
	w := httptest.NewRecorder()
	mw := &maxBytesResponseWriter{ResponseWriter: w, exceeded: new(bool)}

	unwrapped := mw.Unwrap()
	if unwrapped != w {
		t.Error("expected Unwrap to return the original ResponseWriter")
	}
}
