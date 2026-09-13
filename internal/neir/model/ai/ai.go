// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package ai

type AI struct {
	Models         []Model           `json:"models,omitempty"`
	Prompts        []Prompt          `json:"prompts,omitempty"`
	ContextBundles []ContextBundle   `json:"context_bundles,omitempty"`
	Embeddings     []Embedding       `json:"embeddings,omitempty"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

type Model struct {
	Name    string `json:"name"`
	Kind    string `json:"kind,omitempty"`
	Version string `json:"version,omitempty"`
}

type Prompt struct {
	Name     string `json:"name"`
	Template string `json:"template,omitempty"`
	Kind     string `json:"kind,omitempty"`
}

type ContextBundle struct {
	Name    string   `json:"name"`
	Sources []string `json:"sources,omitempty"`
}

type Embedding struct {
	Name      string `json:"name"`
	Dimension int    `json:"dimension,omitempty"`
}
