// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestHealthChecker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHealthChecker(time.Second, time.Second)
	backend := &Backend{
		Name: "test",
		URL:  server.URL,
	}

	if !checker.Check(backend) {
		t.Error("expected healthy backend")
	}
}

func TestHealthCheckerUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker := NewHealthChecker(time.Second, time.Second)
	backend := &Backend{
		Name: "test",
		URL:  server.URL,
	}

	if checker.Check(backend) {
		t.Error("expected unhealthy backend")
	}
}

func TestWeightedLoadBalancer(t *testing.T) {
	wlb := NewWeightedLoadBalancer()
	wlb.AddBackend(&Backend{Name: "a", URL: "http://a.com", Weight: 10, Healthy: true})
	wlb.AddBackend(&Backend{Name: "b", URL: "http://b.com", Weight: 5, Healthy: true})

	selections := make(map[string]int)
	for i := 0; i < 15; i++ {
		b := wlb.Next()
		if b == nil {
			t.Fatal("expected non-nil backend")
		}
		selections[b.Name]++
	}

	if selections["a"] <= selections["b"] {
		t.Errorf("expected more selections for a (weight 10) than b (weight 5), got a=%d b=%d", selections["a"], selections["b"])
	}
}

func TestWeightedLoadBalancerAllUnhealthy(t *testing.T) {
	wlb := NewWeightedLoadBalancer()
	wlb.AddBackend(&Backend{Name: "a", URL: "http://a.com", Weight: 10, Healthy: false})

	b := wlb.Next()
	if b != nil {
		t.Error("expected nil for all unhealthy backends")
	}
}

func TestReverseProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	proxy, err := NewReverseProxy(&Backend{URL: backend.URL}, nil)
	if err != nil {
		t.Fatalf("NewReverseProxy failed: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestReverseProxyUpdateBackend(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	proxy, err := NewReverseProxy(&Backend{URL: backend1.URL}, nil)
	if err != nil {
		t.Fatalf("NewReverseProxy failed: %v", err)
	}

	err = proxy.UpdateBackend(&Backend{URL: backend2.URL})
	if err != nil {
		t.Fatalf("UpdateBackend failed: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLoadBalancingProxy(t *testing.T) {
	var mu sync.Mutex
	counts := make(map[string]int)

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		counts["b1"]++
		mu.Unlock()
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		counts["b2"]++
		mu.Unlock()
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	lb := NewLoadBalancer()
	lb.AddBackend(&Backend{Name: "b1", URL: backend1.URL, Weight: 1, Healthy: true})
	lb.AddBackend(&Backend{Name: "b2", URL: backend2.URL, Weight: 1, Healthy: true})

	lbp := NewLoadBalancingProxy(lb)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		lbp.ServeHTTP(w, req)
	}

	if lbp.RequestCount() != 10 {
		t.Errorf("expected 10 requests, got %d", lbp.RequestCount())
	}
}

func TestLoadBalancingProxyNoBackends(t *testing.T) {
	lb := NewLoadBalancer()
	lbp := NewLoadBalancingProxy(lb)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	lbp.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestRequestLogger(t *testing.T) {
	var logged bool
	logger := NewRequestLogger(func(level, msg string, fields map[string]any) {
		logged = true
		if level != "info" {
			t.Errorf("expected info level, got %s", level)
		}
		if _, ok := fields["method"]; !ok {
			t.Error("expected method field")
		}
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	logger.LogRequest(req, http.StatusOK, time.Millisecond)
	if !logged {
		t.Error("expected log to be called")
	}
}

func TestProxyChain(t *testing.T) {
	chain := NewProxyChain(nil)
	chain.AddProxy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestProxyChainEmpty(t *testing.T) {
	chain := NewProxyChain(nil)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestBackendPool(t *testing.T) {
	pool := NewBackendPool(nil)
	b := pool.GetOrCreate("test", "http://test.com")
	if b == nil {
		t.Fatal("expected non-nil backend")
	}
	if b.Name != "test" {
		t.Errorf("expected name test, got %s", b.Name)
	}

	if pool.HealthyCount() != 1 {
		t.Errorf("expected 1 healthy backend, got %d", pool.HealthyCount())
	}

	pool.Remove("test")
	if pool.HealthyCount() != 0 {
		t.Errorf("expected 0 healthy backends, got %d", pool.HealthyCount())
	}
}

func TestBackendPoolList(t *testing.T) {
	pool := NewBackendPool(nil)
	pool.GetOrCreate("a", "http://a.com")
	pool.GetOrCreate("b", "http://b.com")

	list := pool.List()
	if len(list) != 2 {
		t.Errorf("expected 2 backends, got %d", len(list))
	}
}

func TestCopyBody(t *testing.T) {
	body, err := CopyBody(nil)
	if err != nil {
		t.Fatalf("CopyBody(nil) failed: %v", err)
	}
	if body != nil {
		t.Error("expected nil body")
	}
}
