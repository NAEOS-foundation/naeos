// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

type bridgeAllowControlPlane struct{}

func (bridgeAllowControlPlane) Evaluate(req control.Request) (control.DecisionRecord, error) {
	return control.DecisionRecord{Request: req, Decision: control.DecisionAllow, PolicyID: "bridge-test", RuleID: "allow"}, nil
}

func TestRuntimeBridgeGatewayRejectsReplayBeforeSecondWrite(t *testing.T) {
	root := t.TempDir()
	gw := newRuntimeBridgeGateway(bridgeAllowControlPlane{}, root, gateway.NewInMemoryInvocationStore())

	request := map[string]any{
		"request_id":    "req-bridge-replay",
		"invocation_id": "inv-bridge-replay",
		"actor":        "manus",
		"tool":         "filesystem",
		"action":       "write",
		"resource":     "filesystem",
		"payload": map[string]any{
			"path":    "replay.txt",
			"content": "one write",
		},
	}

	first, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	if first.Status != "completed" {
		t.Fatalf("first request status = %q, want completed: %+v", first.Status, first)
	}
	if first.Observation == nil {
		t.Fatal("first request missing observation")
	}
	if first.Observation.RequestID != request["request_id"] || first.Observation.InvocationID != request["invocation_id"] {
		t.Fatalf("observation identity = (%q, %q), want (%q, %q)",
			first.Observation.RequestID, first.Observation.InvocationID,
			request["request_id"], request["invocation_id"])
	}

	second, err := gw.AuthorizeFromAdapter("json", request)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	if second.Status != "denied" || second.Output != "invocation already consumed" {
		t.Fatalf("second request = %+v, want replay denial", second)
	}

	data, err := os.ReadFile(filepath.Join(root, "replay.txt"))
	if err != nil {
		t.Fatalf("read governed artifact: %v", err)
	}
	if string(data) != "one write" {
		t.Fatalf("artifact content = %q, want one write", string(data))
	}
}
