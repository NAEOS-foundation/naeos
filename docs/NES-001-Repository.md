# NES-001 Repository

## 1. Status
- Status: Draft
- Version: 0.1
- Owner: NAEOS Core Team

## 2. Purpose
This specification defines the repository structure and organizational conventions required to maintain NAEOS artifacts.

## 3. Scope
This document covers the logical organization of repository directories, documentation conventions, contribution rules, and navigation principles.

## 4. Normative References
- NES-000 Foundation
- NAEOS Governance

## 5. Repository Structure
- constitution/: normative policy and constitutional documents.
- governance/: decision-making and review processes.
- kernel/: base runtime and primitive services.
- policy/: policy rules and policy compilation artifacts.
- specification/: core technical specifications.
- docs/: modular documentation set.
- internal/specification/: lexer, parser, normalizer, and resolver components.
- internal/neir/: NEIR domain model, builder, serializer, validator, and versioning support.
- internal/planner/: graph, scheduler, and optimizer logic.
- internal/generation/: engine, templates, targets, and renderers.
- internal/governance/: policy and review enforcement modules.
- internal/knowledge/: graph, index, and provenance support.
- internal/runtime/: execution engine, lifecycle management, and telemetry.
- internal/kernel/: runtime services, registry, and event handling.
- internal/shared/: shared types, errors, contracts, and utilities.
- templates/: reusable templates.
- examples/: reference implementations.

## 6. Requirements
### 6.1 Functional Requirements
- FR-001: The repository shall provide a clear namespace for each artifact class.
- FR-002: The repository shall support discoverability through consistent naming and metadata.

### 6.2 Non-Functional Requirements
- NFR-001: Repository structure shall remain understandable by new contributors.
- NFR-002: Changes shall preserve traceability to higher-level governance documents.

## 7. Contribution Rules
1. Every significant change shall include context and rationale.
2. Documents shall include metadata relevant to ownership and status.
3. Changes shall remain consistent with superior normative documents.

## 8. Acceptance Criteria
- A contributor can locate the relevant domain document within the repository without ambiguity.
- The repository structure supports future extension without reorganization.
