// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
)

const PolicyContextVersion = "v1"

// PolicyContextFields is the explicit, versioned policy-visible surface.
// New NEIR fields MUST NOT become policy-visible implicitly; they must be
// added here as part of an intentional policy-context contract change.
var PolicyContextFields = []string{
	"ai", "apis", "architecture", "components", "deployment",
	"documentation", "domain", "generation", "infrastructure", "inherits",
	"metadata", "modules", "project", "security", "services", "storage",
	"testing", "active_profile",
}

type PolicyContext struct {
	Version string
	Values  map[string]any
}

func (c PolicyContext) Validate() error {
	if c.Version != PolicyContextVersion {
		return fmt.Errorf("unsupported policy context version %q", c.Version)
	}
	if c.Values == nil {
		return fmt.Errorf("policy context values are nil")
	}
	return nil
}

func (c PolicyContext) Digest() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	payload := struct {
		Version string         `json:"version"`
		Values  map[string]any `json:"values"`
	}{Version: c.Version, Values: c.Values}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal policy context digest payload: %w", err)
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

func ContextFromNEIR(neir *model.NEIR) (PolicyContext, error) {
	if neir == nil {
		return PolicyContext{}, fmt.Errorf("NEIR is nil")
	}
	data, err := json.Marshal(neir)
	if err != nil {
		return PolicyContext{}, fmt.Errorf("marshal NEIR policy context: %w", err)
	}
	var allValues map[string]any
	if err := json.Unmarshal(data, &allValues); err != nil {
		return PolicyContext{}, fmt.Errorf("decode NEIR policy context: %w", err)
	}

	values := make(map[string]any, len(PolicyContextFields))
	for _, field := range PolicyContextFields {
		if value, ok := allValues[field]; ok {
			values[field] = value
		}
	}
	return PolicyContext{Version: PolicyContextVersion, Values: values}, nil
}

func PolicyContextFieldNames() []string {
	fields := append([]string(nil), PolicyContextFields...)
	sort.Strings(fields)
	return fields
}
