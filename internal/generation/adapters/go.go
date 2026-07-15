package adapters

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

type GoAdapter struct{}

func init() {
	Register(GoAdapter{})
}

func (GoAdapter) Language() language.Language {
	return language.LanguageGo
}

func cleanModulePath(path string) string {
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, ".\\")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "\\")
	if path == "" {
		return ""
	}
	path = strings.ToLower(strings.TrimSpace(path))
	path = strings.ReplaceAll(path, " ", "-")
	path = strings.ReplaceAll(path, "_", "-")
	return path
}

func pkgName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "/", "")
	return s
}

func (GoAdapter) GenerateProject(projectName string) []engine.Artifact {
	slug := strutil.Slugify(projectName)
	pkg := pkgName(projectName)

	return []engine.Artifact{
		{Path: "README.md", Content: []byte(fmt.Sprintf("# %s\n\nGenerated from NAEOS pipeline (Go).\n\n## Quick Start\n\n```bash\ngo run ./cmd/app\n```\n\n## Test\n\n```bash\ngo test ./...\n```\n", projectName))},
		{Path: "go.mod", Content: []byte(fmt.Sprintf("module github.com/example/%s\n\ngo 1.22\n\nrequire (\n\tgopkg.in/yaml.v3 v3.0.1\n)\n", slug))},
		{Path: ".gitignore", Content: []byte("# Binaries\n*.exe\n*.exe~\n*.dll\n*.so\n*.dylib\n\n# Test binary\n*.test\n\n# Output\n*.out\n\n# Dependency\nvendor/\n\n# IDE\n.idea/\n.vscode/\n*.swp\n*.swo\n\n# OS\n.DS_Store\nThumbs.db\n")},
		{Path: "cmd/app/main.go", Content: []byte(fmt.Sprintf("package main\n\nimport (\n\t\"fmt\"\n\t\"log\"\n\t\"net/http\"\n\n\t\"github.com/example/%s/internal/core\"\n\tcoreconfig \"github.com/example/%s/internal/core/config\"\n\tcorehttp \"github.com/example/%s/internal/core/http\"\n\tcoremiddleware \"github.com/example/%s/internal/core/middleware\"\n)\n\nfunc main() {\n\tcfg := coreconfig.Load(\"config.yaml\")\n\thandler := core.NewHandler(nil)\n\t_ = handler\n\tmux := http.NewServeMux()\n\tmux.HandleFunc(\"/\", func(w http.ResponseWriter, r *http.Request) {\n\t\t_, _ = fmt.Fprintf(w, \"hello from %s on port %%d\", cfg.Port)\n\t})\n\tmux.HandleFunc(\"/health\", func(w http.ResponseWriter, r *http.Request) {\n\t\t_, _ = fmt.Fprintln(w, \"ok\")\n\t})\n\tmux.HandleFunc(\"/api/v1\", func(w http.ResponseWriter, r *http.Request) {\n\t\t_, _ = fmt.Fprintln(w, \"api v1 ready\")\n\t})\n\tmux.HandleFunc(\"/api/v1/resources\", func(w http.ResponseWriter, r *http.Request) {\n\t\t_, _ = fmt.Fprintln(w, \"resources endpoint\")\n\t})\n\t_ = corehttp.Handler{}\n\twrapped := coremiddleware.LoggingMiddleware{}.Wrap(mux)\n\tlog.Printf(\"listening on :%%d\", cfg.Port)\n\tif err := http.ListenAndServe(fmt.Sprintf(\":%%d\", cfg.Port), wrapped); err != nil {\n\t\tlog.Fatal(err)\n\t}\n}\n", slug, slug, slug, slug, projectName))},
		{Path: fmt.Sprintf("%s/package.go", pkg), Content: []byte(fmt.Sprintf("package %s\n\n// %s module.\n", pkg, projectName))},
		{Path: "config.yaml", Content: []byte(fmt.Sprintf("name: %s\nport: 8080\nmode: development\n", slug))},
	}
}

