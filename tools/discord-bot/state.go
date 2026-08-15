package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// State persists the last-seen release tag and the announcement channel ID
// between runs so restarts don't re-announce old releases or lose the setup.
type State struct {
	mu   sync.Mutex
	path string

	lastRelease     string
	announceChannel string
}

// NewState loads state from path (missing or corrupt files start empty).
func NewState(path string) *State {
	s := &State{path: path}
	s.load()
	return s
}

func (s *State) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var raw struct {
		LastRelease     string `json:"last_release"`
		AnnounceChannel string `json:"announce_channel"`
	}
	if json.Unmarshal(data, &raw) == nil {
		s.lastRelease = raw.LastRelease
		s.announceChannel = raw.AnnounceChannel
	}
}

// LastRelease returns the last announced release tag.
func (s *State) LastRelease() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastRelease
}

// SetLastRelease records a new last-seen release tag (persisted on the next save).
func (s *State) SetLastRelease(tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastRelease = tag
}

// AnnounceChannel returns the state-persisted announcement channel ID.
func (s *State) AnnounceChannel() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.announceChannel
}

// SetAnnounceChannel records the announcement channel ID (persisted on the next save).
func (s *State) SetAnnounceChannel(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.announceChannel = channelID
}

// Save persists state to disk.
func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path == "" {
		return errors.New("state path is empty")
	}
	payload, err := json.MarshalIndent(map[string]string{
		"last_release":     s.lastRelease,
		"announce_channel": s.announceChannel,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, payload, 0o600)
}
