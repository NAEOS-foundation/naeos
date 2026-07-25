package hcl

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Spec defines the top-level configuration structure parsed from HCL files.
type Spec struct {
	Project  Project            `json:"project"`
	Services map[string]Service `json:"services"`
	Infra    Infra              `json:"infra"`
}

// Project holds project-level metadata.
type Project struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Service defines a single service declaration.
type Service struct {
	Image    string            `json:"image,omitempty"`
	Port     int               `json:"port,omitempty"`
	Type     string            `json:"type"`
	Env      map[string]string `json:"env,omitempty"`
	Volumes  []string          `json:"volumes,omitempty"`
	Depends  []string          `json:"depends,omitempty"`
	Replicas int               `json:"replicas,omitempty"`
}

// Infra describes infrastructure configuration.
type Infra struct {
	Engine string `json:"engine,omitempty"`
}

// ---------------------------------------------------------------------------
// ParseError: rich error type with line, column, and context
// ---------------------------------------------------------------------------

// ParseError represents an error encountered during HCL parsing, including
// the filename, line number, column, and surrounding context.
type ParseError struct {
	FileName string
	Line     int
	Column   int
	Message  string
	Context  string
}

func (e *ParseError) Error() string {
	if e.FileName != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.FileName, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("line %d col %d: %s", e.Line, e.Column, e.Message)
}

func newParseError(filename string, line, col int, msg, ctx string) *ParseError {
	return &ParseError{FileName: filename, Line: line, Column: col, Message: msg, Context: ctx}
}

// ---------------------------------------------------------------------------
// ParseFile / Parse (original logic)
// ---------------------------------------------------------------------------

// ParseFile reads an HCL file from disk and returns a parsed Spec.
func ParseFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "read file")
	}
	return Parse(data, path)
}

var blockRe = regexp.MustCompile(`^(\w+)\s+"([^"]*)"\s*\{`)
var kvRe = regexp.MustCompile(`^\s*(\w+)\s*=\s*(.+)$`)

// Parse converts raw HCL bytes into a Spec. It supports project, service,
// infra, and nested env{} / volumes / depends sub-blocks inside services.
func Parse(data []byte, filename string) (*Spec, error) {
	spec := &Spec{Services: make(map[string]Service)}
	lines := strings.Split(string(data), "\n")

	var currentBlock string
	var currentLabel string
	var currentSub string
	var inService string
	var errors_ []*ParseError

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if trimmed == "}" {
			if currentSub != "" {
				currentSub = ""
			} else {
				currentBlock = ""
				currentLabel = ""
				inService = ""
			}
			continue
		}

		if m := blockRe.FindStringSubmatch(trimmed); m != nil {
			blockType := m[1]
			blockLabel := m[2]
			if currentBlock == "service" && blockType == "env" {
				currentSub = "env"
				continue
			}
			currentBlock = blockType
			currentLabel = blockLabel
			switch currentBlock {
			case "project":
				spec.Project.Name = currentLabel
			case "service":
				inService = currentLabel
			}
			continue
		}

		if m := kvRe.FindStringSubmatch(trimmed); m != nil {
			col := strings.Index(line, m[1]) + 1
			key := m[1]
			val := strings.TrimSpace(m[2])
			val = strings.Trim(val, `"`)

			if currentSub == "env" && inService != "" {
				svc := spec.Services[inService]
				if svc.Env == nil {
					svc.Env = make(map[string]string)
				}
				svc.Env[key] = val
				spec.Services[inService] = svc
				continue
			}

			switch currentBlock {
			case "project":
				switch key {
				case "version":
					spec.Project.Version = val
				case "description":
					spec.Project.Description = val
				case "name":
					spec.Project.Name = val
				}
			case "service":
				svc := spec.Services[currentLabel]
				switch key {
				case "image":
					svc.Image = val
				case "port":
					p, err := strconv.Atoi(val)
					if err != nil {
						errors_ = append(errors_, newParseError(filename, lineNum, col,
							fmt.Sprintf("invalid port value %q", val), trimmed))
					} else {
						svc.Port = p
					}
				case "type":
					svc.Type = val
				case "replicas":
					r, err := strconv.Atoi(val)
					if err != nil {
						errors_ = append(errors_, newParseError(filename, lineNum, col,
							fmt.Sprintf("invalid replicas value %q", val), trimmed))
					} else {
						svc.Replicas = r
					}
				case "volumes":
					if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
						inner := strings.Trim(val, "[]")
						for _, v := range strings.Split(inner, ",") {
							v = strings.TrimSpace(v)
							v = strings.Trim(v, `"`)
							if v != "" {
								svc.Volumes = append(svc.Volumes, v)
							}
						}
					} else {
						svc.Volumes = append(svc.Volumes, val)
					}
				case "depends":
					if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
						inner := strings.Trim(val, "[]")
						for _, d := range strings.Split(inner, ",") {
							d = strings.TrimSpace(d)
							d = strings.Trim(d, `"`)
							if d != "" {
								svc.Depends = append(svc.Depends, d)
							}
						}
					} else {
						svc.Depends = append(svc.Depends, val)
					}
				}
				spec.Services[currentLabel] = svc
			case "infra":
				switch key {
				case "engine":
					spec.Infra.Engine = val
				}
			}
			continue
		}

		if currentBlock != "" {
			col := 1
			errors_ = append(errors_, newParseError(filename, lineNum, col,
				fmt.Sprintf("unexpected token in %s block", currentBlock), trimmed))
		}
	}

	if len(errors_) > 0 {
		var msgs []string
		for _, e := range errors_ {
			msgs = append(msgs, e.Error())
		}
		return spec, naeoserr.New(naeoserr.ErrParse, fmt.Sprintf("parse errors:\n%s", strings.Join(msgs, "\n")))
	}

	return spec, nil
}

