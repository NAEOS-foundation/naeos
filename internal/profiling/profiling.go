package profiling

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type StageMetrics struct {
	Name      string
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
	Memory    MemStats
	Metadata  map[string]any
}

type MemStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
	PauseTotal uint64 `json:"pause_total_ns"`
}

type PipelineProfile struct {
	Stages     []StageMetrics
	TotalStart time.Time
	TotalEnd   time.Time
	TotalTime  time.Duration
	mu         sync.Mutex
}

func NewProfile() *PipelineProfile {
	return &PipelineProfile{}
}

func (p *PipelineProfile) StartStage(name string) *StageMetrics {
	p.mu.Lock()
	defer p.mu.Unlock()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stage := &StageMetrics{
		Name:      name,
		StartedAt: time.Now(),
		Memory: MemStats{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
			PauseTotal: m.PauseTotalNs,
		},
		Metadata: make(map[string]any),
	}
	p.Stages = append(p.Stages, *stage)
	return &p.Stages[len(p.Stages)-1]
}

func (p *PipelineProfile) EndStage(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.Stages {
		if p.Stages[i].Name == name && p.Stages[i].EndedAt.IsZero() {
			p.Stages[i].EndedAt = time.Now()
			p.Stages[i].Duration = p.Stages[i].EndedAt.Sub(p.Stages[i].StartedAt)
			return
		}
	}
}

func (p *PipelineProfile) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TotalEnd = time.Now()
	p.TotalTime = p.TotalEnd.Sub(p.TotalStart)
}

func (p *PipelineProfile) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TotalStart = time.Now()
}

func (p *PipelineProfile) Summary() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var sb strings.Builder
	sb.WriteString("Pipeline Performance Profile\n")
	sb.WriteString("============================\n\n")
	fmt.Fprintf(&sb, "Total time: %s\n\n", p.TotalTime.Round(time.Microsecond))
	sb.WriteString("Stage Breakdown:\n")
	fmt.Fprintf(&sb, "  %-20s %15s %8s %15s\n", "Stage", "Duration", "%", "Memory Alloc")
	fmt.Fprintf(&sb, "  %-20s %15s %8s %15s\n", "-----", "--------", "--", "------------")
	for _, stage := range p.Stages {
		pct := 0.0
		if p.TotalTime > 0 {
			pct = float64(stage.Duration) / float64(p.TotalTime) * 100
		}
		fmt.Fprintf(&sb, "  %-20s %15s %7.1f%% %15s\n",
			stage.Name,
			stage.Duration.Round(time.Microsecond),
			pct,
			formatBytes(stage.Memory.Alloc))
	}
	return sb.String()
}

func (p *PipelineProfile) SlowestStage() *StageMetrics {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Stages) == 0 {
		return nil
	}
	slowest := &p.Stages[0]
	for i := range p.Stages {
		if p.Stages[i].Duration > slowest.Duration {
			slowest = &p.Stages[i]
		}
	}
	return slowest
}

func (p *PipelineProfile) FastestStage() *StageMetrics {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Stages) == 0 {
		return nil
	}
	fastest := &p.Stages[0]
	for i := range p.Stages {
		if p.Stages[i].Duration < fastest.Duration && p.Stages[i].Duration > 0 {
			fastest = &p.Stages[i]
		}
	}
	return fastest
}

type BenchmarkResult struct {
	Name         string
	Iterations   int
	AvgTime      time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	TotalTime    time.Duration
	P50          time.Duration
	P95          time.Duration
	P99          time.Duration
	StdDev       time.Duration
	Percentiles  map[float64]time.Duration
	allDurations []time.Duration
}

func Benchmark(name string, iterations int, fn func()) *BenchmarkResult {
	result := &BenchmarkResult{
		Name:        name,
		Iterations:  iterations,
		MinTime:     time.Hour,
		Percentiles: make(map[float64]time.Duration),
	}
	durations := make([]time.Duration, iterations)
	for i := 0; i < iterations; i++ {
		start := time.Now()
		fn()
		duration := time.Since(start)
		durations[i] = duration
		result.TotalTime += duration
		if duration < result.MinTime {
			result.MinTime = duration
		}
		if duration > result.MaxTime {
			result.MaxTime = duration
		}
	}
	result.allDurations = durations
	result.AvgTime = result.TotalTime / time.Duration(iterations)
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	result.P50 = percentile(sorted, 50)
	result.P95 = percentile(sorted, 95)
	result.P99 = percentile(sorted, 99)
	result.StdDev = computeStdDev(durations, result.AvgTime)
	result.Percentiles[50] = result.P50
	result.Percentiles[95] = result.P95
	result.Percentiles[99] = result.P99
	return result
}

