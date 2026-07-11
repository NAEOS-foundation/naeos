package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	compileradapters "github.com/NAEOS-foundation/naeos/internal/compiler/adapters"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newCompileCommand() *cobra.Command {
	var target string
	var all bool
	var outDir string

	cmd := &cobra.Command{
		Use:   "compile [spec-file]",
		Short: "Compile spec into AI instruction sets for target tools",
		Long: `Compile a NAEOS specification into AI instruction files for various tools.
Supported targets: copilot, claude, cursor, gemini, codex, opencode.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specFile := args[0]
			content, err := os.ReadFile(specFile)
			if err != nil {
				return fmt.Errorf("read spec: %w", err)
			}

			p := parser.NewParser()
			doc, err := p.Parse(string(content))
			if err != nil {
				return fmt.Errorf("parse spec: %w", err)
			}

			neir, err := buildNEIRFromSpec(doc)
			if err != nil {
				return fmt.Errorf("build NEIR: %w", err)
			}

			comp := compiler.New()
			comp.Register(compileradapters.NewCopilotAdapter())
			comp.Register(compileradapters.NewClaudeAdapter())
			comp.Register(compileradapters.NewCursorAdapter())
			comp.Register(compileradapters.NewGeminiAdapter())
			comp.Register(compileradapters.NewCodexAdapter())
			comp.Register(compileradapters.NewOpenCodeAdapter())

			if all {
				results := comp.CompileAll(neir)
				for tgt, out := range results {
					fmt.Printf("  %s: %s\n", tgt, out.Summary)
					for _, f := range out.Files {
						path := filepath.Join(outDir, f.Path)
						if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
							return fmt.Errorf("create dir: %w", err)
						}
						if err := os.WriteFile(path, []byte(f.Content), 0o644); err != nil {
							return fmt.Errorf("write %s: %w", path, err)
						}
						fmt.Printf("    -> %s\n", path)
					}
				}
				return nil
			}

			t := compiler.Target(target)
			result, err := comp.Compile(neir, t)
			if err != nil {
				return fmt.Errorf("compile: %w", err)
			}

			fmt.Println(result.Summary)
			for _, f := range result.Files {
				path := filepath.Join(outDir, f.Path)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return fmt.Errorf("create dir: %w", err)
				}
				if err := os.WriteFile(path, []byte(f.Content), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", path, err)
				}
				fmt.Printf("  -> %s\n", path)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", "copilot", "target AI tool (copilot, claude, cursor, gemini, codex, opencode)")
	cmd.Flags().BoolVar(&all, "all", false, "compile for all targets")
	cmd.Flags().StringVar(&outDir, "output", ".", "output directory")
	return cmd
}

func buildNEIRFromSpec(doc *parser.SpecDocument) (*model.NEIR, error) {
	neir := &model.NEIR{}

	if doc.Project != "" {
		p := project.Project{}
		data, _ := yaml.Marshal(map[string]string{"name": doc.Project})
		_ = yaml.Unmarshal(data, &p)
		neir.Project = &p
	}

	for _, m := range doc.Modules {
		neir.Modules = append(neir.Modules, module.Module{
			Name:        m.Name,
			Path:        m.Path,
			Description: m.Description,
		})
	}

	for _, s := range doc.Services {
		neir.Services = append(neir.Services, service.Service{
			Name: s.Name,
			Kind: service.ServiceKind(s.Kind),
			Port: s.Port,
		})
	}

	if doc.Architecture != nil {
		neir.Architecture = &architecture.Architecture{
			Pattern:    architecture.Pattern(doc.Architecture.Pattern),
			Principles: doc.Architecture.Principles,
		}
	}

	return neir, nil
}
