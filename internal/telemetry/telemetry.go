package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// SpanStatus represents the status of a span.
type SpanStatus int

const (
	StatusUnset SpanStatus = iota
	StatusOK
	StatusError
)

func (s SpanStatus) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusError:
		return "ERROR"
	default:
		return "UNSET"
	}
}

type Span struct {
	Name       string            `json:"name"`
	ID         string            `json:"id"`
	TraceID    string            `json:"trace_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Status     string            `json:"status"`
	Attributes SpanAttributes   `json:"attributes,omitempty"`
}

type Exporter interface {
	ExportSpans(spans []Span) error
	Flush() error
}

type Config struct {
	Endpoint  string
	Timeout   time.Duration
	BatchSize int
}

type Service struct {
	config   Config
	exporter Exporter
	spans    []Span
	mu       sync.Mutex
}

func NewService(config Config, exporter Exporter) *Service {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	return &Service{
		config:   config,
		exporter: exporter,
		spans:    make([]Span, 0, config.BatchSize),
	}
}

func (s *Service) StartSpan(name string) *Span {
	return &Span{
		Name:      name,
		ID:        generateID(),
		TraceID:   generateID(),
		StartTime: time.Now(),
		Status:    "ok",
	}
}

func (s *Service) StartSpanWithParent(name, parentID string) *Span {
	return &Span{
		Name:      name,
		ID:        generateID(),
		TraceID:   generateID(),
		ParentID:  parentID,
		StartTime: time.Now(),
		Status:    "ok",
	}
}

func (s *Service) EndSpan(span *Span) {
	span.EndTime = time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spans = append(s.spans, *span)
	if len(s.spans) >= s.config.BatchSize {
		s.flushUnsafe()
	}
}

func (s *Service) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushUnsafe()
}

func (s *Service) flushUnsafe() error {
	if len(s.spans) == 0 {
		return nil
	}
	batch := make([]Span, len(s.spans))
	copy(batch, s.spans)
	s.spans = s.spans[:0]
	return s.exporter.ExportSpans(batch)
}

func (s *Service) SpanCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.spans)
}

type HTTPExporter struct {
	endpoint string
	client   *http.Client
	spans    []Span
	mu       sync.Mutex
}

func NewHTTPExporter(endpoint string, timeout time.Duration) *HTTPExporter {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &HTTPExporter{
		endpoint: endpoint,
		client:   &http.Client{Timeout: timeout},
		spans:    make([]Span, 0),
	}
}

func (e *HTTPExporter) ExportSpans(spans []Span) error {
	e.mu.Lock()
	e.spans = append(e.spans, spans...)
	e.mu.Unlock()

	return e.Flush()
}

func (e *HTTPExporter) Flush() error {
	e.mu.Lock()
	if len(e.spans) == 0 {
		e.mu.Unlock()
		return nil
	}
	batch := make([]Span, len(e.spans))
	copy(batch, e.spans)
	e.spans = e.spans[:0]
	e.mu.Unlock()

	data, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshal spans: %w", err)
	}

	resp, err := e.client.Post(e.endpoint+"/v1/traces", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("export spans: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("export failed with status %d", resp.StatusCode)
	}
	return nil
}

var idCounter uint64

func generateID() string {
	idCounter++
	return fmt.Sprintf("span-%d", idCounter)
}

// --- Sampler ---

type Sampler interface {
	Sample(name string) bool
}

type AlwaysSample struct{}

func (AlwaysSample) Sample(string) bool { return true }

type NeverSample struct{}

func (NeverSample) Sample(string) bool { return false }

type ProbabilisticSampler struct {
	Rate float64
}

func NewProbabilisticSampler(rate float64) *ProbabilisticSampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &ProbabilisticSampler{Rate: rate}
}

func (p *ProbabilisticSampler) Sample(string) bool {
	return rand.Float64() < p.Rate
}

type RateLimiterSampler struct {
	maxPerSecond int
	counter      int64
	windowStart  time.Time
	mu           sync.Mutex
}

func NewRateLimiterSampler(maxPerSecond int) *RateLimiterSampler {
	if maxPerSecond <= 0 {
		maxPerSecond = 1
	}
	return &RateLimiterSampler{
		maxPerSecond: maxPerSecond,
		windowStart:  time.Now(),
	}
}

func (r *RateLimiterSampler) Sample(string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.windowStart) >= time.Second {
		r.counter = 0
		r.windowStart = now
	}
	if int(r.counter) >= r.maxPerSecond {
		return false
	}
	r.counter++
	return true
}

// --- SpanAttributes ---

type attrValue struct {
	t    string
	s    string
	i    int64
	f    float64
	b    bool
}

type SpanAttributes struct {
	mu   sync.RWMutex
	attrs map[string]attrValue
}

func (a *SpanAttributes) Set(key string, value interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.attrs == nil {
		a.attrs = make(map[string]attrValue)
	}
	switch v := value.(type) {
	case string:
		a.attrs[key] = attrValue{t: "string", s: v}
	case int64:
		a.attrs[key] = attrValue{t: "int64", i: v}
	case float64:
		a.attrs[key] = attrValue{t: "float64", f: v}
	case bool:
		a.attrs[key] = attrValue{t: "bool", b: v}
	}
}

func (a *SpanAttributes) Get(key string) (interface{}, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	v, ok := a.attrs[key]
	if !ok {
		return nil, false
	}
	switch v.t {
	case "string":
		return v.s, true
	case "int64":
		return v.i, true
	case "float64":
		return v.f, true
	case "bool":
		return v.b, true
	}
	return nil, false
}

func (a *SpanAttributes) Has(key string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.attrs[key]
	return ok
}

func (a *SpanAttributes) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.attrs)
}

// MarshalJSON for SpanAttributes.
func (a *SpanAttributes) MarshalJSON() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	m := make(map[string]interface{}, len(a.attrs))
	for k, v := range a.attrs {
		switch v.t {
		case "string":
			m[k] = v.s
		case "int64":
			m[k] = v.i
		case "float64":
			m[k] = v.f
		case "bool":
			m[k] = v.b
		}
	}
	return json.Marshal(m)
}

// --- MetricsCollector ---

type MetricType int

const (
	MetricCounter MetricType = iota
	MetricGauge
	MetricHistogram
)

type metricEntry struct {
	metricType MetricType
	counter    int64
	gauge      float64
	histogram  []float64
	mu         sync.Mutex
}

type MetricsCollector struct {
	metrics map[string]*metricEntry
	mu      sync.RWMutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*metricEntry),
	}
}

func (mc *MetricsCollector) getOrCreate(name string, mt MetricType) *metricEntry {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if e, ok := mc.metrics[name]; ok {
		return e
	}
	e := &metricEntry{metricType: mt}
	mc.metrics[name] = e
	return e
}

func (mc *MetricsCollector) IncrCounter(name string, delta int64) {
	e := mc.getOrCreate(name, MetricCounter)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.counter += delta
}

func (mc *MetricsCollector) GetCounter(name string) int64 {
	mc.mu.RLock()
	e, ok := mc.metrics[name]
	mc.mu.RUnlock()
	if !ok {
		return 0
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.counter
}

func (mc *MetricsCollector) SetGauge(name string, value float64) {
	e := mc.getOrCreate(name, MetricGauge)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.gauge = value
}

func (mc *MetricsCollector) GetGauge(name string) float64 {
	mc.mu.RLock()
	e, ok := mc.metrics[name]
	mc.mu.RUnlock()
	if !ok {
		return 0
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.gauge
}

func (mc *MetricsCollector) RecordHistogram(name string, value float64) {
	e := mc.getOrCreate(name, MetricHistogram)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.histogram = append(e.histogram, value)
}

func (mc *MetricsCollector) GetHistogram(name string) []float64 {
	mc.mu.RLock()
	e, ok := mc.metrics[name]
	mc.mu.RUnlock()
	if !ok {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]float64, len(e.histogram))
	copy(out, e.histogram)
	return out
}

func (mc *MetricsCollector) Export() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	result := make(map[string]interface{}, len(mc.metrics))
	for name, e := range mc.metrics {
		e.mu.Lock()
		switch e.metricType {
		case MetricCounter:
			result[name] = map[string]interface{}{"type": "counter", "value": e.counter}
		case MetricGauge:
			result[name] = map[string]interface{}{"type": "gauge", "value": e.gauge}
		case MetricHistogram:
			h := make([]float64, len(e.histogram))
			copy(h, e.histogram)
			result[name] = map[string]interface{}{"type": "histogram", "values": h}
		}
		e.mu.Unlock()
	}
	return result
}

// --- MultiExporter ---

type MultiExporter struct {
	exporters []Exporter
}

func NewMultiExporter(exporters ...Exporter) *MultiExporter {
	return &MultiExporter{exporters: exporters}
}

func (m *MultiExporter) ExportSpans(spans []Span) error {
	for _, exp := range m.exporters {
		if err := exp.ExportSpans(spans); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiExporter) Flush() error {
	for _, exp := range m.exporters {
		if err := exp.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// --- BatchProcessor ---

type BatchProcessor struct {
	exporter    Exporter
	batchSize   int
	maxQueue    int
	flushTicker *time.Ticker
	queue       []Span
	mu          sync.Mutex
	done        chan struct{}
	onFlush     func([]Span)
}

func NewBatchProcessor(exporter Exporter, batchSize int, flushInterval time.Duration, maxQueue int) *BatchProcessor {
	if batchSize <= 0 {
		batchSize = 10
	}
	if maxQueue <= 0 {
		maxQueue = 1000
	}
	bp := &BatchProcessor{
		exporter:  exporter,
		batchSize: batchSize,
		maxQueue:  maxQueue,
		queue:     make([]Span, 0, batchSize),
		done:      make(chan struct{}),
	}
	if flushInterval > 0 {
		bp.flushTicker = time.NewTicker(flushInterval)
		go bp.backgroundFlush()
	}
	return bp
}

func (bp *BatchProcessor) backgroundFlush() {
	for {
		select {
		case <-bp.flushTicker.C:
			bp.Flush()
		case <-bp.done:
			return
		}
	}
}

func (bp *BatchProcessor) AddSpan(span Span) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	if len(bp.queue) >= bp.maxQueue {
		return fmt.Errorf("queue full: %d spans", len(bp.queue))
	}
	bp.queue = append(bp.queue, span)
	if len(bp.queue) >= bp.batchSize {
		return bp.flushUnsafe()
	}
	return nil
}

func (bp *BatchProcessor) Flush() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.flushUnsafe()
}

func (bp *BatchProcessor) flushUnsafe() error {
	if len(bp.queue) == 0 {
		return nil
	}
	batch := make([]Span, len(bp.queue))
	copy(batch, bp.queue)
	bp.queue = bp.queue[:0]
	if bp.onFlush != nil {
		bp.onFlush(batch)
	}
	return bp.exporter.ExportSpans(batch)
}

func (bp *BatchProcessor) QueueSize() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return len(bp.queue)
}

func (bp *BatchProcessor) Stop() {
	if bp.flushTicker != nil {
		bp.flushTicker.Stop()
	}
	close(bp.done)
}

// --- Tracer ---

type Tracer struct {
	serviceName string
	exporter    Exporter
}

func NewTracer(serviceName string, exporter Exporter) *Tracer {
	return &Tracer{
		serviceName: serviceName,
		exporter:    exporter,
	}
}

func (t *Tracer) StartSpan(name string) *Span {
	return &Span{
		Name:      name,
		ID:        generateID(),
		TraceID:   generateID(),
		StartTime: time.Now(),
		Status:    StatusOK.String(),
		Labels:    map[string]string{"service": t.serviceName},
	}
}

func (t *Tracer) StartSpanWithParent(name, parentID, traceID string) *Span {
	return &Span{
		Name:      name,
		ID:        generateID(),
		TraceID:   traceID,
		ParentID:  parentID,
		StartTime: time.Now(),
		Status:    StatusOK.String(),
		Labels:    map[string]string{"service": t.serviceName},
	}
}

func (t *Tracer) EndSpan(span *Span) error {
	span.EndTime = time.Now()
	return t.exporter.ExportSpans([]Span{*span})
}

func (t *Tracer) Flush() error {
	return t.exporter.Flush()
}

// --- ConsoleExporter ---

type ConsoleExporter struct {
	writer io.Writer
}

func NewConsoleExporter(w io.Writer) *ConsoleExporter {
	return &ConsoleExporter{writer: w}
}

func (c *ConsoleExporter) ExportSpans(spans []Span) error {
	for _, span := range spans {
		data, err := json.Marshal(span)
		if err != nil {
			return fmt.Errorf("marshal span: %w", err)
		}
		_, err = fmt.Fprintf(c.writer, "%s\n", data)
		if err != nil {
			return fmt.Errorf("write span: %w", err)
		}
	}
	return nil
}

func (c *ConsoleExporter) Flush() error {
	if f, ok := c.writer.(interface{ Flush() error }); ok {
		return f.Flush()
	}
	return nil
}

// --- InMemoryExporter ---

type InMemoryExporter struct {
	spans [][]Span
	mu    sync.Mutex
}

func NewInMemoryExporter() *InMemoryExporter {
	return &InMemoryExporter{}
}

func (e *InMemoryExporter) ExportSpans(spans []Span) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	batch := make([]Span, len(spans))
	copy(batch, spans)
	e.spans = append(e.spans, batch)
	return nil
}

func (e *InMemoryExporter) Flush() error { return nil }

func (e *InMemoryExporter) AllSpans() []Span {
	e.mu.Lock()
	defer e.mu.Unlock()
	var all []Span
	for _, batch := range e.spans {
		all = append(all, batch...)
	}
	return all
}

func (e *InMemoryExporter) BatchCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.spans)
}

func (e *InMemoryExporter) TotalSpanCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	total := 0
	for _, batch := range e.spans {
		total += len(batch)
	}
	return total
}

func (e *InMemoryExporter) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = e.spans[:0]
}

// Ensure unused import suppression.
var _ = atomic.LoadInt64
