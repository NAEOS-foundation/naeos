package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: naeos <init|run|inspect|validate|doctor|repair|scaffold>")
	}

	subcommand := args[0]
	switch subcommand {
	case "init":
		return runInit(args[1:])
	case "run":
		return runPipeline(args[1:])
	case "inspect":
		return runInspect(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "repair":
		return runRepair(args[1:])
	case "scaffold":
		return runScaffold(args[1:])
	case "export":
		return runExport(args[1:])
	case "preview":
		return runPreview(args[1:])
	default:
		return fmt.Errorf("unknown command %q", subcommand)
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	output := fs.String("output", "config.example.yaml", "path for the generated config file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	content := strings.Join([]string{
		"pipeline:",
		"  name: naeos-dev",
		"  mode: development",
		"  verbose: true",
		"  output_dir: ./out",
	}, "\n") + "\n"

	if err := os.WriteFile(*output, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("created %s\n", *output)
	return nil
}

func runPipeline(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	input := fs.String("input", "", "specification input or file path to process")
	outputFormat := fs.String("output", "text", "output format: text, json, or yaml")
	outputFile := fs.String("output-file", "", "optional file path to write the formatted output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}
	if *input == "" {
		return fmt.Errorf("missing required --input")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specInput, err := resolveInput(*input)
	if err != nil {
		return fmt.Errorf("resolve input: %w", err)
	}

	p := pipeline.New(cfg)
	result, err := p.Run(specInput)
	if err != nil {
		return fmt.Errorf("pipeline run failed: %w", err)
	}

	payload := map[string]any{
		"pipeline":   cfg.Name,
		"mode":       cfg.Mode,
		"verbose":    cfg.Verbose,
		"output_dir": cfg.OutputDir,
		"artifacts":  len(result.Artifacts),
		"tasks":      len(result.Tasks),
	}

	var rendered []byte
	switch strings.ToLower(*outputFormat) {
	case "json":
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json output: %w", err)
		}
		rendered = append(data, '\n')
	case "yaml":
		data, err := yaml.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode yaml output: %w", err)
		}
		rendered = data
	default:
		rendered = []byte(fmt.Sprintf("pipeline=%s mode=%s verbose=%t output_dir=%s\nartifacts=%d tasks=%d\n", result.NEIR.Project, cfg.Mode, cfg.Verbose, cfg.OutputDir, len(result.Artifacts), len(result.Tasks)))
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, rendered, 0o644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(rendered); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	input := fs.String("input", "", "specification input or file path to process")
	outputFormat := fs.String("output", "text", "output format: text, json, or yaml")
	outputFile := fs.String("output-file", "", "optional file path to write the formatted output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}
	if *input == "" {
		return fmt.Errorf("missing required --input")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specInput, err := resolveInput(*input)
	if err != nil {
		return fmt.Errorf("resolve input: %w", err)
	}

	p := pipeline.New(cfg)
	result, err := p.Run(specInput)
	if err != nil {
		return fmt.Errorf("inspect run failed: %w", err)
	}

	projectName := ""
	if result.NEIR != nil {
		projectName = fmt.Sprint(result.NEIR.Project)
	}
	payload := map[string]any{
		"pipeline":    cfg.Name,
		"mode":        cfg.Mode,
		"verbose":     cfg.Verbose,
		"output_dir":  cfg.OutputDir,
		"project":     projectName,
		"input":       specInput,
		"artifacts":   len(result.Artifacts),
		"tasks":       len(result.Tasks),
		"source_words": len(strings.Fields(result.Source)),
	}

	var rendered []byte
	switch strings.ToLower(*outputFormat) {
	case "json":
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json output: %w", err)
		}
		rendered = append(data, '\n')
	case "yaml":
		data, err := yaml.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode yaml output: %w", err)
		}
		rendered = data
	default:
		rendered = []byte(fmt.Sprintf("pipeline=%s mode=%s verbose=%t output_dir=%s\nproject=%s artifacts=%d tasks=%d input=%q\n", cfg.Name, cfg.Mode, cfg.Verbose, cfg.OutputDir, projectName, len(result.Artifacts), len(result.Tasks), specInput))
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, rendered, 0o644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(rendered); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	input := fs.String("input", "", "specification input to process")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}
	if *input == "" {
		return fmt.Errorf("missing required --input")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	_ = cfg
	fmt.Printf("config loaded successfully from %s\n", *configPath)
	fmt.Printf("input received: %s\n", *input)
	return nil
}

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	status := map[string]any{
		"config":            *configPath,
		"pipeline_name":     cfg.Name,
		"mode":              cfg.Mode,
		"output_dir":        cfg.OutputDir,
		"output_dir_exists": false,
	}
	if cfg.OutputDir != "" {
		_, err := os.Stat(cfg.OutputDir)
		status["output_dir_exists"] = err == nil
	}

	payload, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("encode doctor output: %w", err)
	}
	fmt.Println(string(payload))
	return nil
}

func resolveInput(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	info, err := os.Stat(input)
	if err != nil {
		if os.IsNotExist(err) {
			return input, nil
		}
		return "", err
	}
	if info.IsDir() {
		return input, nil
	}

	content, err := os.ReadFile(input)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	input := fs.String("input", "", "specification input or file path to process")
	outputDir := fs.String("output-dir", "", "directory to write exported artifacts")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}
	if *input == "" {
		return fmt.Errorf("missing required --input")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if *outputDir == "" {
		*outputDir = cfg.OutputDir
	}
	if *outputDir == "" {
		return fmt.Errorf("missing output directory")
	}

	specInput, err := resolveInput(*input)
	if err != nil {
		return fmt.Errorf("resolve input: %w", err)
	}

	p := pipeline.New(cfg)
	result, err := p.Run(specInput)
	if err != nil {
		return fmt.Errorf("export run failed: %w", err)
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		return fmt.Errorf("create export dir: %w", err)
	}
	for _, artifact := range result.Artifacts {
		artifactPath := filepath.Join(*outputDir, artifact.Path)
		if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
			return fmt.Errorf("create artifact dir: %w", err)
		}
		if err := os.WriteFile(artifactPath, artifact.Content, 0o644); err != nil {
			return fmt.Errorf("write artifact %s: %w", artifact.Path, err)
		}
	}

	fmt.Printf("exported %d artifacts to %s\n", len(result.Artifacts), *outputDir)
	return nil
}

