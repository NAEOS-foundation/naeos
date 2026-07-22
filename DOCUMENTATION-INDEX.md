# Documentation Index

This document serves as the master index for navigating the NAEOS repository.

## 1. Core documents
- [README.md](README.md) — project summary and main entry point.
- [GETTING-STARTED.md](GETTING-STARTED.md) — onboarding guide.
- [CONTRIBUTING.md](CONTRIBUTING.md) — contribution guidelines.

## 2. Concepts and architecture
- [specification/NAEOS-SPEC-001.md](specification/NAEOS-SPEC-001.md) — core specification overview.
- [Reference Architecture/NAEOS-NRA-001.md](Reference%20Architecture/NAEOS-NRA-001.md) — reference architecture.
- [ARCHITECTURE-OVERVIEW.md](ARCHITECTURE-OVERVIEW.md) — conceptual architecture overview.

## 3. Governance and constitution
- [constitution/NAEOS-CON-001.md](constitution/NAEOS-CON-001.md) — engineering constitution.
- [governance/NAEOS-GOV-001.md](governance/NAEOS-GOV-001.md) — project charter.
- [governance/NAEOS-GOV-002.md](governance/NAEOS-GOV-002.md) — NAEOS long-term vision.
- [governance/NAEOS-GOV-003.md](governance/NAEOS-GOV-003.md) — NAEOS official mission.
- [governance/NAEOS-GOV-004.md](governance/NAEOS-GOV-004.md) — NAEOS manifesto.
- [governance/NAEOS-GOV-005.md](governance/NAEOS-GOV-005.md) — NAEOS core principles.
- [governance/NAEOS-GOV-006.md](governance/NAEOS-GOV-006.md) — governance model.
- [governance/NAEOS-GOV-007.md](governance/NAEOS-GOV-007.md) — roadmap.
- [governance/NAEOS-GOV-008.md](governance/NAEOS-GOV-008.md) — versioning policy.

## 4. Policy, Kernel, and Profile
- [kernel/NAEOS-KER-001.md](kernel/NAEOS-KER-001.md) — kernel specification.

### 4.1 Policy System Documentation
- [policy/NAEOS-POL-001.md](policy/NAEOS-POL-001.md) — spesifikasi policy compiler (high-level overview).
- [policy/NAEOS-POL-002.md](policy/NAEOS-POL-002.md) — panduan penulisan & definisi policy (cara membuat policies).
- [policy/NAEOS-POL-003.md](policy/NAEOS-POL-003.md) — contoh policy konkret (security, testing, documentation, compliance, deployment).
- [policy/NAEOS-POL-004.md](policy/NAEOS-POL-004.md) — best practices (design principles, organization, governance).
- [policy/NAEOS-POL-005.md](policy/NAEOS-POL-005.md) — policy compiler & engine (technical deep dive, architecture, graph resolution).
- [policy/NAEOS-POL-006.md](policy/NAEOS-POL-006.md) — evaluasi & enforcement (strategies, CI/CD integration, remediation).
- [policy/NAEOS-POL-007.md](policy/NAEOS-POL-007.md) — troubleshooting & FAQ (common issues, solutions, best practices Q&A).

### 4.2 Profile System Documentation
- [profile/NAEOS-PRO-001.md](profile/NAEOS-PRO-001.md) — profile system specification (high-level overview).
- [profile/NAEOS-PRO-002.md](profile/NAEOS-PRO-002.md) — implementation & setup guide (creating and activating profiles).
- [profile/NAEOS-PRO-003.md](profile/NAEOS-PRO-003.md) — concrete profile examples (base, startup, enterprise, fintech, healthcare, saas, microservices).
- [profile/NAEOS-PRO-004.md](profile/NAEOS-PRO-004.md) — best practices (design principles, policy design, governance).
- [profile/NAEOS-PRO-005.md](profile/NAEOS-PRO-005.md) — API & CLI reference (complete commands, usage examples, SDK).
- [profile/NAEOS-PRO-006.md](profile/NAEOS-PRO-006.md) — migration & upgrade guide (upgrade strategies, major version migration, rollback).
- [profile/NAEOS-PRO-007.md](profile/NAEOS-PRO-007.md) — troubleshooting & FAQ (common issues, solutions, best practices Q&A).

