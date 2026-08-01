package benchmark

import (
	"bufio"
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

func RunBenchmarks(packagePath string, output io.Writer) (*Baseline, error) {
	args := []string{"test", "-bench=.", "-benchmem", "-count=5", "-run=^$", "-json", packagePath}
	cmd := exec.Command("go", args...)

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
	go func() {
		defer wg.Done()
		io.Copy(output, stderr)
	}()

	results := &Baseline{}
	benchData := make(map[string]*benchMetrics)
	benchCounts := make(map[string]int)

	scanner := bufio.NewScanner(io.TeeReader(stdout, output))
	for scanner.Scan() {
		line := scanner.Text()
		var ev testEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		if ev.Action != "output" || !strings.HasPrefix(ev.Output, "Benchmark") {
			continue
		}
		text := strings.TrimSpace(ev.Output)
		parts := strings.Fields(text)
		if len(parts) < 5 {
			continue
		}
		name := parts[0]
		if idx := strings.Index(name, "/"); idx > 0 {
			name = name[:idx]
		}

		nsOp, _ := strconv.ParseFloat(strings.ReplaceAll(parts[2], ",", ""), 64)
		allocs, _ := strconv.ParseFloat(strings.ReplaceAll(parts[4], ",", ""), 64)
		var bytesOp float64
		if len(parts) >= 7 {
			bytesOp, _ = strconv.ParseFloat(strings.ReplaceAll(parts[6], ",", ""), 64)
		}

		if _, ok := benchData[name]; !ok {
			benchData[name] = &benchMetrics{}
		}
		benchData[name].nsOp += nsOp
		benchData[name].allocs += allocs
		benchData[name].bytes += bytesOp
		benchCounts[name]++
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("go test: %w", err)
	}

	for name, m := range benchData {
		count := benchCounts[name]
		if count == 0 {
			continue
		}
		results.Results = append(results.Results, Result{
			Name:        name,
			NsPerOp:     m.nsOp / float64(count),
			AllocsPerOp: m.allocs / float64(count),
			BytesPerOp:  m.bytes / float64(count),
		})
	}

	return results, nil
}
