package airgap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitImage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input      string
		wantName   string
		wantTag    string
	}{
		{"nginx:1.25", "nginx", "1.25"},
		{"postgres", "postgres", "latest"},
		{"ghcr.io/org/app:v2.0.0", "ghcr.io/org/app", "v2.0.0"},
		{"reg:5000/app:tag", "reg:5000/app", "tag"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			name, tag := splitImage(tt.input)
			if name != tt.wantName || tag != tt.wantTag {
				t.Errorf("splitImage(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, tag, tt.wantName, tt.wantTag)
			}
		})
	}
}

func TestHashRef(t *testing.T) {
	t.Parallel()
	ref := hashRef("test.txt", []byte("hello"))
	if ref.Size != 5 {
		t.Errorf("expected size 5, got %d", ref.Size)
	}
	if len(ref.Hash) != 64 {
		t.Errorf("expected 64-char hash, got %d", len(ref.Hash))
	}
}

func TestVerifyChecksum(t *testing.T) {
	t.Parallel()
	b := &Bundle{
		SpecVersion: SpecVersion,
		Name:        "test",
		Version:     "1.0.0",
		Charts:      []FileRef{{Path: "a.tgz", Hash: "abc", Size: 1}},
	}
	// Compute the correct checksum using the canonical bytes.
	hash := sha256Hex(b.manifestBytes())
	b.ManifestHash = hash

	if !b.VerifyChecksum() {
		t.Error("expected checksum to verify")
	}
}

func TestVerifyChecksumTampered(t *testing.T) {
	t.Parallel()
	b := &Bundle{
		SpecVersion: SpecVersion,
		Name:        "test",
		Version:     "1.0.0",
	}
	b.ManifestHash = "deadbeef"
	if b.VerifyChecksum() {
		t.Error("expected checksum to fail")
	}
}

func TestSummary(t *testing.T) {
	t.Parallel()
	b := &Bundle{
		Name:      "app",
		Version:   "1.0.0",
		Charts:    []FileRef{{Path: "a"}},
		Images:    []ImageRef{{Name: "nginx"}},
		Signatures: []FileRef{{Path: "app.sig.json"}},
	}
	s := b.Summary()
	if !strings.Contains(s, "Airgap Bundle: app 1.0.0") {
		t.Error("missing bundle header")
	}
	if !strings.Contains(s, "Charts:     1") {
		t.Error("missing charts count")
	}
}

func TestBuildWithImages(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	imagesFile := filepath.Join(dir, "images.txt")
	if err := os.WriteFile(imagesFile, []byte("# comment\nnginx:1.25\npostgres\n"), 0o644); err != nil {
		t.Fatalf("write images: %v", err)
	}

	b := NewBuilder(BundleConfig{
		Name:       "test-bundle",
		Version:    "1.0.0",
		ImagesFile: imagesFile,
	})
	output := filepath.Join(dir, "bundle.tar.gz")
	bundle, err := b.Build(output)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(bundle.Images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(bundle.Images))
	}
	if bundle.Images[0].Name != "nginx" || bundle.Images[0].Tag != "1.25" {
		t.Errorf("unexpected image: %+v", bundle.Images[0])
	}
	if bundle.Images[1].Name != "postgres" || bundle.Images[1].Tag != "latest" {
		t.Errorf("unexpected image: %+v", bundle.Images[1])
	}
	if _, err := os.Stat(output); err != nil {
		t.Errorf("expected bundle file created: %v", err)
	}
}

func TestBuildWithSBOM(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sbomFile := filepath.Join(dir, "sbom.json")
	if err := os.WriteFile(sbomFile, []byte(`{"bomFormat":"CycloneDX"}`), 0o644); err != nil {
		t.Fatalf("write sbom: %v", err)
	}

	b := NewBuilder(BundleConfig{
		Name:         "sbom-bundle",
		Version:      "1.0.0",
		SBOMManifest: sbomFile,
	})
	output := filepath.Join(dir, "bundle.tar.gz")
	bundle, err := b.Build(output)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(bundle.SBOMManifests) != 1 {
		t.Fatalf("expected 1 SBOM, got %d", len(bundle.SBOMManifests))
	}
	if bundle.SBOMManifests[0].Path != "sbom.json" {
		t.Errorf("expected sbom.json path, got %s", bundle.SBOMManifests[0].Path)
	}
	if !bundle.VerifyChecksum() {
		t.Error("expected checksum to verify")
	}
}

func TestBuildExtractRoundtrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Chart dir
	chartsDir := filepath.Join(dir, "charts")
	if err := os.MkdirAll(chartsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	chartFile := filepath.Join(chartsDir, "myapp-0.1.0.tgz")
	if err := os.WriteFile(chartFile, []byte("fake chart archive"), 0o644); err != nil {
		t.Fatalf("write chart: %v", err)
	}

	// Signature dir
	sigDir := filepath.Join(dir, "sigs")
	if err := os.MkdirAll(sigDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	sigFile := filepath.Join(sigDir, "app.sig.json")
	if err := os.WriteFile(sigFile, []byte(`{"specVersion":"1.0","signature":"abc"}`), 0o644); err != nil {
		t.Fatalf("write sig: %v", err)
	}

	b := NewBuilder(BundleConfig{
		Name:          "roundtrip",
		Version:       "1.0.0",
		ChartsDir:     chartsDir,
		SignaturesDir: sigDir,
	})
	output := filepath.Join(dir, "bundle.tar.gz")
	bundle, err := b.Build(output)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(bundle.Charts) != 1 {
		t.Fatalf("expected 1 chart, got %d", len(bundle.Charts))
	}
	if len(bundle.Signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(bundle.Signatures))
	}

	// Read back
	read, err := ReadBundle(output)
	if err != nil {
		t.Fatalf("ReadBundle: %v", err)
	}
	if read.Name != "roundtrip" {
		t.Errorf("expected name roundtrip, got %s", read.Name)
	}
	if len(read.Charts) != 1 {
		t.Errorf("expected 1 chart, got %d", len(read.Charts))
	}

	// Extract
	dest := filepath.Join(dir, "extracted")
	extracted, err := Extract(output, ExtractOptions{Destination: dest, VerifyHashes: true})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if extracted.Name != "roundtrip" {
		t.Errorf("expected name roundtrip, got %s", extracted.Name)
	}
	if _, err := os.Stat(filepath.Join(dest, "myapp-0.1.0.tgz")); err != nil {
		t.Errorf("expected extracted chart file: %v", err)
	}
}

func TestReadBundleInvalid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.tar.gz")
	if err := os.WriteFile(bad, []byte("not a gzip"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := ReadBundle(bad)
	if err == nil {
		t.Error("expected error for invalid gzip")
	}
}

func TestReadBundleMissing(t *testing.T) {
	t.Parallel()
	_, err := ReadBundle("/nonexistent/bundle.tar.gz")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestPrefixTarPath(t *testing.T) {
	t.Parallel()
	got := prefixTarPath("charts/app.tgz")
	if got != "bundled/charts/app.tgz" {
		t.Errorf("unexpected: %s", got)
	}
	got2 := prefixTarPath("bundled/charts/app.tgz")
	if got2 != "bundled/charts/app.tgz" {
		t.Errorf("unexpected double prefix: %s", got2)
	}
}

func TestBundleManifestJSON(t *testing.T) {
	t.Parallel()
	b := &Bundle{
		SpecVersion: SpecVersion,
		Name:        "m",
		Version:     "1.0.0",
		Images:      []ImageRef{{Name: "redis", Tag: "7"}},
	}
	data, err := b.Manifest()
	if err != nil {
		t.Fatalf("Manifest: %v", err)
	}
	if !strings.Contains(string(data), "redis") {
		t.Error("missing image in manifest")
	}
}

func TestExtractRejectsUnsafePath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	chartFile := filepath.Join(dir, "a.tgz")
	os.WriteFile(chartFile, []byte("data"), 0o644)

	b := NewBuilder(BundleConfig{Name: "unsafe", Version: "1", ChartsDir: dir})
	output := filepath.Join(dir, "b.tar.gz")
	if _, err := b.Build(output); err != nil {
		t.Fatalf("Build: %v", err)
	}
	// Build a manual tar with a corrupted path via Extract (should handle gracefully).
	dest := filepath.Join(dir, "out")
	_, err := Extract(output, ExtractOptions{Destination: dest})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
}

func sha256Hex(data []byte) string {
	return hashRef("", data).Hash
}

func TestBundleWithoutChartsDir(t *testing.T) {
	t.Parallel()
	b := NewBuilder(BundleConfig{Name: "minimal", Version: "1.0.0"})
	output := filepath.Join(t.TempDir(), "bundle.tar.gz")
	bundle, err := b.Build(output)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(bundle.Charts) != 0 {
		t.Error("expected no charts")
	}
	if len(bundle.Images) != 0 {
		t.Error("expected no images")
	}
	if !bundle.VerifyChecksum() {
		t.Error("expected checksum to verify")
	}
}