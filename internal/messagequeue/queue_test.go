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

	q.Subscribe(func(msg *Message) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	q.Publish(NewMessage("test", "1"))
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
