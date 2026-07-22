package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

type spanContextKey struct{}

// ContextWithSpan returns a new context carrying the given span.
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

// SpanFromContext returns the span stored in the context, if any.
func SpanFromContext(ctx context.Context) (*Span, bool) {
	span, ok := ctx.Value(spanContextKey{}).(*Span)
	return span, ok
}

// InjectTraceToCmd sets environment variables for W3C Trace Context propagation
// on cmd if a span is found in ctx.
func InjectTraceToCmd(ctx context.Context, cmd *exec.Cmd) {
	span, ok := SpanFromContext(ctx)
	if !ok || span == nil {
		return
	}
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("TRACEPARENT=00-%s-%s-01", span.TraceID, span.SpanID),
	)
}

var idCounter atomic.Int64

func generateID() string {
	id := idCounter.Add(1)
	return fmt.Sprintf("%016x", id)
}

type Span struct {
	TraceID    string
	SpanID     string
	ParentID   string
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	Attributes map[string]any
	Events     []SpanEvent
	Status     SpanStatus
}

type SpanEvent struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]any
}

type SpanStatus struct {
	Code    SpanStatusCode
	Message string
}

type SpanStatusCode int

const (
	SpanStatusUnset SpanStatusCode = 0
	SpanStatusOK    SpanStatusCode = 1
	SpanStatusError SpanStatusCode = 2
)

type Exporter interface {
	ExportSpans(spans []*Span) error
	ExportMetrics(metrics []*Metric) error
	ExportLogs(entries []LogEntry) error
}

type JSONExporter struct {
	mu      sync.Mutex
	spans   []*Span
	metrics []*Metric
	logs    []LogEntry
}

func NewJSONExporter() *JSONExporter {
	return &JSONExporter{}
}

func (e *JSONExporter) ExportSpans(spans []*Span) error {
	e.mu.Lock()
	e.spans = append(e.spans, spans...)
	e.mu.Unlock()
	return nil
}

func (e *JSONExporter) ExportMetrics(metrics []*Metric) error {
	e.mu.Lock()
	e.metrics = append(e.metrics, metrics...)
	e.mu.Unlock()
	return nil
}

func (e *JSONExporter) ExportLogs(entries []LogEntry) error {
	e.mu.Lock()
	e.logs = append(e.logs, entries...)
	e.mu.Unlock()
	return nil
}

func (e *JSONExporter) Flush() ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	data := map[string]any{
		"spans":   e.spans,
		"metrics": e.metrics,
		"logs":    e.logs,
	}
	return json.MarshalIndent(data, "", "  ")
}

type FileExporter struct {
	path string
	mu   sync.Mutex
	file *os.File
}

func NewFileExporter(path string) (*FileExporter, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return &FileExporter{path: path, file: f}, nil
}

func (e *FileExporter) ExportSpans(spans []*Span) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, span := range spans {
		data, err := json.Marshal(span)
		if err != nil {
			continue
		}
		_, _ = e.file.Write(append(data, '\n'))
	}
	return nil
}

func (e *FileExporter) ExportMetrics(metrics []*Metric) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, m := range metrics {
		data, err := json.Marshal(m)
		if err != nil {
			continue
		}
		_, _ = e.file.Write(append(data, '\n'))
	}
	return nil
}

func (e *FileExporter) ExportLogs(entries []LogEntry) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		_, _ = e.file.Write(append(data, '\n'))
	}
	return nil
}

func (e *FileExporter) Close() error {
	return e.file.Close()
}

type MemoryCleanup struct {
	tracer  *Tracer
	maxAge  time.Duration
	stopCh  chan struct{}
	running bool
	mu      sync.Mutex
}

func NewMemoryCleanup(tracer *Tracer, maxAge time.Duration) *MemoryCleanup {
	return &MemoryCleanup{
		tracer: tracer,
		maxAge: maxAge,
		stopCh: make(chan struct{}),
	}
}

