package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/hcl"
	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	var inputFile string
	var outputPath string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import specifications from HCL format to NAEOS YAML/JSON",
		Long: `Import a specification written in HCL format and convert it to
the NAEOS YAML or JSON format for use with the pipeline.

Supported HCL blocks:
  project "name" { version = "1.0.0" }
  service "name" { port = 8080; type = "backend" }
  infra "name" { engine = "docker" }

Example:
  naeos import --input spec.hcl
  naeos import --input spec.hcl --output spec.yaml --format yaml
  naeos import --input spec.hcl --format json --output-file result.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := hcl.ParseFile(inputFile)
			if err != nil {
				return fmt.Errorf("parse HCL: %w", err)
			}

			var output []byte
			switch outputFormat {
			case "json":
				output, err = json.MarshalIndent(spec, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal JSON: %w", err)
				}
				output = append(output, '\n')
			default:
				output, err = hcl.ToYAML(spec)
				if err != nil {
					return fmt.Errorf("convert to YAML: %w", err)
				}
				output = append(output, '\n')
			}

			if outputPath != "" {
				if err := os.WriteFile(outputPath, output, 0o644); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
				cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Imported %s → %s (%s format)\n", inputFile, outputPath, outputFormat)))
			} else {
				cmd.OutOrStdout().Write(output)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "path to HCL input file (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "path to output file")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "yaml", "output format: yaml, json")
	_ = cmd.MarkFlagRequired("input")

	return cmd
}
