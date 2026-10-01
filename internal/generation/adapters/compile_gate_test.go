// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
)

// writeArtifacts materializes generated artifacts under dir.
func writeArtifacts(t *testing.T, dir string, artifacts []engine.Artifact) {
	t.Helper()
	for _, a := range artifacts {
		path := filepath.Join(dir, filepath.FromSlash(a.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, a.Content, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}

// goBuild runs `go build ./...` inside dir using the module cache already
// populated by the parent NAEOS module. GOFLAGS=-mod=mod lets go.sum be
// completed from the local cache without network access, and GOPROXY=off
// guarantees the generated tree cannot silently resolve its own imports from
// the network.
func goBuild(t *testing.T, dir string) error {
	t.Helper()
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOPROXY=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &buildError{output: string(out), err: err}
	}
	return nil
}

type buildError struct {
	output string
	err    error
}

func (e *buildError) Error() string {
	return e.err.Error() + ": " + e.output
}

// requireGoToolchain skips when no usable go binary is on PATH.
func requireGoToolchain(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "js" || runtime.GOOS == "plan9" {
		t.Skip("no go toolchain for " + runtime.GOOS)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
}

// seedGoMod replaces the generated go.mod with one that pins the module path and
// declares the yaml dependency, then seeds go.sum from the NAEOS module so the
// build resolves entirely from the local module cache.
func seedGoMod(t *testing.T, dir, modulePath string) {
	t.Helper()

	gomod := "module " + modulePath + "\n\ngo 1.22\n\nrequire gopkg.in/yaml.v3 v3.0.1\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	seed, err := os.ReadFile(filepath.Join(repoRoot(t), "go.sum"))
	if err != nil {
		t.Fatalf("read repo go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), seed, 0o644); err != nil {
		t.Fatalf("write go.sum: %v", err)
	}
}

// repoRoot walks up from the test's working directory to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root")
		}
		dir = parent
	}
}

// TestGeneratedGoOutputCompiles is the compile gate for the Go generator.
//
// Generated artifacts used to be emitted without ever being compiled, so the
// Go adapter shipped an entrypoint that called a two-value function in a
// one-value assignment and imported a module path that no adapter ever wrote.
// Both shipped to users because nothing in the pipeline or its tests compiled
// the result. This test generates a project for representative specifications
// and compiles the output.
func TestGeneratedGoOutputCompiles(t *testing.T) {
	requireGoToolchain(t)

	cases := []struct {
		name       string
		project    string
		moduleName string
		modulePath string
	}{
		{
			name:       "declared module path outside internal",
			project:    "demo-app",
			moduleName: "auth",
			modulePath: "./auth",
		},
		{
			name:       "declared module path under internal",
			project:    "demo-app",
			moduleName: "core",
			modulePath: "./internal/core",
		},
		{
			name:       "bare module name without path prefix",
			project:    "demo-app",
			moduleName: "core",
			modulePath: "core",
		},
		{
			name:       "multi word module name",
			project:    "demo-app",
			moduleName: "Order Service",
			modulePath: "./order-service",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := GoAdapter{}

			var artifacts []engine.Artifact
			artifacts = append(artifacts, adapter.GenerateProject(tc.project, ModuleRef{
				Name: tc.moduleName,
				Path: tc.modulePath,
			})...)
			artifacts = append(artifacts, adapter.GenerateModule(tc.moduleName, tc.modulePath, tc.project)...)

			dir := t.TempDir()
			writeArtifacts(t, dir, artifacts)

			seedGoMod(t, dir, "github.com/example/"+slugOf(tc.project))

			if err := goBuild(t, dir); err != nil {
				t.Fatalf("generated Go project must compile: %v", err)
			}
		})
	}
}

// TestGeneratedGoOutputCompilesWithoutModule covers the scaffold path, where the
// entrypoint is generated before any module is known. It must not import a
// module that does not exist.
func TestGeneratedGoOutputCompilesWithoutModule(t *testing.T) {
	requireGoToolchain(t)

	adapter := GoAdapter{}
	artifacts := adapter.GenerateProject("standalone-app")

	dir := t.TempDir()
	writeArtifacts(t, dir, artifacts)
	seedGoMod(t, dir, "github.com/example/standalone-app")

	if err := goBuild(t, dir); err != nil {
		t.Fatalf("generated Go project must compile without a module: %v", err)
	}

	main, err := os.ReadFile(filepath.Join(dir, "cmd", "app", "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if strings.Contains(string(main), "/internal/core") {
		t.Error("standalone entrypoint must not import a generated module")
	}
}

// TestGeneratedGoEntrypointHandlesConfigError pins the two-value Load contract.
func TestGeneratedGoEntrypointHandlesConfigError(t *testing.T) {
	adapter := GoAdapter{}

	var main string
	for _, a := range adapter.GenerateProject("demo-app", ModuleRef{Name: "core", Path: "./core"}) {
		if a.Path == "cmd/app/main.go" {
			main = string(a.Content)
		}
	}
	if main == "" {
		t.Fatal("cmd/app/main.go was not generated")
	}

	if !strings.Contains(main, "cfg, err := coreconfig.Load(") {
		t.Error("entrypoint must handle the error returned by config.Load")
	}
	if !strings.Contains(main, "log.Fatal") {
		t.Error("entrypoint must fail closed when config cannot be loaded")
	}
}

func slugOf(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '_', r == '/', r == '-':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
