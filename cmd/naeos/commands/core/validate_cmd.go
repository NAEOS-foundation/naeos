// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	clihelpers "github.com/NAEOS-foundation/naeos/cmd/naeos/internal/clihelpers"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func NewValidateCommand() *cobra.Command {
	var configPath, input, inputFile, outputFormat, outputFile string
	var languages []string

	cmd := &cobra.Command{
		Use:     "validate",
		Aliases: []string{"v"},
		Short:   "Validate a specification using the NAEOS pipeline",
		Long: `Validate a specification file through the NAEOS pipeline without generating artifacts.

Output formats:
  text  — human-readable text output (default)
  json  — structured JSON with error codes and field locations

Error codes:
  SPEC_EMPTY        — specification is empty
  SPEC_INVALID_YAML — specification contains invalid YAML
  PROJECT_MISSING   — project section is missing
  PROJECT_NAME_MISSING — project name is missing
  SERVICE_DUPLICATE — duplicate service name
  PORT_INVALID      — invalid port number
  PIPELINE_FAILED   — pipeline validation failed

Example:
  naeos validate --input spec.yaml
  naeos validate --input spec.yaml --output json
  naeos v --input-file spec.yaml --output json --output-file result.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputValue, err := clihelpers.LoadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, err := clihelpers.LoadPipelineConfig(configPath, clihelpers.CLIVerbose, languages, clihelpers.CLIDryRun, "")
			if err != nil {
				return err
			}

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			result, err := p.Validate(inputValue)
			if err != nil {
				vr := clihelpers.ValidationResult{
					Valid:  false,
					Status: "invalid",
					Errors: []clihelpers.ValidationError{{
						Code:    "PIPELINE_FAILED",
						Message: err.Error(),
					}},
					Summary: "validation failed",
				}
				return clihelpers.RenderValidation(cmd, vr, outputFormat, outputFile)
			}

			vr := clihelpers.ValidationResult{
				Valid:   true,
				Status:  "valid",
				Summary: fmt.Sprintf("valid — project: %s, services: %d", result.NEIR.Project.Name, len(result.NEIR.Services)),
			}

			if inputValue == "" {
				vr.Valid = false
				vr.Status = "invalid"
				vr.Errors = append(vr.Errors, clihelpers.ValidationError{
					Code:    "SPEC_EMPTY",
					Message: "specification is empty",
				})
				vr.Summary = "validation failed — empty spec"
			}

			return clihelpers.RenderValidation(cmd, vr, outputFormat, outputFile)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text, json")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write the output")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language for code generation")
	return cmd
}

func NewInspectCommand() *cobra.Command {
	var configPath, input, inputFile, outputFormat, outputFile string

	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect the NAEOS pipeline result",
		Long: `Inspect the pipeline result showing project details, artifacts, and tasks.

Example:
  naeos inspect --config config.yaml --input spec.yaml
  naeos inspect --config config.yaml --input-file spec.yaml --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputValue, err := clihelpers.LoadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, err := clihelpers.LoadPipelineConfig(configPath, clihelpers.CLIVerbose, nil, clihelpers.CLIDryRun, "")
			if err != nil {
				return err
			}

			specInput, err := clihelpers.ResolveInput(inputValue)
			if err != nil {
				return fmt.Errorf("resolve input: %w", err)
			}

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			result, err := p.Run(specInput)
			if err != nil {
				return fmt.Errorf("inspect run failed: %w", err)
			}

			projectName := ""
			if result.NEIR != nil && result.NEIR.Project != nil {
				projectName = result.NEIR.Project.Name
			}

			payload := map[string]any{
				"pipeline":     cfg.Name,
				"mode":         cfg.Mode,
				"verbose":      cfg.Verbose,
				"output_dir":   cfg.OutputDir,
				"project":      projectName,
				"input":        specInput,
				"artifacts":    len(result.Artifacts),
				"tasks":        len(result.Tasks),
				"source_words": len(strings.Fields(result.Source)),
			}

			rendered, err := clihelpers.RenderOutput(payload, outputFormat, func() []byte {
				return []byte(fmt.Sprintf("pipeline=%s mode=%s verbose=%t output_dir=%s\nproject=%s artifacts=%d tasks=%d input=%q\n", cfg.Name, cfg.Mode, cfg.Verbose, cfg.OutputDir, projectName, len(result.Artifacts), len(result.Tasks), specInput))
			})
			if err != nil {
				return err
			}

			return clihelpers.WriteOrPrint(cmd, rendered, outputFile)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input or file path to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text, json, or yaml")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write the formatted output")
	return cmd
}
