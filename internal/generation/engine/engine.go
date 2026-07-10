package engine

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/neir/builder"
)

type GeneratorEngine interface {
	Generate(neir any) ([]Artifact, error)
}

type Artifact struct {
	Path    string
	Content []byte
}

type DefaultEngine struct{}

func NewEngine() GeneratorEngine {
	return DefaultEngine{}
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

func (DefaultEngine) Generate(neir any) ([]Artifact, error) {
	if neir == nil {
		return nil, fmt.Errorf("neir is nil")
	}

	projectName := fmt.Sprint(neir)
	var modules []any
	var services []any

	if neirMap, ok := neir.(map[string]any); ok {
		if project, ok := neirMap["project"]; ok {
			projectName = fmt.Sprint(project)
		}
		if rawModules, ok := neirMap["modules"].([]any); ok {
			modules = rawModules
		}
		if rawServices, ok := neirMap["services"].([]any); ok {
			services = rawServices
		}
	} else if neirStruct, ok := neir.(*builder.NEIR); ok {
		projectName = fmt.Sprint(neirStruct.Project)
		modules = neirStruct.Modules
		if metadata, ok := neirStruct.Metadata["source"].(map[string]any); ok {
			if rawServices, ok := metadata["services"].([]any); ok {
				services = rawServices
			}
		}
	}

	artifacts := []Artifact{{
		Path: "README.md",
		Content: []byte(fmt.Sprintf("# %s\n\nGenerated from NAEOS pipeline.\n\n## Overview\n\nThis project was scaffolded with NAEOS and includes a minimal Go entrypoint, container build support, and CI workflow defaults.\n\n## Project Structure\n\n- cmd/app/main.go - application entrypoint\n- Dockerfile - container build definition\n- .github/workflows/ci.yml - CI workflow\n- spec.yaml - source specification\n\n## Quick Start\n\n1. Review spec.yaml\n2. Run `go test ./...`\n3. Build the app with `go build ./cmd/app`\n4. Run the binary with `./app`\n\n## Deployment\n\nThe generated Dockerfile and CI workflow provide a starting point for shipping the service in a containerized environment.\n", projectName)),
	}, {
		Path: "Dockerfile",
		Content: []byte("FROM golang:1.22-alpine\nWORKDIR /app\nCOPY . .\nRUN go build ./cmd/app\nCMD [\"/app/app\"]\n"),
	}, {
		Path: ".github/workflows/ci.yml",
		Content: []byte("name: ci\n\non: [push, pull_request]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-go@v5\n        with:\n          go-version: '1.22'\n      - run: go test ./...\n"),
	}, {
		Path: "go.mod",
		Content: []byte(fmt.Sprintf("module github.com/example/%s\n\ngo 1.22\n", projectName)),
	}, {
		Path: "cmd/app/main.go",
		Content: []byte(fmt.Sprintf("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello from %s\")\n}\n", projectName)),
	}}

	for _, module := range modules {
		var name string
		var path string

		if moduleMap, ok := module.(map[string]any); ok {
			name = fmt.Sprint(moduleMap["name"])
			path = fmt.Sprint(moduleMap["path"])
		} else if moduleStruct, ok := module.(map[string]string); ok {
			name = moduleStruct["name"]
			path = moduleStruct["path"]
		}

		if name == "" {
			continue
		}
		if path == "" {
			path = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		}
		moduleDir := strings.TrimPrefix(path, "./")
		if moduleDir == "" {
			moduleDir = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		}
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/README.md", moduleDir),
			Content: []byte(fmt.Sprintf("# %s\n", name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/package.go", moduleDir),
			Content: []byte(fmt.Sprintf("package %s\n\n// %s module placeholder.\n", slugify(name), name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/config.yaml", moduleDir),
			Content: []byte(fmt.Sprintf("name: %s\nmodule: %s\n", name, name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/handler.go", moduleDir),
			Content: []byte(fmt.Sprintf("package %s\n\n// Handler is a small starter implementation for the %s module.\ntype Handler struct {\n\tservice Service\n}\n\nfunc NewHandler(service Service) *Handler {\n\treturn &Handler{service: service}\n}\n", slugify(name), name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/repository.go", moduleDir),
			Content: []byte(fmt.Sprintf("package %s\n\n// Repository interface describes the persistence boundary for the %s module.\ntype Repository interface {\n\tList() []string\n}\n", slugify(name), name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/service.go", moduleDir),
			Content: []byte(fmt.Sprintf("package %s\n\n// Service interface describes the application behavior for the %s module.\ntype Service interface {\n\tHandle() string\n}\n", slugify(name), name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/domain/model.go", moduleDir),
			Content: []byte(fmt.Sprintf("package domain\n\n// Model is a sample domain object for the %s module.\ntype Model struct {\n\tName string\n}\n", name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/http/handler.go", moduleDir),
			Content: []byte(fmt.Sprintf("package http\n\nimport \"fmt\"\n\n// Handler is a starter HTTP handler for the %s module.\ntype Handler struct{}\n\nfunc (h Handler) ServeHTTP(w interface{}, r interface{}) {\n\tfmt.Println(\"handler for %s\")\n}\n", name, name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/http/router.go", moduleDir),
			Content: []byte(fmt.Sprintf("package http\n\n// Router is a starter router stub for the %s module.\ntype Router struct{}\n", name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/middleware/logging.go", moduleDir),
			Content: []byte(fmt.Sprintf("package middleware\n\n// LoggingMiddleware is a starter middleware stub for the %s module.\ntype LoggingMiddleware struct{}\n", name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/config/config.go", moduleDir),
			Content: []byte(fmt.Sprintf("package config\n\n// Config is a starter configuration container for the %s module.\ntype Config struct {\n\tPort int\n}\n", name)),
		})
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/config/load.go", moduleDir),
			Content: []byte(fmt.Sprintf("package config\n\n// Load returns a starter configuration for the %s module.\nfunc Load() Config {\n\treturn Config{Port: 8080}\n}\n", name)),
		})
	}

	for _, service := range services {
		var name string
		var port int
		if serviceMap, ok := service.(map[string]any); ok {
			name = fmt.Sprint(serviceMap["name"])
			if rawPort, ok := serviceMap["port"].(int); ok {
				port = rawPort
			}
		}
		if name == "" {
			continue
		}
		serviceDir := fmt.Sprintf("internal/%s", slugify(name))
		artifacts = append(artifacts, Artifact{
			Path: fmt.Sprintf("%s/config.yaml", serviceDir),
			Content: []byte(fmt.Sprintf("name: %s\nport: %d\n", name, port)),
		})
	}

	return artifacts, nil
}