// ---------------------------------------------------------------------------
// ToYAML (original logic)
// ---------------------------------------------------------------------------

// ToYAML serializes a Spec to a simple YAML-like representation.
func ToYAML(spec *Spec) ([]byte, error) {
	var out []byte
	out = append(out, []byte("project:\n")...)
	out = append(out, []byte(fmt.Sprintf("  name: %s\n", spec.Project.Name))...)
	if spec.Project.Version != "" {
		out = append(out, []byte(fmt.Sprintf("  version: %s\n", spec.Project.Version))...)
	}
	if spec.Project.Description != "" {
		out = append(out, []byte(fmt.Sprintf("  description: %s\n", spec.Project.Description))...)
	}

	if len(spec.Services) > 0 {
		out = append(out, []byte("services:\n")...)
		for name, svc := range spec.Services {
			out = append(out, []byte(fmt.Sprintf("  - name: %s\n", name))...)
			if svc.Image != "" {
				out = append(out, []byte(fmt.Sprintf("    image: %s\n", svc.Image))...)
			}
			if svc.Port != 0 {
				out = append(out, []byte(fmt.Sprintf("    port: %d\n", svc.Port))...)
			}
			if svc.Type != "" {
				out = append(out, []byte(fmt.Sprintf("    type: %s\n", svc.Type))...)
			}
		}
	}

	if spec.Infra.Engine != "" {
		out = append(out, []byte("infra:\n")...)
		out = append(out, []byte(fmt.Sprintf("  engine: %s\n", spec.Infra.Engine))...)
	}

	return out, nil
}

// ---------------------------------------------------------------------------
// Validate
// ---------------------------------------------------------------------------

// Validate checks a Spec for common issues and returns a list of errors.
// Checks performed: missing project name, duplicate services (already a map so
// this catches map collisions at parse time), port range validation, invalid
// replica counts, and missing service types.
func Validate(spec *Spec) []ParseError {
	var errs []ParseError

	if spec.Project.Name == "" {
		errs = append(errs, ParseError{Message: "project name is required"})
	}

	seen := make(map[string]bool)
	for name, svc := range spec.Services {
		if seen[name] {
			errs = append(errs, ParseError{
				Message: fmt.Sprintf("duplicate service name: %s", name),
			})
		}
		seen[name] = true

		if svc.Port < 0 || svc.Port > 65535 {
			errs = append(errs, ParseError{
				Message: fmt.Sprintf("service %q has invalid port %d (must be 0-65535)", name, svc.Port),
			})
		}

		if svc.Type == "" {
			errs = append(errs, ParseError{
				Message: fmt.Sprintf("service %q is missing a type", name),
			})
		}

		if svc.Replicas < 0 {
			errs = append(errs, ParseError{
				Message: fmt.Sprintf("service %q has negative replica count %d", name, svc.Replicas),
			})
		}

		for _, dep := range svc.Depends {
			if dep == name {
				errs = append(errs, ParseError{
					Message: fmt.Sprintf("service %q depends on itself", name),
				})
			}
			if _, ok := spec.Services[dep]; !ok {
				errs = append(errs, ParseError{
					Message: fmt.Sprintf("service %q depends on unknown service %q", name, dep),
				})
			}
		}
	}

	return errs
}

