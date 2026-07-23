package broker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

const brokerConfigDir = ".config/naeos"
const brokerConfigFile = "brokers.json"

type SavedBroker struct {
	Name   string  `json:"name"`
	Driver string  `json:"driver"`
	Config *Config `json:"config"`
}

type ConnectionStore struct {
	mu      sync.RWMutex
	dir     string
	entries []SavedBroker
}

func NewConnectionStore() *ConnectionStore {
	home, err := os.UserHomeDir()
	if err != nil {
		return &ConnectionStore{dir: brokerConfigDir}
	}
	return &ConnectionStore{dir: filepath.Join(home, brokerConfigDir)}
}

func (s *ConnectionStore) filePath() string {
	return filepath.Join(s.dir, brokerConfigFile)
}

func (s *ConnectionStore) load() error {
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			s.entries = nil
			return nil
		}
		return fmt.Errorf("read broker connections file: %w", err)
	}
	return json.Unmarshal(data, &s.entries)
}

func (s *ConnectionStore) save() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create broker config dir: %w", err)
	}
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal broker connections: %w", err)
	}
	return os.WriteFile(s.filePath(), data, 0o600)
}

func (s *ConnectionStore) Add(name, driver string, config *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.load(); err != nil {
		return err
	}

	for i, e := range s.entries {
		if e.Name == name {
			s.entries[i].Driver = driver
			s.entries[i].Config = config
			return s.save()
		}
	}

	s.entries = append(s.entries, SavedBroker{Name: name, Driver: driver, Config: config})
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

	return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("broker connection %q", name))
	return naeoserr.Wrap(naeoserr.ErrNotFound, fmt.Sprintf("broker connection %q not found", name), nil)
}

func (s *ConnectionStore) Get(name string) (*SavedBroker, error) {
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
	return nil, naeoserr.Wrap(naeoserr.ErrNotFound, fmt.Sprintf("broker connection %q not found", name), nil)
}

func (s *ConnectionStore) List() ([]SavedBroker, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.load(); err != nil {
		return nil, err
	}
	result := make([]SavedBroker, len(s.entries))
	copy(result, s.entries)
	return result, nil
}