## 5. Supporting documents
- [GLOSSARY.md](GLOSSARY.md) — glossary of important terms.
- [ROADMAP.md](ROADMAP.md) — project development roadmap.
- [CHANGELOG.md](CHANGELOG.md) — version history and release notes.

## 6. Architecture Decision Records (ADRs)
- [docs/adr/001-why-go-for-runtime.md](docs/adr/001-why-go-for-runtime.md) — ADR-001: Why Go for the Runtime
- [docs/adr/002-why-neir-as-central-model.md](docs/adr/002-why-neir-as-central-model.md) — ADR-002: Why NEIR as the Central Model
- [docs/adr/003-why-mcp-for-ai-integration.md](docs/adr/003-why-mcp-for-ai-integration.md) — ADR-003: Why MCP for AI Integration
- [docs/adr/004-database-layer.md](docs/adr/004-database-layer.md) — ADR-004: Database Layer (PostgreSQL, MySQL, SQLite)
- [docs/adr/005-websocket-communication.md](docs/adr/005-websocket-communication.md) — ADR-005: WebSocket Communication
- [docs/adr/006-distributed-task-execution.md](docs/adr/006-distributed-task-execution.md) — ADR-006: Distributed Task Execution
- [docs/adr/007-prompt-library.md](docs/adr/007-prompt-library.md) — ADR-007: Prompt Library
- [docs/adr/008-wasm-plugin-sandbox.md](docs/adr/008-wasm-plugin-sandbox.md) — ADR-008: WASM Plugin Sandbox

## 7. Templates and processes
- [templates/ADR-template.md](templates/ADR-template.md) — Architecture Decision Record template.
- [templates/RFC-template.md](templates/RFC-template.md) — Request for Comments template.
- [examples/adr-example.md](examples/adr-example.md) — completed ADR example.
- [examples/rfc-example.md](examples/rfc-example.md) — completed RFC example.

