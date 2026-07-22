package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type failingDB struct {
	Database
}

func (f failingDB) Connect(config *Config) error {
	return errors.New("connection refused")
}

func (f failingDB) Close() error {
	return errors.New("close error")
}

func TestConnectAllError(t *testing.T) {
	m := NewManager()
	m.Register("failing", failingDB{})
	err := m.ConnectAll(map[string]*Config{
		"failing": {Host: "localhost"},
	})
	if err == nil {
		t.Fatal("expected error from failing db")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestManagerCloseAllWithErrorDetail(t *testing.T) {
	m := NewManager()
	m.Register("failing", failingDB{})
	err := m.CloseAll()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStoreLoadFileNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	if err := s.load(); err != nil {
		t.Fatalf("expected no error for missing file: %v", err)
	}
}

func TestStoreSaveDirError(t *testing.T) {
	s := &ConnectionStore{dir: "/nonexistent/deep/path"}
	err := s.Add("test", "pg", &Config{})
	if err == nil {
		t.Error("expected error for unwritable dir")
	}
}

func TestStoreRemoveNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	err := s.Remove("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}

func TestStoreListLoadError(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "connections.json"), []byte("{bad json}"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := s.List()
	if err == nil {
		t.Error("expected error from corrupted file")
	}
}

func TestStoreGetLoadError(t *testing.T) {
	dir := t.TempDir()
	s := &ConnectionStore{dir: dir}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "connections.json"), []byte("{bad json}"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := s.Get("x")
	if err == nil {
		t.Error("expected error from corrupted file")
	}
}

func TestIsTransientErrorMisc(t *testing.T) {
	if isTransientError(context.Canceled) {
		t.Error("expected false for context.Canceled")
	}
	if !isTransientError(context.DeadlineExceeded) {
		t.Error("expected true for DeadlineExceeded")
	}
}

type tempNetError struct{}

func (tempNetError) Timeout() bool   { return false }
func (tempNetError) Temporary() bool { return true }
func (tempNetError) Error() string   { return "temporary" }

func TestIsTransientErrorNetTemporary(t *testing.T) {
	if !isTransientError(tempNetError{}) {
		t.Error("expected true for temporary net.Error")
	}
}

type notNetError struct{}

func (notNetError) Error() string { return "some error" }

func TestIsTransientErrorNonNet(t *testing.T) {
	if isTransientError(notNetError{}) {
		t.Error("expected false for non-net error")
	}
}

type netOpError struct{}

func (netOpError) Error() string   { return "dial tcp: connection refused" }
func (netOpError) Timeout() bool   { return false }
func (netOpError) Temporary() bool { return false }

func TestIsTransientErrorDialRefused(t *testing.T) {
	if !isTransientError(netOpError{}) {
		t.Error("expected true for dial refused")
	}
}

func TestNewConnectionStoreHomeDir(t *testing.T) {
	s := NewConnectionStore()
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestWithRetryExceeded(t *testing.T) {
	attempts := 0
	err := WithRetry(context.Background(), 2, time.Millisecond, func(ctx context.Context) error {
		attempts++
		return errors.New("connection refused")
	})
	if err == nil {
		t.Error("expected error")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts (0..maxRetries), got %d", attempts)
	}
}

func TestLoadMigrationsFileError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.up.sql"), []byte("SELECT 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	migs, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if len(migs) != 0 {
		t.Errorf("expected 0 migrations for unversioned file, got %d", len(migs))
	}
}
