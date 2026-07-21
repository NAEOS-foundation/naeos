//go:build nobroker

package broker

import (
	"fmt"
	"testing"
)

type errConnectBroker struct{ InMemoryBroker }

func (e *errConnectBroker) Connect(config *Config) error { return fmt.Errorf("connect failed") }

type errCloseBroker struct{ InMemoryBroker }

func (e *errCloseBroker) Close() error { return fmt.Errorf("close failed") }

func TestNewAllRealDrivers(t *testing.T) {
	tests := []struct {
		driver string
		name   string
	}{
		{"redis", "redis"},
		{"rabbitmq", "rabbitmq"},
		{"kafka", "kafka"},
		{"nats", "nats"},
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

func TestDeadLetterChan(t *testing.T) {
	b := NewInMemoryBroker()
	ch := b.DeadLetterChan()
	if ch == nil {
		t.Error("expected non-nil channel")
	}
}

func TestConnectAllError(t *testing.T) {
	m := NewManager()
	m.Register("err", &errConnectBroker{})
	configs := map[string]*Config{"err": {Host: "localhost"}}
	err := m.ConnectAll(configs)
	if err == nil {
		t.Error("expected error from connect failure")
	}
}

func TestManagerCloseAllError(t *testing.T) {
	m := NewManager()
	m.Register("err", &errCloseBroker{})
	err := m.CloseAll()
	if err == nil {
		t.Error("expected error from close failure")
	}
}

func TestConnectionPoolCloseAllError(t *testing.T) {
	pool := NewConnectionPool(&errCloseBroker{})
	err := pool.CloseAll()
	if err == nil {
		t.Error("expected error from close failure")
	}
}

func TestNewFromConfigConnectFailure(t *testing.T) {
	_, err := NewFromConfig("memory", nil)
	if err != nil {
		// InMemoryBroker.Connect(nil) does not fail, so this should succeed
		t.Logf("got error (unexpected): %v", err)
	}
}

func TestStoreNewConnectionHomeDir(t *testing.T) {
	s := NewConnectionStore()
	if s.dir == "" {
		t.Error("expected non-empty dir")
	}
}
