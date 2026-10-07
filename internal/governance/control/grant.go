// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package control

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type grantIdentity struct {
	Request       Request  `json:"request"`
	Decision      Decision `json:"decision"`
	PolicyID      string   `json:"policy_id"`
	PolicyVersion string   `json:"policy_version"`
	RuleID        string   `json:"rule_id"`
}

// bindGrantDigest creates a deterministic integrity binding for an
// authorization decision. Mutable evidence such as reasons and timestamps
// are intentionally excluded; the digest represents the authority granted,
// not its audit presentation.
func bindGrantDigest(rec *DecisionRecord) error {
	identity := grantIdentity{
		Request:       rec.Request,
		Decision:      rec.Decision,
		PolicyID:      rec.PolicyID,
		PolicyVersion: rec.PolicyVersion,
		RuleID:        rec.RuleID,
	}
	payload, err := json.Marshal(identity)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(payload)
	rec.GrantDigest = hex.EncodeToString(sum[:])
	return nil
}

func grantDigestMatches(rec DecisionRecord) bool {
	expected := rec
	expected.GrantDigest = ""
	return grantDigestFor(expected) == rec.GrantDigest
}

func grantDigestFor(rec DecisionRecord) string {
	identity := grantIdentity{
		Request:       rec.Request,
		Decision:      rec.Decision,
		PolicyID:      rec.PolicyID,
		PolicyVersion: rec.PolicyVersion,
		RuleID:        rec.RuleID,
	}
	payload, err := json.Marshal(identity)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
