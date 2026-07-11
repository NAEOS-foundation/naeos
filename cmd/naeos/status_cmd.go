package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cfgpkg "github.com/NAEOS-foundation/naeos/pkg/config"
)

func newStatusCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current pipeline and project status",
		Long: `Display the current status of the NAEOS project and pipeline configuration.

Example:
  naeos status
  naeos status --config config.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			fileCfg, err := cfgpkg.LoadFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "NAEOS Status\n")
			fmt.Fprintf(out, "%s\n", "================================")
			fmt.Fprintf(out, "Config:        %s\n", resolved)
			fmt.Fprintf(out, "Pipeline:      %s\n", fileCfg.Pipeline.Name)
			fmt.Fprintf(out, "Mode:          %s\n", fileCfg.Pipeline.Mode)
			fmt.Fprintf(out, "Output Dir:    %s\n", fileCfg.Pipeline.OutputDir)
			fmt.Fprintf(out, "Languages:     %s\n", joinStrings(fileCfg.Pipeline.Language))
			fmt.Fprintf(out, "Verbose:       %t\n", fileCfg.Pipeline.Verbose)
			fmt.Fprintf(out, "Checked At:    %s\n", time.Now().Format(time.RFC3339))
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	return cmd
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	return strings.Join(ss, ", ")
}
