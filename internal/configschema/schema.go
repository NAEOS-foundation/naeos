package configschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Schema struct {
	Type        string              `json:"type"`
	Properties  map[string]Property `json:"properties"`
	Required    []string            `json:"required"`
	Description string              `json:"description,omitempty"`
	Title       string              `json:"title,omitempty"`
	Version     string              `json:"version,omitempty"`
}

type Property struct {
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Default     any                 `json:"default,omitempty"`
	Enum        []any               `json:"enum,omitempty"`
	MinLength   *int                `json:"min_length,omitempty"`
	MaxLength   *int                `json:"max_length,omitempty"`
	Pattern     string              `json:"pattern,omitempty"`
	Minimum     *float64            `json:"minimum,omitempty"`
	Maximum     *float64            `json:"maximum,omitempty"`
	Items       *Property           `json:"items,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
	Deprecated  bool                `json:"deprecated,omitempty"`
	ReadOnly    bool                `json:"read_only,omitempty"`
	Examples    []any               `json:"examples,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidatorFunc func(value any, prop Property) *ValidationError

type SchemaBuilder struct {
	schema      *Schema
	validators  map[string]ValidatorFunc
	customTypes map[string]func(any) bool
}

func NewBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{
			Type:       "object",
			Properties: make(map[string]Property),
			Required:   []string{},
		},
		validators:  make(map[string]ValidatorFunc),
		customTypes: make(map[string]func(any) bool),
	}
}

func (b *SchemaBuilder) Title(title string) *SchemaBuilder {
	b.schema.Title = title
	return b
}

func (b *SchemaBuilder) Description(desc string) *SchemaBuilder {
	b.schema.Description = desc
	return b
}

func (b *SchemaBuilder) Version(version string) *SchemaBuilder {
	b.schema.Version = version
	return b
}

func (b *SchemaBuilder) AddProperty(name string, prop Property) *SchemaBuilder {
	b.schema.Properties[name] = prop
	return b
}

func (b *SchemaBuilder) Required(fields ...string) *SchemaBuilder {
	b.schema.Required = append(b.schema.Required, fields...)
	return b
}

func (b *SchemaBuilder) AddValidator(typeName string, fn ValidatorFunc) *SchemaBuilder {
	b.validators[typeName] = fn
	return b
}

func (b *SchemaBuilder) AddCustomType(typeName string, check func(any) bool) *SchemaBuilder {
	b.customTypes[typeName] = check
	return b
}

func (b *SchemaBuilder) Build() *Schema {
	return b.schema
}

func DefaultSchema() *Schema {
	return NewBuilder().
		Title("NAEOS Default Config").
		Description("Default configuration schema for NAEOS projects").
		Version("1.0.0").
		AddProperty("name", Property{Type: "string", Description: "project name", MinLength: intPtr(1), MaxLength: intPtr(128)}).
		AddProperty("version", Property{Type: "string", Description: "project version", Pattern: `^\d+\.\d+\.\d+(-\w+)?$`}).
		AddProperty("description", Property{Type: "string", Description: "project description", MaxLength: intPtr(1024)}).
		AddProperty("output_dir", Property{Type: "string", Description: "output directory", Default: "./output"}).
		AddProperty("mode", Property{Type: "string", Description: "pipeline mode", Default: "standard", Enum: []any{"standard", "parallel", "dry-run"}}).
		AddProperty("verbose", Property{Type: "boolean", Description: "verbose output", Default: false}).
		AddProperty("languages", Property{Type: "array", Description: "target languages", Items: &Property{Type: "string"}}).
		AddProperty("dry_run", Property{Type: "boolean", Description: "dry run mode", Default: false}).
		AddProperty("port", Property{Type: "number", Description: "server port", Minimum: float64Ptr(1), Maximum: float64Ptr(65535)}).
		AddProperty("host", Property{Type: "string", Description: "server host", Default: "localhost"}).
		Required("name").
		Build()
}

func ValidateFile(path string) ([]ValidationError, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrConfig, "read config")
	}
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return ValidateData(data, "yaml")
	case ".json":
		return ValidateData(data, "json")
	default:
		return ValidateData(data, "json")
	}
}

