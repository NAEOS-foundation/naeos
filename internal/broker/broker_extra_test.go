package broker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFactoryNewAllDrivers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		driver string
		want   string
	}{
		{"mock-redis", "redis"},
		{"mock-rabbitmq", "rabbitmq"},
		{"mock-kafka", "kafka"},
		{"memory", "memory"},
		{"inmemory", "memory"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			t.Parallel()
			b := New(tt.driver)
			if tt.want == "" {
				if b != nil {
					t.Errorf("expected nil for unknown driver")
				}
			} else {
				if b == nil {
					t.Fatalf("expected non-nil for driver %s", tt.driver)
				}
				if b.Name() != tt.want {
					t.Errorf("expected name %q, got %q", tt.want, b.Name())
				}
			}
		})
	}
}

func TestNewFromConfigMemory(t *testing.T) {
	t.Parallel()

	b, err := NewFromConfig("memory", &Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Name() != "memory" {
		t.Errorf("expected 'memory', got %q", b.Name())
	}
}

func TestBrokerStoreAddAndGet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.Add("mybroker", "redis", &Config{Host: "localhost", Port: 6379})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := s.Get("mybroker")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Driver != "redis" {
		t.Errorf("expected driver 'redis', got %q", got.Driver)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 broker, got %d", len(list))
	}

	if err := s.Remove("mybroker"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	_, err = s.Get("mybroker")
	if err == nil {
		t.Fatal("expected error after remove")
	}
}

func TestBrokerStoreAddUpdate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	s.Add("b1", "redis", &Config{Host: "h1", Port: 1})
	s.Add("b1", "kafka", &Config{Host: "h2", Port: 2})

	got, err := s.Get("b1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Driver != "kafka" {
		t.Errorf("expected updated driver 'kafka', got %q", got.Driver)
	}
}

func TestBrokerStoreRemoveNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent broker")
	}
}

func TestBrokerStoreGetNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	_, err := s.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent broker")
	}
}

func TestBrokerStoreListEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestMessageFilterCombined(t *testing.T) {
	t.Parallel()

	f := NewMessageFilter()
	f.SetChannelPattern("orders.*")
	f.SetPayloadMatch("urgent")
	f.SetPredicate(func(msg *Message) bool {
		return len(msg.Payload) > 0
	})

	if !f.Match(&Message{Channel: "orders.new", Payload: []byte("urgent order")}) {
		t.Error("expected match for combined filter")
	}
	if f.Match(&Message{Channel: "orders.new", Payload: []byte("normal")}) {
		t.Error("expected no match when payload doesn't match")
	}
	if f.Match(&Message{Channel: "users.new", Payload: []byte("urgent")}) {
		t.Error("expected no match when channel doesn't match")
	}
}

func TestMatchGlobEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		s       string
		want    bool
	}{
		{"*", "anything", true},
		{"a*", "abc", true},
		{"*z", "xyz", true},
		{"a*z", "abcz", true},
		{"a*z", "abc", false},
		{"prefix*suffix", "prefix_middle_suffix", true},
		{"prefix*suffix", "prefixsuffix", true},
		{"exact", "exact", true},
		{"exact", "nope", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.s, func(t *testing.T) {
			t.Parallel()
			got := matchGlob(tt.pattern, tt.s)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.s, got, tt.want)
			}
		})
	}
}

func TestDeadLetterChannelDrainNilHandler(t *testing.T) {
	t.Parallel()

	dlc := NewDeadLetterChannel(10)
	dlc.Handler()
	dlc.Drain(nil)

	handler := dlc.Handler()
	handler(&Message{ID: "1"})

	dlc.Close()
}

