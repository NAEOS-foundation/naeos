package create

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToSpec(t *testing.T) {
	cfg := &ProjectConfig{
		Name:          "My Project",
		ModulePath:    "./my-project",
		Description:   "A test project",
		Language:      "go",
		Architecture:  "hexagonal",
		Deployment:    "rolling",
		Port:          8080,
		EnableTesting: true,
	}
	spec := cfg.ToSpec()
	if !strings.Contains(spec, "project: my-project") {
		t.Error("expected project name in spec")
	}
	if !strings.Contains(spec, "pattern: hexagonal") {
		t.Error("expected architecture in spec")
	}
	if !strings.Contains(spec, "port: 8080") {
		t.Error("expected port in spec")
	}
	if !strings.Contains(spec, "strategy: unit") {
		t.Error("expected testing strategy in spec")
	}
}

func TestScaffolderGo(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		Name:          "Test Project",
		ModulePath:    "./test-project",
		Language:      "go",
		Architecture:  "hexagonal",
		Deployment:    "rolling",
		Port:          8080,
		OutputDir:     dir,
		EnableDocker:  true,
		EnableCI:      true,
		EnableTesting: true,
	}
	s := NewScaffolder(true)
	files, err := s.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected files to be generated")
	}
	paths := make(map[string]bool)
	for _, f := range files {
		paths[f.Path] = true
	}
	if !paths[filepath.Join(dir, "test-project.spec.yaml")] {
		t.Error("expected spec file")
	}
	if !paths[filepath.Join(dir, "go.mod")] {
		t.Error("expected go.mod")
	}
	if !paths[filepath.Join(dir, "main.go")] {
		t.Error("expected main.go")
	}
	if !paths[filepath.Join(dir, "Dockerfile")] {
		t.Error("expected Dockerfile")
	}
	if !paths[filepath.Join(dir, ".github", "workflows", "ci.yml")] {
		t.Error("expected CI workflow")
	}
	if !paths[filepath.Join(dir, "main_test.go")] {
		t.Error("expected test file")
	}
}

func TestScaffolderTypeScript(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		Name:         "TS App",
		ModulePath:   "./ts-app",
		Language:     "typescript",
		OutputDir:    dir,
		Port:         3000,
		EnableDocker: false,
		EnableCI:     false,
	}
	s := NewScaffolder(false)
	files, err := s.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	found := false
	for _, f := range files {
		if strings.HasSuffix(f.Path, "package.json") {
			found = true
			if !strings.Contains(f.Content, "ts-app") {
				t.Error("expected package.json to contain project name")
			}
		}
	}
	if !found {
		t.Error("expected package.json")
	}
}

func TestScaffolderPython(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		Name:         "Py App",
		ModulePath:   "./py-app",
		Language:     "python",
		OutputDir:    dir,
		Port:         5000,
		EnableDocker: false,
		EnableCI:     false,
	}
	s := NewScaffolder(false)
	files, err := s.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	found := false
	for _, f := range files {
		if strings.HasSuffix(f.Path, "pyproject.toml") {
			found = true
		}
	}
	if !found {
		t.Error("expected pyproject.toml")
	}
}

func TestScaffolderExecute(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		Name:         "Exec Test",
		ModulePath:   "./exec-test",
		Language:     "go",
		Architecture: "hexagonal",
		Deployment:   "rolling",
		Port:         8080,
		OutputDir:    dir,
		EnableDocker: false,
		EnableCI:     false,
	}
	s := NewScaffolder(false)
	if err := s.Execute(cfg); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); os.IsNotExist(err) {
		t.Error("expected main.go to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); os.IsNotExist(err) {
		t.Error("expected go.mod to be created")
	}
}

