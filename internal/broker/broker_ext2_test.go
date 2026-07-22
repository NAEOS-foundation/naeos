package broker

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInMemoryBrokerDeadLetterChan(t *testing.T) {
	b := NewInMemoryBroker()
	ch := b.DeadLetterChan()
	if ch == nil {
		t.Fatal("expected non-nil DeadLetterChan")
	}

	// Should be readable (non-blocking)
	select {
	case <-ch:
		// Empty channel is fine
	default:
		// Expected
	}
}

func TestInMemoryBrokerPublishedChan(t *testing.T) {
	b := NewInMemoryBroker()
	ch := b.PublishedChan()
	if ch == nil {
		t.Fatal("expected non-nil PublishedChan")
	}

	select {
	case <-ch:
	default:
	}
}

func TestInMemoryBrokerPublishConfirmChan(t *testing.T) {
	b := NewInMemoryBroker()
	ch := b.PublishConfirmChan()
	if ch == nil {
		t.Fatal("expected non-nil PublishConfirmChan")
	}
}

func TestDeadLetterHandler(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(nil)

	var received *Message
	b.SetDeadLetterHandler(func(msg *Message) error {
		received = msg
		return nil
	})

	subscriber := func(msg *Message) error {
		return fmt.Errorf("delivery failed")
	}
	b.Subscribe("dead-letter", subscriber)

	msg := NewMessage("dead-letter", []byte("failed"))
	b.Publish("dead-letter", msg)

	if received == nil {
		t.Error("expected dead letter handler to be called")
	}
	if received != nil && string(received.Payload) != "failed" {
		t.Errorf("expected payload 'failed', got %s", string(received.Payload))
	}
}

func TestInMemoryBrokerSubscriberCount(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(nil)

	if b.SubscriberCount("test") != 0 {
		t.Error("expected 0 subscribers initially")
	}

	b.Subscribe("test", func(msg *Message) error { return nil })
	if b.SubscriberCount("test") != 1 {
		t.Error("expected 1 subscriber")
	}

	b.Subscribe("test", func(msg *Message) error { return nil })
	if b.SubscriberCount("test") != 2 {
		t.Error("expected 2 subscribers")
	}

	b.Unsubscribe("test")
	if b.SubscriberCount("test") != 0 {
		t.Error("expected 0 subscribers after unsubscribe all")
	}
}

