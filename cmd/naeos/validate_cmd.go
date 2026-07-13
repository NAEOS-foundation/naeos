package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Line    int    `json:"line,omitempty"`
}

type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Status   string            `json:"status"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
	Summary  string            `json:"summary"`
}

func newValidateCommand() *cobra.Command {
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
			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun)
			if err != nil {
				return err
			}

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			result, err := p.Validate(inputValue)
			if err != nil {
				vr := ValidationResult{
					Valid:  false,
					Status: "invalid",
					Errors: []ValidationError{{
						Code:    "PIPELINE_FAILED",
						Message: err.Error(),
					}},
					Summary: "validation failed",
				}
				return renderValidation(cmd, vr, outputFormat, outputFile)
			}

			vr := ValidationResult{
				Valid:   true,
				Status:  "valid",
				Summary: fmt.Sprintf("valid — project: %s, services: %d", result.NEIR.Project.Name, len(result.NEIR.Services)),
			}

			if inputValue == "" {
				vr.Valid = false
				vr.Status = "invalid"
				vr.Errors = append(vr.Errors, ValidationError{
					Code:    "SPEC_EMPTY",
					Message: "specification is empty",
				})
				vr.Summary = "validation failed — empty spec"
			}

			return renderValidation(cmd, vr, outputFormat, outputFile)
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

func renderValidation(cmd *cobra.Command, vr ValidationResult, format, filePath string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(vr, "", "  ")
		if err != nil {
			return err
		}
		if filePath != "" {
			return writeToFile(filePath, data)
		}
		cmd.OutOrStdout().Write(data)
		cmd.OutOrStdout().Write([]byte("\n"))
		return nil
	case "yaml":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("status: %s\n", vr.Status))
		sb.WriteString(fmt.Sprintf("valid: %t\n", vr.Valid))
		if vr.Summary != "" {
			sb.WriteString(fmt.Sprintf("summary: %s\n", vr.Summary))
		}
		for _, e := range vr.Errors {
			sb.WriteString(fmt.Sprintf("errors:\n  - code: %s\n    message: %s\n", e.Code, e.Message))
		}
		if filePath != "" {
			return writeToFile(filePath, []byte(sb.String()))
		}
		cmd.OutOrStdout().Write([]byte(sb.String()))
		return nil
	default:
		var sb strings.Builder
		if vr.Valid {
			sb.WriteString("✓ ")
		} else {
			sb.WriteString("✗ ")
		}
		sb.WriteString(vr.Summary)
		sb.WriteString("\n")

		for _, e := range vr.Errors {
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", e.Code, e.Message))
		}
		for _, w := range vr.Warnings {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", w))
		}

		if filePath != "" {
			return writeToFile(filePath, []byte(sb.String()))
		}
		cmd.OutOrStdout().Write([]byte(sb.String()))
		return nil
	}
}

func writeToFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func newInspectCommand() *cobra.Command {
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
			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, nil, cliDryRun)
			if err != nil {
				return err
			}

			specInput, err := resolveInput(inputValue)
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

			rendered, err := renderOutput(payload, outputFormat, func() []byte {
				return []byte(fmt.Sprintf("pipeline=%s mode=%s verbose=%t output_dir=%s\nproject=%s artifacts=%d tasks=%d input=%q\n", cfg.Name, cfg.Mode, cfg.Verbose, cfg.OutputDir, projectName, len(result.Artifacts), len(result.Tasks), specInput))
			})
			if err != nil {
				return err
			}

			return writeOrPrint(cmd, rendered, outputFile)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input or file path to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text, json, or yaml")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write the formatted output")
	return cmd
}
