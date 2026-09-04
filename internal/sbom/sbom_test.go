package sbom

import (
	"testing"
)

func TestNewBOM(t *testing.T) {
	t.Parallel()
	bom := NewBOM()
	if bom.BOMFormat != "CycloneDX" {
		t.Errorf("expected bomFormat CycloneDX, got %s", bom.BOMFormat)
	}
	if bom.SpecVersion != SpecVersion {
		t.Errorf("expected specVersion %s, got %s", SpecVersion, bom.SpecVersion)
	}
	if bom.Version != 1 {
		t.Errorf("expected version 1, got %d", bom.Version)
	}
	if bom.SerialNumber == "" {
		t.Error("expected non-empty serial number")
	}
}

func TestNewSerialNumber(t *testing.T) {
	t.Parallel()
	sn := NewSerialNumber()
	if len(sn) != 45 { // urn:uuid: + 36 chars
		t.Errorf("unexpected serial number length: %d", len(sn))
	}
	sn2 := NewSerialNumber()
	if sn == sn2 {
		t.Error("expected different serial numbers")
	}
}

func TestTimestamp(t *testing.T) {
	t.Parallel()
	ts := Timestamp()
	if ts == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestPurl(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		typ     string
		pkgName string
		version string
		want    string
	}{
		{
			name:    "application",
			typ:     "pkg",
			pkgName: "myapp",
			version: "1.0.0",
			want:    "pkg:pkg/myapp@1.0.0",
		},
		{
			name:    "library",
			typ:     "pkg",
			pkgName: "MyLib",
			version: "2.1.0",
			want:    "pkg:pkg/mylib@2.1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Purl(tt.typ, tt.pkgName, tt.version)
			if got != tt.want {
				t.Errorf("Purl(%q, %q, %q) = %q, want %q", tt.typ, tt.pkgName, tt.version, got, tt.want)
			}
		})
	}
}

func TestGeneratorConfigDefaults(t *testing.T) {
	t.Parallel()
	cfg := GeneratorConfig{}.Config()
	if cfg.ToolName != "naeos" {
		t.Errorf("expected default tool name 'naeos', got %q", cfg.ToolName)
	}
}

