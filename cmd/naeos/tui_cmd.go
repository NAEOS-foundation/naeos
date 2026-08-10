package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var tuiLanguages = []string{"Go", "TypeScript", "Python", "Java", "Rust"}
var tuiCloudProviders = []string{"AWS", "GCP", "Azure", "None"}

func newTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Terminal user interface tools",
		Long:  "Terminal user interface tools including the project wizard and live dashboard.",
	}
	cmd.AddCommand(newTUIWizardCommand())
	cmd.AddCommand(newTUIDashboardCommand())
	return cmd
}

func newTUIWizardCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "wizard",
		Short: "Interactive project wizard",
		Long:  "Walk through project configuration step by step and generate a spec.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI()
		},
	}
}

func runTUI() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        NAEOS Project Wizard              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// 1. Project name
	fmt.Print("Project name: ")
	scanner.Scan()
	projectName := strings.TrimSpace(scanner.Text())
	if projectName == "" {
		return fmt.Errorf("project name is required")
	}

	// 2. Language
	fmt.Println()
	fmt.Println("Select language:")
	for i, lang := range tuiLanguages {
		fmt.Printf("  %d) %s\n", i+1, lang)
	}
	fmt.Print("Choice [1]: ")
	scanner.Scan()
	langChoice := strings.TrimSpace(scanner.Text())
	if langChoice == "" {
		langChoice = "1"
	}
	langIdx := 0
	_, _ = fmt.Sscanf(langChoice, "%d", &langIdx)
	if langIdx < 1 || langIdx > len(tuiLanguages) {
		return fmt.Errorf("invalid language choice: %s", langChoice)
	}
	language := tuiLanguages[langIdx-1]

	// 3. Modules
	fmt.Print("\nModules (comma-separated, e.g. api,auth,db): ")
	scanner.Scan()
	modulesRaw := strings.TrimSpace(scanner.Text())
	var modules []string
	if modulesRaw != "" {
		for _, m := range strings.Split(modulesRaw, ",") {
			m = strings.TrimSpace(m)
			if m != "" {
				modules = append(modules, m)
			}
		}
	}

	// 4. Services
	fmt.Println()
	var services []string
	for {
		fmt.Print("Service name (empty to finish): ")
		scanner.Scan()
		svcName := strings.TrimSpace(scanner.Text())
		if svcName == "" {
			break
		}
		fmt.Print("  Port: ")
		scanner.Scan()
		port := strings.TrimSpace(scanner.Text())
		if port == "" {
			port = "8080"
		}
		services = append(services, fmt.Sprintf("  - name: %s\n    port: %s\n    type: backend", svcName, port))
	}

	// 5. Cloud provider
	fmt.Println()
	fmt.Println("Cloud provider:")
	for i, p := range tuiCloudProviders {
		fmt.Printf("  %d) %s\n", i+1, p)
	}
	fmt.Print("Choice [4]: ")
	scanner.Scan()
	cloudChoice := strings.TrimSpace(scanner.Text())
	if cloudChoice == "" {
		cloudChoice = "4"
	}
	cloudIdx := 0
	_, _ = fmt.Sscanf(cloudChoice, "%d", &cloudIdx)
	if cloudIdx < 1 || cloudIdx > len(tuiCloudProviders) {
		return fmt.Errorf("invalid cloud choice: %s", cloudChoice)
	}
	cloudProvider := tuiCloudProviders[cloudIdx-1]

	// Build spec YAML
	var sb strings.Builder
	sb.WriteString("pipeline:\n")
	fmt.Fprintf(&sb, "  name: %s\n", projectName)
	sb.WriteString("  mode: development\n")
	sb.WriteString("  verbose: true\n")
	sb.WriteString("  output_dir: ./out\n")
	fmt.Fprintf(&sb, "  language:\n    - %s\n", strings.ToLower(language))

	if len(modules) > 0 {
		sb.WriteString("modules:\n")
		for _, m := range modules {
			fmt.Fprintf(&sb, "  - %s\n", m)
		}
	}

	if len(services) > 0 {
		sb.WriteString("services:\n")
		for _, s := range services {
			sb.WriteString(s + "\n")
		}
	}

	if cloudProvider != "None" {
		fmt.Fprintf(&sb, "cloud:\n  provider: %s\n", strings.ToLower(cloudProvider))
	}

	specYAML := sb.String()

	// 6. Action
	fmt.Println()
	fmt.Println("Spec generated. What would you like to do?")
	fmt.Println("  1) Validate")
	fmt.Println("  2) Run pipeline")
	fmt.Println("  3) Save to file")
	fmt.Print("Choice [3]: ")
	scanner.Scan()
	actionChoice := strings.TrimSpace(scanner.Text())
	if actionChoice == "" {
		actionChoice = "3"
	}

	switch actionChoice {
	case "1":
		fmt.Println("\n--- Generated Spec ---")
		fmt.Print(specYAML)
		fmt.Println("\n--- Validation ---")
		tmpFile := ".naeos_tui_spec.yaml"
		if err := os.WriteFile(tmpFile, []byte(specYAML), 0o600); err != nil {
			return fmt.Errorf("write temp spec: %w", err)
		}
		defer os.Remove(tmpFile)
		fmt.Printf("Wrote temp spec to %s\n", tmpFile)
		fmt.Println("Run: naeos validate --input " + tmpFile)
	case "2":
		fmt.Println("\n--- Generated Spec ---")
		fmt.Print(specYAML)
		tmpFile := ".naeos_tui_spec.yaml"
		if err := os.WriteFile(tmpFile, []byte(specYAML), 0o600); err != nil {
			return fmt.Errorf("write temp spec: %w", err)
		}
		defer os.Remove(tmpFile)
		fmt.Printf("Run: naeos run --input %s\n", tmpFile)
	case "3":
		outFile := projectName + ".yaml"
		if err := os.WriteFile(outFile, []byte(specYAML), 0o600); err != nil {
			return fmt.Errorf("write spec: %w", err)
		}
		fmt.Printf("Saved to %s\n", outFile)
	default:
		return fmt.Errorf("invalid choice: %s", actionChoice)
	}

	return nil
}
