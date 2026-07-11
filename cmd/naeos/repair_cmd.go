package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newRepairCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "repair",
		Short: "Repair the NAEOS output directory",
		Long: `Repair and recreate the output directory structure.

Example:
  naeos repair --config config.yaml
  naeos repair`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			cfg, err := pipeline.ConfigFromFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if cfg.OutputDir == "" {
				return fmt.Errorf("output_dir is not configured")
			}
			if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
				return fmt.Errorf("create output dir: %w", err)
			}

			readmePath := filepath.Join(cfg.OutputDir, "README.md")
			if _, err := os.Stat(readmePath); err != nil {
				if err := os.WriteFile(readmePath, []byte("# Repaired output\n"), 0o600); err != nil {
					return fmt.Errorf("write repair readme: %w", err)
				}
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "repaired %s\n", cfg.OutputDir)
			return err
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	return cmd
}