func runPreview(args []string) error {
	fs := flag.NewFlagSet("preview", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	input := fs.String("input", "", "specification input or file path to process")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}
	if *input == "" {
		return fmt.Errorf("missing required --input")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specInput, err := resolveInput(*input)
	if err != nil {
		return fmt.Errorf("resolve input: %w", err)
	}

	p := pipeline.New(cfg)
	result, err := p.Run(specInput)
	if err != nil {
		return fmt.Errorf("preview run failed: %w", err)
	}

	fmt.Printf("preview for %s\n", cfg.Name)
	for _, artifact := range result.Artifacts {
		fmt.Printf("- %s\n", artifact.Path)
	}
	return nil
}

func runRepair(args []string) error {
	fs := flag.NewFlagSet("repair", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.OutputDir == "" {
		return fmt.Errorf("output_dir is not configured")
	}
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	readmePath := cfg.OutputDir + string(os.PathSeparator) + "README.md"
	if _, err := os.Stat(readmePath); err != nil {
		if err := os.WriteFile(readmePath, []byte("# Repaired output\n"), 0o644); err != nil {
			return fmt.Errorf("write repair readme: %w", err)
		}
	}

	fmt.Printf("repaired %s\n", cfg.OutputDir)
	return nil
}

func runScaffold(args []string) error {
	fs := flag.NewFlagSet("scaffold", flag.ContinueOnError)
	name := fs.String("name", "", "project name for the scaffold")
	output := fs.String("output", "", "directory where scaffold files will be created")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("missing required --name")
	}
	if *output == "" {
		*output = *name
	}

	if err := os.MkdirAll(*output, 0o755); err != nil {
		return fmt.Errorf("create scaffold dir: %w", err)
	}

	readme := fmt.Sprintf("# %s\n\nGenerated by NAEOS scaffold.\n\n## Quick start\n\n- Review spec.yaml\n- Run `go mod tidy`\n- Run `go run ./cmd/app`\n- Visit `http://localhost:8080/health`\n- Visit `http://localhost:8080/api/v1`\n- Visit `http://localhost:8080/api/v1/resources`\n", *name)
	spec := fmt.Sprintf("project: %s\nmodules:\n  - name: core\n    path: ./internal/core\nservices:\n  - name: api\n    kind: http\n    port: 8080\n", *name)
	makefile := fmt.Sprintf(".PHONY: help\n\nhelp:\n\t@echo 'Available targets:'\n\t@echo '  make help'\n\t@echo '  make scaffold'\n\nscaffold:\n\t@echo 'Scaffold complete for %s'\n", *name)
	gitignore := "# Build artifacts\n/bin/\n/out/\n*.log\n"
	dockerfile := "FROM golang:1.22-alpine\nWORKDIR /app\nCOPY . .\nRUN go build ./cmd/app\nCMD [\"/app/app\"]\n"
	ciWorkflow := "name: ci\n\non: [push, pull_request]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-go@v5\n        with:\n          go-version: '1.22'\n      - run: go test ./...\n"
	goMod := "module github.com/example/" + *name + "\n\ngo 1.22\n"
	mainGo := "package main\n\nimport (\n\t\"fmt\"\n\t\"log\"\n\t\"net/http\"\n\n\t\"github.com/example/" + *name + "/internal/core\"\n\tcoreconfig \"github.com/example/" + *name + "/internal/core/config\"\n\tcorehttp \"github.com/example/" + *name + "/internal/core/http\"\n\tcoremiddleware \"github.com/example/" + *name + "/internal/core/middleware\"\n)\n\nfunc main() {\n\tcfg := coreconfig.Load(\"config.yaml\")\n\thandler := core.NewHandler(nil)\n\t_ = handler\n\tmux := http.NewServeMux()\n\tmux.HandleFunc(\"/\", func(w http.ResponseWriter, r *http.Request) {\n\t\t_, _ = fmt.Fprintf(w, \"hello from scaffold on port %d\", cfg.Port)\n\t})\n\tmux.HandleFunc(\"/health\", func(w http.ResponseWriter, r *http.Request) {\n\t\t_, _ = fmt.Fprintln(w, \"ok\")\n\t})\n\tmux.HandleFunc(\"/api/v1\", func(w http.ResponseWriter, r *http.Request) {\n\t\t_, _ = fmt.Fprintln(w, \"api v1 ready\")\n\t})\n\t_ = corehttp.Handler{}\n\twrapped := coremiddleware.LoggingMiddleware{}.Wrap(mux)\n\tlog.Printf(\"listening on :%d\", cfg.Port)\n\tif err := http.ListenAndServe(fmt.Sprintf(\":%d\", cfg.Port), wrapped); err != nil {\n\t\tlog.Fatal(err)\n\t}\n}\n"
	coreModuleFiles := map[string]string{
		"internal/core/README.md": "# core\n\nStarter module generated by NAEOS.\n",
		"internal/core/package.go": "package core\n\n// core module placeholder.\n",
		"internal/core/config.yaml": "name: core\nmodule: core\n",
		"internal/core/handler.go": "package core\n\n// Handler is a small starter implementation for the core module.\ntype Handler struct {\n\tservice Service\n}\n\nfunc NewHandler(service Service) *Handler {\n\treturn &Handler{service: service}\n}\n",
		"internal/core/repository.go": "package core\n\n// Repository interface describes the persistence boundary for the core module.\ntype Repository interface {\n\tList() []string\n}\n",
		"internal/core/service.go": "package core\n\n// Service interface describes the application behavior for the core module.\ntype Service interface {\n\tHandle() string\n}\n",
		"internal/core/domain/model.go": "package domain\n\n// Model is a sample domain object for the core module.\ntype Model struct {\n\tName string\n}\n",
		"internal/core/http/handler.go": "package http\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n)\n\n// Handler is a starter HTTP handler for the core module.\ntype Handler struct{}\n\nfunc (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {\n\t_, _ = fmt.Fprintln(w, \"handler for core\")\n}\n",
		"internal/core/http/router.go": "package http\n\n// Router is a starter router stub for the core module.\ntype Router struct{}\n",
		"internal/core/middleware/logging.go": "package middleware\n\nimport \"net/http\"\n\n// LoggingMiddleware wraps the next handler with a simple logging stub.\ntype LoggingMiddleware struct{}\n\nfunc (m LoggingMiddleware) Wrap(next http.Handler) http.Handler {\n\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n\t\tnext.ServeHTTP(w, r)\n\t})\n}\n",
		"internal/core/config/config.go": "package config\n\n// Config is a starter configuration container for the core module.\ntype Config struct {\n\tPort int\n}\n",
		"internal/core/config/load.go": "package config\n\nimport (\n\t\"encoding/json\"\n\t\"fmt\"\n\t\"os\"\n\n\t\"gopkg.in/yaml.v3\"\n)\n\n// Load returns a starter configuration for the core module.\nfunc Load(path string) Config {\n\tif path == \"\" {\n\t\tpath = \"config.yaml\"\n\t}\n\tdata, err := os.ReadFile(path)\n\tif err != nil {\n\t\treturn Config{Port: 8080}\n\t}\n\tvar cfg Config\n\tswitch ext := fmt.Sprintf(\"%s\", path); ext {\n\tcase \"config.json\":\n\t\t_ = json.Unmarshal(data, &cfg)\n\tdefault:\n\t\t_ = yaml.Unmarshal(data, &cfg)\n\t}\n\tif cfg.Port == 0 {\n\t\tcfg.Port = 8080\n\t}\n\treturn cfg\n}\n",
	}

	files := map[string]string{
		"README.md":               readme,
		"spec.yaml":               spec,
		"Makefile":                makefile,
		".gitignore":             gitignore,
		"Dockerfile":             dockerfile,
		".github/workflows/ci.yml": ciWorkflow,
		"go.mod":                 goMod,
		"cmd/app/main.go":        mainGo,
		"config.yaml":            "port: 8080\n",
		"config.json":            "{\"port\": 8080}\n",
	}

	for fileName, content := range files {
		path := filepath.Join(*output, fileName)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create scaffold dir for %s: %w", fileName, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write scaffold %s: %w", fileName, err)
		}
	}
	for fileName, content := range coreModuleFiles {
		path := filepath.Join(*output, fileName)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create scaffold dir for %s: %w", fileName, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write scaffold %s: %w", fileName, err)
		}
	}

	fmt.Printf("scaffolded %s\n", *output)
	return nil
}
