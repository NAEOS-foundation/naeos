package profiling

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

type HeapSnapshot struct {
	Alloc        uint64    `json:"alloc"`
	TotalAlloc   uint64    `json:"total_alloc"`
	Sys          uint64    `json:"sys"`
	NumGC        uint32    `json:"num_gc"`
	PauseTotalNs uint64    `json:"pause_total_ns"`
	NumForcedGC  uint32    `json:"num_forced_gc"`
	Timestamp    time.Time `json:"timestamp"`
	Label        string    `json:"label,omitempty"`
}

type HeapDiff struct {
	Label           string        `json:"label"`
	AllocDelta      int64         `json:"alloc_delta"`
	TotalAllocDelta uint64        `json:"total_alloc_delta"`
	SysDelta        int64         `json:"sys_delta"`
	GCCount         uint32        `json:"gc_count"`
	PauseTotalDelta time.Duration `json:"pause_total_delta"`
	Duration        time.Duration `json:"duration"`
	LeakScore       float64       `json:"leak_score"`
}

type MemProfiler struct {
	mu        sync.Mutex
	snapshots []HeapSnapshot
	startedAt time.Time
}

func NewMemProfiler() *MemProfiler {
	return &MemProfiler{
		snapshots: make([]HeapSnapshot, 0, 16),
	}
}

func takeHeapSnapshot(label string) HeapSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return HeapSnapshot{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
		NumForcedGC:  m.NumForcedGC,
		Timestamp:    time.Now(),
		Label:        label,
	}
}

func (m *MemProfiler) Snapshot(label string) HeapSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	snap := takeHeapSnapshot(label)
	m.snapshots = append(m.snapshots, snap)
	return snap
}

func (m *MemProfiler) Snapshots() []HeapSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]HeapSnapshot, len(m.snapshots))
	copy(out, m.snapshots)
	return out
}

func (m *MemProfiler) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startedAt = time.Now()
	m.snapshots = m.snapshots[:0]
	m.snapshots = append(m.snapshots, takeHeapSnapshot("start"))
}

func (m *MemProfiler) Diffs() []HeapDiff {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.snapshots) < 2 {
		return nil
	}
	diffs := make([]HeapDiff, 0, len(m.snapshots)-1)
	for i := 1; i < len(m.snapshots); i++ {
		prev := m.snapshots[i-1]
		cur := m.snapshots[i]
		diff := HeapDiff{
			Label:           cur.Label,
			AllocDelta:      int64(cur.Alloc) - int64(prev.Alloc), //nolint:gosec // safe: alloc diff fits int64
			TotalAllocDelta: cur.TotalAlloc - prev.TotalAlloc,
			SysDelta:        int64(cur.Sys) - int64(prev.Sys), //nolint:gosec // safe: sys diff fits int64
			GCCount:         cur.NumGC - prev.NumGC,
			PauseTotalDelta: time.Duration(cur.PauseTotalNs - prev.PauseTotalNs), //nolint:gosec // safe: pause diff fits int64
			Duration:        cur.Timestamp.Sub(prev.Timestamp),
		}
		if diff.Duration > 0 {
			allocPerSec := float64(diff.AllocDelta) / diff.Duration.Seconds()
			gcPauseRatio := float64(diff.PauseTotalDelta) / float64(diff.Duration)
			diff.LeakScore = allocPerSec * gcPauseRatio / 1024
			if diff.LeakScore < 0 {
				diff.LeakScore = 0
			}
		}
		diffs = append(diffs, diff)
	}
	return diffs
}

func (m *MemProfiler) DetectLeaks(threshold float64) []HeapDiff {
	diffs := m.Diffs()
	if len(diffs) == 0 {
		return nil
	}
	var leaks []HeapDiff
	for _, d := range diffs {
		if d.LeakScore > threshold && d.AllocDelta > 0 {
			leaks = append(leaks, d)
		}
	}
	return leaks
}

func (m *MemProfiler) Summary() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var sb strings.Builder
	sb.WriteString("Memory Profile Summary\n")
	sb.WriteString("======================\n\n")

	for i, snap := range m.snapshots {
		fmt.Fprintf(&sb, "  [%d] %s\n", i, snap.Label)
		fmt.Fprintf(&sb, "       Alloc:      %s\n", formatBytes(snap.Alloc))
		fmt.Fprintf(&sb, "       TotalAlloc: %s\n", formatBytes(snap.TotalAlloc))
		fmt.Fprintf(&sb, "       Sys:        %s\n", formatBytes(snap.Sys))
		fmt.Fprintf(&sb, "       NumGC:      %d\n", snap.NumGC)
		fmt.Fprintf(&sb, "       PauseTotal: %s\n", time.Duration(snap.PauseTotalNs).Round(time.Microsecond)) //nolint:gosec // safe: pause total fits int64
	}

	diffs := m.computeDiffsLocked()
	if len(diffs) > 0 {
		sb.WriteString("\nHeap Diffs:\n")
		for _, d := range diffs {
			fmt.Fprintf(&sb, "  %s:\n", d.Label)
			fmt.Fprintf(&sb, "       Alloc delta:      %+d (%s)\n", d.AllocDelta, formatBytes(uint64(abs64(d.AllocDelta)))) //nolint:gosec // safe: abs delta fits uint64
			fmt.Fprintf(&sb, "       Sys delta:        %+d (%s)\n", d.SysDelta, formatBytes(uint64(abs64(d.SysDelta))))     //nolint:gosec // safe: abs delta fits uint64
			fmt.Fprintf(&sb, "       GC count:         %d\n", d.GCCount)
			fmt.Fprintf(&sb, "       Pause total:      %s\n", d.PauseTotalDelta.Round(time.Microsecond))
			fmt.Fprintf(&sb, "       Leak score:       %.2f\n", d.LeakScore)
		}
	}
	return sb.String()
}

