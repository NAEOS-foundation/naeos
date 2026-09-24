// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/governance/dependencyrisk"
)

type requestFile struct {
	PolicyVersion string   `json:"policy_version"`
	SchemaVersion string   `json:"schema_version"`
	Ecosystem     string   `json:"ecosystem"`
	Name          string   `json:"name"`
	VersionChange string   `json:"version_change"`
	Paths         []string `json:"paths"`
	Evidence      bool     `json:"evidence_available"`
}

type policyFile struct {
	PolicyID      string `json:"policy_id"`
	PolicyVersion string `json:"policy_version"`
	SchemaVersion string `json:"schema_version"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	policyBytes, err := os.ReadFile("governance/dependency-risk-policy.json")
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	var policy policyFile
	if err := json.Unmarshal(policyBytes, &policy); err != nil {
		return fmt.Errorf("parse policy: %w", err)
	}
	if policy.PolicyID == "" || policy.PolicyVersion != "1.0.0" || policy.SchemaVersion != "1.0.0" {
		return errors.New("unsupported dependency-risk policy version")
	}

	requestPath := os.Getenv("NAEOS_DEPENDENCY_RISK_REQUEST")
	if requestPath == "" {
		return errors.New("NAEOS_DEPENDENCY_RISK_REQUEST is required for dependency changes")
	}
	requestPath, err = safeRelativePath(requestPath)
	if err != nil {
		return fmt.Errorf("invalid request path: %w", err)
	}
	requestBytes, err := os.ReadFile(requestPath)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	var req requestFile
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		return fmt.Errorf("parse request: %w", err)
	}
	if req.PolicyVersion != policy.PolicyVersion || req.SchemaVersion != policy.SchemaVersion {
		return errors.New("request policy/schema version is unsupported")
	}

	result := dependencyrisk.Classify(dependencyrisk.Request{
		Ecosystem:         req.Ecosystem,
		Name:              req.Name,
		VersionChange:     dependencyrisk.VersionChange(req.VersionChange),
		Paths:             req.Paths,
		EvidenceAvailable: req.Evidence,
	})
	evidence := struct {
		PolicyID      string                `json:"policy_id"`
		PolicyVersion string                `json:"policy_version"`
		Result        dependencyrisk.Result `json:"result"`
	}{policy.PolicyID, policy.PolicyVersion, result}

	out := os.Getenv("NAEOS_DEPENDENCY_RISK_OUTPUT")
	if out == "" {
		out = "dependency-risk-evidence.json"
	}
	out, err = safeRelativePath(out)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(out, data, 0o600); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	fmt.Printf("dependency risk: decision=%s risk=%s criticality=%s output=%s\n", result.Decision, result.Risk, result.Criticality, out)
	if result.Decision == dependencyrisk.Deny {
		return errors.New("dependency risk policy denied the change")
	}
	return nil
}

func safeRelativePath(value string) (string, error) {
	clean := filepath.Clean(value)
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("path must be relative and remain within the workspace")
	}
	return clean, nil
}
