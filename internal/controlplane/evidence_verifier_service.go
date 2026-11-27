// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0
package controlplane
import (
 "context"
 "crypto/subtle"
 "encoding/hex"
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "strings"
 "time"
)
var ErrAuthoritativeEvidenceNotVerified = errors.New("authoritative execution evidence is not verified")
type AuthoritativeExecutionObservation struct { ExecutionID string; Outcome ExecutionOutcome; EvidenceRef string; Authority string; EvidenceSHA256 string; ObservedAt time.Time; ValidUntil time.Time }
type AuthoritativeExecutionEvidenceSource interface { ObserveExecution(context.Context,string,ExecutionOutcome,string)(AuthoritativeExecutionObservation,error) }
type ExecutionEvidencePolicy interface { Version() string; EvaluateExecutionEvidence(context.Context,AuthoritativeExecutionObservation,ExecutionOutcome) error }
type EvidenceVerifierReplayGuard interface { Claim(context.Context,string,string)(bool,error) }
type ProviderNeutralEvidenceVerificationService struct { source AuthoritativeExecutionEvidenceSource; policy ExecutionEvidencePolicy; guard EvidenceVerifierReplayGuard; scope string; token string; now func()time.Time }
func NewProviderNeutralEvidenceVerificationService(source AuthoritativeExecutionEvidenceSource, policy ExecutionEvidencePolicy, guard EvidenceVerifierReplayGuard, scope, token string)(*ProviderNeutralEvidenceVerificationService,error) {
 if source==nil || policy==nil || guard==nil || strings.TrimSpace(policy.Version())=="" || strings.TrimSpace(scope)=="" || len(strings.TrimSpace(scope))>128 || strings.TrimSpace(token)=="" { return nil,errors.New("source, policy, durable replay guard, verifier scope, and credential are required") }
 return &ProviderNeutralEvidenceVerificationService{source:source,policy:policy,guard:guard,scope:strings.TrimSpace(scope),token:token,now:time.Now},nil
}
func (s *ProviderNeutralEvidenceVerificationService) Handler()http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
 if s==nil||s.source==nil||s.policy==nil||s.guard==nil||s.scope==""||s.token=="" {writeJSON(w,503,map[string]string{"error":"evidence verification service is not configured"});return}
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");writeJSON(w,405,map[string]string{"error":"method not allowed"});return}
 token,ok:=bearerToken(r.Header.Get("Authorization"));if !ok||len(token)!=len(s.token)||subtle.ConstantTimeCompare([]byte(token),[]byte(s.token))!=1{writeJSON(w,401,map[string]string{"error":"service credential required"});return}
 r.Body=http.MaxBytesReader(w,r.Body,maxEvidenceVerifierRequestBytes);dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields();var req evidenceVerificationRequest
 if err:=dec.Decode(&req);err!=nil{writeJSON(w,400,map[string]string{"error":"invalid evidence verification request"});return}
 if err:=dec.Decode(new(any));!errors.Is(err,io.EOF){writeJSON(w,400,map[string]string{"error":"request body must contain exactly one JSON object"});return}
 if len(req.RequestID)!=32{writeJSON(w,400,map[string]string{"error":"request ID must be 128-bit hexadecimal"});return}
 if _,err:=hex.DecodeString(req.RequestID);err!=nil{writeJSON(w,400,map[string]string{"error":"request ID must be 128-bit hexadecimal"});return}
 req.RequestID=strings.ToLower(req.RequestID);req.ExecutionID=strings.TrimSpace(req.ExecutionID);req.EvidenceRef=strings.TrimSpace(req.EvidenceRef)
 if req.ExecutionID==""||len(req.ExecutionID)>256||(req.Outcome!=ExecutionOutcomeCommitted&&req.Outcome!=ExecutionOutcomeNotCommitted)||len(req.EvidenceRef)>4096{writeJSON(w,400,map[string]string{"error":"invalid execution binding or expected outcome"});return}
 claimed,err:=s.guard.Claim(r.Context(),s.scope,req.RequestID);if err!=nil{writeJSON(w,503,map[string]string{"error":"durable replay protection unavailable"});return};if !claimed{writeJSON(w,409,map[string]string{"error":"request ID already used"});return}
 obs,err:=s.source.ObserveExecution(r.Context(),req.ExecutionID,req.Outcome,req.EvidenceRef);if err!=nil{if errors.Is(err,ErrAuthoritativeEvidenceNotVerified){writeJSON(w,422,map[string]string{"error":"authoritative execution evidence is not verified"});return};writeJSON(w,503,map[string]string{"error":"authoritative evidence source unavailable"});return}
 now:=s.now().UTC();if obs.ExecutionID!=req.ExecutionID||obs.Outcome!=req.Outcome||!validEvidenceReference(obs.EvidenceRef)||strings.TrimSpace(obs.Authority)==""||!sha256HexPattern.MatchString(obs.EvidenceSHA256)||obs.ObservedAt.IsZero()||obs.ValidUntil.IsZero()||obs.ObservedAt.After(now.Add(30*time.Second))||!obs.ValidUntil.After(now)||obs.ValidUntil.Before(obs.ObservedAt){writeJSON(w,422,map[string]string{"error":"provider observation is unbound, incomplete, or expired"});return}
 if err:=s.policy.EvaluateExecutionEvidence(r.Context(),obs,req.Outcome);err!=nil{writeJSON(w,422,map[string]string{"error":"evidence does not satisfy verification policy"});return}
 writeJSON(w,200,evidenceVerificationResponse{RequestID:req.RequestID,ExecutionID:obs.ExecutionID,Outcome:obs.Outcome,Status:"VERIFIED",EvidenceRef:obs.EvidenceRef,Authority:obs.Authority,EvidenceSHA256:obs.EvidenceSHA256,ObservedAt:obs.ObservedAt,ValidUntil:obs.ValidUntil,VerificationPolicyVersion:s.policy.Version()})
})}
