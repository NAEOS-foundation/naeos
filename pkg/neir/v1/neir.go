// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

// Package neir defines the version 1 public JSON contract for serialized NEIR
// model documents. Project, modules, services, and metadata are typed. Other
// model sections remain raw JSON until their public field-level contracts are
// designed. Unknown fields are retained when documents are decoded and encoded.
package neir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const (
	// SchemaVersion identifies the public model schema family.
	SchemaVersion = "v1"
	// SchemaID is the canonical schema identifier for this package's document shape.
	SchemaID = "https://registry.naeos.dev/schemaregistry/neir.json"
)

// Document is a public representation of one serialized NEIR model. Sections
// without a stable public DTO are kept as RawMessage and can be inspected or
// passed through without importing NAEOS implementation packages.
type Document struct {
	Project        *Project        `json:"project,omitempty"`
	Architecture   json.RawMessage `json:"architecture,omitempty"`
	Domain         json.RawMessage `json:"domain,omitempty"`
	Modules        []Module        `json:"modules,omitempty"`
	Components     json.RawMessage `json:"components,omitempty"`
	Services       []Service       `json:"services,omitempty"`
	APIs           json.RawMessage `json:"apis,omitempty"`
	Storage        json.RawMessage `json:"storage,omitempty"`
	Infrastructure json.RawMessage `json:"infrastructure,omitempty"`
	Security       json.RawMessage `json:"security,omitempty"`
	AI             json.RawMessage `json:"ai,omitempty"`
	Documentation  json.RawMessage `json:"documentation,omitempty"`
	Deployment     json.RawMessage `json:"deployment,omitempty"`
	Testing        json.RawMessage `json:"testing,omitempty"`
	Metadata       *Metadata       `json:"metadata,omitempty"`
	Generation     json.RawMessage `json:"generation,omitempty"`
	ActiveProfile  string          `json:"active_profile,omitempty"`
	Inherits       string          `json:"inherits,omitempty"`

	// AdditionalProperties stores top-level JSON fields unknown to this DTO.
	AdditionalProperties map[string]json.RawMessage `json:"-"`
}

// Project describes the stable public project identity and common attributes.
type Project struct {
	Name                 string                     `json:"name"`
	Version              string                     `json:"version,omitempty"`
	Description          string                     `json:"description,omitempty"`
	License              string                     `json:"license,omitempty"`
	Authors              []string                   `json:"authors,omitempty"`
	Repository           string                     `json:"repository,omitempty"`
	Tags                 []string                   `json:"tags,omitempty"`
	Attributes           map[string]string          `json:"attributes,omitempty"`
	AdditionalProperties map[string]json.RawMessage `json:"-"`
}

// Module describes one module in a NEIR document.
type Module struct {
	Name                 string                     `json:"name"`
	Path                 string                     `json:"path,omitempty"`
	Description          string                     `json:"description,omitempty"`
	Condition            string                     `json:"condition,omitempty"`
	Packages             []string                   `json:"packages,omitempty"`
	Dependencies         []string                   `json:"dependencies,omitempty"`
	Attributes           map[string]string          `json:"attributes,omitempty"`
	AdditionalProperties map[string]json.RawMessage `json:"-"`
}

// Service describes one service and its public endpoint list.
type Service struct {
	Name                 string                     `json:"name"`
	Kind                 string                     `json:"kind,omitempty"`
	Port                 int                        `json:"port,omitempty"`
	Description          string                     `json:"description,omitempty"`
	Condition            string                     `json:"condition,omitempty"`
	Endpoints            []Endpoint                 `json:"endpoints,omitempty"`
	Middleware           []string                   `json:"middleware,omitempty"`
	Attributes           map[string]string          `json:"attributes,omitempty"`
	AdditionalProperties map[string]json.RawMessage `json:"-"`
}

// Endpoint describes an operation exposed by a service.
type Endpoint struct {
	Method               string                     `json:"method,omitempty"`
	Path                 string                     `json:"path,omitempty"`
	Action               string                     `json:"action,omitempty"`
	AdditionalProperties map[string]json.RawMessage `json:"-"`
}

// Metadata contains version and provenance fields for the NEIR document.
type Metadata struct {
	NEIRVersion          string                     `json:"neir_version,omitempty"`
	SchemaVersion        string                     `json:"schema_version,omitempty"`
	ProjectVersion       string                     `json:"project_version,omitempty"`
	CreatedAt            *time.Time                 `json:"created_at,omitempty"`
	ModifiedAt           *time.Time                 `json:"modified_at,omitempty"`
	Tags                 []string                   `json:"tags,omitempty"`
	Labels               map[string]string          `json:"labels,omitempty"`
	Source               *SourceRef                 `json:"source,omitempty"`
	AdditionalProperties map[string]json.RawMessage `json:"-"`
}

// SourceRef identifies the origin of a NEIR document.
type SourceRef struct {
	Kind                 string                     `json:"kind,omitempty"`
	Ref                  string                     `json:"ref,omitempty"`
	AdditionalProperties map[string]json.RawMessage `json:"-"`
}

// Decode parses a JSON NEIR document and retains unknown fields.
func Decode(data []byte) (*Document, error) {
	var document Document
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode NEIR v1 document: %w", err)
	}
	return &document, nil
}

