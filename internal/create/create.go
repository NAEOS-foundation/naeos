package create

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Wizard struct {
	reader *bufio.Reader
}

type ProjectConfig struct {
	Name           string
	ModulePath     string
	Language       string
	Architecture   string
	Deployment     string
	Port           int
	OutputDir      string
	Description    string
	EnableAuth     bool
	EnableTesting  bool
	EnableDocker   bool
	EnableCI       bool
	DryRun         bool
	NonInteractive bool
}

func NewWizard() *Wizard {
	return &Wizard{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (w *Wizard) Run() (*ProjectConfig, error) {
	cfg := &ProjectConfig{}

	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║     NAEOS Project Creation Wizard    ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()

	cfg.Name = w.askRequired("Project name")
	cfg.ModulePath = w.askDefault("Module path", "./"+strings.ToLower(strings.ReplaceAll(cfg.Name, " ", "-")))
	cfg.Description = w.askDefault("Description", "A NAEOS project")
	cfg.Language = w.askChoice("Language", []string{"go", "typescript", "python", "java", "rust"}, "go")
	cfg.Architecture = w.askChoice("Architecture pattern", []string{"hexagonal", "layered", "clean", "event-driven", "cqrs", "monolith"}, "hexagonal")
	cfg.Deployment = w.askChoice("Deployment strategy", []string{"rolling", "blue-green", "canary", "recreate"}, "rolling")
	cfg.Port = w.askInt("Default port")
	cfg.OutputDir = w.askDefault("Output directory", cfg.Name)
	cfg.EnableAuth = w.askYesNo("Enable authentication", false)
	cfg.EnableTesting = w.askYesNo("Enable test generation", true)
	cfg.EnableDocker = w.askYesNo("Generate Dockerfile", true)
	cfg.EnableCI = w.askYesNo("Generate CI workflow", true)

	fmt.Println()
	fmt.Println("Configuration complete!")
	return cfg, nil
}

func (w *Wizard) askRequired(prompt string) string {
	for {
		fmt.Printf("%s: ", prompt)
		text, _ := w.reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
		fmt.Println("  This field is required.")
	}
}

func (w *Wizard) askDefault(prompt, defaultVal string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultVal)
	text, _ := w.reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	return text
}

func (w *Wizard) askChoice(prompt string, options []string, defaultVal string) string {
	fmt.Printf("%s:\n", prompt)
	for i, opt := range options {
		marker := "  "
		if opt == defaultVal {
			marker = "→ "
		}
		fmt.Printf("  %s%d) %s\n", marker, i+1, opt)
	}
	fmt.Printf("  Choose [1-%d] (default: %s): ", len(options), defaultVal)
	text, _ := w.reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	var idx int
	if _, err := fmt.Sscanf(text, "%d", &idx); err == nil && idx >= 1 && idx <= len(options) {
		return options[idx-1]
	}
	return defaultVal
}

func (w *Wizard) askInt(prompt string) int {
	const defaultVal = 8080
	fmt.Printf("%s [%d]: ", prompt, defaultVal)
	text, _ := w.reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	var val int
	if _, err := fmt.Sscanf(text, "%d", &val); err == nil {
		return val
	}
	return defaultVal
}

func (w *Wizard) askYesNo(prompt string, defaultVal bool) bool {
	defaultStr := "y/N"
	if defaultVal {
		defaultStr = "Y/n"
	}
	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	text, _ := w.reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return defaultVal
	}
	return text == "y" || text == "yes"
}

func (cfg *ProjectConfig) ToSpec() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "project: %s\n", strings.ToLower(strings.ReplaceAll(cfg.Name, " ", "-")))
	if cfg.Description != "" {
		fmt.Fprintf(&sb, "description: %s\n", cfg.Description)
	}
	sb.WriteString("\nmodules:\n")
	fmt.Fprintf(&sb, "  - name: core\n    path: %s\n", cfg.ModulePath)
	sb.WriteString("\nservices:\n")
	fmt.Fprintf(&sb, "  - name: api\n    kind: http\n    port: %d\n", cfg.Port)
	sb.WriteString("\narchitecture:\n")
	fmt.Fprintf(&sb, "  pattern: %s\n", cfg.Architecture)
	sb.WriteString("\ndeployment:\n")
	fmt.Fprintf(&sb, "  strategy: %s\n", cfg.Deployment)
	if cfg.EnableTesting {
		sb.WriteString("\ntesting:\n")
		sb.WriteString("  strategy: unit\n")
		sb.WriteString("  coverage: standard\n")
	}
	return sb.String()
}

