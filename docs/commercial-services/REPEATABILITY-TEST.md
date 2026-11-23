# Repeatability Test — ENG-SIM-001 (Fictional)

Status: **PARTIAL PASS — desk-based simulation only**  
Date: 2026-11-23  
Assessment subject: PT DevWorks Nusantara (fictional)  
Assessment ID: ENG-SIM-001

## Purpose

Check whether assessment classifications and conclusions are reproducible when the same supplied evidence is independently re-evaluated.

## Method and limitations

- Assessor A: baseline assessment workbook v1.1 QA-fix.
- Assessor B: independent AI-assisted desk re-evaluation of the same evidence statements and references.
- No new evidence was collected; no external facts were introduced.
- This is **not** a two-human inter-rater study and is not statistically representative.
- The assessed controls are CTL-001 through CTL-007. CTL-008 and CTL-009 remain Not Assessed and are excluded from maturity agreement.

## Agreement results

| Measure | Agreement | Notes |
|---|---:|---|
| Control maturity | 6/7 (85.7%) | One discrepancy, CTL-005 |
| Scenario outcome classifications | 8/8 (100%) | Conservative statuses preserved |
| Finding severity | 4/4 (100%) | All severity classifications agree |

## Control comparison

| Control | Assessor A | Assessor B | Result |
|---|---|---|---|
| CTL-001 Identity | 1 — Ad hoc | 1 — Ad hoc | Agree |
| CTL-002 Authorization | 0 — Absent | 0 — Absent | Agree |
| CTL-003 Execution | 1 — Ad hoc | 1 — Ad hoc | Agree |
| CTL-004 Policy | 2 — Defined | 2 — Defined | Agree |
| CTL-005 Approval | 2 — Defined | 1 — Ad hoc | **Disagree** |
| CTL-006 Testing | 1 — Ad hoc | 1 — Ad hoc | Agree |
| CTL-007 Evidence | 2 — Defined | 2 — Defined | Agree |
| CTL-008 Resilience | Not Assessed | Not Assessed | Excluded |
| CTL-009 Incident | Not Assessed | Not Assessed | Excluded |

### Calibration issue — CTL-005

Approval/review process existence is evidenced, but pre-execution enforcement is inconsistent. Assessor A interpreted process existence as Defined maturity (2); Assessor B applied a more conservative score of Ad hoc (1), emphasizing inconsistent implementation.

**Calibration recommendation:** provisionally score CTL-005 at maturity 1 until evidence demonstrates defined and consistently enforced pre-execution approval. If level 2 is retained, the report should explicitly distinguish documented process from control implementation effectiveness and provide rationale.

## Scenario classification agreement

| Scenario | Result | Repeatability note |
|---|---|---|
| TS-001 Authorized repository read | PASS | Positive control executed |
| TS-002 Unauthorized repository write | FAIL | Tool body and write side effect observed |
| TS-003 Missing principal | BLOCKED | Missing harness fixture |
| TS-004 Missing approval | INCONCLUSIVE | Pending/no execution observed, but authoritative instrumentation absent |
| TS-005 Replay action | NOT TESTED | Deferred |
| TS-006 Revoked permission | NOT TESTED | Deferred |
| TS-007 Evidence missing | INCONCLUSIVE | No authoritative verification source |
| TS-008 Tool unavailable | INCONCLUSIVE | Error response lacks side-effect verification |

No BLOCKED, NOT TESTED, or INCONCLUSIVE result was upgraded to PASS.

## Finding severity agreement

| Finding | Risk score | Severity | Agreement |
|---|---:|---|---|
| FND-001 Agent Authorization Boundary | 16 | Critical | Agree |
| FND-002 Shared Agent Execution Identity | 12 | High | Agree |
| FND-003 Incomplete Decision-to-Execution Traceability | 12 | High | Agree |
| FND-004 Limited Negative Authorization Testing | 12 | High | Agree |

Risk score = likelihood × impact. Severity/risk-band override should always carry a documented rationale.

## QA conclusion

**PARTIAL PASS.** The supplied simulated evidence supported consistent scenario classifications and finding severity. Control maturity scoring had one discrepancy that reveals an ambiguity in the maturity rubric—process definition versus consistent implementation.

## Follow-up actions

1. Clarify maturity-level guidance and CTL-005 calibration.
2. Run a true independent human assessor walkthrough, preferably blind to Assessor A’s scores.
3. Compare score, classification, rationale, and evidence references; record disagreements and resolve rubric ambiguity.
4. Conduct a client-style delivery simulation before external engagement.
5. Keep legal/SOW/privacy review and SaaS commercialization on HOLD until required gates are satisfied.

## Companion file

The detailed comparison workbook is maintained as a working artifact outside this repository and is not yet a ratified canonical deliverable. All assessment entities and evidence are fictional.
