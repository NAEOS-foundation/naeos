package diff

import (
	"os"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func TestComputeDiffAdded(t *testing.T) {
	d := ComputeDiff("", "hello world", "new.txt")
	if d.Type != ChangeAdded {
		t.Errorf("expected ChangeAdded, got %s", d.Type)
	}
	if d.NewSize != 11 {
		t.Errorf("expected size 11, got %d", d.NewSize)
	}
}

func TestComputeDiffRemoved(t *testing.T) {
	d := ComputeDiff("hello world", "", "deleted.txt")
	if d.Type != ChangeRemoved {
		t.Errorf("expected ChangeRemoved, got %s", d.Type)
	}
	if d.OldSize != 11 {
		t.Errorf("expected size 11, got %d", d.OldSize)
	}
}

func TestComputeDiffUnchanged(t *testing.T) {
	d := ComputeDiff("same content", "same content", "same.txt")
	if d.Type != ChangeUnchanged {
		t.Errorf("expected ChangeUnchanged, got %s", d.Type)
	}
}

func TestComputeDiffModified(t *testing.T) {
	d := ComputeDiff("old content", "new content", "modified.txt")
	if d.Type != ChangeModified {
		t.Errorf("expected ChangeModified, got %s", d.Type)
	}
	if d.OldSize != 11 || d.NewSize != 11 {
		t.Errorf("expected sizes 11/11, got %d/%d", d.OldSize, d.NewSize)
	}
}

func TestComputeDiffBothEmpty(t *testing.T) {
	d := ComputeDiff("", "", "empty.txt")
	if d.Type != ChangeUnchanged {
		t.Errorf("expected ChangeUnchanged, got %s", d.Type)
	}
}

func TestSummary(t *testing.T) {
	diffs := []*FileDiff{
		{Type: ChangeAdded},
		{Type: ChangeRemoved},
		{Type: ChangeModified},
		{Type: ChangeUnchanged},
		{Type: ChangeModified},
	}
	added, removed, modified, unchanged := Summary(diffs)
	if added != 1 || removed != 1 || modified != 2 || unchanged != 1 {
		t.Errorf("expected 1,1,2,1 got %d,%d,%d,%d", added, removed, modified, unchanged)
	}
}

func TestSummaryEmpty(t *testing.T) {
	added, removed, modified, unchanged := Summary([]*FileDiff{})
	if added != 0 || removed != 0 || modified != 0 || unchanged != 0 {
		t.Errorf("expected all zeros, got %d,%d,%d,%d", added, removed, modified, unchanged)
	}
}

func TestFormatDiffAdded(t *testing.T) {
	d := &FileDiff{Path: "new.txt", Type: ChangeAdded}
	output := FormatDiff(d)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormatDiffRemoved(t *testing.T) {
	d := &FileDiff{Path: "del.txt", Type: ChangeRemoved, Lines: []DiffLine{{Type: ChangeRemoved, Content: "gone"}}}
	output := FormatDiff(d)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormatDiffModified(t *testing.T) {
	d := &FileDiff{
		Path: "mod.txt", Type: ChangeModified,
		OldSize: 5, NewSize: 10,
		Lines: []DiffLine{
			{Type: ChangeRemoved, Content: "old"},
			{Type: ChangeAdded, Content: "new"},
		},
	}
	output := FormatDiff(d)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormatDiffUnchanged(t *testing.T) {
	d := &FileDiff{Path: "same.txt", Type: ChangeUnchanged}
	output := FormatDiff(d)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormatDiffNil(t *testing.T) {
	output := FormatDiff(nil)
	if output != "" {
		t.Error("expected empty output for nil")
	}
}

func TestFormatUnifiedAdded(t *testing.T) {
	d := &FileDiff{Path: "new.txt", Type: ChangeAdded, Lines: []DiffLine{{Content: "hello"}, {Content: "world"}}}
	output := FormatUnified(d, 3)
	if output == "" {
		t.Fatal("expected non-empty unified diff")
	}
}

func TestFormatUnifiedRemoved(t *testing.T) {
	d := &FileDiff{Path: "del.txt", Type: ChangeRemoved, Lines: []DiffLine{{Content: "bye"}}}
	output := FormatUnified(d, 3)
	if output == "" {
		t.Fatal("expected non-empty unified diff")
	}
}

func TestFormatUnifiedModified(t *testing.T) {
	d := &FileDiff{
		Path: "mod.txt", Type: ChangeModified,
		Lines: []DiffLine{
			{OldNum: 1, NewNum: 0, Type: ChangeRemoved, Content: "old"},
			{OldNum: 0, NewNum: 1, Type: ChangeAdded, Content: "new"},
			{OldNum: 2, NewNum: 2, Type: ChangeUnchanged, Content: "keep"},
		},
	}
	output := FormatUnified(d, 1)
	if output == "" {
		t.Fatal("expected non-empty unified diff")
	}
}

func TestFormatUnifiedNil(t *testing.T) {
	if FormatUnified(nil, 3) != "" {
		t.Error("expected empty for nil")
	}
}

func TestFormatUnifiedUnchanged(t *testing.T) {
	d := &FileDiff{Path: "same.txt", Type: ChangeUnchanged}
	if FormatUnified(d, 3) != "" {
		t.Error("expected empty for unchanged")
	}
}

func TestComputeLineDiff(t *testing.T) {
	old := []string{"line1", "line2", "line3"}
	new := []string{"line1", "modified", "line3", "line4"}
	lines := computeLineDiff(old, new)
	if len(lines) == 0 {
		t.Error("expected non-empty diff lines")
	}
}

func TestComputeLineDiffBothEmpty(t *testing.T) {
	lines := computeLineDiff(nil, nil)
	if lines != nil {
		t.Error("expected nil for both empty")
	}
}

func TestComputeLineDiffOneEmpty(t *testing.T) {
	lines := computeLineDiff([]string{"a"}, nil)
	if len(lines) == 0 {
		t.Error("expected non-empty for one side empty")
	}
}

func TestAddedLines(t *testing.T) {
	lines := addedLines("a\nb\nc")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if l.Type != ChangeAdded {
			t.Errorf("expected ChangeAdded, got %s", l.Type)
		}
	}
}

func TestRemovedLines(t *testing.T) {
	lines := removedLines("a\nb")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if l.Type != ChangeRemoved {
			t.Errorf("expected ChangeRemoved, got %s", l.Type)
		}
	}
}

func TestCountLinesEmpty(t *testing.T) {
	if countLines("") != 0 {
		t.Error("expected 0 for empty string")
	}
}

func TestCountLinesNonEmpty(t *testing.T) {
	if countLines("a\nb") != 2 {
		t.Error("expected 2 lines")
	}
}

func TestApplyPatchAdded(t *testing.T) {
	d := ComputeDiff("", "new content", "new.txt")
	result := ApplyPatch("original", d)
	if result != "original\nnew content\n" {
		t.Errorf("expected appended content, got %q", result)
	}
}

func TestApplyPatchRemoved(t *testing.T) {
	d := ComputeDiff("old content", "", "del.txt")
	result := ApplyPatch("original", d)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestApplyPatchUnchanged(t *testing.T) {
	d := ComputeDiff("same", "same", "same.txt")
	result := ApplyPatch("original", d)
	if result != "original" {
		t.Errorf("expected 'original', got %q", result)
	}
}

func TestApplyPatchModified(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nmodified\nline3"
	d := ComputeDiff(old, new, "mod.txt")
	result := ApplyPatch(old, d)
	if result != new {
		t.Errorf("expected %q, got %q", new, result)
	}
}

func TestApplyPatchNil(t *testing.T) {
	result := ApplyPatch("original", nil)
	if result != "original" {
		t.Error("expected original for nil diff")
	}
}

func TestComputeDirectoryDiff(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	writeFile(t, oldDir, "keep.txt", "same")
	writeFile(t, oldDir, "remove.txt", "to remove")
	writeFile(t, newDir, "keep.txt", "same")
	writeFile(t, newDir, "added.txt", "new file")

	diffs := ComputeDirectoryDiff(oldDir, newDir, []string{"keep.txt", "remove.txt", "added.txt"})
	if len(diffs) != 3 {
		t.Fatalf("expected 3 diffs, got %d", len(diffs))
	}
	if diffs[0].Type != ChangeUnchanged {
		t.Errorf("expected keep.txt unchanged, got %s", diffs[0].Type)
	}
	if diffs[1].Type != ChangeRemoved {
		t.Errorf("expected remove.txt removed, got %s", diffs[1].Type)
	}
	if diffs[2].Type != ChangeAdded {
		t.Errorf("expected added.txt added, got %s", diffs[2].Type)
	}
}

func TestDiffNEIREmpty(t *testing.T) {
	diff := ComputeNEIRDiff(&model.NEIR{}, &model.NEIR{})
	if diff.Summary != "no changes" {
		t.Errorf("expected 'no changes', got %q", diff.Summary)
	}
	if diff.ProjectDiff == nil {
		t.Fatal("expected non-nil ProjectDiff")
	}
	if diff.ProjectDiff.NameChanged {
		t.Error("expected NameChanged=false for empty NEIRs")
	}
}

func TestDiffModulesSame(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp", Version: "1.0.0"},
		Modules: []module.Module{
			{Name: "core", Path: "./core"},
			{Name: "util", Path: "./util"},
		},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp", Version: "1.0.0"},
		Modules: []module.Module{
			{Name: "core", Path: "./core"},
			{Name: "util", Path: "./util"},
		},
	}
	diff := ComputeNEIRDiff(old, new)
	if diff.Summary != "no changes" {
		t.Errorf("expected 'no changes', got %q", diff.Summary)
	}
	if diff.ProjectDiff.NameChanged {
		t.Error("expected no project name change")
	}
}

func TestDiffModulesAdded(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Modules: []module.Module{
			{Name: "core"},
		},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Modules: []module.Module{
			{Name: "core"},
			{Name: "new-module"},
		},
	}
	diff := ComputeNEIRDiff(old, new)
	if diff.ServicesDiff == nil {
		t.Fatal("expected non-nil ServicesDiff")
	}
	if len(diff.ServicesDiff.Added) != 0 || len(diff.ServicesDiff.Removed) != 0 {
		t.Errorf("expected no service changes for module diff, added=%d removed=%d",
			len(diff.ServicesDiff.Added), len(diff.ServicesDiff.Removed))
	}
	if !containsStrInDiff(diff.Summary, "no changes") {
		t.Errorf("expected 'no changes' in summary, got %q", diff.Summary)
	}
}

func TestDiffServicesRemoved(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080, Kind: service.KindHTTP},
			{Name: "worker", Port: 9090, Kind: service.KindWorker},
		},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080, Kind: service.KindHTTP},
		},
	}
	diff := ComputeNEIRDiff(old, new)
	if diff.ServicesDiff == nil {
		t.Fatal("expected non-nil ServicesDiff")
	}
	if len(diff.ServicesDiff.Removed) != 1 {
		t.Fatalf("expected 1 removed service, got %d", len(diff.ServicesDiff.Removed))
	}
	if diff.ServicesDiff.Removed[0].Name != "worker" {
		t.Errorf("expected removed service 'worker', got %q", diff.ServicesDiff.Removed[0].Name)
	}
	if !containsStrInDiff(diff.Summary, "-1 services") {
		t.Errorf("expected '-1 services' in summary, got %q", diff.Summary)
	}
}

func containsStrInDiff(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(dir+"/"+name, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