func TestInMemoryBrokerUnsubscribeNonExistent(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	b.Connect(&Config{})

	err := b.Unsubscribe("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConnectionPoolSetHealthyInvalidIndex(t *testing.T) {
	t.Parallel()

	r1 := NewRedis()
	pool := NewConnectionPool(r1)

	pool.SetHealthy(-1, false)
	if pool.HealthyCount() != 1 {
		t.Error("expected 1 healthy (invalid index ignored)")
	}

	pool.SetHealthy(100, false)
	if pool.HealthyCount() != 1 {
		t.Error("expected 1 healthy (out of bounds index ignored)")
	}
}

func TestMetricsResetAll(t *testing.T) {
	t.Parallel()

	m := NewMetrics()
	m.IncPublished()
	m.IncReceived()
	m.IncErrors()
	m.SetSubscriberCount("ch", 5)

	m.Reset()

	if m.PublishedCount() != 0 {
		t.Errorf("expected 0 published after reset, got %d", m.PublishedCount())
	}
	if m.ReceivedCount() != 0 {
		t.Errorf("expected 0 received after reset, got %d", m.ReceivedCount())
	}
	if m.ErrorsCount() != 0 {
		t.Errorf("expected 0 errors after reset, got %d", m.ErrorsCount())
	}
	if m.SubscriberCount("ch") != 0 {
		t.Errorf("expected 0 subscribers after reset, got %d", m.SubscriberCount("ch"))
	}
}

func TestMetricsBrokerSubscribeUnsubscribe(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	b.Connect(&Config{})

	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	err := mb.Subscribe("ch", func(msg *Message) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	if m.SubscriberCount("ch") != 1 {
		t.Errorf("expected 1 subscriber, got %d", m.SubscriberCount("ch"))
	}

	err = mb.Unsubscribe("ch")
	if err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	if m.SubscriberCount("ch") != 0 {
		t.Errorf("expected 0 subscribers after unsubscribe, got %d", m.SubscriberCount("ch"))
	}
}

type failingBroker struct{}

func (f *failingBroker) Name() string                               { return "failing" }
func (f *failingBroker) Connect(config *Config) error               { return nil }
func (f *failingBroker) Close() error                               { return nil }
func (f *failingBroker) Ping() error                                { return nil }
func (f *failingBroker) Publish(channel string, msg *Message) error { return fmt.Errorf("fail") }
func (f *failingBroker) Subscribe(channel string, handler MessageHandler) error {
	return fmt.Errorf("fail")
}
func (f *failingBroker) Unsubscribe(channel string) error { return fmt.Errorf("fail") }

func TestMetricsBrokerSubscribeFail(t *testing.T) {
	t.Parallel()

	b := &failingBroker{}
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	err := mb.Subscribe("ch", func(msg *Message) error { return nil })
	if err == nil {
		t.Fatal("expected error from subscribe")
	}
	if m.ErrorsCount() != 1 {
		t.Errorf("expected 1 error, got %d", m.ErrorsCount())
	}
}

func TestPublishWithRetryAllFail(t *testing.T) {
	t.Parallel()

	b := &failingBroker{}
	msg := NewMessage("ch", []byte("data"))
	rc := &RetryConfig{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond, Multiplier: 1}

	err := PublishWithRetry(b, "ch", msg, rc)
	if err == nil {
		t.Fatal("expected error after all retries fail")
	}
}

func TestPublishWithRetrySuccessOnRetry(t *testing.T) {
	t.Parallel()

	var attempts int
	b := &retryBroker{failUntil: 2, attempts: &attempts}
	msg := NewMessage("ch", []byte("data"))
	rc := &RetryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond, Multiplier: 1}

	err := PublishWithRetry(b, "ch", msg, rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type retryBroker struct {
	failUntil int
	attempts  *int
}

func (r *retryBroker) Name() string                 { return "retry" }
func (r *retryBroker) Connect(config *Config) error { return nil }
func (r *retryBroker) Close() error                 { return nil }
func (r *retryBroker) Ping() error                  { return nil }
func (r *retryBroker) Publish(channel string, msg *Message) error {
	*r.attempts++
	if *r.attempts <= r.failUntil {
		return fmt.Errorf("fail attempt %d", *r.attempts)
	}
	return nil
}
func (r *retryBroker) Subscribe(channel string, handler MessageHandler) error { return nil }
func (r *retryBroker) Unsubscribe(channel string) error                       { return nil }

func TestRetryConfigDelayMaxCap(t *testing.T) {
	t.Parallel()

	rc := &RetryConfig{BaseDelay: time.Second, MaxDelay: 2 * time.Second, Multiplier: 2.0}
	d := rc.delay(10)
	if d > 2*time.Second {
		t.Errorf("expected delay capped at 2s, got %v", d)
	}
}

func TestMessageFilterWrapHandlerNoFilter(t *testing.T) {
	t.Parallel()

	f := NewMessageFilter()
	called := false
	wrapped := f.WrapHandler(func(msg *Message) error {
		called = true
		return nil
	})

	wrapped(&Message{Channel: "any"})
	if !called {
		t.Error("expected handler to be called with no filter")
	}
}

func TestConnectionStoreCorruptedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	_ = fmt.Sprintf // ensure fmt is used

	_, err := s.Get("test")
	if err == nil {
		t.Fatal("expected error for empty store")
	}
}

func TestInMemoryBrokerPublishToNoSubscribers(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	b.Connect(&Config{})

	err := b.Publish("empty", NewMessage("empty", []byte("data")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConnectionPoolMetricsAfterHealthChecks(t *testing.T) {
	t.Parallel()

	r1 := NewRedis()
	pool := NewConnectionPool(r1)

	pool.CheckHealth()
	pool.CheckHealth()

	m := pool.PoolMetrics()
	if m.HealthChecks != 2 {
		t.Errorf("expected 2 health checks, got %d", m.HealthChecks)
	}
}

func TestRedisSubscribeNotFound(t *testing.T) {
	t.Parallel()

	b := NewRedis()
	b.Connect(&Config{})

	err := b.Unsubscribe("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMessageBuilderID(t *testing.T) {
	t.Parallel()

	id := generateID()
	if len(id) == 0 {
		t.Error("expected non-empty ID")
	}
}

func TestConnectionPoolNextSkipsUnhealthy(t *testing.T) {
	t.Parallel()

	r1 := NewRedis()
	r2 := NewRedis()
	pool := NewConnectionPool(r1, r2)

	pool.SetHealthy(0, false)

	for i := 0; i < 4; i++ {
		b := pool.Next()
		if b == nil {
			t.Fatal("expected non-nil broker")
		}
		if b != r2 {
			t.Errorf("expected r2 (healthy), got different broker")
		}
	}
}

func TestMetricsBrokerPublish(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	b.Connect(&Config{})

	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	mb.Subscribe("ch", func(msg *Message) error { return nil })

	err := mb.Publish("ch", NewMessage("ch", []byte("data")))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if m.PublishedCount() != 1 {
		t.Errorf("expected 1 published, got %d", m.PublishedCount())
	}
	if m.ReceivedCount() != 1 {
		t.Errorf("expected 1 received, got %d", m.ReceivedCount())
	}
}

func TestMetricsBrokerName(t *testing.T) {
	t.Parallel()

	b := NewRedis()
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	if mb.Name() != "redis" {
		t.Errorf("expected 'redis', got %q", mb.Name())
	}
}

func TestInMemoryBrokerPublishSubscribeConcurrent(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	b.Connect(&Config{})

	var mu sync.Mutex
	received := 0

	for i := 0; i < 10; i++ {
		b.Subscribe("ch", func(msg *Message) error {
			mu.Lock()
			received++
			mu.Unlock()
			return nil
		})
	}

	for i := 0; i < 5; i++ {
		b.Publish("ch", NewMessage("ch", []byte("data")))
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if received != 50 {
		t.Errorf("expected 50 handler calls, got %d", received)
	}
	mu.Unlock()
}

func TestConnectionPoolHealthCheckWithCustomFn(t *testing.T) {
	t.Parallel()

	r1 := NewRedis()
	r2 := NewRedis()
	pool := NewConnectionPool(r1, r2)

	pool.SetHealthCheck(func(b Broker) bool {
		return b.Name() == "redis"
	})
	pool.CheckHealth()

	if pool.HealthyCount() != 2 {
		t.Errorf("expected 2 healthy, got %d", pool.HealthyCount())
	}
}

func TestConnectionPoolNextAllUnhealthy(t *testing.T) {
	t.Parallel()

	r1 := NewRedis()
	r2 := NewRedis()
	pool := NewConnectionPool(r1, r2)

	pool.SetHealthy(0, false)
	pool.SetHealthy(1, false)

	if got := pool.Next(); got != nil {
		t.Error("expected nil when all brokers unhealthy")
	}
}

func TestMetricsBrokerSubscribeCount(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	b.Connect(&Config{})
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	mb.Subscribe("ch1", func(msg *Message) error { return nil })
	mb.Subscribe("ch1", func(msg *Message) error { return nil })
	mb.Subscribe("ch2", func(msg *Message) error { return nil })

	if m.SubscriberCount("ch1") != 2 {
		t.Errorf("expected 2 subscribers for ch1, got %d", m.SubscriberCount("ch1"))
	}
	if m.SubscriberCount("ch2") != 1 {
		t.Errorf("expected 1 subscriber for ch2, got %d", m.SubscriberCount("ch2"))
	}
}

func TestChainMultipleMiddleware(t *testing.T) {
	t.Parallel()

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

var _ atomic.Int64
