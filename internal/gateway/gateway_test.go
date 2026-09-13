// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter()

	if !rl.Allow("user1", 5, time.Minute) {
		t.Error("expected first request to be allowed")
	}

	for i := 0; i < 4; i++ {
		rl.Allow("user1", 5, time.Minute)
	}

	if rl.Allow("user1", 5, time.Minute) {
		t.Error("expected request to be blocked")
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := NewRateLimiter()

	for i := 0; i < 5; i++ {
		rl.Allow("user1", 5, time.Minute)
	}

	rl.Reset("user1")

	if !rl.Allow("user1", 5, time.Minute) {
		t.Error("expected request to be allowed after reset")
	}
}

func TestRateLimiterGetUsage(t *testing.T) {
	rl := NewRateLimiter()

	rl.Allow("user1", 5, time.Minute)
	rl.Allow("user1", 5, time.Minute)

	usage, ok := rl.GetUsage("user1")
	if !ok {
		t.Error("expected usage to be found")
	}
	if usage != 2 {
		t.Errorf("expected usage 2, got %d", usage)
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, time.Second)

	if cb.State() != CircuitClosed {
		t.Error("expected closed state")
	}

	if !cb.Allow() {
		t.Error("expected to allow requests")
	}
}

func TestCircuitBreakerOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, time.Second)

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Error("expected open state")
	}

	if cb.Allow() {
		t.Error("expected to block requests")
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, 10*time.Millisecond)

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	time.Sleep(20 * time.Millisecond)

	if !cb.Allow() {
		t.Error("expected to allow requests after timeout")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, time.Second)

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Error("expected closed state after reset")
	}
}

func TestLoadBalancer(t *testing.T) {
	lb := NewLoadBalancer()

	lb.AddBackend(&Backend{Name: "b1", URL: "http://b1:8080", Healthy: true})
	lb.AddBackend(&Backend{Name: "b2", URL: "http://b2:8080", Healthy: true})
	lb.AddBackend(&Backend{Name: "b3", URL: "http://b3:8080", Healthy: true})

	b := lb.Next()
	if b == nil {
		t.Fatal("expected backend")
	}
	if b.Name != "b2" {
		t.Errorf("expected b2, got %s", b.Name)
	}

	b = lb.Next()
	if b.Name != "b3" {
		t.Errorf("expected b3, got %s", b.Name)
	}
}

func TestLoadBalancerNoHealthy(t *testing.T) {
	lb := NewLoadBalancer()

	lb.AddBackend(&Backend{Name: "b1", Healthy: false})

	b := lb.Next()
	if b != nil {
		t.Error("expected nil for no healthy backends")
	}
}

func TestLoadBalancerRemove(t *testing.T) {
	lb := NewLoadBalancer()

	lb.AddBackend(&Backend{Name: "b1", Healthy: true})
	lb.AddBackend(&Backend{Name: "b2", Healthy: true})

	lb.RemoveBackend("b1")

	backends := lb.List()
	if len(backends) != 1 {
		t.Errorf("expected 1 backend, got %d", len(backends))
	}
}

func TestGateway(t *testing.T) {
	g := New()

	g.AddLoadBalancer("api", &LoadBalancer{})
	lb, ok := g.GetLoadBalancer("api")
	if !ok {
		t.Fatal("expected load balancer")
	}

	lb.AddBackend(&Backend{Name: "b1", URL: "http://b1:8080", Healthy: true})

	req := &Request{
		ID:       "req1",
		ClientID: "client1",
		Service:  "api",
		Path:     "/test",
	}

	resp, err := g.Route(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGatewayRateLimit(t *testing.T) {
	g := New()

	for i := 0; i < 100; i++ {
		g.RateLimiter().Allow("client1", 100, time.Minute)
	}

	req := &Request{
		ID:       "req1",
		ClientID: "client1",
		Service:  "api",
	}

	_, err := g.Route(req)
	if err == nil {
		t.Error("expected rate limit error")
	}
}

func TestGatewayCircuitBreaker(t *testing.T) {
	g := New()

	cb := NewCircuitBreaker("api", 3, 2, time.Second)
	g.AddCircuitBreaker("api", cb)

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	req := &Request{
		ID:       "req1",
		ClientID: "client1",
		Service:  "api",
	}

	_, err := g.Route(req)
	if err == nil {
		t.Error("expected circuit breaker error")
	}
}
