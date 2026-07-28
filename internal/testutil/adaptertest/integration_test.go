package adaptertest

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	adapters "github.com/NAEOS-foundation/naeos/internal/generation/adapters"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
)

type stubAdapter struct {
	target compiler.Target
}

func (s *stubAdapter) Target() compiler.Target {
	return s.target
}

func (s *stubAdapter) Compile(neir *model.NEIR) (*compiler.CompiledOutput, error) {
	return &compiler.CompiledOutput{
		Target:  s.target,
		Summary: "compiled " + neir.Project.Name + " for " + string(s.target),
		Files: []compiler.OutputFile{
			{Path: string(s.target) + "/instructions.md", Content: "# " + string(s.target) + " instructions\n", Kind: "instructions"},
			{Path: string(s.target) + "/context.md", Content: "## Context\nProject: " + neir.Project.Name + "\n", Kind: "context"},
		},
	}, nil
}

func TestCompilerAdapterIntegration(t *testing.T) {
	n := FullNEIR()
	c := compiler.New()

	adapters := []compiler.Target{
		compiler.TargetCopilot,
		compiler.TargetClaude,
		compiler.TargetCursor,
		compiler.TargetGemini,
		compiler.TargetCodex,
		compiler.TargetOpenCode,
		compiler.TargetWindsurf,
	}

	for _, target := range adapters {
		c.Register(&stubAdapter{target: target})
	}

	results := c.CompileAll(n)

	for target, output := range results {
		ValidateCompilerOutput(t, output)
		if output.Target != target {
			t.Errorf("target mismatch: expected %s, got %s", target, output.Target)
		}
		if len(output.Files) < 1 {
			t.Errorf("expected at least 1 file for %s", target)
		}
	}
}

func TestGenerationAdapterIntegration(t *testing.T) {
	n := FullNEIR()

	langs := []language.Language{
		language.LanguageGo,
		language.LanguagePython,
		language.LanguageTypeScript,
		language.LanguageRust,
		language.LanguageJava,
	}

	for _, lang := range langs {
		adapter, ok := adapters.Get(lang)
		if !ok {
			t.Fatalf("no adapter for language %s", lang)
		}

		t.Run(string(lang), func(t *testing.T) {
			projectName := n.Project.Name
			if projectName == "" {
				projectName = "test-project"
			}

			artifacts := adapter.GenerateProject(projectName)
			ValidateArtifacts(t, artifacts)

			for _, mod := range n.Modules {
				modArtifacts := adapter.GenerateModule(mod.Name, mod.Path, projectName)
				ValidateArtifacts(t, modArtifacts)
			}

			for _, svc := range n.Services {
				svcArtifacts := adapter.GenerateService(svc.Name, string(svc.Kind), svc.Port, projectName)
				ValidateArtifacts(t, svcArtifacts)
			}

			dockerArtifacts := adapter.GenerateDockerfile(projectName)
			ValidateArtifacts(t, dockerArtifacts)

			ciArtifacts := adapter.GenerateCI(projectName)
			ValidateArtifacts(t, ciArtifacts)
		})
	}
}

func TestFullPipelineIntegration(t *testing.T) {
	n := FullNEIR()

	compileOK := false
	generateOK := false

	c := compiler.New()
	c.Register(&stubAdapter{target: compiler.TargetCopilot})
	c.Register(&stubAdapter{target: compiler.TargetClaude})

	compileResults := c.CompileAll(n)
	for target, output := range compileResults {
		if output != nil && len(output.Files) > 0 {
			compileOK = true
		}
		_ = target
	}
	if !compileOK {
		t.Error("compilation produced no valid output")
	}

	adapter, ok := adapters.Get(language.LanguageGo)
	if ok {
		artifacts := adapter.GenerateProject(n.Project.Name)
		if len(artifacts) > 0 {
			generateOK = true
		}
	}
	if !generateOK {
		t.Error("generation produced no valid output")
	}
}
