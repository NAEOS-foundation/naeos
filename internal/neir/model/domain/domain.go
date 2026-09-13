// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package domain

type Domain struct {
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	BoundedContexts []BoundedContext  `json:"bounded_contexts,omitempty"`
	Aggregates      []Aggregate       `json:"aggregates,omitempty"`
	Entities        []Entity          `json:"entities,omitempty"`
	ValueObjects    []ValueObject     `json:"value_objects,omitempty"`
	Attributes      map[string]string `json:"attributes,omitempty"`
}

type BoundedContext struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Modules     []string `json:"modules,omitempty"`
}

type Aggregate struct {
	Name       string   `json:"name"`
	RootEntity string   `json:"root_entity,omitempty"`
	Entities   []string `json:"entities,omitempty"`
}

type Entity struct {
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type ValueObject struct {
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes,omitempty"`
}
