package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type TemplateManager struct {
	templatesDir string
	templates    map[string]*template.Template
}

func NewManager(templatesDir string) *TemplateManager {
	return &TemplateManager{
		templatesDir: templatesDir,
		templates:    make(map[string]*template.Template),
	}
}

type TemplateInfo struct {
	Name     string
	Path     string
	IsCustom bool
}

func (m *TemplateManager) List() ([]TemplateInfo, error) {
	var result []TemplateInfo

	builtins := []string{"readme", "dockerfile", "ci", "main", "handler", "service", "repository", "config", "test", "middleware"}
	for _, name := range builtins {
		result = append(result, TemplateInfo{
			Name:     name,
			Path:     fmt.Sprintf("builtin:%s", name),
			IsCustom: false,
		})
	}

	if m.templatesDir != "" {
		entries, err := os.ReadDir(m.templatesDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
					name := strings.TrimSuffix(entry.Name(), ".tmpl")
					result = append(result, TemplateInfo{
						Name:     name,
						Path:     filepath.Join(m.templatesDir, entry.Name()),
						IsCustom: true,
					})
				}
			}
		}
	}

	return result, nil
}

func (m *TemplateManager) Get(name string) (*template.Template, error) {
	if cached, ok := m.templates[name]; ok {
		return cached, nil
	}

	if m.templatesDir != "" {
		customPath := filepath.Join(m.templatesDir, name+".tmpl")
		if _, err := os.Stat(customPath); err == nil {
			tmpl, err := template.ParseFiles(customPath)
			if err != nil {
				return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse custom template %s", name)
			}
			m.templates[name] = tmpl
			return tmpl, nil
		}
	}

	tmplStr := m.getBuiltinTemplate(name)
	if tmplStr == "" {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("template %s not found", name))
	}

	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse builtin template %s", name)
	}
	m.templates[name] = tmpl
	return tmpl, nil
}

func (m *TemplateManager) Render(name string, data any) (string, error) {
	tmpl, err := m.Get(name)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", naeoserr.Wrapf(err, naeoserr.ErrInternal, "execute template %s", name)
	}
	return sb.String(), nil
}

func (m *TemplateManager) AddCustom(name, content string) error {
	if m.templatesDir == "" {
		return naeoserr.New(naeoserr.ErrConfig, "no templates directory configured")
	}
	if err := os.MkdirAll(m.templatesDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(m.templatesDir, name+".tmpl")
	return os.WriteFile(path, []byte(content), 0o600)
}

func (m *TemplateManager) RemoveCustom(name string) error {
	if m.templatesDir == "" {
		return naeoserr.New(naeoserr.ErrConfig, "no templates directory configured")
	}
	path := filepath.Join(m.templatesDir, name+".tmpl")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("template %s not found", name))
	}
	delete(m.templates, name)
	return os.Remove(path)
}

func (m *TemplateManager) getBuiltinTemplate(name string) string {
	builtins := map[string]string{
		"readme": `# {{.ProjectName}}

{{.Description}}

## Quick Start

1. Review the specification
2. Install dependencies
3. Run the application

## Project Structure

- cmd/app/main.go - application entrypoint
- Dockerfile - container build definition
- .github/workflows/ci.yml - CI workflow
`,
		"dockerfile": `FROM {{.BaseImage}}

WORKDIR /app
COPY . .
{{.BuildCommand}}
CMD [{{.RunCommand}}]
`,
		"ci": `name: ci

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '{{.GoVersion}}'
      - run: go test ./...
`,
		"main": `package main

import "fmt"

func main() {
	fmt.Println("hello from {{.ProjectName}}")
}
`,
		"handler": `package {{.PackageName}}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
`,
		"service": `package {{.PackageName}}

type Service interface {
	Handle() string
}
`,
		"repository": `package {{.PackageName}}

type Repository interface {
	List() []string
}
`,
		"config": `package {{.PackageName}}

type Config struct {
	Port int
}

func Load() Config {
	return Config{Port: 8080}
}
`,
		"test": `package {{.PackageName}}

import "testing"

func Test{{.TestName}}(t *testing.T) {
	t.Log("test for {{.TestName}}")
}
`,
		"middleware": `package {{.PackageName}}

import "net/http"

type LoggingMiddleware struct{}

func (m LoggingMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
`,
	}
	return builtins[name]
}
