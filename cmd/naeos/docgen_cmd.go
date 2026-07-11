package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/docgen"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func newDocsGenCommand() *cobra.Command {
	var inputFile, input, outputFormat, outputFile string

	cmd := &cobra.Command{
		Use:   "docgen",
		Short: "Generate documentation from specification",
		Long: `Auto-generate API docs, module docs, and architecture docs from specs.

Example:
  naeos docgen --input-file spec.yaml
  naeos docgen --input-file spec.yaml --output api
  naeos docgen --input-file spec.yaml --output modules
  naeos docgen --input-file spec.yaml --output architecture`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			p := parser.NewParser()
			doc, err := p.Parse(inputValue)
			if err != nil {
				return fmt.Errorf("failed to parse specification: %w", err)
			}

			gen := docgen.NewDocGenerator()
			var output string

			switch outputFormat {
			case "api":
				output = gen.GenerateAPIDoc(doc)
			case "modules":
				output = gen.GenerateModuleDocs(doc)
			default:
				output = gen.GenerateFromSpec(doc)
			}

			return writeOrPrint(cmd, []byte(output), outputFile)
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "specification input")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "full", "output type: full, api, modules")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write output")

	return cmd
}
