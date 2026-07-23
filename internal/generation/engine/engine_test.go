package engine

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/testutil"
)

func TestGeneratorCreatesArtifactsFromNEIR(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "acme-api"},
		Modules: []module.Module{{Name: "auth", Path: "./internal/auth"}},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(artifacts) < 1 {
		t.Fatalf("expected at least one artifact, got %d", len(artifacts))
	}
	foundModule := false
	for _, a := range artifacts {
		if a.Path == "internal/auth/README.md" {
			foundModule = true
		}
	}
	if !foundModule {
		t.Error("expected module README artifact")
	}
}

func TestGenerateForLanguageGo(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "acme-api"},
		Modules: []module.Module{{Name: "auth", Path: "./internal/auth"}},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageGo)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}

	if len(artifacts) < 3 {
		t.Fatalf("expected at least 3 artifacts (go.mod, main.go, Dockerfile), got %d", len(artifacts))
	}

	foundGoMod := false
	foundMain := false
	foundDockerfile := false
	for _, a := range artifacts {
		if a.Path == "go.mod" {
			foundGoMod = true
			if !testutil.ContainsBytes(a.Content, "module github.com/example/acme-api") {
				t.Errorf("go.mod should contain module path")
			}
		}
		if a.Path == "src/main.go" {
			foundMain = true
			if !testutil.ContainsBytes(a.Content, "hello from acme-api") {
				t.Errorf("main.go should contain project name")
			}
		}
		if a.Path == "Dockerfile" {
			foundDockerfile = true
			if !testutil.ContainsBytes(a.Content, "golang:1.22") {
				t.Errorf("Dockerfile should use Go image")
			}
		}
	}

	if !foundGoMod {
		t.Error("expected go.mod artifact")
	}
	if !foundMain {
		t.Error("expected src/main.go artifact")
	}
	if !foundDockerfile {
		t.Error("expected Dockerfile artifact")
	}
}

func TestGenerateForLanguageGoIncludesLicenseHeader(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "acme-api"},
		Modules: []module.Module{{Name: "auth", Path: "./internal/auth"}},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageGo)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}
	if len(artifacts) == 0 {
		t.Fatal("expected artifacts from Go generation")
	}

	for _, a := range artifacts {
		if strings.HasSuffix(a.Path, ".go") {
			content := string(a.Content)
			if !strings.Contains(content, "Copyright 2026 NAEOS Foundation") {
				t.Errorf("expected license header in %s", a.Path)
			}
			if !strings.Contains(content, "package ") {
				t.Errorf("expected package declaration in %s", a.Path)
			}
		}
	}
}

func TestGenerateForLanguageTypeScript(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "web-app"},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageTypeScript)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}

	foundPackageJson := false
	foundMain := false
	for _, a := range artifacts {
		if a.Path == "package.json" {
			foundPackageJson = true
			if !testutil.ContainsBytes(a.Content, "web-app") {
				t.Errorf("package.json should contain project name")
			}
		}
		if a.Path == "src/main.ts" {
			foundMain = true
			if !testutil.ContainsBytes(a.Content, "hello from web-app") {
				t.Errorf("main.ts should contain project name")
			}
		}
	}

	if !foundPackageJson {
		t.Error("expected package.json artifact")
	}
	if !foundMain {
		t.Error("expected src/main.ts artifact")
	}
}

func TestGenerateForLanguagePython(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "ml-service"},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguagePython)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}

	foundPyproject := false
	for _, a := range artifacts {
		if a.Path == "pyproject.toml" {
			foundPyproject = true
			if !testutil.ContainsBytes(a.Content, "ml-service") {
				t.Errorf("pyproject.toml should contain project name")
			}
		}
	}

	if !foundPyproject {
		t.Error("expected pyproject.toml artifact")
	}
}

func TestGenerateForLanguageWithModules(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "fullstack"},
		Modules: []module.Module{
			{Name: "auth", Path: "./internal/auth"},
			{Name: "api", Path: "./internal/api"},
		},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageGo)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}

	authFiles := 0
	apiFiles := 0
	for _, a := range artifacts {
		if testutil.Contains(a.Path, "auth") {
			authFiles++
		}
		if testutil.Contains(a.Path, "api") {
			apiFiles++
		}
	}

	if authFiles == 0 {
		t.Error("expected auth module files")
	}
	if apiFiles == 0 {
		t.Error("expected api module files")
	}
}

