// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"encoding/json"
	"fmt"
)

// JSONToolAdapter is a protocol-neutral adapter for agent tool calls represented
// as JSON objects. It is intentionally narrow: normalization never grants
// authorization; the resulting ToolRequest must still cross the gateway.
type JSONToolAdapter struct{}

func (JSONToolAdapter) Name() string { return "json" }

func (JSONToolAdapter) NormalizeTool(raw any) (ToolRequest, error) {
	data, ok := raw.(map[string]any)
	if !ok {
		return ToolRequest{}, fmt.Errorf("JSON adapter expects an object")
	}
	request := ToolRequest{}
	encoded, err := json.Marshal(data)
	if err != nil {
		return ToolRequest{}, err
	}
	if err := json.Unmarshal(encoded, &request); err != nil {
		return ToolRequest{}, fmt.Errorf("invalid tool request: %w", err)
	}
	request.RequestID = stringValue(data["request_id"])
	request.InvocationID = stringValue(data["invocation_id"])
	if request.Tool == "" || request.Action == "" {
		return ToolRequest{}, fmt.Errorf("tool and action are required")
	}
	return request, nil
}

func (JSONToolAdapter) OnDecision(result ExecutionResult) error { return nil }
