package monitoring

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Metric Types

type MetricType string

const (
	Counter   MetricType = "counter"
	Gauge     MetricType = "gauge"
	Histogram MetricType = "histogram"
	Summary   MetricType = "summary"
)

var defaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    map[string]string
	Help      string
	Buckets   []float64
	Quantiles []float64
	CreatedAt time.Time
}

type MetricFamily struct {
	Name      string
	Type      MetricType
	Help      string
	Metrics   []*Metric
	Buckets   []float64
	Counts    []uint64
	Sum       float64
	Count     uint64
	CreatedAt time.Time
}

// Registry

type Registry struct {
	metrics         map[string]*MetricFamily
	mu              sync.RWMutex
	maxCardinality  int
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
	cleanupRunning  bool
	cleanupMu       sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		metrics:         make(map[string]*MetricFamily),
		maxCardinality:  1000,
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}
}

func NewRegistryWithOptions(maxCardinality int, cleanupInterval time.Duration) *Registry {
	if maxCardinality <= 0 {
		maxCardinality = 1000
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 5 * time.Minute
	}
	return &Registry{
		metrics:         make(map[string]*MetricFamily),
		maxCardinality:  maxCardinality,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan struct{}),
	}
}

func (r *Registry) StartCleanup(minAge time.Duration) {
	r.cleanupMu.Lock()
	defer r.cleanupMu.Unlock()
	if r.cleanupRunning {
		return
	}
	r.cleanupRunning = true
	r.stopCleanup = make(chan struct{})

	go func() {
		ticker := time.NewTicker(r.cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.CleanupStale(minAge)
			case <-r.stopCleanup:
				return
			}
		}
	}()
}

func (r *Registry) StopCleanup() {
	r.cleanupMu.Lock()
	defer r.cleanupMu.Unlock()
	if !r.cleanupRunning {
		return
	}
	r.cleanupRunning = false
	close(r.stopCleanup)
}

func (r *Registry) CleanupStale(minAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-minAge)
	for name, family := range r.metrics {
		filtered := family.Metrics[:0]
		for _, m := range family.Metrics {
			if m.CreatedAt.IsZero() || m.CreatedAt.After(cutoff) {
				filtered = append(filtered, m)
			}
		}
		if len(filtered) == 0 {
			delete(r.metrics, name)
		} else {
			family.Metrics = filtered
		}
	}
}

func (r *Registry) Register(name string, metricType MetricType, help string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.metrics[name]; exists {
		return
	}

	family := &MetricFamily{
		Name:      name,
		Type:      metricType,
		Help:      help,
		CreatedAt: time.Now(),
	}
	if metricType == Histogram {
		family.Buckets = make([]float64, len(defaultBuckets))
		copy(family.Buckets, defaultBuckets)
		family.Counts = make([]uint64, len(defaultBuckets)+1)
	}
	r.metrics[name] = family
}

func (r *Registry) RegisterWithBuckets(name string, help string, buckets []float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.metrics[name]; exists {
		return
	}

	family := &MetricFamily{
		Name:      name,
		Type:      Histogram,
		Help:      help,
		CreatedAt: time.Now(),
	}
	family.Buckets = make([]float64, len(buckets))
	copy(family.Buckets, buckets)
	family.Counts = make([]uint64, len(buckets)+1)
	r.metrics[name] = family
}

func (r *Registry) CounterInc(name string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	if r.cardinalityForLocked(name) >= r.maxCardinality {
		return
	}

	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			m.Value++
			return
		}
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:      name,
		Type:      Counter,
		Value:     1,
		Labels:    labels,
		CreatedAt: time.Now(),
	})
}

func (r *Registry) CounterAdd(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	if r.cardinalityForLocked(name) >= r.maxCardinality {
		return
	}

	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			m.Value += value
			return
		}
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:      name,
		Type:      Counter,
		Value:     value,
		Labels:    labels,
		CreatedAt: time.Now(),
	})
}

func (r *Registry) GaugeSet(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	if r.cardinalityForLocked(name) >= r.maxCardinality {
		return
	}

	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			m.Value = value
			return
		}
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:      name,
		Type:      Gauge,
		Value:     value,
		Labels:    labels,
		CreatedAt: time.Now(),
	})
}

