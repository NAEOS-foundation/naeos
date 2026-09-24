// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package investordemo

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultControlPlaneRateLimit = 30
	defaultControlPlaneWindow    = time.Minute
)

type controlPlaneSecurity struct {
	allowedOrigins map[string]struct{}
	token          string
	limit          int
	window         time.Duration
	mu             sync.Mutex
	hits           map[string][]time.Time
}

func newControlPlaneSecurity(allowedOrigins string, token string) *controlPlaneSecurity {
	origins := map[string]struct{}{}
	for _, origin := range strings.Split(allowedOrigins, ",") {
		origin = strings.TrimSpace(strings.TrimRight(origin, "/"))
		if origin != "" {
			origins[origin] = struct{}{}
		}
	}
	return &controlPlaneSecurity{
		allowedOrigins: origins,
		token:          strings.TrimSpace(token),
		limit:          defaultControlPlaneRateLimit,
		window:         defaultControlPlaneWindow,
		hits:           make(map[string][]time.Time),
	}
}

func (s *controlPlaneSecurity) originAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if _, ok := s.allowedOrigins[origin]; ok {
		return true
	}
	return false
}

func (s *controlPlaneSecurity) authorize(r *http.Request) bool {
	if s.token == "" {
		return true
	}
	const prefix = "Bearer "
	return strings.HasPrefix(r.Header.Get("Authorization"), prefix) &&
		strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), prefix)) == s.token
}

func (s *controlPlaneSecurity) allowRate(ip string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	hits := s.hits[ip]
	cutoff := now.Add(-s.window)
	kept := hits[:0]
	for _, hit := range hits {
		if hit.After(cutoff) {
			kept = append(kept, hit)
		}
	}
	if len(kept) >= s.limit {
		s.hits[ip] = kept
		return false
	}
	s.hits[ip] = append(kept, now)
	return true
}

func requestClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}

func (s *controlPlaneSecurity) protectDecision(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if !s.originAllowed(origin) {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return false
	}
	if !s.authorize(r) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return false
	}
	if !s.allowRate(requestClientIP(r), time.Now().UTC()) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return false
	}
	return true
}

func (s *controlPlaneSecurity) corsOrigin(origin string) string {
	if origin == "" {
		return ""
	}
	if s.originAllowed(origin) {
		return origin
	}
	return ""
}
