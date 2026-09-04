package airgap

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// SpecVersion is the airgap bundle format version.
const SpecVersion = "1.0"

// Bundle is an air-gapped distribution bundle containing charts, images,
// SBOMs, and signatures required to deploy NAEOS offline.
type Bundle struct {
	SpecVersion string    `json:"specVersion"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedAt   string    `json:"createdAt"`
	Charts       []FileRef `json:"charts,omitempty"`
	Images       []ImageRef `json:"images,omitempty"`
	SBOMManifests []FileRef `json:"sboms,omitempty"`
	Signatures   []FileRef `json:"signatures,omitempty"`
	ManifestHash string  `json:"manifestHash"`
}

// FileRef references an embedded file in the bundle tar.
type FileRef struct {
	Path   string `json:"path"`
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Source string `json:"-"`
}

// ImageRef describes a container image that must be loaded in the
// air-gapped environment.
type ImageRef struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// BundleConfig configures bundle creation.
type BundleConfig struct {
	Name    string
	Version string
	// ChartsDir is scanned for *.tgz chart archives.
	ChartsDir string
	// ImagesFile lists image name:tag entries (one per line).
	ImagesFile string
	// SBOMManifest is a CycloneDX SBOM JSON file to include.
	SBOMManifest string
	// SignaturesDir is scanned for *.sig.json signature bundles.
	SignaturesDir string
}

// FileInfo reports the contents of a file during bundle building.
type FileInfo struct {
	Path string
	Size int64
}

// Builder creates air-gapped distribution bundles.
type Builder struct {
	cfg BundleConfig
}

// NewBuilder creates an airgap bundle builder.
func NewBuilder(cfg BundleConfig) *Builder {
	return &Builder{cfg: cfg}
}

// Build creates an airgap bundle tar.gz file at the given path.
func (b *Builder) Build(outputPath string) (*Bundle, error) {
	bundle := &Bundle{
		SpecVersion: SpecVersion,
		Name:        b.cfg.Name,
		Version:     b.cfg.Version,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := b.collectCharts(bundle); err != nil {
		return nil, err
	}
	if err := b.collectImages(bundle); err != nil {
		return nil, err
	}
	if err := b.collectSBOM(bundle); err != nil {
		return nil, err
	}
	if err := b.collectSignatures(bundle); err != nil {
		return nil, err
	}

	manifestData := bundle.manifestBytes()
	hash := sha256.Sum256(manifestData)
	bundle.ManifestHash = hex.EncodeToString(hash[:])

	if err := WriteBundle(bundle, outputPath); err != nil {
		return nil, err
	}
	return bundle, nil
}

func (b *Builder) collectCharts(bundle *Bundle) error {
	if b.cfg.ChartsDir == "" {
		return nil
	}
	return walkDirFunc(b.cfg.ChartsDir, []string{".tgz", ".tar.gz"}, func(path, rel string, content []byte) error {
		ref := hashRef(rel, content)
		ref.Source = path
		bundle.Charts = append(bundle.Charts, ref)
		return nil
	})
}

func (b *Builder) collectImages(bundle *Bundle) error {
	if b.cfg.ImagesFile == "" {
		return nil
	}
	data, err := os.ReadFile(b.cfg.ImagesFile)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "read images file")
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, tag := splitImage(line)
		bundle.Images = append(bundle.Images, ImageRef{Name: name, Tag: tag})
	}
	return nil
}

func (b *Builder) collectSBOM(bundle *Bundle) error {
	if b.cfg.SBOMManifest == "" {
		return nil
	}
	content, err := os.ReadFile(b.cfg.SBOMManifest)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "read SBOM")
	}
	ref := hashRef("sbom.json", content)
	ref.Source = b.cfg.SBOMManifest
	bundle.SBOMManifests = append(bundle.SBOMManifests, ref)
	return nil
}

func (b *Builder) collectSignatures(bundle *Bundle) error {
	if b.cfg.SignaturesDir == "" {
		return nil
	}
	return walkDirFunc(b.cfg.SignaturesDir, []string{".sig.json"}, func(path, rel string, content []byte) error {
		ref := hashRef(rel, content)
		ref.Source = path
		bundle.Signatures = append(bundle.Signatures, ref)
		return nil
	})
}

// Manifest returns the JSON representation of the manifest (indented).
func (b *Bundle) Manifest() ([]byte, error) {
	return json.MarshalIndent(b, "", "  ")
}

// manifestBytes returns the canonical compact JSON used for hashing.
func (b *Bundle) manifestBytes() []byte {
	copy := *b
	copy.ManifestHash = ""
	data, _ := json.Marshal(copy)
	return data
}

// VerifyChecksum recomputes the manifest hash and compares it.
func (b *Bundle) VerifyChecksum() bool {
	hash := sha256.Sum256(b.manifestBytes())
	return hex.EncodeToString(hash[:]) == b.ManifestHash
}

// Summary builds a human-readable summary of the bundle.
func (b *Bundle) Summary() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Airgap Bundle: %s %s\n", b.Name, b.Version)
	fmt.Fprintf(&sb, "  Format:     %s\n", b.SpecVersion)
	fmt.Fprintf(&sb, "  Created:    %s\n", b.CreatedAt)
	fmt.Fprintf(&sb, "  Charts:     %d\n", len(b.Charts))
	fmt.Fprintf(&sb, "  Images:     %d\n", len(b.Images))
	fmt.Fprintf(&sb, "  SBOMs:      %d\n", len(b.SBOMManifests))
	fmt.Fprintf(&sb, "  Signatures: %d\n", len(b.Signatures))
	return sb.String()
}

func hashRef(path string, content []byte) FileRef {
	h := sha256.Sum256(content)
	return FileRef{
		Path: path,
		Hash: hex.EncodeToString(h[:]),
		Size: int64(len(content)),
	}
}

func splitImage(line string) (string, string) {
	idx := strings.LastIndex(line, ":")
	if idx < 0 {
		return line, "latest"
	}
	return line[:idx], line[idx+1:]
}

func walkDirFunc(root string, exts []string, fn func(path, rel string, content []byte) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		matched := false
		for _, ext := range exts {
			if strings.HasSuffix(path, ext) {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		return fn(path, rel, content)
	})
}

// WriteBundle writes the manifest and embedded files to a tar.gz archive.
func WriteBundle(bundle *Bundle, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create %s", outputPath)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	// Collect all file references.
	type fileEntry struct {
		ref FileRef
	}
	var entries []fileEntry
	for _, c := range bundle.Charts {
		entries = append(entries, fileEntry{ref: c})
	}
	for _, s := range bundle.SBOMManifests {
		entries = append(entries, fileEntry{ref: s})
	}
	for _, s := range bundle.Signatures {
		entries = append(entries, fileEntry{ref: s})
	}

	// Sort for reproducible output.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ref.Path < entries[j].ref.Path
	})

	for _, e := range entries {
		var content []byte
		var err error
		if e.ref.Source != "" {
			content, err = os.ReadFile(e.ref.Source)
		} else {
			content, err = os.ReadFile(e.ref.Path)
		}
		if err != nil {
			continue
		}
		hdr := &tar.Header{
			Name:    prefixTarPath(e.ref.Path),
			Mode:    0o644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "write tar header")
		}
		if _, err := tw.Write(content); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "write tar data")
		}
	}

	manifestData, _ := bundle.Manifest()
	hdr := &tar.Header{
		Name:    "manifest.json",
		Mode:    0o644,
		Size:    int64(len(manifestData)),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "write manifest header")
	}
	if _, err := tw.Write(manifestData); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "write manifest data")
	}

	return nil
}

func prefixTarPath(p string) string {
	if strings.HasPrefix(p, "bundled/") {
		return p
	}
	return "bundled/" + strings.TrimPrefix(p, string(filepath.Separator))
}

// ReadBundle reads an airgap bundle tar.gz archive and extracts the manifest.
func ReadBundle(path string) (*Bundle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "open %s", path)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "gzip open")
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read tar")
		}
		if hdr.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read manifest")
			}
			bundle := new(Bundle)
			if err := json.Unmarshal(data, bundle); err != nil {
				return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse manifest")
			}
			return bundle, nil
		}
	}
	return nil, naeoserr.New(naeoserr.ErrNotFound, "manifest.json not found in bundle")
}