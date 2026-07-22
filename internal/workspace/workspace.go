package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

type Module struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type Workspace struct {
	Name    string   `json:"name"`
	Root    string   `json:"root"`
	Modules []Module `json:"modules"`
}

type Manager struct {
	rootDir string
}

func NewManager(rootDir string) *Manager {
	return &Manager{rootDir: rootDir}
}

func (m *Manager) Init(name string) (*Workspace, error) {
	if name == "" {
		return nil, fmt.Errorf("workspace name is required")
	}
	wsDir := filepath.Join(m.rootDir, name)
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create workspace dir: %w", err)
	}
	return &Workspace{
		Name: name,
		Root: wsDir,
	}, nil
}

func (m *Manager) AddModule(name, path, description string, deps []string) error {
	if name == "" {
		return fmt.Errorf("module name is required")
	}
	if path == "" {
		return fmt.Errorf("module path is required")
	}
	fullPath := filepath.Join(m.rootDir, path)
	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		return fmt.Errorf("create module dir: %w", err)
	}
	return nil
}

func (m *Manager) RemoveModule(name string) error {
	if name == "" {
		return fmt.Errorf("module name is required")
	}
	modules, err := m.ListModules()
	if err != nil {
		return err
	}
	for _, mod := range modules {
		if mod.Name == name {
			return os.RemoveAll(filepath.Join(m.rootDir, mod.Path))
		}
	}
	return fmt.Errorf("module %q not found", name)
}

func (m *Manager) ListModules() ([]Module, error) {
	entries, err := os.ReadDir(m.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var modules []Module
	for _, entry := range entries {
		if entry.IsDir() {
			modules = append(modules, Module{
				Name: entry.Name(),
				Path: entry.Name(),
			})
		}
	}
	return modules, nil
}