type Scaffolder struct {
	dryRun bool
	files  []ScaffoldFile
}

type ScaffoldFile struct {
	Path    string
	Content string
	Mode    os.FileMode
}

func NewScaffolder(dryRun bool) *Scaffolder {
	return &Scaffolder{dryRun: dryRun}
}

func (s *Scaffolder) Generate(cfg *ProjectConfig) ([]ScaffoldFile, error) {
	projectName := strings.ToLower(strings.ReplaceAll(cfg.Name, " ", "-"))
	baseDir := cfg.OutputDir

	s.files = nil
	s.addFile(filepath.Join(baseDir, projectName+".spec.yaml"), s.generateSpec(cfg))
	s.addFile(filepath.Join(baseDir, ".gitignore"), s.generateGitignore(cfg))
	s.addFile(filepath.Join(baseDir, "README.md"), s.generateREADME(cfg))
	s.addFile(filepath.Join(baseDir, "Makefile"), s.generateMakefile(cfg))

	switch cfg.Language {
	case "go":
		s.addGoProject(cfg, baseDir)
	case "typescript":
		s.addTypeScriptProject(cfg, projectName, baseDir)
	case "python":
		s.addPythonProject(cfg, projectName, baseDir)
	}

	if cfg.EnableDocker {
		s.addFile(filepath.Join(baseDir, "Dockerfile"), s.generateDockerfile(cfg))
		s.addFile(filepath.Join(baseDir, "docker-compose.yml"), s.generateDockerCompose(cfg))
	}
	if cfg.EnableCI {
		s.addFile(filepath.Join(baseDir, ".github", "workflows", "ci.yml"), s.generateCIWorkflow(cfg))
	}

	return s.files, nil
}

func (s *Scaffolder) addFile(path, content string) {
	s.files = append(s.files, ScaffoldFile{Path: path, Content: content, Mode: 0o644})
}

func (s *Scaffolder) addGoProject(cfg *ProjectConfig, baseDir string) {
	s.addFile(filepath.Join(baseDir, "go.mod"), fmt.Sprintf("module %s\n\ngo 1.25\n", cfg.ModulePath))
	s.addFile(filepath.Join(baseDir, "main.go"), s.generateMainGo(cfg))
	if cfg.EnableTesting {
		s.addFile(filepath.Join(baseDir, "main_test.go"), s.generateMainTest(cfg))
	}
}

func (s *Scaffolder) addTypeScriptProject(cfg *ProjectConfig, name, baseDir string) {
	pkgJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "0.1.0",
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js",
    "test": "jest"
  }
}
`, name)
	s.addFile(filepath.Join(baseDir, "package.json"), pkgJSON)
	s.addFile(filepath.Join(baseDir, "tsconfig.json"), `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true
  }
}
`)
	s.addFile(filepath.Join(baseDir, "src", "index.ts"), s.generateTSIndex(cfg))
}

func (s *Scaffolder) addPythonProject(cfg *ProjectConfig, name, baseDir string) {
	s.addFile(filepath.Join(baseDir, "pyproject.toml"), fmt.Sprintf(`[project]
name = "%s"
version = "0.1.0"
requires-python = ">=3.10"
`, name))
	s.addFile(filepath.Join(baseDir, "src", "__init__.py"), "")
	s.addFile(filepath.Join(baseDir, "src", "main.py"), s.generatePythonMain(cfg))
}

func (s *Scaffolder) generateSpec(cfg *ProjectConfig) string {
	return cfg.ToSpec()
}

func (s *Scaffolder) generateGitignore(cfg *ProjectConfig) string {
	lines := []string{
		"# Build artifacts",
		"*.exe", "*.dll", "*.so", "*.dylib",
		"/" + cfg.OutputDir,
	}
	switch cfg.Language {
	case "go":
		lines = append(lines, "", "# Go", "*.test", "cover.out", "vendor/")
	case "typescript":
		lines = append(lines, "", "# Node", "node_modules/", "dist/", ".tsbuildinfo")
	case "python":
		lines = append(lines, "", "# Python", "__pycache__/", "*.pyc", ".venv/", "*.egg-info/")
	}
	lines = append(lines, "", "# IDE", ".idea/", ".vscode/", "*.swp", "", "# OS", ".DS_Store", "Thumbs.db")
	return strings.Join(lines, "\n") + "\n"
}

func (s *Scaffolder) generateREADME(cfg *ProjectConfig) string {
	return fmt.Sprintf(`# %s

%s

## Getting Started

`+"```bash"+`
make build
make test
`+"```"+`

## Project Structure

- Architecture: %s
- Language: %s
- Deployment: %s

