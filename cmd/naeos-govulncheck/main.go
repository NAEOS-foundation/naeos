// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type traceFrame struct {
	Module   string `json:"module"`
	Version  string `json:"version"`
	Package  string `json:"package"`
	Function string `json:"function"`
	Receiver string `json:"receiver"`
}

type govulnFinding struct {
	OSV   string       `json:"osv"`
	Trace []traceFrame `json:"trace"`
}

type govulnEvent struct {
	Config  json.RawMessage `json:"config"`
	Finding *govulnFinding  `json:"finding"`
}

type findingSet map[string]govulnFinding

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get repository root: %w", err)
	}

	baseSHA, err := baseRevision(ctx, root)
	if err != nil {
		return err
	}
	baseline := make(findingSet)
	if baseSHA != "" {
		baseDir, cleanup, err := addBaseWorktree(ctx, root, baseSHA)
		if err != nil {
			return err
		}
		defer cleanup()
		baseline, err = scan(ctx, baseDir)
		if err != nil {
			return fmt.Errorf("scan base %s: %w", baseSHA, err)
		}
		fmt.Printf("Govulncheck baseline: %s\n", baseSHA)
	} else {
		fmt.Println("No baseline commit available; treating all reachable findings as new.")
	}

	head, err := scan(ctx, root)
	if err != nil {
		return fmt.Errorf("scan head: %w", err)
	}

	fmt.Printf("Reachable vulnerabilities in base: %s\n", formatIDs(baseline))
	fmt.Printf("Reachable vulnerabilities in head: %s\n", formatIDs(head))

	added := newFindings(baseline, head)
	if len(added) > 0 {
		for _, finding := range added {
			fmt.Printf("New reachable finding: %s (%s)\n", finding.OSV, describeTrace(finding.Trace))
		}
		return fmt.Errorf("govulncheck found %d new reachable finding(s)", len(added))
	}

	fmt.Println("No new reachable vulnerabilities compared with base.")
	return nil
}

func baseRevision(ctx context.Context, root string) (string, error) {
	base := strings.TrimSpace(os.Getenv("BASE_SHA"))
	if base != "" && strings.Trim(base, "0") != "" {
		return base, nil
	}

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD^")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

func addBaseWorktree(ctx context.Context, root, revision string) (string, func(), error) {
	worktree, err := os.MkdirTemp("", "naeos-govulncheck-base-")
	if err != nil {
		return "", nil, fmt.Errorf("create base worktree path: %w", err)
	}
	if err := os.Remove(worktree); err != nil {
		return "", nil, fmt.Errorf("prepare base worktree path: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "--detach", worktree, revision)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(worktree)
		return "", nil, fmt.Errorf("add base worktree: %w: %s", err, strings.TrimSpace(string(output)))
	}

	cleanup := func() {
		cmd := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", worktree)
		cmd.Dir = root
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "remove base worktree: %v: %s\n", err, strings.TrimSpace(string(output)))
		}
	}
	return worktree, cleanup, nil
}

func scan(ctx context.Context, root string) (findingSet, error) {
	module, err := modulePath(ctx, root)
	if err != nil {
		return nil, err
	}

	scanner := strings.TrimSpace(os.Getenv("GOVULNCHECK"))
	if scanner == "" {
		scanner = "govulncheck"
	}
	cmd := exec.CommandContext(ctx, scanner, "-format=json", "./...")
	cmd.Dir = root
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run govulncheck: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if stderr.Len() > 0 {
		fmt.Fprint(os.Stderr, stderr.String())
	}

	findings, err := parseFindings(&stdout, module)
	if err != nil {
		return nil, fmt.Errorf("parse govulncheck output: %w", err)
	}
	return findings, nil
}

func modulePath(ctx context.Context, root string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-m")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve Go module path: %w", err)
	}
	module := strings.TrimSpace(string(output))
	if module == "" || strings.ContainsAny(module, "\r\n") {
		return "", errors.New("resolve Go module path: expected one module")
	}
	return module, nil
}

func parseFindings(reader io.Reader, module string) (findingSet, error) {
	decoder := json.NewDecoder(reader)
	findings := make(findingSet)
	hasConfig := false
	for {
		var event govulnEvent
		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(event.Config) > 0 && !bytes.Equal(bytes.TrimSpace(event.Config), []byte("null")) {
			hasConfig = true
		}
		if event.Finding == nil || event.Finding.OSV == "" || !reachesModule(event.Finding.Trace, module) {
			continue
		}
		finding := *event.Finding
		findings[findingKey(finding)] = finding
	}
	if !hasConfig {
		return nil, errors.New("missing govulncheck config event")
	}
	return findings, nil
}

func reachesModule(trace []traceFrame, module string) bool {
	for _, frame := range trace {
		if frame.Module == module {
			return true
		}
	}
	return false
}

func findingKey(finding govulnFinding) string {
	encoded, _ := json.Marshal(finding)
	return string(encoded)
}

func newFindings(baseline, head findingSet) []govulnFinding {
	added := make([]govulnFinding, 0)
	for key, finding := range head {
		if _, exists := baseline[key]; !exists {
			added = append(added, finding)
		}
	}
	sort.Slice(added, func(i, j int) bool {
		return findingKey(added[i]) < findingKey(added[j])
	})
	return added
}

func formatIDs(findings findingSet) string {
	ids := make(map[string]struct{}, len(findings))
	for _, finding := range findings {
		ids[finding.OSV] = struct{}{}
	}
	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	if len(ordered) == 0 {
		return "none"
	}
	return strings.Join(ordered, ", ")
}

func describeTrace(trace []traceFrame) string {
	if len(trace) == 0 {
		return "trace unavailable"
	}
	source := trace[0]
	caller := trace[len(trace)-1]
	version := source.Version
	if version == "" {
		version = "unknown version"
	}
	return fmt.Sprintf("%s@%s reached %s.%s", source.Module, version, caller.Package, caller.Function)
}
