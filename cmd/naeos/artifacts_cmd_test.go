package main

import (
	"strings"
	"testing"
)

func TestArtifactsListUsesCommandOutput(t *testing.T) {
	root := newRootCommand()
	output, err := executeCommand(root, "artifacts", "list")
	if err != nil {
		t.Fatalf("execute artifacts list failed: %v", err)
	}
	if !strings.Contains(output, "No artifacts tracked.") {
		t.Fatalf("expected artifacts list output in command buffer, got %q", output)
	}
}
