package cicd

import (
	"fmt"
	"strings"
)

type JenkinsGenerator struct{}

func (g *JenkinsGenerator) Name() string {
	return "Jenkins"
}

func (g *JenkinsGenerator) Generate(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("pipeline {\n")
	sb.WriteString("    agent any\n\n")

	// Environment
	if len(config.Secrets) > 0 {
		sb.WriteString("    environment {\n")
		for _, secret := range config.Secrets {
			fmt.Fprintf(&sb, "        %s = credentials('%s')\n", strings.ToUpper(secret), secret)
		}
		sb.WriteString("    }\n\n")
	}

	// Stages
	sb.WriteString("    stages {\n")

	// Build stage
	sb.WriteString("        stage('Build') {\n")
	sb.WriteString("            steps {\n")
	for _, lang := range config.Languages {
		if cmd := buildCommand(lang); cmd != "" {
			for _, line := range strings.Split(cmd, " && ") {
				fmt.Fprintf(&sb, "                sh '%s'\n", line)
			}
		}
	}
	sb.WriteString("            }\n")
	sb.WriteString("        }\n\n")

	// Test stage
	sb.WriteString("        stage('Test') {\n")
	sb.WriteString("            steps {\n")
	for _, lang := range config.Languages {
		if cmd := testCommand(lang); cmd != "" {
			fmt.Fprintf(&sb, "                sh '%s'\n", cmd)
		}
	}
	sb.WriteString("            }\n")
	sb.WriteString("        }\n\n")

	// Deploy stage
	sb.WriteString("        stage('Deploy') {\n")
	sb.WriteString("            when {\n")
	sb.WriteString("                branch 'main'\n")
	sb.WriteString("            }\n")
	sb.WriteString("            steps {\n")
	sb.WriteString("                echo 'Deploying...'\n")
	for _, step := range config.Steps {
		fmt.Fprintf(&sb, "                sh '%s'\n", step.Command)
	}
	sb.WriteString("            }\n")
	sb.WriteString("        }\n")

	sb.WriteString("    }\n\n")

	// Post
	sb.WriteString("    post {\n")
	sb.WriteString("        always {\n")
	sb.WriteString("            cleanWs()\n")
	sb.WriteString("        }\n")
	sb.WriteString("        success {\n")
	sb.WriteString("            echo 'Pipeline succeeded!'\n")
	sb.WriteString("        }\n")
	sb.WriteString("        failure {\n")
	sb.WriteString("            echo 'Pipeline failed!'\n")
	sb.WriteString("        }\n")
	sb.WriteString("    }\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}
