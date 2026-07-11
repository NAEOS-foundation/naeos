package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func loadInput(input, inputFile string) (string, error) {
	if input == "" && inputFile == "" {
		return "", fmt.Errorf("missing required --input or --input-file")
	}
	if input != "" && inputFile != "" {
		return "", fmt.Errorf("cannot use both --input and --input-file")
	}
	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return "", fmt.Errorf("read input file: %w", err)
		}
		return string(data), nil
	}
	return input, nil
}

func resolveInput(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	info, err := os.Stat(input)
	if err != nil {
		if os.IsNotExist(err) {
			return input, nil
		}
		return "", err
	}
	if info.IsDir() {
		return input, nil
	}

	content, err := os.ReadFile(input)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func renderOutput(data any, format string, defaultFn func() []byte) ([]byte, error) {
	switch format {
	case "json":
		result, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encode json output: %w", err)
		}
		return append(result, '\n'), nil
	case "yaml":
		result, err := yaml.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("encode yaml output: %w", err)
		}
		return result, nil
	default:
		return defaultFn(), nil
	}
}

func writeOrPrint(cmd *cobra.Command, data []byte, filePath string) error {
	if filePath != "" {
		return os.WriteFile(filePath, data, 0o600)
	}
	_, err := cmd.OutOrStdout().Write(data)
	return err
}

func writeFileInDir(dir, fileName, content string) error {
	path := filepath.Join(dir, fileName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create scaffold dir for %s: %w", fileName, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write scaffold %s: %w", fileName, err)
	}
	return nil
}

func resolveConfigPath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}

	candidates := []string{
		"config.yaml",
		"config.yml",
		"config.json",
		"naeos.yaml",
		"naeos.yml",
		"naeos.json",
		".naeos/config.yaml",
		".naeos/config.yml",
		".naeos/config.json",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("missing required --config (no config file found in current directory)")
}

func mustConfigPath(cmd *cobra.Command, configPath string) (string, error) {
	resolved, err := resolveConfigPath(configPath)
	if err != nil {
		return "", err
	}
	return resolved, nil
}

func loadPipelineConfig(configPath string, verbose bool, languages []string, dryRun bool) (*pipeline.Config, string, error) {
	resolved, err := resolveConfigPath(configPath)
	if err != nil {
		return nil, "", err
	}

	cfg, err := pipeline.ConfigFromFile(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}
	if verbose {
		cfg.Verbose = true
	}
	if len(languages) > 0 {
		cfg.Languages = languages
	}
	if dryRun {
		cfg.DryRun = true
	}
	return &cfg, resolved, nil
}
