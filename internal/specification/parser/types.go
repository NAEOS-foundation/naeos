package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Type System v3

type SpecType string

const (
	TypeString  SpecType = "string"
	TypeInteger SpecType = "integer"
	TypeFloat   SpecType = "float"
	TypeBoolean SpecType = "boolean"
	TypeArray   SpecType = "array"
	TypeObject  SpecType = "object"
	TypeRef     SpecType = "ref"
	TypeUnion   SpecType = "union"
)

type TypeDefinition struct {
	Name     string
	Type     SpecType
	Required bool
	Default  any
	Constraints []Constraint
	Items    *TypeDefinition
	Properties map[string]*TypeDefinition
	Ref      string
	Union    []*TypeDefinition
}

type Constraint struct {
	Type  string
	Value any
}

type ValidationRule struct {
	Name     string
	Field    string
	Operator string
	Value    any
	Message  string
}

type ValidationSchema struct {
	Types      map[string]*TypeDefinition
	Rules      []*ValidationRule
	Required   []string
	Unique     []string
	Custom     []*CustomValidator
}

type CustomValidator struct {
	Name   string
	Func   string
	Params map[string]any
}

// Type Registry

type TypeRegistry struct {
	types map[string]*TypeDefinition
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]*TypeDefinition),
	}
}

func (r *TypeRegistry) Register(name string, def *TypeDefinition) {
	r.types[name] = def
}

func (r *TypeRegistry) Get(name string) (*TypeDefinition, bool) {
	def, ok := r.types[name]
	return def, ok
}

func (r *TypeRegistry) Resolve(ref string) (*TypeDefinition, error) {
	if def, ok := r.types[ref]; ok {
		return def, nil
	}
	return nil, fmt.Errorf("type not found: %s", ref)
}

// Type Builder

type TypeBuilder struct {
	name string
	def  *TypeDefinition
}

func NewType(name string, t SpecType) *TypeBuilder {
	return &TypeBuilder{
		name: name,
		def: &TypeDefinition{
			Name: name,
			Type: t,
		},
	}
}

func (b *TypeBuilder) Required() *TypeBuilder {
	b.def.Required = true
	return b
}

func (b *TypeBuilder) Default(val any) *TypeBuilder {
	b.def.Default = val
	return b
}

func (b *TypeBuilder) Min(min int) *TypeBuilder {
	b.def.Constraints = append(b.def.Constraints, Constraint{Type: "min", Value: min})
	return b
}

func (b *TypeBuilder) Max(max int) *TypeBuilder {
	b.def.Constraints = append(b.def.Constraints, Constraint{Type: "max", Value: max})
	return b
}

func (b *TypeBuilder) Pattern(pattern string) *TypeBuilder {
	b.def.Constraints = append(b.def.Constraints, Constraint{Type: "pattern", Value: pattern})
	return b
}

func (b *TypeBuilder) OneOf(values ...string) *TypeBuilder {
	b.def.Constraints = append(b.def.Constraints, Constraint{Type: "oneOf", Value: values})
	return b
}

func (b *TypeBuilder) Items(items *TypeBuilder) *TypeBuilder {
	b.def.Items = items.Build()
	return b
}

func (b *TypeBuilder) Union(types ...*TypeBuilder) *TypeBuilder {
	for _, t := range types {
		b.def.Union = append(b.def.Union, t.Build())
	}
	return b
}

func (b *TypeBuilder) Properties(props map[string]*TypeBuilder) *TypeBuilder {
	b.def.Properties = make(map[string]*TypeDefinition)
	for k, v := range props {
		b.def.Properties[k] = v.Build()
	}
	return b
}

func (b *TypeBuilder) Ref(ref string) *TypeBuilder {
	b.def.Ref = ref
	return b
}

func (b *TypeBuilder) Build() *TypeDefinition {
	return b.def
}

// Validation Engine

type ValidationEngine struct {
	schema *ValidationSchema
}

func NewValidationEngine(schema *ValidationSchema) *ValidationEngine {
	return &ValidationEngine{schema: schema}
}

func (v *ValidationEngine) Validate(data map[string]any) []ValidationError {
	var errors []ValidationError

	// Check required fields
	for _, field := range v.schema.Required {
		if _, ok := data[field]; !ok {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("field '%s' is required", field),
			})
		}
	}

	// Check unique fields
	seen := make(map[any]bool)
	for _, field := range v.schema.Unique {
		val, ok := data[field]
		if !ok {
			continue
		}
		if seen[val] {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("field '%s' must be unique", field),
			})
		}
		seen[val] = true
	}

	// Validate types
	for name, def := range v.schema.Types {
		val, ok := data[name]
		if !ok {
			if def.Required {
				errors = append(errors, ValidationError{
					Field:   name,
					Message: fmt.Sprintf("field '%s' is required", name),
				})
			}
			continue
		}

		if err := v.validateType(val, def); err != nil {
			errors = append(errors, ValidationError{
				Field:   name,
				Message: err.Error(),
			})
		}
	}

	// Apply custom rules
	for _, rule := range v.schema.Rules {
		if err := v.applyRule(data, rule); err != nil {
			errors = append(errors, ValidationError{
				Field:   rule.Field,
				Message: rule.Message,
			})
		}
	}

	return errors
}