func TestGenerateForLanguageNilNEIR(t *testing.T) {
	engine := NewEngine()
	_, err := engine.GenerateForLanguage(nil, language.LanguageGo)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestGenerateForLanguageInvalidLanguage(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
	}

	engine := NewEngine()
	_, err := engine.GenerateForLanguage(neir, "invalid-lang")
	if err == nil {
		t.Fatal("expected error for invalid language")
	}
}

func TestGenerateNilNEIR(t *testing.T) {
	engine := NewEngine()
	_, err := engine.Generate(nil)
	if err == nil {
		t.Fatal("expected error for nil NEIR")
	}
}

func TestGenerateMapInput(t *testing.T) {
	neir := map[string]any{
		"project": "my-project",
		"modules": []any{
			map[string]any{"name": "auth", "path": "./internal/auth"},
			map[string]any{"name": "api", "path": "./internal/api"},
		},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(artifacts) == 0 {
		t.Fatal("expected artifacts from map input")
	}
	foundAuth := false
	foundApi := false
	for _, a := range artifacts {
		if strings.Contains(a.Path, "auth") {
			foundAuth = true
		}
		if strings.Contains(a.Path, "api") {
			foundApi = true
		}
	}
	if !foundAuth {
		t.Error("expected auth module artifact")
	}
	if !foundApi {
		t.Error("expected api module artifact")
	}
}

func TestGenerateMapInputWithServices(t *testing.T) {
	neir := map[string]any{
		"project": "svc-proj",
		"services": []any{
			map[string]any{"name": "api-gateway", "port": 8080, "kind": "http"},
			map[string]any{"name": "worker", "port": 9090, "kind": "grpc"},
		},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(artifacts) < 2 {
		t.Fatalf("expected at least 2 service artifacts, got %d", len(artifacts))
	}
	for _, a := range artifacts {
		if !strings.Contains(a.Path, "config.yaml") {
			continue
		}
		content := string(a.Content)
		if !strings.Contains(content, "api-gateway") && !strings.Contains(content, "worker") {
			t.Errorf("service config should contain service name, got: %s", content)
		}
	}
}

func TestGenerateNEIRWithArchitecture(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "arch-proj"},
		Architecture: &architecture.Architecture{
			Pattern: architecture.PatternClean,
		},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	foundArchDoc := false
	for _, a := range artifacts {
		if a.Path == "docs/architecture.md" {
			foundArchDoc = true
			if !strings.Contains(string(a.Content), "clean") {
				t.Error("architecture doc should contain pattern name")
			}
			if !strings.Contains(string(a.Content), "arch-proj") {
				t.Error("architecture doc should contain project name")
			}
		}
	}
	if !foundArchDoc {
		t.Error("expected architecture doc artifact")
	}
}

func TestGenerateNEIRWithDeployment(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "deploy-proj"},
		Deployment: &deployment.Deployment{
			Strategy: deployment.StrategyRolling,
		},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	foundDockerCompose := false
	for _, a := range artifacts {
		if a.Path == "docker-compose.yml" {
			foundDockerCompose = true
			content := string(a.Content)
			if !strings.Contains(content, "services:") {
				t.Error("docker-compose.yml should contain services")
			}
			if strings.Contains(content, "version:") {
				t.Error("docker-compose.yml should not contain deprecated version key")
			}
		}
	}
	if !foundDockerCompose {
		t.Error("expected docker-compose.yml artifact")
	}
}

func TestGenerateNEIRWithDeploymentContent(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "deploy-proj"},
		Deployment: &deployment.Deployment{
			Strategy: deployment.StrategyBlueGreen,
		},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	for _, a := range artifacts {
		if a.Path == "docker-compose.yml" {
			content := string(a.Content)
			if !strings.Contains(content, "services:") {
				t.Error("docker-compose.yml should contain 'services:'")
			}
			if !strings.Contains(content, "build: .") {
				t.Error("docker-compose.yml should contain 'build: .'")
			}
		}
	}
}

func TestGenerateForLanguageJava(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "java-app"},
		Modules: []module.Module{{Name: "auth", Path: "./internal/auth"}},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageJava)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}

	foundPom := false
	foundMain := false
	foundDockerfile := false
	for _, a := range artifacts {
		if a.Path == "pom.xml" {
			foundPom = true
			if !strings.Contains(string(a.Content), "java-app") {
				t.Error("pom.xml should contain project name")
			}
		}
		if a.Path == "src/main.java" {
			foundMain = true
			if !strings.Contains(string(a.Content), "java-app") {
				t.Error("main.java should contain project name")
			}
		}
		if a.Path == "Dockerfile" {
			foundDockerfile = true
			if !strings.Contains(string(a.Content), "eclipse-temurin") {
				t.Error("Dockerfile should use Java image")
			}
		}
	}
	if !foundPom {
		t.Error("expected pom.xml artifact")
	}
	if !foundMain {
		t.Error("expected src/main.java artifact")
	}
	if !foundDockerfile {
		t.Error("expected Dockerfile artifact")
	}
}

