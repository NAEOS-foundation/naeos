# NAEOS External Validation Missions

**Version:** 1.0.0  
**Cohort:** #001

These missions extend [`docs/EXTERNAL-VALIDATION.md`](./EXTERNAL-VALIDATION.md). Mission 001 uses the existing Golden Path evidence chain; Missions 002 and 003 exercise specific governance and evidence boundaries when those workflows are available at the evaluated commit.

## Mission 001 — Verified Golden Path

### Objective
Determine whether an external engineer can independently execute the documented NAEOS Golden Path and verify its acceptance criteria.

### Procedure
1. Use a clean checkout.
2. Record the exact NAEOS commit and toolchain.
3. Follow the external validation runbook without maintainer walkthrough.
4. Build and execute the documented demo.
5. Inspect the generated evidence chain.
6. Complete the acceptance matrix.
7. Submit the result using the external validation result template.

### Minimum Evidence
- commit SHA;
- environment/toolchain;
- command sequence;
- demo exit status;
- generated evidence;
- observed run ID, specification hash, and NEIR hash;
- acceptance matrix;
- deviations.

### Result
Use: `VERIFIED`, `PARTIAL`, `FAILED`, `BLOCKED`, `AMBIGUOUS`, or `NOT TESTED`.

## Mission 002 — Policy Change During Execution

### Objective
Determine how the evaluated NAEOS workflow behaves when policy or authorization changes after execution has started.

### Validation Question
> What happens when authorization or policy changes while an execution is already in progress?

### Preconditions
Use only a documented, supported NAEOS scenario available at the evaluated commit. Do not invent undocumented production semantics.

Record:
```text
NAEOS Version:
Commit:
Initial Policy:
Initial Authorization:
Capability:
Execution ID:
```

### Procedure
1. Establish and record the initial policy/authorization state.
2. Start the supported execution scenario.
3. Record evidence that execution entered the initial state.
4. Apply the documented policy change.
5. Record the policy version/change event.
6. Observe actual behavior.
7. Record resulting execution state and evidence.
8. If supported, repeat from clean state.

### Observation Questions
Record facts for questions such as: Did execution continue? Was the capability blocked? Was authorization re-evaluated? Was an event emitted? Was evidence recorded? Was prior authorization reused?

Do not infer security impact beyond the observed evidence.

### Minimum Evidence
- initial policy;
- execution state;
- policy change;
- relevant logs/events;
- resulting state;
- reproduction steps.

## Mission 003 — Evidence & Handoff Verification

### Objective
Determine whether another engineer can reconstruct what happened from the evidence and handoff artifacts produced by the evaluated workflow.

### Validation Question
> Can an independent engineer reconstruct what happened from the recorded evidence and handoff state?

### Procedure
1. Execute the supported workflow that produces the relevant evidence.
2. Record execution ID, version, policy, capability, timestamp, and result.
3. Inspect the evidence artifact.
4. Where a handoff artifact exists, inspect its documented fields.
5. Ask an independent reviewer to reconstruct the execution without undocumented maintainer context.
6. Record missing or ambiguous information.

### Reconstruction Questions
The reviewer should be able to answer, based on available evidence:
```text
What happened?
Why was it allowed?
Which policy/version applied?
Which capability was authorized?
What was executed?
What evidence supports the result?
```

### Handoff Fields
Where supported by the evaluated workflow, inspect: version; payload/reference; policy; capability set; provenance; expiry/replay state; integrity/digest information.

Do not treat a field as present or authoritative unless the evaluated commit actually produces and documents it.

### Minimum Evidence
- evidence artifact;
- handoff artifact, if applicable;
- relevant logs;
- policy/version;
- execution identifier;
- reconstruction notes.

## Cross-Mission Submission
Use this structure for every mission:
```text
Validator ID:
Mission ID:
NAEOS Version:
Commit:
Environment:
Date:

Objective:

Preconditions:

Steps:

Expected Behavior:

Actual Behavior:

Result:

Evidence:

Finding:

Reproduction Steps:

Recommendation:

Maintainer Notes:
```

## Evidence Integrity
Do not remove failure evidence before submission, silently replace evidence after discovering the expected result, hide blockers, or change the environment without documenting the change.

If evidence changes:
```text
Evidence v1
    ↓
Reason for Update
    ↓
Evidence v2
```
The change must remain traceable.

## Mission Completion
A mission is complete when the first attempt is recorded, evidence is submitted, the result is classified, findings are reviewed, and any required re-test is completed.