// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"math"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

type digestBindingControlPlane struct {
	seen   map[string]any
	mutate func()
}

func (c *digestBindingControlPlane) Evaluate(req control.Request) (control.DecisionRecord, error) {
	c.seen = cloneMap(req.Context)
	return control.DecisionRecord{
		Request: req, Decision: control.DecisionAllow, PolicyID: "p1",
		PolicyVersion: "v1", RuleID: "allow",
	}, nil
}

func (c *digestBindingControlPlane) ValidateDecision(req control.Request, issued control.DecisionRecord) (control.DecisionRecord, error) {
	if c.mutate != nil {
		c.mutate()
	}
	if issued.Request.Context[candidateDigestContextKeyV1] != req.Context[candidateDigestContextKeyV1] {
		return control.DecisionRecord{}, errCandidateTest("digest changed during revalidation")
	}
	return issued, nil
}

type errCandidateTest string

func (e errCandidateTest) Error() string { return string(e) }

type digestCaptureSandbox struct {
	calls int
	last  ToolRequest
}

func (s *digestCaptureSandbox) Execute(req ToolRequest) (string, error) {
	s.calls++
	s.last = req
	return "ok", nil
}

func TestCandidateDigestV1BindsNormalizedRequestToAuthorization(t *testing.T) {
	cp := &digestBindingControlPlane{}
	sb := &digestCaptureSandbox{}
	gw := New(cp, sb)
	req := ToolRequest{
		RequestID: "req-digest-01", InvocationID: "inv-digest-01",
		Tool: "filesystem", Action: "write", Resource: "filesystem",
		Actor: "agent", Capability: "filesystem.write",
		Payload: map[string]any{"path": "a.txt", "content": "hello"},
		Context: map[string]any{
			"tenant":                     "test",
			candidateDigestContextKeyV1:  "attacker-controlled",
			candidateProfileContextKeyV1: "attacker-controlled",
		},
	}
	result, err := gw.Authorize(req)
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if result.Status != "completed" || sb.calls != 1 {
		t.Fatalf("expected one execution, status=%s calls=%d", result.Status, sb.calls)
	}
	if result.CandidateDigest == "" || result.CandidateDigestProfile != CandidateDigestProfileV1 {
		t.Fatalf("candidate digest metadata missing: %+v", result)
	}
	if cp.seen[candidateDigestContextKeyV1] != result.CandidateDigest ||
		cp.seen[candidateProfileContextKeyV1] != CandidateDigestProfileV1 {
		t.Fatalf("authorization context is not bound to gateway digest: %#v", cp.seen)
	}
	if sb.last.Payload["content"] != "hello" {
		t.Fatalf("unexpected executed payload: %#v", sb.last.Payload)
	}
	if _, ok := sb.last.Context[candidateDigestContextKeyV1]; ok {
		t.Fatal("untrusted candidate digest context must not reach sandbox")
	}
	if _, ok := sb.last.Context[candidateProfileContextKeyV1]; ok {
		t.Fatal("untrusted candidate profile context must not reach sandbox")
	}
}

func TestCandidateDigestV1ExecutesIsolatedSnapshotWhenCallerMutatesNestedInput(t *testing.T) {
	req := ToolRequest{
		RequestID: "req-digest-02", InvocationID: "inv-digest-02",
		Tool: "filesystem", Action: "write", Resource: "filesystem",
		Payload: map[string]any{
			"document": map[string]any{"blocks": []any{"approved"}},
		},
	}
	cp := &digestBindingControlPlane{mutate: func() {
		document := req.Payload["document"].(map[string]any)
		document["blocks"].([]any)[0] = "mutated"
	}}
	sb := &digestCaptureSandbox{}
	gw := New(cp, sb)
	result, err := gw.Authorize(req)
	if err != nil {
		t.Fatalf("authorized snapshot should execute independently of caller mutation: %v", err)
	}
	if result.Status != "completed" || sb.calls != 1 {
		t.Fatalf("expected one execution of snapshot, status=%s calls=%d", result.Status, sb.calls)
	}
	document := sb.last.Payload["document"].(map[string]any)
	if got := document["blocks"].([]any)[0]; got != "approved" {
		t.Fatalf("sandbox did not receive the authorized snapshot: got %v", got)
	}
}

func TestCandidateDigestV1RejectsNonFiniteNumbers(t *testing.T) {
	for _, n := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := CandidateDigestV1(ToolRequest{
			Tool: "mock.transfer", Action: "execute",
			Payload: map[string]any{"amount": n},
		})
		if err == nil {
			t.Fatalf("non-finite number %v was accepted", n)
		}
	}
}

func TestCandidateDigestV1IgnoresMapInsertionOrder(t *testing.T) {
	a := ToolRequest{Tool: "mock", Action: "run", Payload: map[string]any{"z": 2, "a": 1}}
	b := ToolRequest{Tool: "mock", Action: "run", Payload: map[string]any{"a": 1, "z": 2}}
	da, err := CandidateDigestV1(a)
	if err != nil {
		t.Fatal(err)
	}
	db, err := CandidateDigestV1(b)
	if err != nil {
		t.Fatal(err)
	}
	if da != db {
		t.Fatalf("map order changed digest: %s != %s", da, db)
	}
}
