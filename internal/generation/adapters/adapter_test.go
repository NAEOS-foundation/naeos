package adapters

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func TestGenerateForNEIR_Nil(t *testing.T) {
	t.Parallel()
	artifacts, err := GenerateForNEIR(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 0 {
		t.Fatalf("expected 0 artifacts, got %d", len(artifacts))
	}
}

func TestGenerateForNEIR_EmptyNEIR(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{}
	artifacts, err := GenerateForNEIR(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) == 0 {
		t.Fatal("expected artifacts, got none")
	}
}

func TestGenerateForNEIR_WithGeneration(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test-proj"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
		},
		Services: []service.Service{
			{Name: "api", Kind: "http", Port: 8080},
		},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguageGo},
		},
	}
	artifacts, err := GenerateForNEIR(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) == 0 {
		t.Fatal("expected artifacts, got none")
	}
}

func TestGenerateForNEIR_UnknownLanguageSkipped(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test-proj"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core"},
		},
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{"fortran77"},
		},
	}
	artifacts, err := GenerateForNEIR(neir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 0 {
		t.Fatalf("expected 0 artifacts for unknown language, got %d", len(artifacts))
	}
}

func TestResolveLanguages_Empty(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{}
	langs := resolveLanguages(neir)
	if len(langs) != 1 {
		t.Fatalf("expected 1 default language, got %d", len(langs))
	}
	if langs[0] != language.LanguageGo {
		t.Fatalf("expected go as default, got %s", langs[0])
	}
}

func TestResolveLanguages_FromGeneration(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Generation: &generation.GenerationConfig{
			Languages: []language.Language{language.LanguagePython, language.LanguageRust},
		},
	}
	langs := resolveLanguages(neir)
	if len(langs) != 2 {
		t.Fatalf("expected 2 languages, got %d", len(langs))
	}
	if langs[0] != language.LanguagePython || langs[1] != language.LanguageRust {
		t.Fatalf("unexpected languages: %v", langs)
	}
}

func TestAll(t *testing.T) {
	t.Parallel()
	all := All()
	if len(all) == 0 {
		t.Fatal("expected registered adapters, got none")
	}
	count := 0
	for _, adapters := range all {
		count += len(adapters)
	}
	if count != 7 {
		t.Fatalf("expected 7 registered adapters, got %d", count)
	}
}
