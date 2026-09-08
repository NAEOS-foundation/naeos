package benchmark

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResultJSONRoundtrip(t *testing.T) {
	t.Parallel()
	tt := []Result{
		{Name: "BenchmarkParser", NsPerOp: 123.5, AllocsPerOp: 4, BytesPerOp: 512},
		{Name: "BenchmarkCompile/go", NsPerOp: 0.5, AllocsPerOp: 0, BytesPerOp: 0},
		{},
	}
	for _, tc := range tt {
		data, err := json.Marshal(tc)
		if err != nil {
			t.Fatalf("marshal %+v: %v", tc, err)
		}
		var got Result
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", string(data), err)
		}
		if got != tc {
			t.Errorf("roundtrip mismatch: got %+v want %+v", got, tc)
		}
	}
}

func TestBaselineJSONRoundtrip(t *testing.T) {
	t.Parallel()
	b := Baseline{Commit: "abc123", Date: "2026-09-05", Results: []Result{
		{Name: "BenchmarkA", NsPerOp: 1.0},
		{Name: "BenchmarkB", NsPerOp: 2.0},
	}}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Baseline
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Results) != 2 || got.Results[0].Name != "BenchmarkA" {
		t.Errorf("unexpected roundtrip result: %+v", got)
	}
}

func TestLoadBaseline(t *testing.T) {
	t.Parallel()

	t.Run("valid file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "baseline.json")
		b := Baseline{Commit: "deadbeef", Results: []Result{{Name: "BenchmarkX", NsPerOp: 10}}}
		if err := b.SaveBaseline(path); err != nil {
			t.Fatalf("SaveBaseline: %v", err)
		}
		got, err := LoadBaseline(path)
		if err != nil {
			t.Fatalf("LoadBaseline: %v", err)
		}
		if got.Commit != "deadbeef" || len(got.Results) != 1 || got.Results[0].Name != "BenchmarkX" {
			t.Errorf("loaded baseline mismatch: %+v", got)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		_, err := LoadBaseline(filepath.Join(t.TempDir(), "nope.json"))
		if err == nil {
			t.Fatal("expected error for missing file")
		}
		if !strings.Contains(err.Error(), "read baseline") {
			t.Errorf("expected wrapped read baseline error, got %q", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		_, err := LoadBaseline(path)
		if err == nil {
			t.Fatal("expected error for invalid json")
		}
		if !strings.Contains(err.Error(), "unmarshal baseline") {
			t.Errorf("expected wrapped unmarshal error, got %q", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "empty.json")
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		_, err := LoadBaseline(path)
		if err == nil {
			t.Fatal("expected error for empty file")
		}
		if !strings.Contains(err.Error(), "unmarshal baseline") {
			t.Errorf("expected wrapped unmarshal error, got %q", err)
		}
	})
}

func TestSaveBaseline(t *testing.T) {
	t.Parallel()

	t.Run("writes file with content", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "out.json")
		b := Baseline{Commit: "abcd", Results: []Result{{Name: "BenchmarkY", NsPerOp: 7}}}
		if err := b.SaveBaseline(path); err != nil {
			t.Fatalf("SaveBaseline: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if !strings.Contains(string(data), "BenchmarkY") {
			t.Errorf("file content missing benchmark name: %s", string(data))
		}
		var loaded Baseline
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("unmarshal written file: %v", err)
		}
	})

	t.Run("file permission 0600", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "perm.json")
		b := Baseline{Commit: "x"}
		if err := b.SaveBaseline(path); err != nil {
			t.Fatalf("SaveBaseline: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("expected 0600 permission, got %o", perm)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		t.Parallel()
		b := Baseline{Commit: "x"}
		err := b.SaveBaseline(filepath.Join(t.TempDir(), "no", "dir", "x.json"))
		if err == nil {
			t.Fatal("expected error for invalid path")
		}
	})
}

func TestCompare(t *testing.T) {
	t.Parallel()

	base := &Baseline{Results: []Result{
		{Name: "BenchmarkA", NsPerOp: 100, AllocsPerOp: 10, BytesPerOp: 1000},
		{Name: "BenchmarkB", NsPerOp: 50, AllocsPerOp: 5, BytesPerOp: 500},
		{Name: "BenchmarkSlow", NsPerOp: 1000},
	}}

	t.Run("no regression", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkA", NsPerOp: 100, AllocsPerOp: 10, BytesPerOp: 1000},
			{Name: "BenchmarkB", NsPerOp: 50, AllocsPerOp: 5, BytesPerOp: 500},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 0 {
			t.Errorf("expected no regressions, got %+v", r)
		}
	})

	t.Run("nsPerOp regression flagged", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkA", NsPerOp: 150, AllocsPerOp: 10, BytesPerOp: 1000},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 1 {
			t.Fatalf("expected 1 regression, got %+v", r)
		}
		if r[0].Name != "BenchmarkA/nsPerOp" || r[0].Old != 100 || r[0].New != 150 {
			t.Errorf("unexpected regression: %+v", r[0])
		}
		if r[0].Delta != 50 {
			t.Errorf("expected delta 50, got %v", r[0].Delta)
		}
	})

	t.Run("allocsPerOp regression flagged", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkB", NsPerOp: 50, AllocsPerOp: 20, BytesPerOp: 500},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 1 || r[0].Name != "BenchmarkB/allocsPerOp" {
			t.Fatalf("expected allocs regression, got %+v", r)
		}
	})

	t.Run("improvement not flagged", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkA", NsPerOp: 50, AllocsPerOp: 10, BytesPerOp: 1000},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 0 {
			t.Errorf("expected no regression on improvement, got %+v", r)
		}
	})

	t.Run("exactly 10%% not flagged", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkB", NsPerOp: 55, AllocsPerOp: 5, BytesPerOp: 500},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 0 {
			t.Errorf("expected delta==10 to be excluded, got %+v", r)
		}
	})

	t.Run("just over 10%% flagged", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkB", NsPerOp: 56, AllocsPerOp: 5, BytesPerOp: 500},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 1 || r[0].Name != "BenchmarkB/nsPerOp" {
			t.Errorf("expected regression at 12%% increase, got %+v", r)
		}
	})

	t.Run("zero baseline values skipped", func(t *testing.T) {
		t.Parallel()
		base := &Baseline{Results: []Result{
			{Name: "BenchmarkC", NsPerOp: 0, AllocsPerOp: 0, BytesPerOp: 0},
		}}
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkC", NsPerOp: 100, AllocsPerOp: 100, BytesPerOp: 100},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 0 {
			t.Errorf("expected zero baseline to be skipped, got %+v", r)
		}
	})

	t.Run("new benchmark in current ignored", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkNEW", NsPerOp: 9999, AllocsPerOp: 9999, BytesPerOp: 9999},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 0 {
			t.Errorf("expected new-only benchmark to be ignored, got %+v", r)
		}
	})

	t.Run("multiple regressions in one result", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkA", NsPerOp: 200, AllocsPerOp: 20, BytesPerOp: 2000},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 2 {
			t.Errorf("expected 2 regressions (ns + allocs), got %+v", r)
		}
	})

	t.Run("bytesPerOp increase not flagged", func(t *testing.T) {
		t.Parallel()
		cur := Baseline{Results: []Result{
			{Name: "BenchmarkB", NsPerOp: 50, AllocsPerOp: 5, BytesPerOp: 5000},
		}}
		r, err := base.Compare(cur)
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 0 {
			t.Errorf("expected bytes increase to be ignored, got %+v", r)
		}
	})

	t.Run("empty baseline and current", func(t *testing.T) {
		t.Parallel()
		b := &Baseline{}
		r, err := b.Compare(Baseline{})
		if err != nil {
			t.Fatalf("Compare: %v", err)
		}
		if len(r) != 0 {
			t.Errorf("expected empty result, got %+v", r)
		}
	})
}

func TestRegressionFieldTags(t *testing.T) {
	t.Parallel()
	r := Regression{Name: "B/ns", Old: 1, New: 2, Delta: 100}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Regression
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Name != "B/ns" || got.Old != 1 || got.New != 2 || got.Delta != 100 {
		t.Errorf("regression roundtrip mismatch: %+v", got)
	}
}
