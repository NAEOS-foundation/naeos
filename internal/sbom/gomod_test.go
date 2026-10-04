// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sbom

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testGoMod = `module github.com/NAEOS-foundation/naeos

go 1.26.0

require (
	github.com/spf13/cobra v1.10.2
	golang.org/x/sys v0.47.0 // indirect
)

require gopkg.in/yaml.v3 v3.0.1
`

const testGoSum = `github.com/spf13/cobra v1.10.2 h1:abcd123=
github.com/spf13/cobra v1.10.2/go.mod h1:efgh=
golang.org/x/sys v0.47.0 h1:xyz456=
gopkg.in/yaml.v3 v3.0.1 h1:hash789=
`

func writeModuleTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(testGoMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), []byte(testGoSum), 0o644); err != nil {
		t.Fatalf("write go.sum: %v", err)
	}
	return dir
}

func TestParseGoMod(t *testing.T) {
	t.Parallel()
	refs := parseGoMod([]byte(testGoMod))
	if len(refs) != 3 {
		t.Fatalf("expected 3 module refs, got %d", len(refs))
	}
	if refs[0].Path != "github.com/spf13/cobra" || refs[0].Version != "v1.10.2" {
		t.Errorf("unexpected first ref: %+v", refs[0])
	}
	if refs[0].Indirect {
		t.Errorf("cobra should not be indirect: %+v", refs[0])
	}
	if !refs[1].Indirect {
		t.Errorf("golang.org/x/sys should be indirect: %+v", refs[1])
	}
	if refs[2].Path != "gopkg.in/yaml.v3" {
		t.Errorf("expected single-line require parsed: %+v", refs[2])
	}
}

func TestParseGoSum(t *testing.T) {
	t.Parallel()
	sums := parseGoSum([]byte(testGoSum))
	if len(sums) != 3 {
		t.Fatalf("expected 3 module sums, got %d", len(sums))
	}
	if got := sums["github.com/spf13/cobra@v1.10.2"]; got != "abcd123=" {
		t.Errorf("expected h1 content 'abcd123=', got %q", got)
	}
	if got := sums["golang.org/x/sys@v0.47.0"]; got != "xyz456=" {
		t.Errorf("expected x/sys hash, got %q", got)
	}
	if _, ok := sums["github.com/spf13/cobra@v1.10.2/go.mod"]; ok {
		t.Error("go.mod pseudo-entries should be excluded")
	}
}

func TestGeneratorFromGoModules(t *testing.T) {
	t.Parallel()
	dir := writeModuleTestDir(t)
	gen := NewGenerator(GeneratorConfig{
		Project:     "NAEOS",
		Version:     "3.5.0",
		ToolVersion: "3.5.0",
	})
	bom, err := gen.FromGoModules(dir)
	if err != nil {
		t.Fatalf("FromGoModules: %v", err)
	}
	if bom.ComponentCount() != 4 { // 3 modules + root component
		t.Errorf("expected 4 components, got %d", bom.ComponentCount())
	}

	byName := make(map[string]Component)
	for _, c := range bom.Components {
		byName[c.Name] = c
	}

	cobra, ok := byName["github.com/spf13/cobra"]
	if !ok {
		t.Fatal("expected cobra component")
	}
	if cobra.Type != Library {
		t.Errorf("expected library type, got %s", cobra.Type)
	}
	if len(cobra.Licenses) != 1 || cobra.Licenses[0].License == nil || cobra.Licenses[0].License.ID != "Apache-2.0" {
		t.Errorf("expected CycloneDX Apache-2.0 license choice, got %+v", cobra.Licenses)
	}
	if !strings.HasPrefix(cobra.Purl, "pkg:golang/github.com/spf13/cobra@v1.10.2") {
		t.Errorf("unexpected purl %q", cobra.Purl)
	}
	if len(cobra.Hashes) != 1 || cobra.Hashes[0].Alg != "H1" {
		t.Errorf("expected go.sum H1 hash, got %+v", cobra.Hashes)
	}

	if comp := byName["golang.org/x/sys"]; comp.Properties[0].Name != "go.module" || comp.Properties[0].Value != "indirect" {
		t.Errorf("expected indirect property, got %+v", comp.Properties)
	}
	if comp := byName["gopkg.in/yaml.v3"]; len(comp.Licenses) != 1 || comp.Licenses[0].Expression != "MIT AND Apache-2.0" {
		t.Errorf("expected CycloneDX SPDX license expression, got %+v", comp.Licenses)
	}
	if bom.Metadata.Component == nil || len(bom.Metadata.Component.Licenses) != 1 ||
		bom.Metadata.Component.Licenses[0].License == nil ||
		bom.Metadata.Component.Licenses[0].License.ID != "Apache-2.0" {
		t.Errorf("expected Apache-2.0 license on the root application, got %+v", bom.Metadata.Component)
	}

	data, err := Marshal(bom)
	if err != nil {
		t.Fatalf("Marshal SBOM: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode SBOM JSON: %v", err)
	}
	components := document["components"].([]any)
	for _, raw := range components {
		component := raw.(map[string]any)
		if _, ok := component["license"]; ok {
			t.Fatalf("component %q contains non-CycloneDX singular license property", component["name"])
		}
		if _, ok := component["licenses"].([]any); !ok {
			t.Fatalf("component %q has no CycloneDX licenses array", component["name"])
		}
	}
}

func TestGeneratorFromGoModulesMissingSum(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(testGoMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	gen := NewGenerator(GeneratorConfig{Project: "NAEOS", Version: "1.0.0"})
	bom, err := gen.FromGoModules(dir)
	if err != nil {
		t.Fatalf("FromGoModules without go.sum: %v", err)
	}
	// populateHash fills a deterministic hash when go.sum is absent.
	if len(bom.Components) == 0 {
		t.Fatal("expected components")
	}
	if len(bom.Components[0].Hashes) == 0 {
		t.Error("expected deterministic hash populated")
	}
}

func TestGeneratorFromGoModulesRejectsUnknownLicense(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n\ngo 1.26.0\n\nrequire example.com/unreviewed v1.2.3\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), nil, 0o644); err != nil {
		t.Fatalf("write go.sum: %v", err)
	}

	gen := NewGenerator(GeneratorConfig{Project: "NAEOS", Version: "1.0.0"})
	if _, err := gen.FromGoModules(dir); err == nil || !strings.Contains(err.Error(), "no audited SPDX license for Go module example.com/unreviewed@v1.2.3") {
		t.Fatalf("expected explicit missing-license error, got %v", err)
	}
}

func TestGoModuleLicensed(t *testing.T) {
	t.Parallel()
	if !GoModuleLicensed("github.com/spf13/cobra") {
		t.Error("expected cobra listed as licensed")
	}
	if GoModuleLicensed("example.invalid/module") {
		t.Error("expected unknown module not listed")
	}
	if got := goModuleLicense["github.com/DATA-DOG/go-sqlmock"]; got != "BSD-3-Clause" {
		t.Errorf("expected go-sqlmock license to match its upstream BSD-3-Clause text, got %q", got)
	}
}

func TestCurrentGoModulesHaveCuratedLicenses(t *testing.T) {
	t.Parallel()

	modData, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatalf("read repository go.mod: %v", err)
	}
	for _, ref := range parseGoMod(modData) {
		if !GoModuleLicensed(ref.Path) {
			t.Errorf("Go module %s@%s has no curated SPDX license", ref.Path, ref.Version)
		}
	}
}