func (mc *MemoryCleanup) Start() {
	mc.mu.Lock()
	if mc.running {
		mc.mu.Unlock()
		return
	}
	mc.running = true
	mc.mu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mc.cleanup()
			case <-mc.stopCh:
				return
			}
		}
	}()
}

func (mc *MemoryCleanup) Stop() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.running {
		close(mc.stopCh)
		mc.running = false
	}
}

func (mc *MemoryCleanup) cleanup() {
	cutoff := time.Now().Add(-mc.maxAge)
	mc.tracer.mu.Lock()
	defer mc.tracer.mu.Unlock()

	var remaining []*Span
	for _, span := range mc.tracer.spans {
		if span.EndTime.IsZero() || span.EndTime.After(cutoff) {
			remaining = append(remaining, span)
		}
	}
	mc.tracer.spans = remaining
}

type HistogramValue struct {
	Count   int64
	Sum     float64
	Min     float64
	Max     float64
	Buckets []HistogramBucket
}

type HistogramBucket struct {
	UpperBound float64
	Count      int64
}

type MetricsCollector struct {
	name       string
	metrics    map[string]*Metric
	histograms map[string]*HistogramValue
	mu         sync.RWMutex
}

func NewMetricsCollector(name string) *MetricsCollector {
	return &MetricsCollector{
		name:       name,
		metrics:    make(map[string]*Metric),
		histograms: make(map[string]*HistogramValue),
	}
}

func (mc *MetricsCollector) Counter(name string, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := name + labelsKey(labels)
	if m, ok := mc.metrics[key]; ok {
		m.Value++
	} else {
		mc.metrics[key] = &Metric{
			Name:   name,
			Type:   MetricCounter,
			Value:  1,
			Labels: labels,
		}
	}
}

func (mc *MetricsCollector) Gauge(name string, value float64, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := name + labelsKey(labels)
	mc.metrics[key] = &Metric{
		Name:   name,
		Type:   MetricGauge,
		Value:  value,
		Labels: labels,
	}
}

func (mc *MetricsCollector) Histogram(name string, value float64, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := name + labelsKey(labels)
	h, ok := mc.histograms[key]
	if !ok {
		h = &HistogramValue{
			Buckets: []HistogramBucket{
				{UpperBound: 0.01},
				{UpperBound: 0.05},
				{UpperBound: 0.1},
				{UpperBound: 0.5},
				{UpperBound: 1.0},
				{UpperBound: 5.0},
				{UpperBound: 10.0},
				{UpperBound: 50.0},
				{UpperBound: 100.0},
			},
		}
		mc.histograms[key] = h
	}

	h.Count++
	h.Sum += value
	if h.Count == 1 || value < h.Min {
		h.Min = value
	}
	if h.Count == 1 || value > h.Max {
		h.Max = value
	}

	for i := range h.Buckets {
		if value <= h.Buckets[i].UpperBound {
			h.Buckets[i].Count++
		}
	}

	mc.metrics[key] = &Metric{
		Name:   name,
		Type:   MetricHistogram,
		Value:  value,
		Labels: labels,
	}
}

func (mc *MetricsCollector) GetMetrics() []*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := make([]*Metric, 0, len(mc.metrics))
	for _, m := range mc.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

func (mc *MetricsCollector) GetHistogram(name string) *HistogramValue {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	h, ok := mc.histograms[name]
	if !ok {
		return nil
	}
	return h
}

func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metrics = make(map[string]*Metric)
	mc.histograms = make(map[string]*HistogramValue)
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}

type Tracer struct {
	name  string
	spans []*Span
	mu    sync.RWMutex
}

func NewTracer(name string) *Tracer {
	return &Tracer{
		name:  name,
		spans: make([]*Span, 0),
	}
}

func (t *Tracer) StartSpan(name string) *Span {
	span := &Span{
		TraceID:    generateID(),
		SpanID:     generateID(),
		Name:       name,
		StartTime:  time.Now(),
		Attributes: make(map[string]any),
		Events:     make([]SpanEvent, 0),
	}

	t.mu.Lock()
	t.spans = append(t.spans, span)
	t.mu.Unlock()

	return span
}

