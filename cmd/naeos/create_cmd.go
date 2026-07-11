package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/create"
)

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Interactive project creation wizard",
		Long: `Launch an interactive wizard to create a new NAEOS project.

Example:
  naeos create`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			wizard := create.NewWizard()
			cfg, err := wizard.Run()
			if err != nil {
				return err
			}

			specContent := cfg.ToSpec()

			fmt.Fprintf(cmd.OutOrStdout(), "\nGenerated specification:\n%s\n", specContent)

			specFile := filepath.Join(cfg.OutputDir, "spec.yaml")
			if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(specFile, []byte(specContent), 0o600); err != nil {
				return err
			}

			configContent := fmt.Sprintf("pipeline:\n  name: %s\n  mode: development\n  output_dir: %s\n", cfg.Name, cfg.OutputDir)
			configFile := filepath.Join(cfg.OutputDir, "config.yaml")
			if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created project %s in %s/\n", cfg.Name, cfg.OutputDir)
			fmt.Fprintf(cmd.OutOrStdout(), "  spec:    %s\n", specFile)
			fmt.Fprintf(cmd.OutOrStdout(), "  config:  %s\n", configFile)
			return nil
		},
	}
	return cmd
}
