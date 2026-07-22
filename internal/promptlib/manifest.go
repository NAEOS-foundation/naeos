package promptlib

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest represents the registry of all prompt templates in the library.
type Manifest struct {
	Prompts []PromptMeta `yaml:"prompts"`
}

// PromptMeta holds metadata about a single prompt template.
type PromptMeta struct {
	Name        string `yaml:"name"`
	Kind        string `yaml:"kind"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	Target      string `yaml:"target,omitempty"`
	Path        string `yaml:"path"`
}

// LoadManifest parses a YAML manifest from bytes.
func LoadManifest(data []byte) (*Manifest, error) {
	var m Manifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

// FindByName returns the PromptMeta for the given name, or nil if not found.
func (m *Manifest) FindByName(name string) *PromptMeta {
	for i := range m.Prompts {
		if m.Prompts[i].Name == name {
			return &m.Prompts[i]
		}
	}
	return nil
}

// FilterByKind returns all PromptMeta entries matching the given kind.
func (m *Manifest) FilterByKind(kind string) []PromptMeta {
	var result []PromptMeta
	for _, p := range m.Prompts {
		if p.Kind == kind {
			result = append(result, p)
		}
	}
	return result
}

// LoadPromptsFromDir loads all YAML prompt files from a directory tree.
// It walks subdirectories looking for .yaml and .yml files.
func LoadPromptsFromDir(dir string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open root %s: %w", dir, err)
	}
	defer root.Close()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("rel %s: %w", path, err)
		}
		f, err := root.Open(rel)
		if err != nil {
			return fmt.Errorf("open %s: %w", rel, err)
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		result[path] = data
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, err)
	}
	return result, nil
}
