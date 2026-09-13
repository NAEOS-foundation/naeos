// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package contextbundle

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type DependencyEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type SecurityContext struct {
	AuthMethod   string   `json:"auth_method,omitempty"`
	AuthProvider string   `json:"auth_provider,omitempty"`
	AuthModel    string   `json:"auth_model,omitempty"`
	Roles        []string `json:"roles,omitempty"`
}

type CloudResource struct {
	Provider string `json:"provider"`
	Type     string `json:"type"`
	Name     string `json:"name"`
}

type Bundle struct {
	Project         string            `json:"project"`
	Summary         string            `json:"summary"`
	Modules         []ModuleContext   `json:"modules"`
	Services        []ServiceContext  `json:"services"`
	Languages       []string          `json:"languages"`
	Targets         []string          `json:"targets"`
	NEIR            string            `json:"neir,omitempty"`
	Raw             string            `json:"raw,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	DependencyGraph []DependencyEdge  `json:"dependency_graph,omitempty"`
	Security        *SecurityContext  `json:"security,omitempty"`
	Cloud           []CloudResource   `json:"cloud,omitempty"`
}

type ModuleContext struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type ServiceContext struct {
	Name      string            `json:"name"`
	Kind      string            `json:"kind"`
	Port      int               `json:"port,omitempty"`
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

	for _, mod := range doc.Modules {
		for _, dep := range mod.Dependencies {
			bundle.DependencyGraph = append(bundle.DependencyGraph, DependencyEdge{
				From: mod.Name,
				To:   dep,
				Kind: "module",
			})
		}
	}

	bundle.Security = extractSecurityFromDoc(doc)
	bundle.Cloud = extractCloudFromDoc(doc)

	bundle.Summary = g.buildSummary(bundle)
	bundle.Metadata["generated_by"] = "naeos-context-bundle"
	bundle.Metadata["module_count"] = fmt.Sprintf("%d", len(bundle.Modules))
	bundle.Metadata["service_count"] = fmt.Sprintf("%d", len(bundle.Services))

	return bundle
}

func extractSecurityFromDoc(doc *parser.SpecDocument) *SecurityContext {
	if doc.Data == nil {
		return nil
	}
	m, ok := doc.Data.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := m["security"]
	if !ok {
		return nil
	}
	sm, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	sc := &SecurityContext{}
	if v, ok := sm["auth_method"].(string); ok {
		sc.AuthMethod = v
	}
	if v, ok := sm["auth_provider"].(string); ok {
		sc.AuthProvider = v
	}
	if v, ok := sm["auth_model"].(string); ok {
		sc.AuthModel = v
	}
	if roles, ok := sm["roles"].([]any); ok {
		for _, r := range roles {
			if s, ok := r.(string); ok {
				sc.Roles = append(sc.Roles, s)
			}
		}
	}
	return sc
}

func extractCloudFromDoc(doc *parser.SpecDocument) []CloudResource {
	if doc.Data == nil {
		return nil
	}
	m, ok := doc.Data.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := m["cloud"]
	if !ok {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	var resources []CloudResource
	for _, item := range list {
		rm, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cr := CloudResource{}
		if v, ok := rm["provider"].(string); ok {
			cr.Provider = v
		}
		if v, ok := rm["type"].(string); ok {
			cr.Type = v
		}
		if v, ok := rm["name"].(string); ok {
			cr.Name = v
		}
		resources = append(resources, cr)
	}
	return resources
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

	fmt.Fprintf(&sb, "# %s — AI Context Bundle\n\n", b.Project)

	if b.Summary != "" {
		fmt.Fprintf(&sb, "## Summary\n%s\n\n", b.Summary)
	}

	if len(b.Modules) > 0 {
		sb.WriteString("## Modules\n\n")
		for _, m := range b.Modules {
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

	if len(b.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, s := range b.Services {
			fmt.Fprintf(&sb, "- **%s** (kind=%s", s.Name, s.Kind)
			if s.Port > 0 {
				fmt.Fprintf(&sb, ", port=%d", s.Port)
			}
			sb.WriteString(")\n")
			for _, ep := range s.Endpoints {
				fmt.Fprintf(&sb, "  - %s %s", ep.Method, ep.Path)
				if ep.Action != "" {
					fmt.Fprintf(&sb, " → %s", ep.Action)
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	if len(b.Targets) > 0 {
		fmt.Fprintf(&sb, "## Targets\n%s\n\n", strings.Join(b.Targets, ", "))
	}

	if b.NEIR != "" {
		sb.WriteString("## NEIR\n```json\n")
		sb.WriteString(b.NEIR)
		sb.WriteString("\n```\n\n")
	}

	if len(b.DependencyGraph) > 0 {
		sb.WriteString("## Dependency Graph\n\n")
		for _, e := range b.DependencyGraph {
			fmt.Fprintf(&sb, "- `%s` → `%s` (kind=%s)\n", e.From, e.To, e.Kind)
		}
		sb.WriteString("\n")
	}

	if b.Security != nil {
		sb.WriteString("## Security\n\n")
		if b.Security.AuthMethod != "" {
			fmt.Fprintf(&sb, "- Auth Method: %s\n", b.Security.AuthMethod)
		}
		if b.Security.AuthProvider != "" {
			fmt.Fprintf(&sb, "- Auth Provider: %s\n", b.Security.AuthProvider)
		}
		if b.Security.AuthModel != "" {
			fmt.Fprintf(&sb, "- Auth Model: %s\n", b.Security.AuthModel)
		}
		if len(b.Security.Roles) > 0 {
			fmt.Fprintf(&sb, "- Roles: %s\n", strings.Join(b.Security.Roles, ", "))
		}
		sb.WriteString("\n")
	}

	if len(b.Cloud) > 0 {
		sb.WriteString("## Cloud Resources\n\n")
		for _, cr := range b.Cloud {
			fmt.Fprintf(&sb, "- **%s** (provider=%s, type=%s)\n", cr.Name, cr.Provider, cr.Type)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (b *Bundle) ToPlainText() string {
	var sb strings.Builder

	if b.Project != "" {
		fmt.Fprintf(&sb, "Project: %s\n", b.Project)
	}
	fmt.Fprintf(&sb, "Modules: %d, Services: %d\n", len(b.Modules), len(b.Services))

	if len(b.Languages) > 0 {
		fmt.Fprintf(&sb, "Languages: %s\n", strings.Join(b.Languages, ", "))
	}

	for _, m := range b.Modules {
		fmt.Fprintf(&sb, "  Module: %s (%s)\n", m.Name, m.Path)
		if len(m.Dependencies) > 0 {
			fmt.Fprintf(&sb, "    deps: %s\n", strings.Join(m.Dependencies, ", "))
		}
	}

	for _, s := range b.Services {
		fmt.Fprintf(&sb, "  Service: %s kind=%s port=%d\n", s.Name, s.Kind, s.Port)
	}

	if len(b.DependencyGraph) > 0 {
		sb.WriteString("Dependency Graph:\n")
		for _, e := range b.DependencyGraph {
			fmt.Fprintf(&sb, "  %s -> %s (kind=%s)\n", e.From, e.To, e.Kind)
		}
	}

	if b.Security != nil {
		sb.WriteString("Security:\n")
		if b.Security.AuthMethod != "" {
			fmt.Fprintf(&sb, "  Auth Method: %s\n", b.Security.AuthMethod)
		}
		if b.Security.AuthProvider != "" {
			fmt.Fprintf(&sb, "  Auth Provider: %s\n", b.Security.AuthProvider)
		}
		if b.Security.AuthModel != "" {
			fmt.Fprintf(&sb, "  Auth Model: %s\n", b.Security.AuthModel)
		}
		if len(b.Security.Roles) > 0 {
			fmt.Fprintf(&sb, "  Roles: %s\n", strings.Join(b.Security.Roles, ", "))
		}
	}

	if len(b.Cloud) > 0 {
		sb.WriteString("Cloud Resources:\n")
		for _, cr := range b.Cloud {
			fmt.Fprintf(&sb, "  %s (provider=%s, type=%s)\n", cr.Name, cr.Provider, cr.Type)
		}
	}

	return sb.String()
}

func (b *Bundle) buildTargets(languages []string) []string {
	targets := []string{"markdown", "plain", "json"}
	seen := make(map[string]bool)
	for _, t := range targets {
		seen[t] = true
	}
	for _, lang := range languages {
		key := "lang-" + strings.ToLower(lang)
		if !seen[key] {
			targets = append(targets, key)
			seen[key] = true
		}
	}
	sort.Strings(targets)
	return targets
}

func (b *Bundle) buildSummary() string {
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

func (b *Bundle) SupportedTargets() []string {
	targets := make([]string, 0, 4)
	targets = append(targets, "markdown", "plain", "json")
	if b.NEIR != "" {
		targets = append(targets, "neir")
	}
	sort.Strings(targets)
	return targets
}

func (b *Bundle) EstimateTokens() int {
	content := b.ToPlainText()
	words := strings.Fields(content)
	return len(words) * 4 / 3
}

func (b *Bundle) ToJSON() string {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (b *Bundle) FilterByModule(names []string) *Bundle {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	filtered := &Bundle{
		Project:   b.Project,
		Summary:   b.Summary,
		Metadata:  b.Metadata,
		Security:  b.Security,
		Cloud:     b.Cloud,
		NEIR:      b.NEIR,
		Raw:       b.Raw,
		Languages: b.Languages,
		Targets:   b.Targets,
	}
	for _, m := range b.Modules {
		if nameSet[m.Name] {
			filtered.Modules = append(filtered.Modules, m)
		}
	}
	filtered.Services = append(filtered.Services, b.Services...)
	for _, e := range b.DependencyGraph {
		if nameSet[e.From] || nameSet[e.To] {
			filtered.DependencyGraph = append(filtered.DependencyGraph, e)
		}
	}
	return filtered
}

func (b *Bundle) FilterByService(kinds []string) *Bundle {
	kindSet := make(map[string]bool, len(kinds))
	for _, k := range kinds {
		kindSet[strings.ToLower(k)] = true
	}
	filtered := &Bundle{
		Project:         b.Project,
		Summary:         b.Summary,
		Metadata:        b.Metadata,
		Security:        b.Security,
		Cloud:           b.Cloud,
		NEIR:            b.NEIR,
		Raw:             b.Raw,
		Languages:       b.Languages,
		Targets:         b.Targets,
		Modules:         b.Modules,
		DependencyGraph: b.DependencyGraph,
	}
	for _, s := range b.Services {
		if kindSet[strings.ToLower(s.Kind)] {
			filtered.Services = append(filtered.Services, s)
		}
	}
	return filtered
}

func (b *Bundle) Merge(other *Bundle) *Bundle {
	merged := &Bundle{
		Project:  b.Project,
		Summary:  b.Summary,
		Metadata: make(map[string]string),
		Security: b.Security,
		Cloud:    append([]CloudResource{}, b.Cloud...),
		NEIR:     b.NEIR,
		Raw:      b.Raw,
	}
	for k, v := range b.Metadata {
		merged.Metadata[k] = v
	}
	if other != nil {
		if other.Project != "" {
			merged.Project = other.Project
		}
		merged.Modules = append(merged.Modules, b.Modules...)
		merged.Modules = append(merged.Modules, other.Modules...)
		merged.Services = append(merged.Services, b.Services...)
		merged.Services = append(merged.Services, other.Services...)
		merged.Languages = append(merged.Languages, b.Languages...)
		merged.Languages = append(merged.Languages, other.Languages...)
		merged.DependencyGraph = append(merged.DependencyGraph, b.DependencyGraph...)
		merged.DependencyGraph = append(merged.DependencyGraph, other.DependencyGraph...)
		merged.Cloud = append(merged.Cloud, other.Cloud...)
		for k, v := range other.Metadata {
			merged.Metadata[k] = v
		}
		if other.Security != nil {
			merged.Security = other.Security
		}
	} else {
		merged.Modules = append([]ModuleContext{}, b.Modules...)
		merged.Services = append([]ServiceContext{}, b.Services...)
		merged.Languages = append([]string{}, b.Languages...)
		merged.DependencyGraph = append([]DependencyEdge{}, b.DependencyGraph...)
	}
	seenLang := make(map[string]bool)
	var uniqueLangs []string
	for _, l := range merged.Languages {
		if !seenLang[l] {
			seenLang[l] = true
			uniqueLangs = append(uniqueLangs, l)
		}
	}
	merged.Languages = uniqueLangs
	merged.Targets = b.buildTargets(merged.Languages)
	merged.Summary = merged.buildSummary()
	merged.Metadata["token_estimate"] = fmt.Sprintf("%d", merged.EstimateTokens())
	return merged
}
