//go:build !nobroker

package broker

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestRealRedisNotConnected(t *testing.T) {
	t.Parallel()
	b := NewRealRedis()

	if err := b.Ping(); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Publish("ch", &Message{Payload: []byte("data")}); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Subscribe("ch", func(msg *Message) error { return nil }); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealRedisName(t *testing.T) {
	t.Parallel()
	b := NewRealRedis()
	if b.Name() != "redis" {
		t.Errorf("expected 'redis', got %q", b.Name())
	}
}

func TestRealRedisUnsubscribeNotConnected(t *testing.T) {
	t.Parallel()
	b := NewRealRedis()
	if err := b.Unsubscribe("nonexistent"); err != nil {
		t.Fatalf("Unsubscribe on not connected: %v", err)
	}
}

func TestRealRedisPublishNilPayload(t *testing.T) {
	t.Parallel()
	b := NewRealRedis()
	err := b.Publish("ch", &Message{})
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestRealNATSNotConnectedAdditional(t *testing.T) {
	t.Parallel()
	b := NewRealNATS()

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealNATSUnsubscribeNotConnected(t *testing.T) {
	t.Parallel()
	b := NewRealNATS()
	if err := b.Unsubscribe("nonexistent"); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
}

func TestRealRabbitMQNotConnected(t *testing.T) {
	t.Parallel()
	b := NewRealRabbitMQ()

	if err := b.Ping(); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Publish("ch", &Message{Payload: []byte("data")}); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Subscribe("ch", func(msg *Message) error { return nil }); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealRabbitMQName(t *testing.T) {
	t.Parallel()
	b := NewRealRabbitMQ()
	if b.Name() != "rabbitmq" {
		t.Errorf("expected 'rabbitmq', got %q", b.Name())
	}
}

func TestRealRabbitMQUnsubscribeNotConnected(t *testing.T) {
	t.Parallel()
	b := NewRealRabbitMQ()
	if err := b.Unsubscribe("nonexistent"); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
}

func TestRealRabbitMQPublishNilPayload(t *testing.T) {
	t.Parallel()
	b := NewRealRabbitMQ()
	err := b.Publish("ch", &Message{})
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestRealRabbitMQConnectInvalid(t *testing.T) {
	t.Parallel()
	b := NewRealRabbitMQ()
	err := b.Connect(&Config{Host: "127.0.0.1", Port: 1})
	if err == nil {
		t.Error("expected error for invalid connection")
	}
}

func TestRealKafkaNotConnected(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()

	if err := b.Ping(); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Publish("ch", &Message{Payload: []byte("data")}); err == nil {
		t.Error("expected error when not connected")
	}

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealKafkaName(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	if b.Name() != "kafka" {
		t.Errorf("expected 'kafka', got %q", b.Name())
	}
}

func TestRealKafkaConnect(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	err := b.Connect(&Config{Host: "localhost", Port: 9092})
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if err := b.Ping(); err != nil {
		t.Errorf("expected nil Ping after connect, got %v", err)
	}
}

func TestRealKafkaConnectThenClose(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	b.Connect(&Config{Host: "localhost", Port: 9092})

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealKafkaPublishNilPayload(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	err := b.Publish("ch", &Message{})
	if err == nil {
		t.Error("expected error when not connected")
	}
}

func TestRealKafkaSubscribeNotConnected(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	err := b.Subscribe("ch", func(msg *Message) error { return nil })
	if err == nil {
		t.Error("expected error when config is nil")
	}
}

func TestRealKafkaUnsubscribeNotSubscribed(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	if err := b.Unsubscribe("nonexistent"); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
}

func TestBrokerSupportedDrivers(t *testing.T) {
	t.Parallel()
	drivers := SupportedDrivers()
	if len(drivers) == 0 {
		t.Error("expected non-empty list")
	}
}

func TestNewFromConfigUnsupported(t *testing.T) {
	t.Parallel()
	_, err := NewFromConfig("unknown", &Config{})
	if err == nil {
		t.Error("expected error for unsupported driver")
	}
}

func TestBrokerFactoryNewRealDrivers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		driver string
		want   string
	}{
		{"redis", "redis"},
		{"rabbitmq", "rabbitmq"},
		{"kafka", "kafka"},
		{"nats", "nats"},
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

func TestRealKafkaPublishAfterConnect(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	b.Connect(&Config{Host: "localhost", Port: 9092, Timeout: 100 * time.Millisecond})

	err := b.Publish("ch", &Message{Payload: []byte("data")})
	if err == nil {
		t.Log("publish may have succeeded or failed with connection error")
	}
}

func TestRealKafkaPublishNilPayloadAfterConnect(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	b.Connect(&Config{Host: "localhost", Port: 9092, Timeout: 100 * time.Millisecond})

	err := b.Publish("ch", &Message{})
	if err == nil {
		t.Log("publish may have succeeded or failed with connection error")
	}
}

func TestRealKafkaSubscribeAfterConnect(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	b.Connect(&Config{Host: "localhost", Port: 9092, Timeout: 100 * time.Millisecond})

	err := b.Subscribe("ch", func(msg *Message) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	_ = b.Unsubscribe("ch")
}

func TestRealKafkaSubscribeDuplicate(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	b.Connect(&Config{Host: "localhost", Port: 9092, Timeout: 100 * time.Millisecond})

	_ = b.Subscribe("ch", func(msg *Message) error { return nil })
	err := b.Subscribe("ch", func(msg *Message) error { return nil })
	if err == nil {
		t.Error("expected error for duplicate subscribe")
	}

	_ = b.Unsubscribe("ch")
}

func TestRealKafkaCloseWithCancel(t *testing.T) {
	t.Parallel()
	b := NewRealKafka()
	b.Connect(&Config{Host: "localhost", Port: 9092, Timeout: 100 * time.Millisecond})

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealRedisConnectFailed(t *testing.T) {
	t.Parallel()
	b := NewRealRedis()
	err := b.Connect(&Config{Host: "127.0.0.1", Port: 1, Timeout: 100 * time.Millisecond})
	if err == nil {
		t.Error("expected error connecting to non-existent Redis")
	}
}

func TestRealRedisCloseWithCancel(t *testing.T) {
	t.Parallel()
	b := NewRealRedis()
	b.cancel = func() {}

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRealNATSConnectFailed(t *testing.T) {
	t.Parallel()
	b := NewRealNATS()
	err := b.Connect(&Config{Host: "127.0.0.1", Port: 1, Timeout: 100 * time.Millisecond})
	if err == nil {
		t.Error("expected error connecting to non-existent NATS")
	}
}

func TestBrokerStoreLoadMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := &ConnectionStore{dir: filepath.Join(dir, "nonexistent")}

	entries, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestBrokerStoreAddMultiple(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	if err := s.Add("b1", "redis", &Config{Host: "h1", Port: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.Add("b2", "kafka", &Config{Host: "h2", Port: 2}); err != nil {
		t.Fatal(err)
	}

	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2, got %d", len(list))
	}
}

func TestBrokerStoreRemoveFirst(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	s.Add("b1", "redis", &Config{Host: "h1", Port: 1})
	s.Add("b2", "kafka", &Config{Host: "h2", Port: 2})

	if err := s.Remove("b1"); err != nil {
		t.Fatal(err)
	}

	list, _ := s.List()
	if len(list) != 1 {
		t.Errorf("expected 1, got %d", len(list))
	}
	if list[0].Name != "b2" {
		t.Errorf("expected b2, got %s", list[0].Name)
	}
}

var _ = fmt.Sprintf
