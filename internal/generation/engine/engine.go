package engine

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	xlanguage "golang.org/x/text/language"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

type GeneratorEngine interface {
	Generate(neir any) ([]Artifact, error)
	GenerateForLanguage(neir *model.NEIR, lang language.Language) ([]Artifact, error)
}

type Artifact struct {
	Path    string
	Content []byte
}

const goLicenseHeader = `// Copyright 2026 NAEOS Foundation
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

`

type DefaultEngine struct{}

func NewEngine() GeneratorEngine {
	return DefaultEngine{}
}

func (DefaultEngine) Generate(neir any) ([]Artifact, error) {
	if neir == nil {
		return nil, fmt.Errorf("neir is nil")
	}

	var projectName string
	var modules []any
	var services []any
	var archPattern string
	var deployStrategy string

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
	} else if neirStruct, ok := neir.(*model.NEIR); ok {
		if neirStruct.Project != nil {
			projectName = neirStruct.Project.Name
		}
		for _, m := range neirStruct.Modules {
			modules = append(modules, map[string]any{"name": m.Name, "path": m.Path})
		}
		for _, s := range neirStruct.Services {
			services = append(services, map[string]any{"name": s.Name, "port": s.Port, "kind": string(s.Kind)})
		}
		if neirStruct.Architecture != nil {
			archPattern = string(neirStruct.Architecture.Pattern)
		}
		if neirStruct.Deployment != nil {
			deployStrategy = string(neirStruct.Deployment.Strategy)
		}
	}

	var artifacts []Artifact

	if deployStrategy != "" {
		artifacts = append(artifacts, Artifact{
			Path:    "docker-compose.yml",
			Content: []byte("services:\n  app:\n    build: .\n    ports:\n      - '8080:8080'\n    deploy:\n      replicas: 2\n"),
		})
	}

	if archPattern != "" {
		artifacts = append(artifacts, Artifact{
			Path:    "docs/architecture.md",
			Content: []byte(fmt.Sprintf("# Architecture\n\nPattern: %s\n\n## Overview\n\nThis document describes the architectural decisions for %s.\n", archPattern, projectName)),
		})
	}

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
		pkg := strutil.Slugify(name)
		artifacts = append(artifacts, Artifact{
			Path:    fmt.Sprintf("%s/README.md", moduleDir),
			Content: []byte(fmt.Sprintf("# %s\n\nModule for %s project.\n", name, projectName)),
		})
		artifacts = append(artifacts, Artifact{
			Path:    fmt.Sprintf("%s/package.go", moduleDir),
			Content: []byte(fmt.Sprintf("package %s\n\n// %s module.\n", pkg, name)),
		})
		artifacts = append(artifacts, Artifact{
			Path:    fmt.Sprintf("%s/config.yaml", moduleDir),
			Content: []byte(fmt.Sprintf("name: %s\nmodule: %s\n", name, name)),
		})
	}

	for _, svc := range services {
		var name string
		var port int
		var kind string
		if serviceMap, ok := svc.(map[string]any); ok {
			name = fmt.Sprint(serviceMap["name"])
			if rawPort, ok := serviceMap["port"].(int); ok {
				port = rawPort
			}
			kind = fmt.Sprint(serviceMap["kind"])
		}
		if name == "" {
			continue
		}
		serviceDir := fmt.Sprintf("internal/%s", strutil.Slugify(name))
		artifacts = append(artifacts, Artifact{
			Path:    fmt.Sprintf("%s/config.yaml", serviceDir),
			Content: []byte(fmt.Sprintf("name: %s\nport: %d\nkind: %s\n", name, port, kind)),
		})
	}

	return artifacts, nil
}

