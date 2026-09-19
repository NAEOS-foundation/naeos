// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package messagequeue

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueuePublishConsume(t *testing.T) {
	q := NewQueue("test", 10)
	received := make(chan string, 1)

	q.Subscribe(func(msg *Message) error {
		received <- msg.Payload.(string)
		return nil
	})

	q.Publish(NewMessage("test", "hello"))

	select {
	case r := <-received:
		if r != "hello" {
			t.Errorf("expected 'hello', got %q", r)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}

	q.Stop()
}

func TestQueueRetry(t *testing.T) {
	q := NewQueue("test", 10)
	var attempts int32

	q.Subscribe(func(msg *Message) error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("retry")
	})

	q.Publish(NewMessage("test", "fail"))
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts (1+2 retries with MaxRetries=3), got %d", atomic.LoadInt32(&attempts))
	}

	q.Stop()
}

func TestQueueDeadLetter(t *testing.T) {
	q := NewQueue("test", 10)

	q.Subscribe(func(msg *Message) error {
		return errors.New("permanent failure")
	})

	msg := NewMessage("test", "dead")
	msg.MaxRetries = 1
	q.Publish(msg)
	time.Sleep(100 * time.Millisecond)

	dead := q.DeadLetters()
	if len(dead) != 1 {
		t.Errorf("expected 1 dead letter, got %d", len(dead))
	}

	q.Stop()
}

func TestQueueStats(t *testing.T) {
	q := NewQueue("test", 10)

	q.Subscribe(func(msg *Message) error {
		return nil
	})

	q.Publish(NewMessage("test", "1"))
	q.Publish(NewMessage("test", "2"))
	time.Sleep(50 * time.Millisecond)

	stats := q.Stats()
	if stats.Published != 2 {
		t.Errorf("expected 2 published, got %d", stats.Published)
	}
	if stats.Consumed != 2 {
		t.Errorf("expected 2 consumed, got %d", stats.Consumed)
	}

	q.Stop()
}

func TestQueueMetrics(t *testing.T) {
	q := NewQueue("test", 10)

	q.Subscribe(func(msg *Message) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	q.Publish(NewMessage("test", "1"))
	time.Sleep(50 * time.Millisecond)

	metrics := q.Metrics()
	if metrics.LatencyCount != 1 {
		t.Errorf("expected 1 latency count, got %d", metrics.LatencyCount)
	}

	q.Stop()
}

func TestTopicPubSub(t *testing.T) {
	topic := NewTopic("events")
	received := make(chan string, 1)

	topic.Subscribe("sub1", func(msg *Message) error {
		received <- msg.Payload.(string)
		return nil
	})

	topic.Publish(NewMessage("events", "data"))

	select {
	case r := <-received:
		if r != "data" {
			t.Errorf("expected 'data', got %q", r)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}

	if topic.Subscribers() != 1 {
		t.Errorf("expected 1 subscriber, got %d", topic.Subscribers())
	}
}

func TestTopicUnsubscribe(t *testing.T) {
	topic := NewTopic("events")

	topic.Subscribe("sub1", func(msg *Message) error {
		return nil
	})

	if topic.Subscribers() != 1 {
		t.Errorf("expected 1 subscriber, got %d", topic.Subscribers())
	}

	topic.Unsubscribe("sub1")

	if topic.Subscribers() != 0 {
		t.Errorf("expected 0 subscribers, got %d", topic.Subscribers())
	}
}

func TestBroker(t *testing.T) {
	broker := NewBroker()

	topic := broker.CreateTopic("logs")
	if topic == nil {
		t.Fatal("expected topic")
	}

	topic2 := broker.CreateTopic("logs")
	if topic != topic2 {
		t.Error("expected same topic instance")
	}

	topics := broker.ListTopics()
	if len(topics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(topics))
	}

	broker.DeleteTopic("logs")
	_, ok := broker.GetTopic("logs")
	if ok {
		t.Error("expected topic to be deleted")
	}
}

func TestBrokerPublishNotFound(t *testing.T) {
	broker := NewBroker()

	err := broker.Publish("nonexistent", NewMessage("test", "data"))
	if !errors.Is(err, ErrTopicNotFound) {
		t.Errorf("expected ErrTopicNotFound, got %v", err)
	}
}

func TestBrokerSubscribeAutoCreate(t *testing.T) {
	broker := NewBroker()

	received := make(chan struct{}, 1)
	broker.Subscribe("auto-topic", "sub1", func(msg *Message) error {
		received <- struct{}{}
		return nil
	})

	broker.Publish("auto-topic", NewMessage("auto-topic", "data"))

	select {
	case <-received:
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for message")
	}

	broker.Stop()
}

func TestQueueFull(t *testing.T) {
	q := NewQueue("test", 1)

	if err := q.Publish(NewMessage("test", "1")); err != nil {
		t.Fatalf("first publish: %v", err)
	}

	err := q.Publish(NewMessage("test", "2"))
	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}

	q.Stop()
}

