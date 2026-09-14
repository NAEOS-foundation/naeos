// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBusPublishSubscribe(t *testing.T) {
	bus := NewBus()
	var received string

	bus.Subscribe("test", func(e Event) error {
		received = e.Payload.(string)
		return nil
	})

	bus.Publish("test", "hello")

	if received != "hello" {
		t.Errorf("expected 'hello', got %q", received)
	}
}

func TestBusWildcard(t *testing.T) {
	bus := NewBus()
	var count int32

	bus.Subscribe("*", func(e Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	bus.Publish("topic1", "a")
	bus.Publish("topic2", "b")
	bus.Publish("topic3", "c")

	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("expected 3 wildcard calls, got %d", atomic.LoadInt32(&count))
	}
}

func TestBusMiddleware(t *testing.T) {
	bus := NewBus()
	var order []string

	bus.Use(func(e Event, next Handler) error {
		order = append(order, "before")
		return next(e)
	})

	bus.Subscribe("test", func(e Event) error {
		order = append(order, "handler")
		return nil
	})

	bus.Publish("test", nil)

	if len(order) != 2 || order[0] != "before" || order[1] != "handler" {
		t.Errorf("expected [before handler], got %v", order)
	}
}

func TestBusHistory(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("test", func(e Event) error { return nil })
	bus.Publish("test", "a")
	bus.Publish("test", "b")

	history := bus.History()
	if len(history) != 2 {
		t.Errorf("expected 2 history events, got %d", len(history))
	}
	if history[0].ID == "" {
		t.Error("expected non-empty event ID")
	}
}

func TestBusHistorySince(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("test", func(e Event) error { return nil })

	bus.Publish("test", "old")
	time.Sleep(10 * time.Millisecond)
	cutoff := time.Now()
	time.Sleep(10 * time.Millisecond)
	bus.Publish("test", "new")

	history := bus.HistorySince(cutoff)
	if len(history) != 1 {
		t.Errorf("expected 1 event since cutoff, got %d", len(history))
	}
}

func TestBusHistoryByTopic(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("a", func(e Event) error { return nil })
	bus.Subscribe("b", func(e Event) error { return nil })

	bus.Publish("a", "1")
	bus.Publish("b", "2")
	bus.Publish("a", "3")

	history := bus.HistoryByTopic("a")
	if len(history) != 2 {
		t.Errorf("expected 2 events for topic a, got %d", len(history))
	}
}

func TestBusHistoryFilter(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("*", func(e Event) error { return nil })

	bus.Publish("api.v1", "a")
	bus.Publish("api.v2", "b")
	bus.Publish("db", "c")

	filtered := bus.HistoryFilter("api.*", time.Time{}, 10)
	if len(filtered) != 2 {
		t.Errorf("expected 2 api events, got %d", len(filtered))
	}

	limited := bus.HistoryFilter("", time.Time{}, 1)
	if len(limited) != 1 {
		t.Errorf("expected 1 limited event, got %d", len(limited))
	}
}

func TestBusClearHistory(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("test", func(e Event) error { return nil })
	bus.Publish("test", "a")

	bus.ClearHistory()

	if len(bus.History()) != 0 {
		t.Error("expected empty history after clear")
	}
}

func TestAsyncHandler(t *testing.T) {
	var count int32

	ah := NewAsyncHandler(func(e Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}, 2, 10)

	ah.Start()

	for i := 0; i < 10; i++ {
		ah.ch <- Event{Topic: "test"}
	}

	time.Sleep(50 * time.Millisecond)
	ah.Stop()

	if atomic.LoadInt32(&count) != 10 {
		t.Errorf("expected 10 processed, got %d", atomic.LoadInt32(&count))
	}
}

func TestAsyncHandlerDeadLetter(t *testing.T) {
	ah := NewAsyncHandler(func(e Event) error {
		return errors.New("fail")
	}, 1, 10)

	ah.Start()

	ah.ch <- Event{Topic: "test"}
	time.Sleep(50 * time.Millisecond)
	ah.Stop()

	if ah.DeadLetterCount() != 1 {
		t.Errorf("expected 1 dead letter, got %d", ah.DeadLetterCount())
	}

	dl := ah.DeadLetters()
	if len(dl) != 1 {
		t.Errorf("expected 1 dead letter event, got %d", len(dl))
	}
}

func TestAsyncHandlerConcurrency(t *testing.T) {
	var count int32

	ah := NewAsyncHandler(func(e Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}, 4, 100)

	ah.Start()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ah.ch <- Event{Topic: "test"}
		}()
	}
	wg.Wait()

	time.Sleep(100 * time.Millisecond)
	ah.Stop()

	if atomic.LoadInt32(&count) != 100 {
		t.Errorf("expected 100, got %d", atomic.LoadInt32(&count))
	}
}

func TestBusNilHandler(t *testing.T) {
	bus := NewBus()
	err := bus.Subscribe("test", nil)
	if err == nil {
		t.Error("expected error for nil handler")
	}
}

func TestBusUnsubscribeNotFound(t *testing.T) {
	bus := NewBus()
	err := bus.Unsubscribe("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent topic")
	}
}

func TestBusSubscriberCount(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("test", func(e Event) error { return nil })
	bus.Subscribe("test", func(e Event) error { return nil })

	if bus.SubscriberCount("test") != 2 {
		t.Errorf("expected 2 subscribers, got %d", bus.SubscriberCount("test"))
	}
}

func TestBusTopics(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("a", func(e Event) error { return nil })
	bus.Subscribe("b", func(e Event) error { return nil })

	topics := bus.Topics()
	if len(topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(topics))
	}
}

func TestEventID(t *testing.T) {
	bus := NewBus()

	bus.Subscribe("test", func(e Event) error { return nil })

	bus.Publish("test", "a")
	bus.Publish("test", "b")

	history := bus.History()
	if history[0].ID == history[1].ID {
		t.Error("expected different event IDs")
	}
}

func TestMatchTopic(t *testing.T) {
	tests := []struct {
		topic   string
		pattern string
		want    bool
	}{
		{"api.v1", "*", true},
		{"api.v1", "api.*", true},
		{"db.v1", "api.*", false},
		{"api.v1", "api.v1", true},
		{"api.v2", "api.v1", false},
	}

	for _, tt := range tests {
		got := matchTopic(tt.topic, tt.pattern)
		if got != tt.want {
			t.Errorf("matchTopic(%q, %q) = %v, want %v", tt.topic, tt.pattern, got, tt.want)
		}
	}
}

func TestBusMaxHistory(t *testing.T) {
	bus := NewBus()
	bus.maxHistory = 5

	bus.Subscribe("test", func(e Event) error { return nil })

	for i := 0; i < 10; i++ {
		bus.Publish("test", i)
	}

	if len(bus.History()) != 5 {
		t.Errorf("expected 5 history events (max), got %d", len(bus.History()))
	}
}
