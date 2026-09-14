// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"sync"
	"time"
)

type TelemetryEvent struct {
	Name      string
	Timestamp time.Time
	Payload   map[string]any
}

type Metrics struct {
	Events     int
	LastEvent  TelemetryEvent
	StartTime  time.Time
	EventTypes map[string]int
}

type TelemetrySink interface {
	Emit(event TelemetryEvent) error
	Metrics() Metrics
}

type DefaultTelemetry struct {
	mu      sync.RWMutex
	events  []TelemetryEvent
	metrics Metrics
}

func NewTelemetry() *DefaultTelemetry {
	return &DefaultTelemetry{
		metrics: Metrics{
			StartTime:  time.Now(),
			EventTypes: make(map[string]int),
		},
	}
}

func (t *DefaultTelemetry) Emit(event TelemetryEvent) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	t.events = append(t.events, event)
	t.metrics.Events++
	t.metrics.LastEvent = event
	t.metrics.EventTypes[event.Name]++

	return nil
}

func (t *DefaultTelemetry) Metrics() Metrics {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.metrics
}

func (t *DefaultTelemetry) Events() []TelemetryEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]TelemetryEvent, len(t.events))
	copy(result, t.events)
	return result
}

func (t *DefaultTelemetry) EventsByName(name string) []TelemetryEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var result []TelemetryEvent
	for _, e := range t.events {
		if e.Name == name {
			result = append(result, e)
		}
	}
	return result
}

func (t *DefaultTelemetry) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = nil
	t.metrics = Metrics{
		StartTime:  time.Now(),
		EventTypes: make(map[string]int),
	}
}
