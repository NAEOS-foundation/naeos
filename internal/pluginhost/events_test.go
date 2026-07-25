package pluginhost

import (
	"errors"
	"testing"
)

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus()
	if bus == nil {
		t.Fatal("expected non-nil bus")
	}
	if len(bus.subscriptions) != 0 {
		t.Error("expected empty subscriptions")
	}
}

func TestSubscribeAndEmit(t *testing.T) {
	bus := NewEventBus()
	called := false
	bus.Subscribe(EventBeforeParse, "plugin-a", func(pluginName string, data *EventData) error {
		called = true
		if pluginName != "plugin-a" {
			t.Errorf("expected plugin-a, got %s", pluginName)
		}
		if data.PipelineID != "p1" {
			t.Errorf("expected pipeline p1, got %s", data.PipelineID)
		}
		return nil
	})

	errs := bus.Emit(EventBeforeParse, &EventData{PipelineID: "p1"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestEmitNoSubscribers(t *testing.T) {
	bus := NewEventBus()
	errs := bus.Emit(EventAfterParse, &EventData{})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestEmitMultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	count := 0
	bus.Subscribe(EventAfterGenerate, "p1", func(_ string, _ *EventData) error {
		count++
		return nil
	})
	bus.Subscribe(EventAfterGenerate, "p2", func(_ string, _ *EventData) error {
		count++
		return nil
	})

	bus.Emit(EventAfterGenerate, &EventData{})
	if count != 2 {
		t.Errorf("expected 2 calls, got %d", count)
	}
}

func TestEmitHandlerError(t *testing.T) {
	bus := NewEventBus()
	bus.Subscribe(EventBeforeGenerate, "failing", func(_ string, _ *EventData) error {
		return errors.New("handler failed")
	})

	errs := bus.Emit(EventBeforeGenerate, &EventData{})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Error() != "PLUGIN_ERROR: plugin failing: handler failed" {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewEventBus()
	called := false
	bus.Subscribe(EventOnPipelineComplete, "p1", func(_ string, _ *EventData) error {
		called = true
		return nil
	})

	bus.Unsubscribe(EventOnPipelineComplete, "p1")
	bus.Emit(EventOnPipelineComplete, &EventData{})
	if called {
		t.Error("expected handler not to be called after unsubscribe")
	}
}

func TestUnsubscribeNonExistent(t *testing.T) {
	bus := NewEventBus()
	bus.Unsubscribe(EventBeforeParse, "nonexistent")
}

func TestSubscribers(t *testing.T) {
	bus := NewEventBus()
	bus.Subscribe(EventAfterParse, "a", func(_ string, _ *EventData) error { return nil })
	bus.Subscribe(EventAfterParse, "b", func(_ string, _ *EventData) error { return nil })

	names := bus.Subscribers(EventAfterParse)
	if len(names) != 2 {
		t.Errorf("expected 2 subscribers, got %d", len(names))
	}
}

func TestSubscribersEmpty(t *testing.T) {
	bus := NewEventBus()
	names := bus.Subscribers(EventBeforeGenerate)
	if len(names) != 0 {
		t.Errorf("expected 0 subscribers, got %d", len(names))
	}
}

func TestHasSubscribers(t *testing.T) {
	bus := NewEventBus()
	if bus.HasSubscribers(EventBeforeParse) {
		t.Error("expected no subscribers")
	}

	bus.Subscribe(EventBeforeParse, "p", func(_ string, _ *EventData) error { return nil })
	if !bus.HasSubscribers(EventBeforeParse) {
		t.Error("expected subscribers")
	}
}

func TestOverwriteSubscription(t *testing.T) {
	bus := NewEventBus()
	callCount := 0
	bus.Subscribe(EventBeforeParse, "p1", func(_ string, _ *EventData) error {
		callCount++
		return nil
	})
	bus.Subscribe(EventBeforeParse, "p1", func(_ string, _ *EventData) error {
		callCount += 10
		return nil
	})

	bus.Emit(EventBeforeParse, &EventData{})
	if callCount != 10 {
		t.Errorf("expected callCount 10 (overwrite), got %d", callCount)
	}
}

// --- PluginEventBus Tests ---

func TestNewPluginEventBus(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	if peb == nil {
		t.Fatal("expected non-nil PluginEventBus")
	}
}

func TestPluginEventBusOnPipelineStart(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	peb.OnPipelineStart("p1")
}

func TestPluginEventBusOnPipelineComplete(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	called := false
	bus.Subscribe(EventOnPipelineComplete, "obs", func(_ string, data *EventData) error {
		called = true
		if data.PipelineID != "p1" {
			t.Errorf("expected pipeline p1, got %s", data.PipelineID)
		}
		if data.Artifacts != 5 {
			t.Errorf("expected 5 artifacts, got %d", data.Artifacts)
		}
		return nil
	})

	peb.OnPipelineComplete("p1", 5, "1.2s")
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestPluginEventBusOnPipelineFailed(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	called := false
	bus.Subscribe(EventOnPipelineFailed, "obs", func(_ string, data *EventData) error {
		called = true
		if data.Error != "something broke" {
			t.Errorf("expected error msg, got %s", data.Error)
		}
		return nil
	})

	peb.OnPipelineFailed("p1", "something broke")
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestPluginEventBusOnArtifactGenerated(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	peb.OnArtifactGenerated("file.go", "/out/file.go")
}

func TestPluginEventBusEmitBeforeParse(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	called := false
	bus.Subscribe(EventBeforeParse, "p1", func(_ string, data *EventData) error {
		called = true
		if data.Stage != "parse" {
			t.Errorf("expected stage parse, got %s", data.Stage)
		}
		return nil
	})

	errs := peb.EmitBeforeParse("p1")
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestPluginEventBusEmitAfterParse(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	called := false
	bus.Subscribe(EventAfterParse, "p1", func(_ string, _ *EventData) error {
		called = true
		return nil
	})

	peb.EmitAfterParse("p1")
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestPluginEventBusEmitBeforeGenerate(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	called := false
	bus.Subscribe(EventBeforeGenerate, "p1", func(_ string, _ *EventData) error {
		called = true
		return nil
	})

	peb.EmitBeforeGenerate("p1")
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestPluginEventBusEmitAfterGenerate(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	called := false
	bus.Subscribe(EventAfterGenerate, "p1", func(_ string, _ *EventData) error {
		called = true
		return nil
	})

	peb.EmitAfterGenerate("p1")
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestPluginEventBusMultipleEvents(t *testing.T) {
	bus := NewEventBus()
	peb := NewPluginEventBus(bus)
	events := make([]PipelineEvent, 0)

	handler := func(_ string, _ *EventData) error {
		return nil
	}

	bus.Subscribe(EventBeforeParse, "p1", handler)
	bus.Subscribe(EventAfterParse, "p1", handler)
	bus.Subscribe(EventBeforeGenerate, "p1", handler)
	bus.Subscribe(EventAfterGenerate, "p1", handler)
	bus.Subscribe(EventOnPipelineComplete, "p1", handler)

	_ = peb.EmitBeforeParse("p1")
	events = append(events, EventBeforeParse)
	_ = peb.EmitAfterParse("p1")
	events = append(events, EventAfterParse)
	_ = peb.EmitBeforeGenerate("p1")
	events = append(events, EventBeforeGenerate)
	_ = peb.EmitAfterGenerate("p1")
	events = append(events, EventAfterGenerate)
	peb.OnPipelineComplete("p1", 0, "0s")
	events = append(events, EventOnPipelineComplete)

	for _, e := range events {
		if !bus.HasSubscribers(e) {
			t.Errorf("expected subscribers for %s", e)
		}
	}
}
