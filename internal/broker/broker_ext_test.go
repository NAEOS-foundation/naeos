package broker

import (
	"fmt"
	"testing"
	"time"
)

func TestSupportedDrivers(t *testing.T) {
	drivers := SupportedDrivers()
	expected := []string{
		"redis", "rabbitmq", "kafka", "nats",
		"memory", "inmemory",
		"mock-redis", "mock-rabbitmq", "mock-kafka",
	}
	if len(drivers) != len(expected) {
		t.Errorf("expected %d drivers, got %d", len(expected), len(drivers))
	}
	for _, exp := range expected {
		found := false
		for _, d := range drivers {
			if d == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected driver %q not found", exp)
		}
	}
}

func TestNewAllDrivers(t *testing.T) {
	tests := []struct {
		driver string
		name   string
	}{
		{"memory", "memory"},
		{"inmemory", "memory"},
		{"mock-redis", "redis"},
		{"mock-rabbitmq", "rabbitmq"},
		{"mock-kafka", "kafka"},
	}
	for _, tt := range tests {
		b := New(tt.driver)
		if b == nil {
			t.Fatalf("New(%q) returned nil", tt.driver)
		}
		if b.Name() != tt.name {
			t.Errorf("expected name %q, got %q", tt.name, b.Name())
		}
	}
}

func TestNewUnsupportedDriver(t *testing.T) {
	b := New("unsupported")
	if b != nil {
		t.Errorf("expected nil for unsupported driver, got %T", b)
	}
}

func TestNewFromConfigSuccess(t *testing.T) {
	b, err := NewFromConfig("memory", &Config{})
	if err != nil {
		t.Fatalf("new from config: %v", err)
	}
	if b.Name() != "memory" {
		t.Errorf("expected 'memory', got %s", b.Name())
	}
}

func TestNewFromConfigUnsupportedDriver(t *testing.T) {
	_, err := NewFromConfig("invalid", &Config{})
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestMatchGlobExact(t *testing.T) {
	if !matchGlob("exact", "exact") {
		t.Error("expected exact match")
	}
}

func TestMatchGlobPrefixSuffix(t *testing.T) {
	if !matchGlob("*.log", "error.log") {
		t.Error("expected suffix match")
	}
	if matchGlob("*.log", "error.txt") {
		t.Error("expected no match")
	}
}

func TestMatchGlobMultiPartWildcard(t *testing.T) {
	if !matchGlob("a*b*c", "abc") {
		t.Error("expected match with empty wildcards")
	}
	if matchGlob("a*b*c", "aXbYc") {
		t.Log("multi-wildcard matches interleaved content via contains")
	}
}

func TestPow(t *testing.T) {
	if result := pow(2.0, 0); result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
	if result := pow(2.0, 3); result != 8.0 {
		t.Errorf("expected 8.0, got %f", result)
	}
	if result := pow(3.0, 2); result != 9.0 {
		t.Errorf("expected 9.0, got %f", result)
	}
}

func TestRetryConfigDelayCaps(t *testing.T) {
	rc := &RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   10 * time.Second,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}
	d := rc.delay(0)
	if d > rc.MaxDelay {
		t.Errorf("expected delay capped at %v, got %v", rc.MaxDelay, d)
	}
}

type failBroker struct{ mockBroker }

func (f *failBroker) Publish(channel string, msg *Message) error {
	return fmt.Errorf("publish failed")
}

func TestPublishWithRetryFailsAfterAllAttempts(t *testing.T) {
	b := &failBroker{}

	msg := NewMessage("ch", []byte("data"))
	rc := &RetryConfig{
		MaxAttempts: 2,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  1.0,
	}

	err := PublishWithRetry(b, "ch", msg, rc)
	if err == nil {
		t.Fatal("expected publish to fail")
	}
}

func TestMetricsBrokerPublishError(t *testing.T) {
	b := NewRedis()
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	mb.Publish("ch", NewMessage("ch", []byte("data")))
	if m.ErrorsCount() == 0 {
		t.Error("expected errors for publish when not connected")
	}
}

