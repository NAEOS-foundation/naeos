// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
)

// PolicyContextVersion identifies the shape of the evaluator input contract.
// A policy runtime MUST treat a version change as a contract change.
const PolicyContextVersion = "v1"

type PolicyContext struct {
	Version string
	Values  map[string]any
}

// ContextFromNEIR builds the policy evaluator context from the complete
// top-level NEIR contract. JSON tags are the canonical field names, so newly
// supported NEIR domains become visible to policy evaluation without another
// hand-maintained allow-list in the pipeline.
func ContextFromNEIR(neir *model.NEIR) (PolicyContext, error) {
	if neir == nil {
		return PolicyContext{}, fmt.Errorf("NEIR is nil")
	}

	data, err := json.Marshal(neir)
	if err != nil {
		return PolicyContext{}, fmt.Errorf("marshal NEIR policy context: %w", err)
	}

	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		return PolicyContext{}, fmt.Errorf("decode NEIR policy context: %w", err)
	}

	return PolicyContext{
		Version: PolicyContextVersion,
		Values:  values,
	}, nil
}
