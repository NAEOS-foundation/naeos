package sbom

// ComponentType identifies the type of a software component.
type ComponentType string

const (
	Application ComponentType = "application"
	Library     ComponentType = "library"
	Framework   ComponentType = "framework"
	Container   ComponentType = "container"
	Platform    ComponentType = "platform"
	File        ComponentType = "file"
)

// SpecVersion is the CycloneDX specification version emitted.
const SpecVersion = "1.5"

// BOM is a CycloneDX Software Bill of Materials document.
type BOM struct {
	BOMFormat    string       `json:"bomFormat"`
	SpecVersion  string       `json:"specVersion"`
	SerialNumber string       `json:"serialNumber,omitempty"`
	Version      int          `json:"version"`
	Metadata     Metadata     `json:"metadata,omitempty"`
	Components   []Component  `json:"components,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
}

// Metadata captures document-level information about the BOM.
type Metadata struct {
	Timestamp  string         `json:"timestamp,omitempty"`
	Tools      []Tool         `json:"tools,omitempty"`
	Component  *Component     `json:"component,omitempty"`
	Properties []Property     `json:"properties,omitempty"`
}

// Tool describes the tool that produced the BOM.
type Tool struct {
	Vendor  string `json:"vendor,omitempty"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Component is a software element tracked in the BOM.
type Component struct {
	Type       ComponentType  `json:"type"`
	Name       string         `json:"name"`
	BOMRef     string         `json:"bom-ref,omitempty"`
	Group      string         `json:"group,omitempty"`
	Version    string         `json:"version,omitempty"`
	Supplier   string         `json:"supplier,omitempty"`
	License    string         `json:"license,omitempty"`
	Hashes     []Hash         `json:"hashes,omitempty"`
	Purl       string         `json:"purl,omitempty"`
	FileName   string         `json:"fileName,omitempty"`
	Path       string         `json:"path,omitempty"`
	Properties []Property     `json:"properties,omitempty"`
}

// Property is an arbitrary key-value annotation.
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Hash is a checksum over component content.
type Hash struct {
	Alg string `json:"alg"`
	Val string `json:"content"`
}

// Dependency records a component-to-component relationship in the BOM graph.
type Dependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

// NewBOM creates a new CycloneDX BOM document with a fresh serial number.
func NewBOM() *BOM {
	return &BOM{
		BOMFormat:   "CycloneDX",
		SpecVersion: SpecVersion,
		SerialNumber: NewSerialNumber(),
		Version:     1,
	}
}

// ComponentCount returns the number of components in a BOM, optionally
// including the top-level metadata component.
func (b *BOM) ComponentCount() int {
	n := len(b.Components)
	if b.Metadata.Component != nil {
		n++
	}
	return n
}
