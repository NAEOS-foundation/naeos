package configschema

import (
	"encoding/json"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/testutil"
)

func TestDefaultSchema(t *testing.T) {
	schema := DefaultSchema()
	if schema.Type != "object" {
		t.Errorf("expected type object, got %s", schema.Type)
	}
	if len(schema.Properties) == 0 {
		t.Error("expected properties in default schema")
	}
	if len(schema.Required) == 0 || schema.Required[0] != "name" {
		t.Error("expected 'name' in required")
	}
}

func TestSchemaBuilder(t *testing.T) {
	schema := NewBuilder().
		Title("Test Schema").
		Description("A test").
		Version("2.0").
		AddProperty("field1", Property{Type: "string", Description: "test field"}).
		AddProperty("field2", Property{Type: "number", Minimum: float64Ptr(0), Maximum: float64Ptr(100)}).
		Required("field1").
		Build()
	if schema.Title != "Test Schema" {
		t.Errorf("expected title Test Schema, got %s", schema.Title)
	}
	if schema.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", schema.Version)
	}
	if _, ok := schema.Properties["field1"]; !ok {
		t.Error("expected field1 in properties")
	}
	if len(schema.Required) != 1 || schema.Required[0] != "field1" {
		t.Error("expected field1 in required")
	}
}

func TestValidateConfig(t *testing.T) {
	config := map[string]any{
		"name":    "myproject",
		"verbose": true,
	}
	errs := ValidateConfig(config)
	for _, e := range errs {
		t.Errorf("unexpected error: %s: %s", e.Field, e.Message)
	}
}

func TestValidateConfigMissingRequired(t *testing.T) {
	config := map[string]any{
		"verbose": true,
	}
	errs := ValidateConfig(config)
	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
		}
	}
	if !found {
		t.Error("expected missing required field 'name' error")
	}
}

func TestValidateTypeEnum(t *testing.T) {
	schema := NewBuilder().
		AddProperty("mode", Property{Type: "string", Enum: []any{"a", "b", "c"}}).
		Build()
	config := map[string]any{"mode": "d"}
	errs := ValidateWithSchema(config, schema)
	found := false
	for _, e := range errs {
		if e.Field == "mode" {
			found = true
		}
	}
	if !found {
		t.Error("expected enum validation error")
	}
	config["mode"] = "a"
	errs = ValidateWithSchema(config, schema)
	for _, e := range errs {
		t.Errorf("unexpected error: %s", e.Message)
	}
}

func TestValidateStringConstraints(t *testing.T) {
	schema := NewBuilder().
		AddProperty("name", Property{
			Type:      "string",
			MinLength: intPtr(3),
			MaxLength: intPtr(10),
			Pattern:   `^[a-z]+$`,
		}).
		Build()
	tests := []struct {
		value    string
		wantErrs int
	}{
		{"ab", 1},
		{"abcdefghjkl", 1},
		{"abc123", 1},
		{"abc", 0},
	}
	for _, tt := range tests {
		config := map[string]any{"name": tt.value}
		errs := ValidateWithSchema(config, schema)
		strErrs := 0
		for _, e := range errs {
			if e.Field == "name" {
				strErrs++
			}
		}
		if strErrs != tt.wantErrs {
			t.Errorf("value %s: expected %d errors, got %d", tt.value, tt.wantErrs, strErrs)
		}
	}
}

func TestValidateNumberConstraints(t *testing.T) {
	schema := NewBuilder().
		AddProperty("port", Property{
			Type:    "number",
			Minimum: float64Ptr(1),
			Maximum: float64Ptr(65535),
		}).
		Build()
	tests := []struct {
		value    any
		wantErrs int
	}{
		{0, 1},
		{99999, 1},
		{8080, 0},
	}
	for _, tt := range tests {
		config := map[string]any{"port": tt.value}
		errs := ValidateWithSchema(config, schema)
		numErrs := 0
		for _, e := range errs {
			if e.Field == "port" {
				numErrs++
			}
		}
		if numErrs != tt.wantErrs {
			t.Errorf("value %v: expected %d errors, got %d", tt.value, tt.wantErrs, numErrs)
		}
	}
}

func TestValidateArrayItems(t *testing.T) {
	schema := NewBuilder().
		AddProperty("tags", Property{
			Type:  "array",
			Items: &Property{Type: "string"},
		}).
		Build()
	config := map[string]any{"tags": []any{"a", "b", 123}}
	errs := ValidateWithSchema(config, schema)
	itemErrs := 0
	for _, e := range errs {
		if e.Field == "tags[2]" {
			itemErrs++
		}
	}
	if itemErrs != 1 {
		t.Errorf("expected 1 item error, got %d", itemErrs)
	}
}

func TestValidateNestedObject(t *testing.T) {
	schema := NewBuilder().
		AddProperty("server", Property{
			Type: "object",
			Properties: map[string]Property{
				"host": {Type: "string"},
				"port": {Type: "number", Minimum: float64Ptr(1)},
			},
			Required: []string{"host"},
		}).
		Build()
	config := map[string]any{
		"server": map[string]any{"port": 8080},
	}
	errs := ValidateWithSchema(config, schema)
	found := false
	for _, e := range errs {
		if e.Field == "server.host" {
			found = true
		}
	}
	if !found {
		t.Error("expected nested required field error")
	}
}

func TestToJSONSchema(t *testing.T) {
	schema := DefaultSchema()
	jsonSchema := schema.ToJSONSchema()
	if jsonSchema["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Error("expected json-schema draft-07")
	}
	props, ok := jsonSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties map")
	}
	nameProp, ok := props["name"].(map[string]any)
	if !ok {
		t.Fatal("expected name property")
	}
	if nameProp["type"] != "string" {
		t.Errorf("expected type string, got %v", nameProp["type"])
	}
	req, ok := jsonSchema["required"].([]string)
	if !ok || len(req) == 0 || req[0] != "name" {
		t.Error("expected required field name")
	}
	_, err := json.Marshal(jsonSchema)
	if err != nil {
		t.Errorf("failed to marshal JSON schema: %v", err)
	}
}

func TestGenerateDocumentation(t *testing.T) {
	schema := DefaultSchema()
	doc := schema.GenerateDocumentation()
	if len(doc) == 0 {
		t.Error("expected non-empty documentation")
	}
	if !testutil.Contains(doc, "## Required Fields") {
		t.Error("expected Required Fields section")
	}
	if !testutil.Contains(doc, "### `name`") {
		t.Error("expected name field docs")
	}
}

func TestDeprecatedField(t *testing.T) {
	schema := NewBuilder().
		AddProperty("old_field", Property{Type: "string", Deprecated: true}).
		Build()
	config := map[string]any{"old_field": "value"}
	errs := ValidateWithSchema(config, schema)
	found := false
	for _, e := range errs {
		if e.Field == "old_field" && testutil.Contains(e.Message, "deprecated") {
			found = true
		}
	}
	if !found {
		t.Error("expected deprecation warning")
	}
}
