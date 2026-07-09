package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// File holds pipeline configuration loaded from disk.
type File struct {
	Pipeline Pipeline `json:"pipeline" yaml:"pipeline"`
}

// Pipeline contains the configurable components for the NAEOS pipeline.
type Pipeline struct {
	Name      string `json:"name" yaml:"name"`
	Mode      string `json:"mode" yaml:"mode"`
	Verbose   bool   `json:"verbose" yaml:"verbose"`
	OutputDir string `json:"output_dir" yaml:"output_dir"`
}

// LoadFile reads configuration from a JSON or YAML file.
func LoadFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg File
	if err := parse(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parse(data []byte, out *File) error {
	if len(data) == 0 {
		return fmt.Errorf("config is empty")
	}

	if err := json.Unmarshal(data, out); err == nil {
		return nil
	}

	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	return nil
}
