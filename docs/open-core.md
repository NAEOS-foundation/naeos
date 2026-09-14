# NAEOS Open-Core Boundary

- Status: Draft
- Version: 0.1
- Owner: NAEOS Foundation

## 1. Purpose

This document defines the proposed boundary between the open-source NAEOS Core
and future commercial offerings (NAEOS Cloud, NAEOS Enterprise). It exists to
make licensing expectations explicit for contributors, adopters, and investors.

This is a governance document. It does not change the license of any code in
this repository.

## 2. Current State

As of this document, the entire first-party repository is licensed under the
Apache License 2.0 (`LICENSE`, `NOTICE`). This includes:

- the NAEOS CLI and daemon (`cmd/`)
- the kernel, specification engine, policy engine, and runtime (`internal/`)
- the compiler, validator, graph, NEIR model, and registry tooling
- integrations, governance, audit, and compliance modules
- examples, documentation, and generated artifacts

No first-party component is currently under a restrictive or commercial
license.

## 3. Boundary Principles

1. Code contributed to this repository is expected to be Apache-2.0
   (inbound = outbound, enforced by DCO).
2. The open-source core must remain independently useful without NAEOS Cloud.
3. The commercial boundary must not retract rights already granted under the
   Apache-2.0 license of this repository.
4. New proprietary components must live outside this repository or in clearly
   separated directories with explicit licensing.
5. The NAEOS brand and trademark are protected separately and are not granted
   by the Apache License.

## 4. Candidate Boundary

| Component | Recommended Model | Rationale |
|---|---|---|
| NAEOS Core (specification engine, NEIR, compiler, validator, runtime) | Apache-2.0 | Differentiating open-source value; drives adoption and contributor growth |
| Governance, policy, audit, compliance modules | Apache-2.0 / Review | Community value; keep open unless enterprise obligations require separation |
| CLI, SDK, plugins SDK | Apache-2.0 | Adoption surface; must stay open |
| Integrations and adapters | Apache-2.0 | Ecosystem growth |
| NAEOS Cloud (hosted specification/planning/generation services) | Commercial SaaS terms | Hosted service; Apache-2.0 does not require hosting the code publicly |
| NAEOS Enterprise (admin, SSO, advanced governance, support) | Commercial terms / proprietary where justified | Enterprise controls; separation keeps core lean |
| Marketplace and registry hosting | Commercial / governed | Operated by the Foundation; content licensing differs from code |

## 5. Guardrails

- Do not relicense, re-move, or re-license already-committed Apache-2.0 code.
- Keep the open-source boundary to the components the community actively
  consumes (`docs/open-core.md` is intentionally conservative).
- Publish an SBOM with each release so commercial and open boundaries remain
  auditable.
- Keep the trademark policy separate from licensing (see `NOTICE`).

## 6. Contribution Implications

Contributors submit under DCO with inbound = outbound (Apache-2.0). This means:

- The Foundation may incorporate contributions into NAEOS Core and into
  commercial products built on NAEOS Core, under the Apache-2.0 terms already
  granted to every other user.
- Contributions do not transfer ownership; the contributor retains copyright
  and grants a license to everyone, including the Foundation.
- Commercial components built on top of the open core do not require
  contributor permission, because Apache-2.0 already permits this.

## 7. Open Questions

- Registration and ownership of the NAEOS trademark (see `NOTICE`).
- Whether NAEOS Enterprise should be licensed, source-available, or closed.
- Whether specific internal modules (e.g., multitenant, observability, billing)
  should move to commercial components before NAEOS Cloud launches.