func TestQueueConcurrency(t *testing.T) {
	q := NewQueue("test", 100)
	var count int32

	q.Subscribe(func(msg *Message) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Publish(NewMessage("test", "data"))
		}()
	}
	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&count) != 50 {
		t.Errorf("expected 50 consumed, got %d", atomic.LoadInt32(&count))
	}

	q.Stop()
}

// --- Idempotency Tests ---

func TestIdempotencyStoreDedup(t *testing.T) {
	store := NewInMemoryIdempotencyStore(1 * time.Minute)

	isDuplicate, err := store.Dedup("key-1", 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDuplicate {
		t.Error("expected first call to not be a duplicate")
	}

	isDuplicate, err = store.Dedup("key-1", 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isDuplicate {
		t.Error("expected second call with same key to be a duplicate")
	}
}

func TestIdempotencyStoreDifferentKeys(t *testing.T) {
	store := NewInMemoryIdempotencyStore(1 * time.Minute)

	isDuplicate1, err := store.Dedup("key-1", 1*time.Minute)
	if err != nil || isDuplicate1 {
		t.Fatalf("unexpected result for key-1: isDuplicate=%v err=%v", isDuplicate1, err)
	}

	isDuplicate2, err := store.Dedup("key-2", 1*time.Minute)
	if err != nil || isDuplicate2 {
		t.Fatalf("unexpected result for key-2: isDuplicate=%v err=%v", isDuplicate2, err)
	}
}

func TestIdempotencyStoreCleanup(t *testing.T) {
	store := NewInMemoryIdempotencyStore(100 * time.Millisecond)

	_, err := store.Dedup("temp-key", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	store.Cleanup()

	isDuplicate, err := store.Dedup("temp-key", 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDuplicate {
		t.Error("expected cleaned-up key to not be a duplicate")
	}
}

func TestIdempotencyStoreEmptyKey(t *testing.T) {
	store := NewInMemoryIdempotencyStore(1 * time.Minute)

	isDuplicate, err := store.Dedup("", 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDuplicate {
		t.Error("expected empty key to not be a duplicate")
	}
}

func TestNewIdempotentMessage(t *testing.T) {
	msg := NewIdempotentMessage("test", "payload", "idempotency-key-1")
	if msg.IdempotencyKey != "idempotency-key-1" {
		t.Errorf("expected idempotency key 'idempotency-key-1', got %q", msg.IdempotencyKey)
	}
	if msg.Topic != "test" {
		t.Errorf("expected topic 'test', got %q", msg.Topic)
	}
	if msg.Payload != "payload" {
		t.Errorf("expected payload 'payload', got %v", msg.Payload)
	}
	if msg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", msg.MaxRetries)
	}
}

func TestQueueWithIdempotencyStore(t *testing.T) {
	q := NewQueue("test", 10)
	store := NewInMemoryIdempotencyStore(1 * time.Minute)
	q.WithIdempotencyStore(store, 1*time.Minute)

	var count int32
	q.Subscribe(func(msg *Message) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	msg := NewIdempotentMessage("test", "data", "unique-key-1")
	err := q.Publish(msg)
	if err != nil {
		t.Fatalf("unexpected error publishing: %v", err)
	}

	// Publish same idempotency key — should be rejected
	msg2 := NewIdempotentMessage("test", "data", "unique-key-1")
	err = q.Publish(msg2)
	if !errors.Is(err, ErrDuplicateMessage) {
		t.Errorf("expected ErrDuplicateMessage, got %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	q.Stop()

	if count != 1 {
		t.Errorf("expected 1 consumed, got %d", count)
	}
}

func TestQueueIdempotencyStats(t *testing.T) {
	q := NewQueue("test", 10)
	store := NewInMemoryIdempotencyStore(1 * time.Minute)
	q.WithIdempotencyStore(store, 1*time.Minute)

	q.Subscribe(func(msg *Message) error {
		return nil
	})

	// Publish unique message
	msg1 := NewIdempotentMessage("test", "data", "key-1")
	q.Publish(msg1)

	// Publish duplicate
	msg2 := NewIdempotentMessage("test", "data", "key-1")
	q.Publish(msg2)

	time.Sleep(50 * time.Millisecond)
	q.Stop()

	stats := q.Stats()
	if stats.Published != 1 {
		t.Errorf("expected 1 published, got %d", stats.Published)
	}
	if stats.DedupSkipped != 1 {
		t.Errorf("expected 1 dedup skipped, got %d", stats.DedupSkipped)
	}

	idempStats := q.IdempotencyStats()
	if idempStats.DedupHits != 1 {
		t.Errorf("expected 1 idempotency hit, got %d", idempStats.DedupHits)
	}
}

func TestQueueIdempotencyWithoutStore(t *testing.T) {
	q := NewQueue("test", 10)

	var count int32
	q.Subscribe(func(msg *Message) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	// Publish with idempotency key but no store — should work normally
	msg1 := NewIdempotentMessage("test", "data", "key-1")
	err := q.Publish(msg1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg2 := NewIdempotentMessage("test", "data", "key-1")
	err = q.Publish(msg2)
	if err != nil {
		t.Fatalf("unexpected error on second publish without store: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	q.Stop()

	if count != 2 {
		t.Errorf("expected 2 consumed (no store = no dedup), got %d", count)
	}
}

func TestQueueIdempotencyStoreDedupWithTTL(t *testing.T) {
	q := NewQueue("test", 10)
	store := NewInMemoryIdempotencyStore(10 * time.Millisecond)
	q.WithIdempotencyStore(store, 10*time.Millisecond)

	var count int32
	q.Subscribe(func(msg *Message) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	// Publish with TTL
	msg1 := NewIdempotentMessage("test", "data", "ttl-key")
	q.Publish(msg1)

	// Should be rejected while TTL is active
	msg2 := NewIdempotentMessage("test", "data", "ttl-key")
	err := q.Publish(msg2)
	if !errors.Is(err, ErrDuplicateMessage) {
		t.Errorf("expected ErrDuplicateMessage, got %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(30 * time.Millisecond)

	// Should succeed after TTL expires
	msg3 := NewIdempotentMessage("test", "data", "ttl-key")
	err = q.Publish(msg3)
	if err != nil {
		t.Fatalf("unexpected error after TTL expiry: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	q.Stop()

	if count != 2 {
		t.Errorf("expected 2 consumed, got %d", count)
	}
}

func TestQueueProcessedMessageDedup(t *testing.T) {
	q := NewQueue("test", 10)
	store := NewInMemoryIdempotencyStore(1 * time.Minute)
	q.WithIdempotencyStore(store, 1*time.Minute)

	var count int32
	q.Subscribe(func(msg *Message) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	msg := NewMessage("test", "data")
	msg.Processed = true
	q.Publish(msg)

	time.Sleep(50 * time.Millisecond)
	q.Stop()

	stats := q.Stats()
	if stats.DedupSkipped != 1 {
		t.Errorf("expected 1 dedup skipped for pre-processed message, got %d", stats.DedupSkipped)
	}
	if count != 0 {
		t.Errorf("expected 0 consumed, got %d", count)
	}
}
