package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := NewState(path)
	if got := s.LastRelease(); got != "" {
		t.Fatalf("expected empty state, got %q", got)
	}

	s.SetLastRelease("v3.0.0")
	if err := s.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded := NewState(path)
	if got := loaded.LastRelease(); got != "v3.0.0" {
		t.Fatalf("expected v3.0.0 after reload, got %q", got)
	}
}

func TestStateAnnounceChannelRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := NewState(path)
	s.SetAnnounceChannel("123456789")
	if err := s.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded := NewState(path)
	if got := loaded.AnnounceChannel(); got != "123456789" {
		t.Fatalf("expected channel 123456789 after reload, got %q", got)
	}
}

func TestStateIgnoresCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	s := NewState(path)
	if got := s.LastRelease(); got != "" {
		t.Fatalf("expected empty state for corrupt file, got %q", got)
	}
}
