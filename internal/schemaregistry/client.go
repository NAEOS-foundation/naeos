// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package schemaregistry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

const (
	DefaultNEIRSchemaURL = "https://registry.naeos.dev/schemaregistry/latest.json"
	DefaultSpecSchemaURL = "https://registry.naeos.dev/schemaregistry/spec/latest.json"
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
			return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "read local schema")
		}
	} else {
		resp, err := c.httpClient.Get(c.registryURL)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "fetch schema")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, naeoserr.New(naeoserr.ErrNetwork, fmt.Sprintf("fetch schema: HTTP %d", resp.StatusCode))
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "read schema")
		}
	}

	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse schema")
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
		return NEIRValidationResult{}, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "read spec")
	}

	ext := filepath.Ext(specPath)
	var spec map[string]any
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return NEIRValidationResult{}, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse YAML")
		}
	default:
		if err := json.Unmarshal(data, &spec); err != nil {
			return NEIRValidationResult{}, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse JSON")
		}
	}

	delete(spec, "$schema")

	return validateSpec(spec, schema), nil
}

func ValidateSpec(spec map[string]any, schema map[string]any) NEIRValidationResult {
	return validateSpec(spec, schema)
}

func validateSpec(spec map[string]any, schema map[string]any) NEIRValidationResult {
	return ValidateJSONSchema(spec, schema)
}

// ValidateJSONSchema applies full JSON Schema validation to a document. External
// schema references are unavailable unless their resources are embedded in the
// provided schema, so validation cannot trigger implicit network or file loads.
func ValidateJSONSchema(document any, schema map[string]any) NEIRValidationResult {
	result := NEIRValidationResult{Valid: true, Version: NEIRSchemaVersion}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft7)
	schemaURL, _ := schema["$id"].(string)
	if schemaURL == "" {
		schemaURL = "https://naeos.invalid/schema.json"
	}
	if err := compiler.AddResource(schemaURL, schema); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, NEIRValidationError{Message: fmt.Sprintf("register schema: %v", err)})
		return result
	}
	compiled, err := compiler.Compile(schemaURL)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, NEIRValidationError{Message: fmt.Sprintf("compile schema: %v", err)})
		return result
	}
	if err := compiled.Validate(document); err != nil {
		result.Valid = false
		var validationErr *jsonschema.ValidationError
		if errors.As(err, &validationErr) {
			appendSchemaOutputErrors(&result.Errors, validationErr.DetailedOutput())
		} else {
			result.Errors = append(result.Errors, NEIRValidationError{Message: err.Error()})
		}
	}

	return result
}

func appendSchemaOutputErrors(dst *[]NEIRValidationError, output *jsonschema.OutputUnit) {
	if output.Error != nil && len(output.Errors) == 0 {
		*dst = append(*dst, NEIRValidationError{
			Field:   output.InstanceLocation,
			Message: output.Error.String(),
		})
	}
	for i := range output.Errors {
		appendSchemaOutputErrors(dst, &output.Errors[i])
	}
}
