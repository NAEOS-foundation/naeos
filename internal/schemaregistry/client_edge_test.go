package schemaregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewNEIRClientDefaultURL(t *testing.T) {
	c := NewNEIRClient("")
	if c.registryURL != DefaultNEIRSchemaURL {
		t.Errorf("expected default URL, got %q", c.registryURL)
	}
	if c.httpClient == nil || c.httpClient.Timeout <= 0 {
		t.Error("expected configured http client")
	}
}

func TestFetchSchemaVersionHTTP(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"object","properties":{"name":{"type":"string"}}}`)
	}))
	defer srv.Close()

	c := NewNEIRClient(srv.URL + "/schemas/latest.json")
	schema, err := c.FetchSchemaVersion("v2")
	if err != nil {
		t.Fatalf("fetch version: %v", err)
	}
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
	if gotPath != "/schemas/v2/neir.json" {
		t.Errorf("expected path /schemas/v2/neir.json, got %q", gotPath)
	}
}

func TestFetchSchemaVersionEmptyUsesBase(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, `{"type":"object"}`)
	}))
	defer srv.Close()

	c := NewNEIRClient(srv.URL + "/latest.json")
	if _, err := c.FetchSchemaVersion(""); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if gotPath != "/latest.json" {
		t.Errorf("expected base path, got %q", gotPath)
	}
}

func TestFetchSchemaVersionInvalidURL(t *testing.T) {
	c := NewNEIRClient("://bad-url")
	if _, err := c.FetchSchemaVersion("v1"); err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestFetchSchemaHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewNEIRClient(srv.URL)
	if _, err := c.FetchSchema(); err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestFetchSchemaUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := NewNEIRClient(url)
	if _, err := c.FetchSchema(); err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestFetchSchemaFileMissing(t *testing.T) {
	c := NewNEIRClient("file://" + filepath.Join(t.TempDir(), "missing.json"))
	if _, err := c.FetchSchema(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFetchSchemaFileBadJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	c := NewNEIRClient("file://" + path)
	if _, err := c.FetchSchema(); err == nil {
		t.Fatal("expected parse error for bad JSON")
	}
}

func TestValidateSpecValid(t *testing.T) {
	spec := map[string]any{
		"project": "demo",
		"modules": []any{
			map[string]any{"name": "core", "path": "./internal/core"},
		},
	}
	result := ValidateSpec(spec, loadTestSchema(t))
	if !result.Valid {
		t.Fatalf("expected valid, got %v", result.Errors)
	}
}

func TestValidateSpecMissingRequired(t *testing.T) {
	spec := map[string]any{"modules": []any{}}
	result := ValidateSpec(spec, loadTestSchema(t))
	if result.Valid {
		t.Fatal("expected invalid")
	}
	found := false
	for _, e := range result.Errors {
		if e.Field == "project" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing 'project' error, got %v", result.Errors)
	}
}

func TestValidateSpecEnumAndNested(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"mode": map[string]any{
				"type": "string",
				"enum": []any{"fast", "safe"},
			},
			"server": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"host": map[string]any{"type": "string"},
				},
				"required": []any{"host"},
			},
		},
	}

	spec := map[string]any{
		"mode":   "turbo",
		"server": map[string]any{"port": 8080},
	}
	result := ValidateSpec(spec, schema)
	if result.Valid {
		t.Fatal("expected invalid")
	}
	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(result.Errors), result.Errors)
	}

	spec["mode"] = "fast"
	spec["server"] = map[string]any{"host": "localhost"}
	result = ValidateSpec(spec, schema)
	if !result.Valid {
		t.Fatalf("expected valid after fixes, got %v", result.Errors)
	}
}

func TestValidateSpecResolveRef(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"service": map[string]any{"$ref": "#/definitions/Service"},
		},
		"definitions": map[string]any{
			"Service": map[string]any{
				"type":     "object",
				"required": []any{"name"},
			},
		},
	}

	result := ValidateSpec(map[string]any{"service": map[string]any{}}, schema)
	if result.Valid {
		t.Fatal("expected invalid for $ref target missing required field")
	}
	found := false
	for _, e := range result.Errors {
		if e.Field == "service.name" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected service.name error, got %v", result.Errors)
	}
}

func TestValidateNEIRValueStringEnum(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"mode": map[string]any{"type": "string", "enum": []any{"a"}},
		},
	}
	errs := validateNEIRValue("mode", "z", schema["properties"].(map[string]any)["mode"].(map[string]any), schema)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestNEIRValidationErrorJSON(t *testing.T) {
	e := NEIRValidationError{Field: "x", Message: "bad"}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back NEIRValidationError
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Field != "x" || back.Message != "bad" {
		t.Errorf("roundtrip mismatch: %+v", back)
	}
}