## 8. Modular documentation structure
- [docs/README.md](docs/README.md) — NAEOS documentation structure map.
- [docs/NES-000-Foundation.md](docs/NES-000-Foundation.md) — foundation.
- [docs/NES-001-Repository.md](docs/NES-001-Repository.md) — repository.
- [docs/NES-002-Kernel.md](docs/NES-002-Kernel.md) — kernel.
- [docs/NES-003-Workspace.md](docs/NES-003-Workspace.md) — workspace.
- [docs/NES-004-Bootstrap.md](docs/NES-004-Bootstrap.md) — bootstrap.
- [docs/NES-005-Blueprint.md](docs/NES-005-Blueprint.md) — blueprint.
- [docs/NES-006-Template.md](docs/NES-006-Template.md) — template.
- [docs/NES-007-Generator.md](docs/NES-007-Generator.md) — generator.
- [docs/NES-008-Registry.md](docs/NES-008-Registry.md) — registry.
- [docs/NES-009-Plugin.md](docs/NES-009-Plugin.md) — plugin.
- [docs/NES-010-Knowledge.md](docs/NES-010-Knowledge.md) — knowledge.
- [docs/NES-011-Graph.md](docs/NES-011-Graph.md) — graph.
- [docs/NES-012-Policy.md](docs/NES-012-Policy.md) — policy.
- [docs/NES-013-Compiler.md](docs/NES-013-Compiler.md) — compiler.
- [docs/NES-014-Validator.md](docs/NES-014-Validator.md) — validator.
- [docs/NES-015-Runtime.md](docs/NES-015-Runtime.md) — runtime.
- [docs/NES-016-AI.md](docs/NES-016-AI.md) — AI.
- [docs/NES-017-Studio.md](docs/NES-017-Studio.md) — studio.
- [docs/NES-018-Cloud.md](docs/NES-018-Cloud.md) — cloud.
- [docs/NES-019-SDK.md](docs/NES-019-SDK.md) — SDK.
- [docs/NES-020-Security.md](docs/NES-020-Security.md) — security.
- [docs/NES-021-Testing.md](docs/NES-021-Testing.md) — testing.
- [docs/NES-022-Release.md](docs/NES-022-Release.md) — release.
- [docs/NES-028-CLI-Reference.md](docs/NES-028-CLI-Reference.md) — CLI reference (commands, flags, output).
- [docs/NES-029-Configuration.md](docs/NES-029-Configuration.md) — pipeline configuration reference.
- [docs/NES-030-Specification-Language.md](docs/NES-030-Specification-Language.md) — NAEOS specification language.
- [docs/NES-031-Errors.md](docs/NES-031-Errors.md) — error code catalog.
- [docs/NES-032-Telemetry.md](docs/NES-032-Telemetry.md) — telemetry and metrics reference.
- [docs/NES-033-Testing-Guide.md](docs/NES-033-Testing-Guide.md) — testing guide.
- [docs/NES-034-Event-Bus.md](docs/NES-034-Event-Bus.md) — internal event bus (pub/sub).
- [docs/NES-035-Version-Management.md](docs/NES-035-Version-Management.md) — SemVer version management.
- [docs/NES-036-Template-Renderer.md](docs/NES-036-Template-Renderer.md) — template rendering engine.
- [docs/NES-037-Knowledge-Graph-Provenance.md](docs/NES-037-Knowledge-Graph-Provenance.md) — knowledge graph and provenance tracking.
- [docs/NES-038-Shared-Types-Contracts.md](docs/NES-038-Shared-Types-Contracts.md) — shared types and contracts.
- [docs/NES-039-SDK-MultiLanguage.md](docs/NES-039-SDK-MultiLanguage.md) — multi-language SDK specification (Go, TS, Python, Java, Rust).
- [docs/NES-040-Output-Adapter-Architecture.md](docs/NES-040-Output-Adapter-Architecture.md) — output adapter architecture for language extensions.
- [docs/NES-042-Database.md](docs/NES-042-Database.md) — database layer (PostgreSQL, MySQL, SQLite).
- [docs/NES-043-WebSocket.md](docs/NES-043-WebSocket.md) — WebSocket real-time communication.
- [docs/NES-044-EventSourcing.md](docs/NES-044-EventSourcing.md) — event sourcing and aggregate snapshots.
- [docs/NES-045-Distributed.md](docs/NES-045-Distributed.md) — distributed task execution.
- [docs/NES-046-ConfigHotReload.md](docs/NES-046-ConfigHotReload.md) — configuration hot-reload.
- [docs/NES-047-PipelineCache.md](docs/NES-047-PipelineCache.md) — pipeline result caching.
- [docs/NES-048-PipelineMiddleware.md](docs/NES-048-PipelineMiddleware.md) — composable pipeline middleware.
- [docs/NES-049-AuditLogging.md](docs/NES-049-AuditLogging.md) — audit logging layer.
- [docs/NES-050-HCLParser.md](docs/NES-050-HCLParser.md) — HCL configuration parser.
- [docs/NES-051-ProfileDetection.md](docs/NES-051-ProfileDetection.md) — automatic language/framework detection.
- [docs/NES-052-CICD.md](docs/NES-052-CICD.md) — CI/CD pipeline automation.
- [docs/NES-053-WASMPlugin.md](docs/NES-053-WASMPlugin.md) — WASM plugin sandboxed execution.

## 9. Reading recommendations

### 9.1 For beginners (understanding the project)
Suggested reading order:
1. [README.md](README.md)
2. [GETTING-STARTED.md](GETTING-STARTED.md)
3. [specification/NAEOS-SPEC-001.md](specification/NAEOS-SPEC-001.md)
4. [constitution/NAEOS-CON-001.md](constitution/NAEOS-CON-001.md)
5. [policy/NAEOS-POL-001.md](policy/NAEOS-POL-001.md)

### 9.2 For the policy system
To understand and use the policy system, read in this order:
1. [policy/NAEOS-POL-001.md](policy/NAEOS-POL-001.md) — understand basic policy compiler concepts
2. [policy/NAEOS-POL-003.md](policy/NAEOS-POL-003.md) — see concrete policy examples
3. [policy/NAEOS-POL-002.md](policy/NAEOS-POL-002.md) — writing and defining policies
4. [policy/NAEOS-POL-004.md](policy/NAEOS-POL-004.md) — policy design best practices
5. [policy/NAEOS-POL-006.md](policy/NAEOS-POL-006.md) — evaluation and enforcement strategies
6. [policy/NAEOS-POL-005.md](policy/NAEOS-POL-005.md) — technical deep dive (optional, for developers)
7. [policy/NAEOS-POL-007.md](policy/NAEOS-POL-007.md) — troubleshooting when needed

