package cicd

import (
	"fmt"
	"strings"
)

type GitHubActionsGenerator struct{}

func (g *GitHubActionsGenerator) Name() string {
	return "GitHub Actions"
}

func (g *GitHubActionsGenerator) Generate(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("name: CI/CD Pipeline\n\n")

	// Triggers
	sb.WriteString("on:\n")
	if config.Trigger.OnPush {
		sb.WriteString("  push:\n    branches: [main, develop]\n")
	}
	if config.Trigger.OnPR {
		sb.WriteString("  pull_request:\n    branches: [main]\n")
	}
	if config.Trigger.OnRelease {
		sb.WriteString("  release:\n    types: [created]\n")
	}
	if config.Trigger.Schedule != "" {
		fmt.Fprintf(&sb, "  schedule:\n    - cron: '%s'\n", config.Trigger.Schedule)
	}
	sb.WriteString("\n")

	// Jobs
	sb.WriteString("jobs:\n")
	sb.WriteString("  build:\n")
	sb.WriteString("    runs-on: ubuntu-latest\n\n")

	// Steps
	sb.WriteString("    steps:\n")
	sb.WriteString("      - uses: actions/checkout@v4\n\n")

	for _, lang := range config.Languages {
		switch lang {
		case "go":
			sb.WriteString("      - name: Set up Go\n")
			sb.WriteString("        uses: actions/setup-go@v5\n")
			sb.WriteString("        with:\n")
			sb.WriteString("          go-version: '1.22'\n\n")
		case "node", "typescript":
			sb.WriteString("      - name: Set up Node.js\n")
			sb.WriteString("        uses: actions/setup-node@v4\n")
			sb.WriteString("        with:\n")
			sb.WriteString("          node-version: '20'\n\n")
		case "python":
			sb.WriteString("      - name: Set up Python\n")
			sb.WriteString("        uses: actions/setup-python@v5\n")
			sb.WriteString("        with:\n")
			sb.WriteString("          python-version: '3.12'\n\n")
		case "java":
			sb.WriteString("      - name: Set up Java\n")
			sb.WriteString("        uses: actions/setup-java@v4\n")
			sb.WriteString("        with:\n")
			sb.WriteString("          java-version: '21'\n")
			sb.WriteString("          distribution: 'temurin'\n\n")
		case "rust":
			sb.WriteString("      - name: Set up Rust\n")
			sb.WriteString("        uses: dtolnay/rust-toolchain@stable\n\n")
		}
		if cmd := buildCommand(lang); cmd != "" {
			fmt.Fprintf(&sb, "      - name: Build\n        run: %s\n\n", cmd)
		}
		if cmd := testCommand(lang); cmd != "" {
			fmt.Fprintf(&sb, "      - name: Test\n        run: %s\n\n", cmd)
		}
	}

	// Custom steps
	for _, step := range config.Steps {
		fmt.Fprintf(&sb, "      - name: %s\n", step.Name)
		fmt.Fprintf(&sb, "        run: %s\n\n", step.Command)
	}

	return sb.String(), nil
}