func TestScaffolderDryRun(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		Name:         "Dry Run",
		ModulePath:   "./dry-run",
		Language:     "go",
		Architecture: "hexagonal",
		Deployment:   "rolling",
		Port:         8080,
		OutputDir:    dir,
	}
	s := NewScaffolder(true)
	if err := s.Execute(cfg); err != nil {
		t.Fatalf("Execute dry-run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Error("file should not exist in dry-run mode")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		cfg     ProjectConfig
		wantErr bool
	}{
		{ProjectConfig{Name: "ok", ModulePath: "./ok", Language: "go", Architecture: "hexagonal", Port: 8080}, false},
		{ProjectConfig{Name: "", ModulePath: "./ok", Language: "go", Architecture: "hexagonal", Port: 8080}, true},
		{ProjectConfig{Name: "ok", ModulePath: "", Language: "go", Architecture: "hexagonal", Port: 8080}, true},
		{ProjectConfig{Name: "ok", ModulePath: "./ok", Language: "invalid", Architecture: "hexagonal", Port: 8080}, true},
		{ProjectConfig{Name: "ok", ModulePath: "./ok", Language: "go", Architecture: "invalid", Port: 8080}, true},
		{ProjectConfig{Name: "ok", ModulePath: "./ok", Language: "go", Architecture: "hexagonal", Port: 0}, true},
	}
	for _, tt := range tests {
		errs := ValidateConfig(&tt.cfg)
		hasErr := len(errs) > 0
		if hasErr != tt.wantErr {
			t.Errorf("config %+v: expected error=%v, got errors=%v", tt.cfg, tt.wantErr, errs)
		}
	}
}

func TestGenerateGitignore(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateGitignore(&ProjectConfig{Language: "go", OutputDir: "dist"})
	if !strings.Contains(content, "*.test") {
		t.Error("expected Go gitignore entries")
	}
	if !strings.Contains(content, "cover.out") {
		t.Error("expected cover.out in gitignore")
	}
	content = s.generateGitignore(&ProjectConfig{Language: "typescript", OutputDir: "dist"})
	if !strings.Contains(content, "node_modules/") {
		t.Error("expected node_modules in gitignore")
	}
}

func TestGenerateDockerfile(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateDockerfile(&ProjectConfig{Language: "go", Port: 8080})
	if !strings.Contains(content, "golang:1.25-alpine") {
		t.Error("expected Go Dockerfile")
	}
	content = s.generateDockerfile(&ProjectConfig{Language: "typescript", Port: 3000})
	if !strings.Contains(content, "node:20-alpine") {
		t.Error("expected Node Dockerfile")
	}
	content = s.generateDockerfile(&ProjectConfig{Language: "python", Port: 5000})
	if !strings.Contains(content, "python:3.12-slim") {
		t.Error("expected Python Dockerfile")
	}
	content = s.generateDockerfile(&ProjectConfig{Language: "java", Port: 8080})
	if !strings.Contains(content, "python") {
		t.Error("expected default Dockerfile for Java")
	}
}

func TestToSpecWithoutTesting(t *testing.T) {
	cfg := &ProjectConfig{
		Name:          "test",
		Description:   "",
		Language:      "go",
		Architecture:  "clean",
		Deployment:    "canary",
		Port:          9090,
		EnableTesting: false,
	}
	spec := cfg.ToSpec()
	if strings.Contains(spec, "testing:") {
		t.Error("expected no testing section when disabled")
	}
	if strings.Contains(spec, "description:") {
		t.Error("expected no description when empty")
	}
}

func TestGenerateDockerCompose(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateDockerCompose(&ProjectConfig{Port: 8080})
	if !strings.Contains(content, `"8080:8080"`) {
		t.Error("expected port mapping in docker-compose")
	}
}

func TestGenerateCIWorkflow(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateCIWorkflow(&ProjectConfig{Name: "My App"})
	if !strings.Contains(content, "name: CI") {
		t.Error("expected CI workflow header")
	}
	if !strings.Contains(content, "my-app") {
		t.Error("expected project name in CI")
	}
}

func TestGenerateREADME(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateREADME(&ProjectConfig{
		Name:         "Test App",
		Description:  "A test",
		Architecture: "clean",
		Language:     "go",
		Deployment:   "rolling",
	})
	if !strings.Contains(content, "Test App") {
		t.Error("expected project name in README")
	}
	if !strings.Contains(content, "clean") {
		t.Error("expected architecture in README")
	}
}

func TestGenerateMakefile(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateMakefile(&ProjectConfig{Name: "My App"})
	if !strings.Contains(content, "my-app") {
		t.Error("expected binary name in Makefile")
	}
}

func TestGenerateMainGo(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateMainGo(&ProjectConfig{Name: "My API", Port: 8080})
	if !strings.Contains(content, "My API") {
		t.Error("expected app name in main.go")
	}
	if !strings.Contains(content, `port = "8080"`) {
		t.Error("expected default port in main.go")
	}
}

func TestGenerateTSIndex(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateTSIndex(&ProjectConfig{Name: "TS App", Port: 3000})
	if !strings.Contains(content, "TS App") {
		t.Error("expected app name in TS index")
	}
	if !strings.Contains(content, "3000") {
		t.Error("expected port in TS index")
	}
}

