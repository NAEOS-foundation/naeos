package api

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a token-bucket rate limiter per client.
type RateLimiter struct {
	clients    map[string]*clientRecord
	mu         sync.Mutex
	rate       int
	window     time.Duration
	maxClients int
	done       chan struct{}
}

type clientRecord struct {
	tokens   int
	lastSeen time.Time
}

const defaultMaxClients = 10000

// NewRateLimiter creates a rate limiter with the given requests per window.
func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:    make(map[string]*clientRecord),
		rate:       requestsPerWindow,
		window:     window,
		maxClients: defaultMaxClients,
		done:       make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Allow reports whether the given client has remaining rate limit tokens.
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	record, exists := rl.clients[clientID]
	if !exists {
		if len(rl.clients) >= rl.maxClients {
			return false
		}
		rl.clients[clientID] = &clientRecord{
			tokens:   rl.rate - 1,
			lastSeen: now,
		}
		return true
	}

	elapsed := now.Sub(record.lastSeen)
	refills := int(elapsed/rl.window) * rl.rate
	if refills > 0 {
		record.tokens += refills
		if record.tokens > rl.rate {
			record.tokens = rl.rate
		}
		record.lastSeen = now
	}

	if record.tokens <= 0 {
		return false
	}

	record.tokens--
	record.lastSeen = now
	return true
}

// Rate returns the maximum requests per window.
func (rl *RateLimiter) Rate() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.rate
}

// Remaining returns the number of remaining requests for the given client.
func (rl *RateLimiter) Remaining(clientID string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	record, exists := rl.clients[clientID]
	if !exists {
		return rl.rate
	}

	elapsed := now.Sub(record.lastSeen)
	refills := int(elapsed/rl.window) * rl.rate
	tokens := record.tokens + refills
	if tokens > rl.rate {
		tokens = rl.rate
	}
	return tokens
}

// ResetTime returns the timestamp when the rate limit window resets for the given client.
func (rl *RateLimiter) ResetTime(clientID string) time.Time {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	record, exists := rl.clients[clientID]
	if !exists {
		return time.Now()
	}
	return record.lastSeen.Add(rl.window)
}

// Reset clears all client records from the rate limiter.
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*clientRecord)
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for id, record := range rl.clients {
				if now.Sub(record.lastSeen) > rl.window*10 {
					delete(rl.clients, id)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Stop signals the cleanup goroutine to stop.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// Middleware wraps an http.Handler with per-client rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientID = forwarded
		}

		if !rl.Allow(clientID) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", rl.window.String())
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"success":false,"error":"rate limit exceeded"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		level := slog.LevelInfo
		if rec.status >= 500 {
			level = slog.LevelError
		} else if rec.status >= 400 {
			level = slog.LevelWarn
		}
		slog.LogAttrs(r.Context(), level, "request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.String("duration", duration.String()),
			slog.String("request_id", RequestIDFromContext(r.Context())),
			slog.String("component", "api-server"),
		)
	})
}

type maxBytesBody struct {
	io.ReadCloser
	exceeded *bool
}

func (b *maxBytesBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			*b.exceeded = true
		}
	}
	return n, err
}

type maxBytesResponseWriter struct {
	http.ResponseWriter
	exceeded *bool
}

func (mw *maxBytesResponseWriter) WriteHeader(code int) {
	if *mw.exceeded {
		code = http.StatusRequestEntityTooLarge
	}
	mw.ResponseWriter.WriteHeader(code)
}

func (mw *maxBytesResponseWriter) Unwrap() http.ResponseWriter {
	return mw.ResponseWriter
}
