package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/testrunner"
)

func newTestCommand() *cobra.Command {
	var workingDir string
	var languages []string
	var verbose bool

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
  naeos test --dir ./my-project --verbose`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := testrunner.TestConfig{
				WorkingDir: workingDir,
				Languages:  languages,
				Verbose:    verbose,
			}

			runner := testrunner.NewRunner(config)
			results, err := runner.RunAll()
			if err != nil {
				return fmt.Errorf("test run failed: %w", err)
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

	return cmd
}
