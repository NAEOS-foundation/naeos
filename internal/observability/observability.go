package observability

import (
	"sync"
	"time"
)

// Tracing

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

// Logging

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

// Metrics

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

type MetricsCollector struct {
	name    string
	metrics map[string]*Metric
	mu      sync.RWMutex
}

func NewMetricsCollector(name string) *MetricsCollector {
	return &MetricsCollector{
		name:    name,
		metrics: make(map[string]*Metric),
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

func labelsKey(labels map[string]string) string {
	key := ""
	for k, v := range labels {
		key += k + "=" + v + ","
	}
	return key
}

// Observability Stack

type Stack struct {
	Tracer   *Tracer
	Logger   *Logger
	Metrics  *MetricsCollector
}

func NewStack(name string) *Stack {
	return &Stack{
		Tracer:  NewTracer(name),
		Logger:  NewLogger(name, LogLevelInfo),
		Metrics: NewMetricsCollector(name),
	}
}

// Helpers

func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}