func (v *ValidationEngine) validateType(val any, def *TypeDefinition) error {
	switch def.Type {
	case TypeString:
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
		return v.validateString(s, def)
	case TypeInteger:
		// Accept both int and float64 (YAML/JSON numbers)
		switch val.(type) {
		case int, int64, float64:
			return v.validateNumber(val, def)
		default:
			return fmt.Errorf("expected integer, got %T", val)
		}
	case TypeFloat:
		switch val.(type) {
		case float64, int, int64:
			return nil
		default:
			return fmt.Errorf("expected float, got %T", val)
		}
	case TypeBoolean:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", val)
		}
		return nil
	case TypeArray:
		arr, ok := val.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", val)
		}
		if def.Items != nil {
			for i, item := range arr {
				if err := v.validateType(item, def.Items); err != nil {
					return fmt.Errorf("item %d: %w", i, err)
				}
			}
		}
		return nil
	case TypeObject:
		obj, ok := val.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object, got %T", val)
		}
		if def.Properties != nil {
			for name, propDef := range def.Properties {
				if propVal, ok := obj[name]; ok {
					if err := v.validateType(propVal, propDef); err != nil {
						return fmt.Errorf("property '%s': %w", name, err)
					}
				} else if propDef.Required {
					return fmt.Errorf("property '%s' is required", name)
				}
			}
		}
		return nil
	case TypeRef:
		if def.Ref == "" {
			return fmt.Errorf("ref type must have a reference")
		}
		return nil
	case TypeUnion:
		for _, unionType := range def.Union {
			if err := v.validateType(val, unionType); err == nil {
				return nil
			}
		}
		return fmt.Errorf("value does not match any type in union")
	}

	return nil
}

func (v *ValidationEngine) validateString(s string, def *TypeDefinition) error {
	for _, c := range def.Constraints {
		switch c.Type {
		case "pattern":
			pattern, ok := c.Value.(string)
			if !ok {
				continue
			}
			matched, err := regexp.MatchString(pattern, s)
			if err != nil {
				return fmt.Errorf("invalid pattern: %s", pattern)
			}
			if !matched {
				return fmt.Errorf("string does not match pattern: %s", pattern)
			}
		case "oneOf":
			values, ok := c.Value.([]string)
			if !ok {
				continue
			}
			found := false
			for _, v := range values {
				if s == v {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("string must be one of: %s", strings.Join(values, ", "))
			}
		case "min":
			min, ok := c.Value.(int)
			if ok && len(s) < min {
				return fmt.Errorf("string length must be at least %d", min)
			}
		case "max":
			max, ok := c.Value.(int)
			if ok && len(s) > max {
				return fmt.Errorf("string length must be at most %d", max)
			}
		}
	}
	return nil
}

func (v *ValidationEngine) validateNumber(val any, def *TypeDefinition) error {
	var num float64
	switch n := val.(type) {
	case int:
		num = float64(n)
	case int64:
		num = float64(n)
	case float64:
		num = n
	default:
		return fmt.Errorf("expected number, got %T", val)
	}

	for _, c := range def.Constraints {
		switch c.Type {
		case "min":
			if min, ok := c.Value.(int); ok && num < float64(min) {
				return fmt.Errorf("number must be at least %d", min)
			}
		case "max":
			if max, ok := c.Value.(int); ok && num > float64(max) {
				return fmt.Errorf("number must be at most %d", max)
			}
		}
	}
	return nil
}

func (v *ValidationEngine) applyRule(data map[string]any, rule *ValidationRule) error {
	val, ok := data[rule.Field]
	if !ok {
		return nil
	}

	switch rule.Operator {
	case "equals":
		if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", rule.Value) {
			return fmt.Errorf("field '%s' must equal %v", rule.Field, rule.Value)
		}
	case "notEquals":
		if fmt.Sprintf("%v", val) == fmt.Sprintf("%v", rule.Value) {
			return fmt.Errorf("field '%s' must not equal %v", rule.Field, rule.Value)
		}
	case "contains":
		s, ok := val.(string)
		if ok {
			substr, ok := rule.Value.(string)
			if ok && !strings.Contains(s, substr) {
				return fmt.Errorf("field '%s' must contain '%s'", rule.Field, substr)
			}
		}
	case "matches":
		s, ok := val.(string)
		if ok {
			pattern, ok := rule.Value.(string)
			if ok {
				matched, _ := regexp.MatchString(pattern, s)
				if !matched {
					return fmt.Errorf("field '%s' must match pattern '%s'", rule.Field, pattern)
				}
			}
		}
	}

	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
