# NAEOS Repository Architecture

**Status:** Stable
**Purpose:** Canonical map of the repository's structural domains and document boundaries.

## 1. Repository role

The NAEOS repository is the canonical open-source reference implementation of the **NAEOS AI Engineering Operating System**.

It intentionally contains implementation, normative engineering artifacts, experiments, documentation, website content, and project operations in one repository. The repository is broad by design; its domains are separated by responsibility rather than by repository count.

## 2. Structural domains

| Domain | Primary paths | Responsibility |
|---|---|---|
| Core implementation | `cmd/`, `pkg/`, `internal/` | Executable NAEOS implementation |
| Architecture & specifications | `Reference Architecture/`, `specification/`, `NAEOS-MTS-001.md`, `docs/NES-*.md` | Normative and technical contracts |
| Governance | `constitution/`, `governance/`, `policy/`, `profile/` | Constraints, policy, profiles, and project governance |
| Kernel | `kernel/` | Kernel architecture and kernel-specific engineering artifacts |
| Extensions | `plugins/`, `extensions/`, `docs/plugin-sdk.md`, `docs/NES-053-WASMPlugin.md` | Plugin and extension mechanisms |
| Evidence & verification | `experiments/`, evidence-related `docs/`, validation workflows | Reproducible observations and verification evidence |
| Examples | `examples/` | Reproducible demonstrations and contributor starting points |
| Public documentation | root `*.md`, `docs/` | Onboarding, reference, guides, ADRs, and project records |
| Website | `site/` | Public web experience and published documentation |
| Project operations | `.github/`, `brand/`, selected planning/GTM documents | CI, contribution workflows, release, security, and distribution |

## 3. Authority boundaries

The repository follows this precedence when documents conflict:

1. Constitution
2. Reference Architecture
3. Master Technical Specification
4. Engineering Specifications and ADRs
5. Implemented source code
6. Experiments and observations
7. Public guides and marketing material

The authoritative rules are defined in [DOCUMENTATION-AUTHORITY.md](DOCUMENTATION-AUTHORITY.md). This file is a structural map, not a replacement for that authority model.

## 4. Canonical entry points

### New visitor

`README.md` → `START-HERE.md` → `GETTING-STARTED.md`

### Architecture

`DOCUMENTATION-AUTHORITY.md` → `Reference Architecture/NAEOS-NRA-001.md` → `NAEOS-MTS-001.md` → relevant `docs/NES-*.md`

### Running NAEOS

`GETTING-STARTED.md` → `docs/GOLDEN-PATH.md` → `examples/demo-cli/`

### Contributing

`CONTRIBUTING.md` → relevant issue/template → implementation or documentation domain → pull request

### Website

`site/` is the publication layer. Website content must derive from canonical project terminology and must not become an independent architectural authority.

## 5. Document classes

Every new document should belong to one of these classes:

- **Normative** — defines a requirement or contract.
- **Reference** — explains an implemented contract or interface.
- **Guide** — helps users or contributors perform a task.
- **Example** — demonstrates a reproducible workflow.
- **Experiment** — records an observable, falsifiable result.
- **Project record** — roadmap, changelog, GTM, partner, or historical record.
- **Marketing** — communicates the project externally without redefining technical authority.

If a document does not have a clear class, it should not be added yet.

## 6. Duplication policy

Do not create parallel canonical documents for the same concept.

Examples:

- One canonical reference architecture.
- One canonical documentation authority model.
- One canonical brand/positioning source of truth.
- One canonical roadmap authority; public summaries link to it.
- CLI command pages may be generated/derived references and must not become competing specifications.

When two documents cover the same concept, prefer linking to the authoritative source over copying the full definition.

## 7. New-folder rule

Do not introduce a new top-level directory merely to make the repository look smaller or cleaner.

Create a new top-level domain only when at least one of these is true:

- it has a distinct lifecycle or build boundary;
- it has a distinct security/governance boundary;
- it has a stable contributor ownership boundary;
- it contains multiple artifacts that form a coherent domain.

Otherwise, place the artifact under the existing domain and update the documentation index.

## 8. Contributor navigation principle

A contributor should not need to understand the entire repository before making a small change. Each domain should expose a clear entry document and point back to the canonical authority hierarchy.

The repository is therefore **wide but disciplined**: breadth is intentional; ambiguity is not.

## 9. Change control

Structural changes should preserve these invariants:

- runtime behavior remains unchanged unless explicitly intended;
- canonical terminology remains consistent;
- authority relationships remain explicit;
- generated/derived documents remain distinguishable from normative documents;
- links and contributor paths remain valid.

`DOCUMENTATION-INDEX.md` remains the navigation index. This document explains the architecture behind that index.
