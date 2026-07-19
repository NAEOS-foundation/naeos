package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/internal/cicd"
)

var (
	cicdPlatform  string
	cicdOutput    string
	cicdProject   string
	cicdLanguages string
	cicdInputFile string
)

func newCICDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cicd",
		Short: "Generate CI/CD pipeline configuration",
		Long:  `Generate CI/CD pipeline configuration for GitHub Actions, GitLab CI, or Jenkins.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			platform := cicd.CICDPlatform(cicdPlatform)
			gen, err := cicd.GetGenerator(platform)
			if err != nil {
				return err
			}

			config := &cicd.PipelineConfig{
				Project:   cicdProject,
				Platform:  platform,
				Languages: strings.Split(cicdLanguages, ","),
				Trigger: cicd.TriggerConfig{
					OnPush: true,
					OnPR:   true,
				},
			}

			if cicdInputFile != "" {
				data, err := os.ReadFile(cicdInputFile)
				if err != nil {
					return fmt.Errorf("read input file: %w", err)
				}
				ext := strings.TrimPrefix(cicdInputFile, ".")
				if strings.HasSuffix(ext, "yaml") || strings.HasSuffix(ext, "yml") {
					if err := yaml.Unmarshal(data, config); err != nil {
						return fmt.Errorf("parse YAML input: %w", err)
					}
				} else {
					if err := json.Unmarshal(data, config); err != nil {
						return fmt.Errorf("parse JSON input: %w", err)
					}
				}
			}

			output, err := gen.Generate(config)
			if err != nil {
				return err
			}

			if cicdOutput != "" {
				return os.WriteFile(cicdOutput, []byte(output), 0o600)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&cicdPlatform, "platform", "p", "github", "CI/CD platform (github, gitlab, jenkins)")
	cmd.Flags().StringVarP(&cicdOutput, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&cicdProject, "project", "myapp", "Project name for pipeline config")
	cmd.Flags().StringVar(&cicdLanguages, "languages", "go", "Comma-separated list of languages (go, python, node, etc.)")
	cmd.Flags().StringVar(&cicdInputFile, "input-file", "", "Path to YAML/JSON spec file to override config")

	return cmd
}
