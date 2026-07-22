package adaptertest

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
)

func TestBasicNEIR(t *testing.T) {
	n := BasicNEIR()
	if n.Project.Name != "test-project" {
		t.Errorf("expected test-project, got %s", n.Project.Name)
	}
	if len(n.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(n.Services))
	}
	if len(n.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(n.Modules))
	}
}

func TestFullNEIR(t *testing.T) {
	n := FullNEIR()
	if n.Project == nil {
		t.Fatal("expected non-nil project")
	}
	if n.Architecture == nil {
		t.Error("expected non-nil architecture")
	}
	if n.Security == nil {
		t.Error("expected non-nil security")
	}
	if n.Deployment == nil {
		t.Error("expected non-nil deployment")
	}
	if n.Testing == nil {
		t.Error("expected non-nil testing")
	}
	if len(n.APIs) == 0 {
		t.Error("expected at least one API")
	}
	if len(n.Storage) == 0 {
		t.Error("expected at least one storage")
	}
}

func TestContains(t *testing.T) {
	if !Contains("hello world", "world") {
		t.Error("expected substring match")
	}
	if Contains("hello", "xyz") {
		t.Error("expected no match")
	}
	if !Contains("", "") {
		t.Error("expected empty match")
	}
}

func TestValidateCompilerOutput(t *testing.T) {
	t.Run("valid output", func(t *testing.T) {
		output := &compiler.CompiledOutput{
			Target:  "test",
			Summary: "compilation summary",
			Files: []compiler.OutputFile{
				{Path: "test.md", Content: "content", Kind: "docs"},
			},
		}
		ValidateCompilerOutput(t, output)
	})

}

func TestValidateArtifacts(t *testing.T) {
	artifacts := []engine.Artifact{
		{Path: "main.go", Content: []byte("package main")},
	}
	ValidateArtifacts(t, artifacts)
}

func TestAssertFileContains(t *testing.T) {
	artifacts := []engine.Artifact{
		{Path: "main.go", Content: []byte("package main\nfunc main() {}")},
	}

	if !AssertFileContains(t, artifacts, "main.go", "func main") {
		t.Error("expected to find func main")
	}
}
