package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newValidateCommand() *cobra.Command {
	var configPath, input, inputFile, outputFormat, outputFile string
	var languages []string

	cmd := &cobra.Command{
		Use:     "validate",
		Aliases: []string{"v"},
		Short:   "Validate a specification using the NAEOS pipeline",
		Long: `Validate a specification file through the NAEOS pipeline without generating artifacts.

Example:
  naeos validate --config config.yaml --input spec.yaml
  naeos v --config config.yaml --input-file spec.yaml --output json`,
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
				return fmt.Errorf("pipeline validate failed: %w", err)
			}

			payload := map[string]any{
				"pipeline":   cfg.Name,
				"mode":       cfg.Mode,
				"verbose":    cfg.Verbose,
				"output_dir": cfg.OutputDir,
				"status":     "valid",
				"project":    result.NEIR.Project,
				"source_len": len(result.Source),
			}

			if len(languages) > 0 {
				payload["languages"] = languages
			}

			rendered, err := renderOutput(payload, outputFormat, func() []byte {
				return []byte(fmt.Sprintf("config=%s mode=%s verbose=%t output_dir=%s\nstatus=valid project=%v source_len=%d\n",
					cfg.Name, cfg.Mode, cfg.Verbose, cfg.OutputDir, result.NEIR.Project, len(result.Source)))
			})
			if err != nil {
				return err
			}

			return writeOrPrint(cmd, rendered, outputFile)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text, json, or yaml")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write the formatted output")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language for code generation (go, typescript, python, java, rust)")
	return cmd
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