## License

Apache 2.0
`, cfg.Name, cfg.Description, cfg.Architecture, cfg.Language, cfg.Deployment)
}

func (s *Scaffolder) generateMakefile(cfg *ProjectConfig) string {
	return fmt.Sprintf(`.PHONY: build test clean run

BINARY := %s

build:
	go build -o $(BINARY) .

test:
	go test -v -race ./...

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY)
`, strings.ToLower(strings.ReplaceAll(cfg.Name, " ", "-")))
}

func (s *Scaffolder) generateDockerfile(cfg *ProjectConfig) string {
	switch cfg.Language {
	case "go":
		return fmt.Sprintf(`FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/server /server
EXPOSE %d
CMD ["/server"]
`, cfg.Port)
	case "typescript":
		return fmt.Sprintf(`FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
EXPOSE %d
CMD ["node", "dist/index.js"]
`, cfg.Port)
	default:
		return fmt.Sprintf(`FROM python:3.12-slim
WORKDIR /app
COPY . .
RUN pip install --no-cache-dir -r requirements.txt 2>/dev/null || true
EXPOSE %d
CMD ["python", "src/main.py"]
`, cfg.Port)
	}
}

func (s *Scaffolder) generateDockerCompose(cfg *ProjectConfig) string {
	return fmt.Sprintf(`version: "3.8"

services:
  app:
    build: .
    ports:
      - "%d:%d"
    environment:
      - PORT=%d
    restart: unless-stopped
`, cfg.Port, cfg.Port, cfg.Port)
}

func (s *Scaffolder) generateCIWorkflow(cfg *ProjectConfig) string {
	name := strings.ToLower(strings.ReplaceAll(cfg.Name, " ", "-"))
	return fmt.Sprintf(`name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build
        run: go build -o %s .
      - name: Test
        run: go test -v -race ./...
`, name)
}

func (s *Scaffolder) generateMainGo(cfg *ProjectConfig) string {
	return fmt.Sprintf(`package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "%d"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s!")
	})

	fmt.Printf("Server starting on :%%s\\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v\\n", err)
		os.Exit(1)
	}
}
`, cfg.Port, cfg.Name)
}

func (s *Scaffolder) generateMainTest(cfg *ProjectConfig) string {
	return `package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
`
}

func (s *Scaffolder) generateTSIndex(cfg *ProjectConfig) string {
	return fmt.Sprintf(`import http from "http";

const server = http.createServer((req, res) => {
  res.writeHead(200, { "Content-Type": "text/plain" });
  res.end("Hello from %s!");
});

const port = process.env.PORT || "%d";
server.listen(port, () => {
  console.log("Server running on port " + port);
});
`, cfg.Name, cfg.Port)
}

func (s *Scaffolder) generatePythonMain(cfg *ProjectConfig) string {
	return fmt.Sprintf(`import os
from http.server import HTTPServer, BaseHTTPRequestHandler

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.end_headers()
        self.wfile.write(b"Hello from %s!")

if __name__ == "__main__":
    port = int(os.environ.get("PORT", %d))
    server = HTTPServer(("0.0.0.0", port), Handler)
    print(f"Server running on port {port}")
    server.serve_forever()
`, cfg.Name, cfg.Port)
}

func (s *Scaffolder) Execute(cfg *ProjectConfig) error {
	files, err := s.Generate(cfg)
	if err != nil {
		return err
	}
	for _, f := range files {
		if s.dryRun {
			fmt.Printf("[dry-run] %s (%d bytes)\n", f.Path, len(f.Content))
			continue
		}
		dir := filepath.Dir(f.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
		if err := os.WriteFile(f.Path, []byte(f.Content), f.Mode); err != nil {
			return fmt.Errorf("write %s: %w", f.Path, err)
		}
	}
	return nil
}

func ValidateConfig(cfg *ProjectConfig) []string {
	var errs []string
	if cfg.Name == "" {
		errs = append(errs, "name is required")
	}
	if cfg.ModulePath == "" {
		errs = append(errs, "module path is required")
	}
	validLangs := map[string]bool{"go": true, "typescript": true, "python": true, "java": true, "rust": true}
	if !validLangs[cfg.Language] {
		errs = append(errs, "unsupported language: "+cfg.Language)
	}
	validArch := map[string]bool{"hexagonal": true, "layered": true, "clean": true, "event-driven": true, "cqrs": true, "monolith": true}
	if !validArch[cfg.Architecture] {
		errs = append(errs, "unsupported architecture: "+cfg.Architecture)
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, "port must be between 1 and 65535")
	}
	return errs
}