func (DefaultEngine) GenerateForLanguage(neir *model.NEIR, lang language.Language) ([]Artifact, error) {
	if neir == nil {
		return nil, fmt.Errorf("neir is nil")
	}
	if !language.IsValid(lang) {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	projectName := ""
	if neir.Project != nil {
		projectName = neir.Project.Name
	}
	slug := strutil.Slugify(projectName)

	var artifacts []Artifact

	exts := language.Extensions(lang)
	ext := ".go"
	if len(exts) > 0 {
		ext = exts[0]
	}
	buildFile := language.BuildFile(lang)

	if buildFile != "" {
		artifacts = append(artifacts, Artifact{
			Path:    buildFile,
			Content: []byte(generateBuildFile(lang, slug)),
		})
	}

	artifacts = append(artifacts, Artifact{
		Path:    fmt.Sprintf("src/main%s", ext),
		Content: []byte(generateMainFile(lang, projectName)),
	})

	artifacts = append(artifacts, Artifact{
		Path:    "Dockerfile",
		Content: []byte(generateDockerfile(lang)),
	})

	for _, m := range neir.Modules {
		artifacts = append(artifacts, Artifact{
			Path:    fmt.Sprintf("src/%s/main%s", strutil.Slugify(m.Name), ext),
			Content: []byte(generateModuleFile(lang, m.Name, projectName)),
		})
	}

	return artifacts, nil
}

func generateBuildFile(lang language.Language, slug string) string {
	switch lang {
	case language.LanguageGo:
		return fmt.Sprintf("module github.com/example/%s\n\ngo 1.22\n", slug)
	case language.LanguageTypeScript:
		return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "main": "src/main.ts",
  "scripts": {
    "build": "tsc",
    "start": "node dist/main.js"
  }
}
`, slug)
	case language.LanguagePython:
		return fmt.Sprintf(`[project]
name = "%s"
version = "1.0.0"
requires-python = ">=3.10"
`, slug)
	case language.LanguageJava:
		return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project>
  <groupId>com.example</groupId>
  <artifactId>%s</artifactId>
  <version>1.0.0</version>
</project>
`, slug)
	case language.LanguageRust:
		return fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
edition = "2021"
`, slug)
	default:
		return ""
	}
}

func generateMainFile(lang language.Language, projectName string) string {
	switch lang {
	case language.LanguageGo:
		return goLicenseHeader + fmt.Sprintf("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello from %s\")\n}\n", projectName)
	case language.LanguageTypeScript:
		return fmt.Sprintf("console.log('hello from %s');\n", projectName)
	case language.LanguagePython:
		return fmt.Sprintf("print('hello from %s')\n", projectName)
	case language.LanguageJava:
		return fmt.Sprintf("public class App {\n    public static void main(String[] args) {\n        System.out.println(\"hello from %s\");\n    }\n}\n", projectName)
	case language.LanguageRust:
		return fmt.Sprintf("fn main() {\n    println!(\"hello from %s\");\n}\n", projectName)
	default:
		return ""
	}
}

func generateDockerfile(lang language.Language) string {
	switch lang {
	case language.LanguageGo:
		return "FROM golang:1.22-alpine\nWORKDIR /app\nCOPY . .\nRUN go build ./src/main.go\nCMD [\"./main\"]\n"
	case language.LanguageTypeScript:
		return "FROM node:20-alpine\nWORKDIR /app\nCOPY package*.json ./\nRUN npm install\nCOPY . .\nRUN npm run build\nCMD [\"node\", \"dist/main.js\"]\n"
	case language.LanguagePython:
		return "FROM python:3.12-slim\nWORKDIR /app\nCOPY . .\nCMD [\"python\", \"src/main.py\"]\n"
	case language.LanguageJava:
		return "FROM eclipse-temurin:21-jdk\nWORKDIR /app\nCOPY . .\nRUN javac src/main.java\nCMD [\"java\", \"src/App\"]\n"
	case language.LanguageRust:
		return "FROM rust:1.75-slim\nWORKDIR /app\nCOPY . .\nRUN cargo build --release\nCMD [\"./target/release/app\"]\n"
	default:
		return ""
	}
}

func generateModuleFile(lang language.Language, moduleName, _ string) string {
	pkg := strutil.Slugify(moduleName)
	switch lang {
	case language.LanguageGo:
		return goLicenseHeader + fmt.Sprintf("package %s\n\n// %s module.\n", pkg, moduleName)
	case language.LanguageTypeScript:
		return fmt.Sprintf("// %s module\nexport {};\n", moduleName)
	case language.LanguagePython:
		return fmt.Sprintf("# %s module\n", moduleName)
	case language.LanguageJava:
		title := cases.Title(xlanguage.English).String(moduleName)
		return fmt.Sprintf("public class %s {\n    // %s module\n}\n", title, moduleName)
	case language.LanguageRust:
		return fmt.Sprintf("// %s module\n", moduleName)
	default:
		return ""
	}
}