func (t *Tracer) StartSpanWithParent(name, parentID string) *Span {
	var traceID string
	t.mu.RLock()
	for _, s := range t.spans {
		if s.SpanID == parentID {
			traceID = s.TraceID
			break
		}
	}
	t.mu.RUnlock()

	if traceID == "" {
		traceID = generateID()
	}

	span := &Span{
		TraceID:    traceID,
		SpanID:     generateID(),
		ParentID:   parentID,
		Name:       name,
		StartTime:  time.Now(),
		Attributes: make(map[string]any),
		Events:     make([]SpanEvent, 0),
	}

	t.mu.Lock()
	t.spans = append(t.spans, span)
	t.mu.Unlock()

	return span
}

func (t *Tracer) GetSpans() []*Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.spans
}

func (t *Tracer) GetSpansByTrace(traceID string) []*Span {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var spans []*Span
	for _, span := range t.spans {
		if span.TraceID == traceID {
			spans = append(spans, span)
		}
	}
	return spans
}

func (t *Tracer) EndSpan(span *Span) {
	span.EndTime = time.Now()
}

func (t *Tracer) AddEvent(span *Span, name string, attributes map[string]any) {
	event := SpanEvent{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: attributes,
	}
	span.Events = append(span.Events, event)
}

func (t *Tracer) SetStatus(span *Span, code SpanStatusCode, message string) {
	span.Status = SpanStatus{
		Code:    code,
		Message: message,
	}
}

func (t *Tracer) SpanCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.spans)
}

func (t *Tracer) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = nil
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type LogEntry struct {
	Timestamp  time.Time
	Level      LogLevel
	Message    string
	Attributes map[string]any
	Source     string
}

type Logger struct {
	name    string
	entries []LogEntry
	level   LogLevel
	mu      sync.RWMutex
}

func NewLogger(name string, level LogLevel) *Logger {
	return &Logger{
		name:    name,
		entries: make([]LogEntry, 0),
		level:   level,
	}
}

func (l *Logger) Debug(msg string, attrs map[string]any) {
	l.log(LogLevelDebug, msg, attrs)
}

func (l *Logger) Info(msg string, attrs map[string]any) {
	l.log(LogLevelInfo, msg, attrs)
}

func (l *Logger) Warn(msg string, attrs map[string]any) {
	l.log(LogLevelWarn, msg, attrs)
}

func (l *Logger) Error(msg string, attrs map[string]any) {
	l.log(LogLevelError, msg, attrs)
}

func (l *Logger) log(level LogLevel, msg string, attrs map[string]any) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      level,
		Message:    msg,
		Attributes: attrs,
		Source:     l.name,
	}

	l.mu.Lock()
	l.entries = append(l.entries, entry)
	l.mu.Unlock()
}

func (l *Logger) GetEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.entries
}

func (l *Logger) GetEntriesByLevel(level LogLevel) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var entries []LogEntry
	for _, entry := range l.entries {
		if entry.Level == level {
			entries = append(entries, entry)
		}
	}
	return entries
}

type MetricType string

const (
	MetricCounter   MetricType = "counter"
	MetricGauge     MetricType = "gauge"
	MetricHistogram MetricType = "histogram"
)

type Metric struct {
	Name   string
	Type   MetricType
	Value  float64
	Labels map[string]string
}

type Stack struct {
	Tracer  *Tracer
	Logger  *Logger
	Metrics *MetricsCollector
}

func NewStack(name string) *Stack {
	return &Stack{
		Tracer:  NewTracer(name),
		Logger:  NewLogger(name, LogLevelInfo),
		Metrics: NewMetricsCollector(name),
	}
}

func labelsKey(labels map[string]string) string {
	key := ""
	for k, v := range labels {
		key += k + "=" + v + ","
	}
	return key
}
