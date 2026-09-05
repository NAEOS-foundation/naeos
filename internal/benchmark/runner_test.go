package benchmark

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseBenchLine(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name   string
		line   string
		want   string // expected benchmark name; "" means expect ok=false
		em     benchMetrics
		ok     bool
		reason string
	}{
		{
			name: "standard line",
			line: `{"Action":"output","Output":"BenchmarkParser-8           123456             100.5 ns/op         128 B/op          4 allocs/op\n"}`,
			want: "BenchmarkParser-8",
			em:   benchMetrics{nsOp: 100.5, bytes: 128, allocs: 4},
			ok:   true,
		},
		{
			name: "comma-formatted numbers",
			line: `{"Action":"output","Output":"BenchmarkCompile-16        10                 1,234.5 ns/op       512 B/op          8 allocs/op\n"}`,
			want: "BenchmarkCompile-16",
			em:   benchMetrics{nsOp: 1234.5, bytes: 512, allocs: 8},
			ok:   true,
		},
		{
			name: "sub-benchmark collapsed to parent",
			line: `{"Action":"output","Output":"BenchmarkParser/sub-4     50000             2000 ns/op         256 B/op          2 allocs/op\n"}`,
			want: "BenchmarkParser",
			em:   benchMetrics{nsOp: 2000, bytes: 256, allocs: 2},
			ok:   true,
		},
		{
			name: "bytes only no allocs column",
			line: `{"Action":"output","Output":"BenchmarkFoo-8            100000             50 ns/op         64 B/op\n"}`,
			want: "BenchmarkFoo-8",
			em:   benchMetrics{nsOp: 50, bytes: 64},
			ok:   true,
		},
		{
			name: "non-benchmark output",
			line: `{"Action":"output","Output":"some log line\n"}`,
			ok:   false,
		},
		{
			name: "non-output action",
			line: `{"Action":"ok","Package":"pkg"}`,
			ok:   false,
		},
		{
			name:   "malformed json",
			line:   `this is not json`,
			ok:     false,
			reason: "malformed",
		},
		{
			name: "too few tokens",
			line: `{"Action":"output","Output":"BenchmarkAbc-8 12345\n"}`,
			ok:   false,
		},
		{
			name: "valid json wrong shape",
			line: `{"foo":"bar"}`,
			ok:   false,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			name, m, ok := parseBenchLine(tc.line)
			if ok != tc.ok {
				t.Fatalf("ok=%v, want %v (%s)", ok, tc.ok, tc.reason)
			}
			if !tc.ok {
				return
			}
			if name != tc.want {
				t.Errorf("name=%q, want %q", name, tc.want)
			}
			if m != tc.em {
				t.Errorf("metrics=%+v, want %+v", m, tc.em)
			}
		})
	}
}

func TestParseBenchLineLeadingSpaces(t *testing.T) {
	t.Parallel()
	line := `{"Action":"output","Output":"  BenchmarkTrim-4   9999   42.5 ns/op   512 B/op   3 allocs/op\n"}`
	name, m, ok := parseBenchLine(line)
	if ok {
		t.Fatalf("expected leading-space line to be skipped, got name=%q metrics=%+v", name, m)
	}
	if name != "" || m != (benchMetrics{}) {
		t.Errorf("expected zero return values, got name=%q metrics=%+v", name, m)
	}
}

func TestParseBenchLineEmptyOutput(t *testing.T) {
	t.Parallel()
	line := `{"Action":"output","Output":""}`
	if _, _, ok := parseBenchLine(line); ok {
		t.Error("expected ok=false for empty output")
	}
}

func TestSummarize(t *testing.T) {
	t.Parallel()

	t.Run("averages across counts", func(t *testing.T) {
		t.Parallel()
		data := map[string]*benchMetrics{
			"BenchmarkA": {nsOp: 300, allocs: 30, bytes: 3000},
		}
		counts := map[string]int{"BenchmarkA": 3}
		got := summarize(data, counts)
		if len(got) != 1 {
			t.Fatalf("expected 1 result, got %+v", got)
		}
		r := got[0]
		if r.NsPerOp != 100 || r.AllocsPerOp != 10 || r.BytesPerOp != 1000 {
			t.Errorf("unexpected average: %+v", r)
		}
	})

	t.Run("zero count skipped", func(t *testing.T) {
		t.Parallel()
		data := map[string]*benchMetrics{"BenchmarkZ": {nsOp: 5}}
		counts := map[string]int{"BenchmarkZ": 0}
		if got := summarize(data, counts); len(got) != 0 {
			t.Errorf("expected empty result, got %+v", got)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		got := summarize(map[string]*benchMetrics{}, map[string]int{})
		if len(got) != 0 {
			t.Errorf("expected empty result, got %+v", got)
		}
	})

	t.Run("multiple benchmarks both included", func(t *testing.T) {
		t.Parallel()
		data := map[string]*benchMetrics{
			"BenchmarkA": {nsOp: 10, allocs: 1, bytes: 100},
			"BenchmarkB": {nsOp: 20, allocs: 2, bytes: 200},
		}
		counts := map[string]int{"BenchmarkA": 1, "BenchmarkB": 1}
		got := summarize(data, counts)
		byName := make(map[string]Result, len(got))
		for _, r := range got {
			byName[r.Name] = r
		}
		if len(byName) != 2 {
			t.Fatalf("expected 2 distinct results, got %+v", got)
		}
		if r := byName["BenchmarkA"]; r.NsPerOp != 10 || r.AllocsPerOp != 1 || r.BytesPerOp != 100 {
			t.Errorf("BenchmarkA mismatch: %+v", r)
		}
		if r := byName["BenchmarkB"]; r.NsPerOp != 20 || r.AllocsPerOp != 2 || r.BytesPerOp != 200 {
			t.Errorf("BenchmarkB mismatch: %+v", r)
		}
	})
}

func TestParseBenchLineIntegrationOrdering(t *testing.T) {
	t.Parallel()
	lines := []string{
		`{"Action":"output","Output":"BenchmarkA-4   10   100.0 ns/op   10 B/op   1 allocs/op\n"}`,
		`{"Action":"output","Output":"BenchmarkB-4   10   200.0 ns/op   20 B/op   2 allocs/op\n"}`,
		`{"Action":"output","Output":"BenchmarkA-4   10   300.0 ns/op   30 B/op   3 allocs/op\n"}`,
		`{"Action":"output","Output":"BenchmarkA-4   10   500.0 ns/op   50 B/op   5 allocs/op\n"}`,
		`{"Action":"fail","Package":"bg"}`,
		`garbage line`,
	}
	scanner := bufio.NewScanner(strings.NewReader(strings.Join(lines, "\n")))
	benchData := make(map[string]*benchMetrics)
	benchCounts := make(map[string]int)
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

	got := summarize(benchData, benchCounts)
	if len(got) != 2 {
		t.Fatalf("expected 2 benchmarks, got %+v", got)
	}
	want := map[string]Result{
		"BenchmarkA-4": {NsPerOp: 300, AllocsPerOp: 3, BytesPerOp: 30},
		"BenchmarkB-4": {NsPerOp: 200, AllocsPerOp: 2, BytesPerOp: 20},
	}
	for _, r := range got {
		w := want[r.Name]
		if r.NsPerOp != w.NsPerOp || r.AllocsPerOp != w.AllocsPerOp || r.BytesPerOp != w.BytesPerOp {
			t.Errorf("benchmark %s summary mismatch: got %+v want %+v", r.Name, r, w)
		}
	}
}