// ---------------------------------------------------------------------------
// ToJSON
// ---------------------------------------------------------------------------

// ToJSON serializes a Spec to indented JSON.
func ToJSON(spec *Spec) ([]byte, error) {
	return json.MarshalIndent(spec, "", "  ")
}

// ---------------------------------------------------------------------------
// ToHCL: round-trip a Spec back to HCL text
// ---------------------------------------------------------------------------

// ToHCL converts a Spec into a string that can be re-parsed by Parse.
func ToHCL(spec *Spec) string {
	var b strings.Builder

	fmt.Fprintf(&b, `project "%s" {`+"\n", spec.Project.Name)
	if spec.Project.Version != "" {
		fmt.Fprintf(&b, "  version = %q\n", spec.Project.Version)
	}
	if spec.Project.Description != "" {
		fmt.Fprintf(&b, "  description = %q\n", spec.Project.Description)
	}
	b.WriteString("}\n\n")

	names := sortedKeys(spec.Services)
	for _, name := range names {
		svc := spec.Services[name]
		fmt.Fprintf(&b, `service "%s" {`+"\n", name)
		if svc.Image != "" {
			fmt.Fprintf(&b, "  image = %q\n", svc.Image)
		}
		if svc.Port != 0 {
			fmt.Fprintf(&b, "  port = %d\n", svc.Port)
		}
		if svc.Type != "" {
			fmt.Fprintf(&b, "  type = %q\n", svc.Type)
		}
		if svc.Replicas > 0 {
			fmt.Fprintf(&b, "  replicas = %d\n", svc.Replicas)
		}
		if len(svc.Volumes) > 0 {
			fmt.Fprintf(&b, "  volumes = [%s]\n", formatStringSlice(svc.Volumes))
		}
		if len(svc.Depends) > 0 {
			fmt.Fprintf(&b, "  depends = [%s]\n", formatStringSlice(svc.Depends))
		}
		if len(svc.Env) > 0 {
			b.WriteString("  env \"vars\" {\n")
			envKeys := make([]string, 0, len(svc.Env))
			for k := range svc.Env {
				envKeys = append(envKeys, k)
			}
			sort.Strings(envKeys)
			for _, k := range envKeys {
				fmt.Fprintf(&b, "    %s = %q\n", k, svc.Env[k])
			}
			b.WriteString("  }\n")
		}
		b.WriteString("}\n\n")
	}

	if spec.Infra.Engine != "" {
		b.WriteString("infra \"main\" {\n")
		fmt.Fprintf(&b, "  engine = %q\n", spec.Infra.Engine)
		b.WriteString("}\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// ---------------------------------------------------------------------------
// MergeSpecs
// ---------------------------------------------------------------------------

// MergeSpecs merges src into dst, returning a new Spec. For conflicts the
// non-zero value from src wins. Service maps are unioned; when both specs
// define the same service name the fields from src take precedence.
func MergeSpecs(dst, src *Spec) *Spec {
	out := &Spec{
		Project:  dst.Project,
		Services: make(map[string]Service),
		Infra:    dst.Infra,
	}

	if src.Project.Name != "" {
		out.Project.Name = src.Project.Name
	}
	if src.Project.Version != "" {
		out.Project.Version = src.Project.Version
	}
	if src.Project.Description != "" {
		out.Project.Description = src.Project.Description
	}

	for name, svc := range dst.Services {
		out.Services[name] = svc
	}
	for name, srcSvc := range src.Services {
		if dstSvc, ok := out.Services[name]; ok {
			merged := mergeServices(dstSvc, srcSvc)
			out.Services[name] = merged
		} else {
			out.Services[name] = srcSvc
		}
	}

	if src.Infra.Engine != "" {
		out.Infra.Engine = src.Infra.Engine
	}

	return out
}

func mergeServices(dst, src Service) Service {
	out := dst
	if src.Image != "" {
		out.Image = src.Image
	}
	if src.Port != 0 {
		out.Port = src.Port
	}
	if src.Type != "" {
		out.Type = src.Type
	}
	if src.Replicas != 0 {
		out.Replicas = src.Replicas
	}

	if len(src.Volumes) > 0 {
		out.Volumes = append(out.Volumes, src.Volumes...)
	}

	if len(src.Depends) > 0 {
		out.Depends = append(out.Depends, src.Depends...)
	}

	if len(src.Env) > 0 {
		if out.Env == nil {
			out.Env = make(map[string]string)
		}
		for k, v := range src.Env {
			out.Env[k] = v
		}
	}

	return out
}

// ---------------------------------------------------------------------------
// SpecDiff
// ---------------------------------------------------------------------------

// DiffKind describes the type of difference.
type DiffKind string

const (
	DiffAdded   DiffKind = "added"
	DiffRemoved DiffKind = "removed"
	DiffChanged DiffKind = "changed"
)

// DiffEntry represents a single difference between two Specs.
type DiffEntry struct {
	Kind     DiffKind
	Path     string
	OldValue string
	NewValue string
}

func (d DiffEntry) String() string {
	switch d.Kind {
	case DiffAdded:
		return fmt.Sprintf("+ %s = %s", d.Path, d.NewValue)
	case DiffRemoved:
		return fmt.Sprintf("- %s = %s", d.Path, d.OldValue)
	case DiffChanged:
		return fmt.Sprintf("~ %s: %s -> %s", d.Path, d.OldValue, d.NewValue)
	}
	return ""
}

// SpecDiff compares two Specs and returns a list of differences.
func SpecDiff(a, b *Spec) []DiffEntry {
	var diffs []DiffEntry

	if a.Project.Name != b.Project.Name {
		diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: "project.name",
			OldValue: a.Project.Name, NewValue: b.Project.Name})
	}
	if a.Project.Version != b.Project.Version {
		diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: "project.version",
			OldValue: a.Project.Version, NewValue: b.Project.Version})
	}
	if a.Project.Description != b.Project.Description {
		diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: "project.description",
			OldValue: a.Project.Description, NewValue: b.Project.Description})
	}

	if a.Infra.Engine != b.Infra.Engine {
		diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: "infra.engine",
			OldValue: a.Infra.Engine, NewValue: b.Infra.Engine})
	}

	allNames := make(map[string]bool)
	for n := range a.Services {
		allNames[n] = true
	}
	for n := range b.Services {
		allNames[n] = true
	}

	sorted := make([]string, 0, len(allNames))
	for n := range allNames {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)

	for _, name := range sorted {
		svcA, okA := a.Services[name]
		svcB, okB := b.Services[name]
		prefix := "services." + name

		if okA && !okB {
			diffs = append(diffs, DiffEntry{Kind: DiffRemoved, Path: prefix, OldValue: "(exists)"})
			continue
		}
		if !okA && okB {
			diffs = append(diffs, DiffEntry{Kind: DiffAdded, Path: prefix, NewValue: "(exists)"})
			continue
		}

		if svcA.Image != svcB.Image {
			diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: prefix + ".image",
				OldValue: svcA.Image, NewValue: svcB.Image})
		}
		if svcA.Port != svcB.Port {
			diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: prefix + ".port",
				OldValue: strconv.Itoa(svcA.Port), NewValue: strconv.Itoa(svcB.Port)})
		}
		if svcA.Type != svcB.Type {
			diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: prefix + ".type",
				OldValue: svcA.Type, NewValue: svcB.Type})
		}
		if svcA.Replicas != svcB.Replicas {
			diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: prefix + ".replicas",
				OldValue: strconv.Itoa(svcA.Replicas), NewValue: strconv.Itoa(svcB.Replicas)})
		}
		if !stringSliceEqual(svcA.Volumes, svcB.Volumes) {
			diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: prefix + ".volumes",
				OldValue: strings.Join(svcA.Volumes, ","), NewValue: strings.Join(svcB.Volumes, ",")})
		}
		if !stringSliceEqual(svcA.Depends, svcB.Depends) {
			diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: prefix + ".depends",
				OldValue: strings.Join(svcA.Depends, ","), NewValue: strings.Join(svcB.Depends, ",")})
		}
		if !stringMapEqual(svcA.Env, svcB.Env) {
			diffs = append(diffs, DiffEntry{Kind: DiffChanged, Path: prefix + ".env",
				OldValue: formatMap(svcA.Env), NewValue: formatMap(svcB.Env)})
		}
	}

	return diffs
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func sortedKeys(m map[string]Service) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatStringSlice(ss []string) string {
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, ", ")
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func stringMapEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