func (m *MemProfiler) computeDiffsLocked() []HeapDiff {
	if len(m.snapshots) < 2 {
		return nil
	}
	diffs := make([]HeapDiff, 0, len(m.snapshots)-1)
	for i := 1; i < len(m.snapshots); i++ {
		prev := m.snapshots[i-1]
		cur := m.snapshots[i]
		diff := HeapDiff{
			Label:           cur.Label,
			AllocDelta:      int64(cur.Alloc) - int64(prev.Alloc), //nolint:gosec // safe: alloc diff fits int64
			TotalAllocDelta: cur.TotalAlloc - prev.TotalAlloc,
			SysDelta:        int64(cur.Sys) - int64(prev.Sys), //nolint:gosec // safe: sys diff fits int64
			GCCount:         cur.NumGC - prev.NumGC,
			PauseTotalDelta: time.Duration(cur.PauseTotalNs - prev.PauseTotalNs), //nolint:gosec // safe: pause diff fits int64
			Duration:        cur.Timestamp.Sub(prev.Timestamp),
		}
		if diff.Duration > 0 {
			allocPerSec := float64(diff.AllocDelta) / diff.Duration.Seconds()
			gcPauseRatio := float64(diff.PauseTotalDelta) / float64(diff.Duration)
			diff.LeakScore = allocPerSec * gcPauseRatio / 1024
			if diff.LeakScore < 0 {
				diff.LeakScore = 0
			}
		}
		diffs = append(diffs, diff)
	}
	return diffs
}

type LeakReport struct {
	Suspected bool        `json:"suspected"`
	Stages    []string    `json:"stages,omitempty"`
	Delta     int64       `json:"delta_bytes"`
	Details   string      `json:"details,omitempty"`
	Diffs     []HeapDiff  `json:"diffs,omitempty"`
	GCStats   GCLeakStats `json:"gc_stats,omitempty"`
}

type GCLeakStats struct {
	TotalPause time.Duration `json:"total_pause"`
	PausePerGC time.Duration `json:"pause_per_gc"`
	GCCount    uint32        `json:"gc_count"`
	GCPressure float64       `json:"gc_pressure"`
}

func (m *MemProfiler) Analyze() LeakReport {
	m.mu.Lock()
	defer m.mu.Unlock()
	report := LeakReport{}

	if len(m.snapshots) < 2 {
		return report
	}

	first := m.snapshots[0]
	last := m.snapshots[len(m.snapshots)-1]
	allocDelta := int64(last.Alloc) - int64(first.Alloc) //nolint:gosec // safe: alloc diff fits int64

	diffs := m.computeDiffsLocked()
	report.Diffs = diffs

	totalPause := time.Duration(last.PauseTotalNs - first.PauseTotalNs) //nolint:gosec // safe: pause diff fits int64
	gcCount := last.NumGC - first.NumGC

	if gcCount > 0 {
		report.GCStats = GCLeakStats{
			TotalPause: totalPause,
			PausePerGC: totalPause / time.Duration(gcCount),
			GCCount:    gcCount,
			GCPressure: float64(totalPause) / float64(time.Since(first.Timestamp)),
		}
	}

	if allocDelta > 0 {
		allocPerSec := float64(allocDelta) / time.Since(first.Timestamp).Seconds()
		isSustained := true
		var growingStages []string
		for i := 1; i < len(m.snapshots); i++ {
			delta := int64(m.snapshots[i].Alloc) - int64(m.snapshots[i-1].Alloc) //nolint:gosec // safe: alloc diff fits int64
			if delta <= 0 {
				isSustained = false
			}
			if delta > 0 {
				growingStages = append(growingStages, m.snapshots[i].Label)
			}
		}

		if isSustained && allocPerSec > 1024 {
			report.Suspected = true
			report.Stages = growingStages
			report.Delta = allocDelta
			report.Details = fmt.Sprintf(
				"sustained allocation growth detected: +%s across %d stages (%.1f KB/s)",
				formatBytes(uint64(allocDelta)), len(growingStages), allocPerSec/1024, //nolint:gosec // safe: abs alloc fits uint64
			)
		}
	}

	return report
}

func (m *MemProfiler) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots = make([]HeapSnapshot, 0, 16)
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
