package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/docs"
)

func newDocsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate project documentation",
		Long: `Generate API documentation and architecture diagrams.

Example:
  naeos docs api --project my-app
  naeos docs architecture --project my-app -o ./docs`,
	}

	var projectName, outputDir string

	docsAPI := &cobra.Command{
		Use:   "api",
		Short: "Generate API documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := docs.NewGenerator(projectName, nil)
			endpoints := []docs.Endpoint{
				{Method: "GET", Path: "/health", Description: "Health check"},
				{Method: "GET", Path: "/api/v1", Description: "API root"},
				{Method: "GET", Path: "/api/v1/resources", Description: "List resources"},
			}
			content := gen.GenerateAPIDocs(endpoints)
			if outputDir != "" {
				os.MkdirAll(outputDir, 0o755)
				os.WriteFile(filepath.Join(outputDir, "api.md"), []byte(content), 0o600)
				fmt.Fprintf(cmd.OutOrStdout(), "Generated api.md in %s\n", outputDir)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), content)
			}
			return nil
		},
	}

	docsArch := &cobra.Command{
		Use:   "architecture",
		Short: "Generate architecture diagram",
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := docs.NewGenerator(projectName, nil)
			content := gen.GenerateArchitectureDiagram(
				[]string{"api", "worker"},
				[]string{"core", "auth", "data"},
			)
			if outputDir != "" {
				os.MkdirAll(outputDir, 0o755)
				os.WriteFile(filepath.Join(outputDir, "architecture.md"), []byte(content), 0o600)
				fmt.Fprintf(cmd.OutOrStdout(), "Generated architecture.md in %s\n", outputDir)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), content)
			}
			return nil
		},
	}

	cmd.AddCommand(docsAPI)
	cmd.AddCommand(docsArch)
	cmd.PersistentFlags().StringVarP(&projectName, "project", "p", "project", "project name")
	cmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "output directory")
	return cmd
}
