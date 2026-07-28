package promptlib

import (
	"bytes"
	"fmt"
	"text/template"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"

	"gopkg.in/yaml.v3"
)

// LLMPrompt represents a prompt template for LLM interactions.
type LLMPrompt struct {
	Name        string       `yaml:"name"`
	Kind        string       `yaml:"kind"`
	Version     string       `yaml:"version,omitempty"`
	Description string       `yaml:"description,omitempty"`
	Provider    string       `yaml:"provider,omitempty"`
	System      string       `yaml:"system,omitempty"`
	User        string       `yaml:"user"`
	Variables   []Variable   `yaml:"variables,omitempty"`
	Constraints *Constraints `yaml:"constraints,omitempty"`
}

// CompilerTemplate represents a prompt template for AI instruction compilation.
type CompilerTemplate struct {
	Name      string     `yaml:"name"`
	Kind      string     `yaml:"kind"`
	Version   string     `yaml:"version,omitempty"`
	Target    string     `yaml:"target"`
	Files     []FileSpec `yaml:"files"`
	Variables []Variable `yaml:"variables,omitempty"`
}

// Variable defines a template variable with metadata.
type Variable struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Constraints holds LLM generation constraints.
type Constraints struct {
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty"`
}

// FileSpec defines a single output file in a compiler template.
type FileSpec struct {
	Path     string `yaml:"path"`
	Kind     string `yaml:"kind"`
	Template string `yaml:"template"`
}

// RenderedLLM holds the rendered output of an LLM prompt.
type RenderedLLM struct {
	System      string
	User        string
	MaxTokens   int
	Temperature float64
}

// RenderedFile holds a single rendered output file from a compiler template.
type RenderedFile struct {
	Path    string
	Content string
	Kind    string
}

// ParseLLMPrompt parses YAML data into an LLMPrompt.
func ParseLLMPrompt(data []byte) (*LLMPrompt, error) {
	var p LLMPrompt
	if err := parseYAML(data, &p); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse LLM prompt")
	}
	if p.Name == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "LLM prompt name is required")
	}
	if p.User == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "LLM prompt user template is required")
	}
	if p.Constraints == nil {
		p.Constraints = &Constraints{MaxTokens: 1024, Temperature: 0.3}
	}
	if p.Constraints.MaxTokens == 0 {
		p.Constraints.MaxTokens = 1024
	}
	if p.Constraints.Temperature == 0 {
		p.Constraints.Temperature = 0.3
	}
	return &p, nil
}

// ParseCompilerTemplate parses YAML data into a CompilerTemplate.
func ParseCompilerTemplate(data []byte) (*CompilerTemplate, error) {
	var t CompilerTemplate
	if err := parseYAML(data, &t); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse compiler template")
	}
	if t.Name == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "compiler template name is required")
	}
	if t.Target == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "compiler template target is required")
	}
	if len(t.Files) == 0 {
		return nil, naeoserr.New(naeoserr.ErrValidation, "compiler template must have at least one file")
	}
	for i, f := range t.Files {
		if f.Path == "" {
			return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("file[%d] path is required", i))
		}
		if f.Template == "" {
			return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("file[%d] template is required", i))
		}
	}
	return &t, nil
}

// RenderLLM renders an LLM prompt with the given data variables.
func RenderLLM(p *LLMPrompt, data map[string]any) (*RenderedLLM, error) {
	if p == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "nil LLM prompt")
	}

	user, err := renderTemplate("user", p.User, data)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "render user template")
	}

	var system string
	if p.System != "" {
		system, err = renderTemplate("system", p.System, data)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "render system template")
		}
	}

	return &RenderedLLM{
		System:      system,
		User:        user,
		MaxTokens:   p.Constraints.MaxTokens,
		Temperature: p.Constraints.Temperature,
	}, nil
}

// RenderCompiler renders a compiler template with NEIR data.
func RenderCompiler(t *CompilerTemplate, neir *model.NEIR) ([]RenderedFile, error) {
	if t == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "nil compiler template")
	}
	if neir == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "nil NEIR")
	}

	ctx := buildNEIRContext(neir)
	var files []RenderedFile

	for _, fs := range t.Files {
		content, err := renderTemplate(fs.Path, fs.Template, ctx)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "render file %s", fs.Path)
		}
		files = append(files, RenderedFile{
			Path:    fs.Path,
			Content: content,
			Kind:    fs.Kind,
		})
	}

	return files, nil
}

func renderTemplate(name, tmplStr string, data any) (string, error) {
	tmpl, err := template.New(name).Funcs(FuncMap).Parse(tmplStr)
	if err != nil {
		return "", naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse template %s", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", naeoserr.Wrapf(err, naeoserr.ErrInternal, "execute template %s", name)
	}

	return buf.String(), nil
}

func parseYAML(data []byte, v any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(v)
}

// NEIRContext provides a view model for rendering NEIR data in templates.
type NEIRContext struct {
	Project        projectView
	Architecture   architectureView
	Modules        []moduleView
	Components     []componentView
	Services       []serviceView
	APIs           []apiView
	Security       *securityView
	Deployment     *deploymentView
	Testing        *testingView
	Storage        []storageView
	Infrastructure *infrastructureView
	AI             *aiView
	Documentation  *documentationView
}

type projectView struct {
	Name        string
	Description string
	Version     string
}

type architectureView struct {
	Pattern    string
	Principles []string
}

type moduleView struct {
	Name         string
	Path         string
	Description  string
	Dependencies []string
}

type componentView struct {
	Name   string
	Kind   string
	Module string
}

type serviceView struct {
	Name      string
	Kind      string
	Port      int
	Endpoints []endpointView
}

