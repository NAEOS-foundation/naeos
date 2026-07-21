package gateway

import (
	"testing"
	"time"
)

func TestCircuitBreakerRecordSuccessInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, time.Millisecond)

	// Force state to half-open
	cb.state = CircuitHalfOpen
	cb.successCount = 1

	cb.RecordSuccess()

	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after RecordSuccess reaches threshold, got %s", cb.State())
	}
}

func TestCircuitBreakerRecordSuccessInOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, time.Second)
	cb.state = CircuitOpen

	cb.RecordSuccess()

	if cb.failureCount != 0 {
		t.Error("expected failure count to be reset")
	}
}

func TestGatewayNoBackends(t *testing.T) {
	g := New()
	g.AddLoadBalancer("api", NewLoadBalancer())

	req := &Request{
		ID:       "req1",
		ClientID: "client1",
		Service:  "api",
	}
	_, err := g.Route(req)
	if err == nil {
		t.Error("expected error for no healthy backends")
	}
}

func TestGatewayNoService(t *testing.T) {
	g := New()
	req := &Request{
		ID:       "req1",
		ClientID: "client1",
		Service:  "nonexistent",
	}
	resp, err := g.Route(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRateLimiterBlockedDuration(t *testing.T) {
	rl := NewRateLimiter()
	rl.Allow("user1", 1, time.Minute)
	rl.Allow("user1", 1, time.Minute) // triggers block

	if rl.Allow("user1", 1, time.Minute) {
		t.Error("expected blocked request to remain blocked within window")
	}
}
