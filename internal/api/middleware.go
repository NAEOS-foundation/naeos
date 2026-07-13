package api

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	clients map[string]*clientRecord
	mu      sync.Mutex
	rate    int
	window  time.Duration
}

type clientRecord struct {
	tokens   int
	lastSeen time.Time
}

func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientRecord),
		rate:    requestsPerWindow,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	record, exists := rl.clients[clientID]
	if !exists {
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

func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*clientRecord)
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
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

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientID = forwarded
		}

		if !rl.Allow(clientID) {
			w.Header().Set("Retry-After", rl.window.String())
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