type endpointView struct {
	Method  string
	Path    string
	Action  string
	Summary string
}

type apiView struct {
	Name      string
	Version   string
	Protocol  string
	Endpoints []endpointView
}

type securityView struct {
	Authentication *authView
	Authorization  *authzView
	Encryption     *encryptionView
}

type authView struct {
	Method   string
	Provider string
}

type authzView struct {
	Model string
	Roles []string
}

type encryptionView struct {
	InTransit bool
	AtRest    bool
}

type deploymentView struct {
	Strategy     string
	Environments []envView
}

type envView struct {
	Name string
	Kind string
}

type testingView struct {
	Strategy   string
	Frameworks []string
}

type storageView struct {
	Name        string
	Type        string
	Provider    string
	Collections []collectionView
}

type collectionView struct {
	Name string
}

type infrastructureView struct {
	Provider  string
	Region    string
	Resources []resourceView
}

type resourceView struct {
	Name string
	Kind string
}

type aiView struct {
	Models []aiModelView
}

type aiModelView struct {
	Name    string
	Kind    string
	Version string
}

type documentationView struct {
	ADRs []docEntryView
	RFCs []docEntryView
}

type docEntryView struct {
	Title string
}

func buildNEIRContext(neir *model.NEIR) *NEIRContext {
	ctx := &NEIRContext{}

	if neir.Project != nil {
		ctx.Project = projectView{
			Name:        neir.Project.Name,
			Description: neir.Project.Description,
			Version:     neir.Project.Version,
		}
	}

	if neir.Architecture != nil {
		ctx.Architecture = architectureView{
			Pattern:    string(neir.Architecture.Pattern),
			Principles: neir.Architecture.Principles,
		}
	}

	for _, m := range neir.Modules {
		ctx.Modules = append(ctx.Modules, moduleView{
			Name:         m.Name,
			Path:         m.Path,
			Description:  m.Description,
			Dependencies: m.Dependencies,
		})
	}

	for _, c := range neir.Components {
		ctx.Components = append(ctx.Components, componentView{
			Name:   c.Name,
			Kind:   string(c.Kind),
			Module: c.Module,
		})
	}

	for _, s := range neir.Services {
		sv := serviceView{
			Name: s.Name,
			Kind: string(s.Kind),
			Port: s.Port,
		}
		for _, ep := range s.Endpoints {
			sv.Endpoints = append(sv.Endpoints, endpointView{
				Method: ep.Method,
				Path:   ep.Path,
				Action: ep.Action,
			})
		}
		ctx.Services = append(ctx.Services, sv)
	}

	for _, a := range neir.APIs {
		av := apiView{
			Name:     a.Name,
			Version:  a.Version,
			Protocol: string(a.Protocol),
		}
		for _, ep := range a.Endpoints {
			av.Endpoints = append(av.Endpoints, endpointView{
				Method:  ep.Method,
				Path:    ep.Path,
				Summary: ep.Summary,
			})
		}
		ctx.APIs = append(ctx.APIs, av)
	}

	if neir.Security != nil {
		sv := &securityView{}
		if neir.Security.Authentication != nil {
			sv.Authentication = &authView{
				Method:   neir.Security.Authentication.Method,
				Provider: neir.Security.Authentication.Provider,
			}
		}
		if neir.Security.Authorization != nil {
			sv.Authorization = &authzView{
				Model: neir.Security.Authorization.Model,
				Roles: neir.Security.Authorization.Roles,
			}
		}
		if neir.Security.Encryption != nil {
			sv.Encryption = &encryptionView{
				InTransit: neir.Security.Encryption.InTransit,
				AtRest:    neir.Security.Encryption.AtRest,
			}
		}
		ctx.Security = sv
	}

	if neir.Deployment != nil {
		dv := &deploymentView{Strategy: string(neir.Deployment.Strategy)}
		for _, env := range neir.Deployment.Environments {
			dv.Environments = append(dv.Environments, envView{
				Name: env.Name,
				Kind: env.Kind,
			})
		}
		ctx.Deployment = dv
	}

	if neir.Testing != nil {
		ctx.Testing = &testingView{
			Strategy:   string(neir.Testing.Strategy),
			Frameworks: neir.Testing.Frameworks,
		}
	}

	for _, st := range neir.Storage {
		sv := storageView{
			Name:     st.Name,
			Type:     string(st.Type),
			Provider: st.Provider,
		}
		for _, col := range st.Collections {
			sv.Collections = append(sv.Collections, collectionView{Name: col.Name})
		}
		ctx.Storage = append(ctx.Storage, sv)
	}

	if neir.Infrastructure != nil {
		iv := &infrastructureView{
			Provider: string(neir.Infrastructure.Provider),
			Region:   neir.Infrastructure.Region,
		}
		for _, r := range neir.Infrastructure.Resources {
			iv.Resources = append(iv.Resources, resourceView{
				Name: r.Name,
				Kind: r.Kind,
			})
		}
		ctx.Infrastructure = iv
	}

	if neir.AI != nil {
		av := &aiView{}
		for _, m := range neir.AI.Models {
			av.Models = append(av.Models, aiModelView{
				Name:    m.Name,
				Kind:    m.Kind,
				Version: m.Version,
			})
		}
		ctx.AI = av
	}

	if neir.Documentation != nil {
		dv := &documentationView{}
		for _, adr := range neir.Documentation.ADRs {
			dv.ADRs = append(dv.ADRs, docEntryView{Title: adr.Title})
		}
		for _, rfc := range neir.Documentation.RFCs {
			dv.RFCs = append(dv.RFCs, docEntryView{Title: rfc.Title})
		}
		ctx.Documentation = dv
	}

	return ctx
}