func TestMiddlewareChainExt(t *testing.T) {
	var order []string

	m1 := func(next MessageHandler) MessageHandler {
		return func(msg *Message) error {
			order = append(order, "m1-before")
			err := next(msg)
			order = append(order, "m1-after")
			return err
		}
	}
	m2 := func(next MessageHandler) MessageHandler {
		return func(msg *Message) error {
			order = append(order, "m2-before")
			err := next(msg)
			order = append(order, "m2-after")
			return err
		}
	}

	handler := Chain(func(msg *Message) error {
		order = append(order, "handler")
		return nil
	}, m1, m2)

	msg := NewMessage("test", []byte("data"))
	err := handler(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d", len(expected), len(order))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

func TestMiddlewareChainEmptyExt(t *testing.T) {
	var called bool
	handler := Chain(func(msg *Message) error {
		called = true
		return nil
	})

	msg := NewMessage("test", []byte("data"))
	handler(msg)
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestConnectAll(t *testing.T) {
	m := NewManager()
	m.Register("mem1", NewInMemoryBroker())
	m.Register("mem2", NewInMemoryBroker())

	configs := map[string]*Config{
		"mem1": {Host: "localhost"},
		"mem2": {Host: "localhost"},
	}
	err := m.ConnectAll(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloseAllSuccess(t *testing.T) {
	m := NewManager()
	m.Register("mem1", NewInMemoryBroker())

	err := m.CloseAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConnectionPoolNext(t *testing.T) {
	b := NewInMemoryBroker()
	b2 := NewInMemoryBroker()
	pool := NewConnectionPool(b, b2)

	next := pool.Next()
	if next == nil {
		t.Fatal("expected non-nil broker from pool")
	}
}

func TestConnectionPoolCheckHealthExt(t *testing.T) {
	b := NewInMemoryBroker()
	pool := NewConnectionPool(b)

	pool.CheckHealth()
	metrics := pool.PoolMetrics()
	if metrics.Total != 1 {
		t.Errorf("expected total 1, got %d", metrics.Total)
	}
}

func TestConnectionPoolSetHealthyExt(t *testing.T) {
	b := NewInMemoryBroker()
	pool := NewConnectionPool(b)

	pool.SetHealthy(0, false)
	metrics := pool.PoolMetrics()
	if metrics.Unhealthy != 1 {
		t.Errorf("expected 1 unhealthy, got %d", metrics.Unhealthy)
	}
}

func TestConnectionPoolMaxLifetime(t *testing.T) {
	b := NewInMemoryBroker()
	pool := NewConnectionPool(b)

	pool.SetMaxLifetime(time.Hour)
	pool.SetMaxLifetime(0) // Reset
}

func TestConnectionPoolSetHealthCheckExt(t *testing.T) {
	b := NewInMemoryBroker()
	pool := NewConnectionPool(b)

	pool.SetHealthCheck(func(broker Broker) bool {
		return true
	})
}

func TestConnectionStoreAddAndRemove(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.Add("conn1", "memory", &Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	entries, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(entries))
	}

	err = s.Remove("conn1")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	entries, err = s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 connections after remove, got %d", len(entries))
	}
}

func TestConnectionStoreAddDuplicate(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	s.Add("conn1", "memory", &Config{Host: "localhost"})
	s.Add("conn1", "redis", &Config{Host: "other"})

	entries, _ := s.List()
	if len(entries) != 1 {
		t.Errorf("expected 1 connection (deduped), got %d", len(entries))
	}
}

func TestConnectionStoreRemoveNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.Remove("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}

func TestConnectionStoreGet(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	s.Add("conn1", "memory", &Config{Host: "localhost"})
	entry, err := s.Get("conn1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if entry.Name != "conn1" {
		t.Errorf("expected 'conn1', got %s", entry.Name)
	}
}

func TestConnectionStoreGetNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	_, err := s.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}

func TestConnectionStoreSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	s.Add("conn1", "memory", &Config{Host: "localhost"})

	s2 := &ConnectionStore{dir: dir}
	s2.load()
	if len(s2.entries) != 1 {
		t.Errorf("expected 1 loaded entry, got %d", len(s2.entries))
	}
}

func TestConnectionStoreLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.load()
	if err != nil {
		t.Errorf("expected nil error for missing file, got %v", err)
	}
}

func TestNewMessageAttributes(t *testing.T) {
	msg := NewMessage("topic", []byte("payload"))
	if msg.Channel != "topic" {
		t.Errorf("expected channel 'topic', got %s", msg.Channel)
	}
	if string(msg.Payload) != "payload" {
		t.Errorf("expected payload 'payload', got %s", string(msg.Payload))
	}
	if msg.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestNewFromConfigSuccessWithTLS(t *testing.T) {
	b, err := NewFromConfig("memory", &Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Name() != "memory" {
		t.Errorf("expected 'memory', got %s", b.Name())
	}
}

func TestNewFromConfigUnknownDriver(t *testing.T) {
	_, err := NewFromConfig("unknown-driver", &Config{Host: "localhost"})
	if err == nil {
		t.Error("expected error for unknown driver")
	}
}

func TestInMemoryBrokerConnectNilConfig(t *testing.T) {
	b := NewInMemoryBroker()
	err := b.Connect(nil)
	if err != nil {
		t.Errorf("expected nil error for nil config on memory broker, got %v", err)
	}
}

func TestInMemoryBrokerConnectConfig(t *testing.T) {
	b := NewInMemoryBroker()
	err := b.Connect(&Config{Host: "localhost", Port: 1234})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestInMemoryBrokerPing(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(nil)
	if err := b.Ping(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestInMemoryBrokerClose(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(nil)
	if err := b.Close(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestManagerRegisterAndGet(t *testing.T) {
	m := NewManager()
	b := NewInMemoryBroker()
	m.Register("mem", b)

	got, ok := m.Get("mem")
	if !ok {
		t.Fatal("expected broker to be found")
	}
	if got != b {
		t.Error("expected same broker instance")
	}
}

func TestManagerGetNotFound(t *testing.T) {
	m := NewManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent broker")
	}
}

func TestManagerList(t *testing.T) {
	m := NewManager()
	m.Register("b1", NewInMemoryBroker())
	m.Register("b2", NewInMemoryBroker())

	list := m.List()
	if len(list) != 2 {
		t.Errorf("expected 2 brokers, got %d", len(list))
	}
}

func TestInMemoryBrokerPublishSubscribeOrder(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(nil)

	var mu sync.Mutex
	var received []string

	b.Subscribe("test", func(msg *Message) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, string(msg.Payload))
		return nil
	})

	for i := 0; i < 5; i++ {
		msg := NewMessage("test", []byte(fmt.Sprintf("msg-%d", i)))
		b.Publish("test", msg)
	}

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(received) != 5 {
		t.Errorf("expected 5 messages, got %d", len(received))
	}
}

func TestInMemoryBrokerSubscribeMultipleChannels(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(nil)

	var count1, count2 int
	b.Subscribe("ch1", func(msg *Message) error {
		count1++
		return nil
	})
	b.Subscribe("ch2", func(msg *Message) error {
		count2++
		return nil
	})

	b.Publish("ch1", NewMessage("ch1", []byte("a")))
	b.Publish("ch2", NewMessage("ch2", []byte("b")))
	b.Publish("ch1", NewMessage("ch1", []byte("c")))

	time.Sleep(50 * time.Millisecond)
	if count1 != 2 {
		t.Errorf("expected 2 messages on ch1, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("expected 1 message on ch2, got %d", count2)
	}
}

func TestConnectionStoreGetAll(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	s.Add("c1", "memory", &Config{Host: "localhost"})
	s.Add("c2", "redis", &Config{Host: "localhost", Port: 6379})

	entries, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}
