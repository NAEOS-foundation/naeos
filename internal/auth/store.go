package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const authConfigDir = ".config/naeos"
const authConfigFile = "auth.json"

type SavedUser struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at,omitempty"`
}

type UserStore struct {
	mu      sync.RWMutex
	dir     string
	entries []SavedUser
}

func NewUserStore() *UserStore {
	home, err := os.UserHomeDir()
	if err != nil {
		return &UserStore{dir: authConfigDir}
	}
	return &UserStore{dir: filepath.Join(home, authConfigDir)}
}

func (s *UserStore) filePath() string {
	return filepath.Join(s.dir, authConfigFile)
}

func (s *UserStore) load() error {
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			s.entries = nil
			return nil
		}
		return fmt.Errorf("read auth file: %w", err)
	}
	return json.Unmarshal(data, &s.entries)
}

func (s *UserStore) save() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create auth config dir: %w", err)
	}
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal users: %w", err)
	}
	return os.WriteFile(s.filePath(), data, 0o600)
}

func (s *UserStore) Add(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.load(); err != nil {
		return err
	}

	for i, e := range s.entries {
		if e.ID == user.ID {
			s.entries[i] = SavedUser{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Roles: user.Roles,
			}
			return s.save()
		}
	}

	entry := SavedUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Roles: user.Roles,
	}
	if !user.CreatedAt.IsZero() {
		entry.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	s.entries = append(s.entries, entry)
	return s.save()
}

func (s *UserStore) Get(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.load(); err != nil {
		return nil, false
	}

	for _, e := range s.entries {
		if e.ID == id {
			return &User{
				ID:    e.ID,
				Name:  e.Name,
				Email: e.Email,
				Roles: e.Roles,
			}, true
		}
	}
	return nil, false
}

func (s *UserStore) List() ([]SavedUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.load(); err != nil {
		return nil, err
	}
	return s.entries, nil
}
