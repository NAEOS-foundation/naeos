# Assessment Maturity Rubric

Status: Draft — internal calibration  
Version: 1.0.0-draft

## Scoring rule

Score each control from 0–5 using evidence of both process and actual control behavior. A higher score requires evidence for the preceding levels and stronger implementation/verification. Documentation alone does not establish effective enforcement. Use the lowest level fully supported by the evidence for the control's intended objective; do not average evidence gaps into a higher score.

| Level | Label | Minimum evidence / interpretation |
|---|---|---|
| 0 | Absent | Control is not present, or evidence demonstrates it is absent. If not assessed due to missing evidence, use Not Assessed—not 0. |
| 1 | Ad hoc | Informal, inconsistent, manual, or reactive practices. No repeatable control design or reliable enforcement is evidenced. |
| 2 | Defined | Documented process/design, roles, rules, and expected behavior exist. Implementation or enforcement may be partial, inconsistent, or not verified. |
| 3 | Implemented | Control is deployed in the assessed scope and operates as designed; implementation evidence is available. Documentation alone is insufficient. |
| 4 | Verified | Implementation has been tested against defined positive and negative criteria; results and evidence support effectiveness in the assessed scope. |
| 5 | Operationally Validated | Sustained effectiveness is demonstrated in operational use, including monitoring, exceptions, maintenance, and relevant failure/recovery conditions over a stated period. |

## Conservative scoring and evidence

- **Not Assessed** is a status, not a maturity score. Do not score lack of access/evidence as 0 unless absence is established.
- Select the highest level for which all minimum evidence is satisfied. If evidence only supports a documented design but enforcement is inconsistent, score no higher than 2; if no repeatable design is evidenced, score 1.
- Record the maturity score, status, evidence IDs, limitations, and rationale separately. Distinguish control design maturity from observed effectiveness.
- A scenario FAIL is direct evidence of an effectiveness gap and must be reflected in the relevant control score/risk analysis; it cannot be offset by policy documentation.
- For authorization and approval controls, a denied decision must be observed before the tool body or protected side effect. A post-execution rejection is not effective pre-execution enforcement.
- If two assessors disagree, compare evidence references and level criteria; document the disagreement and rationale. Do not silently average scores. Escalate rubric ambiguity for calibration.
- Overrides or exceptions must include rationale, approver, scope, and residual risk; never use them to erase observed failure.

## CTL-005 calibration example — Approval before sensitive action

If an approval workflow is documented but approval enforcement is inconsistent, score **1 — Ad hoc** when repeatability/defined operation is not evidenced. Score **2 — Defined** only when the documented design and roles are sufficiently defined; note that implementation effectiveness remains unverified or inconsistent. Score **3 or higher** requires evidence that enforcement is deployed and operates in scope. Any unauthorized execution observed is a FAIL for that scenario and must be disclosed, regardless of maturity score.

## Repeatability and independence

A desk-based second review is useful for detecting ambiguity but is not a human inter-rater study. For a genuine independent walkthrough, provide the same evidence package to a second assessor without the first assessor's scores, then compare maturity, status, severity, rationale, and evidence references. Record the agreement measure and limitations; small samples should not be treated as statistically representative.
