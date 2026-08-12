# Documentation Index

This master index maps high-level documentation to files and docs/* entries.

Last updated: 2026-08-12

## 1. Core documents
- README.md — project summary and main entry point (English-first)
- GETTING-STARTED.md — onboarding guide
- CONTRIBUTING.md — contribution guidelines and code of conduct
- WHITEPAPER-EN.md — whitepaper (English)
- WHITEPAPER.md — whitepaper (Bahasa Indonesia)

## 2. Concepts & architecture
- docs/NES-000-Foundation.md — foundation overview
- docs/ARCHITECTURE-OVERVIEW.md — conceptual architecture
- Reference Architecture/ — reference patterns and examples

## 3. Governance & constitution
- constitution/ — engineering constitution and governance artifacts
- governance/ — governance documents and policies

## 4. Policy, Kernel, and Profile
- docs/NES-002-Kernel.md — kernel specification and API reference
- policy/ — policy system documentation (see docs/NES-012-Policy.md)
- profile/ — profile system documentation (see docs/NES-019-SDK-MultiLanguage.md)

## 5. Supporting documents
- GLOSSARY.md — glossary of important terms
- ROADMAP.md — project development roadmap
- CHANGELOG.md — version history and release notes

## 6. Architecture Decision Records (ADRs)
- docs/adr/ — ADRs (see docs/adr/ for files and templates)

## 7. Templates and processes
- templates/ADR-template.md — ADR template
- templates/RFC-template.md — RFC template
- examples/adr-example.md — completed ADR example

## 8. Modular documentation (docs/)
A canonical set of NES documents live under docs/ and are the authoritative English-first sources.

- docs/README.md — documentation structure map
- docs/NES-000-Foundation.md
- docs/NES-001-Repository.md
- docs/NES-002-Kernel.md
- docs/NES-003-Workspace.md
- docs/NES-004-Bootstrap.md
- docs/NES-005-Blueprint.md
- docs/NES-006-Template.md
- docs/NES-007-Generator.md
- docs/NES-008-Registry.md
- docs/NES-009-Plugin.md
- docs/NES-010-Knowledge.md
- docs/NES-011-Graph.md
- docs/NES-012-Policy.md
- docs/NES-013-Compiler.md
- docs/NES-014-Validator.md
- docs/NES-015-Runtime.md
- docs/NES-016-AI.md
- docs/NES-017-Studio.md
- docs/NES-018-Cloud.md
- docs/NES-019-SDK-MultiLanguage.md
- docs/NES-020-Security.md
- docs/NES-021-Testing.md
- docs/NES-022-Release.md
- docs/NES-028-CLI-Reference.md
- docs/NES-029-Configuration.md
- docs/NES-030-Specification-Language.md
- docs/NES-031-Errors.md
- docs/NES-032-Telemetry.md
- docs/NES-033-Testing-Guide.md
- docs/NES-034-Event-Bus.md
- docs/NES-035-Version-Management.md
- docs/NES-036-Template-Renderer.md
- docs/NES-037-Knowledge-Graph-Provenance.md
- docs/NES-038-Shared-Types-Contracts.md
- docs/NES-039-SDK-MultiLanguage.md
- docs/NES-040-Output-Adapter-Architecture.md
- docs/NES-042-Database.md
- docs/NES-043-WebSocket.md
- docs/NES-044-EventSourcing.md
- docs/NES-045-Distributed.md
- docs/NES-046-ConfigHotReload.md
- docs/NES-047-PipelineCache.md
- docs/NES-048-PipelineMiddleware.md
- docs/NES-049-AuditLogging.md
- docs/NES-050-HCLParser.md
- docs/NES-051-ProfileDetection.md
- docs/NES-052-CICD.md
- docs/NES-053-WASMPlugin.md

## 9. Reading recommendations
### For beginners
1. WHITEPAPER-EN.md
2. README.md
3. GETTING-STARTED.md
4. docs/NES-000-Foundation.md

### For CLI users
1. docs/NES-028-CLI-Reference.md
2. docs/NES-029-Configuration.md
3. docs/NES-030-Specification-Language.md
4. examples/spec-minimal.yaml

## 10. Notes
- This index is intentionally concise: the authoritative content lives under docs/ (English-first). If a document exists only in Bahasa Indonesia it should be moved under docs/id/ with `-id.md` suffix.
- To propose structural changes to the documentation layout, open a PR against the docs/cleanup branch.
