// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package docgen

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type DocGenerator struct{}

func NewDocGenerator() *DocGenerator {
	return &DocGenerator{}
}

func (g *DocGenerator) GenerateFromSpec(doc *parser.SpecDocument) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# %s\n\n", doc.Project)
	sb.WriteString("Auto-generated documentation from specification.\n\n")

	if doc.Architecture != nil {
		sb.WriteString("## Architecture\n\n")
		fmt.Fprintf(&sb, "- Pattern: %s\n", doc.Architecture.Pattern)
		if len(doc.Architecture.Principles) > 0 {
			fmt.Fprintf(&sb, "- Principles: %s\n", strings.Join(doc.Architecture.Principles, ", "))
		}
		sb.WriteString("\n")
	}

	if len(doc.Modules) > 0 {
		sb.WriteString("## Modules\n\n")
		for _, m := range doc.Modules {
			fmt.Fprintf(&sb, "### %s\n\n", m.Name)
			if m.Description != "" {
				fmt.Fprintf(&sb, "%s\n\n", m.Description)
			}
			fmt.Fprintf(&sb, "- Path: `%s`\n", m.Path)
			if len(m.Dependencies) > 0 {
				fmt.Fprintf(&sb, "- Dependencies: %s\n", strings.Join(m.Dependencies, ", "))
			}
			sb.WriteString("\n")
		}
	}

	if len(doc.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range doc.Services {
			fmt.Fprintf(&sb, "### %s\n\n", s.Name)
			fmt.Fprintf(&sb, "- Kind: %s\n", s.Kind)
			if s.Port > 0 {
				fmt.Fprintf(&sb, "- Port: %d\n", s.Port)
			}
			if len(s.Endpoints) > 0 {
				sb.WriteString("\n**Endpoints:**\n\n")
				for _, ep := range s.Endpoints {
					fmt.Fprintf(&sb, "- `%s %s` → %s\n", ep.Method, ep.Path, ep.Action)
				}
			}
			sb.WriteString("\n")
		}
	}

	if doc.Deployment != nil {
		sb.WriteString("## Deployment\n\n")
		fmt.Fprintf(&sb, "- Strategy: %s\n", doc.Deployment.Strategy)
		if len(doc.Deployment.Environments) > 0 {
			fmt.Fprintf(&sb, "- Environments: %s\n", strings.Join(doc.Deployment.Environments, ", "))
		}
		sb.WriteString("\n")
	}

	if doc.Testing != nil {
		sb.WriteString("## Testing\n\n")
		fmt.Fprintf(&sb, "- Strategy: %s\n", doc.Testing.Strategy)
		if doc.Testing.Coverage != "" {
			fmt.Fprintf(&sb, "- Coverage target: %s%%\n", doc.Testing.Coverage)
		}
		sb.WriteString("\n")
	}

	if doc.Generation != nil && len(doc.Generation.Languages) > 0 {
		sb.WriteString("## Generation\n\n")
		fmt.Fprintf(&sb, "- Languages: %s\n", strings.Join(doc.Generation.Languages, ", "))
		if doc.Generation.OutputDir != "" {
			fmt.Fprintf(&sb, "- Output: %s\n", doc.Generation.OutputDir)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *DocGenerator) GenerateFromNEIR(neir *model.NEIR) string {
	var sb strings.Builder

	if neir.Project != nil {
		fmt.Fprintf(&sb, "# %s\n\n", neir.Project.Name)
		if neir.Project.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", neir.Project.Description)
		}
		if neir.Project.Version != "" {
			fmt.Fprintf(&sb, "Version: %s\n\n", neir.Project.Version)
		}
	}

	if len(neir.Modules) > 0 {
		sb.WriteString("## Modules\n\n")
		for _, m := range neir.Modules {
			fmt.Fprintf(&sb, "- **%s** (`%s`)", m.Name, m.Path)
			if m.Description != "" {
				fmt.Fprintf(&sb, " — %s", m.Description)
			}
			sb.WriteString("\n")
			if len(m.Dependencies) > 0 {
				fmt.Fprintf(&sb, "  Dependencies: %s\n", strings.Join(m.Dependencies, ", "))
			}
		}
		sb.WriteString("\n")
	}

	if len(neir.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range neir.Services {
			fmt.Fprintf(&sb, "- **%s** (kind=%s", s.Name, string(s.Kind))
			if s.Port > 0 {
				fmt.Fprintf(&sb, ", port=%d", s.Port)
			}
			sb.WriteString(")\n")
			for _, ep := range s.Endpoints {
				fmt.Fprintf(&sb, "  - `%s %s` → %s\n", ep.Method, ep.Path, ep.Action)
			}
		}
		sb.WriteString("\n")
	}

	if neir.Generation != nil {
		sb.WriteString("## Generation Config\n\n")
		var langs []string
		for _, l := range neir.Generation.Languages {
			langs = append(langs, string(l))
		}
		if len(langs) > 0 {
			fmt.Fprintf(&sb, "- Languages: %s\n", strings.Join(langs, ", "))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *DocGenerator) GenerateAPIDoc(doc *parser.SpecDocument) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# %s — API Reference\n\n", doc.Project)

	for _, svc := range doc.Services {
		if len(svc.Endpoints) > 0 {
			fmt.Fprintf(&sb, "## %s\n\n", svc.Name)
			for _, ep := range svc.Endpoints {
				fmt.Fprintf(&sb, "### `%s %s`\n\n", ep.Method, ep.Path)
				if ep.Action != "" {
					fmt.Fprintf(&sb, "**Action:** %s\n\n", ep.Action)
				}
			}
		}
	}

	return sb.String()
}

func (g *DocGenerator) GenerateModuleDocs(doc *parser.SpecDocument) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# %s — Module Documentation\n\n", doc.Project)

	for _, m := range doc.Modules {
		fmt.Fprintf(&sb, "## %s\n\n", m.Name)
		if m.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", m.Description)
		}
		fmt.Fprintf(&sb, "**Path:** `%s`\n\n", m.Path)
		if len(m.Dependencies) > 0 {
			sb.WriteString("**Dependencies:**\n\n")
			for _, dep := range m.Dependencies {
				fmt.Fprintf(&sb, "- %s\n", dep)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (g *DocGenerator) SupportedLanguages() []language.Language {
	return []language.Language{
		language.LanguageGo,
		language.LanguageTypeScript,
		language.LanguagePython,
		language.LanguageJava,
		language.LanguageRust,
	}
}
