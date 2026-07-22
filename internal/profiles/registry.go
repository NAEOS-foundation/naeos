package profiles

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

//go:embed profiles.json
var builtinProfilesJSON []byte

type Profile struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Industry     string            `json:"industry"`
	Version      string            `json:"version"`
	Modules      []ModuleTemplate  `json:"modules"`
	Services     []ServiceTemplate `json:"services"`
	Architecture ArchTemplate      `json:"architecture"`
	Security     SecurityTemplate  `json:"security"`
	Deployment   DeployTemplate    `json:"deployment"`
	Testing      TestTemplate      `json:"testing"`
	Tags         []string          `json:"tags"`
}

type ModuleTemplate struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type ServiceTemplate struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Port        int    `json:"port"`
	Description string `json:"description,omitempty"`
}

type ArchTemplate struct {
	Pattern    string   `json:"pattern"`
	Principles []string `json:"principles,omitempty"`
}

type SecurityTemplate struct {
	Authentication string   `json:"authentication,omitempty"`
	Authorization  string   `json:"authorization,omitempty"`
	Roles          []string `json:"roles,omitempty"`
	Encryption     bool     `json:"encryption,omitempty"`
}

type DeployTemplate struct {
	Strategy     string   `json:"strategy"`
	Environments []string `json:"environments,omitempty"`
}

type TestTemplate struct {
	Strategy   string   `json:"strategy"`
	Coverage   string   `json:"coverage,omitempty"`
	Frameworks []string `json:"frameworks,omitempty"`
}

type Registry struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
}

func NewRegistry() *Registry {
	r := &Registry{
		profiles: make(map[string]*Profile),
	}
	r.loadBuiltin()
	return r
}

func (r *Registry) loadBuiltin() {
	r.mu.Lock()
	defer r.mu.Unlock()
	var profiles []Profile
	if err := json.Unmarshal(builtinProfilesJSON, &profiles); err != nil {
		return
	}
	for i := range profiles {
		r.profiles[profiles[i].ID] = &profiles[i]
	}
}

func (r *Registry) Register(p *Profile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[p.ID] = p.Clone()
}

func (r *Registry) Get(id string) (*Profile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[id]
	return p, ok
}

func (r *Registry) List() []Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Profile, 0, len(r.profiles))
	for _, p := range r.profiles {
		result = append(result, *p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func (r *Registry) Search(query string) []Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	query = strings.ToLower(query)
	var result []Profile
	for _, p := range r.profiles {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) ||
			strings.Contains(strings.ToLower(p.Industry), query) ||
			strings.Contains(strings.ToLower(p.ID), query) {
			result = append(result, *p)
		}
	}
	return result
}

func (r *Registry) ByIndustry(industry string) []Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Profile
	for _, p := range r.profiles {
		if strings.EqualFold(p.Industry, industry) {
			result = append(result, *p)
		}
	}
	return result
}

func (p *Profile) Clone() *Profile {
	out := *p
	out.Modules = make([]ModuleTemplate, len(p.Modules))
	copy(out.Modules, p.Modules)
	out.Services = make([]ServiceTemplate, len(p.Services))
	copy(out.Services, p.Services)
	out.Tags = make([]string, len(p.Tags))
	copy(out.Tags, p.Tags)
	return &out
}

func (r *Registry) ToSpecYAML(p *Profile) string {
	var sb strings.Builder

	slug := strings.ToLower(strings.ReplaceAll(p.Name, " ", "-"))
	fmt.Fprintf(&sb, "project: %s\n", slug)
	fmt.Fprintf(&sb, "description: %s\n\n", p.Description)

	sb.WriteString("modules:\n")
	for _, m := range p.Modules {
		fmt.Fprintf(&sb, "  - name: %s\n    path: %s\n    description: %s\n", m.Name, m.Path, m.Description)
		if len(m.Dependencies) > 0 {
			sb.WriteString("    dependencies:\n")
			for _, d := range m.Dependencies {
				fmt.Fprintf(&sb, "      - %s\n", d)
			}
		}
	}

	sb.WriteString("\nservices:\n")
	for _, s := range p.Services {
		fmt.Fprintf(&sb, "  - name: %s\n    kind: %s\n    port: %d\n", s.Name, s.Kind, s.Port)
		if s.Description != "" {
			fmt.Fprintf(&sb, "    description: %s\n", s.Description)
		}
	}

	fmt.Fprintf(&sb, "\narchitecture:\n  pattern: %s\n", p.Architecture.Pattern)
	if len(p.Architecture.Principles) > 0 {
		sb.WriteString("  principles:\n")
		for _, pr := range p.Architecture.Principles {
			fmt.Fprintf(&sb, "    - %s\n", pr)
		}
	}

	if p.Security.Authentication != "" {
		fmt.Fprintf(&sb, "\nsecurity:\n  authentication:\n    method: %s\n", p.Security.Authentication)
		if p.Security.Authorization != "" {
			fmt.Fprintf(&sb, "  authorization:\n    model: %s\n", p.Security.Authorization)
		}
		if len(p.Security.Roles) > 0 {
			sb.WriteString("    roles:\n")
			for _, role := range p.Security.Roles {
				fmt.Fprintf(&sb, "      - %s\n", role)
			}
		}
	}

	fmt.Fprintf(&sb, "\ndeployment:\n  strategy: %s\n", p.Deployment.Strategy)
	if len(p.Deployment.Environments) > 0 {
		sb.WriteString("  environments:\n")
		for _, env := range p.Deployment.Environments {
			fmt.Fprintf(&sb, "    - %s\n", env)
		}
	}

	fmt.Fprintf(&sb, "\ntesting:\n  strategy: %s\n", p.Testing.Strategy)
	if p.Testing.Coverage != "" {
		fmt.Fprintf(&sb, "  coverage: %s\n", p.Testing.Coverage)
	}
	if len(p.Testing.Frameworks) > 0 {
		sb.WriteString("  frameworks:\n")
		for _, f := range p.Testing.Frameworks {
			fmt.Fprintf(&sb, "    - %s\n", f)
		}
	}

	return sb.String()
}
