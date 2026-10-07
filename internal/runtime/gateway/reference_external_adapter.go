// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"encoding/json"
	"fmt"
)

// AdoptionEnvelope is the minimal protocol-neutral envelope accepted by the
// reference external-agent adapter. Provider-specific transports must map into
// this shape before entering the governed gateway.
type AdoptionEnvelope struct {
	Contract     string         `json:"contract"`
	Version      string         `json:"version"`
	RequestID    string         `json:"request_id"`
	InvocationID string         `json:"invocation_id"`
	Actor        string         `json:"actor"`
	Capability   string         `json:"capability"`
	Tool         string         `json:"tool"`
	Action       string         `json:"action"`
	Resource     string         `json:"resource"`
	Environment  string         `json:"environment"`
	Payload      map[string]any `json:"payload"`
	Context      map[string]any `json:"context"`
}

// ReferenceExternalAdapter demonstrates the provider-neutral adoption boundary.
// It normalizes an adoption envelope only; it never evaluates policy or executes.
type ReferenceExternalAdapter struct{}

func (ReferenceExternalAdapter) Name() string { return "reference-external-agent" }

func (ReferenceExternalAdapter) NormalizeTool(raw any) (ToolRequest, error) {
	data, ok := raw.(map[string]any)
	if !ok {
		return ToolRequest{}, fmt.Errorf("reference adapter expects an object")
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return ToolRequest{}, fmt.Errorf("encode adoption envelope: %w", err)
	}

	var envelope AdoptionEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return ToolRequest{}, fmt.Errorf("decode adoption envelope: %w", err)
	}
	if envelope.Contract != "naeos.agent-adoption" {
		return ToolRequest{}, fmt.Errorf("unsupported adoption contract %q", envelope.Contract)
	}
	if envelope.Version != "1.0" {
		return ToolRequest{}, fmt.Errorf("unsupported adoption contract version %q", envelope.Version)
	}
	if envelope.RequestID == "" || envelope.InvocationID == "" {
		return ToolRequest{}, fmt.Errorf("request_id and invocation_id are required")
	}
	if envelope.Tool == "" || envelope.Action == "" {
		return ToolRequest{}, fmt.Errorf("tool and action are required")
	}

	context := cloneMap(envelope.Context)
	if context == nil {
		context = make(map[string]any)
	}
	context["request_id"] = envelope.RequestID

	return ToolRequest{
		RequestID:    envelope.RequestID,
		InvocationID: envelope.InvocationID,
		Capability:   envelope.Capability,
		Tool:         envelope.Tool,
		Action:       envelope.Action,
		Resource:     envelope.Resource,
		Environment:  envelope.Environment,
		Actor:        envelope.Actor,
		Payload:      cloneMap(envelope.Payload),
		Context:      context,
	}, nil
}

func (ReferenceExternalAdapter) OnDecision(ExecutionResult) error { return nil }

func cloneMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}
