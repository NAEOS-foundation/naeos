package adapters

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

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
		return nil, nil
	}

	languages := resolveLanguages(neir)
	if len(languages) <= 1 {
		return generateSequential(neir, languages)
	}

	return generateParallel(neir, languages)
}

func generateSequential(neir *model.NEIR, languages []language.Language) ([]engine.Artifact, error) {
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

func generateParallel(neir *model.NEIR, languages []language.Language) ([]engine.Artifact, error) {
	var mu sync.Mutex
	var allArtifacts []engine.Artifact

	g, _ := errgroup.WithContext(context.Background())
	for _, lang := range languages {
		lang := lang
		g.Go(func() error {
			adapter, ok := Get(lang)
			if !ok {
				return nil
			}
			artifacts := generateWithAdapter(adapter, neir)
			mu.Lock()
			allArtifacts = append(allArtifacts, artifacts...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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
	var mu sync.Mutex

	projectName := ""
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	artifacts = append(artifacts, adapter.GenerateProject(projectName)...)
	artifacts = append(artifacts, adapter.GenerateDockerfile(projectName)...)
	artifacts = append(artifacts, adapter.GenerateCI(projectName)...)

	if neir.Deployment != nil && string(neir.Deployment.Strategy) != "" {
		artifacts = append(artifacts, adapter.GenerateDockerCompose(projectName)...)
	}

	if neir.Architecture != nil && string(neir.Architecture.Pattern) != "" {
		artifacts = append(artifacts, adapter.GenerateArchitectureDoc(projectName, string(neir.Architecture.Pattern))...)
	}

	g, _ := errgroup.WithContext(context.Background())
	for _, m := range neir.Modules {
		m := m
		g.Go(func() error {
			moduleArtifacts := adapter.GenerateModule(m.Name, m.Path, projectName)
			mu.Lock()
			artifacts = append(artifacts, moduleArtifacts...)
			mu.Unlock()
			return nil
		})
	}

	for _, s := range neir.Services {
		s := s
		g.Go(func() error {
			serviceArtifacts := adapter.GenerateService(s.Name, string(s.Kind), s.Port, projectName)
			mu.Lock()
			artifacts = append(artifacts, serviceArtifacts...)
			mu.Unlock()
			return nil
		})
	}

	_ = g.Wait()
	return artifacts
}
