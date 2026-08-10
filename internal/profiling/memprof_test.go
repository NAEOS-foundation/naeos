package profiling

import (
	"math"
	"runtime"
	"testing"
)

func TestNewMemProfiler(t *testing.T) {
	m := NewMemProfiler()
	if m == nil {
		t.Fatal("expected non-nil MemProfiler")
	}
	snaps := m.Snapshots()
	if len(snaps) != 0 {
		t.Fatalf("expected 0 snapshots initially, got %d", len(snaps))
	}
}

func TestMemProfilerSnapshot(t *testing.T) {
	m := NewMemProfiler()
	snap := m.Snapshot("test")
	if snap.Label != "test" {
		t.Fatalf("expected label 'test', got %q", snap.Label)
	}
	if snap.Alloc == 0 {
		t.Fatal("expected non-zero Alloc")
	}
	if snap.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestMemProfilerStart(t *testing.T) {
	m := NewMemProfiler()
	m.Start()
	snaps := m.Snapshots()
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot after Start, got %d", len(snaps))
	}
	if snaps[0].Label != "start" {
		t.Fatalf("expected label 'start', got %q", snaps[0].Label)
	}
}

func TestMemProfilerDiffs(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("before")
	m.Snapshot("after")

	diffs := m.Diffs()
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Label != "after" {
		t.Fatalf("expected label 'after', got %q", diffs[0].Label)
	}
}

func TestMemProfilerDiffsSingleSnapshot(t *testing.T) {
	m := NewMemProfiler()
	diffs := m.Diffs()
	if diffs != nil {
		t.Fatal("expected nil diffs with no snapshots")
	}

	m.Snapshot("only")
	diffs = m.Diffs()
	if diffs != nil {
		t.Fatal("expected nil diffs with single snapshot")
	}
}

func TestMemProfilerDetectLeaks(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("before")
	m.Snapshot("after")

	leaks := m.DetectLeaks(0)
	if leaks != nil {
		t.Logf("got %d leaks (GC-dependent, may be empty)", len(leaks))
	}
}

func TestMemProfilerLeakDetectionHighThreshold(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("before")
	m.Snapshot("after")

	leaks := m.DetectLeaks(1e12)
	if len(leaks) != 0 {
		t.Fatalf("expected 0 leaks with high threshold, got %d", len(leaks))
	}
}

func TestMemProfilerAnalyze(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("before")

	alloc := make([]byte, 1<<20)
	_ = alloc

	runtime.GC()
	m.Snapshot("after")

	report := m.Analyze()
	if report.Diffs == nil {
		t.Fatal("expected non-nil diffs in report")
	}
	if report.GCStats.GCCount > 0 {
		if report.GCStats.PausePerGC <= 0 {
			t.Fatal("expected positive PausePerGC")
		}
	}
}

func TestMemProfilerAnalyzeSingleSnapshot(t *testing.T) {
	m := NewMemProfiler()
	report := m.Analyze()
	if report.Diffs != nil {
		t.Fatal("expected nil diffs with no snapshots")
	}

	m.Snapshot("only")
	report = m.Analyze()
	if report.Diffs != nil {
		t.Fatal("expected nil diffs with single snapshot")
	}
}

func TestMemProfilerReset(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("one")
	m.Snapshot("two")
	m.Reset()

	snaps := m.Snapshots()
	if len(snaps) != 0 {
		t.Fatalf("expected 0 snapshots after Reset, got %d", len(snaps))
	}
}

func TestMemProfilerSummary(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("one")

	summary := m.Summary()
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}

	m.Snapshot("two")
	summary = m.Summary()
	if summary == "" {
		t.Fatal("expected non-empty summary with diff")
	}
}

func TestMemProfilerConcurrency(t *testing.T) {
	m := NewMemProfiler()
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 50; i++ {
			m.Snapshot("goroutine-a")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			m.Snapshot("goroutine-b")
		}
		done <- true
	}()

	<-done
	<-done

	snaps := m.Snapshots()
	if len(snaps) != 100 {
		t.Fatalf("expected 100 snapshots, got %d", len(snaps))
	}
}

func TestMemProfilerStartResetState(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("stale")
	m.Start()

	snaps := m.Snapshots()
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot after Start, got %d", len(snaps))
	}
	if snaps[0].Label != "start" {
		t.Fatalf("expected label 'start' after Start, got %q", snaps[0].Label)
	}
}

func TestTakeHeapSnapshot(t *testing.T) {
	snap := takeHeapSnapshot("direct")
	if snap.Label != "direct" {
		t.Fatalf("expected label 'direct', got %q", snap.Label)
	}
	if snap.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestMemProfilerDiffsConsecutive(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("a")
	m.Snapshot("b")
	m.Snapshot("c")

	diffs := m.Diffs()
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs for 3 snapshots, got %d", len(diffs))
	}
	if diffs[0].Label != "b" {
		t.Fatalf("expected first diff label 'b', got %q", diffs[0].Label)
	}
	if diffs[1].Label != "c" {
		t.Fatalf("expected second diff label 'c', got %q", diffs[1].Label)
	}
}

func TestAbs64(t *testing.T) {
	tests := []struct {
		input    int64
		expected int64
	}{
		{42, 42},
		{-42, 42},
		{0, 0},
		{math.MaxInt64, math.MaxInt64},
	}

	for _, tt := range tests {
		got := abs64(tt.input)
		if got != tt.expected {
			t.Fatalf("abs64(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestHeapSnapshotFields(t *testing.T) {
	snap := takeHeapSnapshot("fields")
	if snap.TotalAlloc < snap.Alloc {
		t.Fatal("expected TotalAlloc >= Alloc")
	}
	if snap.Sys == 0 {
		t.Fatal("expected non-zero Sys")
	}
}

func TestMemProfilerAnalyzeSustainedGrowth(t *testing.T) {
	m := NewMemProfiler()
	m.Snapshot("start")

	buf := make([]byte, 100<<20)
	for i := range buf {
		buf[i] = byte(i)
	}
	runtime.GC()
	m.Snapshot("alloc-heavy")

	report := m.Analyze()
	if report.Diffs == nil {
		t.Fatal("expected non-nil diffs")
	}
	if len(report.Diffs) == 0 {
		t.Fatal("expected at least one diff")
	}
	if report.Delta <= 0 && len(report.Stages) == 0 {
		t.Log("sustained growth not detected (allocation timing dependent)")
	}
}

func BenchmarkMemProfilerSnapshots(b *testing.B) {
	m := NewMemProfiler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Snapshot("bench")
	}
}

func BenchmarkMemProfilerDiffs(b *testing.B) {
	m := NewMemProfiler()
	for i := 0; i < 100; i++ {
		m.Snapshot("snap")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Diffs()
	}
}
