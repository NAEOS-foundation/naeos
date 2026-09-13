// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package language

type Language string

const (
	LanguageGo         Language = "go"
	LanguageTypeScript Language = "typescript"
	LanguagePython     Language = "python"
	LanguageJava       Language = "java"
	LanguageRust       Language = "rust"
)

var supported = map[Language]bool{
	LanguageGo:         true,
	LanguageTypeScript: true,
	LanguagePython:     true,
	LanguageJava:       true,
	LanguageRust:       true,
}

func IsValid(lang Language) bool {
	return supported[lang]
}

func All() []Language {
	return []Language{
		LanguageGo,
		LanguageTypeScript,
		LanguagePython,
		LanguageJava,
		LanguageRust,
	}
}

func Extensions(lang Language) []string {
	switch lang {
	case LanguageGo:
		return []string{".go", ".mod", ".sum"}
	case LanguageTypeScript:
		return []string{".ts", ".tsx", ".js", ".json"}
	case LanguagePython:
		return []string{".py", ".toml", ".cfg"}
	case LanguageJava:
		return []string{".java", ".xml", ".gradle"}
	case LanguageRust:
		return []string{".rs", ".toml"}
	default:
		return nil
	}
}

func BuildFile(lang Language) string {
	switch lang {
	case LanguageGo:
		return "go.mod"
	case LanguageTypeScript:
		return "package.json"
	case LanguagePython:
		return "pyproject.toml"
	case LanguageJava:
		return "pom.xml"
	case LanguageRust:
		return "Cargo.toml"
	default:
		return ""
	}
}

func DockerBaseImage(lang Language) string {
	switch lang {
	case LanguageGo:
		return "golang:1.22-alpine"
	case LanguageTypeScript:
		return "node:22-alpine"
	case LanguagePython:
		return "python:3.12-slim"
	case LanguageJava:
		return "eclipse-temurin:21-jdk-alpine"
	case LanguageRust:
		return "rust:1.78-alpine"
	default:
		return "alpine:latest"
	}
}

func DockerRuntimeImage(lang Language) string {
	switch lang {
	case LanguageGo:
		return "alpine:3.19"
	case LanguageTypeScript:
		return "node:22-alpine"
	case LanguagePython:
		return "python:3.12-slim"
	case LanguageJava:
		return "eclipse-temurin:21-jre-alpine"
	case LanguageRust:
		return "alpine:3.19"
	default:
		return "alpine:latest"
	}
}
