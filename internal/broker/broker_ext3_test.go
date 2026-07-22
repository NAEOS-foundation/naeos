package broker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Mock brokers for error path testing

type connectFailBroker struct {
	InMemoryBroker
}

func (c *connectFailBroker) Connect(config *Config) error {
	return fmt.Errorf("connect failed")
}

type closeFailBroker struct {
	InMemoryBroker
}

func (c *closeFailBroker) Close() error {
	return fmt.Errorf("close failed")
}

func TestManagerConnectAllError(t *testing.T) {
	m := NewManager()
	m.Register("fail", &connectFailBroker{})
	err := m.ConnectAll(map[string]*Config{"fail": {}})
	if err == nil {
		t.Fatal("expected error from ConnectAll")
	}
}

func TestManagerCloseAllError(t *testing.T) {
	m := NewManager()
	b := &closeFailBroker{}
	m.Register("fail", b)
	// Connect first so it's in the brokers map
	b.connected = true
	err := m.CloseAll()
	if err == nil {
		t.Fatal("expected error from CloseAll")
	}
}

func TestConnectionPoolCloseAllError(t *testing.T) {
	b := &closeFailBroker{}
	pool := NewConnectionPool(b)
	err := pool.CloseAll()
	if err == nil {
		t.Fatal("expected error from pool CloseAll")
	}
}

func TestNewRealDriverBranches(t *testing.T) {
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
			t.Errorf("expected %q, got %q", tt.name, b.Name())
		}
	}
}

func TestNewFromConfigConnectFailure(t *testing.T) {
	_, err := NewFromConfig("redis", &Config{Host: "127.0.0.1", Port: 1})
	if err == nil {
		t.Fatal("expected error from connect failure")
	}
}

func TestStoreAddLoadErrorExt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "brokers.json"), []byte("not valid json"), 0o600)
	s := &ConnectionStore{dir: dir}
	err := s.Add("test", "memory", &Config{})
	if err == nil {
		t.Fatal("expected error from corrupt file")
	}
}

func TestStoreRemoveLoadErrorExt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "brokers.json"), []byte("{bad json"), 0o600)
	s := &ConnectionStore{dir: dir}
	err := s.Remove("test")
	if err == nil {
		t.Fatal("expected error from corrupt file")
	}
}

func TestStoreListLoadErrorExt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "brokers.json"), []byte("garbage"), 0o600)
	s := &ConnectionStore{dir: dir}
	_, err := s.List()
	if err == nil {
		t.Fatal("expected error from corrupt file")
	}
}

func TestStoreGetLoadErrorExt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "brokers.json"), []byte("{{bad"), 0o600)
	s := &ConnectionStore{dir: dir}
	_, err := s.Get("test")
	if err == nil {
		t.Fatal("expected error from corrupt file")
	}
}

func TestNewConnectionStoreHomeDirError(t *testing.T) {
	t.Setenv("HOME", "")
	s := NewConnectionStore()
	if s.dir != brokerConfigDir {
		t.Errorf("expected fallback dir %q, got %q", brokerConfigDir, s.dir)
	}
}

func TestInMemoryBrokerPublishChannelOverflow(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})
	for i := 0; i < 257; i++ {
		b.Publish("ch", NewMessage("ch", []byte("x")))
	}
}

func TestConnectionStoreLoadNoFile(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	err := s.load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
}

func TestConnectionStoreSaveMkdirError(t *testing.T) {
	s := &ConnectionStore{dir: "/proc/fake/path/that/cant/be/created"}
	err := s.save()
	if err == nil {
		t.Fatal("expected error from save with bad dir")
	}
}

func TestConnectionStoreAddNewAndRemove(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	err := s.Add("my-broker", "memory", &Config{Host: "localhost", Port: 1883})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	brokers, err := s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(brokers) != 1 {
		t.Fatalf("expected 1 broker, got %d", len(brokers))
	}

	got, err := s.Get("my-broker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Driver != "memory" {
		t.Errorf("expected driver memory, got %s", got.Driver)
	}

	err = s.Remove("my-broker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	brokers, err = s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(brokers) != 0 {
		t.Errorf("expected 0 brokers after remove, got %d", len(brokers))
	}
}

func TestConnectionStoreGetNotFoundExt(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	_, err := s.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestConnectionStoreRemoveNotFoundExt(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	err := s.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestConnectionStoreUpdateExisting(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}

	s.Add("broker1", "memory", &Config{Host: "h1", Port: 1})
	s.Add("broker1", "redis", &Config{Host: "h2", Port: 2})

	brokers, _ := s.List()
	if len(brokers) != 1 {
		t.Fatalf("expected 1 broker after update, got %d", len(brokers))
	}
	if brokers[0].Driver != "redis" {
		t.Errorf("expected driver redis after update, got %s", brokers[0].Driver)
	}
}

func TestConnectionStoreLoadReadError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "brokers.json")
	os.WriteFile(path, []byte("[]"), 0o000) // no read permissions
	s := &ConnectionStore{dir: dir}
	err := s.load()
	if err == nil {
		// If running as root, permissions won't restrict
		t.Skip("running as root, permissions not enforced")
	}
}

func TestConnectionStoreLoadValidJSON(t *testing.T) {
	dir := t.TempDir()
	data := []json.RawMessage{[]byte(`{"name":"b1","driver":"memory","config":null}`)}
	b, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filepath.Join(dir, "brokers.json"), b, 0o600)
	s := &ConnectionStore{dir: dir}
	err := s.load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConnectionPoolCheckHealthAllBrokers(t *testing.T) {
	b := NewInMemoryBroker()
	b.Connect(&Config{})
	pool := NewConnectionPool(b)

	metrics := pool.PoolMetrics()
	if metrics.Total != 1 {
		t.Errorf("expected 1 total, got %d", metrics.Total)
	}
	if metrics.Healthy != 1 {
		t.Errorf("expected 1 healthy, got %d", metrics.Healthy)
	}
}
