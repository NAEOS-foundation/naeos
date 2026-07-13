package pluginhost

import (
	"fmt"
	"sync"
)

// Logger provides structured logging for plugins.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

// MetricsCollector provides metrics collection for plugins.
type MetricsCollector interface {
	CounterInc(name string, labels map[string]string)
	GaugeSet(name string, value float64, labels map[string]string)
	HistogramObserve(name string, value float64, labels map[string]string)
}

// EventEmitter provides inter-plugin event communication.
type EventEmitter interface {
	Emit(event string, data any)
	On(event string, handler func(data any))
}

// SimpleLogger is a Logger implementation that writes to stdout.
type SimpleLogger struct {
	prefix string
}

// NewSimpleLogger creates a new SimpleLogger with the given prefix.
func NewSimpleLogger(prefix string) *SimpleLogger {
	return &SimpleLogger{prefix: prefix}
}

func (l *SimpleLogger) Info(msg string, args ...any) {
	fmt.Printf("[%s] INFO: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

func (l *SimpleLogger) Warn(msg string, args ...any) {
	fmt.Printf("[%s] WARN: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

func (l *SimpleLogger) Error(msg string, args ...any) {
	fmt.Printf("[%s] ERROR: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

func (l *SimpleLogger) Debug(msg string, args ...any) {
	fmt.Printf("[%s] DEBUG: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

// SimpleMetrics is a no-op MetricsCollector.
type SimpleMetrics struct {
	mu        sync.RWMutex
	counters  map[string]float64
	gauges    map[string]float64
	histograms map[string][]float64
}

// NewSimpleMetrics creates a new SimpleMetrics.
func NewSimpleMetrics() *SimpleMetrics {
	return &SimpleMetrics{
		counters:  make(map[string]float64),
		gauges:    make(map[string]float64),
		histograms: make(map[string][]float64),
	}
}

func (m *SimpleMetrics) CounterInc(name string, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := metricKey(name, labels)
	m.counters[key]++
}

func (m *SimpleMetrics) GaugeSet(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := metricKey(name, labels)
	m.gauges[key] = value
}

func (m *SimpleMetrics) HistogramObserve(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := metricKey(name, labels)
	m.histograms[key] = append(m.histograms[key], value)
}

func (m *SimpleMetrics) Snapshot() (counters, gauges map[string]float64, histograms map[string][]float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counters = make(map[string]float64, len(m.counters))
	for k, v := range m.counters {
		counters[k] = v
	}

	gauges = make(map[string]float64, len(m.gauges))
	for k, v := range m.gauges {
		gauges[k] = v
	}

	histograms = make(map[string][]float64, len(m.histograms))
	for k, v := range m.histograms {
		histograms[k] = append([]float64{}, v...)
	}

	return counters, gauges, histograms
}

func metricKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += "|" + k + "=" + v
	}
	return key
}

// SimpleEventEmitter is an in-memory EventEmitter.
type SimpleEventEmitter struct {
	handlers map[string][]func(data any)
	mu       sync.RWMutex
}

// NewSimpleEventEmitter creates a new SimpleEventEmitter.
func NewSimpleEventEmitter() *SimpleEventEmitter {
	return &SimpleEventEmitter{
		handlers: make(map[string][]func(data any)),
	}
}

func (e *SimpleEventEmitter) Emit(event string, data any) {
	e.mu.RLock()
	handlers := e.handlers[event]
	e.mu.RUnlock()

	for _, handler := range handlers {
		handler(data)
	}
}

func (e *SimpleEventEmitter) On(event string, handler func(data any)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[event] = append(e.handlers[event], handler)
}
