package kernel

import (
	"errors"
	"testing"
)

type simpleService struct{ name string }

type lifecycleService struct {
	name        string
	initErr     error
	startErr    error
	stopErr     error
	initialized bool
	started     bool
	stopped     bool
}

func (s *lifecycleService) Initialize() error { s.initialized = true; return s.initErr }
func (s *lifecycleService) Start() error      { s.started = true; return s.startErr }
func (s *lifecycleService) Stop() error       { s.stopped = true; return s.stopErr }

func TestRegisterEmptyName(t *testing.T) {
	k := NewKernel()
	err := k.Register("", "service")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRegisterNilService(t *testing.T) {
	k := NewKernel()
	err := k.Register("svc", nil)
	if err == nil {
		t.Error("expected error for nil service")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	k := NewKernel()
	k.Register("svc", "service1")
	err := k.Register("svc", "service2")
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestResolveNotFound(t *testing.T) {
	k := NewKernel()
	_, err := k.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent service")
	}
}

func TestRegisteredServicesEmpty(t *testing.T) {
	k := NewKernel()
	names := k.RegisteredServices()
	if len(names) != 0 {
		t.Errorf("expected empty, got %v", names)
	}
}

func TestRegisteredServicesMultiple(t *testing.T) {
	k := NewKernel()
	k.Register("z", "last")
	k.Register("a", "first")
	k.Register("m", "middle")

	names := k.RegisteredServices()
	if len(names) != 3 {
		t.Fatalf("expected 3, got %d", len(names))
	}
	if names[0] != "a" || names[1] != "m" || names[2] != "z" {
		t.Errorf("expected sorted order, got %v", names)
	}
}

func TestTopics(t *testing.T) {
	k := NewKernel()
	topics := k.Topics()
	if len(topics) != 0 {
		t.Errorf("expected empty topics, got %v", topics)
	}
}

func TestStartAlreadyStarted(t *testing.T) {
	k := NewKernel()
	k.Register("svc", &lifecycleService{name: "svc"})
	k.Start()
	err := k.Start()
	if err == nil {
		t.Error("expected error for double start")
	}
}

func TestStartServiceInitFailure(t *testing.T) {
	k := NewKernel()
	k.Register("svc", &lifecycleService{name: "svc", initErr: errors.New("init failed")})
	err := k.Start()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStartServiceStartFailure(t *testing.T) {
	k := NewKernel()
	k.Register("svc", &lifecycleService{name: "svc", startErr: errors.New("start failed")})
	err := k.Start()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStopNotStarted(t *testing.T) {
	k := NewKernel()
	err := k.Stop()
	if err == nil {
		t.Error("expected error for stop without start")
	}
}

func TestStopServiceFailure(t *testing.T) {
	k := NewKernel()
	svc := &lifecycleService{name: "svc", stopErr: errors.New("stop failed")}
	k.Register("svc", svc)
	k.Start()
	err := k.Stop()
	if err == nil {
		t.Fatal("expected error")
	}
	if !svc.stopped {
		t.Error("expected service to be stopped despite error")
	}
}

func TestStartStopWithNonLifecycleService(t *testing.T) {
	k := NewKernel()
	k.Register("svc", &simpleService{name: "svc"})
	if err := k.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := k.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
}

func TestEmitTelemetryEmptyName(t *testing.T) {
	k := NewKernel()
	err := k.EmitTelemetry(TelemetryEvent{Name: ""})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestEmitTelemetryRecordsMetrics(t *testing.T) {
	k := NewKernel()
	k.EmitTelemetry(TelemetryEvent{Name: "test.event", Timestamp: 100, Payload: map[string]any{"key": "val"}})
	m := k.Metrics()
	if m.Events != 1 {
		t.Errorf("expected 1 event, got %d", m.Events)
	}
	if m.LastEvent.Name != "test.event" {
		t.Errorf("expected test.event, got %s", m.LastEvent.Name)
	}
}

func TestPublishSubscribeRoundTrip(t *testing.T) {
	k := NewKernel()
	var received any
	k.Subscribe("topic", func(payload any) {
		received = payload
	})
	k.Publish("topic", "data")
	if received != "data" {
		t.Errorf("expected 'data', got %v", received)
	}
}

func TestKernelMultipleServicesStartStop(t *testing.T) {
	k := NewKernel()
	s1 := &lifecycleService{name: "s1"}
	s2 := &lifecycleService{name: "s2"}
	k.Register("s1", s1)
	k.Register("s2", s2)

	if err := k.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if !s1.initialized || !s2.initialized {
		t.Error("expected both services to be initialized")
	}
	if !s1.started || !s2.started {
		t.Error("expected both services to be started")
	}

	if err := k.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if !s1.stopped || !s2.stopped {
		t.Error("expected both services to be stopped")
	}
}

func TestKernelConcurrentSafe(t *testing.T) {
	k := NewKernel()
	k.Register("svc", &simpleService{name: "svc"})

	t.Run("parallel start stop", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 5; i++ {
			k.Start()
			k.Stop()
		}
	})
	t.Run("parallel resolve", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 10; i++ {
			k.Resolve("svc")
		}
	})
	t.Run("parallel register", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 5; i++ {
			k.Register("reg", "service")
		}
	})
}