func TestGenerateBasic(t *testing.T) {
	t.Parallel()
	gen := NewGenerator(GeneratorConfig{
		Project:     "myproject",
		Version:     "1.0.0",
		ToolVersion: "3.3.0",
	})
	bom, err := gen.Generate([]Component{
		{Name: "foo", Type: Library, Version: "0.1.0", Purl: Purl("pkg", "foo", "0.1.0")},
		{Name: "bar", Type: Library, Version: "0.2.0"},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if bom.Metadata.Component == nil {
		t.Fatal("expected metadata component")
	}
	if bom.Metadata.Component.Name != "myproject" {
		t.Errorf("expected metadata component name 'myproject', got %q", bom.Metadata.Component.Name)
	}
	if bom.ComponentCount() != 3 {
		t.Errorf("expected 3 components (2 + metadata), got %d", bom.ComponentCount())
	}
	if len(bom.Metadata.Tools) != 1 || bom.Metadata.Tools[0].Name != "naeos" {
		t.Error("expected naeos tool in metadata")
	}
	// Components are sorted by name.
	if bom.Components[0].Name != "bar" {
		t.Errorf("expected sorted first component 'bar', got %q", bom.Components[0].Name)
	}
	if bom.Components[1].Name != "foo" {
		t.Errorf("expected sorted second component 'foo', got %q", bom.Components[1].Name)
	}
}

func TestGenerateEmpty(t *testing.T) {
	t.Parallel()
	gen := NewGenerator(GeneratorConfig{})
	bom, err := gen.Generate(nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if bom.ComponentCount() != 0 {
		t.Errorf("expected 0 components, got %d", bom.ComponentCount())
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	gen := NewGenerator(GeneratorConfig{Project: "p", Version: "1"})
	bom, err := gen.Generate([]Component{
		{Name: "comp", Type: Library},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, err := Marshal(bom)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	bom2, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if bom2.ComponentCount() != 2 {
		t.Errorf("expected 2 components, got %d", bom2.ComponentCount())
	}
}

func TestUnmarshalInvalidFormat(t *testing.T) {
	t.Parallel()
	_, err := Unmarshal([]byte(`{"bomFormat":"SPDX"}`))
	if err == nil {
		t.Error("expected error for non-CycloneDX document")
	}
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := Unmarshal([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGenerateHashPopulated(t *testing.T) {
	t.Parallel()
	gen := NewGenerator(GeneratorConfig{})
	bom, err := gen.Generate([]Component{
		{Name: "h", Type: Library, FileName: "h.go", Path: "/h.go"},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	c := bom.Components[0]
	if len(c.Hashes) == 0 {
		t.Error("expected hash to be populated")
	}
	if c.Hashes[0].Alg != "SHA-256" {
		t.Errorf("expected SHA-256, got %s", c.Hashes[0].Alg)
	}
}

func TestGenerateHashPreserved(t *testing.T) {
	t.Parallel()
	gen := NewGenerator(GeneratorConfig{})
	custom := Hash{Alg: "SHA-512", Val: "abc123"}
	gen2 := gen
	bom, err := gen2.Generate([]Component{
		{Name: "h", Type: Library, Hashes: []Hash{custom}},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	c := bom.Components[0]
	if len(c.Hashes) != 1 || c.Hashes[0].Alg != "SHA-512" {
		t.Error("expected custom hash preserved")
	}
}

func TestGenerateDependencies(t *testing.T) {
	t.Parallel()
	gen := NewGenerator(GeneratorConfig{Project: "myapp"})
	bom, err := gen.Generate([]Component{
		{Name: "x", Type: Library},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(bom.Dependencies) != 1 || bom.Dependencies[0].Ref != "myapp" {
		t.Error("expected root dependency on metadata component")
	}
}

func TestComponentOptions(t *testing.T) {
	t.Parallel()
	c := Component{Name: "c", Type: Library}
	WithPurl("pkg:pkg/c@1")(&c)
	WithSupplier("TestOrg")(&c)
	WithProperty("k", "v")(&c)

	gen := NewGenerator(GeneratorConfig{})
	bom, err := gen.Generate([]Component{c})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	got := bom.Components[0]
	if got.Purl != "pkg:pkg/c@1" {
		t.Errorf("expected purl, got %q", got.Purl)
	}
	if got.Supplier != "TestOrg" {
		t.Errorf("expected supplier, got %q", got.Supplier)
	}
	if len(got.Properties) != 1 || got.Properties[0].Name != "k" {
		t.Error("expected property")
	}
}

func TestComponentCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		bom      *BOM
		expected int
	}{
		{
			name:     "no components",
			bom:      &BOM{},
			expected: 0,
		},
		{
			name: "with metadata component",
			bom: &BOM{
				Metadata: Metadata{Component: &Component{Name: "root"}},
				Components: []Component{
					{Name: "a"},
					{Name: "b"},
				},
			},
			expected: 3,
		},
		{
			name: "without metadata component",
			bom: &BOM{
				Components: []Component{{Name: "x"}},
			},
			expected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.bom.ComponentCount()
			if got != tt.expected {
				t.Errorf("ComponentCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPopulateHash(t *testing.T) {
	t.Parallel()
	c := Component{Name: "test"}
	got := populateHash(c)
	if len(got.Hashes) != 1 || got.Hashes[0].Alg != "SHA-256" {
		t.Error("expected SHA-256 hash populated")
	}
	// Same input should produce same hash (deterministic).
	got2 := populateHash(c)
	if got.Hashes[0].Val != got2.Hashes[0].Val {
		t.Error("expected deterministic hash")
	}
}

func TestPopulateHashExisting(t *testing.T) {
	t.Parallel()
	existing := Hash{Alg: "SHA-512", Val: "existing"}
	c := Component{Name: "test", Hashes: []Hash{existing}}
	got := populateHash(c)
	if len(got.Hashes) != 1 || got.Hashes[0].Alg != "SHA-512" {
		t.Error("expected existing hash preserved")
	}
}

func TestEncodeJSON(t *testing.T) {
	t.Parallel()
	data, err := EncodeJSON(map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("EncodeJSON: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
}
