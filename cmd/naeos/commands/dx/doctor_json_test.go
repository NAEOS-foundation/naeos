// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package dx

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func TestRenderDoctorJSONHealthy(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	results := []checkResult{
		{Name: "Go", Status: "pass", Detail: "go1.25"},
		{Name: "Git", Status: "pass"},
	}
	if err := renderDoctorJSON(cmd, results); err != nil {
		t.Fatal(err)
	}

	var report map[string]any
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if report["status"] != "healthy" {
		t.Errorf("expected healthy, got %v", report["status"])
	}
	if report["passed"] != float64(2) {
		t.Errorf("expected passed=2, got %v", report["passed"])
	}
}

func TestRenderDoctorJSONDegraded(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	results := []checkResult{
		{Name: "Go", Status: "pass"},
		{Name: "Node", Status: "warn", Detail: "old version"},
	}
	if err := renderDoctorJSON(cmd, results); err != nil {
		t.Fatal(err)
	}

	var report map[string]any
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if report["status"] != "degraded" {
		t.Errorf("expected degraded, got %v", report["status"])
	}
	if report["warned"] != float64(1) {
		t.Errorf("expected warned=1, got %v", report["warned"])
	}
}

func TestRenderDoctorJSONUnhealthy(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	results := []checkResult{
		{Name: "Go", Status: "pass"},
		{Name: "Docker", Status: "fail", Detail: "not running"},
		{Name: "Broker", Status: "fail"},
	}
	if err := renderDoctorJSON(cmd, results); err != nil {
		t.Fatal(err)
	}

	var report map[string]any
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if report["status"] != "unhealthy" {
		t.Errorf("expected unhealthy, got %v", report["status"])
	}
	if report["failed"] != float64(2) {
		t.Errorf("expected failed=2, got %v", report["failed"])
	}
	if report["version"] == "" || report["go"] == "" || report["platform"] == "" {
		t.Error("expected version/go/platform fields")
	}
}
