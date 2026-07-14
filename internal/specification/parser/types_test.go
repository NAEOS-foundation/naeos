package parser

import (
	"testing"
)

func TestTypeRegistry(t *testing.T) {
	reg := NewTypeRegistry()

	def := NewType("name", TypeString).Required().Build()
	reg.Register("name", def)

	got, ok := reg.Get("name")
	if !ok {
		t.Fatal("expected type to be found")
	}
	if got.Name != "name" {
		t.Errorf("expected name 'name', got %s", got.Name)
	}
	if got.Type != TypeString {
		t.Errorf("expected type string, got %s", got.Type)
	}
}

func TestTypeRegistryResolve(t *testing.T) {
	reg := NewTypeRegistry()
	reg.Register("name", NewType("name", TypeString).Build())

	def, err := reg.Resolve("name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "name" {
		t.Errorf("expected name 'name', got %s", def.Name)
	}

	_, err = reg.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent type")
	}
}

func TestTypeBuilder(t *testing.T) {
	def := NewType("age", TypeInteger).
		Required().
		Min(0).
		Max(150).
		Build()

	if def.Name != "age" {
		t.Errorf("expected name 'age', got %s", def.Name)
	}
	if def.Type != TypeInteger {
		t.Errorf("expected type integer, got %s", def.Type)
	}
	if !def.Required {
		t.Error("expected required to be true")
	}
	if len(def.Constraints) != 2 {
		t.Errorf("expected 2 constraints, got %d", len(def.Constraints))
	}
}

func TestValidationEngineRequired(t *testing.T) {
	schema := &ValidationSchema{
		Required: []string{"name", "email"},
		Types: map[string]*TypeDefinition{
			"name":  NewType("name", TypeString).Required().Build(),
			"email": NewType("email", TypeString).Required().Build(),
		},
	}

	engine := NewValidationEngine(schema)

	// Missing required fields (Required list + type Required)
	var data map[string]any
	var errors []ValidationError

	data = map[string]any{
		"age": 25,
	}
	errors = engine.Validate(data)
	if len(errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errors))
	}

	// All required fields present
	data = map[string]any{
		"name":  "John",
		"email": "john@example.com",
	}
	errors = engine.Validate(data)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestValidationEngineTypes(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"name":    NewType("name", TypeString).Required().Build(),
			"age":     NewType("age", TypeInteger).Build(),
			"active":  NewType("active", TypeBoolean).Build(),
			"tags":    NewType("tags", TypeArray).Items(NewType("item", TypeString)).Build(),
		},
	}

	engine := NewValidationEngine(schema)

	data := map[string]any{
		"name":   "John",
		"age":    25,
		"active": true,
		"tags":   []any{"admin", "user"},
	}
	errors := engine.Validate(data)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}

	// Wrong types
	data = map[string]any{
		"name":   123,
		"age":    "not a number",
		"active": "not a bool",
	}
	errors = engine.Validate(data)
	if len(errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errors))
	}
}

func TestValidationEnginePattern(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"email": NewType("email", TypeString).Pattern(`^[a-z]+@[a-z]+\.[a-z]+$`).Build(),
		},
	}

	engine := NewValidationEngine(schema)

	data := map[string]any{
		"email": "john@example.com",
	}
	errors := engine.Validate(data)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}

	data = map[string]any{
		"email": "not-an-email",
	}
	errors = engine.Validate(data)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestValidationEngineOneOf(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"status": NewType("status", TypeString).OneOf("active", "inactive", "pending").Build(),
		},
	}

	engine := NewValidationEngine(schema)

	data := map[string]any{
		"status": "active",
	}
	errors := engine.Validate(data)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}

	data = map[string]any{
		"status": "unknown",
	}
	errors = engine.Validate(data)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestValidationEngineRules(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"port": NewType("port", TypeInteger).Build(),
		},
		Rules: []*ValidationRule{
			{
				Name:     "valid-port",
				Field:    "port",
				Operator: "min",
				Value:    1,
				Message:  "port must be at least 1",
			},
		},
	}

	engine := NewValidationEngine(schema)

	data := map[string]any{
		"port": 8080,
	}
	errors := engine.Validate(data)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestValidationEngineUnion(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"value": NewType("value", TypeUnion).Union(
				NewType("str", TypeString),
				NewType("num", TypeInteger),
			).Build(),
		},
	}

	engine := NewValidationEngine(schema)

	// String value
	data := map[string]any{
		"value": "hello",
	}
	errors := engine.Validate(data)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}

	// Integer value
	data = map[string]any{
		"value": 123,
	}
	errors = engine.Validate(data)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestSchemaVersionV3(t *testing.T) {
	result := CheckSpecVersion("0.3.0")
	if !result.Valid {
		t.Errorf("expected valid, got: %s", result.Message)
	}
}
