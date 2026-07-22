package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const connectionsDir = ".naeos/db"
const connectionsFile = "connections.json"

type SavedConnection struct {
	Name   string  `json:"name"`
	Driver string  `json:"driver"`
	Config *Config `json:"config"`
}

type ConnectionStore struct {
	mu      sync.RWMutex
	dir     string
	entries []SavedConnection
}

func NewConnectionStore() *ConnectionStore {
	home, err := os.UserHomeDir()
	if err != nil {
		return &ConnectionStore{dir: connectionsDir}
	}
	return &ConnectionStore{dir: filepath.Join(home, connectionsDir)}
}

func (s *ConnectionStore) filePath() string {
	return filepath.Join(s.dir, connectionsFile)
}

func (s *ConnectionStore) load() error {
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			s.entries = nil
			return nil
		}
		return fmt.Errorf("read connections file: %w", err)
	}
	return json.Unmarshal(data, &s.entries)
}

func (s *ConnectionStore) save() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create connections dir: %w", err)
	}
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal connections: %w", err)
	}
	return os.WriteFile(s.filePath(), data, 0o600)
}

func (s *ConnectionStore) Add(name, driver string, config *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.load(); err != nil {
		return err
	}

	for _, e := range s.entries {
		if e.Name == name {
			return fmt.Errorf("connection %q already exists", name)
		}
	}

	s.entries = append(s.entries, SavedConnection{Name: name, Driver: driver, Config: config})
	return s.save()
}

func (s *ConnectionStore) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.load(); err != nil {
		return err
	}

	for i, e := range s.entries {
		if e.Name == name {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("connection %q not found", name)
}

func (s *ConnectionStore) Get(name string) (*SavedConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.load(); err != nil {
		return nil, err
	}

	for i := range s.entries {
		if s.entries[i].Name == name {
			return &s.entries[i], nil
		}
	}
	return nil, fmt.Errorf("connection %q not found", name)
}

func (s *ConnectionStore) List() ([]SavedConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.load(); err != nil {
		return nil, err
	}
	result := make([]SavedConnection, len(s.entries))
	copy(result, s.entries)
	return result, nil
}
