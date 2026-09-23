// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

func TestPipelineH4SecurityContextTrace(t *testing.T) {
	const spec = `project: security-trace
modules:
  - name: core
    path: ./internal/core
security:
  tls: "1.3"
services:
  - name: api
    kind: http
    port: 8080
`

	p, err := New(Config{
		Policies: []policy.Rule{
			{RuleID: "must-have-tls", Condition: "exists:security", Action: "block", Enabled: true},
		},
	})
	if err != nil {
		t.Fatalf("create pipeline failed: %v", err)
	}

	// Stage 1-5: exercise the real parser -> normalizer -> resolver -> builder
	// -> policy-context path through Pipeline.Run.
	result, err := p.Run(spec)
	if err != nil {
		t.Fatalf("security context was lost before policy evaluation: %v", err)
	}
	if result == nil || result.NEIR == nil {
		t.Fatal("expected non-nil pipeline result and NEIR")
	}
	if result.NEIR.Security == nil {
		t.Fatal("trace: NEIR.Security is nil")
	}
	if result.NEIR.Security.Attributes["tls"] != "1.3" {
		t.Fatalf("trace: NEIR.Security.Attributes[tls] = %q, want 1.3", result.NEIR.Security.Attributes["tls"])
	}
	if result.PolicyContextVersion == "" {
		t.Fatal("trace: policy context version was not recorded")
	}
	if result.PolicyContextDigest == "" {
		t.Fatal("trace: policy context digest was not recorded")
	}
	if result.GovernanceStatus != "evaluated" {
		t.Fatalf("trace: governance status = %q, want evaluated", result.GovernanceStatus)
	}
	if len(result.PolicyResults) != 1 {
		t.Fatalf("trace: policy results = %d, want 1", len(result.PolicyResults))
	}
	if !result.PolicyResults[0].Passed {
		t.Fatalf("trace: policy result failed: %#v", result.PolicyResults[0])
	}
	if strings.Contains(result.PolicyResults[0].Message, "key security not found in context") {
		t.Fatal("trace: evaluator still reports security missing from policy context")
	}
}
