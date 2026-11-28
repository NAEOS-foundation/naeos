---
title: "Governance & Compliance: Policy Controls and Evidence Boundaries"
description: "Understand NAEOS policy, authorization, execution evidence, and how to evaluate governance capabilities without assuming compliance certification."
date: 2026-07-25
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Engineering governance needs more than policy files or generated compliance reports. Teams need clear authorization boundaries, traceable execution, inspectable evidence, and independent verification.

## Governance in NAEOS

NAEOS is an open-source, vendor-neutral engineering control plane for AI coding agents. Its current public proof path demonstrates policy and authorization boundaries, controlled execution, durable evidence, and independent verification.

```
Specification → NEIR → Validation + Policy
→ Agent Intent → Authorized Execution
→ Observation → Evidence → Independent Verification
```

## Evaluate controls by evidence

When evaluating governance capabilities, distinguish:
- **Policy evaluation** — what rules and inputs were evaluated, and what decision was returned.
- **Authorization** — what action, identity, scope, and capability were permitted or denied.
- **Execution** — whether an authorized action actually ran and what outcome was observed.
- **Evidence** — what records are retained and how integrity and identity bindings are verified.
- **Compliance mapping** — how implemented controls map to a framework, including gaps and scope.

A policy example or framework mapping is not, by itself, proof of an implemented control or certification. NAEOS does not claim blanket SOC 2, HIPAA, or GDPR compliance. Organizations must assess the applicable scope, configuration, operating processes, and independent evidence for their own environment.

## Technical evaluation

Start with the public Control Plane and reproducible Golden Path. Inspect the decision and execution records, then run the independent verifier against the serialized evidence. Use the repository's current documentation and release baseline rather than assuming older feature descriptions remain current.

- [FAQ](/faq/)
- [About NAEOS](/about/)
- [Whitepaper](/whitepaper/)
- [Repository](https://github.com/NAEOS-foundation/naeos)