func (r *Registry) GaugeInc(name string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			m.Value++
			return
		}
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:      name,
		Type:      Gauge,
		Value:     1,
		Labels:    labels,
		CreatedAt: time.Now(),
	})
}

func (r *Registry) GaugeDec(name string, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			m.Value--
			return
		}
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:      name,
		Type:      Gauge,
		Value:     -1,
		Labels:    labels,
		CreatedAt: time.Now(),
	})
}

func (r *Registry) HistogramObserve(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	family.Sum += value
	family.Count++

	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			m.Value = value
			return
		}
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:      name,
		Type:      Histogram,
		Value:     value,
		Labels:    labels,
		Buckets:   family.Buckets,
		CreatedAt: time.Now(),
	})
}

func (r *Registry) HistogramObserveWithBuckets(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	family.Sum += value
	family.Count++

	for i, bound := range family.Buckets {
		if value <= bound {
			family.Counts[i]++
			break
		}
	}
	if value > family.Buckets[len(family.Buckets)-1] {
		family.Counts[len(family.Buckets)]++
	}

	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			m.Value = value
			return
		}
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:      name,
		Type:      Histogram,
		Value:     value,
		Labels:    labels,
		Buckets:   family.Buckets,
		CreatedAt: time.Now(),
	})
}

func (r *Registry) GetFamilies() []*MetricFamily {
	r.mu.RLock()
	defer r.mu.RUnlock()

	families := make([]*MetricFamily, 0, len(r.metrics))
	for _, f := range r.metrics {
		families = append(families, f)
	}
	return families
}

func (r *Registry) GetFamily(name string) (*MetricFamily, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	f, ok := r.metrics[name]
	return f, ok
}

func (r *Registry) CounterValue(name string, labels map[string]string) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	family, ok := r.metrics[name]
	if !ok {
		return 0
	}
	key := labelsKey(labels)
	for _, m := range family.Metrics {
		if labelsKey(m.Labels) == key {
			return m.Value
		}
	}
	return 0
}

func (r *Registry) GaugeValue(name string, labels map[string]string) float64 {
	return r.CounterValue(name, labels)
}

func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.metrics, name)
}

