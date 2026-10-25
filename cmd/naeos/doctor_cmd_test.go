// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

func TestDoctorJSONViaCommand(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "doctor", "--output-format", "json", "--quick")
	if err != nil {
		t.Fatalf("doctor --format json failed: %v", err)
	}
	if !strings.Contains(output, `"status"`) {
		t.Fatalf("expected JSON report, got %q", output)
	}
}

func TestDoctorHelp(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "doctor", "--help")
	if err != nil {
		t.Fatalf("doctor --help failed: %v", err)
	}
	if !strings.Contains(output, "doctor") {
		t.Fatalf("expected doctor help, got %q", output)
	}
}
