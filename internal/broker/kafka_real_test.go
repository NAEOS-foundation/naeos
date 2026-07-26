//go:build !nobroker

package broker

import (
	"testing"
)

func TestRealKafkaName(t *testing.T) {
	b := NewRealKafka()
	if b.Name() != "kafka" {
		t.Errorf("expected name 'kafka', got %s", b.Name())
	}
}

func TestRealKafkaNotConnected(t *testing.T) {
	b := NewRealKafka()

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
