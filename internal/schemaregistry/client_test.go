package schemaregistry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateNEIRSpecValid(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(spec, []byte("project: my-project\nmodules:\n  - name: core\n    path: ./internal/core\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	schema := loadTestSchema(t)
	result, err := ValidateNEIRSpec(spec, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateNEIRSpecMissingRequired(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(spec, []byte("services:\n  - name: api\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	schema := loadTestSchema(t)
	result, err := ValidateNEIRSpec(spec, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid due to missing required fields")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestValidateNEIRSpecInvalidEnum(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(spec, []byte("project: test\nmodules:\n  - name: core\n    path: ./internal/core\narchitecture:\n  pattern: invalid-pattern\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	schema := loadTestSchema(t)
	result, err := ValidateNEIRSpec(spec, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid due to invalid enum")
	}
}

func TestValidateNEIRSpecJSON(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(spec, []byte(`{"project":"test","modules":[{"name":"core","path":"./internal/core"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	schema := loadTestSchema(t)
	result, err := ValidateNEIRSpec(spec, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateNEIRSpecWithSchemaRef(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(spec, []byte("$schema: https://naeos.dev/schemaregistry/latest.json\nproject: test\nmodules:\n  - name: core\n    path: ./internal/core\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	schema := loadTestSchema(t)
	result, err := ValidateNEIRSpec(spec, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateNEIRSpecNotExist(t *testing.T) {
	_, err := ValidateNEIRSpec("/nonexistent/path.yaml", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func loadTestSchema(t *testing.T) map[string]any {
	t.Helper()
	c := NewNEIRClient("file://../../site/static/schemaregistry/latest.json")
	schema, err := c.FetchSchema()
	if err != nil {
		t.Fatalf("load test schema: %v", err)
	}
	return schema
}

func TestNEIRClientFetchSchema(t *testing.T) {
	c := NewNEIRClient("file://../../site/static/schemaregistry/latest.json")
	schema, err := c.FetchSchema()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema["title"] != "NEIR Specification" {
		t.Errorf("expected 'NEIR Specification', got %v", schema["title"])
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties to be a map")
	}
	if _, ok := props["project"]; !ok {
		t.Error("expected project property")
	}
	if _, ok := props["modules"]; !ok {
		t.Error("expected modules property")
	}
}

func TestNEIRValidationResultJSON(t *testing.T) {
	r := NEIRValidationResult{
		Valid:   false,
		Version: "v1",
		Errors: []NEIRValidationError{
			{Field: "project", Message: "required field 'project' is missing"},
		},
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded NEIRValidationResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Valid != r.Valid {
		t.Errorf("expected valid=%v, got %v", r.Valid, decoded.Valid)
	}
	if len(decoded.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(decoded.Errors))
	}
	if decoded.Errors[0].Field != "project" {
		t.Errorf("expected field 'project', got %q", decoded.Errors[0].Field)
	}
}