func (r *Registry) FormatPrometheus() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sb strings.Builder
	for _, family := range r.metrics {
		fmt.Fprintf(&sb, "# HELP %s %s\n", family.Name, family.Help)
		fmt.Fprintf(&sb, "# TYPE %s %s\n", family.Name, family.Type)

		if family.Type == Histogram && len(family.Buckets) > 0 {
			cumCount := uint64(0)
			for i, bound := range family.Buckets {
				cumCount += family.Counts[i]
				fmt.Fprintf(&sb, "%s_bucket{le=\"%g\"} %d\n", family.Name, bound, cumCount)
			}
			cumCount += family.Counts[len(family.Buckets)]
			fmt.Fprintf(&sb, "%s_bucket{le=\"+Inf\"} %d\n", family.Name, cumCount)
			fmt.Fprintf(&sb, "%s_sum %g\n", family.Name, family.Sum)
			fmt.Fprintf(&sb, "%s_count %d\n", family.Name, family.Count)
		} else {
			for _, m := range family.Metrics {
				if len(m.Labels) > 0 {
					fmt.Fprintf(&sb, "%s{%s} %f\n", family.Name, formatLabels(m.Labels), m.Value)
				} else {
					fmt.Fprintf(&sb, "%s %f\n", family.Name, m.Value)
				}
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func labelsKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	key := ""
	for _, k := range keys {
		key += k + "=" + labels[k] + ","
	}
	return key
}

func formatLabels(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := ""
	for _, k := range keys {
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf(`%s="%s"`, k, labels[k])
	}
	return result
}

func (r *Registry) cardinalityForLocked(name string) int {
	family, ok := r.metrics[name]
	if !ok {
		return 0
	}
	return len(family.Metrics)
}

// statusResponseWriter captures the HTTP status code

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (w *statusResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Collector

type Collector struct {
	registry *Registry
}

func NewCollector(registry *Registry) *Collector {
	return &Collector{registry: registry}
}

func (c *Collector) Collect() *MetricsSnapshot {
	families := c.registry.GetFamilies()
	snapshot := &MetricsSnapshot{
		Timestamp: time.Now(),
		Families:  families,
	}
	return snapshot
}

type MetricsSnapshot struct {
	Timestamp time.Time
	Families  []*MetricFamily
}

// Default Metrics

type Metrics struct {
	registry *Registry
}

func NewMetrics() *Metrics {
	reg := NewRegistry()

	reg.Register("naeos_requests_total", Counter, "Total HTTP requests")
	reg.Register("naeos_request_duration_seconds", Histogram, "HTTP request duration")
	reg.Register("naeos_pipelines_total", Counter, "Total pipeline runs")
	reg.Register("naeos_pipeline_duration_seconds", Histogram, "Pipeline run duration")
	reg.Register("naeos_spec_validations_total", Counter, "Total spec validations")
	reg.Register("naeos_artifacts_generated_total", Counter, "Total artifacts generated")
	reg.Register("naeos_active_websocket_connections", Gauge, "Active WebSocket connections")
	reg.Register("naeos_uptime_seconds", Gauge, "Server uptime in seconds")
	reg.Register("naeos_go_goroutines", Gauge, "Number of goroutines")
	reg.Register("naeos_go_memory_alloc_bytes", Gauge, "Current memory allocation in bytes")
	reg.Register("naeos_go_memory_sys_bytes", Gauge, "Total memory obtained from OS")
	reg.Register("naeos_go_gc_pauses_total", Counter, "Total number of GC pauses")

	return &Metrics{registry: reg}
}

func (m *Metrics) RecordRuntime() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	m.registry.GaugeSet("naeos_go_goroutines", float64(runtime.NumGoroutine()), nil)
	m.registry.GaugeSet("naeos_go_memory_alloc_bytes", float64(mem.Alloc), nil)
	m.registry.GaugeSet("naeos_go_memory_sys_bytes", float64(mem.Sys), nil)
	m.registry.CounterAdd("naeos_go_gc_pauses_total", float64(mem.NumGC), nil)
}

func (m *Metrics) StartRuntimeCollector(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			m.RecordRuntime()
		}
	}()
}

func (m *Metrics) Registry() *Registry {
	return m.registry
}

func (m *Metrics) IncRequests(method, path, status string) {
	m.registry.CounterInc("naeos_requests_total", map[string]string{
		"method": method,
		"path":   path,
		"status": status,
	})
}

func (m *Metrics) ObserveRequestDuration(method, path string, duration float64) {
	m.registry.HistogramObserveWithBuckets("naeos_request_duration_seconds", duration, map[string]string{
		"method": method,
		"path":   path,
	})
}

func (m *Metrics) IncPipelines(status string) {
	m.registry.CounterInc("naeos_pipelines_total", map[string]string{
		"status": status,
	})
}

func (m *Metrics) ObservePipelineDuration(duration float64) {
	m.registry.HistogramObserveWithBuckets("naeos_pipeline_duration_seconds", duration, nil)
}

func (m *Metrics) IncSpecValidations(valid bool) {
	status := "success"
	if !valid {
		status = "failure"
	}
	m.registry.CounterInc("naeos_spec_validations_total", map[string]string{
		"status": status,
	})
}

func (m *Metrics) IncArtifacts() {
	m.registry.CounterInc("naeos_artifacts_generated_total", nil)
}

func (m *Metrics) SetWebSocketConnections(count int) {
	m.registry.GaugeSet("naeos_active_websocket_connections", float64(count), nil)
}

func (m *Metrics) SetUptime(seconds float64) {
	m.registry.GaugeSet("naeos_uptime_seconds", seconds, nil)
}

// HTTP Handlers

func PrometheusHandler(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(registry.FormatPrometheus()))
	}
}

func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	}
}

func ReadyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ready"}`)
	}
}

// Middleware

func MetricsMiddleware(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := newStatusResponseWriter(w)

			next.ServeHTTP(sw, r)

			duration := time.Since(start).Seconds()
			status := fmt.Sprintf("%d", sw.statusCode)
			metrics.IncRequests(r.Method, r.URL.Path, status)
			metrics.ObserveRequestDuration(r.Method, r.URL.Path, duration)
		})
	}
}
