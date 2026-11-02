// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

// Package nopsadapter validates the read-only NOPS engineering-intelligence
// exchange contract. It deliberately performs no network or execution work.
package nopsadapter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	ContractName    = "naeos.nops-engineering-intelligence"
	ContractVersion = "0.1.0"

	CapabilityCodeAudit       = "code.audit"
	CapabilityDeploymentCheck = "deployment.analyze"
	CapabilityRefactorProposal = "refactor.propose"

	DefaultMaxFiles       = 64
	DefaultMaxFileBytes   = 1 << 20
	DefaultMaxPayloadBytes = 4 << 20
	DefaultMaxPathBytes   = 512
)

var (
	commitDigestPattern = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
	fileDigestPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	resultDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

// Request is the versioned NOPS request envelope. Source content is supplied
// by NAEOS; NOPS must not resolve or clone arbitrary repositories.
type Request struct {
	Contract        string    `json:"contract"`
	ContractVersion string    `json:"contract_version"`
	RequestID       string    `json:"request_id"`
	CreatedAt       string    `json:"created_at"`
	Capability      string    `json:"capability"`
	Source          Source    `json:"source"`
	Analysis        *Analysis `json:"analysis,omitempty"`
}

type Source struct {
	Repository string      `json:"repository,omitempty"`
	Revision   string      `json:"revision"`
	Files      []SourceFile `json:"files"`
}

type SourceFile struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Content string `json:"content"`
}

type Analysis struct {
	Categories     []string `json:"categories,omitempty"`
	ProviderPolicy string   `json:"provider_policy,omitempty"`
}

// Limits bounds untrusted source snapshots before any future transport call.
type Limits struct {
	MaxFiles        int
	MaxFileBytes    int
	MaxPayloadBytes int
	MaxPathBytes    int
}

func DefaultLimits() Limits {
	return Limits{
		MaxFiles: DefaultMaxFiles, MaxFileBytes: DefaultMaxFileBytes,
		MaxPayloadBytes: DefaultMaxPayloadBytes, MaxPathBytes: DefaultMaxPathBytes,
	}
}

