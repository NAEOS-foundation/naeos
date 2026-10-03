// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/sbom"
)

func TestSBOMVerifyRequiresLicensesWhenRequested(t *testing.T) {
	t.Parallel()

	bom, err := sbom.NewGenerator(sbom.GeneratorConfig{
		Project: "NAEOS",
		Version: "1.0.0",
	}).Generate([]sbom.Component{{Type: sbom.Library, Name: "dependency"}})
	if err != nil {
		t.Fatalf("generate BOM: %v", err)
	}

	path := filepath.Join(t.TempDir(), "bom.json")
	if err := sbom.Write(bom, path); err != nil {
		t.Fatalf("write BOM: %v", err)
	}

	cmd := newSBomVerifyCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"--require-licenses", path})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "SBOM verification failed") {
		t.Fatalf("expected license verification failure, got %v", err)
	}
	if !strings.Contains(output.String(), "licenses") || !strings.Contains(output.String(), "FAIL") {
		t.Fatalf("expected failed license check in output, got %q", output.String())
	}
}

func TestSBOMVerifyAcceptsCycloneDXLicenseChoices(t *testing.T) {
	t.Parallel()

	bom, err := sbom.NewGenerator(sbom.GeneratorConfig{}).Generate([]sbom.Component{{
		Type:     sbom.Library,
		Name:     "dependency",
		Licenses: []sbom.LicenseChoice{{Expression: "MIT AND Apache-2.0"}},
	}})
	if err != nil {
		t.Fatalf("generate BOM: %v", err)
	}

	path := filepath.Join(t.TempDir(), "bom.json")
	if err := sbom.Write(bom, path); err != nil {
		t.Fatalf("write BOM: %v", err)
	}

	cmd := newSBomVerifyCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"--require-licenses", "--output", "json", path})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("verify BOM: %v\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), `"name": "licenses"`) || !strings.Contains(output.String(), `"passed": true`) {
		t.Fatalf("expected passing license check in output, got %q", output.String())
	}
}
