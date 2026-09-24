// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/governance/changerisk"
)

type policyFile struct {
	PolicyID      string `json:"policy_id"`
	PolicyVersion string `json:"policy_version"`
	SchemaVersion string `json:"schema_version"`
}

type evidenceFile struct {
	PolicyID      string   `json:"policy_id"`
	PolicyVersion string   `json:"policy_version"`
	BaseSHA       string   `json:"base_sha"`
	HeadSHA       string   `json:"head_sha"`
	Result        changerisk.Result `json:"result"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	policyBytes, err := os.ReadFile("governance/change-risk-policy.json")
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	var policy policyFile
	if err := json.Unmarshal(policyBytes, &policy); err != nil {
		return fmt.Errorf("parse policy: %w", err)
	}
	if policy.PolicyID == "" || policy.PolicyVersion != "1.0.0" || policy.SchemaVersion != "1.0.0" {
		return errors.New("unsupported change-risk policy version")
	}

	base := strings.TrimSpace(os.Getenv("BASE_SHA"))
	if base == "" {
		return errors.New("BASE_SHA is required for change risk governance")
	}
	head := strings.TrimSpace(os.Getenv("HEAD_SHA"))
	if head == "" {
		head = "HEAD"
	}
	if err := verifyCommit(base); err != nil {
		return fmt.Errorf("invalid base commit: %w", err)
	}
	if err := verifyCommit(head); err != nil {
		return fmt.Errorf("invalid head commit: %w", err)
	}

	paths, err := gitOutput("diff", "--name-only", base+"..."+head)
	if err != nil {
		return fmt.Errorf("collect changed paths: %w", err)
	}
	additions, deletions, err := diffStats(base, head)
	if err != nil {
		return fmt.Errorf("collect diff stats: %w", err)
	}

	pathList := splitLines(paths)
	result := changerisk.Classify(changerisk.Request{
		Paths:             pathList,
		Additions:         additions,
		Deletions:         deletions,
		EvidenceAvailable: true,
	})
	evidence := evidenceFile{
		PolicyID:      policy.PolicyID,
		PolicyVersion: policy.PolicyVersion,
		BaseSHA:       base,
		HeadSHA:       head,
		Result:        result,
	}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	data = append(data, '\n')

	out := os.Getenv("NAEOS_CHANGE_RISK_OUTPUT")
	if out == "" {
		out = "change-risk-evidence.json"
	}
	out, err = safeRelativePath(out)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	if err := os.WriteFile(out, data, 0o600); err != nil { //nolint:gosec // output is constrained to a workspace-relative path
		return fmt.Errorf("write evidence: %w", err)
	}

	fmt.Printf("change risk: decision=%s risk=%s criticality=%s files=%d additions=%d deletions=%d output=%s\n",
		result.Decision, result.Risk, result.Criticality, result.ChangedFiles, additions, deletions, out)
	if result.Decision == changerisk.Deny {
		return errors.New("change risk policy denied the change")
	}
	return nil
}

func verifyCommit(ref string) error {
	if strings.HasPrefix(ref, "-") {
		return errors.New("commit ref must not start with '-'")
	}
	_, err := gitOutput("rev-parse", "--verify", ref+"^{commit}")
	return err
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...) //nolint:gosec // refs are validated with git rev-parse before use
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func diffStats(base, head string) (int, int, error) {
	output, err := gitOutput("diff", "--numstat", base+"..."+head)
	if err != nil {
		return 0, 0, err
	}
	additions, deletions := 0, 0
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || fields[0] == "-" || fields[1] == "-" {
			continue
		}
		added, err := strconv.Atoi(fields[0])
		if err != nil {
			return 0, 0, fmt.Errorf("parse additions: %w", err)
		}
		deleted, err := strconv.Atoi(fields[1])
		if err != nil {
			return 0, 0, fmt.Errorf("parse deletions: %w", err)
		}
		additions += added
		deletions += deleted
	}
	return additions, deletions, scanner.Err()
}

func splitLines(value string) []string {
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func safeRelativePath(value string) (string, error) {
	clean := filepath.Clean(value)
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("path must be relative and remain within the workspace")
	}
	return clean, nil
}
