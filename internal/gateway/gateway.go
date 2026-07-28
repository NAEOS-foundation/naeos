package gateway

import (
	"fmt"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Rate Limiter

type RateLimiter struct {
	limits map[string]*RateLimit
	mu     sync.RWMutex
}

type RateLimit struct {
	Key       string
	Limit     int
	Window    time.Duration
	Requests  []time.Time
	Blocked   bool
	BlockedAt time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*RateLimit),
	}
}

func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limitObj, exists := rl.limits[key]

	if !exists {
		rl.limits[key] = &RateLimit{
			Key:      key,
			Limit:    limit,
			Window:   window,
			Requests: []time.Time{now},
		}
		return true
	}

	if limitObj.Blocked && now.Sub(limitObj.BlockedAt) < window {
		return false
	}

	windowStart := now.Add(-window)
	valid := make([]time.Time, 0)
	for _, t := range limitObj.Requests {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		limitObj.Blocked = true
		limitObj.BlockedAt = now
		return false
	}

	limitObj.Requests = append(valid, now)
	return true
}

func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limits, key)
}

func (rl *RateLimiter) GetUsage(key string) (int, bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, ok := rl.limits[key]
	if !ok {
		return 0, false
	}
	return len(limit.Requests), true
}

// Circuit Breaker

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

type CircuitBreaker struct {
	name             string
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastFailure      time.Time
	mu               sync.RWMutex
}

func NewCircuitBreaker(name string, failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	if cb.state == CircuitHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitClosed
			cb.successCount = 0
		}
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitOpen
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// Load Balancer

type LoadBalancer struct {
	backends []*Backend
	current  int
	mu       sync.RWMutex
}

type Backend struct {
	Name      string
	URL       string
	Weight    int
	Healthy   bool
	LastCheck time.Time
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{}
}

func (lb *LoadBalancer) AddBackend(backend *Backend) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.backends = append(lb.backends, backend)
}

func (lb *LoadBalancer) RemoveBackend(name string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, b := range lb.backends {
		if b.Name == name {
			lb.backends = append(lb.backends[:i], lb.backends[i+1:]...)
			return
		}
	}
}

func (lb *LoadBalancer) Next() *Backend {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.backends) == 0 {
		return nil
	}

	healthy := make([]*Backend, 0)
	for _, b := range lb.backends {
		if b.Healthy {
			healthy = append(healthy, b)
		}
	}

	if len(healthy) == 0 {
		return nil
	}

	lb.current = (lb.current + 1) % len(healthy)
	return healthy[lb.current]
}

func (lb *LoadBalancer) List() []*Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.backends
}

// Gateway

type Gateway struct {
	rateLimiter     *RateLimiter
	circuitBreakers map[string]*CircuitBreaker
	loadBalancers   map[string]*LoadBalancer
	mu              sync.RWMutex
}

func New() *Gateway {
	return &Gateway{
		rateLimiter:     NewRateLimiter(),
		circuitBreakers: make(map[string]*CircuitBreaker),
		loadBalancers:   make(map[string]*LoadBalancer),
	}
}

func (g *Gateway) RateLimiter() *RateLimiter {
	return g.rateLimiter
}

func (g *Gateway) AddCircuitBreaker(name string, cb *CircuitBreaker) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.circuitBreakers[name] = cb
}

func (g *Gateway) GetCircuitBreaker(name string) (*CircuitBreaker, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cb, ok := g.circuitBreakers[name]
	return cb, ok
}

func (g *Gateway) AddLoadBalancer(name string, lb *LoadBalancer) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.loadBalancers[name] = lb
}

func (g *Gateway) GetLoadBalancer(name string) (*LoadBalancer, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	lb, ok := g.loadBalancers[name]
	return lb, ok
}

func (g *Gateway) Route(request *Request) (*Response, error) {
	// Rate limiting
	if !g.rateLimiter.Allow(request.ClientID, 100, time.Minute) {
		return nil, naeoserr.New(naeoserr.ErrNetwork, "rate limit exceeded")
	}

	// Circuit breaker
	cb, ok := g.GetCircuitBreaker(request.Service)
	if ok && !cb.Allow() {
		return nil, naeoserr.New(naeoserr.ErrNetwork, fmt.Sprintf("circuit breaker open for %s", request.Service))
	}

	// Load balancing
	lb, ok := g.GetLoadBalancer(request.Service)
	if ok {
		backend := lb.Next()
		if backend == nil {
			return nil, naeoserr.New(naeoserr.ErrNetwork, fmt.Sprintf("no healthy backends for %s", request.Service))
		}

		if cb != nil {
			cb.RecordSuccess()
		}

		return &Response{
			StatusCode: 200,
			Body:       fmt.Sprintf("Response from %s", backend.Name),
		}, nil
	}

	return &Response{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}

type Request struct {
	ID       string
	ClientID string
	Service  string
	Path     string
	Method   string
	Headers  map[string]string
	Body     []byte
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}
