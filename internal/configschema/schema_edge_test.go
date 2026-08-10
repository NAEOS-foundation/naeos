package configschema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddValidator(t *testing.T) {
	called := false
	b := NewBuilder()
	b.AddValidator("string", func(value any, prop Property) *ValidationError {
		called = true
		return nil
	})
	if len(b.validators) != 1 {
		t.Errorf("expected 1 validator registered, got %d", len(b.validators))
	}
	if called {
		t.Error("expected validator not to be invoked during registration")
	}
}
func TestAddCustomType(t *testing.T) {
	b := NewBuilder()
	b.AddCustomType("uuid", func(v any) bool {
		s, ok := v.(string)
		return ok && len(s) == 36
	})
	if len(b.customTypes) != 1 {
		t.Errorf("expected 1 custom type registered, got %d", len(b.customTypes))
	}
	if !b.customTypes["uuid"]("123e4567-e89b-12d3-a456-426614174000") {
		t.Error("expected uuid check to pass for valid uuid")
	}
	if b.customTypes["uuid"]("short") {
		t.Error("expected uuid check to fail for invalid value")
	}
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		ext     string
		wantErr bool
	}{
		{"valid yaml", "name: myproject\nverbose: true\n", ".yaml", false},
		{"valid yml", "name: myproject\n", ".yml", false},
		{"valid json", `{"name": "myproject"}`, ".json", false},
		{"unknown ext treated as json", `{"name": "myproject"}`, ".conf", false},
		{"missing required yaml", "verbose: true\n", ".yaml", true},
		{"bad yaml", "name: [unclosed\n", ".yaml", true},
		{"bad json", `{"name": "myproject"`, ".json", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config"+tt.ext)
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("write config: %v", err)
			}
			errs, err := ValidateFile(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no errors, got %v", errs)
			}
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors")
			}
		})
	}
}

func TestValidateFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")
	_, err := ValidateFile(path)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateData(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		format string
		want   int
	}{
		{"valid yaml", "name: myproject\n", "yaml", 0},
		{"valid yml alias", "name: myproject\n", "yml", 0},
		{"yaml missing required", "verbose: true\n", "yaml", 1},
		{"invalid yaml", "name: [broken\n", "yaml", 1},
		{"valid json", `{"name": "myproject"}`, "json", 0},
		{"invalid json", `{"name":`, "json", 1},
		{"json missing required", `{"verbose": true}`, "json", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := ValidateData([]byte(tt.data), tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(errs) != tt.want {
				t.Errorf("expected %d errors, got %d (%v)", tt.want, len(errs), errs)
			}
		})
	}
}

func TestValidateDataUnknownFormatFallsBackToJSON(t *testing.T) {
	errs, err := ValidateData([]byte(`{"name": "myproject"}`), "toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no errors for unknown format, got %v", errs)
	}
}

func TestLoadSchemaFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		ext     string
		wantErr bool
	}{
		{"yaml schema", "type: object\nproperties:\n  name:\n    type: string\n", ".yaml", false},
		{"yml schema", "type: object\n", ".yml", false},
		{"json schema", `{"type": "object", "properties": {"name": {"type": "string"}}}`, ".json", false},
		{"unknown ext json", `{"type": "object"}`, ".cfg", false},
		{"bad yaml", "type: [broken\n", ".yaml", true},
		{"bad json", `{"type":`, ".json", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "schema"+tt.ext)
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("write schema: %v", err)
			}
			schema, err := LoadSchemaFromFile(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if schema == nil || schema.Type != "object" {
				t.Errorf("expected object schema, got %+v", schema)
			}
		})
	}
}

func TestLoadSchemaFromFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	_, err := LoadSchemaFromFile(path)
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
}
