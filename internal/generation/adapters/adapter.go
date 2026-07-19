package adapters

import (
	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
)

type OutputAdapter interface {
	Language() language.Language
	Framework() string
	GenerateProject(projectName string) []engine.Artifact
	GenerateModule(moduleName, modulePath, projectName string) []engine.Artifact
	GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []engine.Artifact
	GenerateDockerfile(projectName string) []engine.Artifact
	GenerateCI(projectName string) []engine.Artifact
	GenerateDockerCompose(projectName string) []engine.Artifact
	GenerateArchitectureDoc(projectName, pattern string) []engine.Artifact
}

var adapters = map[language.Language][]OutputAdapter{}

func Register(adapter OutputAdapter) {
	lang := adapter.Language()
	adapters[lang] = append(adapters[lang], adapter)
}

func Get(lang language.Language) (OutputAdapter, bool) {
	adapters, ok := adapters[lang]
	if !ok || len(adapters) == 0 {
		return nil, false
	}
	for _, a := range adapters {
		if a.Framework() == "" {
			return a, true
		}
	}
	return adapters[0], true
}

func GetFramework(lang language.Language, framework string) (OutputAdapter, bool) {
	adapters, ok := adapters[lang]
	if !ok {
		return nil, false
	}
	for _, a := range adapters {
		if a.Framework() == framework {
			return a, true
		}
	}
	return nil, false
}

func All() map[language.Language][]OutputAdapter {
	result := make(map[language.Language][]OutputAdapter, len(adapters))
	for k, v := range adapters {
		result[k] = v
	}
	return result
}

func GenerateForNEIR(neir *model.NEIR) ([]engine.Artifact, error) {
	if neir == nil {
		return nil, nil // Nil input — nothing to generate, return empty result
	}

	languages := resolveLanguages(neir)
	var allArtifacts []engine.Artifact

	for _, lang := range languages {
		adapter, ok := Get(lang)
		if !ok {
			continue
		}
		artifacts := generateWithAdapter(adapter, neir)
		allArtifacts = append(allArtifacts, artifacts...)
	}

	return allArtifacts, nil
}

func resolveLanguages(neir *model.NEIR) []language.Language {
	if neir.Generation != nil && len(neir.Generation.Languages) > 0 {
		return neir.Generation.Languages
	}
	return []language.Language{language.LanguageGo}
}

func generateWithAdapter(adapter OutputAdapter, neir *model.NEIR) []engine.Artifact {
	var artifacts []engine.Artifact

	projectName := ""
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	artifacts = append(artifacts, adapter.GenerateProject(projectName)...)
	artifacts = append(artifacts, adapter.GenerateDockerfile(projectName)...)
	artifacts = append(artifacts, adapter.GenerateCI(projectName)...)

	for _, m := range neir.Modules {
		artifacts = append(artifacts, adapter.GenerateModule(m.Name, m.Path, projectName)...)
	}

	for _, s := range neir.Services {
		artifacts = append(artifacts, adapter.GenerateService(s.Name, string(s.Kind), s.Port, projectName)...)
	}

	if neir.Deployment != nil && string(neir.Deployment.Strategy) != "" {
		artifacts = append(artifacts, adapter.GenerateDockerCompose(projectName)...)
	}

	if neir.Architecture != nil && string(neir.Architecture.Pattern) != "" {
		artifacts = append(artifacts, adapter.GenerateArchitectureDoc(projectName, string(neir.Architecture.Pattern))...)
	}

	return artifacts
}
