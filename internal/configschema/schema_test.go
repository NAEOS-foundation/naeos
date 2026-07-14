package configschema

import (
	"os"
	"encoding/json"
	"testing"
)

func TestValidateConfigRequired(t *testing.T) {
	config := map[string]any{
		"description": "no name",
	}
	errs := ValidateConfig(config)
	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected required field error for 'name', got %v", errs)
	}
}

func TestValidateConfigTypes(t *testing.T) {
	config := map[string]any{
		"name":    123,
		"verbose": "notbool",
	}
	errs := ValidateConfig(config)
	if len(errs) < 2 {
		t.Errorf("expected >=2 type errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateConfigValid(t *testing.T) {
	config := map[string]any{
		"name":    "myproject",
		"version": "1.0.0",
		"verbose": true,
	}
	errs := ValidateConfig(config)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateDataJSON(t *testing.T) {
	data := []byte(`{"name":"test","verbose":true}`)
	errs := ValidateData(data, "json")
	if len(errs) > 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateDataYAML(t *testing.T) {
	data := []byte("name: test\nverbose: true")
	errs := ValidateData(data, "yaml")
	if len(errs) > 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateDataInvalidJSON(t *testing.T) {
	data := []byte(`{not json}`)
	errs := ValidateData(data, "json")
	if len(errs) == 0 {
		t.Error("expected errors for invalid JSON")
	}
}

func TestValidateDataMissingRequired(t *testing.T) {
	data, _ := json.Marshal(map[string]any{
		"version": "1.0.0",
	})
	errs := ValidateData(data, "json")
	if len(errs) == 0 {
		t.Error("expected missing required error")
	}
}

func TestDefaultSchema(t *testing.T) {
	s := DefaultSchema()
	if s.Type != "object" {
		t.Errorf("expected type 'object', got %s", s.Type)
	}
	if len(s.Required) == 0 {
		t.Error("expected required fields")
	}
	if _, ok := s.Properties["name"]; !ok {
		t.Error("expected 'name' property")
	}
}

func TestValidateFileYAML(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.yaml"
	data := []byte("name: test\nversion: \"1.0\"\nverbose: true")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	errs, err := ValidateFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) > 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidateFileJSON(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.json"
	data := []byte(`{"name": "test", "version": "1.0", "verbose": true}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	errs, err := ValidateFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) > 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidateFileUnknown(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.txt"
	data := []byte(`{"name": "test"}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	// .txt falls through to JSON validation by default
	errs, err := ValidateFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) > 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidateFileNotFound(t *testing.T) {
	_, err := ValidateFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestValidateDataInvalidYAML(t *testing.T) {
	data := []byte("{{invalid yaml: [}")
	errs := ValidateData(data, "yaml")
	if len(errs) == 0 {
		t.Error("expected errors for invalid YAML")
	}
	if errs[0].Field != "_root" {
		t.Errorf("expected field '_root', got %s", errs[0].Field)
	}
}

func TestValidateTypeNumber(t *testing.T) {
	schema := DefaultSchema()
	// int type
	if !validateType(42, "number") {
		t.Error("expected int to be valid for 'number'")
	}
	// int64 type
	if !validateType(int64(100), "number") {
		t.Error("expected int64 to be valid for 'number'")
	}
	// float64 type
	if !validateType(float64(3.14), "number") {
		t.Error("expected float64 to be valid for 'number'")
	}
	// string should not be valid for number
	if validateType("not a number", "number") {
		t.Error("expected string to be invalid for 'number'")
	}
	// exercise ValidateConfig with number-typed value in a known property
	config := map[string]any{
		"name":    "project",
		"verbose": float64(1),
	}
	errs := ValidateConfig(config)
	// verbose expects boolean; number(1) is not boolean
	found := false
	for _, e := range errs {
		if e.Field == "verbose" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected type error for verbose with numeric value, got %v", errs)
	}
	_ = schema
}

func TestValidateTypeObject(t *testing.T) {
	// map[string]any should be valid for "object" type
	obj := map[string]any{"key": "value"}
	if !validateType(obj, "object") {
		t.Error("expected map[string]any to be valid for 'object'")
	}
	// a slice should not be valid for "object"
	slice := []any{"a", "b"}
	if validateType(slice, "object") {
		t.Error("expected []any to be invalid for 'object'")
	}
	// string should not be valid for "object"
	if validateType("string", "object") {
		t.Error("expected string to be invalid for 'object'")
	}
}

func TestValidateTypeUnknown(t *testing.T) {
	// Unknown type should always return true (no validation applied)
	if !validateType(123, "unknown_type") {
		t.Error("expected any value to be valid for unknown type")
	}
	if !validateType("hello", "custom") {
		t.Error("expected any value to be valid for custom type")
	}
	if !validateType(nil, "nonexistent") {
		t.Error("expected nil to be valid for nonexistent type")
	}
}
