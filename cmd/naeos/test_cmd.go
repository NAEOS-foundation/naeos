package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/testrunner"
)

func newTestCommand() *cobra.Command {
	var workingDir string
	var languages []string
	var verbose bool
	var parallel bool
	var timeout int
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests for generated code",
		Long: `Run tests across all detected or specified languages.

Automatically detects project languages and runs appropriate test commands:
  - Go: go test -v ./...
  - TypeScript/Node: npm test / pnpm test
  - Python: python -m pytest -v
  - Java: mvn test / ./gradlew test
  - Rust: cargo test --verbose

Example:
  naeos test
  naeos test --language go --language typescript
  naeos test --dir ./my-project --verbose
  naeos test --parallel --timeout 30
  naeos test --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := testrunner.TestConfig{
				WorkingDir: workingDir,
				Languages:  languages,
				Verbose:    verbose,
				Parallel:   parallel,
				Timeout:    timeout,
			}

			runner := testrunner.NewRunner(config)
			results, err := runner.RunAll()
			if err != nil {
				return fmt.Errorf("test run failed: %w", err)
			}

			if outputFormat == "json" {
				return renderTestResultsJSON(cmd, results)
			}

			output := testrunner.FormatResults(results)
			fmt.Fprint(cmd.OutOrStdout(), output)

			for _, r := range results {
				if !r.Passed {
					return fmt.Errorf("tests failed for %s", r.Language)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&workingDir, "dir", ".", "working directory for tests")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language (go, typescript, python, java, rust)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose test output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "run tests for different languages in parallel")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "test timeout in seconds (0 = no timeout)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json")

	return cmd
}

func renderTestResultsJSON(cmd *cobra.Command, results []testrunner.TestResult) error {
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}
	status := "passed"
	if failed > 0 {
		status = "failed"
	}

	report := map[string]any{
		"status":  status,
		"passed":  passed,
		"failed":  failed,
		"total":   len(results),
		"results": results,
	}
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	cmd.OutOrStdout().Write(data)
	cmd.OutOrStdout().Write([]byte("\n"))
	return nil
}
