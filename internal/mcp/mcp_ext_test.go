package mcp

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

func TestSetArtifactStore(t *testing.T) {
	s := newTestServer()
	store := artifacts.NewStore(t.TempDir())
	s.SetArtifactStore(store)
	if s.store != store {
		t.Error("expected store to be set")
	}
}

func TestSetPluginManager(t *testing.T) {
	s := newTestServer()
	mgr := pluginhost.NewManager(t.TempDir())
	s.SetPluginManager(mgr)
	if s.pluginMgr != mgr {
		t.Error("expected plugin manager to be set")
	}
}

func TestHandleListArtifactsWithStore(t *testing.T) {
	s := newTestServer()
	store := artifacts.NewStore(t.TempDir())
	store.Add("test.txt", []byte("hello"), artifacts.KindCode)
	s.SetArtifactStore(store)

	result, err := s.handleListArtifacts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestHandleListPluginsWithManager(t *testing.T) {
	s := newTestServer()
	mgr := pluginhost.NewManager(t.TempDir())
	s.SetPluginManager(mgr)

	result, err := s.handleListPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty content")
	}
}
