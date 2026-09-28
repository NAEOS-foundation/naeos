// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pdr

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "strings"
 "github.com/NAEOS-foundation/naeos/internal/controlplane"
 "github.com/NAEOS-foundation/naeos/internal/evidence"
 "github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

type Record struct {
 RecordVersion string `json:"record_version"`
 RecordID string `json:"record_id"`
 PolicyID string `json:"policy_id"`
 PolicyVersion string `json:"policy_version"`
 SchemaVersion string `json:"schema_version"`
 DecisionID string `json:"decision_id"`
 Outcome string `json:"outcome"`
 Evidence []EvidenceRef `json:"evidence"`
 Verification Verification `json:"verification"`
 ChangeClassification *Classification `json:"change_classification,omitempty"`
}
type EvidenceRef struct { EvidenceID string `json:"evidence_id"`; Digest string `json:"digest"`; Kind string `json:"kind,omitempty"` }
type Verification struct { Status string `json:"status,omitempty"`; EvidenceDigest string `json:"evidence_digest,omitempty"` }
type Classification struct { Risk string `json:"risk"`; Criticality string `json:"criticality"`; Decision string `json:"decision"` }
type Dependencies struct { Policies *policy.Registry; Evidence *evidence.EvidenceStore; Ledger *controlplane.Ledger }
type Result struct { RecordID string `json:"record_id"`; DecisionID string `json:"decision_id"`; PolicyID string `json:"policy_id"`; PolicyVersion string `json:"policy_version"`; EvidenceIDs []string `json:"evidence_ids"`; LedgerEventID string `json:"ledger_event_id"` }

func Verify(data []byte, deps Dependencies) (Result,error) {
 var rec Record
 if err:=json.Unmarshal(data,&rec);err!=nil{return Result{},fmt.Errorf("decode PDR: %w",err)}
 if err:=verifyRecord(rec,deps);err!=nil{return Result{},err}
 event,_:=deps.Ledger.Decision(rec.DecisionID)
 return Result{RecordID:rec.RecordID,DecisionID:rec.DecisionID,PolicyID:rec.PolicyID,PolicyVersion:rec.PolicyVersion,EvidenceIDs:evidenceIDs(rec.Evidence),LedgerEventID:event.ID},nil
}
func verifyRecord(rec Record,deps Dependencies) error {
 if rec.RecordVersion!="1.0.0"||rec.SchemaVersion!="1.0.0"{return fmt.Errorf("unsupported PDR/schema version: record=%q schema=%q",rec.RecordVersion,rec.SchemaVersion)}
 if rec.RecordID==""||rec.PolicyID==""||rec.PolicyVersion==""||rec.DecisionID==""{return fmt.Errorf("PDR identity bindings are required")}
 if len(rec.Evidence)==0{return fmt.Errorf("PDR requires at least one evidence reference")}
 if deps.Policies==nil||deps.Evidence==nil||deps.Ledger==nil{return fmt.Errorf("PDR verifier dependencies are required")}
 active,ok:=deps.Policies.GetActive(rec.PolicyID);if !ok{return fmt.Errorf("active policy %s is unavailable",rec.PolicyID)}
 if active.Version!=rec.PolicyVersion{return fmt.Errorf("stale policy binding: PDR=%s active=%s",rec.PolicyVersion,active.Version)}
 event,ok:=deps.Ledger.Decision(rec.DecisionID);if !ok{return fmt.Errorf("decision %s is not present in the control-plane ledger",rec.DecisionID)}
 expected:=decisionForOutcome(rec);if event.Decision!=expected{return fmt.Errorf("ledger decision mismatch: PDR=%s ledger=%s",expected,event.Decision)}
 if idx,err:=deps.Evidence.Verify();err!=nil{return fmt.Errorf("evidence integrity verification failed at index %d: %w",idx,err)}
 seen:=make(map[string]struct{},len(rec.Evidence))
 for _,ref:=range rec.Evidence{
  if ref.EvidenceID==""||ref.Digest==""{return fmt.Errorf("PDR evidence references require id and digest")}
  if _,ok:=seen[ref.EvidenceID];ok{return fmt.Errorf("duplicate PDR evidence reference %s",ref.EvidenceID)}
  seen[ref.EvidenceID]=struct{}{}
  if len(ref.Digest)!=sha256.Size*2{return fmt.Errorf("invalid evidence digest for %s",ref.EvidenceID)}
  if _,err:=hex.DecodeString(ref.Digest);err!=nil{return fmt.Errorf("invalid evidence digest for %s: %w",ref.EvidenceID,err)}
  ev:=deps.Evidence.ByID(ref.EvidenceID);if ev==nil{return fmt.Errorf("referenced evidence %s is unavailable",ref.EvidenceID)}
  if !strings.EqualFold(ev.Hash,ref.Digest){return fmt.Errorf("evidence digest mismatch for %s",ref.EvidenceID)}
 }
 if rec.Verification.Status=="passed"{
  if rec.Verification.EvidenceDigest==""{return fmt.Errorf("passed verification requires evidence digest")}
  found:=false;for _,ref:=range rec.Evidence{if strings.EqualFold(ref.Digest,rec.Verification.EvidenceDigest){found=true;break}}
  if !found{return fmt.Errorf("verification digest is not bound to PDR evidence")}
 }
 if rec.Outcome=="verified"&&rec.Verification.Status!="passed"{return fmt.Errorf("verified outcome requires passed verification")}
 return nil
}
func decisionForOutcome(rec Record)controlplane.DecisionStatus{
 if rec.ChangeClassification!=nil{switch rec.ChangeClassification.Decision{case "deny":return controlplane.DecisionDeny;case "require_review":return controlplane.DecisionPending;case "allow":return controlplane.DecisionAllow}}
 switch rec.Outcome{case "deny":return controlplane.DecisionDeny;case "require_review":return controlplane.DecisionPending;default:return controlplane.DecisionAllow}
}
func evidenceIDs(refs []EvidenceRef)[]string{out:=make([]string,0,len(refs));for _,ref:=range refs{out=append(out,ref.EvidenceID)};return out}
