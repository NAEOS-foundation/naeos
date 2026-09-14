// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Action captures a single task or tool invocation associated with a session.
type Action struct {
	ID         string         `json:"id"`
	SessionID  string         `json:"session_id"`
	AgentID    string         `json:"agent_id,omitempty"`
	Type       string         `json:"type"`
	Target     string         `json:"target"`
	Reason     string         `json:"reason,omitempty"`
	SpecRefs   []string       `json:"spec_refs,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
	Decision   string         `json:"decision,omitempty"`
	PolicyID   string         `json:"policy_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// Session represents an autonomous task session and its visible state.
type Session struct {
	ID                   string    `json:"id"`
	AgentID              string    `json:"agent_id"`
	Project              string    `json:"project,omitempty"`
	Workspace            string    `json:"workspace,omitempty"`
	Specification        string    `json:"specification,omitempty"`
	SpecificationVersion string    `json:"specification_version,omitempty"`
	Status               string    `json:"status"`
	Permissions          []string  `json:"permissions,omitempty"`
	Denied               []string  `json:"denied,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Actions              []Action  `json:"actions,omitempty"`
}

type storeFile struct {
	Sessions []Session `json:"sessions"`
}

// Store is a lightweight persisted session store for the control-plane
// foundation. It intentionally uses existing NAEOS JSON conventions and can be
// expanded later without changing the public shape.
type Store struct {
	path     string
	mu       sync.RWMutex
	sessions []Session
}

func NewStore(path string) *Store {
	if path == "" {
		path = filepath.Join(os.TempDir(), "naeos-agent-store.json")
	}
	return &Store{path: path}
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.sessions = nil
			return nil
		}
		return fmt.Errorf("read store: %w", err)
	}
	if len(data) == 0 {
		s.sessions = nil
		return nil
	}

	var raw storeFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse store: %w", err)
	}
	s.sessions = raw.Sessions
	return nil
}

func (s *Store) Save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(storeFile{Sessions: s.sessions}, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("marshal store: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create store dir: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	return nil
}

func (s *Store) CreateSession(session Session) (Session, error) {
	if session.AgentID == "" {
		return Session{}, fmt.Errorf("agent_id is required")
	}
	if session.Status == "" {
		session.Status = "active"
	}
	if session.ID == "" {
		session.ID = newID("sess")
	}
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.sessions {
		if s.sessions[i].ID == session.ID {
			return Session{}, fmt.Errorf("session %s already exists", session.ID)
		}
	}
	s.sessions = append(s.sessions, session)
	return session, nil
}

func (s *Store) GetSession(id string) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, session := range s.sessions {
		if session.ID == id {
			return session, true
		}
	}
	return Session{}, false
}

func (s *Store) ListSessions() []Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Session, len(s.sessions))
	copy(out, s.sessions)
	return out
}

func (s *Store) ListActions() []Action {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []Action
	for _, session := range s.sessions {
		out = append(out, session.Actions...)
	}
	return out
}

func (s *Store) ListSessionActions(sessionID string) []Action {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, session := range s.sessions {
		if session.ID == sessionID {
			out := make([]Action, len(session.Actions))
			copy(out, session.Actions)
			return out
		}
	}
	return nil
}

func (s *Store) AppendAction(sessionID string, action Action) (Action, error) {
	if sessionID == "" {
		return Action{}, fmt.Errorf("session_id is required")
	}
	if action.Type == "" || action.Target == "" {
		return Action{}, fmt.Errorf("action type and target are required")
	}
	if action.ID == "" {
		action.ID = newID("act")
	}
	if action.CreatedAt.IsZero() {
		action.CreatedAt = time.Now().UTC()
	}
	if action.SessionID == "" {
		action.SessionID = sessionID
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.sessions {
		if s.sessions[i].ID != sessionID {
			continue
		}
		s.sessions[i].UpdatedAt = time.Now().UTC()
		s.sessions[i].Actions = append(s.sessions[i].Actions, action)
		return action, nil
	}
	return Action{}, fmt.Errorf("session %s not found", sessionID)
}

func (s *Store) GetAction(id string) (Action, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, session := range s.sessions {
		for _, action := range session.Actions {
			if action.ID == id {
				return action, true
			}
		}
	}
	return Action{}, false
}

func (s *Store) DeleteSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, session := range s.sessions {
		if session.ID == id {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("session %s not found", id)
}

func (s *Store) Path() string { return s.path }

func newID(prefix string) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

func SortSessionsByCreatedAtDescending(sessions []Session) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})
}
