//go:build !nobroker

package broker

import (
	"testing"
)

func TestRealNATSName(t *testing.T) {
	b := NewRealNATS()
	if b.Name() != "nats" {
		t.Errorf("expected name 'nats', got %s", b.Name())
	}
}

func TestRealNATSConnectFailure(t *testing.T) {
	b := NewRealNATS()
	err := b.Connect(&Config{Host: "192.0.2.1", Port: 1, Timeout: 1})
	if err == nil {
		t.Error("expected error when connecting to unreachable host")
	}
}

func TestRealNATSNotConnected(t *testing.T) {
	b := NewRealNATS()

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
