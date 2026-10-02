# NAEOS Engineering Note #001
## External Validation: Testing the Engineering Boundary, Not the Claim

**Status:** Open for External Validation  
**Program:** External Validator Cohort #001  
**Scope:** Reproducibility, policy enforcement, evidence, and handoff  
**Budget:** $0  
**Validation principle:** **Evidence first. Claims second.**

---

## 1. Why This Note Exists

AI coding agents can generate software quickly.

The harder engineering problem is determining whether an agent was actually allowed to perform an action, under which policy, with which authorization, and what evidence remains afterward.

NAEOS is being developed around this problem.

Rather than asking the community to trust the architecture or accept a set of claims, NAEOS is open to independent technical validation.

The objective is straightforward:

> **Can an external engineer reproduce the documented control path from a clean checkout and independently verify the resulting evidence?**

If the answer is no, that result is valuable.

---

## 2. What We Are Testing

The external validation focuses on a bounded engineering path:

```
Specification
    ↓
NEIR
    ↓
Validation
    ↓
Policy Decision
    ↓
AI Context
    ↓
Authorized Generation
    ↓
Artifacts
    ↓
Evidence
```

The validator is not asked to endorse NAEOS.

They are asked to determine whether the documented behavior can actually be reproduced.

The validation therefore covers:

- reproducibility from a fixed commit;
- policy enforcement;
- validation behavior;
- authorization boundaries;
- generated artifacts;
- evidence generation;
- evidence integrity;
- traceability between execution and recorded evidence;
- documentation clarity;
- reproducibility friction.

---

## 3. The Independent Attempt

The first attempt must be performed independently.

The validator starts from a clean checkout and a fixed NAEOS commit.

They record:

- commit SHA;
- operating system and architecture;
- Go/toolchain version;
- commands executed;
- exit status;
- generated artifacts;
- `run_id`;
- `specification_hash`;
- `neir_hash`;
- evidence verification results;
- deviations from the documented procedure.

Maintainer assistance is intentionally excluded from the first attempt.

This creates a useful distinction between:

> **“The maintainer can make it work.”**

and

> **“An external engineer can reproduce it from the documentation.”**

The second is the property this program is designed to measure.

---

## 4. What Counts as Evidence

A successful command is not sufficient evidence by itself.

The validation record should connect the observed behavior to inspectable artifacts.

Examples include:

- specification files;
- generated NEIR;
- validation output;
- policy rejection evidence;
- AI context;
- authorized generation output;
- execution trace;
- generated artifacts;
- `run_id`;
- specification hash;
- NEIR hash;
- EvidenceBundle;
- EvidenceBundle verification result.

The goal is not simply to show that something happened.

The goal is to establish:

> **what happened, under which conditions, and what evidence allows another engineer to inspect it.**

---

## 5. What We Want Validators to Challenge

A useful validator is not someone who confirms that everything works.

We want engineers to challenge assumptions.

Examples:

- Does the Golden Path actually work from a fresh checkout?
- Is the policy boundary observable?
- Can a denied operation be demonstrated clearly?
- Can evidence be traced back to the corresponding execution?
- Are hashes and identifiers consistent?
- Are documentation steps sufficient without maintainer intervention?
- Are there undocumented assumptions?
- What happens when expected conditions are not satisfied?
- Can another engineer independently reconstruct what happened?

A failure is not a bad validation result.

A failure that is reproducible is valuable engineering evidence.

---

## 6. Validation Outcomes

Results are recorded using explicit categories:

| Result | Meaning |
|---|---|
| **VERIFIED** | Documented behavior reproduced with supporting evidence |
| **PARTIAL** | Some documented behavior reproduced, with material deviation |
| **FAILED** | Expected behavior could not be reproduced |
| **BLOCKED** | Validation could not proceed because of an external blocker |
| **AMBIGUOUS** | Evidence was insufficient to establish the result |
| **NOT TESTED** | The validator did not execute the relevant test |

These categories intentionally avoid turning validation into a subjective score.

---

## 7. What This Does Not Prove

External validation of a repository workflow does not automatically establish:

- production readiness;
- customer adoption;
- enterprise compliance;
- security of every external AI agent integration;
- suitability for every deployment environment;
- absence of undiscovered vulnerabilities.

Those are separate engineering questions requiring separate evidence.

The purpose of this cohort is narrower:

> **Test whether the documented engineering behavior is independently reproducible and inspectable.**

---

## 8. Cohort #001

**Target:** 5 independent validators  
**Missions:** 3  
**Duration:** approximately 14 days  
**Compensation:** none  
**Output:** public evidence-based validation report

The missions cover:

1. **Verified Golden Path**
2. **Policy Change During Execution**
3. **Evidence & Handoff Verification**

Mission 001 is the initial entry point.

The canonical program protocol is defined in:

- `docs/EXTERNAL-VALIDATOR-PROGRAM.md`
- `docs/EXTERNAL-VALIDATOR-BRIEF.md`
- `docs/EXTERNAL-VALIDATION-MISSIONS.md`

The reproducible runbook remains:

- `docs/EXTERNAL-VALIDATION.md`

---

## 9. Fixed Evaluation Boundary

The cohort intake fixes the initial evaluation target to a specific repository commit.

Validators should record the exact commit they evaluate and must not silently mix evidence from different commits.

The evaluation should preserve the distinction between:

- repository proof;
- local workflow behavior;
- production-readiness claims;
- customer/adoption claims;
- broader security claims.

Evidence from one boundary must not be generalized into another claim without additional validation.

---

## 10. The Engineering Question

The central question is intentionally small:

> **Can another engineer reproduce the control boundary and verify its evidence without relying on the maintainer's interpretation?**

That question is more useful than whether the architecture sounds convincing.

NAEOS will be evaluated by what can be reproduced, inspected, and challenged.

**Evidence determines the result.**
