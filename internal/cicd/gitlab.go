package cicd

import (
	"fmt"
	"strings"
)

type GitLabCIGenerator struct{}

func (g *GitLabCIGenerator) Name() string {
	return "GitLab CI"
}

func (g *GitLabCIGenerator) Generate(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("stages:\n")
	sb.WriteString("  - build\n")
	sb.WriteString("  - test\n")
	sb.WriteString("  - deploy\n\n")

	// Default image
	if len(config.Languages) > 0 {
		if img := langImage(config.Languages[0]); img != "" {
			fmt.Fprintf(&sb, "image: %s\n\n", img)
		}
	}

	// Build job
	sb.WriteString("build:\n")
	sb.WriteString("  stage: build\n")
	sb.WriteString("  script:\n")
	for _, lang := range config.Languages {
		if cmd := buildCommand(lang); cmd != "" {
			for _, line := range strings.Split(cmd, " && ") {
				fmt.Fprintf(&sb, "    - %s\n", line)
			}
		}
	}
	sb.WriteString("\n")

	// Test job
	sb.WriteString("test:\n")
	sb.WriteString("  stage: test\n")
	sb.WriteString("  script:\n")
	for _, lang := range config.Languages {
		if cmd := testCommand(lang); cmd != "" {
			fmt.Fprintf(&sb, "    - %s\n", cmd)
		}
	}
	sb.WriteString("\n")

	// Deploy job
	sb.WriteString("deploy:\n")
	sb.WriteString("  stage: deploy\n")
	sb.WriteString("  only:\n")
	sb.WriteString("    - main\n")
	sb.WriteString("  script:\n")
	sb.WriteString("    - echo 'Deploying...'\n")

	// Custom steps
	for _, step := range config.Steps {
		fmt.Fprintf(&sb, "    - %s\n", step.Command)
	}
	sb.WriteString("\n")

	return sb.String(), nil
}
