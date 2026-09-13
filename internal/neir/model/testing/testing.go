// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package testing

type TestingStrategy string

const (
	StrategyUnit        TestingStrategy = "unit"
	StrategyIntegration TestingStrategy = "integration"
	StrategyE2E         TestingStrategy = "e2e"
	StrategyContract    TestingStrategy = "contract"
)

type Testing struct {
	Strategy   TestingStrategy   `json:"strategy,omitempty"`
	Frameworks []string          `json:"frameworks,omitempty"`
	Coverage   *Coverage         `json:"coverage,omitempty"`
	Fixtures   []Fixture         `json:"fixtures,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Coverage struct {
	MinPercent float64 `json:"min_percent,omitempty"`
}

type Fixture struct {
	Name string `json:"name"`
	Kind string `json:"kind,omitempty"`
	Path string `json:"path,omitempty"`
}
