package benchmark

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type testEvent struct {
	Action string `json:"action"`
	Bench  bool   `json:"Benchmark"`
	Output string `json:"output"`
}

type benchMetrics struct {
	nsOp   float64
	allocs float64
	bytes  float64
}

// syncWriter serializes writes from concurrent goroutines to the same
// underlying writer (parsed benchmark output and copied stderr).
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func newSyncWriter(w io.Writer) *syncWriter {
	return &syncWriter{w: w}
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

var runCommand = exec.CommandContext

func RunBenchmarks(packagePath string, output io.Writer) (*Baseline, error) {
	args := []string{"test", "-bench=.", "-benchmem", "-count=5", "-run=^$", "-json", packagePath}
	cmd := runCommand(context.Background(), "go", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	out := newSyncWriter(output)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(out, stderr)
	}()

	results := &Baseline{}
	benchData := make(map[string]*benchMetrics)
	benchCounts := make(map[string]int)

	scanner := bufio.NewScanner(io.TeeReader(stdout, out))
	for scanner.Scan() {
		name, m, ok := parseBenchLine(scanner.Text())
		if !ok {
			continue
		}
		if _, exists := benchData[name]; !exists {
			benchData[name] = &benchMetrics{}
		}
		benchData[name].nsOp += m.nsOp
		benchData[name].allocs += m.allocs
		benchData[name].bytes += m.bytes
		benchCounts[name]++
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("go test: %w", err)
	}

	results.Results = summarize(benchData, benchCounts)
	return results, nil
}

func parseBenchLine(line string) (string, benchMetrics, bool) {
	var ev testEvent
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return "", benchMetrics{}, false
	}
	if ev.Action != "output" || !strings.HasPrefix(ev.Output, "Benchmark") {
		return "", benchMetrics{}, false
	}
	text := strings.TrimSpace(ev.Output)
	parts := strings.Fields(text)
	if len(parts) < 5 {
		return "", benchMetrics{}, false
	}
	name := parts[0]
	if idx := strings.Index(name, "/"); idx > 0 {
		name = name[:idx]
	}

	m := benchMetrics{}
	m.nsOp, _ = strconv.ParseFloat(strings.ReplaceAll(parts[2], ",", ""), 64)
	m.bytes, _ = strconv.ParseFloat(strings.ReplaceAll(parts[4], ",", ""), 64)
	if len(parts) >= 7 {
		m.allocs, _ = strconv.ParseFloat(strings.ReplaceAll(parts[6], ",", ""), 64)
	}
	return name, m, true
}

func summarize(data map[string]*benchMetrics, counts map[string]int) []Result {
	var results []Result
	for name, m := range data {
		count := counts[name]
		if count == 0 {
			continue
		}
		results = append(results, Result{
			Name:        name,
			NsPerOp:     m.nsOp / float64(count),
			AllocsPerOp: m.allocs / float64(count),
			BytesPerOp:  m.bytes / float64(count),
		})
	}
	return results
}