func ValidateData(data []byte, format string) ([]ValidationError, error) {
	var config map[string]any
	switch format {
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return []ValidationError{{Field: "_root", Message: "invalid YAML: " + err.Error()}}, nil
		}
	default:
		if err := json.Unmarshal(data, &config); err != nil {
			return []ValidationError{{Field: "_root", Message: "invalid JSON: " + err.Error()}}, nil
		}
	}
	return ValidateConfig(config), nil
}

func ValidateConfig(config map[string]any) []ValidationError {
	schema := DefaultSchema()
	return ValidateWithSchema(config, schema)
}

func ValidateWithSchema(config map[string]any, schema *Schema) []ValidationError {
	var errors []ValidationError
	for _, required := range schema.Required {
		if _, ok := config[required]; !ok {
			errors = append(errors, ValidationError{
				Field:   required,
				Message: fmt.Sprintf("required field '%s' is missing", required),
			})
		}
	}
	for key, val := range config {
		prop, ok := schema.Properties[key]
		if !ok {
			continue
		}
		if prop.Deprecated {
			errors = append(errors, ValidationError{
				Field:   key,
				Message: fmt.Sprintf("field '%s' is deprecated", key),
			})
		}
		errs := validateProperty(key, val, prop)
		errors = append(errors, errs...)
	}
	return errors
}

func validateProperty(field string, value any, prop Property) []ValidationError {
	var errors []ValidationError
	if !validateType(value, prop.Type) {
		return append(errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("field '%s' should be of type %s", field, prop.Type),
		})
	}
	if len(prop.Enum) > 0 {
		found := false
		for _, enumVal := range prop.Enum {
			if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", enumVal) {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("field '%s' must be one of %v", field, prop.Enum),
			})
		}
	}
	switch prop.Type {
	case "string":
		errors = append(errors, validateString(field, value.(string), prop)...)
	case "number":
		errors = append(errors, validateNumber(field, value, prop)...)
	case "array":
		errors = append(errors, validateArray(field, value.([]any), prop)...)
	case "object":
		errors = append(errors, validateObject(field, value.(map[string]any), prop)...)
	}
	return errors
}

func validateString(field, value string, prop Property) []ValidationError {
	var errors []ValidationError
	if prop.MinLength != nil && len(value) < *prop.MinLength {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("field '%s' must be at least %d characters", field, *prop.MinLength),
		})
	}
	if prop.MaxLength != nil && len(value) > *prop.MaxLength {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("field '%s' must be at most %d characters", field, *prop.MaxLength),
		})
	}
	if prop.Pattern != "" {
		matched, err := regexp.MatchString(prop.Pattern, value)
		if err != nil {
			return append(errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("invalid pattern for field '%s': %v", field, err),
			})
		}
		if !matched {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("field '%s' does not match pattern %s", field, prop.Pattern),
			})
		}
	}
	return errors
}

func validateNumber(field string, value any, prop Property) []ValidationError {
	var errors []ValidationError
	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float64:
		num = v
	default:
		return errors
	}
	if prop.Minimum != nil && num < *prop.Minimum {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("field '%s' must be at least %v", field, *prop.Minimum),
		})
	}
	if prop.Maximum != nil && num > *prop.Maximum {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("field '%s' must be at most %v", field, *prop.Maximum),
		})
	}
	return errors
}

func validateArray(field string, value []any, prop Property) []ValidationError {
	var errors []ValidationError
	if prop.Items != nil {
		for i, item := range value {
			itemField := fmt.Sprintf("%s[%d]", field, i)
			errs := validateProperty(itemField, item, *prop.Items)
			errors = append(errors, errs...)
		}
	}
	return errors
}

func validateObject(field string, value map[string]any, prop Property) []ValidationError {
	var errors []ValidationError
	for _, required := range prop.Required {
		if _, ok := value[required]; !ok {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.%s", field, required),
				Message: fmt.Sprintf("required field '%s' is missing in '%s'", required, field),
			})
		}
	}
	for key, val := range value {
		childProp, ok := prop.Properties[key]
		if !ok {
			continue
		}
		childField := fmt.Sprintf("%s.%s", field, key)
		errs := validateProperty(childField, val, childProp)
		errors = append(errors, errs...)
	}
	return errors
}

