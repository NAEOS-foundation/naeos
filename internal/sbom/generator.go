package sbom

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// GeneratorConfig configures SBOM generation for a build output.
type GeneratorConfig struct {
	Project     string
	Version     string
	ToolName    string
	ToolVersion string
}

// Config validates the generator configuration and applies defaults.
func (c GeneratorConfig) Config() GeneratorConfig {
	if c.ToolName == "" {
		c.ToolName = "naeos"
	}
	return c
}

// ComponentOption mutates a Component during collection.
type ComponentOption func(*Component)

// WithPurl attaches a package URL to a component.
func WithPurl(purl string) ComponentOption {
	return func(c *Component) { c.Purl = purl }
}

// WithSupplier attaches a supplier string to a component.
func WithSupplier(sup string) ComponentOption {
	return func(c *Component) { c.Supplier = sup }
}

// WithProperty attaches a key/value property to a component.
func WithProperty(name, value string) ComponentOption {
	return func(c *Component) {
		c.Properties = append(c.Properties, Property{Name: name, Value: value})
	}
}

// Generator produces CycloneDX BOM documents from a set of components.
type Generator struct {
	cfg GeneratorConfig
}

// NewGenerator creates an SBOM generator with the given configuration.
func NewGenerator(cfg GeneratorConfig) *Generator {
	return &Generator{cfg: cfg.Config()}
}

// Generate builds a BOM for an application and its component inventory.
func (g *Generator) Generate(components []Component) (*BOM, error) {
	bom := NewBOM()
	bom.Metadata.Timestamp = Timestamp()
	bom.Metadata.Tools = []Tool{{
		Name:    g.cfg.ToolName,
		Version: g.cfg.ToolVersion,
	}}
	if g.cfg.Project != "" {
		bom.Metadata.Component = &Component{
			Type:  Application,
			Name:  g.cfg.Project,
			Version: g.cfg.Version,
			Purl:  Purl("pkg", g.cfg.Project, g.cfg.Version),
		}
	}

	for _, c := range components {
		c = populateHash(c)
		bom.Components = append(bom.Components, c)
	}

	sort.Slice(bom.Components, func(i, j int) bool {
		return bom.Components[i].Name < bom.Components[j].Name
	})
	if bom.Metadata.Component != nil {
		bom.Dependencies = []Dependency{{Ref: bom.Metadata.Component.Name}}
	}
	return bom, nil
}

// FromDir scans a directory tree and adds every file as a File component,
// computing a SHA-256 hash for each. This produces a file-level SBOM
// when no package manager inventory exists.
func (g *Generator) FromDir(root string) (*BOM, error) {
	var comps []Component
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h := sha256.Sum256(content)
		comps = append(comps, Component{
			Type:     File,
			Name:     filepath.Base(path),
			FileName: rel,
			Path:     path,
			Hashes: []Hash{{
				Alg: "SHA-256",
				Val: hex.EncodeToString(h[:]),
			}},
		})
		return nil
	})
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "walk %s", root)
	}
	return g.Generate(comps)
}

func populateHash(c Component) Component {
	if len(c.Hashes) > 0 {
		return c
	}
	key := c.FileName + ":" + c.Path + ":" + c.Name
	h := sha256.Sum256([]byte(key))
	c.Hashes = []Hash{{
		Alg: "SHA-256",
		Val: hex.EncodeToString(h[:]),
	}}
	return c
}

// Write persists a BOM to disk as indented JSON.
func Write(bom *BOM, path string) error {
	data, err := Marshal(bom)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create dir")
	}
	return os.WriteFile(path, data, 0o600)
}

// Marshal serializes a BOM to indented JSON.
func Marshal(bom *BOM) ([]byte, error) {
	return EncodeJSON(bom)
}

// Load reads a BOM from an indented JSON file.
func Load(path string) (*BOM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read %s", path)
	}
	bom, err := Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return bom, nil
}

// Unmarshal parses a BOM from JSON bytes.
func Unmarshal(data []byte) (*BOM, error) {
	bom := new(BOM)
	if err := DecodeJSON(data, bom); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse BOM")
	}
	if bom.BOMFormat != "CycloneDX" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "not a CycloneDX document")
	}
	return bom, nil
}

// Purl builds a minimal package-url string from a shared prefix.
func Purl(typ, name, version string) string {
	ns := strings.ToLower(name)
	return fmt.Sprintf("pkg:%s/%s@%s", typ, ns, version)
}