func (GoAdapter) GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact {
	dir := cleanModulePath(modulePath)
	if dir == "" {
		dir = strutil.Slugify(moduleName)
	}
	pkg := pkgName(moduleName)

	return []engine.Artifact{
		{Path: fmt.Sprintf("%s/handler.go", dir), Content: []byte(fmt.Sprintf("package %s\n\ntype Handler struct {\n\tservice Service\n}\n\nfunc NewHandler(service Service) *Handler {\n\treturn &Handler{service: service}\n}\n", pkg))},
		{Path: fmt.Sprintf("%s/repository.go", dir), Content: []byte(fmt.Sprintf("package %s\n\ntype Repository interface {\n\tList() []string\n}\n", pkg))},
		{Path: fmt.Sprintf("%s/service.go", dir), Content: []byte(fmt.Sprintf("package %s\n\ntype Service interface {\n\tHandle() string\n}\n", pkg))},
		{Path: fmt.Sprintf("%s/domain/model.go", dir), Content: []byte("package domain\n\ntype Model struct {\n\tName string\n}\n")},
		{Path: fmt.Sprintf("%s/http/handler.go", dir), Content: []byte(fmt.Sprintf("package http\n\nimport (\n\t\"encoding/json\"\n\t\"net/http\"\n)\n\ntype Handler struct{}\n\nfunc (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode(map[string]string{\"module\": %q, \"status\": \"ok\"})\n}\n", moduleName))},
		{Path: fmt.Sprintf("%s/http/router.go", dir), Content: []byte(fmt.Sprintf("package http\n\nimport \"net/http\"\n\ntype Router struct {\n\thandler Handler\n}\n\nfunc NewRouter() *Router {\n\treturn &Router{handler: Handler{}}\n}\n\nfunc (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {\n\tr.handler.ServeHTTP(w, req)\n}\n"))},
		{Path: fmt.Sprintf("%s/middleware/logging.go", dir), Content: []byte("package middleware\n\nimport \"net/http\"\n\ntype LoggingMiddleware struct{}\n\nfunc (m LoggingMiddleware) Wrap(next http.Handler) http.Handler {\n\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n\t\tnext.ServeHTTP(w, r)\n\t})\n}\n")},
		{Path: fmt.Sprintf("%s/config/config.go", dir), Content: []byte(fmt.Sprintf("package config\n\ntype Config struct {\n\tName string `yaml:\"name\"`\n\tPort int    `yaml:\"port\"`\n\tMode string `yaml:\"mode\"`\n}\n"))},
		{Path: fmt.Sprintf("%s/config/load.go", dir), Content: []byte(fmt.Sprintf("package config\n\nimport (\n\t\"os\"\n\n\t\"gopkg.in/yaml.v3\"\n)\n\nfunc Load(path string) (*Config, error) {\n\tdata, err := os.ReadFile(path)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\tvar cfg Config\n\tif err := yaml.Unmarshal(data, &cfg); err != nil {\n\t\treturn nil, err\n\t}\n\treturn &cfg, nil\n}\n"))},
		{Path: fmt.Sprintf("%s/handler_test.go", dir), Content: []byte(fmt.Sprintf("package %s\n\nimport \"testing\"\n\nfunc TestHandler(t *testing.T) {\n\th := NewHandler(nil)\n\tif h == nil {\n\t\tt.Fatal(\"handler should not be nil\")\n\t}\n}\n", pkg))},
	}
}

func (GoAdapter) GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact {
	dir := fmt.Sprintf("internal/%s", strutil.Slugify(serviceName))
	pkg := pkgName(serviceName)

	var artifacts []engine.Artifact
	artifacts = append(artifacts, engine.Artifact{
		Path:    fmt.Sprintf("%s/config.yaml", dir),
		Content: []byte(fmt.Sprintf("name: %s\nport: %d\nkind: %s\n", serviceName, servicePort, serviceKind)),
	})

	if serviceKind == "http" || serviceKind == "" {
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("%s/server.go", dir),
			Content: []byte(fmt.Sprintf("package %s\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n)\n\nfunc Run(port int) error {\n\thandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n\t\tw.Header().Set(\"Content-Type\", \"application/json\")\n\t\tfmt.Fprintf(w, `{\"service\":\"%%s\",\"status\":\"ok\"}`, %q)\n\t})\n\taddr := fmt.Sprintf(\":%%d\", port)\n\tfmt.Printf(\"%%s listening on %%s\\n\", %q, addr)\n\treturn http.ListenAndServe(addr, handler)\n}\n", pkg, serviceName, serviceName)),
		})
		artifacts = append(artifacts, engine.Artifact{
			Path:    fmt.Sprintf("%s/server_test.go", dir),
			Content: []byte(fmt.Sprintf("package %s\n\nimport \"testing\"\n\nfunc TestRun(t *testing.T) {\n\tt.Log(\"test for %s server\")\n}\n", pkg, serviceName)),
		})
	}

	return artifacts
}

func (GoAdapter) GenerateDockerfile(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "Dockerfile",
		Content: []byte("FROM golang:1.22-alpine AS build\nWORKDIR /app\nCOPY . .\nRUN go build ./cmd/app\n\nFROM alpine:3.19\nCOPY --from=build /app/app /app/app\nCMD [\"/app/app\"]\n"),
	}}
}

func (GoAdapter) GenerateCI(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    ".github/workflows/ci.yml",
		Content: []byte("name: ci\n\non: [push, pull_request]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-go@v5\n        with:\n          go-version: '1.22'\n      - run: go test ./...\n"),
	}}
}

func (GoAdapter) GenerateDockerCompose(projectName string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docker-compose.yml",
		Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '8080:8080'\n"),
	}}
}

func (GoAdapter) GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact {
	return []engine.Artifact{{
		Path:    "docs/architecture.md",
		Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\nProject: %s\n", pattern, projectName)),
	}}
}
