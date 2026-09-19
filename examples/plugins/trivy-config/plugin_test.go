// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestScanSummarizesFindings(t *testing.T) {
	original := runTrivy
	runTrivy = func(context.Context, string, ...string) ([]byte, error) {
		return json.RawMessage(`{"Results":[{"Target":"Dockerfile","Type":"dockerfile","Misconfigurations":[{"Severity":"HIGH"},{"Severity":"LOW"}]}]}`), nil
	}
	t.Cleanup(func() { runTrivy = original })

	result, err := scan(context.Background(), "Dockerfile")
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if result["ok"] != false || result["finding_count"] != 2 || result["blocking_count"] != 1 {
		t.Fatalf("unexpected summary: %#v", result)
	}
	counts := result["severity_counts"].(map[string]int)
	if counts["HIGH"] != 1 || counts["LOW"] != 1 {
		t.Fatalf("unexpected severity counts: %#v", counts)
	}
}

func TestScanDoesNotBlockOnLowFindings(t *testing.T) {
	original := runTrivy
	runTrivy = func(context.Context, string, ...string) ([]byte, error) {
		return json.RawMessage(`{"Results":[{"Target":"Dockerfile","Type":"dockerfile","Misconfigurations":[{"Severity":"LOW"},{"Severity":"MEDIUM"}]}]}`), nil
	}
	t.Cleanup(func() { runTrivy = original })

	result, err := scan(context.Background(), "Dockerfile")
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if result["ok"] != true || result["finding_count"] != 2 || result["blocking_count"] != 0 {
		t.Fatalf("unexpected non-blocking summary: %#v", result)
	}
}

func TestScanRejectsInvalidJSON(t *testing.T) {
	original := runTrivy
	runTrivy = func(context.Context, string, ...string) ([]byte, error) {
		return []byte("not-json"), nil
	}
	t.Cleanup(func() { runTrivy = original })

	if _, err := scan(context.Background(), "Dockerfile"); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestPluginRequiresTarget(t *testing.T) {
	if _, err := New().Execute("scan", nil); err == nil {
		t.Fatal("expected missing target error")
	}
}

func TestPluginRejectsUnknownAction(t *testing.T) {
	if _, err := New().Execute("unknown", nil); err == nil {
		t.Fatal("expected unknown action error")
	}
}
