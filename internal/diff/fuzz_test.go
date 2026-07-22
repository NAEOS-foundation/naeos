package diff

import (
	"testing"
)

func FuzzComputeDiff(f *testing.F) {
	f.Add("hello world", "hello world", "same.txt")
	f.Add("old content", "new content", "modified.txt")
	f.Add("something", "", "removed.txt")
	f.Add("", "new file", "added.txt")
	f.Add("", "", "empty.txt")
	f.Add("line1\nline2\nline3", "line1\nmodified\nline3", "multi.txt")

	f.Fuzz(func(t *testing.T, old, new, path string) {
		d := ComputeDiff(old, new, path)
		if d == nil {
			t.Fatal("diff should not be nil")
		}
		if d.Path != path {
			t.Errorf("expected path %q, got %q", path, d.Path)
		}
		if d.Type != ChangeAdded && d.Type != ChangeRemoved && d.Type != ChangeModified && d.Type != ChangeUnchanged {
			t.Errorf("invalid type: %s", d.Type)
		}
	})
}

func FuzzFormatDiff(f *testing.F) {
	f.Add("new.txt", "hello", string(ChangeAdded))
	f.Add("del.txt", "gone", string(ChangeRemoved))
	f.Add("mod.txt", "changed", string(ChangeModified))
	f.Add("same.txt", "same", string(ChangeUnchanged))

	f.Fuzz(func(t *testing.T, path, content, changeType string) {
		d := &FileDiff{
			Path:    path,
			Type:    ChangeType(changeType),
			NewSize: len(content),
			Lines:   addedLines(content),
		}
		output := FormatDiff(d)
		_ = output
	})
}

func FuzzApplyPatch(f *testing.F) {
	f.Add("original line", "new line", "test.txt")
	f.Add("", "added", "new.txt")
	f.Add("to delete", "", "del.txt")
	f.Add("line1\nline2", "line1\nchanged", "mod.txt")

	f.Fuzz(func(t *testing.T, old, new, path string) {
		d := ComputeDiff(old, new, path)
		result := ApplyPatch(old, d)
		_ = result
	})
}

func FuzzFormatUnified(f *testing.F) {
	f.Add("line1\nline2", "line1\nchanged", "test.txt", 3)
	f.Add("only", "different", "test2.txt", 1)

	f.Fuzz(func(t *testing.T, old, new, path string, contextLines int) {
		d := ComputeDiff(old, new, path)
		output := FormatUnified(d, contextLines)
		_ = output
	})
}
