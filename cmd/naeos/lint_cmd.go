package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/lint"
)

func newLintCommand() *cobra.Command {
	var inputFile string
	var fix bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint a specification file",
		Long: `Lint a NAEOS specification file for issues and optionally auto-fix them.

Example:
  naeos lint --input-file spec.yaml
  naeos lint --input-file spec.yaml --fix
  naeos lint --input-file spec.yaml --output json`,
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

			type lintIssue struct {
				Line     int    `json:"line"`
				Severity string `json:"severity"`
				Rule     string `json:"rule"`
				Message  string `json:"message"`
			}

			type lintOutput struct {
				File      string      `json:"file"`
				Issues    []lintIssue `json:"issues"`
				IssueCount int        `json:"issue_count"`
			}

			var issues []lintIssue
			for _, issue := range result.Issues {
				issues = append(issues, lintIssue{
					Line:     issue.Line,
					Severity: string(issue.Severity),
					Rule:     issue.Rule,
					Message:  issue.Message,
				})
			}

			output := lintOutput{
				File:       inputFile,
				Issues:     issues,
				IssueCount: len(result.Issues),
			}

			if outputFormat == "json" {
				data, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal lint result: %w", err)
				}
				cmd.OutOrStdout().Write(data)
				cmd.OutOrStdout().Write([]byte("\n"))
				return nil
			}

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
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json")
	return cmd
}