func (r *BenchmarkResult) String() string {
	return fmt.Sprintf("Benchmark %s: %d iterations, avg=%s, min=%s, max=%s, p50=%s, p95=%s, p99=%s, stddev=%s",
		r.Name, r.Iterations,
		r.AvgTime.Round(time.Microsecond),
		r.MinTime.Round(time.Microsecond),
		r.MaxTime.Round(time.Microsecond),
		r.P50.Round(time.Microsecond),
		r.P95.Round(time.Microsecond),
		r.P99.Round(time.Microsecond),
		r.StdDev.Round(time.Microsecond))
}

func (r *BenchmarkResult) Histogram() string {
	buckets := []time.Duration{
		time.Microsecond,
		10 * time.Microsecond,
		100 * time.Microsecond,
		time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		time.Second,
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Histogram: %s\n", r.Name)
	for i, upper := range buckets {
		lower := time.Duration(0)
		if i > 0 {
			lower = buckets[i-1] //nolint:gosec // G602: bounds are safe, i > 0 is checked
		}
		count := 0
		for _, d := range r.durations() {
			if d >= lower && d < upper {
				count++
			}
		}
		bar := strings.Repeat("#", count*40/max(r.Iterations, 1))
		fmt.Fprintf(&sb, "  [%10s, %10s) %4d %s\n", lower, upper, count, bar)
	}
	return sb.String()
}

func (r *BenchmarkResult) Durations() []time.Duration {
	return r.allDurations
}

func (r *BenchmarkResult) durations() []time.Duration {
	return r.allDurations
}

type ProfileSnapshot struct {
	Name       string          `json:"name"`
	TotalTime  time.Duration   `json:"total_time"`
	StageCount int             `json:"stage_count"`
	StageDurs  []time.Duration `json:"stage_durations"`
	StageNames []string        `json:"stage_names"`
	Timestamp  time.Time       `json:"timestamp"`
}

func (p *PipelineProfile) Snapshot() *ProfileSnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	snap := &ProfileSnapshot{
		TotalTime:  p.TotalTime,
		StageCount: len(p.Stages),
		Timestamp:  time.Now(),
	}
	for _, s := range p.Stages {
		snap.StageNames = append(snap.StageNames, s.Name)
		snap.StageDurs = append(snap.StageDurs, s.Duration)
	}
	return snap
}

func SaveProfile(path string, profile *PipelineProfile) error {
	snap := profile.Snapshot()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func LoadProfile(path string) (*ProfileSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap ProfileSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

type ProfileComparison struct {
	StageName   string        `json:"stage_name"`
	Baseline    time.Duration `json:"baseline"`
	Compare     time.Duration `json:"compare"`
	Diff        time.Duration `json:"diff"`
	PercentDiff float64       `json:"percent_diff"`
	Regression  bool          `json:"regression"`
}

func CompareProfiles(baseline, compare *ProfileSnapshot) []ProfileComparison {
	var comps []ProfileComparison
	baselineMap := make(map[string]time.Duration)
	for i, name := range baseline.StageNames {
		if i < len(baseline.StageDurs) {
			baselineMap[name] = baseline.StageDurs[i]
		}
	}
	for i, name := range compare.StageNames {
		compareDur := time.Duration(0)
		if i < len(compare.StageDurs) {
			compareDur = compare.StageDurs[i]
		}
		baselineDur, exists := baselineMap[name]
		if !exists {
			baselineDur = 0
		}
		diff := compareDur - baselineDur
		var pctDiff float64
		if baselineDur > 0 {
			pctDiff = float64(diff) / float64(baselineDur) * 100
		}
		comps = append(comps, ProfileComparison{
			StageName:   name,
			Baseline:    baselineDur,
			Compare:     compareDur,
			Diff:        diff,
			PercentDiff: pctDiff,
			Regression:  diff > 0 && pctDiff > 10,
		})
	}
	return comps
}

type GCPauseTracker struct {
	pauses []time.Duration
	mu     sync.Mutex
}

func NewGCPauseTracker() *GCPauseTracker {
	return &GCPauseTracker{}
}

func (t *GCPauseTracker) Snapshot() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.mu.Lock()
	defer t.mu.Unlock()
	pauseNs := m.PauseTotalNs
	if pauseNs > uint64(math.MaxInt64) {
		pauseNs = uint64(math.MaxInt64)
	}
	t.pauses = append(t.pauses, time.Duration(pauseNs))
}

func (t *GCPauseTracker) Pauses() []time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]time.Duration, len(t.pauses))
	copy(out, t.pauses)
	return out
}

func (t *GCPauseTracker) TotalPause() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	var total time.Duration
	for _, p := range t.pauses {
		total += p
	}
	return total
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func computeStdDev(durations []time.Duration, mean time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sumSq float64
	for _, d := range durations {
		diff := float64(d - mean)
		sumSq += diff * diff
	}
	return time.Duration(math.Sqrt(sumSq / float64(len(durations))))
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
