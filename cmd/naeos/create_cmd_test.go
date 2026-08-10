package main

import (
	"os"
	"strings"
	"testing"
)

func TestCreateWizardGeneratesProject(t *testing.T) {
	dir := t.TempDir()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	input := "demo-app\n" + // project name
		"\n" + // module path (default)
		"\n" + // description (default)
		"\n" + // language (default: go)
		"\n" + // architecture (default: hexagonal)
		"\n" + // deployment (default: rolling)
		"\n" + // port (default: 8080)
		dir + "\n" + // output directory -> temp dir
		"\n" + // enable auth (default: no)
		"\n" + // enable testing (default: yes)
		"\n" + // enable docker (default: yes)
		"\n" // enable ci (default: yes)
	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("write wizard input: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	root := NewRootCommand()
	output, err := executeCommand(root, "create")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !strings.Contains(output, "Generated specification") {
		t.Fatalf("expected generated specification, got %q", output)
	}
	if !strings.Contains(output, "Created project demo-app") {
		t.Fatalf("expected project creation message, got %q", output)
	}

	if _, err := os.Stat(dir + "/spec.yaml"); err != nil {
		t.Fatalf("expected spec.yaml: %v", err)
	}
	if _, err := os.Stat(dir + "/config.yaml"); err != nil {
		t.Fatalf("expected config.yaml: %v", err)
	}
}

func TestCreateWizardRequiresProjectName(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	if _, err := w.WriteString("\ndemo-app\n\n\n\n\n\n" + t.TempDir() + "\n\n\n\n\n"); err != nil {
		t.Fatalf("write wizard input: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	root := NewRootCommand()
	output, err := executeCommand(root, "create")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !strings.Contains(output, "Created project demo-app") {
		t.Fatalf("expected project creation after empty name retry, got %q", output)
	}
}
