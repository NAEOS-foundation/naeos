package broker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	b := NewRedis()

	if b.Name() != "redis" {
		t.Errorf("expected name 'redis', got %s", b.Name())
	}

	err := b.Connect(&Config{Host: "localhost", Port: 6379})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = b.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var received *Message
	var wg sync.WaitGroup
	wg.Add(1)

	b.Subscribe("test", func(msg *Message) error {
		received = msg
		wg.Done()
		return nil
	})

	msg := NewMessage("test", []byte("hello"))
	b.Publish("test", msg)

	wg.Wait()

	if received == nil {
		t.Fatal("expected message to be received")
	}

	b.Unsubscribe("test")

	err = b.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRabbitMQ(t *testing.T) {
	b := NewRabbitMQ()

	if b.Name() != "rabbitmq" {
		t.Errorf("expected name 'rabbitmq', got %s", b.Name())
	}

	err := b.Connect(&Config{Host: "localhost", Port: 5672})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = b.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var received *Message
	var wg sync.WaitGroup
	wg.Add(1)

	b.Subscribe("test", func(msg *Message) error {
		received = msg
		wg.Done()
		return nil
	})

	msg := NewMessage("test", []byte("hello"))
	b.Publish("test", msg)

	wg.Wait()

	if received == nil {
		t.Fatal("expected message to be received")
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKafka(t *testing.T) {
	b := NewKafka()

	if b.Name() != "kafka" {
		t.Errorf("expected name 'kafka', got %s", b.Name())
	}

	err := b.Connect(&Config{Host: "localhost", Port: 9092})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = b.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var received *Message
	var wg sync.WaitGroup
	wg.Add(1)

	b.Subscribe("test", func(msg *Message) error {
		received = msg
		wg.Done()
		return nil
	})

	msg := NewMessage("test", []byte("hello"))
	b.Publish("test", msg)

	wg.Wait()

	if received == nil {
		t.Fatal("expected message to be received")
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotConnected(t *testing.T) {
	b := NewRedis()

	err := b.Publish("test", &Message{})
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = b.Subscribe("test", func(msg *Message) error { return nil })
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = b.Ping()
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestManager(t *testing.T) {
	m := NewManager()

	redis := NewRedis()
	rabbitmq := NewRabbitMQ()
	kafka := NewKafka()

	m.Register("redis", redis)
	m.Register("rabbitmq", rabbitmq)
	m.Register("kafka", kafka)

	got, ok := m.Get("redis")
	if !ok {
		t.Fatal("expected broker to be found")
	}
	if got.Name() != "redis" {
		t.Errorf("expected 'redis', got %s", got.Name())
	}

	names := m.List()
	if len(names) != 3 {
		t.Errorf("expected 3 brokers, got %d", len(names))
	}

	m.Remove("redis")
	_, ok = m.Get("redis")
	if ok {
		t.Error("expected broker to be removed")
	}
}

func TestManagerConnectAll(t *testing.T) {
	m := NewManager()

	redis := NewRedis()
	m.Register("redis", redis)

	configs := map[string]*Config{
		"redis": {Host: "localhost", Port: 6379},
	}

	err := m.ConnectAll(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerCloseAll(t *testing.T) {
	m := NewManager()

	redis := NewRedis()
	redis.Connect(&Config{})
	m.Register("redis", redis)

	err := m.CloseAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage("test", []byte("hello"))
	if msg == nil {
		t.Fatal("expected message to be created")
	}
	if msg.Channel != "test" {
		t.Errorf("expected channel 'test', got %s", msg.Channel)
	}
	if string(msg.Payload) != "hello" {
		t.Errorf("expected payload 'hello', got %s", string(msg.Payload))
	}
}

func TestMessageTimestamp(t *testing.T) {
	before := time.Now()
	msg := NewMessage("test", []byte("hello"))
	after := time.Now()

	if msg.Timestamp.Before(before) || msg.Timestamp.After(after) {
		t.Error("expected timestamp between before and after")
	}
}

func TestInMemoryBrokerBasic(t *testing.T) {
	b := NewInMemoryBroker()
	if b.Name() != "memory" {
		t.Errorf("expected name 'memory', got %s", b.Name())
	}

	err := b.Connect(&Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = b.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var received *Message
	b.Subscribe("test", func(msg *Message) error {
		received = msg
		return nil
	})

	msg := NewMessage("test", []byte("hello"))
	err = b.Publish("test", msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received == nil {
		t.Fatal("expected message to be received")
	}
	if string(received.Payload) != "hello" {
		t.Errorf("expected payload 'hello', got %s", string(received.Payload))
	}

	if b.SubscriberCount("test") != 1 {
		t.Errorf("expected 1 subscriber, got %d", b.SubscriberCount("test"))
	}

	b.Unsubscribe("test")
	if b.SubscriberCount("test") != 0 {
		t.Errorf("expected 0 subscribers after unsubscribe, got %d", b.SubscriberCount("test"))
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInMemoryBrokerMultipleSubscribers(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})

	count := 0
	var mu sync.Mutex

	b.Subscribe("ch", func(msg *Message) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})
	b.Subscribe("ch", func(msg *Message) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})

	msg := NewMessage("ch", []byte("data"))
	b.Publish("ch", msg)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if count != 2 {
		t.Errorf("expected 2 handler calls, got %d", count)
	}
	mu.Unlock()
}

func TestInMemoryBrokerDeadLetter(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})

	var dlqReceived *Message
	b.SetDeadLetterHandler(func(msg *Message) error {
		dlqReceived = msg
		return nil
	})

	b.Subscribe("ch", func(msg *Message) error {
		return fmt.Errorf("handler failed")
	})

	msg := NewMessage("ch", []byte("fail"))
	b.Publish("ch", msg)
	time.Sleep(10 * time.Millisecond)

	if dlqReceived == nil {
		t.Fatal("expected dead letter handler to be called")
	}
}

func TestInMemoryBrokerNotConnected(t *testing.T) {
	b := NewInMemoryBroker()
	err := b.Publish("ch", NewMessage("ch", []byte("x")))
	if err == nil {
		t.Error("expected error when not connected")
	}
	err = b.Subscribe("ch", func(msg *Message) error { return nil })
	if err == nil {
		t.Error("expected error when not connected")
	}
	err = b.Ping()
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestInMemoryBrokerPublishNilMessage(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})
	err := b.Publish("ch", nil)
	if err == nil {
		t.Error("expected error for nil message")
	}
}

func TestInMemoryBrokerChannels(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})

	b.Subscribe("ch", func(msg *Message) error { return nil })
	msg := NewMessage("ch", []byte("test"))
	b.Publish("ch", msg)

	select {
	case <-b.PublishConfirmChan():
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected publish confirm")
	}

	select {
	case published := <-b.PublishedChan():
		if string(published.Payload) != "test" {
			t.Errorf("expected payload 'test', got %s", string(published.Payload))
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected published message")
	}
}

func TestConnectionPoolBasic(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	pool := NewConnectionPool(r1, r2)

	if pool.Len() != 2 {
		t.Errorf("expected 2, got %d", pool.Len())
	}
	if pool.HealthyCount() != 2 {
		t.Errorf("expected 2 healthy, got %d", pool.HealthyCount())
	}
}

func TestConnectionPoolRoundRobin(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	r1.Connect(&Config{})
	r2.Connect(&Config{})
	pool := NewConnectionPool(r1, r2)

	w1 := pool.Next()
	w2 := pool.Next()
	if w1 == w2 {
		t.Error("expected different brokers from round-robin")
	}
}

func TestConnectionPoolEmpty(t *testing.T) {
	pool := NewConnectionPool()
	if pool.Next() != nil {
		t.Error("expected nil for empty pool")
	}
}

func TestConnectionPoolSetHealthy(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	pool := NewConnectionPool(r1, r2)

	pool.SetHealthy(0, false)
	if pool.HealthyCount() != 1 {
		t.Errorf("expected 1 healthy, got %d", pool.HealthyCount())
	}
}

func TestConnectionPoolCheckHealth(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	r1.Connect(&Config{})
	pool := NewConnectionPool(r1, r2)
	pool.CheckHealth()
}

func TestDeadLetterChannel(t *testing.T) {
	dlc := NewDeadLetterChannel(10)
	if dlc.Len() != 0 {
		t.Errorf("expected 0, got %d", dlc.Len())
	}

	handler := dlc.Handler()
	err := handler(&Message{ID: "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dlc.Len() != 1 {
		t.Errorf("expected 1 message, got %d", dlc.Len())
	}

	done := make(chan struct{})
	dlc.Drain(func(msg *Message) error {
		close(done)
		return nil
	})
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected drain handler to be called")
	}
	dlc.Close()
}

func TestDeadLetterChannelFull(t *testing.T) {
	dlc := NewDeadLetterChannel(1)
	handler := dlc.Handler()
	handler(&Message{ID: "1"})
	err := handler(&Message{ID: "2"})
	if err == nil {
		t.Error("expected error when dead letter queue full")
	}
}

func TestMetrics(t *testing.T) {
	m := NewMetrics()
	m.IncPublished()
	m.IncReceived()
	m.IncErrors()
	m.IncPublished()

	if m.PublishedCount() != 2 {
		t.Errorf("expected 2, got %d", m.PublishedCount())
	}
	if m.ReceivedCount() != 1 {
		t.Errorf("expected 1, got %d", m.ReceivedCount())
	}
	if m.ErrorsCount() != 1 {
		t.Errorf("expected 1, got %d", m.ErrorsCount())
	}

	m.SetSubscriberCount("ch1", 3)
	if m.SubscriberCount("ch1") != 3 {
		t.Errorf("expected 3, got %d", m.SubscriberCount("ch1"))
	}

	m.Reset()
	if m.PublishedCount() != 0 {
		t.Errorf("expected 0 after reset, got %d", m.PublishedCount())
	}
}

func TestMetricsBroker(t *testing.T) {
	b := NewRedis()
	b.Connect(&Config{})
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	if mb.Name() != "redis" {
		t.Errorf("expected 'redis', got %s", mb.Name())
	}

	mb.Subscribe("ch", func(msg *Message) error { return nil })
	mb.Publish("ch", NewMessage("ch", []byte("data")))
	time.Sleep(10 * time.Millisecond)

	if m.PublishedCount() != 1 {
		t.Errorf("expected 1 published, got %d", m.PublishedCount())
	}
}

func TestMiddlewareChain(t *testing.T) {
	var order []string
	handler := func(msg *Message) error {
		order = append(order, "handler")
		return nil
	}

	mw1 := func(next MessageHandler) MessageHandler {
		return func(msg *Message) error {
			order = append(order, "mw1")
			return next(msg)
		}
	}
	mw2 := func(next MessageHandler) MessageHandler {
		return func(msg *Message) error {
			order = append(order, "mw2")
			return next(msg)
		}
	}

	chained := Chain(handler, mw1, mw2)
	chained(&Message{})

	if len(order) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(order))
	}
	if order[0] != "mw1" || order[1] != "mw2" || order[2] != "handler" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestMessageFilterChannelPattern(t *testing.T) {
	f := NewMessageFilter()
	f.SetChannelPattern("orders.*")

	if !f.Match(&Message{Channel: "orders.new"}) {
		t.Error("expected match")
	}
	if f.Match(&Message{Channel: "users.created"}) {
		t.Error("expected no match")
	}
}

func TestMessageFilterPayloadMatch(t *testing.T) {
	f := NewMessageFilter()
	f.SetPayloadMatch("error")

	if !f.Match(&Message{Payload: []byte("something error occurred")}) {
		t.Error("expected match")
	}
	if f.Match(&Message{Payload: []byte("all good")}) {
		t.Error("expected no match")
	}
}

func TestMessageFilterPredicate(t *testing.T) {
	f := NewMessageFilter()
	f.SetPredicate(func(msg *Message) bool {
		return string(msg.Payload) == "important"
	})

	if !f.Match(&Message{Payload: []byte("important")}) {
		t.Error("expected match")
	}
	if f.Match(&Message{Payload: []byte("spam")}) {
		t.Error("expected no match")
	}
}

func TestMessageFilterNilMessage(t *testing.T) {
	f := NewMessageFilter()
	if f.Match(nil) {
		t.Error("expected false for nil message")
	}
}

func TestMessageFilterWrapHandler(t *testing.T) {
	f := NewMessageFilter()

	var matched bool
	wrapped := f.WrapHandler(func(msg *Message) error {
		matched = true
		return nil
	})

	wrapped(&Message{Payload: []byte("any")})
	if !matched {
		t.Error("expected handler to be called when no filter set")
	}

	f.SetPayloadMatch("valid")
	var validCalled bool
	wrappedValid := f.WrapHandler(func(msg *Message) error {
		validCalled = true
		return nil
	})

	wrappedValid(&Message{Payload: []byte("valid data")})
	if !validCalled {
		t.Error("expected handler to be called for matching message")
	}

	var invalidCalled bool
	wrappedInvalid := f.WrapHandler(func(msg *Message) error {
		invalidCalled = true
		return nil
	})

	wrappedInvalid(&Message{Payload: []byte("bad")})
	if invalidCalled {
		t.Error("expected handler NOT to be called for non-matching message")
	}
}

func TestMatchGlob(t *testing.T) {
	if !matchGlob("orders.*", "orders.new") {
		t.Error("expected match")
	}
	if matchGlob("orders.*", "users.new") {
		t.Error("expected no match")
	}
	if !matchGlob("exact", "exact") {
		t.Error("expected exact match")
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	rc := DefaultRetryConfig()
	if rc.MaxAttempts != 3 {
		t.Errorf("expected 3, got %d", rc.MaxAttempts)
	}
	if rc.BaseDelay != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", rc.BaseDelay)
	}
}

func TestRetryConfigDelay(t *testing.T) {
	rc := DefaultRetryConfig()
	d0 := rc.delay(0)
	d1 := rc.delay(1)
	if d1 <= d0 {
		t.Error("expected increasing delay")
	}
}

func TestPublishWithRetryNilConfig(t *testing.T) {
	b := NewRedis()
	b.Connect(&Config{})
	msg := NewMessage("ch", []byte("data"))

	handler := func(msg *Message) error { return nil }
	b.Subscribe("ch", handler)

	err := PublishWithRetry(b, "ch", msg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerListEmpty(t *testing.T) {
	m := NewManager()
	names := m.List()
	if len(names) != 0 {
		t.Errorf("expected empty list, got %d", len(names))
	}
}

func TestManagerGetNonExistent(t *testing.T) {
	m := NewManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected false for non-existent broker")
	}
}

func TestConnectionPoolSetMaxLifetime(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	pool := NewConnectionPool(r1, r2)

	pool.SetMaxLifetime(0)
	if pool.HealthyCount() != 2 {
		t.Errorf("expected 2 healthy with no lifetime limit, got %d", pool.HealthyCount())
	}

	pool.SetMaxLifetime(1 * time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	if pool.HealthyCount() != 0 {
		t.Errorf("expected 0 healthy with expired lifetime, got %d", pool.HealthyCount())
	}
}

func TestConnectionPoolStartStopHealthCheck(t *testing.T) {
	r1 := NewRedis()
	pool := NewConnectionPool(r1)

	pool.StartHealthCheck(10 * time.Millisecond)
	defer pool.StopHealthCheck()

	time.Sleep(50 * time.Millisecond)

	pool.StopHealthCheck()
	pool.StopHealthCheck()
}

func TestConnectionPoolMetrics(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	pool := NewConnectionPool(r1, r2)

	m := pool.PoolMetrics()
	if m.Total != 2 {
		t.Errorf("expected total 2, got %d", m.Total)
	}
	if m.Healthy != 2 {
		t.Errorf("expected healthy 2, got %d", m.Healthy)
	}

	pool.SetHealthy(0, false)
	m = pool.PoolMetrics()
	if m.Healthy != 1 {
		t.Errorf("expected healthy 1, got %d", m.Healthy)
	}
	if m.Unhealthy != 1 {
		t.Errorf("expected unhealthy 1, got %d", m.Unhealthy)
	}

	pool.Next()
	pool.Next()
	m = pool.PoolMetrics()
	if m.NextCalls < 2 {
		t.Errorf("expected at least 2 NextCalls, got %d", m.NextCalls)
	}
}

func TestConnectionPoolAllUnhealthy(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	r1.Connect(&Config{})
	r2.Connect(&Config{})
	pool := NewConnectionPool(r1, r2)

	pool.SetHealthy(0, false)
	pool.SetHealthy(1, false)
	if pool.HealthyCount() != 0 {
		t.Errorf("expected 0 healthy, got %d", pool.HealthyCount())
	}
	if got := pool.Next(); got != nil {
		t.Error("expected nil when all brokers unhealthy")
	}
}

func TestConnectionPoolConcurrentNext(t *testing.T) {
	brokers := make([]Broker, 5)
	for i := range brokers {
		brokers[i] = NewRedis()
	}
	pool := NewConnectionPool(brokers...)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if b := pool.Next(); b == nil {
					t.Error("unexpected nil from Next()")
				}
			}
		}()
	}
	wg.Wait()
	metrics := pool.PoolMetrics()
	if metrics.NextCalls <= 0 {
		t.Error("expected non-zero NextCalls")
	}
}

func TestConnectionPoolNextRoundRobinOrder(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	r3 := NewRedis()
	pool := NewConnectionPool(r1, r2, r3)

	seen := make(map[Broker]int)
	for i := 0; i < 6; i++ {
		b := pool.Next()
		seen[b]++
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 unique brokers, got %d", len(seen))
	}
	for _, count := range seen {
		if count != 2 {
			t.Errorf("expected each broker 2 times, got %d", count)
		}
	}
}

func TestConnectionPoolHealthCheckTracksMaxLifetime(t *testing.T) {
	b := &mockBroker{name: "test"}
	pool := NewConnectionPool(b)
	pool.SetMaxLifetime(0)

	pool.SetHealthCheck(func(b Broker) bool { return true })
	pool.CheckHealth()
	if pool.HealthyCount() != 1 {
		t.Errorf("expected 1 healthy after check, got %d", pool.HealthyCount())
	}

	pool.SetMaxLifetime(1 * time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	pool.CheckHealth()
	if pool.HealthyCount() != 0 {
		t.Errorf("expected 0 healthy after lifetime expiry, got %d", pool.HealthyCount())
	}
}

type mockBroker struct {
	name      string
	connected atomic.Bool
}

func (m *mockBroker) Name() string                                           { return m.name }
func (m *mockBroker) Connect(config *Config) error                           { m.connected.Store(true); return nil }
func (m *mockBroker) Close() error                                           { m.connected.Store(false); return nil }
func (m *mockBroker) Ping() error                                            { return nil }
func (m *mockBroker) Publish(channel string, msg *Message) error             { return nil }
func (m *mockBroker) Subscribe(channel string, handler MessageHandler) error { return nil }
func (m *mockBroker) Unsubscribe(channel string) error                       { return nil }

func BenchmarkConnectionPoolNext(b *testing.B) {
	brokers := make([]Broker, 10)
	for i := range brokers {
		brokers[i] = NewRedis()
	}
	pool := NewConnectionPool(brokers...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Next()
	}
}

func BenchmarkConnectionPoolCheckHealth(b *testing.B) {
	brokers := make([]Broker, 10)
	for i := range brokers {
		brokers[i] = NewRedis()
	}
	pool := NewConnectionPool(brokers...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.CheckHealth()
	}
}