### 9.3 For CLI users
To use NAEOS directly, read in this order:
1. [docs/NES-028-CLI-Reference.md](docs/NES-028-CLI-Reference.md) — complete CLI commands
2. [docs/NES-029-Configuration.md](docs/NES-029-Configuration.md) — configuration format
3. [docs/NES-030-Specification-Language.md](docs/NES-030-Specification-Language.md) — writing specifications
4. [examples/spec-minimal.yaml](examples/spec-minimal.yaml) — minimal specification example
5. [examples/spec-full.yaml](examples/spec-full.yaml) — full specification example

### 9.4 For developers & testing
1. [docs/NES-033-Testing-Guide.md](docs/NES-033-Testing-Guide.md) — testing guide
2. [docs/NES-031-Errors.md](docs/NES-031-Errors.md) — error catalog
3. [docs/NES-032-Telemetry.md](docs/NES-032-Telemetry.md) — observability

### 9.5 For the profile system
To understand and use the profile system, read in this order:
1. [profile/NAEOS-PRO-001.md](profile/NAEOS-PRO-001.md) — understand basic concepts
2. [profile/NAEOS-PRO-003.md](profile/NAEOS-PRO-003.md) — see concrete examples
3. [profile/NAEOS-PRO-002.md](profile/NAEOS-PRO-002.md) — implementation practices
4. [profile/NAEOS-PRO-004.md](profile/NAEOS-PRO-004.md) — best practices
5. [profile/NAEOS-PRO-005.md](profile/NAEOS-PRO-005.md) — CLI & API reference
6. [profile/NAEOS-PRO-007.md](profile/NAEOS-PRO-007.md) — troubleshooting when needed
7. [profile/NAEOS-PRO-006.md](profile/NAEOS-PRO-006.md) — upgrade guide when upgrading

## 10. Go package reference documents
The following documents directly reference Go packages in this repository:
- [docs/NES-002-Kernel-API.md](docs/NES-002-Kernel-API.md) — API publik `pkg/kernel`.
- [docs/NES-023-NEIR-Model.md](docs/NES-023-NEIR-Model.md) — model `internal/neir/model`.
- [docs/NES-026-Pipeline.md](docs/NES-026-Pipeline.md) — pipeline `pkg/pipeline`.
- [docs/NES-027-Governance.md](docs/NES-027-Governance.md) — governance `internal/governance`.
- [docs/NES-028-CLI-Reference.md](docs/NES-028-CLI-Reference.md) — CLI `cmd/naeos`.
- [docs/NES-029-Configuration.md](docs/NES-029-Configuration.md) — config `pkg/config`.
- [docs/NES-032-Telemetry.md](docs/NES-032-Telemetry.md) — telemetry `internal/runtime/telemetry`.
- [docs/NES-034-Event-Bus.md](docs/NES-034-Event-Bus.md) — event bus `internal/events`.
- [docs/NES-035-Version-Management.md](docs/NES-035-Version-Management.md) — versioning `internal/neir/version`.
- [docs/NES-036-Template-Renderer.md](docs/NES-036-Template-Renderer.md) — renderer `internal/generation/renderers`.
- [docs/NES-037-Knowledge-Graph-Provenance.md](docs/NES-037-Knowledge-Graph-Provenance.md) — knowledge `internal/knowledge`.
- [docs/NES-038-Shared-Types-Contracts.md](docs/NES-038-Shared-Types-Contracts.md) — shared `internal/shared`.
- [docs/NES-039-SDK-MultiLanguage.md](docs/NES-039-SDK-MultiLanguage.md) — SDK multi-language `internal/generation/adapters`.
- [docs/NES-040-Output-Adapter-Architecture.md](docs/NES-040-Output-Adapter-Architecture.md) — adapter architecture `internal/generation/adapters`.

## 11. Navigation notes
- Use this document as your entry point when searching for a specific topic.
- To understand the overall project, start with the core documents first.
- For contributions, see [CONTRIBUTING.md](CONTRIBUTING.md).
- For the policy system, start with NAEOS-POL-001 and continue as needed.
- For the profile system, start with NAEOS-PRO-001 and continue as needed.
