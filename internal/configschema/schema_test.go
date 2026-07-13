package configschema

import (
	"encoding/json"
	"testing"
)

func TestValidateConfigRequired(t *testing.T) {
	config := map[string]interface{}{
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
	config := map[string]interface{}{
		"name":    123,
		"verbose": "notbool",
	}
	errs := ValidateConfig(config)
	if len(errs) < 2 {
		t.Errorf("expected >=2 type errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateConfigValid(t *testing.T) {
	config := map[string]interface{}{
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
	data, _ := json.Marshal(map[string]interface{}{
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
