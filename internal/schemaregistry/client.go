package schemaregistry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultNEIRSchemaURL = "https://naeos.dev/schemaregistry/latest.json"
	NEIRSchemaVersion    = "v1"
)

type NEIRValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type NEIRValidationResult struct {
	Valid   bool                  `json:"valid"`
	Version string                `json:"version"`
	Errors  []NEIRValidationError `json:"errors,omitempty"`
}

type NEIRClient struct {
	registryURL string
	httpClient  *http.Client
}

func NewNEIRClient(registryURL string) *NEIRClient {
	if registryURL == "" {
		registryURL = DefaultNEIRSchemaURL
	}
	return &NEIRClient{
		registryURL: registryURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *NEIRClient) FetchSchema() (map[string]any, error) {
	var body []byte

	if strings.HasPrefix(c.registryURL, "file://") {
		path := strings.TrimPrefix(c.registryURL, "file://")
		var err error
		body, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read local schema: %w", err)
		}
	} else {
		resp, err := c.httpClient.Get(c.registryURL)
		if err != nil {
			return nil, fmt.Errorf("fetch schema: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch schema: HTTP %d", resp.StatusCode)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read schema: %w", err)
		}
	}

	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	return schema, nil
}

func (c *NEIRClient) FetchSchemaVersion(version string) (map[string]any, error) {
	baseURL := c.registryURL
	if version != "" {
		u, err := url.Parse(c.registryURL)
		if err != nil {
			return nil, err
		}
		u.Path = filepath.Join(filepath.Dir(u.Path), version, "neir.json")
		baseURL = u.String()
	}
	c2 := NewNEIRClient(baseURL)
	return c2.FetchSchema()
}

func ValidateNEIRSpec(specPath string, schema map[string]any) (NEIRValidationResult, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return NEIRValidationResult{}, fmt.Errorf("read spec: %w", err)
	}

	ext := filepath.Ext(specPath)
	var spec map[string]any
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return NEIRValidationResult{}, fmt.Errorf("parse YAML: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &spec); err != nil {
			return NEIRValidationResult{}, fmt.Errorf("parse JSON: %w", err)
		}
	}

	delete(spec, "$schema")

	result := NEIRValidationResult{
		Valid:   true,
		Version: NEIRSchemaVersion,
	}

	props, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]any)

	reqFields := make([]string, 0, len(required))
	for _, r := range required {
		if s, ok := r.(string); ok {
			reqFields = append(reqFields, s)
		}
	}

	for _, field := range reqFields {
		if _, exists := spec[field]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, NEIRValidationError{
				Field:   field,
				Message: fmt.Sprintf("required field '%s' is missing", field),
			})
		}
	}

	for key, val := range spec {
		propSchema, ok := props[key].(map[string]any)
		if ok {
			errs := validateNEIRValue(key, val, propSchema, schema)
			result.Errors = append(result.Errors, errs...)
		}
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result, nil
}

func validateNEIRValue(field string, value any, propSchema map[string]any, rootSchema map[string]any) []NEIRValidationError {
	var errs []NEIRValidationError

	propType, _ := propSchema["type"].(string)

	if ref, ok := propSchema["$ref"].(string); ok {
		defSchema := resolveNEIRRef(ref, rootSchema)
		if defSchema != nil {
			return validateNEIRValue(field, value, defSchema, rootSchema)
		}
	}

	switch propType {
	case "object":
		if valMap, ok := value.(map[string]any); ok {
			subProps, _ := propSchema["properties"].(map[string]any)
			subRequired, _ := propSchema["required"].([]any)

			for _, r := range subRequired {
				if s, ok := r.(string); ok {
					if _, exists := valMap[s]; !exists {
						errs = append(errs, NEIRValidationError{
							Field:   fmt.Sprintf("%s.%s", field, s),
							Message: fmt.Sprintf("required field '%s.%s' is missing", field, s),
						})
					}
				}
			}

			for key, val := range valMap {
				if subProp, ok := subProps[key].(map[string]any); ok {
					subErrs := validateNEIRValue(fmt.Sprintf("%s.%s", field, key), val, subProp, rootSchema)
					errs = append(errs, subErrs...)
				}
			}
		}
	case "array":
		if arr, ok := value.([]any); ok {
			items, _ := propSchema["items"].(map[string]any)
			if items != nil {
				if ref, ok := items["$ref"].(string); ok {
					defSchema := resolveNEIRRef(ref, rootSchema)
					if defSchema != nil {
						items = defSchema
					}
				}
				for i, item := range arr {
					itemField := fmt.Sprintf("%s[%d]", field, i)
					if itemMap, ok := item.(map[string]any); ok {
						subErrs := validateNEIRValue(itemField, itemMap, items, rootSchema)
						errs = append(errs, subErrs...)
					}
				}
			}
		}
	case "string":
		if strVal, ok := value.(string); ok {
			if enumVals, ok := propSchema["enum"].([]any); ok {
				found := false
				for _, e := range enumVals {
					if strVal == fmt.Sprintf("%v", e) {
						found = true
						break
					}
				}
				if !found {
					errs = append(errs, NEIRValidationError{
						Field:   field,
						Message: fmt.Sprintf("'%s' must be one of %v", field, enumVals),
					})
				}
			}
		}
	}

	return errs
}

func resolveNEIRRef(ref string, rootSchema map[string]any) map[string]any {
	if len(ref) < 2 || ref[:2] != "#/" {
		return nil
	}

	parts := splitNEIRPath(ref[2:])
	current := rootSchema
	for _, part := range parts {
		if m, ok := current[part].(map[string]any); ok {
			current = m
		} else {
			return nil
		}
	}
	return current
}

func splitNEIRPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
