package main

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

func TestArtifactStatsMixedFiles(t *testing.T) {
	p := New()
	ctx := &pluginhost.PluginContext{}
	if err := p.Initialize(ctx); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	t.Cleanup(func() { _ = p.Shutdown() })

	result, err := p.Execute("stats", map[string]any{
		"files": []any{
			map[string]any{"path": "src/main.go", "content": "package main\n\nfunc main() {}\n"},
			map[string]any{"path": "web/app.ts", "content": "export const x = 1;\n"},
			map[string]any{"path": "README", "content": "# demo\n"},
		},
	})
	if err != nil {
		t.Fatalf("Execute stats: %v", err)
	}

	r := result.(map[string]any)
	if r["files"] != 3 {
		t.Errorf("expected 3 files, got %v", r["files"])
	}
	if r["lines"] != 8 {
		t.Errorf("expected 8 lines, got %v", r["lines"])
	}

	byExt, ok := r["by_ext"].(map[string]map[string]int)
	if !ok {
		t.Fatalf("expected by_ext map, got %T", r["by_ext"])
	}
	if byExt[".go"]["files"] != 1 || byExt[".go"]["lines"] != 4 || byExt[".ts"]["lines"] != 2 || byExt["(none)"]["files"] != 1 {
		t.Errorf("unexpected per-extension stats: %v", byExt)
	}
}

func TestArtifactStatsEmpty(t *testing.T) {
	p := New()

	result, err := p.Execute("stats", map[string]any{"files": []any{}})
	if err == nil {
		t.Fatal("expected error for empty files list")
	}

	_, err = p.Execute("stats", nil)
	if err == nil {
		t.Fatal("expected error for missing files param")
	}

	_ = result
}

func TestArtifactStatsIgnoresNonMapItems(t *testing.T) {
	p := New()

	result, err := p.Execute("stats", map[string]any{
		"files": []any{"not-a-map", 42},
	})
	if err != nil {
		t.Fatalf("Execute stats: %v", err)
	}

	r := result.(map[string]any)
	if r["files"] != 0 {
		t.Errorf("expected 0 counted files, got %v", r["files"])
	}
}

func TestArtifactStatsPingDescribe(t *testing.T) {
	p := New()

	ping, err := p.Execute("ping", nil)
	if err != nil {
		t.Fatalf("Execute ping: %v", err)
	}
	if ping.(map[string]string)["status"] != "ok" {
		t.Errorf("unexpected ping result: %v", ping)
	}

	desc, err := p.Execute("describe", nil)
	if err != nil {
		t.Fatalf("Execute describe: %v", err)
	}
	if desc.(map[string]any)["name"] != "artifact-stats" {
		t.Errorf("unexpected describe result: %v", desc)
	}
}
