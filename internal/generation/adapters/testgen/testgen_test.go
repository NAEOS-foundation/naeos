// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package testgen

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func TestGenerateTestsNil(t *testing.T) {
	_, err := GenerateTests(nil)
	if err == nil {
		t.Error("expected error for nil NEIR")
	}
}

func TestGenerateTestsEmpty(t *testing.T) {
	neir := &model.NEIR{
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageGo},
		},
	}
	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 0 {
		t.Errorf("expected 0 artifacts for empty NEIR, got %d", len(artifacts))
	}
}

func TestGenerateTestsGo(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "myproject"},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageGo},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core"},
		},
	}

	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(artifacts) < 2 {
		t.Fatalf("expected at least 2 artifacts (module test + integration test), got %d", len(artifacts))
	}

	foundModuleTest := false
	foundIntegration := false
	for _, a := range artifacts {
		if strings.Contains(a.Path, "core_test") {
			foundModuleTest = true
			content := string(a.Content)
			if !strings.Contains(content, "package core") {
				t.Error("expected 'package core' in module test")
			}
			if !strings.Contains(content, "TestCoreModule") {
				t.Error("expected TestCoreModule in test content")
			}
		}
		if strings.Contains(a.Path, "integration_test") {
			foundIntegration = true
		}
	}

	if !foundModuleTest {
		t.Error("expected module test artifact")
	}
	if !foundIntegration {
		t.Error("expected integration test artifact")
	}
}

func TestGenerateTestsTypeScript(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "tsproject"},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageTypeScript},
		},
		Modules: []module.Module{
			{Name: "auth", Path: "./src/auth"},
		},
	}

	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(artifacts) < 2 {
		t.Fatalf("expected at least 2 artifacts, got %d", len(artifacts))
	}

	for _, a := range artifacts {
		content := string(a.Content)
		if strings.Contains(a.Path, "auth_test") {
			if !strings.Contains(content, "vitest") {
				t.Error("expected vitest import in TypeScript test")
			}
		}
	}
}

func TestGenerateTestsPython(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "pyproject"},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguagePython},
		},
		Modules: []module.Module{
			{Name: "utils", Path: "./utils"},
		},
	}

	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, a := range artifacts {
		content := string(a.Content)
		if strings.Contains(a.Path, "utils_test") {
			if !strings.Contains(content, "unittest") {
				t.Error("expected unittest import in Python test")
			}
		}
	}
}

func TestGenerateTestsJava(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "javaproject"},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageJava},
		},
		Modules: []module.Module{
			{Name: "service", Path: "./src/main/java"},
		},
	}

	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, a := range artifacts {
		content := string(a.Content)
		if strings.Contains(a.Path, "service_test") {
			if !strings.Contains(content, "junit") {
				t.Error("expected junit import in Java test")
			}
		}
	}
}

func TestGenerateTestsRust(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "rustproject"},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageRust},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./src/core"},
		},
	}

	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, a := range artifacts {
		content := string(a.Content)
		if strings.Contains(a.Path, "core_test") {
			if !strings.Contains(content, "#[cfg(test)]") {
				t.Error("expected #[cfg(test)] in Rust test")
			}
		}
	}
}

func TestGenerateTestsWithServices(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "svcproject"},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageGo},
		},
		Services: []service.Service{
			{
				Name: "api",
				Kind: service.KindHTTP,
				Port: 8080,
			},
		},
	}

	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundSvcTest := false
	for _, a := range artifacts {
		if strings.Contains(a.Path, "api_test") {
			foundSvcTest = true
			content := string(a.Content)
			if !strings.Contains(content, "package api") {
				t.Error("expected 'package api' in service test")
			}
		}
	}

	if !foundSvcTest {
		t.Error("expected service test artifact")
	}
}

func TestGenerateTestsMultipleLanguages(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "multiproject"},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{
				language.LanguageGo,
				language.LanguageTypeScript,
			},
		},
		Modules: []module.Module{
			{Name: "core", Path: "./core"},
		},
	}

	artifacts, err := GenerateTests(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	goTests := 0
	tsTests := 0
	for _, a := range artifacts {
		if strings.HasSuffix(a.Path, "_test.go") {
			goTests++
		}
		if strings.HasSuffix(a.Path, "_test.ts") {
			tsTests++
		}
	}

	if goTests == 0 {
		t.Error("expected Go test files")
	}
	if tsTests == 0 {
		t.Error("expected TypeScript test files")
	}
}