func TestGenerateForLanguageRust(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "rust-app"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageRust)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}

	foundCargo := false
	foundMain := false
	foundDockerfile := false
	for _, a := range artifacts {
		if a.Path == "Cargo.toml" {
			foundCargo = true
			if !strings.Contains(string(a.Content), "rust-app") {
				t.Error("Cargo.toml should contain project name")
			}
		}
		if a.Path == "src/main.rs" {
			foundMain = true
			if !strings.Contains(string(a.Content), "rust-app") {
				t.Error("main.rs should contain project name")
			}
		}
		if a.Path == "Dockerfile" {
			foundDockerfile = true
			if !strings.Contains(string(a.Content), "rust") {
				t.Error("Dockerfile should use Rust image")
			}
		}
	}
	if !foundCargo {
		t.Error("expected Cargo.toml artifact")
	}
	if !foundMain {
		t.Error("expected src/main.rs artifact")
	}
	if !foundDockerfile {
		t.Error("expected Dockerfile artifact")
	}
}

func TestGenerateModuleWithoutPath(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test-proj"},
		Modules: []module.Module{{Name: "auth"}},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	found := false
	for _, a := range artifacts {
		if strings.Contains(a.Path, "auth") {
			found = true
		}
	}
	if !found {
		t.Error("expected artifact for module without explicit path")
	}
}

func TestGenerateModuleEmptyNameSkipped(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test-proj"},
		Modules: []module.Module{{Name: "", Path: "./internal/empty"}},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	for _, a := range artifacts {
		if strings.Contains(a.Path, "empty") {
			t.Error("module with empty name should be skipped")
		}
	}
}

func TestGenerateServicesViaNEIRStruct(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "svc-struct"},
		Services: []service.Service{
			{Name: "api", Port: 8080, Kind: service.KindHTTP},
			{Name: "grpc-svc", Port: 9090, Kind: service.KindGRPC},
		},
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(artifacts) < 2 {
		t.Fatalf("expected at least 2 service artifacts, got %d", len(artifacts))
	}
	foundApi := false
	foundGrpc := false
	for _, a := range artifacts {
		if strings.Contains(a.Path, "api") && strings.Contains(a.Path, "config.yaml") {
			foundApi = true
			content := string(a.Content)
			if !strings.Contains(content, "8080") {
				t.Error("api config should contain port 8080")
			}
		}
		if strings.Contains(a.Path, "grpc-svc") && strings.Contains(a.Path, "config.yaml") {
			foundGrpc = true
			content := string(a.Content)
			if !strings.Contains(content, "9090") {
				t.Error("grpc-svc config should contain port 9090")
			}
		}
	}
	if !foundApi {
		t.Error("expected api service config")
	}
	if !foundGrpc {
		t.Error("expected grpc-svc service config")
	}
}

func TestGenerateMapInputWithArchitectureAndDeployment(t *testing.T) {
	neir := map[string]any{
		"project": "full-proj",
	}

	engine := NewEngine()
	artifacts, err := engine.Generate(neir)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(artifacts) != 0 {
		t.Fatalf("expected 0 artifacts for map without modules/services/arch/deploy, got %d", len(artifacts))
	}
}

func TestGenerateForLanguageNoProject(t *testing.T) {
	neir := &model.NEIR{}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageGo)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}
	if len(artifacts) == 0 {
		t.Fatal("expected at least main.go and Dockerfile even without project")
	}
}

func TestGenerateForLanguageWithServices(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "svc-proj"},
		Services: []service.Service{
			{Name: "api", Port: 8080, Kind: service.KindHTTP},
		},
		Modules: []module.Module{{Name: "auth", Path: "./internal/auth"}},
	}

	engine := NewEngine()
	artifacts, err := engine.GenerateForLanguage(neir, language.LanguageGo)
	if err != nil {
		t.Fatalf("GenerateForLanguage returned error: %v", err)
	}
	foundModuleFile := false
	for _, a := range artifacts {
		if strings.Contains(a.Path, "auth") && strings.HasSuffix(a.Path, ".go") {
			foundModuleFile = true
			content := string(a.Content)
			if !strings.Contains(content, "package auth") {
				t.Error("module file should contain package declaration")
			}
		}
	}
	if !foundModuleFile {
		t.Error("expected module Go file")
	}
}