func TestGeneratePythonMain(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generatePythonMain(&ProjectConfig{Name: "Py App", Port: 5000})
	if !strings.Contains(content, "Py App") {
		t.Error("expected app name in Python main")
	}
	if !strings.Contains(content, "5000") {
		t.Error("expected port in Python main")
	}
}

func TestGenerateGitignorePython(t *testing.T) {
	s := NewScaffolder(false)
	content := s.generateGitignore(&ProjectConfig{Language: "python", OutputDir: "dist"})
	if !strings.Contains(content, "__pycache__/") {
		t.Error("expected python gitignore entries")
	}
}

func TestScaffolderGenerateWithoutDockerCI(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		Name:         "NoDockerCI",
		ModulePath:   "./nodockerci",
		Language:     "go",
		OutputDir:    dir,
		Port:         8080,
		EnableDocker: false,
		EnableCI:     false,
	}
	s := NewScaffolder(true)
	files, err := s.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, f := range files {
		if strings.Contains(f.Path, "Dockerfile") || strings.Contains(f.Path, "ci.yml") {
			t.Errorf("unexpected file when docker/ci disabled: %s", f.Path)
		}
	}
}

func TestScaffolderJava(t *testing.T) {
	dir := t.TempDir()
	cfg := &ProjectConfig{
		Name:       "JavaApp",
		ModulePath: "./javaapp",
		Language:   "java",
		OutputDir:  dir,
		Port:       8080,
	}
	s := NewScaffolder(true)
	files, err := s.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	paths := make(map[string]bool)
	for _, f := range files {
		paths[f.Path] = true
	}
	if !paths[filepath.Join(dir, "README.md")] {
		t.Error("expected README.md for Java project")
	}
}

func TestValidateConfigWithAllValidLanguages(t *testing.T) {
	langs := []string{"go", "typescript", "python", "java", "rust"}
	for _, lang := range langs {
		errs := ValidateConfig(&ProjectConfig{
			Name: "test", ModulePath: "./test",
			Language: lang, Architecture: "hexagonal", Port: 8080,
		})
		if len(errs) > 0 {
			t.Errorf("expected valid for language %q, got errors: %v", lang, errs)
		}
	}
}

func TestValidateConfigPortTooHigh(t *testing.T) {
	errs := ValidateConfig(&ProjectConfig{
		Name: "test", ModulePath: "./test",
		Language: "go", Architecture: "hexagonal", Port: 70000,
	})
	if len(errs) == 0 {
		t.Error("expected error for port out of range")
	}
}

func TestNewWizard(t *testing.T) {
	w := NewWizard()
	if w == nil {
		t.Fatal("expected non-nil wizard")
	}
	if w.reader == nil {
		t.Error("expected non-nil reader")
	}
}