var documentFields = fields(
	"project", "architecture", "domain", "modules", "components", "services",
	"apis", "storage", "infrastructure", "security", "ai", "documentation",
	"deployment", "testing", "metadata", "generation", "active_profile", "inherits",
)

func (d *Document) UnmarshalJSON(data []byte) error {
	type wire Document
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	extra, err := unknownFields(data, documentFields)
	if err != nil {
		return err
	}
	*d = Document(decoded)
	d.AdditionalProperties = extra
	return nil
}

func (d Document) MarshalJSON() ([]byte, error) {
	type wire Document
	data, err := json.Marshal(wire(d))
	if err != nil {
		return nil, err
	}
	return mergeUnknown(data, d.AdditionalProperties, documentFields)
}

func (p *Project) UnmarshalJSON(data []byte) error {
	type wire Project
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	extra, err := unknownFields(data, fields("name", "version", "description", "license", "authors", "repository", "tags", "attributes"))
	if err != nil {
		return err
	}
	*p = Project(decoded)
	p.AdditionalProperties = extra
	return nil
}

func (p Project) MarshalJSON() ([]byte, error) {
	type wire Project
	data, err := json.Marshal(wire(p))
	if err != nil {
		return nil, err
	}
	return mergeUnknown(data, p.AdditionalProperties, fields("name", "version", "description", "license", "authors", "repository", "tags", "attributes"))
}

func (m *Module) UnmarshalJSON(data []byte) error {
	type wire Module
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	extra, err := unknownFields(data, fields("name", "path", "description", "condition", "packages", "dependencies", "attributes"))
	if err != nil {
		return err
	}
	*m = Module(decoded)
	m.AdditionalProperties = extra
	return nil
}

func (m Module) MarshalJSON() ([]byte, error) {
	type wire Module
	data, err := json.Marshal(wire(m))
	if err != nil {
		return nil, err
	}
	return mergeUnknown(data, m.AdditionalProperties, fields("name", "path", "description", "condition", "packages", "dependencies", "attributes"))
}

func (s *Service) UnmarshalJSON(data []byte) error {
	type wire Service
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	extra, err := unknownFields(data, fields("name", "kind", "port", "description", "condition", "endpoints", "middleware", "attributes"))
	if err != nil {
		return err
	}
	*s = Service(decoded)
	s.AdditionalProperties = extra
	return nil
}

func (s Service) MarshalJSON() ([]byte, error) {
	type wire Service
	data, err := json.Marshal(wire(s))
	if err != nil {
		return nil, err
	}
	return mergeUnknown(data, s.AdditionalProperties, fields("name", "kind", "port", "description", "condition", "endpoints", "middleware", "attributes"))
}

func (e *Endpoint) UnmarshalJSON(data []byte) error {
	type wire Endpoint
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	extra, err := unknownFields(data, fields("method", "path", "action"))
	if err != nil {
		return err
	}
	*e = Endpoint(decoded)
	e.AdditionalProperties = extra
	return nil
}

func (e Endpoint) MarshalJSON() ([]byte, error) {
	type wire Endpoint
	data, err := json.Marshal(wire(e))
	if err != nil {
		return nil, err
	}
	return mergeUnknown(data, e.AdditionalProperties, fields("method", "path", "action"))
}

func (m *Metadata) UnmarshalJSON(data []byte) error {
	type wire Metadata
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	extra, err := unknownFields(data, fields("neir_version", "schema_version", "project_version", "created_at", "modified_at", "tags", "labels", "source"))
	if err != nil {
		return err
	}
	*m = Metadata(decoded)
	m.AdditionalProperties = extra
	return nil
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	type wire Metadata
	data, err := json.Marshal(wire(m))
	if err != nil {
		return nil, err
	}
	return mergeUnknown(data, m.AdditionalProperties, fields("neir_version", "schema_version", "project_version", "created_at", "modified_at", "tags", "labels", "source"))
}

func (s *SourceRef) UnmarshalJSON(data []byte) error {
	type wire SourceRef
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	extra, err := unknownFields(data, fields("kind", "ref"))
	if err != nil {
		return err
	}
	*s = SourceRef(decoded)
	s.AdditionalProperties = extra
	return nil
}

func (s SourceRef) MarshalJSON() ([]byte, error) {
	type wire SourceRef
	data, err := json.Marshal(wire(s))
	if err != nil {
		return nil, err
	}
	return mergeUnknown(data, s.AdditionalProperties, fields("kind", "ref"))
}

func fields(names ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(names))
	for _, name := range names {
		result[name] = struct{}{}
	}
	return result
}

func unknownFields(data []byte, known map[string]struct{}) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	if object == nil {
		return nil, fmt.Errorf("NEIR document must be a JSON object")
	}
	for name := range known {
		delete(object, name)
	}
	if len(object) == 0 {
		return nil, nil
	}
	return object, nil
}

func mergeUnknown(data []byte, extra map[string]json.RawMessage, known map[string]struct{}) ([]byte, error) {
	if len(extra) == 0 {
		return data, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	for name, value := range extra {
		if _, isKnown := known[name]; isKnown {
			continue
		}
		if !json.Valid(value) {
			return nil, fmt.Errorf("additional property %q contains invalid JSON", name)
		}
		object[name] = bytes.Clone(value)
	}
	return json.Marshal(object)
}
