package adaptertest

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
)

func TestContains(t *testing.T) {
	if !Contains("hello world", "world") {
		t.Error("expected contains to be true")
	}
	if Contains("hello world", "xyz") {
		t.Error("expected contains to be false")
	}
}

func TestValidateCompilerOutput(t *testing.T) {
	output := &compiler.CompiledOutput{
		Target:  "go",
		Summary: "compiled successfully",
		Files:   []compiler.OutputFile{{Path: "main.go", Content: "package main"}},
	}
	ValidateCompilerOutput(t, output)
}

func TestValidateArtifacts(t *testing.T) {
	artifacts := []engine.Artifact{
		{Path: "dist/main.go", Content: []byte("package main")},
	}
	ValidateArtifacts(t, artifacts)
}

func TestAssertFileContains(t *testing.T) {
	artifacts := []engine.Artifact{
		{Path: "dist/main.go", Content: []byte("package main")},
	}
	if !AssertFileContains(t, artifacts, "dist/main.go", "package") {
		t.Error("expected file to contain 'package'")
	}
}

func TestBasicNEIR(t *testing.T) {
	neir := BasicNEIR()
	if neir.Project == nil || neir.Project.Name != "test-project" {
		t.Error("expected basic NEIR with test-project")
	}
	if len(neir.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(neir.Modules))
	}
}

func TestFullNEIR(t *testing.T) {
	neir := FullNEIR()
	if neir.Project == nil || neir.Project.Name != "test-project" {
		t.Error("expected full NEIR with test-project")
	}
	if neir.Architecture == nil {
		t.Error("expected architecture in full NEIR")
	}
	if neir.Infrastructure == nil {
		t.Error("expected infrastructure in full NEIR")
	}
}
