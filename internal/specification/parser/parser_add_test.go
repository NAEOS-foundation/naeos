package parser

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestVersionCompatWarningValid(t *testing.T) {
	warning := VersionCompatWarning("0.1.0")
	if warning != "" {
		t.Errorf("expected empty warning, got %q", warning)
	}
}

func TestVersionCompatWarningEmpty(t *testing.T) {
	warning := VersionCompatWarning("")
	if warning != "" {
		t.Errorf("expected empty warning, got %q", warning)
	}
}

func TestVersionCompatWarningTooLow(t *testing.T) {
	warning := VersionCompatWarning("0.0.1")
	if warning == "" {
		t.Error("expected warning for version below minimum")
	}
}

func TestVersionCompatWarningInvalid(t *testing.T) {
	warning := VersionCompatWarning("invalid")
	if warning == "" {
		t.Error("expected warning for invalid version string")
	}
}

func scalarNode(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}

func TestParseYAMLScalarBool(t *testing.T) {
	result, err := parseYAMLScalar(scalarNode("true"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestParseYAMLScalarNull(t *testing.T) {
	result, err := parseYAMLScalar(scalarNode("~"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestParseYAMLScalarNullWord(t *testing.T) {
	result, err := parseYAMLScalar(scalarNode("null"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestParseYAMLScalarInteger(t *testing.T) {
	result, err := parseYAMLScalar(scalarNode("42"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", result)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestParseYAMLScalarFloat(t *testing.T) {
	result, err := parseYAMLScalar(scalarNode("3.14"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := result.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", result)
	}
	if v != 3.14 {
		t.Errorf("expected 3.14, got %f", v)
	}
}

func TestParseYAMLScalarString(t *testing.T) {
	result, err := parseYAMLScalar(scalarNode("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %v", result)
	}
}

func TestParseSchemaVersionNegative(t *testing.T) {
	_, err := ParseSchemaVersion("-1.0.0")
	if err == nil {
		t.Error("expected error for negative version")
	}
}

func TestParseSchemaVersionVPrefixOnly(t *testing.T) {
	_, err := ParseSchemaVersion("v")
	if err == nil {
		t.Error("expected error for v prefix only")
	}
}

func TestCheckSpecVersionTooLow(t *testing.T) {
	result := CheckSpecVersion("0.0.1")
	if result.Valid {
		t.Error("expected invalid for too-low version")
	}
}
