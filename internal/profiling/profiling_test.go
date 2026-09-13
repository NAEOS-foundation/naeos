// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package profiling

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPipelineProfile(t *testing.T) {
	p := NewProfile()
	p.Start()
	p.StartStage("compile")
	time.Sleep(time.Millisecond)
	p.EndStage("compile")
	p.StartStage("test")
	time.Sleep(time.Millisecond)
	p.EndStage("test")
	p.Finish()

	if p.TotalTime < 2*time.Millisecond {
		t.Errorf("expected total time >= 2ms, got %s", p.TotalTime)
	}
	if len(p.Stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(p.Stages))
	}
	if p.Stages[0].Duration == 0 {
		t.Error("expected non-zero duration for compile stage")
	}
	if p.Stages[0].Memory.Alloc == 0 {
		t.Error("expected memory stats to be captured")
	}
	slowest := p.SlowestStage()
	if slowest == nil || slowest.Name != "compile" && slowest.Name != "test" {
		t.Error("expected a slowest stage")
	}
	fastest := p.FastestStage()
	if fastest == nil {
		t.Error("expected a fastest stage")
	}
}

func TestSummary(t *testing.T) {
	p := NewProfile()
	p.Start()
	p.StartStage("a")
	p.EndStage("a")
	p.Finish()
	summary := p.Summary()
	if len(summary) == 0 {
		t.Error("expected non-empty summary")
	}
}

func TestBenchmark(t *testing.T) {
	result := Benchmark("test-bench", 100, func() {
		time.Sleep(time.Microsecond * 10)
	})
	if result.Iterations != 100 {
		t.Errorf("expected 100 iterations, got %d", result.Iterations)
	}
	if result.AvgTime == 0 {
		t.Error("expected non-zero avg time")
	}
	if result.P50 == 0 {
		t.Error("expected non-zero P50")
	}
	if result.P99 == 0 {
		t.Error("expected non-zero P99")
	}
	if result.StdDev == 0 {
		t.Error("expected non-zero stddev")
	}
	str := result.String()
	if len(str) == 0 {
		t.Error("expected non-empty string")
	}
}

func TestBenchmarkHistogram(t *testing.T) {
	result := Benchmark("test-hist", 50, func() {
		time.Sleep(time.Microsecond)
	})
	hist := result.Histogram()
	if len(hist) == 0 {
		t.Error("expected non-empty histogram")
	}
}

func TestProfileSnapshot(t *testing.T) {
	p := NewProfile()
	p.Start()
	p.StartStage("step1")
	p.EndStage("step1")
	p.Finish()
	snap := p.Snapshot()
	if snap.StageCount != 1 {
		t.Errorf("expected 1 stage, got %d", snap.StageCount)
	}
	if len(snap.StageNames) != 1 || snap.StageNames[0] != "step1" {
		t.Error("expected step1 in stage names")
	}
}

func TestSaveAndLoadProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")
	p := NewProfile()
	p.Start()
	p.StartStage("build")
	p.EndStage("build")
	p.Finish()
	if err := SaveProfile(path, p); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}
	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}
	if loaded.StageCount != 1 {
		t.Errorf("expected 1 stage, got %d", loaded.StageCount)
	}
	if loaded.StageNames[0] != "build" {
		t.Errorf("expected build, got %s", loaded.StageNames[0])
	}
}

func TestCompareProfiles(t *testing.T) {
	baseline := &ProfileSnapshot{
		StageNames: []string{"a", "b"},
		StageDurs:  []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}
	compare := &ProfileSnapshot{
		StageNames: []string{"a", "b"},
		StageDurs:  []time.Duration{12 * time.Millisecond, 15 * time.Millisecond},
	}
	comps := CompareProfiles(baseline, compare)
	if len(comps) != 2 {
		t.Fatalf("expected 2 comparisons, got %d", len(comps))
	}
	if !comps[0].Regression {
		t.Error("stage a should be a regression (20% increase > 10% threshold)")
	}
	if comps[1].Regression {
		t.Error("stage b should not be a regression (it decreased)")
	}
}

func TestGCPauseTracker(t *testing.T) {
	tracker := NewGCPauseTracker()
	tracker.Snapshot()
	tracker.Snapshot()
	pauses := tracker.Pauses()
	if len(pauses) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(pauses))
	}
	total := tracker.TotalPause()
	if total < 0 {
		t.Error("expected non-negative total pause")
	}
}

func TestPercentile(t *testing.T) {
	sorted := []time.Duration{
		time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
	}
	p50 := percentile(sorted, 50)
	if p50 != 3*time.Millisecond {
		t.Errorf("expected P50=3ms, got %s", p50)
	}
	p99 := percentile(sorted, 99)
	if p99 != 5*time.Millisecond {
		t.Errorf("expected P99=5ms, got %s", p99)
	}
}

func TestSlowestStageEmpty(t *testing.T) {
	p := NewProfile()
	if p.SlowestStage() != nil {
		t.Error("expected nil for empty profile")
	}
	if p.FastestStage() != nil {
		t.Error("expected nil for empty profile")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		b    uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1048576, "1.0 MiB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.b)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.b, got, tt.want)
		}
	}
}

func TestSaveLoadNonexistent(t *testing.T) {
	_, err := LoadProfile("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error loading nonexistent file")
	}
}
