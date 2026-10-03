# NAEOS Enterprise Control Plane Contract

**Status:** Draft for Enterprise Readiness v2  
**Scope:** E1 — Product contract and control-plane boundary

## Purpose

Define the stable contract between enterprise clients, the NAEOS control plane, policy enforcement, runtime execution, and evidence. Enterprise integrations consume stable NAEOS contracts rather than internal implementation packages.

## Canonical lifecycle

```
Intent → Context/Specification → Policy Evaluation → Capability Authorization
      → Approval (when required) → Execution → Observation → Verification
      → Evidence → Audit/Export
```

Every consequential execution must be attributable to tenant, project, environment, actor/service identity, agent identity, repository/resource, capability, policy identifier and immutable version, authorization decision, approval when required, execution identifier, verification result, and evidence identifier.

## Trust boundaries

- **Control plane:** identity, policy, authorization, approvals, configuration, evidence metadata, audit queries, administration.
- **Execution boundary:** enforce authorization immediately before consequential side effects.
- **Evidence boundary:** record observations and integrity metadata independently from agent conversational state.
- **External systems:** GitHub, CI/CD, AI agents, SIEM, secret managers, IdPs, and infrastructure are explicit integration boundaries.

## Non-negotiable semantics

1. Intent does not grant authorization.
2. Agent capability does not imply permission.
3. Authorization is scoped and versioned.
4. Policy changes cannot silently broaden an already-authorized execution.
5. Handoff cannot increase downstream authority.
6. Consequential execution revalidates its authorization boundary.
7. Unsupported or ambiguous authorization states fail closed where required.
8. Evidence records observations; claims require verification.
9. Evidence integrity must be independently checkable.
10. Tenant and project boundaries are explicit in enterprise-scoped operations.

## Enterprise resource model

```
Organization
  └── Tenant
       ├── Project
       │    ├── Environment
       │    ├── Repository
       │    ├── Agent
       │    ├── Policy
       │    └── Evidence
       └── Identity / Service Account
```

## API principles

The enterprise API must support explicit versioning, idempotency for mutations, correlation IDs, deterministic authorization decisions, structured errors, pagination, optimistic concurrency/version checks where required, audit metadata for administrative mutations, and backward-compatibility rules.

Internal Go package structures are not the long-term API contract.

## Identity and authorization

Distinguish human identity, workload/service identity, agent identity, resource identity, capability, policy, authorization decision, and approval.

RBAC controls administrative access. Capability authorization controls what an agent or workload may execute. They are related but not interchangeable.

## Evidence contract

A minimum evidence record must make it possible to reconstruct:

```
who → requested what → under which policy/version
   → received which authorization → executed what
   → observed what → verified what
```

Evidence must support integrity verification and explicit retention/deletion semantics.

## Deployment contract

Enterprise profiles must support Kubernetes, PostgreSQL, enterprise secret management, TLS, isolated network environments, air-gapped operation where required, backup/restore, and upgrade/rollback.

Availability and disaster-recovery objectives remain deployment-specific until measured and documented.

## E1 exit criteria

- control-plane resource model documented
- API versioning and compatibility rules documented
- authorization semantics documented
- evidence semantics documented
- trust boundaries documented
- contract mapped to implementation packages
- at least one executable contract test
- later enterprise phases reference this contract instead of redefining semantics

## Claims boundary

This contract defines engineering requirements. It is not a claim of certification, regulatory compliance, production scale, security assurance, or availability SLA.
