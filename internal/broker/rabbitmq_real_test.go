//go:build !nobroker

package broker

import (
	"testing"
)

func TestRealRabbitMQName(t *testing.T) {
	b := NewRealRabbitMQ()
	if b.Name() != "rabbitmq" {
		t.Errorf("expected name 'rabbitmq', got %s", b.Name())
	}
}

func TestRealRabbitMQNotConnected(t *testing.T) {
	b := NewRealRabbitMQ()

	err := b.Ping()
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = b.Publish("test", &Message{})
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = b.Subscribe("test", func(msg *Message) error { return nil })
	if err == nil {
		t.Error("expected error when not connected")
	}
}
