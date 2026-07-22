package language

import "testing"

func TestIsValid(t *testing.T) {
	tests := []struct {
		lang Language
		want bool
	}{
		{LanguageGo, true},
		{LanguageTypeScript, true},
		{LanguagePython, true},
		{LanguageJava, true},
		{LanguageRust, true},
		{Language("cobol"), false},
		{Language(""), false},
	}
	for _, tt := range tests {
		got := IsValid(tt.lang)
		if got != tt.want {
			t.Errorf("IsValid(%s) = %v, want %v", tt.lang, got, tt.want)
		}
	}
}

func TestAll(t *testing.T) {
	langs := All()
	if len(langs) != 5 {
		t.Errorf("expected 5 languages, got %d", len(langs))
	}
	seen := make(map[Language]bool)
	for _, l := range langs {
		seen[l] = true
	}
	if !seen[LanguageGo] || !seen[LanguageRust] {
		t.Error("missing Go or Rust in All()")
	}
}

func TestExtensions(t *testing.T) {
	tests := []struct {
		lang Language
		want int
	}{
		{LanguageGo, 3},
		{LanguageTypeScript, 4},
		{LanguagePython, 3},
		{LanguageJava, 3},
		{LanguageRust, 2},
		{Language("unknown"), 0},
	}
	for _, tt := range tests {
		exts := Extensions(tt.lang)
		if len(exts) != tt.want {
			t.Errorf("Extensions(%s) returned %d, want %d", tt.lang, len(exts), tt.want)
		}
	}
}

func TestExtensions_ContainsGo(t *testing.T) {
	exts := Extensions(LanguageGo)
	found := false
	for _, e := range exts {
		if e == ".go" {
			found = true
		}
	}
	if !found {
		t.Error("expected .go extension for Go")
	}
}

func TestBuildFile(t *testing.T) {
	tests := []struct {
		lang Language
		want string
	}{
		{LanguageGo, "go.mod"},
		{LanguageTypeScript, "package.json"},
		{LanguagePython, "pyproject.toml"},
		{LanguageJava, "pom.xml"},
		{LanguageRust, "Cargo.toml"},
		{Language("unknown"), ""},
	}
	for _, tt := range tests {
		got := BuildFile(tt.lang)
		if got != tt.want {
			t.Errorf("BuildFile(%s) = %s, want %s", tt.lang, got, tt.want)
		}
	}
}

func TestDockerBaseImage(t *testing.T) {
	tests := []struct {
		lang Language
		want string
	}{
		{LanguageGo, "golang:1.22-alpine"},
		{LanguageTypeScript, "node:22-alpine"},
		{LanguagePython, "python:3.12-slim"},
		{LanguageJava, "eclipse-temurin:21-jdk-alpine"},
		{LanguageRust, "rust:1.78-alpine"},
		{Language("unknown"), "alpine:latest"},
	}
	for _, tt := range tests {
		got := DockerBaseImage(tt.lang)
		if got != tt.want {
			t.Errorf("DockerBaseImage(%s) = %s, want %s", tt.lang, got, tt.want)
		}
	}
}

func TestDockerRuntimeImage(t *testing.T) {
	tests := []struct {
		lang Language
		want string
	}{
		{LanguageGo, "alpine:3.19"},
		{LanguageTypeScript, "node:22-alpine"},
		{LanguagePython, "python:3.12-slim"},
		{LanguageJava, "eclipse-temurin:21-jre-alpine"},
		{LanguageRust, "alpine:3.19"},
		{Language("unknown"), "alpine:latest"},
	}
	for _, tt := range tests {
		got := DockerRuntimeImage(tt.lang)
		if got != tt.want {
			t.Errorf("DockerRuntimeImage(%s) = %s, want %s", tt.lang, got, tt.want)
		}
	}
}

func TestLanguageConstants(t *testing.T) {
	tests := []struct {
		lang Language
		want string
	}{
		{LanguageGo, "go"},
		{LanguageTypeScript, "typescript"},
		{LanguagePython, "python"},
		{LanguageJava, "java"},
		{LanguageRust, "rust"},
	}
	for _, tt := range tests {
		if string(tt.lang) != tt.want {
			t.Errorf("Language(%s) = %s, want %s", tt.want, string(tt.lang), tt.want)
		}
	}
}
