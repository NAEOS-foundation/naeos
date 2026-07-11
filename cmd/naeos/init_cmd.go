package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a default NAEOS config file",
		Long: `Generate a default NAEOS configuration file.

Example:
  naeos init
  naeos init -o my-config.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			content := strings.Join([]string{
				"pipeline:",
				"  name: naeos-dev",
				"  mode: development",
				"  verbose: true",
				"  output_dir: ./out",
			}, "\n") + "\n"

			if err := os.WriteFile(output, []byte(content), 0o600); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			_, err := fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", output)
			return err
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "config.example.yaml", "path for the generated config file")
	return cmd
}