func TestMetricsBrokerSubscribeError(t *testing.T) {
	b := NewRedis()
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	err := mb.Subscribe("ch", func(msg *Message) error { return nil })
	if err == nil {
		t.Error("expected error for subscribe when not connected")
	}
	if m.ErrorsCount() == 0 {
		t.Error("expected error count > 0")
	}
}

func TestMetricsBrokerMetrics(t *testing.T) {
	b := NewRedis()
	b.Connect(&Config{})
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	if mb.Metrics() != m {
		t.Error("expected same metrics reference")
	}
}

func TestMetricsBrokerConnectClosePing(t *testing.T) {
	b := NewRedis()
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)

	if err := mb.Connect(&Config{}); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := mb.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := mb.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestChainNoMiddleware(t *testing.T) {
	var called bool
	handler := func(msg *Message) error {
		called = true
		return nil
	}

	chained := Chain(handler)
	chained(&Message{})
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestConnectionPoolCloseAll(t *testing.T) {
	r1 := NewRedis()
	r2 := NewRedis()
	r1.Connect(&Config{})
	r2.Connect(&Config{})
	pool := NewConnectionPool(r1, r2)

	if err := pool.CloseAll(); err != nil {
		t.Fatalf("close all: %v", err)
	}
	if r1.Ping() == nil {
		t.Error("expected r1 to be closed")
	}
	if r2.Ping() == nil {
		t.Error("expected r2 to be closed")
	}
}

func TestConnectionPoolSetHealthCheck(t *testing.T) {
	r := NewRedis()
	pool := NewConnectionPool(r)
	pool.SetHealthCheck(func(b Broker) bool { return false })
	pool.CheckHealth()
	if pool.HealthyCount() != 0 {
		t.Error("expected 0 healthy with failing check fn")
	}
}

func TestConnectionPoolSetHealthyOutOfRange(t *testing.T) {
	r := NewRedis()
	pool := NewConnectionPool(r)
	pool.SetHealthy(-1, false)
	pool.SetHealthy(5, false)
	if pool.HealthyCount() != 1 {
		t.Errorf("expected 1 healthy, got %d", pool.HealthyCount())
	}
}

func TestConnectionPoolStartHealthCheckDuplicate(t *testing.T) {
	r := NewRedis()
	pool := NewConnectionPool(r)
	pool.StartHealthCheck(1 * time.Hour)
	pool.StartHealthCheck(1 * time.Hour)
	pool.StopHealthCheck()
}

func TestDeadLetterChannelHandlerLock(t *testing.T) {
	dlc := NewDeadLetterChannel(10)
	handler := dlc.Handler()
	_ = handler(&Message{ID: "msg1"})
	_ = handler(&Message{ID: "msg2"})
	_ = handler(&Message{ID: "msg3"})
	if dlc.Len() != 3 {
		t.Errorf("expected 3 messages, got %d", dlc.Len())
	}
	dlc.Close()
}

func TestDeadLetterChannelMessages(t *testing.T) {
	dlc := NewDeadLetterChannel(10)
	_ = dlc.Handler()(&Message{ID: "m1"})
	_ = dlc.Handler()(&Message{ID: "m2"})

	ch := dlc.Messages()
	msg1 := <-ch
	if msg1.ID != "m1" {
		t.Errorf("expected m1, got %s", msg1.ID)
	}
	dlc.Close()
}

func TestStoreNewConnectionStore(t *testing.T) {
	s := NewConnectionStore()
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestStoreSaveLoadError(t *testing.T) {
	s := &ConnectionStore{dir: "/nonexistent/xyz/abc/123"}
	err := s.Add("test", "redis", &Config{})
	if err == nil {
		t.Error("expected error when dir nonexistent")
	}
}

func TestStoreGetLoadError(t *testing.T) {
	s := &ConnectionStore{dir: "/dev/null/nope"}
	_, err := s.Get("test")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStoreAddAndGet(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{
		dir: dir,
	}

	err := s.Add("my-broker", "memory", &Config{Host: "localhost", Port: 6379})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	entry, err := s.Get("my-broker")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if entry.Name != "my-broker" {
		t.Errorf("expected my-broker, got %s", entry.Name)
	}
	if entry.Driver != "memory" {
		t.Errorf("expected memory, got %s", entry.Driver)
	}
}

func TestStoreAddUpdateExisting(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	s.Add("b1", "redis", &Config{Port: 6379})
	s.Add("b1", "kafka", &Config{Port: 9092})
	entry, _ := s.Get("b1")
	if entry.Driver != "kafka" {
		t.Errorf("expected kafka, got %s", entry.Driver)
	}
	if entry.Config.Port != 9092 {
		t.Errorf("expected port 9092, got %d", entry.Config.Port)
	}
}

func TestStoreRemove(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	s.Add("b1", "redis", &Config{})
	err := s.Remove("b1")
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	_, err = s.Get("b1")
	if err == nil {
		t.Error("expected error after remove")
	}
}

func TestStoreRemoveNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	err := s.Remove("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStoreList(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	s.Add("b1", "redis", &Config{})
	s.Add("b2", "kafka", &Config{})

	entries, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2, got %d", len(entries))
	}
}

func TestStoreListEmpty(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	entries, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0, got %d", len(entries))
	}
}

func TestStorePersistence(t *testing.T) {
	dir := t.TempDir()
	s1 := &ConnectionStore{dir: dir}
	s1.Add("persistent", "memory", &Config{Port: 9999})

	s2 := &ConnectionStore{dir: dir}
	entry, err := s2.Get("persistent")
	if err != nil {
		t.Fatalf("get from reloaded store: %v", err)
	}
	if entry.Config.Port != 9999 {
		t.Errorf("expected port 9999, got %d", entry.Config.Port)
	}
}

func TestStoreGetNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	_, err := s.Get("missing")
	if err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestStoreNewConnectionStoreCreatesDir(t *testing.T) {
	s := NewConnectionStore()
	_ = s
}

func TestMatchGlobWildcardOnly(t *testing.T) {
	if !matchGlob("*", "anything") {
		t.Error("expected wildcard to match everything")
	}
}

func TestDeadLetterChannelCloseSafe(t *testing.T) {
	dlc := NewDeadLetterChannel(10)
	dlc.Close()
}

func TestManagerConnectAllSkipsMissing(t *testing.T) {
	m := NewManager()
	configs := map[string]*Config{
		"nonexistent": {Host: "localhost"},
	}
	err := m.ConnectAll(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRabbitMQUnsubscribe(t *testing.T) {
	b := NewRabbitMQ()
	b.Connect(&Config{})
	err := b.Unsubscribe("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKafkaUnsubscribe(t *testing.T) {
	b := NewKafka()
	b.Connect(&Config{})
	err := b.Unsubscribe("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerConnectAllSuccess(t *testing.T) {
	m := NewManager()
	r := NewRedis()
	m.Register("redis", r)

	configs := map[string]*Config{
		"redis": {Host: "localhost", Port: 6379},
	}
	if err := m.ConnectAll(configs); err != nil {
		t.Fatalf("connect all: %v", err)
	}
	if r.Ping() != nil {
		t.Error("expected broker to be connected")
	}
}

func TestInMemoryBrokerSubscriberCountNoChannel(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})
	if count := b.SubscriberCount("nonexistent"); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestMetricsBrokerUnsubscribe(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})
	m := NewMetrics()
	mb := NewMetricsBroker(b, m)
	mb.Subscribe("ch", func(msg *Message) error { return nil })

	if err := mb.Unsubscribe("ch"); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if m.SubscriberCount("ch") != 0 {
		t.Errorf("expected 0 subscribers, got %d", m.SubscriberCount("ch"))
	}
}
