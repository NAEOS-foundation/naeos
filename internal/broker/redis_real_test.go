//go:build !nobroker

package broker

import (
	"testing"
)

func TestRealRedisName(t *testing.T) {
	b := NewRealRedis()
	if b.Name() != "redis" {
		t.Errorf("expected name 'redis', got %s", b.Name())
	}
}

func TestRealRedisNotConnected(t *testing.T) {
	b := NewRealRedis()

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
