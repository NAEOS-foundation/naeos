// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"testing"
	"time"
)

func TestNewTelemetry(t *testing.T) {
	telemetry := NewTelemetry()
	if telemetry == nil {
		t.Fatal("expected non-nil telemetry")
	}
	m := telemetry.Metrics()
	if m.Events != 0 {
		t.Fatalf("expected 0 events, got %d", m.Events)
	}
}

func TestEmit(t *testing.T) {
	telemetry := NewTelemetry()
	event := TelemetryEvent{
		Name:    "test-event",
		Payload: map[string]any{"key": "value"},
	}
	err := telemetry.Emit(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := telemetry.Metrics()
	if m.Events != 1 {
		t.Fatalf("expected 1 event, got %d", m.Events)
	}
	if m.LastEvent.Name != "test-event" {
		t.Fatalf("expected last event name 'test-event', got %s", m.LastEvent.Name)
	}
}

func TestEmitAutoTimestamp(t *testing.T) {
	telemetry := NewTelemetry()
	event := TelemetryEvent{Name: "test"}
	_ = telemetry.Emit(event)
	m := telemetry.Metrics()
	if m.LastEvent.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestEvents(t *testing.T) {
	telemetry := NewTelemetry()
	_ = telemetry.Emit(TelemetryEvent{Name: "event-1"})
	_ = telemetry.Emit(TelemetryEvent{Name: "event-2"})
	_ = telemetry.Emit(TelemetryEvent{Name: "event-3"})

	events := telemetry.Events()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestEventsByName(t *testing.T) {
	telemetry := NewTelemetry()
	_ = telemetry.Emit(TelemetryEvent{Name: "build"})
	_ = telemetry.Emit(TelemetryEvent{Name: "test"})
	_ = telemetry.Emit(TelemetryEvent{Name: "build"})

	buildEvents := telemetry.EventsByName("build")
	if len(buildEvents) != 2 {
		t.Fatalf("expected 2 build events, got %d", len(buildEvents))
	}
}

func TestEventTypes(t *testing.T) {
	telemetry := NewTelemetry()
	_ = telemetry.Emit(TelemetryEvent{Name: "build"})
	_ = telemetry.Emit(TelemetryEvent{Name: "build"})
	_ = telemetry.Emit(TelemetryEvent{Name: "test"})

	m := telemetry.Metrics()
	if m.EventTypes["build"] != 2 {
		t.Fatalf("expected 2 build events, got %d", m.EventTypes["build"])
	}
	if m.EventTypes["test"] != 1 {
		t.Fatalf("expected 1 test event, got %d", m.EventTypes["test"])
	}
}

func TestReset(t *testing.T) {
	telemetry := NewTelemetry()
	_ = telemetry.Emit(TelemetryEvent{Name: "test"})
	telemetry.Reset()

	m := telemetry.Metrics()
	if m.Events != 0 {
		t.Fatalf("expected 0 events after reset, got %d", m.Events)
	}
	events := telemetry.Events()
	if len(events) != 0 {
		t.Fatalf("expected 0 events after reset, got %d", len(events))
	}
}

func TestMetricsStartTime(t *testing.T) {
	before := time.Now()
	telemetry := NewTelemetry()
	after := time.Now()

	m := telemetry.Metrics()
	if m.StartTime.Before(before) || m.StartTime.After(after) {
		t.Fatal("start time not within expected range")
	}
}
