package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type HealthChecker struct {
	interval time.Duration
	timeout  time.Duration
	stopCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		interval: interval,
		timeout:  timeout,
		stopCh:   make(chan struct{}),
	}
}

func (hc *HealthChecker) Check(backend *Backend) bool {
	if backend.URL == "" {
		return backend.Healthy
	}

	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, backend.URL+"/health", nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: hc.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	backend.LastCheck = time.Now()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (hc *HealthChecker) Start(lb *LoadBalancer) {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = true
	hc.mu.Unlock()

	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				lb.mu.Lock()
				for _, backend := range lb.backends {
					backend.Healthy = hc.Check(backend)
				}
				lb.mu.Unlock()
			case <-hc.stopCh:
				return
			}
		}
	}()
}

func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	if hc.running {
		close(hc.stopCh)
		hc.running = false
	}
}

type WeightedLoadBalancer struct {
	*LoadBalancer
	currentWeight map[string]int
	mu            sync.RWMutex
}

func NewWeightedLoadBalancer() *WeightedLoadBalancer {
	return &WeightedLoadBalancer{
		LoadBalancer:  NewLoadBalancer(),
		currentWeight: make(map[string]int),
	}
}

func (wlb *WeightedLoadBalancer) Next() *Backend {
	wlb.LoadBalancer.mu.Lock()
	defer wlb.LoadBalancer.mu.Unlock()

	healthy := make([]*Backend, 0)
	for _, b := range wlb.backends {
		if b.Healthy {
			healthy = append(healthy, b)
		}
	}

	if len(healthy) == 0 {
		return nil
	}

	wlb.mu.Lock()
	defer wlb.mu.Unlock()

	var selected *Backend
	totalWeight := 0

	for _, b := range healthy {
		totalWeight += b.Weight
	}

	if totalWeight == 0 {
		wlb.current = (wlb.current + 1) % len(healthy)
		return healthy[wlb.current]
	}

	for _, b := range healthy {
		if _, ok := wlb.currentWeight[b.Name]; !ok {
			wlb.currentWeight[b.Name] = 0
		}
		wlb.currentWeight[b.Name] += b.Weight

		if selected == nil || wlb.currentWeight[b.Name] > wlb.currentWeight[selected.Name] {
			selected = b
		}
	}

	if selected != nil {
		wlb.currentWeight[selected.Name] -= totalWeight
	}

	return selected
}

type ProxyConfig struct {
	Timeout       time.Duration
	BufferPool    int
	FlushInterval time.Duration
}

type ReverseProxy struct {
	proxy   *httputil.ReverseProxy
	config  *ProxyConfig
	backend *Backend
	mu      sync.RWMutex
}

func NewReverseProxy(backend *Backend, config *ProxyConfig) (*ReverseProxy, error) {
	if backend == nil || backend.URL == "" {
		return nil, naeoserr.New(naeoserr.ErrNetwork, "backend URL is required")
	}

	target, err := url.Parse(backend.URL)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "invalid backend URL")
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	if config != nil {
		proxy.Transport = &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}
	}

	return &ReverseProxy{
		proxy:   proxy,
		config:  config,
		backend: backend,
	}, nil
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rp.proxy.ServeHTTP(w, r)
}

func (rp *ReverseProxy) UpdateBackend(backend *Backend) error {
	if backend == nil || backend.URL == "" {
		return naeoserr.New(naeoserr.ErrNetwork, "backend URL is required")
	}

	target, err := url.Parse(backend.URL)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "invalid backend URL")
	}

	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.backend = backend
	rp.proxy = httputil.NewSingleHostReverseProxy(target)
	return nil
}

type LoadBalancingProxy struct {
	*ReverseProxy
	loadBalancer *LoadBalancer
	counter      atomic.Int64
}

func NewLoadBalancingProxy(lb *LoadBalancer) *LoadBalancingProxy {
	return &LoadBalancingProxy{
		loadBalancer: lb,
	}
}

func (lbp *LoadBalancingProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lbp.loadBalancer.Next()
	if backend == nil {
		http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
		return
	}

	target, err := url.Parse(backend.URL)
	if err != nil {
		http.Error(w, "invalid backend URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		backend.Healthy = false
		http.Error(w, fmt.Sprintf("backend error: %v", err), http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
	lbp.counter.Add(1)
}

func (lbp *LoadBalancingProxy) RequestCount() int64 {
	return lbp.counter.Load()
}

type RequestLogger struct {
	logger func(level, msg string, fields map[string]any)
	mu     sync.Mutex
}

func NewRequestLogger(logger func(level, msg string, fields map[string]any)) *RequestLogger {
	return &RequestLogger{logger: logger}
}

func (rl *RequestLogger) LogRequest(r *http.Request, statusCode int, duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.logger != nil {
		fields := map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": statusCode,
			"duration_ms": duration.Milliseconds(),
			"remote_addr": r.RemoteAddr,
		}
		rl.logger("info", "request completed", fields)
	}
}

type ProxyChain struct {
	proxies []http.Handler
	logger  *RequestLogger
	mu      sync.RWMutex
}

func NewProxyChain(logger *RequestLogger) *ProxyChain {
	return &ProxyChain{
		logger: logger,
	}
}

func (pc *ProxyChain) AddProxy(handler http.Handler) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.proxies = append(pc.proxies, handler)
}

func (pc *ProxyChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pc.mu.RLock()
	proxies := make([]http.Handler, len(pc.proxies))
	copy(proxies, pc.proxies)
	pc.mu.RUnlock()

	if len(proxies) == 0 {
		http.Error(w, "no proxies configured", http.StatusServiceUnavailable)
		return
	}

	start := time.Now()
	for _, proxy := range proxies {
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		proxy.ServeHTTP(rec, r)

		if rec.written {
			if pc.logger != nil {
				pc.logger.LogRequest(r, rec.statusCode, time.Since(start))
			}
			return
		}
	}

	http.Error(w, "no backend responded", http.StatusBadGateway)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.written = true
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.written {
		rr.written = true
	}
	return rr.ResponseWriter.Write(b)
}

func (rr *responseRecorder) Unwrap() http.ResponseWriter {
	return rr.ResponseWriter
}

type BackendPool struct {
	backends map[string]*Backend
	mu       sync.RWMutex
	factory  func(name, url string) *Backend
}

func NewBackendPool(factory func(name, url string) *Backend) *BackendPool {
	if factory == nil {
		factory = func(name, u string) *Backend {
			return &Backend{
				Name:    name,
				URL:     u,
				Weight:  1,
				Healthy: true,
			}
		}
	}
	return &BackendPool{
		backends: make(map[string]*Backend),
		factory:  factory,
	}
}

func (bp *BackendPool) GetOrCreate(name, backendURL string) *Backend {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if b, ok := bp.backends[name]; ok {
		return b
	}

	b := bp.factory(name, backendURL)
	bp.backends[name] = b
	return b
}

func (bp *BackendPool) Remove(name string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	delete(bp.backends, name)
}

func (bp *BackendPool) List() map[string]*Backend {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	result := make(map[string]*Backend, len(bp.backends))
	for k, v := range bp.backends {
		result[k] = v
	}
	return result
}

func (bp *BackendPool) HealthyCount() int {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	count := 0
	for _, b := range bp.backends {
		if b.Healthy {
			count++
		}
	}
	return count
}

func CopyBody(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return io.ReadAll(r)
}
