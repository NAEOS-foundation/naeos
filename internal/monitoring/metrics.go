package monitoring

import (
	"fmt"
	"net/http"
	"sort"
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

type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      map[string]string
	Help        string
	Buckets     []float64
	Quantiles   []float64
}

type MetricFamily struct {
	Name    string
	Type    MetricType
	Help    string
	Metrics []*Metric
}

// Registry

type Registry struct {
	metrics map[string]*MetricFamily
	mu      sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]*MetricFamily),
	}
}

func (r *Registry) Register(name string, metricType MetricType, help string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.metrics[name]; exists {
		return
	}

	r.metrics[name] = &MetricFamily{
		Name: name,
		Type: metricType,
		Help: help,
	}
}

func (r *Registry) CounterInc(name string, labels map[string]string) {
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
		Name:   name,
		Type:   Counter,
		Value:  1,
		Labels: labels,
	})
}

func (r *Registry) CounterAdd(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
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
		Name:   name,
		Type:   Counter,
		Value:  value,
		Labels: labels,
	})
}

func (r *Registry) GaugeSet(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
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
		Name:   name,
		Type:   Gauge,
		Value:  value,
		Labels: labels,
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
		Name:   name,
		Type:   Gauge,
		Value:  1,
		Labels: labels,
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
		Name:   name,
		Type:   Gauge,
		Value:  -1,
		Labels: labels,
	})
}

func (r *Registry) HistogramObserve(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.metrics[name]
	if !ok {
		return
	}

	family.Metrics = append(family.Metrics, &Metric{
		Name:   name,
		Type:   Histogram,
		Value:  value,
		Labels: labels,
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

func (r *Registry) FormatPrometheus() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := ""
	for _, family := range r.metrics {
		result += fmt.Sprintf("# HELP %s %s\n", family.Name, family.Help)
		result += fmt.Sprintf("# TYPE %s %s\n", family.Name, family.Type)

		for _, m := range family.Metrics {
			if len(m.Labels) > 0 {
				result += fmt.Sprintf("%s{%s} %f\n", family.Name, formatLabels(m.Labels), m.Value)
			} else {
				result += fmt.Sprintf("%s %f\n", family.Name, m.Value)
			}
		}
		result += "\n"
	}
	return result
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

	// Register default metrics
	reg.Register("naeos_requests_total", Counter, "Total HTTP requests")
	reg.Register("naeos_request_duration_seconds", Histogram, "HTTP request duration")
	reg.Register("naeos_pipelines_total", Counter, "Total pipeline runs")
	reg.Register("naeos_pipeline_duration_seconds", Histogram, "Pipeline run duration")
	reg.Register("naeos_spec_validations_total", Counter, "Total spec validations")
	reg.Register("naeos_artifacts_generated_total", Counter, "Total artifacts generated")
	reg.Register("naeos_active_websocket_connections", Gauge, "Active WebSocket connections")
	reg.Register("naeos_uptime_seconds", Gauge, "Server uptime in seconds")

	return &Metrics{registry: reg}
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
	m.registry.HistogramObserve("naeos_request_duration_seconds", duration, map[string]string{
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
	m.registry.HistogramObserve("naeos_pipeline_duration_seconds", duration, nil)
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
		w.Write([]byte(registry.FormatPrometheus()))
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

			next.ServeHTTP(w, r)

			duration := time.Since(start).Seconds()
			metrics.IncRequests(r.Method, r.URL.Path, "200")
			metrics.ObserveRequestDuration(r.Method, r.URL.Path, duration)
		})
	}
}