// DecodeRequest rejects unknown fields, malformed JSON, and trailing values.
func DecodeRequest(data []byte, limits Limits) (Request, error) {
	if limits.MaxPayloadBytes <= 0 {
		limits = DefaultLimits()
	}
	if len(data) > limits.MaxPayloadBytes {
		return Request{}, fmt.Errorf("request exceeds maximum payload size")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var request Request
	if err := decoder.Decode(&request); err != nil {
		return Request{}, fmt.Errorf("decode NOPS request: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Request{}, fmt.Errorf("request contains trailing JSON data")
	}
	if err := ValidateRequest(request, limits); err != nil {
		return Request{}, err
	}
	return request, nil
}

// ValidateRequest verifies envelope identity, capability, immutable revision,
// normalized paths, payload bounds, and each supplied file's SHA-256 digest.
func ValidateRequest(request Request, limits Limits) error {
	if limits.MaxFiles <= 0 || limits.MaxFileBytes <= 0 ||
		limits.MaxPayloadBytes <= 0 || limits.MaxPathBytes <= 0 {
		limits = DefaultLimits()
	}
	if request.Contract != ContractName {
		return fmt.Errorf("unsupported contract %q", request.Contract)
	}
	if request.ContractVersion != ContractVersion {
		return fmt.Errorf("unsupported contract version %q", request.ContractVersion)
	}
	if strings.TrimSpace(request.RequestID) == "" || len(request.RequestID) > 128 {
		return fmt.Errorf("request_id is required and must not exceed 128 bytes")
	}
	if _, err := time.Parse(time.RFC3339, request.CreatedAt); err != nil {
		return fmt.Errorf("created_at must be an RFC3339 timestamp")
	}
	switch request.Capability {
	case CapabilityCodeAudit, CapabilityDeploymentCheck, CapabilityRefactorProposal:
	default:
		return fmt.Errorf("unsupported capability %q", request.Capability)
	}
	if !commitDigestPattern.MatchString(request.Source.Revision) {
		return fmt.Errorf("source.revision must be an immutable 40- or 64-character git commit digest")
	}
	if len(request.Source.Files) == 0 {
		return fmt.Errorf("source.files must contain at least one file")
	}
	if len(request.Source.Files) > limits.MaxFiles {
		return fmt.Errorf("source.files exceeds maximum file count")
	}

	seen := make(map[string]struct{}, len(request.Source.Files))
	totalBytes := 0
	for i, file := range request.Source.Files {
		if file.Path == "" || len(file.Path) > limits.MaxPathBytes ||
			strings.Contains(file.Path, "\\") || strings.ContainsRune(file.Path, 0) ||
			path.IsAbs(file.Path) || path.Clean(file.Path) != file.Path ||
			file.Path == "." || file.Path == ".." || strings.HasPrefix(file.Path, "../") {
			return fmt.Errorf("source.files[%d].path must be a normalized repository-relative path", i)
		}
		if _, exists := seen[file.Path]; exists {
			return fmt.Errorf("source.files contains duplicate path %q", file.Path)
		}
		seen[file.Path] = struct{}{}

		contentBytes := len([]byte(file.Content))
		if contentBytes > limits.MaxFileBytes {
			return fmt.Errorf("source file %q exceeds maximum file size", file.Path)
		}
		totalBytes += contentBytes
		if totalBytes > limits.MaxPayloadBytes {
			return fmt.Errorf("source files exceed maximum aggregate size")
		}
		if !fileDigestPattern.MatchString(file.SHA256) {
			return fmt.Errorf("source file %q has an invalid SHA-256 digest", file.Path)
		}
		sum := sha256.Sum256([]byte(file.Content))
		if hex.EncodeToString(sum[:]) != file.SHA256 {
			return fmt.Errorf("source file %q digest does not match supplied content", file.Path)
		}
	}
	return nil
}

// Finding is an advisory, source-bound NOPS finding.
type Finding struct {
	ID             string   `json:"id"`
	Category       string   `json:"category"`
	Severity       string   `json:"severity"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Location       *Location `json:"location,omitempty"`
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
}

type Location struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type Response struct {
	Contract        string           `json:"contract"`
	ContractVersion string           `json:"contract_version"`
	RequestID       string           `json:"request_id"`
	Capability      string           `json:"capability"`
	Status          string           `json:"status"`
	Source          ResponseSource   `json:"source"`
	Analysis        *ResponseAnalysis `json:"analysis,omitempty"`
	ResultDigest    string           `json:"result_digest"`
}

type ResponseSource struct {
	Repository string `json:"repository,omitempty"`
	Revision   string `json:"revision"`
}

type ResponseAnalysis struct {
	Provider      string    `json:"provider"`
	Model         string    `json:"model"`
	PromptVersion string    `json:"prompt_version"`
	Findings      []Finding `json:"findings"`
}

// DecodeResponse performs strict JSON decoding before response validation.
func DecodeResponse(data []byte, request Request, limits Limits) (Response, error) {
	if limits.MaxPayloadBytes <= 0 {
		limits = DefaultLimits()
	}
	if len(data) > limits.MaxPayloadBytes {
		return Response{}, fmt.Errorf("response exceeds maximum payload size")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var response Response
	if err := decoder.Decode(&response); err != nil {
		return Response{}, fmt.Errorf("decode NOPS response: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Response{}, fmt.Errorf("response contains trailing JSON data")
	}
	if err := ValidateResponse(response, request); err != nil {
		return Response{}, err
	}
	return response, nil
}

// ValidateResponse binds advisory output to the exact request and source
// revision. A completed response is never interpreted as policy approval.
func ValidateResponse(response Response, request Request) error {
	if response.Contract != ContractName || response.ContractVersion != ContractVersion {
		return fmt.Errorf("unsupported NOPS response contract or version")
	}
	if response.RequestID == "" || response.RequestID != request.RequestID {
		return fmt.Errorf("NOPS response request_id does not match request")
	}
	if response.Capability != request.Capability {
		return fmt.Errorf("NOPS response capability does not match request")
	}
	if response.Source.Revision != request.Source.Revision {
		return fmt.Errorf("NOPS response source revision does not match request")
	}
	if request.Source.Repository != "" && response.Source.Repository != request.Source.Repository {
		return fmt.Errorf("NOPS response repository does not match request")
	}
	switch response.Status {
	case "completed", "failed", "rejected":
	default:
		return fmt.Errorf("unsupported NOPS response status %q", response.Status)
	}
	if !resultDigestPattern.MatchString(response.ResultDigest) {
		return fmt.Errorf("NOPS response has an invalid result_digest")
	}
	if response.Status != "completed" {
		if response.Analysis != nil {
			return fmt.Errorf("non-completed NOPS response must not contain analysis")
		}
		return nil
	}
	if response.Analysis == nil || response.Analysis.Findings == nil {
		return fmt.Errorf("completed NOPS response must include an analysis and findings collection")
	}
	if strings.TrimSpace(response.Analysis.Provider) == "" ||
		strings.TrimSpace(response.Analysis.Model) == "" ||
		strings.TrimSpace(response.Analysis.PromptVersion) == "" {
		return fmt.Errorf("completed NOPS response is missing provider provenance")
	}

	for i, finding := range response.Analysis.Findings {
		if finding.ID == "" || finding.Title == "" || finding.Description == "" ||
			finding.Recommendation == "" {
			return fmt.Errorf("finding %d is missing required descriptive fields", i)
		}
		switch finding.Severity {
		case "critical", "high", "medium", "low", "info":
		default:
			return fmt.Errorf("finding %q has unsupported severity", finding.ID)
		}
		switch finding.Confidence {
		case "high", "medium", "low":
		default:
			return fmt.Errorf("finding %q has unsupported confidence", finding.ID)
		}
		if finding.Location != nil {
			var sourceFile *SourceFile
			for fileIndex := range request.Source.Files {
				if request.Source.Files[fileIndex].Path == finding.Location.Path {
					sourceFile = &request.Source.Files[fileIndex]
					break
				}
			}
			if sourceFile == nil {
				return fmt.Errorf("finding %q references a file outside the supplied snapshot", finding.ID)
			}
			lineCount := strings.Count(sourceFile.Content, "\n")
			if len(sourceFile.Content) > 0 && !strings.HasSuffix(sourceFile.Content, "\n") {
				lineCount++
			}
			if finding.Location.StartLine < 1 || finding.Location.EndLine < finding.Location.StartLine ||
				finding.Location.EndLine > lineCount {
				return fmt.Errorf("finding %q has a line range outside the supplied source file", finding.ID)
			}
		}
	}
	return nil
}
