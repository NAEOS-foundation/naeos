package contextbundle

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type Bundle struct {
	Project    string            `json:"project"`
	Summary    string            `json:"summary"`
	Modules    []ModuleContext   `json:"modules"`
	Services   []ServiceContext  `json:"services"`
	Languages  []string          `json:"languages"`
	Targets    []string          `json:"targets"`
	NEIR       string            `json:"neir,omitempty"`
	Raw        string            `json:"raw,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type ModuleContext struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type ServiceContext struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Port     int    `json:"port,omitempty"`
	Endpoints []EndpointContext `json:"endpoints,omitempty"`
}

type EndpointContext struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Action string `json:"action,omitempty"`
}

type Generator struct {
	compiler *compiler.Compiler
}

func NewGenerator(c *compiler.Compiler) *Generator {
	return &Generator{compiler: c}
}

func (g *Generator) GenerateFromNEIR(neir *model.NEIR) *Bundle {
	bundle := &Bundle{
		Metadata: make(map[string]string),
	}

	if neir.Project != nil {
		bundle.Project = neir.Project.Name
	}

	for _, mod := range neir.Modules {
		mc := ModuleContext{
			Name:        mod.Name,
			Path:        mod.Path,
			Description: mod.Description,
		}
		mc.Dependencies = append(mc.Dependencies, mod.Dependencies...)
		bundle.Modules = append(bundle.Modules, mc)
	}

	for _, svc := range neir.Services {
		sc := ServiceContext{
			Name: svc.Name,
			Kind: string(svc.Kind),
			Port: svc.Port,
		}
		for _, ep := range svc.Endpoints {
			sc.Endpoints = append(sc.Endpoints, EndpointContext{
				Method: ep.Method,
				Path:   ep.Path,
				Action: ep.Action,
			})
		}
		bundle.Services = append(bundle.Services, sc)
	}

	if neir.Generation != nil {
		for _, l := range neir.Generation.Languages {
			bundle.Languages = append(bundle.Languages, string(l))
		}
	}

	bundle.Summary = g.buildSummary(bundle)
	bundle.Metadata["generated_by"] = "naeos-context-bundle"
	bundle.Metadata["module_count"] = fmt.Sprintf("%d", len(bundle.Modules))
	bundle.Metadata["service_count"] = fmt.Sprintf("%d", len(bundle.Services))

	return bundle
}

func (g *Generator) GenerateFromSpec(doc *parser.SpecDocument) *Bundle {
	bundle := &Bundle{
		Metadata: make(map[string]string),
	}

	if doc.Project != "" {
		bundle.Project = doc.Project
	}
	if doc.Raw != "" {
		bundle.Raw = doc.Raw
	}

	for _, mod := range doc.Modules {
		bundle.Modules = append(bundle.Modules, ModuleContext{
			Name:         mod.Name,
			Path:         mod.Path,
			Description:  mod.Description,
			Dependencies: mod.Dependencies,
		})
	}

	for _, svc := range doc.Services {
		sc := ServiceContext{
			Name: svc.Name,
			Kind: svc.Kind,
			Port: svc.Port,
		}
		for _, ep := range svc.Endpoints {
			sc.Endpoints = append(sc.Endpoints, EndpointContext{
				Method: ep.Method,
				Path:   ep.Path,
				Action: ep.Action,
			})
		}
		bundle.Services = append(bundle.Services, sc)
	}

	if doc.Generation != nil {
		bundle.Languages = doc.Generation.Languages
	}

	bundle.Summary = g.buildSummary(bundle)
	bundle.Metadata["generated_by"] = "naeos-context-bundle"
	bundle.Metadata["module_count"] = fmt.Sprintf("%d", len(bundle.Modules))
	bundle.Metadata["service_count"] = fmt.Sprintf("%d", len(bundle.Services))

	return bundle
}

func (g *Generator) buildSummary(b *Bundle) string {
	var parts []string

	if b.Project != "" {
		parts = append(parts, fmt.Sprintf("Project: %s", b.Project))
	}

	if len(b.Modules) > 0 {
		names := make([]string, len(b.Modules))
		for i, m := range b.Modules {
			names[i] = m.Name
		}
		parts = append(parts, fmt.Sprintf("Modules: %s", strings.Join(names, ", ")))
	}

	if len(b.Services) > 0 {
		parts = append(parts, fmt.Sprintf("Services: %d", len(b.Services)))
	}

	if len(b.Languages) > 0 {
		parts = append(parts, fmt.Sprintf("Languages: %s", strings.Join(b.Languages, ", ")))
	}

	return strings.Join(parts, "; ")
}

func (b *Bundle) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — AI Context Bundle\n\n", b.Project))

	if b.Summary != "" {
		sb.WriteString(fmt.Sprintf("## Summary\n%s\n\n", b.Summary))
	}

	if len(b.Modules) > 0 {
		sb.WriteString("## Modules\n\n")
		for _, m := range b.Modules {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`)", m.Name, m.Path))
			if m.Description != "" {
				sb.WriteString(fmt.Sprintf(" — %s", m.Description))
			}
			sb.WriteString("\n")
			if len(m.Dependencies) > 0 {
				sb.WriteString(fmt.Sprintf("  Dependencies: %s\n", strings.Join(m.Dependencies, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	if len(b.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range b.Services {
			sb.WriteString(fmt.Sprintf("- **%s** (kind=%s", s.Name, s.Kind))
			if s.Port > 0 {
				sb.WriteString(fmt.Sprintf(", port=%d", s.Port))
			}
			sb.WriteString(")\n")
			for _, ep := range s.Endpoints {
				sb.WriteString(fmt.Sprintf("  - %s %s", ep.Method, ep.Path))
				if ep.Action != "" {
					sb.WriteString(fmt.Sprintf(" → %s", ep.Action))
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	if len(b.Targets) > 0 {
		sb.WriteString(fmt.Sprintf("## Targets\n%s\n\n", strings.Join(b.Targets, ", ")))
	}

	if b.NEIR != "" {
		sb.WriteString("## NEIR\n```json\n")
		sb.WriteString(b.NEIR)
		sb.WriteString("\n```\n\n")
	}

	return sb.String()
}

func (b *Bundle) ToPlainText() string {
	var sb strings.Builder

	if b.Project != "" {
		sb.WriteString(fmt.Sprintf("Project: %s\n", b.Project))
	}
	sb.WriteString(fmt.Sprintf("Modules: %d, Services: %d\n", len(b.Modules), len(b.Services)))

	if len(b.Languages) > 0 {
		sb.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(b.Languages, ", ")))
	}

	for _, m := range b.Modules {
		sb.WriteString(fmt.Sprintf("  Module: %s (%s)\n", m.Name, m.Path))
		if len(m.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("    deps: %s\n", strings.Join(m.Dependencies, ", ")))
		}
	}

	for _, s := range b.Services {
		sb.WriteString(fmt.Sprintf("  Service: %s kind=%s port=%d\n", s.Name, s.Kind, s.Port))
	}

	return sb.String()
}

func (b *Bundle) SupportedTargets() []string {
	targets := make([]string, 0, 4)
	targets = append(targets, "markdown", "plain")
	if b.NEIR != "" {
		targets = append(targets, "json")
	}
	sort.Strings(targets)
	return targets
}
