package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/lint"
)

func newLintCommand() *cobra.Command {
	var inputFile string
	var fix bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint a specification file",
		Long: `Lint a NAEOS specification file for issues and optionally auto-fix them.

Example:
  naeos lint --input-file spec.yaml
  naeos lint --input-file spec.yaml --fix`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" {
				return fmt.Errorf("missing required --input-file")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			l := lint.NewLinter()
			result := l.Lint(inputFile, string(data))

			if len(result.Issues) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: no issues found\n", inputFile)
				return nil
			}

			for _, issue := range result.Issues {
				lineStr := ""
				if issue.Line > 0 {
					lineStr = fmt.Sprintf(":%d", issue.Line)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s [%s] %s: %s\n",
					inputFile, lineStr, issue.Severity, issue.Rule, issue.Message)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\n%d issue(s) found\n", len(result.Issues))

			if fix {
				fixed := lint.Fix(string(data))
				if err := os.WriteFile(inputFile, []byte(fixed), 0o600); err != nil {
					return fmt.Errorf("write fixed file: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Applied fixes to %s\n", inputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file to lint")
	cmd.Flags().BoolVar(&fix, "fix", false, "automatically fix issues where possible")
	return cmd
}
