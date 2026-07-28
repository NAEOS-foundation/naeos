package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newCLIDocsGenCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:    "docsgen",
		Short:  "Generate CLI reference documentation",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return fmt.Errorf("create output dir: %w", err)
			}

			root := cmd.Root()
			root.DisableAutoGenTag = true

			if err := doc.GenMarkdownTree(root, outputDir); err != nil {
				return fmt.Errorf("generate docs: %w", err)
			}

			fmt.Printf("Generated CLI docs in %s/\n", outputDir)

			entries, err := os.ReadDir(outputDir)
			if err != nil {
				return nil
			}
			for _, e := range entries {
				if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
					fmt.Printf("  - %s\n", filepath.Join(outputDir, e.Name()))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "docs/cli", "output directory for generated docs")
	return cmd
}
