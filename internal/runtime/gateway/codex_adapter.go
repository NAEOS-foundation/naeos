// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"encoding/json"
	"fmt"
)

// CodexToolAdapter translates the narrow NAEOS Codex bridge envelope into a
// ToolRequest. It is transport-only: it never authorizes or executes.
type CodexToolAdapter struct{}

func (CodexToolAdapter) Name() string { return "codex" }

func (CodexToolAdapter) NormalizeTool(raw any) (ToolRequest, error) {
	data, ok := raw.(map[string]any)
	if !ok {
		return ToolRequest{}, fmt.Errorf("codex adapter expects an object")
	}
	typ, _ := data["type"].(string)
	if typ != "function_call" {
		return ToolRequest{}, fmt.Errorf("unsupported Codex call type %q", typ)
	}
	name, _ := data["name"].(string)
	if name == "" {
		return ToolRequest{}, fmt.Errorf("codex function name is required")
	}
	args, err := decodeCodexArguments(data["arguments"])
	if err != nil {
		return ToolRequest{}, err
	}
	actor, _ := args["actor"].(string)
	if actor == "" {
		actor, _ = data["actor"].(string)
	}
	if actor == "" {
		actor = "codex"
	}
	action, _ := args["action"].(string)
	if action == "" {
		action = "execute"
	}
	return ToolRequest{
		Capability:  stringValue(args["capability"]),
		Tool:        name,
		Action:      action,
		Resource:    stringValue(args["resource"]),
		Environment: stringValue(args["environment"]),
		Actor:       actor,
		Payload:     mapValue(args["payload"]),
		Context:     mapValue(args["context"]),
	}, nil
}

func decodeCodexArguments(raw any) (map[string]any, error) {
	if raw == nil {
		return map[string]any{}, nil
	}
	switch value := raw.(type) {
	case string:
		var args map[string]any
		if err := json.Unmarshal([]byte(value), &args); err != nil {
			return nil, fmt.Errorf("decode Codex arguments: %w", err)
		}
		return args, nil
	case map[string]any:
		return value, nil
	default:
		return nil, fmt.Errorf("codex arguments must be an object or JSON string")
	}
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	result, _ := value.(string)
	return result
}

func mapValue(value any) map[string]any {
	if value == nil {
		return nil
	}
	result, _ := value.(map[string]any)
	return result
}

func (CodexToolAdapter) OnDecision(result ExecutionResult) error { return nil }
