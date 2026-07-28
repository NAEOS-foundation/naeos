package broker

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestInMemoryBrokerConcurrentPublishSubscribe(t *testing.T) {
	b := NewInMemoryBroker()
	if err := b.Connect(&Config{}); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer b.Close()

	var received atomic.Int64
	var wg sync.WaitGroup
	numMessages := 100
	numSubscribers := 5

	for i := 0; i < numSubscribers; i++ {
		if err := b.Subscribe("test", func(msg *Message) error {
			received.Add(1)
			return nil
		}); err != nil {
			t.Fatalf("Subscribe: %v", err)
		}
	}

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			msg := NewMessage("test", []byte("hello"))
			if err := b.Publish("test", msg); err != nil {
				t.Errorf("Publish: %v", err)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	expected := int64(numMessages * numSubscribers)
	if got := received.Load(); got != expected {
		t.Errorf("expected %d received, got %d", expected, got)
	}
}

func TestInMemoryBrokerConcurrentSubscribeUnsubscribe(t *testing.T) {
	b := NewInMemoryBroker()
	if err := b.Connect(&Config{}); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer b.Close()

	var wg sync.WaitGroup
	numOps := 50

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			handler := func(msg *Message) error { return nil }
			ch := "chan"
			_ = b.Subscribe(ch, handler)
			_ = b.Unsubscribe(ch)
		}(i)
	}

	wg.Wait()
	if count := b.SubscriberCount("chan"); count != 0 {
		t.Errorf("expected 0 subscribers, got %d", count)
	}
}

func TestConnectionPoolConcurrentNextWithMetrics(t *testing.T) {
	brokers := make([]Broker, 10)
	for i := range brokers {
		b := NewInMemoryBroker()
		_ = b.Connect(&Config{})
		brokers[i] = b
	}
	pool := NewConnectionPool(brokers...)

	var wg sync.WaitGroup
	var selected atomic.Int64
	numGoroutines := 20
	iterations := 50

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				b := pool.Next()
				if b != nil {
					selected.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * iterations)
	if got := selected.Load(); got != expected {
		t.Errorf("expected %d selections, got %d", expected, got)
	}

	metrics := pool.PoolMetrics()
	if metrics.NextCalls <= 0 {
		t.Error("expected non-zero NextCalls")
	}
}

func TestConnectionPoolConcurrentHealthCheck(t *testing.T) {
	brokers := make([]Broker, 5)
	for i := range brokers {
		b := NewInMemoryBroker()
		_ = b.Connect(&Config{})
		brokers[i] = b
	}
	pool := NewConnectionPool(brokers...)

	var wg sync.WaitGroup
	numGoroutines := 10

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				pool.CheckHealth()
				pool.HealthyCount()
				pool.Len()
				pool.PoolMetrics()
			}
		}()
	}

	wg.Wait()
}

func TestManagerConcurrentRegisterGetRemove(t *testing.T) {
	m := NewManager()
	var wg sync.WaitGroup
	numOps := 50

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := "broker"
			m.Register(name, NewRedis())
			_, _ = m.Get(name)
			m.List()
			m.Remove(name)
		}(i)
	}

	wg.Wait()
}

func TestMetricsConcurrentIncrement(t *testing.T) {
	metrics := NewMetrics()
	var wg sync.WaitGroup
	numGoroutines := 20
	opsPerGoroutine := 100

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				metrics.IncPublished()
				metrics.IncReceived()
				metrics.IncErrors()
				_ = metrics.PublishedCount()
				_ = metrics.ReceivedCount()
				_ = metrics.ErrorsCount()
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * opsPerGoroutine)
	if got := metrics.PublishedCount(); got != expected {
		t.Errorf("expected %d published, got %d", expected, got)
	}
	if got := metrics.ReceivedCount(); got != expected {
		t.Errorf("expected %d received, got %d", expected, got)
	}
	if got := metrics.ErrorsCount(); got != expected {
		t.Errorf("expected %d errors, got %d", expected, got)
	}
}

func TestMetricsConcurrentSubscriberCount(t *testing.T) {
	metrics := NewMetrics()
	var wg sync.WaitGroup
	numChannels := 20

	for i := 0; i < numChannels; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ch := "channel"
			metrics.SetSubscriberCount(ch, int64(n))
			_ = metrics.SubscriberCount(ch)
		}(i)
	}

	wg.Wait()
}

func TestDeadLetterChannelConcurrent(t *testing.T) {
	dlc := NewDeadLetterChannel(1000)
	var wg sync.WaitGroup
	handler := dlc.Handler()

	numMessages := 50
	numSenders := 10

	for s := 0; s < numSenders; s++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numMessages; i++ {
				msg := NewMessage("test", []byte("payload"))
				_ = handler(msg)
			}
		}()
	}

	wg.Wait()

	expected := numMessages * numSenders
	if got := dlc.Len(); got != expected {
		t.Errorf("expected %d messages, got %d", expected, got)
	}
}

func TestConnectionPoolConcurrentSetHealthy(t *testing.T) {
	brokers := make([]Broker, 5)
	for i := range brokers {
		b := NewInMemoryBroker()
		_ = b.Connect(&Config{})
		brokers[i] = b
	}
	pool := NewConnectionPool(brokers...)

	var wg sync.WaitGroup
	numGoroutines := 10

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				pool.SetHealthy(i%5, i%2 == 0)
				pool.Next()
				pool.HealthyCount()
			}
		}()
	}

	wg.Wait()
}

func TestMetricsBrokerConcurrent(t *testing.T) {
	inner := NewInMemoryBroker()
	if err := inner.Connect(&Config{}); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer inner.Close()

	metrics := NewMetrics()
	mb := NewMetricsBroker(inner, metrics)

	var wg sync.WaitGroup
	numOps := 50

	for i := 0; i < numOps; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = mb.Subscribe("ch", func(msg *Message) error { return nil })
		}()
		go func() {
			defer wg.Done()
			_ = mb.Publish("ch", NewMessage("ch", []byte("data")))
		}()
		go func() {
			defer wg.Done()
			_ = mb.Unsubscribe("ch")
		}()
	}

	wg.Wait()
	_ = mb.Name()
	_ = mb.Ping()
}
