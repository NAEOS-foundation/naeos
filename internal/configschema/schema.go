package configschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Schema struct {
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties"`
	Required   []string               `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func DefaultSchema() *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]Property{
			"name":        {Type: "string", Description: "project name"},
			"version":     {Type: "string", Description: "project version"},
			"description": {Type: "string", Description: "project description"},
			"output_dir":  {Type: "string", Description: "output directory", Default: "./output"},
			"mode":        {Type: "string", Description: "pipeline mode", Default: "standard"},
			"verbose":     {Type: "boolean", Description: "verbose output", Default: false},
			"languages":   {Type: "array", Description: "target languages"},
			"dry_run":     {Type: "boolean", Description: "dry run mode", Default: false},
		},
		Required: []string{"name"},
	}
}

func ValidateFile(path string) ([]ValidationError, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return validateYAML(data)
	case ".json":
		return validateJSON(data)
	default:
		return validateJSON(data)
	}
}

func ValidateData(data []byte, format string) []ValidationError {
	var config map[string]any
	switch format {
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return []ValidationError{{Field: "_root", Message: "invalid YAML: " + err.Error()}}
		}
	default:
		if err := json.Unmarshal(data, &config); err != nil {
			return []ValidationError{{Field: "_root", Message: "invalid JSON: " + err.Error()}}
		}
	}

	return ValidateConfig(config)
}

func ValidateConfig(config map[string]any) []ValidationError {
	schema := DefaultSchema()
	var errors []ValidationError

	for _, required := range schema.Required {
		if _, ok := config[required]; !ok {
			errors = append(errors, ValidationError{
				Field:   required,
				Message: fmt.Sprintf("required field '%s' is missing", required),
			})
		}
	}

	for key, val := range config {
		prop, ok := schema.Properties[key]
		if !ok {
			continue
		}
		if !validateType(val, prop.Type) {
			errors = append(errors, ValidationError{
				Field:   key,
				Message: fmt.Sprintf("field '%s' should be of type %s", key, prop.Type),
			})
		}
	}

	return errors
}

func validateType(val any, expected string) bool {
	switch expected {
	case "string":
		_, ok := val.(string)
		return ok
	case "boolean":
		_, ok := val.(bool)
		return ok
	case "number":
		switch val.(type) {
		case int, int64, float64:
			return true
		}
		return false
	case "array":
		_, ok := val.([]any)
		return ok
	case "object":
		_, ok := val.(map[string]any)
		return ok
	}
	return true
}

func validateYAML(data []byte) ([]ValidationError, error) {
	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return []ValidationError{{Field: "_root", Message: "invalid YAML: " + err.Error()}}, nil
	}
	return ValidateConfig(config), nil
}

func validateJSON(data []byte) ([]ValidationError, error) {
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return []ValidationError{{Field: "_root", Message: "invalid JSON: " + err.Error()}}, nil
	}
	return ValidateConfig(config), nil
}
