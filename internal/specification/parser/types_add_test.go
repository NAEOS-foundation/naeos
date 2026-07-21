package parser

import (
	"testing"
)

func TestValidationEngineFloatAcceptsFloat(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"rate": {Type: TypeFloat},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"rate": 3.14})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineFloatAcceptsInt(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"count": {Type: TypeFloat},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"count": 42})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineFloatRejectsString(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"rate": {Type: TypeFloat},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"rate": "not-a-float"})
	if len(errs) == 0 {
		t.Error("expected error for string")
	}
}

func TestValidationEngineRefValid(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"ext": {Type: TypeRef, Ref: "some.type"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"ext": "ignored"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineRefEmptyRef(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"ext": {Type: TypeRef, Ref: ""},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"ext": "val"})
	if len(errs) == 0 {
		t.Error("expected error for empty ref")
	}
}

func TestValidationEngineObjectWithProperties(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"config": {
				Type: TypeObject,
				Properties: map[string]*TypeDefinition{
					"name": {Type: TypeString},
					"age":  {Type: TypeInteger},
				},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{
		"config": map[string]any{
			"name": "test",
			"age":  30,
		},
	})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineObjectPropertyWrongType(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"config": {
				Type: TypeObject,
				Properties: map[string]*TypeDefinition{
					"age": {Type: TypeInteger},
				},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{
		"config": map[string]any{
			"age": "not-a-number",
		},
	})
	if len(errs) == 0 {
		t.Error("expected error for wrong property type")
	}
}

func TestValidationEngineObjectMissingRequiredProperty(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"config": {
				Type: TypeObject,
				Properties: map[string]*TypeDefinition{
					"name": {Type: TypeString, Required: true},
				},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{
		"config": map[string]any{},
	})
	if len(errs) == 0 {
		t.Error("expected error for missing required property")
	}
}

func TestValidationEngineObjectRejectsNonMap(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"config": {Type: TypeObject},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"config": "string"})
	if len(errs) == 0 {
		t.Error("expected error for non-map")
	}
}

func TestValidationEngineStringMinLength(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"name": {
				Type:        TypeString,
				Constraints: []Constraint{{Type: "min", Value: 3}},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"name": "ab"})
	if len(errs) == 0 {
		t.Error("expected error for short string")
	}
}

func TestValidationEngineStringMaxLength(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"name": {
				Type:        TypeString,
				Constraints: []Constraint{{Type: "max", Value: 5}},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"name": "too long"})
	if len(errs) == 0 {
		t.Error("expected error for long string")
	}
}

func TestValidationEngineStringMinMaxValid(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"name": {
				Type:        TypeString,
				Constraints: []Constraint{{Type: "min", Value: 2}, {Type: "max", Value: 10}},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"name": "hello"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineNumberMinConstraint(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"age": {
				Type:        TypeInteger,
				Constraints: []Constraint{{Type: "min", Value: 18}},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"age": 15})
	if len(errs) == 0 {
		t.Error("expected error for under min")
	}
}

func TestValidationEngineNumberMaxConstraint(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"age": {
				Type:        TypeInteger,
				Constraints: []Constraint{{Type: "max", Value: 150}},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"age": 200})
	if len(errs) == 0 {
		t.Error("expected error for over max")
	}
}

func TestValidationEngineUniqueValid(t *testing.T) {
	schema := &ValidationSchema{
		Unique: []string{"email"},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"email": "a@b.com"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineUniqueDuplicate(t *testing.T) {
	schema := &ValidationSchema{
		Unique: []string{"email"},
	}
	_ = NewValidationEngine(schema)
}

func TestValidationEngineRuleEquals(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "status", Operator: "equals", Value: "active", Message: "must be active"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"status": "inactive"})
	if len(errs) == 0 {
		t.Error("expected error for not equals")
	}
}

func TestValidationEngineRuleEqualsPass(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "status", Operator: "equals", Value: "active", Message: "must be active"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"status": "active"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineRuleNotEquals(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "env", Operator: "notEquals", Value: "prod", Message: "not in prod"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"env": "prod"})
	if len(errs) == 0 {
		t.Error("expected error for notEquals with matching value")
	}
}

func TestValidationEngineRuleContains(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "description", Operator: "contains", Value: "error", Message: "must contain error"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"description": "all good"})
	if len(errs) == 0 {
		t.Error("expected error for not containing substring")
	}
}

