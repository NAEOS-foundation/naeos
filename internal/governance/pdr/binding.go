// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pdr

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

// EvidenceRef identifies one durable EvidenceStore record and the digest
// expected by the PDR.
type EvidenceRef struct {
	ID     string
	Digest string
}

// Record is the minimal authoritative PDR binding surface. The PDR records
// an existing policy decision; it never creates or grants authorization.
type Record struct {
	RecordVersion string
	PolicyID      string
	PolicyVersion string
	SchemaVersion string
	DecisionID    string
	Outcome       string
	Evidence      []EvidenceRef
}

// VerifyBinding independently verifies the PDR against existing evidence and
// the control-plane decision ledger. It fails closed on missing, stale, or
// tampered bindings.
func VerifyBinding(record Record, store *evidence.EvidenceStore, ledger *controlplane.Ledger) error {
	if record.RecordVersion == "" || record.PolicyID == "" || record.PolicyVersion == "" ||
		record.SchemaVersion == "" || record.DecisionID == "" {
		return fmt.Errorf("pdr binding requires record, policy, schema, and decision identifiers")
	}
	if len(record.Evidence) == 0 {
		return fmt.Errorf("pdr binding requires evidence")
	}
	if store == nil || ledger == nil {
		return fmt.Errorf("pdr binding requires evidence store and ledger")
	}

	decision, ok := ledger.Decision(record.DecisionID)
	if !ok {
		return fmt.Errorf("policy decision %q is not present in ledger", record.DecisionID)
	}
	if decision.Metadata["policy_id"] != "" && decision.Metadata["policy_id"] != record.PolicyID {
		return fmt.Errorf("policy id mismatch: pdr=%q ledger=%q", record.PolicyID, decision.Metadata["policy_id"])
	}
	if decision.Metadata["policy_version"] != "" && decision.Metadata["policy_version"] != record.PolicyVersion {
		return fmt.Errorf("policy version mismatch: pdr=%q ledger=%q", record.PolicyVersion, decision.Metadata["policy_version"])
	}

	switch record.Outcome {
	case "allow", "verified":
		if decision.Decision != controlplane.DecisionAllow {
			return fmt.Errorf("pdr outcome %q does not match ledger decision %q", record.Outcome, decision.Decision)
		}
	case "deny":
		if decision.Decision != controlplane.DecisionDeny {
			return fmt.Errorf("pdr outcome %q does not match ledger decision %q", record.Outcome, decision.Decision)
		}
	case "require_review":
		if decision.Decision != controlplane.DecisionPending {
			return fmt.Errorf("pdr outcome %q does not match ledger decision %q", record.Outcome, decision.Decision)
		}
	default:
		return fmt.Errorf("unsupported pdr outcome %q", record.Outcome)
	}

	if _, err := store.Verify(); err != nil {
		return fmt.Errorf("evidence chain verification failed: %w", err)
	}

	for _, ref := range record.Evidence {
		if ref.ID == "" || ref.Digest == "" {
			return fmt.Errorf("malformed evidence reference")
		}
		ev := store.ByID(ref.ID)
		if ev == nil {
			return fmt.Errorf("evidence %q is not resolvable", ref.ID)
		}
		if ev.Hash != ref.Digest {
			return fmt.Errorf("evidence digest mismatch for %q", ref.ID)
		}
		if evidence.RecomputeHash(*ev) != ref.Digest {
			return fmt.Errorf("evidence integrity mismatch for %q", ref.ID)
		}
		if ev.PolicyID != "" && ev.PolicyID != record.PolicyID {
			return fmt.Errorf("evidence policy mismatch for %q", ref.ID)
		}
		if ev.PolicyVersion != "" && ev.PolicyVersion != record.PolicyVersion {
			return fmt.Errorf("evidence policy version mismatch for %q", ref.ID)
		}
	}

	return nil
}
