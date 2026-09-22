# Documentation Authority Model

## Purpose

NAEOS contains public guides, normative specifications, implementation documents, and deterministic experiments. This document defines which source is authoritative when two documents appear to disagree.

## Authority hierarchy

Higher layers constrain lower layers:

```text
Constitution
    ↓
Reference Architecture
    ↓
Master Technical Specification
    ↓
Engineering Specifications / ADRs
    ↓
Implementation
    ↓
Experiments
```

### 1. Constitution

The Constitution defines the highest-level normative engineering principles and constraints.

Primary location:

- `constitution/`

### 2. Reference Architecture

The NAEOS Reference Architecture defines the approved architectural structure and major system boundaries.

Primary document:

- [NAEOS-NRA-001](Reference%20Architecture/NAEOS-NRA-001.md)

### 3. Master Technical Specification

The Master Technical Specification translates the reference architecture into implementable technical contracts.

Primary document:

- [NAEOS-MTS-001](NAEOS-MTS-001.md)

### 4. Engineering Specifications and ADRs

Engineering Specifications and Architecture Decision Records define concrete contracts, decisions, interfaces, and implementation constraints.

Primary locations:

- `docs/NES-*.md`
- `docs/adr/`

### 5. Implementation

Source code is the executable implementation of the applicable normative contracts.

If implementation differs from an applicable normative document, the discrepancy must be treated as an engineering issue rather than silently redefining the specification.

### 6. Experiments

Experiments validate specific, falsifiable claims against named repository components.

Experiments do **not** override normative architecture or specifications. They provide evidence about observed behavior and must explicitly state limitations.

See [experiments/README.md](experiments/README.md).

## Public guides

The following are navigation and onboarding documents:

- [README.md](README.md)
- [START-HERE.md](START-HERE.md)
- [GETTING-STARTED.md](GETTING-STARTED.md)
- [ARCHITECTURE-OVERVIEW.md](ARCHITECTURE-OVERVIEW.md)

They should explain the project consistently with the normative hierarchy, but they are not themselves the source of normative authority.

## Conflict resolution

When a conflict is found:

1. Identify the documents and their authority levels.
2. Prefer the higher-authority document.
3. Open an issue or change request describing the inconsistency.
4. Update the lower-authority document, or intentionally revise the higher-authority document through the appropriate governance process.
5. Add or update a deterministic test/experiment when the conflict concerns executable behavior.

## Versioning terminology

NAEOS uses more than one version namespace.

- **Release version** — software release, e.g. `3.6.0`.
- **Document version** — version of a normative specification, e.g. NRA `1.0.0`.
- **Experiment series version** — engineering/security experiment milestone, e.g. Evidence `V5.6`.

These identifiers must not be treated as a single global version sequence.

## Claim discipline

A document should distinguish:

- normative requirements;
- implemented behavior;
- experimental observations;
- proposed/future capabilities.

Do not describe a proposed architecture as an implemented capability. Do not describe a deterministic experiment as proof of security for every deployment.

## Change checklist

Before changing a normative document, verify:

- terminology is consistent;
- referenced versions are current;
- implementation impact is understood;
- dependent documents are identified;
- tests or experiments are updated when applicable;
- links remain valid.

Before changing a public guide, verify it remains consistent with the normative hierarchy.

