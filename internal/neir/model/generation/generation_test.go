package generation

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
)

func TestGenerationConfig_HasLanguage(t *testing.T) {
	gc := &GenerationConfig{
		Languages: []language.Language{language.LanguageGo, language.LanguagePython},
	}
	if !gc.HasLanguage(language.LanguageGo) {
		t.Error("expected HasLanguage(Go) to be true")
	}
	if gc.HasLanguage(language.LanguageRust) {
		t.Error("expected HasLanguage(Rust) to be false")
	}
}

func TestGenerationConfig_DefaultLanguage(t *testing.T) {
	gc := &GenerationConfig{
		Languages: []language.Language{language.LanguageTypeScript},
	}
	if gc.DefaultLanguage() != language.LanguageTypeScript {
		t.Errorf("expected TypeScript, got %s", gc.DefaultLanguage())
	}
}

func TestGenerationConfig_DefaultLanguageEmpty(t *testing.T) {
	gc := &GenerationConfig{}
	if gc.DefaultLanguage() != language.LanguageGo {
		t.Errorf("expected Go default, got %s", gc.DefaultLanguage())
	}
}

func TestGenerationConfig_ZeroValue(t *testing.T) {
	var gc GenerationConfig
	if gc.Languages != nil {
		t.Error("expected nil Languages")
	}
	if gc.OutputDir != "" {
		t.Error("expected empty OutputDir")
	}
}
