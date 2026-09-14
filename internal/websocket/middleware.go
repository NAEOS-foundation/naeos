// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Middleware func(next http.Handler) http.Handler

type AuthFunc func(r *http.Request) (clientID string, err error)

type AuthMiddleware struct {
	authFunc AuthFunc
}

func NewAuthMiddleware(authFunc AuthFunc) *AuthMiddleware {
	return &AuthMiddleware{authFunc: authFunc}
}

func (a *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := a.authFunc(r); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type RateLimiter struct {
	limits map[string]*rateBucket
	mu     sync.Mutex
	rate   int
	window time.Duration
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*rateBucket),
		rate:   requestsPerWindow,
		window: window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, ok := rl.limits[key]
	if !ok || now.After(bucket.resetAt) {
		rl.limits[key] = &rateBucket{count: 1, resetAt: now.Add(rl.window)}
		return true
	}

	if bucket.count >= rl.rate {
		return false
	}
	bucket.count++
	return true
}

func (rl *RateLimiter) Remaining(key string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, ok := rl.limits[key]
	if !ok || time.Now().After(bucket.resetAt) {
		return rl.rate
	}
	remaining := rl.rate - bucket.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for k, v := range rl.limits {
		if now.After(v.resetAt) {
			delete(rl.limits, k)
		}
	}
}

func (s *Server) WithRateLimit(requestsPerWindow int, window time.Duration) *Server {
	s.rateLimiter = NewRateLimiter(requestsPerWindow, window)
	return s
}

func (s *Server) WithAuth(authFunc AuthFunc) *Server {
	s.authFunc = authFunc
	return s
}

type ClientRateInfo struct {
	Remaining int
	Limited   bool
}

func (s *Server) ClientRateInfo(clientID string) ClientRateInfo {
	if s.rateLimiter == nil {
		return ClientRateInfo{Remaining: -1, Limited: false}
	}
	return ClientRateInfo{
		Remaining: s.rateLimiter.Remaining(clientID),
		Limited:   !s.rateLimiter.Allow(clientID),
	}
}

type MessageInterceptor func(clientID string, msg *Message) bool

func (s *Server) AddInterceptor(interceptor MessageInterceptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interceptors = append(s.interceptors, interceptor)
}

type MetricsCollector struct {
	MessagesSent     int64
	MessagesReceived int64
	ClientsConnected int64
	ErrorsCount      int64
	mu               sync.RWMutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (mc *MetricsCollector) IncrSent(n int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.MessagesSent += n
}

func (mc *MetricsCollector) IncrReceived(n int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.MessagesReceived += n
}

func (mc *MetricsCollector) IncrConnected(n int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.ClientsConnected += n
}

func (mc *MetricsCollector) IncrErrors(n int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.ErrorsCount += n
}

func (mc *MetricsCollector) Snapshot() map[string]int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return map[string]int64{
		"messages_sent":     mc.MessagesSent,
		"messages_received": mc.MessagesReceived,
		"clients_connected": mc.ClientsConnected,
		"errors_count":      mc.ErrorsCount,
	}
}

func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(*m),
	})
}