func TestWizard_askRequired(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("my-project\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askRequired("Project name")
	if result != "my-project" {
		t.Errorf("expected my-project, got %s", result)
	}
}

func TestWizard_askDefault(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askDefault("Module path", "./default")
	if result != "./default" {
		t.Errorf("expected ./default, got %s", result)
	}
}

func TestWizard_askDefaultCustom(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("./custom\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askDefault("Module path", "./default")
	if result != "./custom" {
		t.Errorf("expected ./custom, got %s", result)
	}
}

func TestWizard_askChoiceDefault(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askChoice("Language", []string{"go", "typescript"}, "go")
	if result != "go" {
		t.Errorf("expected go, got %s", result)
	}
}

func TestWizard_askChoiceSelected(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("2\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askChoice("Language", []string{"go", "typescript", "python"}, "go")
	if result != "typescript" {
		t.Errorf("expected typescript, got %s", result)
	}
}

func TestWizard_askChoiceInvalid(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("99\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askChoice("Language", []string{"go", "typescript"}, "go")
	if result != "go" {
		t.Errorf("expected go fallback, got %s", result)
	}
}

func TestWizard_askIntDefault(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askInt("Port")
	if result != 8080 {
		t.Errorf("expected 8080, got %d", result)
	}
}

func TestWizard_askIntCustom(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("3000\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askInt("Port")
	if result != 3000 {
		t.Errorf("expected 3000, got %d", result)
	}
}

func TestWizard_askIntInvalid(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("not-a-number\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askInt("Port")
	if result != 8080 {
		t.Errorf("expected 8080 fallback, got %d", result)
	}
}

func TestWizard_askYesNoDefaultNo(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askYesNo("Enable auth", false)
	if result {
		t.Error("expected false")
	}
}

func TestWizard_askYesNoDefaultYes(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askYesNo("Enable testing", true)
	if !result {
		t.Error("expected true")
	}
}

func TestWizard_askYesNoYes(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("y\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askYesNo("Enable auth", false)
	if !result {
		t.Error("expected true")
	}
}

func TestWizard_askRequiredRetry(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("\nmy-project\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askRequired("Project name")
	if result != "my-project" {
		t.Errorf("expected my-project, got %s", result)
	}
}

func TestNewScaffolder(t *testing.T) {
	s := NewScaffolder(true)
	if s == nil {
		t.Fatal("expected non-nil scaffolder")
	}
	if !s.dryRun {
		t.Error("expected dryRun true")
	}
	s2 := NewScaffolder(false)
	if s2.dryRun {
		t.Error("expected dryRun false")
	}
}

func TestScaffolder_AddFile(t *testing.T) {
	s := NewScaffolder(false)
	s.addFile("test.txt", "hello")
	if len(s.files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(s.files))
	}
	if s.files[0].Path != "test.txt" {
		t.Errorf("expected test.txt, got %s", s.files[0].Path)
	}
	if s.files[0].Content != "hello" {
		t.Errorf("expected hello, got %s", s.files[0].Content)
	}
}

func TestScaffolder_GenerateReturnsSpec(t *testing.T) {
	s := NewScaffolder(true)
	files, err := s.Generate(&ProjectConfig{
		Name: "test", OutputDir: "/tmp/out", Language: "go",
		ModulePath: "./test", Architecture: "hexagonal", Port: 8080,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	foundSpec := false
	for _, f := range files {
		if strings.HasSuffix(f.Path, ".spec.yaml") {
			foundSpec = true
			if !strings.Contains(f.Content, "project: test") {
				t.Error("spec missing project name")
			}
		}
	}
	if !foundSpec {
		t.Error("expected spec file")
	}
}

func TestScaffolder_ExecuteMkdirAllError(t *testing.T) {
	s := NewScaffolder(false)
	s.addFile("/proc/readonly/forbidden/file.txt", "content")
	err := s.Execute(&ProjectConfig{OutputDir: "/out"})
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestScaffolder_ExecuteWriteFileError(t *testing.T) {
	s := NewScaffolder(false)
	s.addFile("/root/test_write_denied.txt", "content")
	err := s.Execute(&ProjectConfig{OutputDir: "/out"})
	if err == nil {
		t.Error("expected error when writing to protected path")
	}
}

func TestWizard_askYesNoNo(t *testing.T) {
	r, w, _ := os.Pipe()
	w.WriteString("n\n")
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	result := wiz.askYesNo("Enable auth", true)
	if result {
		t.Error("expected false")
	}
}

func TestGenerateSpec(t *testing.T) {
	s := NewScaffolder(false)
	cfg := &ProjectConfig{Name: "test", Architecture: "hexagonal", Port: 8080, Language: "go"}
	spec := s.generateSpec(cfg)
	if !strings.Contains(spec, "project: test") {
		t.Error("expected spec content")
	}
}

func TestWizard_Run(t *testing.T) {
	input := "MyProject\n./myproject\nA test project\n1\n1\n1\n8080\nmyproject\nn\n\ny\ny\n"
	r, w, _ := os.Pipe()
	w.WriteString(input)
	w.Close()
	wiz := &Wizard{reader: bufio.NewReader(r)}
	cfg, err := wiz.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cfg.Name != "MyProject" {
		t.Errorf("expected MyProject, got %s", cfg.Name)
	}
	if cfg.Language != "go" {
		t.Errorf("expected go, got %s", cfg.Language)
	}
	if cfg.Architecture != "hexagonal" {
		t.Errorf("expected hexagonal, got %s", cfg.Architecture)
	}
	if cfg.Deployment != "rolling" {
		t.Errorf("expected rolling, got %s", cfg.Deployment)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected 8080, got %d", cfg.Port)
	}
	if cfg.EnableAuth {
		t.Error("expected auth disabled")
	}
	if !cfg.EnableTesting {
		t.Error("expected testing enabled")
	}
	if !cfg.EnableDocker {
		t.Error("expected docker enabled")
	}
	if !cfg.EnableCI {
		t.Error("expected CI enabled")
	}
}
