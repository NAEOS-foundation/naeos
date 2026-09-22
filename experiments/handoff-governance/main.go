// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

// Command handoff-governance tests the NAEOS handoff authorization boundary.
// Keep this executable independent of network services and LLM output.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/investordemo"
)

const (
	agentA = "agent-a"
	agentB = "agent-b"
)

type scenarioResult struct {
	Name     string   `json:"name"`
	Expected string   `json:"expected"`
	Observed string   `json:"observed"`
	Passed   bool     `json:"passed"`
	Checks   []string `json:"checks"`
	Errors   []string `json:"errors,omitempty"`
}

func main() {
	results := run()
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(struct {
		Experiment string           `json:"experiment"`
		Thesis     string           `json:"thesis"`
		Results    []scenarioResult `json:"results"`
		Summary    string           `json:"summary"`
	}{
		Experiment: "NAEOS Agent Handoff Governance v1",
		Thesis:     "A handoff transfers context, not authority.",
		Results:    results,
		Summary:    fmt.Sprintf("%d/%d scenarios passed", passed, len(results)),
	})
	if passed != len(results) {
		os.Exit(2)
	}
}

func baseContract() *investordemo.HandoffContract {
	now := time.Now().UTC()
	return &investordemo.HandoffContract{
		ContractVersion:        "1.0",
		CanonicalVersion:       "1",
		Initiator:              agentA,
		Recipient:              agentB,
		RequestedCapability:    "repository.write",
		AuthorizedCapabilities: []investordemo.Capability{"repository.read", "repository.write"},
		PolicyID:               "POLICY-HANDOFF",
		PolicyVersion:          1,
		Provenance:             map[string]interface{}{"source": agentA, "destination": agentB},
		CreatedAt:              now,
		ExpiresAt:              now.Add(10 * time.Minute),
		ReplayProtection:       investordemo.ReplayProtection{Nonce: "handoff-v1-001", Timestamp: now},
	}
}

func newValidator() *investordemo.HandoffValidator {
	return investordemo.SetupDemoEnvironment().HandoffValidator
}

func sign(v *investordemo.HandoffValidator, c *investordemo.HandoffContract) {
	c.Signature = v.SignContract(c)
}

func result(name, expected, observed string, passed bool, checks []string, errors []string) scenarioResult {
	return scenarioResult{Name: name, Expected: expected, Observed: observed, Passed: passed, Checks: checks, Errors: errors}
}

func run() []scenarioResult {
	results := make([]scenarioResult, 0, 10)

	// 1. Valid signed contract.
	{
		v := newValidator()
		c := baseContract()
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("01-valid-signed-handoff",
			"valid contract accepted",
			fmt.Sprintf("valid=%v errors=%d", r.Valid, len(r.Errors)),
			r.Valid,
			[]string{"contract version accepted", "canonicalization version accepted", "signature verified", "capability set contains request", "provenance matches"},
			r.Errors))
	}

	// 2. Capability widening.
	{
		v := newValidator()
		c := baseContract()
		c.RequestedCapability = "credential.rotate"
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("02-capability-widening",
			"request outside authorized capabilities is rejected",
			fmt.Sprintf("valid=%v widening=%v", r.Valid, r.CapabilityWideningDetected),
			!r.Valid && r.CapabilityWideningDetected,
			[]string{"requested capability checked against authorized set"}, r.Errors))
	}

	// 3. Downstream escalation.
	{
		v := newValidator()
		c := baseContract()
		c.DownstreamHandoff = &investordemo.HandoffContract{
			ContractVersion: "1.0", CanonicalVersion: "1", Initiator: agentB,
			RequestedCapability: "credential.rotate",
		}
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("03-downstream-escalation",
			"downstream request outside parent authority is rejected",
			fmt.Sprintf("valid=%v widening=%v", r.Valid, r.CapabilityWideningDetected),
			!r.Valid && r.CapabilityWideningDetected,
			[]string{"downstream request compared with parent authorized capabilities"}, r.Errors))
	}

	// 4. Replay.
	{
		v := newValidator()
		c1 := baseContract()
		sign(v, c1)
		first := v.ValidateHandoff(c1)
		c2 := baseContract()
		sign(v, c2)
		second := v.ValidateHandoff(c2)
		results = append(results, result("04-replay",
			"same nonce is accepted once and rejected on replay",
			fmt.Sprintf("first=%v second=%v replay=%v", first.Valid, second.Valid, second.ReplayDetected),
			first.Valid && !second.Valid && second.ReplayDetected,
			[]string{"nonce recorded at first validation", "same nonce detected on second validation"}, second.Errors))
	}

	// 5. Payload tampering.
	{
		v := newValidator()
		c := baseContract()
		payload := map[string]interface{}{"change": "update-service"}
		c.Payload = payload
		c.PayloadDigest = investordemo.CalculatePayloadDigestForExperiment(payload)
		c.Payload["change"] = "tampered"
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("05-payload-tampering",
			"payload digest mismatch is rejected",
			fmt.Sprintf("valid=%v tampered=%v", r.Valid, r.PayloadTampered),
			!r.Valid && r.PayloadTampered,
			[]string{"payload re-hashed before acceptance"}, r.Errors))
	}

	// 6. Provenance mismatch.
	{
		v := newValidator()
		c := baseContract()
		c.Provenance["source"] = "untrusted-agent"
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("06-provenance-mismatch",
			"source mismatch is rejected",
			fmt.Sprintf("valid=%v mismatch=%v", r.Valid, r.ProvenanceMismatch),
			!r.Valid && r.ProvenanceMismatch,
			[]string{"provenance source compared with initiator"}, r.Errors))
	}

	// 7. Contract version mismatch.
	{
		v := newValidator()
		c := baseContract()
		c.ContractVersion = "2.0"
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("07-contract-version-mismatch",
			"unsupported contract version fails closed",
			fmt.Sprintf("valid=%v errors=%d", r.Valid, len(r.Errors)),
			!r.Valid,
			[]string{"contract version is an explicit trust boundary"}, r.Errors))
	}

	// 8. Canonicalization mismatch.
	{
		v := newValidator()
		c := baseContract()
		c.CanonicalVersion = "99"
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("08-canonicalization-mismatch",
			"unsupported canonicalization version fails closed",
			fmt.Sprintf("valid=%v errors=%d", r.Valid, len(r.Errors)),
			!r.Valid,
			[]string{"canonicalization version is an explicit trust boundary"}, r.Errors))
	}

	// 9. Expiration.
	{
		v := newValidator()
		c := baseContract()
		c.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		sign(v, c)
		r := v.ValidateHandoff(c)
		results = append(results, result("09-expired-handoff",
			"expired authorization is rejected",
			fmt.Sprintf("valid=%v expired=%v", r.Valid, r.ExpiredDetected),
			!r.Valid && r.ExpiredDetected,
			[]string{"expiration checked before acceptance"}, r.Errors))
	}

	// 10. Signature tampering.
	{
		v := newValidator()
		c := baseContract()
		sign(v, c)
		c.Signature = "tampered-signature"
		r := v.ValidateHandoff(c)
		results = append(results, result("10-signature-tampering",
			"modified signature is rejected",
			fmt.Sprintf("valid=%v errors=%d", r.Valid, len(r.Errors)),
			!r.Valid,
			[]string{"signature binds the contract fields"}, r.Errors))
	}

	return results
}