func validateType(val any, expected string) bool {
	switch expected {
	case "string":
		_, ok := val.(string)
		return ok
	case "boolean":
		_, ok := val.(bool)
		return ok
	case "number":
		switch val.(type) {
		case int, int64, float64:
			return true
		}
		return false
	case "array":
		_, ok := val.([]any)
		return ok
	case "object":
		_, ok := val.(map[string]any)
		return ok
	}
	return true
}

func (s *Schema) ToJSONSchema() map[string]any {
	result := map[string]any{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    s.Type,
	}
	if s.Title != "" {
		result["title"] = s.Title
	}
	if s.Description != "" {
		result["description"] = s.Description
	}
	if len(s.Required) > 0 {
		result["required"] = s.Required
	}
	props := make(map[string]any)
	for name, prop := range s.Properties {
		p := map[string]any{
			"type": prop.Type,
		}
		if prop.Description != "" {
			p["description"] = prop.Description
		}
		if prop.Default != nil {
			p["default"] = prop.Default
		}
		if len(prop.Enum) > 0 {
			p["enum"] = prop.Enum
		}
		if prop.MinLength != nil {
			p["minLength"] = *prop.MinLength
		}
		if prop.MaxLength != nil {
			p["maxLength"] = *prop.MaxLength
		}
		if prop.Pattern != "" {
			p["pattern"] = prop.Pattern
		}
		if prop.Minimum != nil {
			p["minimum"] = *prop.Minimum
		}
		if prop.Maximum != nil {
			p["maximum"] = *prop.Maximum
		}
		if prop.Items != nil {
			itemSchema := map[string]any{"type": prop.Items.Type}
			p["items"] = itemSchema
		}
		if prop.Deprecated {
			p["deprecated"] = true
		}
		props[name] = p
	}
	result["properties"] = props
	return result
}

func (s *Schema) GenerateDocumentation() string {
	var sb strings.Builder
	title := s.Title
	if title == "" {
		title = "Configuration Schema"
	}
	fmt.Fprintf(&sb, "# %s\n\n", title)
	if s.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", s.Description)
	}
	if len(s.Required) > 0 {
		sb.WriteString("## Required Fields\n\n")
		for _, field := range s.Required {
			fmt.Fprintf(&sb, "- `%s`\n", field)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("## Properties\n\n")
	names := make([]string, 0, len(s.Properties))
	for name := range s.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		prop := s.Properties[name]
		fmt.Fprintf(&sb, "### `%s` (%s)\n\n", name, prop.Type)
		if prop.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", prop.Description)
		}
		if prop.Default != nil {
			fmt.Fprintf(&sb, "- Default: `%v`\n", prop.Default)
		}
		if len(prop.Enum) > 0 {
			fmt.Fprintf(&sb, "- Allowed values: %v\n", prop.Enum)
		}
		if prop.MinLength != nil {
			fmt.Fprintf(&sb, "- Min length: %d\n", *prop.MinLength)
		}
		if prop.MaxLength != nil {
			fmt.Fprintf(&sb, "- Max length: %d\n", *prop.MaxLength)
		}
		if prop.Pattern != "" {
			fmt.Fprintf(&sb, "- Pattern: `%s`\n", prop.Pattern)
		}
		if prop.Minimum != nil {
			fmt.Fprintf(&sb, "- Minimum: %v\n", *prop.Minimum)
		}
		if prop.Maximum != nil {
			fmt.Fprintf(&sb, "- Maximum: %v\n", *prop.Maximum)
		}
		if prop.Deprecated {
			sb.WriteString("- **DEPRECATED**\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func LoadSchemaFromFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrConfig, "read schema")
	}
	ext := filepath.Ext(path)
	var schema Schema
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &schema); err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse YAML schema")
		}
	default:
		if err := json.Unmarshal(data, &schema); err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse JSON schema")
		}
	}
	return &schema, nil
}

func intPtr(n int) *int             { return &n }
func float64Ptr(n float64) *float64 { return &n }
