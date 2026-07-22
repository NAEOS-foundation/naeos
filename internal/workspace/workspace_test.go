package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/test")
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
	if m.rootDir != "/tmp/test" {
		t.Errorf("expected rootDir /tmp/test, got %s", m.rootDir)
	}
}

func TestManagerInit(t *testing.T) {
	t.Run("creates workspace directory", func(t *testing.T) {
		dir := t.TempDir()
		m := NewManager(dir)
		ws, err := m.Init("myapp")
		if err != nil {
			t.Fatal(err)
		}
		if ws.Name != "myapp" {
			t.Errorf("expected name myapp, got %s", ws.Name)
		}
		expectedRoot := filepath.Join(dir, "myapp")
		if ws.Root != expectedRoot {
			t.Errorf("expected root %s, got %s", expectedRoot, ws.Root)
		}
		if _, err := os.Stat(expectedRoot); os.IsNotExist(err) {
			t.Error("workspace directory was not created")
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		m := NewManager(t.TempDir())
		_, err := m.Init("")
		if err == nil {
			t.Error("expected error for empty name")
		}
	})

	t.Run("succeeds when directory already exists", func(t *testing.T) {
		dir := t.TempDir()
		m := NewManager(dir)
		os.MkdirAll(filepath.Join(dir, "existing"), 0o755)
		_, err := m.Init("existing")
		if err != nil {
			t.Errorf("expected no error for existing dir, got %v", err)
		}
	})
}

func TestManagerAddModule(t *testing.T) {
	t.Run("creates module directory", func(t *testing.T) {
		dir := t.TempDir()
		m := NewManager(dir)
		err := m.AddModule("auth", "modules/auth", "Auth module", []string{})
		if err != nil {
			t.Fatal(err)
		}
		modPath := filepath.Join(dir, "modules/auth")
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			t.Error("module directory was not created")
		}
	})

	t.Run("returns error for empty module name", func(t *testing.T) {
		m := NewManager(t.TempDir())
		err := m.AddModule("", "path", "desc", nil)
		if err == nil {
			t.Error("expected error for empty name")
		}
	})

	t.Run("returns error for empty path", func(t *testing.T) {
		m := NewManager(t.TempDir())
		err := m.AddModule("auth", "", "desc", nil)
		if err == nil {
			t.Error("expected error for empty path")
		}
	})
}

func TestManagerListModules(t *testing.T) {
	t.Run("returns empty slice for empty directory", func(t *testing.T) {
		dir := t.TempDir()
		m := NewManager(dir)
		mods, err := m.ListModules()
		if err != nil {
			t.Fatal(err)
		}
		if len(mods) != 0 {
			t.Errorf("expected 0 modules, got %d", len(mods))
		}
	})

	t.Run("lists only directories", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "auth"), 0o755)
		os.MkdirAll(filepath.Join(dir, "api"), 0o755)
		os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644)
		m := NewManager(dir)
		mods, err := m.ListModules()
		if err != nil {
			t.Fatal(err)
		}
		if len(mods) != 2 {
			t.Errorf("expected 2 modules, got %d", len(mods))
		}
	})

	t.Run("returns nil for non-existent root", func(t *testing.T) {
		m := NewManager("/nonexistent/path")
		mods, err := m.ListModules()
		if err != nil {
			t.Fatal(err)
		}
		if mods != nil {
			t.Errorf("expected nil, got %v", mods)
		}
	})
}

func TestManagerRemoveModule(t *testing.T) {
	t.Run("removes existing module", func(t *testing.T) {
		dir := t.TempDir()
		m := NewManager(dir)
		modPath := filepath.Join(dir, "auth")
		os.MkdirAll(modPath, 0o755)
		err := m.AddModule("auth", "auth", "desc", nil)
		if err != nil {
			t.Fatal(err)
		}
		err = m.RemoveModule("auth")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(modPath); !os.IsNotExist(err) {
			t.Error("module directory should have been removed")
		}
	})

	t.Run("returns error for non-existent module", func(t *testing.T) {
		dir := t.TempDir()
		m := NewManager(dir)
		err := m.RemoveModule("ghost")
		if err == nil {
			t.Error("expected error for non-existent module")
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		m := NewManager(t.TempDir())
		err := m.RemoveModule("")
		if err == nil {
			t.Error("expected error for empty name")
		}
	})
}
