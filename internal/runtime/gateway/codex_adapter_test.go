// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import "testing"

func TestCodexToolAdapterNormalizesFunctionCall(t *testing.T) {
    req, err := (CodexToolAdapter{}).NormalizeTool(map[string]any{
        "type": "function_call", "name": "file-edit",
        "arguments": map[string]any{"action":"write","resource":"src/app.go","capability":"repository.write","environment":"development","payload":map[string]any{"operation":"replace"}},
    })
    if err != nil { t.Fatalf("NormalizeTool() error = %v", err) }
    if req.Tool != "file-edit" || req.Action != "write" || req.Resource != "src/app.go" { t.Fatalf("unexpected request: %+v", req) }
    if req.Capability != "repository.write" || req.Actor != "codex" { t.Fatalf("unexpected authority fields: %+v", req) }
}

func TestCodexToolAdapterAcceptsJSONStringArguments(t *testing.T) {
    req, err := (CodexToolAdapter{}).NormalizeTool(map[string]any{
        "type":"function_call", "name":"shell", "arguments":"{\"action\":\"execute\",\"resource\":\"go test ./...\"}", "actor":"codex-agent",
    })
    if err != nil { t.Fatalf("NormalizeTool() error = %v", err) }
    if req.Tool != "shell" || req.Action != "execute" || req.Resource != "go test ./..." || req.Actor != "codex-agent" { t.Fatalf("unexpected request: %+v", req) }
}

func TestCodexToolAdapterRejectsNonFunctionCall(t *testing.T) {
    _, err := (CodexToolAdapter{}).NormalizeTool(map[string]any{"type":"tool_result","name":"file-edit"})
    if err == nil { t.Fatal("expected unsupported call type error") }
}

func TestCodexToolAdapterDoesNotAuthorize(t *testing.T) {
    req, err := (CodexToolAdapter{}).NormalizeTool(map[string]any{"type":"function_call","name":"deploy","arguments":map[string]any{"action":"execute"}})
    if err != nil { t.Fatalf("NormalizeTool() error = %v", err) }
    if req.Tool != "deploy" { t.Fatalf("Tool = %q, want deploy", req.Tool) }
}
