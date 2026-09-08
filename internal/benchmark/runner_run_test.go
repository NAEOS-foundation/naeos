package benchmark

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// envHelperProcess marks a re-invocation of this test binary as the fake
// "go test" helper that replays recorded JSON benchmark output.
const envHelperProcess = "NAEOS_BENCH_HELPER_PROCESS"

func TestHelperProcess(t *testing.T) {
	if os.Getenv(envHelperProcess) != "1" {
		return
	}
	if st := os.Getenv("FAKE_STDERR"); st != "" {
		_, _ = os.Stderr.WriteString(st)
	}
	if stdout := os.Getenv("FAKE_STDOUT"); stdout != "" {
		_, _ = os.Stdout.WriteString(stdout)
	}
	os.Exit(0)
}

// fakeRunBenchmarks swaps the package-level runCommand to invoke the helper
// process (instead of the real "go test -bench"), then returns a cleanup.
func fakeRunBenchmarks(stdout, stderr string) func() {
	orig := runCommand
	runCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess")
		cmd.Env = append(os.Environ(),
			envHelperProcess+"=1",
			"FAKE_STDOUT="+stdout,
			"FAKE_STDERR="+stderr,
		)
		return cmd
	}
	return func() { runCommand = orig }
}

func TestRunBenchmarksParsesOutput(t *testing.T) {
	stdout := `{"Action":"output","Output":"BenchmarkParser-8   100   10.0 ns/op   20 B/op   2 allocs/op\n"}` + "\n" +
		`{"Action":"output","Output":"BenchmarkParser-8   100   30.0 ns/op   40 B/op   4 allocs/op\n"}` + "\n" +
		`{"Action":"ok","Package":"pkg"}` + "\n"
	defer fakeRunBenchmarks(stdout, "")()

	var buf bytes.Buffer
	res, err := RunBenchmarks("some/pkg", &buf)
	if err != nil {
		t.Fatalf("RunBenchmarks: %v", err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("expected 1 result, got %+v", res.Results)
	}
	r := res.Results[0]
	if r.Name != "BenchmarkParser-8" {
		t.Errorf("name=%q, want BenchmarkParser-8", r.Name)
	}
	if r.NsPerOp != 20 {
		t.Errorf("nsPerOp=%v, want 20 (avg over 2)", r.NsPerOp)
	}
	if r.BytesPerOp != 30 {
		t.Errorf("bytesPerOp=%v, want 30", r.BytesPerOp)
	}
	if r.AllocsPerOp != 3 {
		t.Errorf("allocsPerOp=%v, want 3", r.AllocsPerOp)
	}
}

func TestRunBenchmarksStderrStreamedToOutput(t *testing.T) {
	defer fakeRunBenchmarks(`{"Action":"ok","Package":"pkg"}`+"\n", "warning: some stderr\n")()

	var buf bytes.Buffer
	if _, err := RunBenchmarks("some/pkg", &buf); err != nil {
		t.Fatalf("RunBenchmarks: %v", err)
	}
	if !strings.Contains(buf.String(), "warning: some stderr") {
		t.Errorf("expected stderr copied to output, got %q", buf.String())
	}
}

func TestRunBenchmarksIgnoresNonBenchmarks(t *testing.T) {
	stdout := `{"Action":"output","Output":"normal log\n"}` + "\n" +
		`{"Action":"output","Output":"BenchmarkOnlyOne-2   10   5.0 ns/op   1 B/op   1 allocs/op\n"}` + "\n" +
		`{"Action":"ok","Package":"pkg"}` + "\n"
	defer fakeRunBenchmarks(stdout, "")()

	res, err := RunBenchmarks("some/pkg", new(bytes.Buffer))
	if err != nil {
		t.Fatalf("RunBenchmarks: %v", err)
	}
	if len(res.Results) != 1 || res.Results[0].Name != "BenchmarkOnlyOne-2" {
		t.Errorf("expected only real benchmark, got %+v", res.Results)
	}
}

func TestRunBenchmarksSubBenchmarkGroupedToParent(t *testing.T) {
	stdout := `{"Action":"output","Output":"BenchmarkParent/child-4   5   50.0 ns/op   5 B/op   5 allocs/op\n"}` + "\n" +
		`{"Action":"ok","Package":"pkg"}` + "\n"
	defer fakeRunBenchmarks(stdout, "")()

	res, err := RunBenchmarks("some/pkg", new(bytes.Buffer))
	if err != nil {
		t.Fatalf("RunBenchmarks: %v", err)
	}
	if len(res.Results) != 1 || res.Results[0].Name != "BenchmarkParent" {
		t.Fatalf("expected parent name, got %+v", res.Results)
	}
}
