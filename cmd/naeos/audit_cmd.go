package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/security"
)

func newAuditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit [file1] [file2] ...",
		Short: "Security audit of generated or source files",
		Long: `Run a security audit on the specified files.

Example:
  naeos audit main.go config.yaml
  naeos audit internal/**/*.go`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			auditor := security.NewAuditor()

			files := make(map[string]string)
			for _, arg := range args {
				data, err := os.ReadFile(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: cannot read %s: %v\n", arg, err)
					continue
				}
				files[arg] = string(data)
			}

			if len(files) == 0 {
				return fmt.Errorf("no files to audit")
			}

			result := auditor.AuditFiles(files)

			for _, f := range result.Finding {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n  File: %s (line %d)\n  %s\n  Remediation: %s\n\n",
					f.Severity, f.Title, f.File, f.Line, f.Description, f.Remediation)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Audit complete: %d findings (%d critical, %d high, %d medium, %d low, %d info)\n",
				result.Summary.Total, result.Summary.Critical, result.Summary.High,
				result.Summary.Medium, result.Summary.Low, result.Summary.Info)
			return nil
		},
	}
	return cmd
}
