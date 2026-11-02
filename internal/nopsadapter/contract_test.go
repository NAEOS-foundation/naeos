// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package nopsadapter

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func testRequest() Request {
	content := "package demo\n"
	sum := sha256.Sum256([]byte(content))
	return Request{
		Contract: ContractName, ContractVersion: ContractVersion,
		RequestID: "req-001", CreatedAt: "2026-11-02T08:00:00Z",
		Capability: CapabilityCodeAudit,
		Source: Source{
			Repository: "NAEOS-foundation/Nops",
			Revision:   strings.Repeat("a", 40),
			Files: []SourceFile{{
				Path: "pkg/demo.go", SHA256: hex.EncodeToString(sum[:]), Content: content,
			}},
		},
	}
}

func TestValidateRequestAcceptsBoundedImmutableSnapshot(t *testing.T) {
	if err := ValidateRequest(testRequest(), DefaultLimits()); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
}

func TestValidateRequestRejectsUnsupportedCapability(t *testing.T) {
	request := testRequest()
	request.Capability = "repository.write"
	if err := ValidateRequest(request, DefaultLimits()); err == nil {
		t.Fatal("expected unsupported capability to fail closed")
	}
}

func TestValidateRequestRejectsPathTraversalAndAbsolutePaths(t *testing.T) {
	for _, candidate := range []string{"../secret.env", "/etc/passwd", "pkg/../secret.go", "pkg\\secret.go"} {
		t.Run(candidate, func(t *testing.T) {
			request := testRequest()
			request.Source.Files[0].Path = candidate
			if err := ValidateRequest(request, DefaultLimits()); err == nil {
				t.Fatalf("expected unsafe path %q to be rejected", candidate)
			}
		})
	}
}

func TestValidateRequestRejectsDigestMismatchAndDuplicatePaths(t *testing.T) {
	t.Run("digest mismatch", func(t *testing.T) {
		request := testRequest()
		request.Source.Files[0].SHA256 = strings.Repeat("0", 64)
		if err := ValidateRequest(request, DefaultLimits()); err == nil {
			t.Fatal("expected digest mismatch to fail")
		}
	})
	t.Run("duplicate path", func(t *testing.T) {
		request := testRequest()
		request.Source.Files = append(request.Source.Files, request.Source.Files[0])
		if err := ValidateRequest(request, DefaultLimits()); err == nil {
			t.Fatal("expected duplicate path to fail")
		}
	})
}

func TestValidateRequestEnforcesFileAndPayloadLimits(t *testing.T) {
	request := testRequest()
	limits := DefaultLimits()
	limits.MaxFileBytes = 2
	if err := ValidateRequest(request, limits); err == nil {
		t.Fatal("expected per-file size limit to fail")
	}
}

func TestDecodeRequestRejectsUnknownFieldsAndTrailingJSON(t *testing.T) {
	encoded := `{"contract":"naeos.nops-engineering-intelligence","contract_version":"0.1.0","request_id":"r","created_at":"2026-11-02T08:00:00Z","capability":"code.audit","source":{"revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","files":[{"path":"x.go","sha256":"` + strings.Repeat("0", 64) + `","content":""}]},"unexpected":true}`
	if _, err := DecodeRequest([]byte(encoded), DefaultLimits()); err == nil {
		t.Fatal("expected unknown field to be rejected")
	}
	if _, err := DecodeRequest([]byte("{} {}"), DefaultLimits()); err == nil {
		t.Fatal("expected trailing JSON value to be rejected")
	}
}

func TestValidateResponseCorrelatesRequestAndSource(t *testing.T) {
	request := testRequest()
	response := Response{
		Contract: ContractName, ContractVersion: ContractVersion,
		RequestID: request.RequestID, Capability: request.Capability,
		Status: "completed",
		Source: ResponseSource{Repository: request.Source.Repository, Revision: request.Source.Revision},
		Analysis: &ResponseAnalysis{
			Provider: "openai", Model: "configured-model", PromptVersion: "audit-v1",
			Findings: []Finding{{
				ID: "finding-1", Category: "security", Severity: "medium",
				Title: "Example", Description: "Evidence-backed detail",
				Location:       &Location{Path: "pkg/demo.go", StartLine: 1, EndLine: 1},
				Recommendation: "Review the source", Confidence: "medium",
			}},
		},
		ResultDigest: "sha256:" + strings.Repeat("b", 64),
	}
	if err := ValidateResponse(response, request); err != nil {
		t.Fatalf("valid response rejected: %v", err)
	}

	response.Source.Revision = strings.Repeat("c", 40)
	if err := ValidateResponse(response, request); err == nil {
		t.Fatal("expected stale source revision to be rejected")
	}
	response.Source.Revision = request.Source.Revision
	response.Analysis.Findings[0].Location.EndLine = 2
	if err := ValidateResponse(response, request); err == nil {
		t.Fatal("expected line range beyond supplied file to be rejected")
	}
}

func TestValidateResponseRejectsUnboundFindingsAndMissingFindings(t *testing.T) {
	request := testRequest()
	response := Response{
		Contract: ContractName, ContractVersion: ContractVersion,
		RequestID: request.RequestID, Capability: request.Capability,
		Status: "completed",
		Source: ResponseSource{Repository: request.Source.Repository, Revision: request.Source.Revision},
		Analysis: &ResponseAnalysis{
			Provider: "openai", Model: "model", PromptVersion: "audit-v1",
			Findings: []Finding{{
				ID: "finding-1", Severity: "high", Confidence: "high",
				Title: "Out-of-snapshot finding", Description: "Not bound to input",
				Location:       &Location{Path: "secrets.env", StartLine: 1, EndLine: 1},
				Recommendation: "Do not accept",
			}},
		},
		ResultDigest: "sha256:" + strings.Repeat("b", 64),
	}
	if err := ValidateResponse(response, request); err == nil {
		t.Fatal("expected finding outside source snapshot to be rejected")
	}

	response.Analysis.Findings = nil
	if err := ValidateResponse(response, request); err == nil {
		t.Fatal("expected missing findings collection to be rejected")
	}
}
