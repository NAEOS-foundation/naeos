package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var initTemplates = map[string]string{
	"basic": `pipeline:
  name: my-project
  mode: development
  verbose: true
  output_dir: ./out
  language:
    - go
`,
	"microservices": `pipeline:
  name: my-microservices
  mode: development
  verbose: true
  output_dir: ./out
  language:
    - go
    - typescript
services:
  - name: api-gateway
    port: 8080
    type: backend
  - name: user-service
    port: 8081
    type: backend
  - name: order-service
    port: 8082
    type: backend
  - name: notification-service
    port: 8083
    type: backend
  - name: web-frontend
    port: 3000
    type: frontend
`,
	"rest-api": `pipeline:
  name: my-rest-api
  mode: development
  verbose: true
  output_dir: ./out
  language:
    - go
services:
  - name: api
    port: 8080
    type: backend
`,
	"fullstack": `pipeline:
  name: my-fullstack
  mode: development
  verbose: true
  output_dir: ./out
  language:
    - go
    - typescript
    - python
services:
  - name: backend
    port: 8080
    type: backend
  - name: frontend
    port: 3000
    type: frontend
  - name: worker
    port: 9090
    type: job
`,
	"kubernetes": `pipeline:
  name: my-k8s-app
  mode: production
  verbose: false
  output_dir: ./deploy
  language:
    - go
infra:
  engine: kubernetes
services:
  - name: api
    port: 8080
    type: backend
    replicas: 3
`,
	"hcl": `project "my-project" {
  version     = "1.0.0"
  description = "A new NAEOS project"
}

service "api" {
  image = "my-project-api"
  port  = 8080
  type  = "backend"
}

service "web" {
  image = "my-project-web"
  port  = 3000
  type  = "frontend"
}

infra "infra" {
  engine = "docker"
}
`,
}

func newInitCommand() *cobra.Command {
	var output string
	var template string
	var projectName string
	var listTemplates bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new NAEOS project or generate config",
		Long: `Initialize a new NAEOS project with a configuration file.

Available templates:
  basic          — Minimal config with Go (default)
  microservices  — Multi-service microservices architecture
  rest-api       — Single REST API service
  fullstack      — Fullstack with backend + frontend + worker
  kubernetes     — Production-ready Kubernetes deployment
  hcl            — HCL format specification

Example:
  naeos init
  naeos init --template microservices
  naeos init --template rest-api --name my-api
  naeos init --list-templates`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listTemplates {
				cmd.OutOrStdout().Write([]byte("Available templates:\n\n"))
				for name := range initTemplates {
					cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-15s %s\n", name, templateDescription(name))))
				}
				return nil
			}

			content, ok := initTemplates[template]
			if !ok {
				return fmt.Errorf("unknown template %q. Use --list-templates to see available templates", template)
			}

			if projectName != "" {
				content = strings.Replace(content, "my-project", projectName, -1)
				content = strings.Replace(content, "my-microservices", projectName, -1)
				content = strings.Replace(content, "my-rest-api", projectName, -1)
				content = strings.Replace(content, "my-fullstack", projectName, -1)
				content = strings.Replace(content, "my-k8s-app", projectName, -1)
			}

			if err := os.WriteFile(output, []byte(content), 0o600); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			ext := "yaml"
			if template == "hcl" {
				ext = "hcl"
			}

			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Created %s (template: %s)\n", output, template)))
			if ext == "yaml" {
				steps := "\nNext steps:\n"
				steps += "  1. Edit " + output + " to customize your project\n"
				steps += "  2. Run 'naeos validate --input spec.yaml' to validate\n"
				steps += "  3. Run 'naeos run' to generate artifacts\n"
				cmd.OutOrStdout().Write([]byte(steps))
			} else {
				steps := "\nNext steps:\n"
				steps += "  1. Edit " + output + " to customize your project\n"
				steps += "  2. Run 'naeos import --input " + output + "' to convert to YAML\n"
				cmd.OutOrStdout().Write([]byte(steps))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "naeos.yaml", "path for the generated config file")
	cmd.Flags().StringVarP(&template, "template", "t", "basic", "template to use (basic, microservices, rest-api, fullstack, kubernetes, hcl)")
	cmd.Flags().StringVarP(&projectName, "name", "n", "", "project name (replaces default in template)")
	cmd.Flags().BoolVar(&listTemplates, "list-templates", false, "list all available templates")
	return cmd
}

func templateDescription(name string) string {
	descriptions := map[string]string{
		"basic":          "Minimal config with Go",
		"microservices":  "Multi-service microservices architecture",
		"rest-api":       "Single REST API service",
		"fullstack":      "Fullstack: backend + frontend + worker",
		"kubernetes":     "Production-ready Kubernetes deployment",
		"hcl":            "HCL format specification",
	}
	if d, ok := descriptions[name]; ok {
		return d
	}
	return ""
}
