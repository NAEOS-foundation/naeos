package main

import (
	"fmt"

	"github.com/spf13/cobra"

	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func newContextCommand() *cobra.Command {
	var inputFile, input, outputFormat, outputFile string

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Generate AI context bundles from specifications",
		Long: `Generate context bundles optimized for LLM consumption.
Produces structured markdown or plain text summaries of your project.

Example:
  naeos context --input-file spec.yaml
  naeos context --input 'project: myapp' --output json
  naeos context --input-file spec.yaml --output markdown`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			p := parser.NewParser(".")
			doc, err := p.Parse(inputValue)
			if err != nil {
				return fmt.Errorf("failed to parse specification: %w", err)
			}

			gen := contextbundle.NewGenerator(nil)
			bundle := gen.GenerateFromSpec(doc)

			rendered, err := renderOutput(bundle, outputFormat, func() []byte {
				return []byte(bundle.ToMarkdown())
			})
			if err != nil {
				return err
			}

			return writeOrPrint(cmd, rendered, outputFile)
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "markdown", "output format: markdown, plain, json, or yaml")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write the output")

	return cmd
}
