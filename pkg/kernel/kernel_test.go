package kernel

import (
	"testing"
)

type testService struct {
	initialized bool
	started     bool
	stopped     bool
}

func (s *testService) Initialize() error {
	s.initialized = true
	return nil
}

func (s *testService) Start() error {
	if !s.initialized {
		return errNotInitialized
	}
	s.started = true
	return nil
}

func (s *testService) Stop() error {
	if !s.started {
		return errNotStarted
	}
	s.stopped = true
	return nil
}

var (
	errNotInitialized = &kernelError{"not initialized"}
	errNotStarted     = &kernelError{"not started"}
)

type kernelError struct {
	msg string
}

func (e *kernelError) Error() string {
	return e.msg
}

func TestKernelRegisterResolve(t *testing.T) {
	k := NewKernel()
	service := &testService{}

	if err := k.Register("demo", service); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resolved, err := k.Resolve("demo")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if resolved != service {
		t.Fatalf("expected resolved service to match registered instance")
	}
}

func TestKernelStartStop(t *testing.T) {
	k := NewKernel()
	service := &testService{}
	if err := k.Register("demo", service); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if err := k.Start(); err != nil {
		t.Fatalf("kernel start failed: %v", err)
	}
	if !service.initialized || !service.started {
		t.Fatal("expected service to be initialized and started")
	}

	if err := k.Stop(); err != nil {
		t.Fatalf("kernel stop failed: %v", err)
	}
	if !service.stopped {
		t.Fatal("expected service to be stopped")
	}
}

func TestKernelEventBus(t *testing.T) {
	k := NewKernel()
	var received any

	if err := k.Subscribe("topic", func(payload any) {
		received = payload
	}); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	k.Publish("topic", "payload")
	if received != "payload" {
		t.Fatalf("expected received payload to be 'payload', got %v", received)
	}
}

func TestKernelTelemetry(t *testing.T) {
	k := NewKernel()
	event := TelemetryEvent{Name: "test-event", Timestamp: 123, Payload: map[string]any{"key": "value"}}
	if err := k.EmitTelemetry(event); err != nil {
		t.Fatalf("emit telemetry failed: %v", err)
	}
	metrics := k.Metrics()
	if metrics.Events != 1 {
		t.Fatalf("expected 1 event, got %d", metrics.Events)
	}
	if metrics.LastEvent.Name != event.Name {
		t.Fatalf("expected last event name %q, got %q", event.Name, metrics.LastEvent.Name)
	}
}