func TestValidationEngineRuleContainsPass(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "description", Operator: "contains", Value: "error", Message: "msg"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"description": "something error occurred"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineRuleMatches(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "email", Operator: "matches", Value: `^\S+@\S+$`, Message: "invalid email"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"email": "not-an-email"})
	if len(errs) == 0 {
		t.Error("expected error for not matching pattern")
	}
}

func TestValidationEngineRuleMatchesPass(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "email", Operator: "matches", Value: `^\S+@\S+$`, Message: "invalid"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"email": "a@b.com"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineRuleFieldNotFound(t *testing.T) {
	schema := &ValidationSchema{
		Rules: []*ValidationRule{
			{Field: "missing", Operator: "equals", Value: "x", Message: "irrelevant"},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"other": "y"})
	if len(errs) != 0 {
		t.Errorf("expected no errors for missing field, got %v", errs)
	}
}

func TestValidationEngineMissingOptionalField(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"name": {Type: TypeString},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineRequiredFieldMissing(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"name": {Type: TypeString, Required: true},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{})
	if len(errs) == 0 {
		t.Error("expected error for missing required field")
	}
}

func TestValidationEngineRequiredFromList(t *testing.T) {
	schema := &ValidationSchema{
		Required: []string{"email"},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{})
	if len(errs) == 0 {
		t.Error("expected error for missing required field from list")
	}
}

func TestValidationErrorError(t *testing.T) {
	e := ValidationError{Field: "name", Message: "is required"}
	if e.Error() != "name: is required" {
		t.Errorf("unexpected Error() output: %s", e.Error())
	}
}

func TestValidationEngineArrayWithoutItems(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"tags": {Type: TypeArray},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"tags": []any{"a", "b", 1}})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationEngineArrayNonArray(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"tags": {Type: TypeArray},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"tags": "not-array"})
	if len(errs) == 0 {
		t.Error("expected error for non-array")
	}
}

func TestValidationEngineIntegerRejectsString(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"count": {Type: TypeInteger},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"count": "not-int"})
	if len(errs) == 0 {
		t.Error("expected error for string instead of integer")
	}
}

func TestValidationEngineBooleanRejectsString(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"flag": {Type: TypeBoolean},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"flag": "true"})
	if len(errs) == 0 {
		t.Error("expected error for string instead of boolean")
	}
}

func TestValidationEngineStringInvalidPattern(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"field": {
				Type:        TypeString,
				Constraints: []Constraint{{Type: "pattern", Value: "[invalid"}},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"field": "test"})
	if len(errs) == 0 {
		t.Error("expected error for invalid pattern")
	}
}

func TestValidationEngineStringConstraintWrongType(t *testing.T) {
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{
			"field": {
				Type:        TypeString,
				Constraints: []Constraint{{Type: "pattern", Value: 42}},
			},
		},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"field": "test"})
	if len(errs) != 0 {
		t.Errorf("expected no error for non-string pattern value, got %v", errs)
	}
}

func TestTypeRegistryResolveNotFound(t *testing.T) {
	r := NewTypeRegistry()
	_, err := r.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestTypeBuilderOneOf(t *testing.T) {
	def := NewType("color", TypeString).OneOf("red", "green", "blue").Build()
	schema := &ValidationSchema{
		Types: map[string]*TypeDefinition{"color": def},
	}
	v := NewValidationEngine(schema)
	errs := v.Validate(map[string]any{"color": "yellow"})
	if len(errs) == 0 {
		t.Error("expected error for not in oneOf")
	}
}

func TestTypeBuilderef(t *testing.T) {
	def := NewType("ext", TypeRef).Ref("my.type").Build()
	if def.Ref != "my.type" {
		t.Errorf("expected my.type, got %s", def.Ref)
	}
}

func TestTypeBuilderDefault(t *testing.T) {
	def := NewType("port", TypeInteger).Default(8080).Build()
	if def.Default != 8080 {
		t.Errorf("expected 8080, got %v", def.Default)
	}
}

func TestTypeBuilderProperties(t *testing.T) {
	def := NewType("config", TypeObject).Properties(map[string]*TypeBuilder{
		"host": NewType("host", TypeString),
		"port": NewType("port", TypeInteger),
	}).Build()
	if len(def.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(def.Properties))
	}
}

func TestTypeBuilderUnion(t *testing.T) {
	def := NewType("id", TypeUnion).Union(
		NewType("id_str", TypeString),
		NewType("id_int", TypeInteger),
	).Build()
	if len(def.Union) != 2 {
		t.Errorf("expected 2 union types, got %d", len(def.Union))
	}
}
