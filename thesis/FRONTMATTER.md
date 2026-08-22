# Front Matter

---

## Title Page

**NAEOS: A Declarative Engineering Runtime for Specification-Driven, AI-Native Software Construction**

*Design, Implementation, and Empirical Evaluation of a Canonical Intermediate Representation, Executable Governance, and Vendor-Neutral AI Integration in an Open-Source Platform*

**A Technical Monograph**

**Author:** Bayu Priatno — Founder & Lead Architect, NAEOS Foundation
**Affiliation:** NAEOS Foundation (independent, non-institutional)
**Version:** 1.0.0 — August 2026
**Artifact:** `github.com/NAEOS-foundation/naeos` (Apache License 2.0)
**Web:** `https://naeos.dev`

**DOI (Zenodo):** [10.5281/zenodo.22060578](https://doi.org/10.5281/zenodo.22060578) (concept — all versions)
**This version (v1.1.0):** [10.5281/zenodo.22061988](https://doi.org/10.5281/zenodo.22061988)
**Preprint:** *[optional: arXiv `cs.SE` identifier after endorsement]*
**License:** Apache 2.0 (code) / CC BY 4.0 (manuscript)

**Keywords:** declarative engineering; model-driven development; intermediate representation; deterministic build pipelines; AI-native software engineering; governance as code; traceability; WebAssembly sandboxing; reproducibility

**Computing Classification:** ACM CCS → Software and its engineering → Software creation and management → Software verification and validation → Compilers; Software architectures; Maintaining software

---

## Publication Note

This monograph is published independently by the platform's author. It is archived with a permanent DOI on **Zenodo** (auto-minted from the NAEOS GitHub releases) and may additionally appear as an **arXiv preprint** (`cs.SE`). Distilled portions are prepared for submission to industry and tool tracks of international software engineering venues (e.g., ICSE-SEIP, FSE-Industry, ASE-Industry), which accept practitioner-authored work without academic affiliation. A practitioner-oriented summary is maintained on the project site under `site/content/blog`.

## Declaration of Originality

I hereby declare that this monograph is my own original work. All sources — published or unpublished — have been fully acknowledged and cited. All quantitative results reported herein were produced by experiments I executed against the open-source NAEOS platform (version recorded in Appendix A), and all reproduction commands are disclosed in the same appendix. This work has not been submitted for any academic degree.

Jakarta, August 2026 — Bayu Priatno

---

## Abstract (Extended — International Format)

Modern software systems increasingly span five programming languages, three cloud targets, and six AI coding agents, while their normative descriptions decay across wikis, tickets, and hand-maintained per-tool configuration files. We characterize this condition as **source-of-truth multiplicity** and ask whether it can be eliminated structurally rather than managed procedurally.

We present **NAEOS**, an open-source declarative engineering runtime that compiles a single human-readable specification through a nine-stage deterministic pipeline into a canonical intermediate representation — **NEIR**, a fourteen-domain model with semver schema governance — from which code (Go, TypeScript, Python, Java, Rust), infrastructure (Docker, Kubernetes, Terraform), documentation, compliance evidence (SOC 2/HIPAA/GDPR), and AI instruction sets (Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode) are derived. Governance norms are compiled into executable rules and enforced by validators and tamper-evident audit chains; third-party extensions execute in a capability-restricted WebAssembly sandbox; agents integrate through one standard Model Context Protocol server rather than per-vendor adapters.

Empirically, two independent compilations produced 79 byte-identical artifacts (SHA-256 pairwise comparison); replicated benchmarks across two hardware platforms and three model scales show 5.8 ms small-spec full runs, ~127 µs validation, linear memory scaling, and stable allocation profiles; parallel adapter fan-out improves large-scale generation by ~6 %. We formalize a **determinism-preserving AI integration** pattern that confines LLM nondeterminism to audited assistive roles without contaminating reproducible derivation.

The contribution set — (1) whole-system IR design, (2) governance-as-executable-rules, (3) AI-target compilation, and (4) an evaluated reference implementation with committed baselines — demonstrates that specification-driven construction can meet the reproducibility demands of regulated environments without sacrificing AI leverage.

---

## Author Biography

**Bayu Priatno** is the founder and lead architect of the **NAEOS Foundation**, the open-source initiative behind the NAEOS declarative engineering runtime ("*Specify Once. Build Anywhere.*"). As project founder, he defined the platform's core vision — the specification as single source of truth, enforced through executable governance — and led the design of its central components: the NEIR intermediate representation, the deterministic compilation pipeline, the multi-agent AI compiler, and the WASM-based plugin sandbox.

He has worked at the intersection of developer tooling, distributed systems, and AI-assisted engineering as an independent engineer and open-source author. His work on NAEOS spans architecture, implementation, and community stewardship, including the NES specification series (NES-000–NES-054), the project's Engineering Constitution, and its governance model — an unusual practice of applying a platform's own principles to itself.

His professional interests include model-driven engineering, intermediate representation design, reproducible build systems, supply-chain security, and the architectural containment of nondeterministic AI in engineering toolchains. Contact: `https://naeos.dev` · GitHub: `github.com/NAEOS-foundation/naeos`

---

## Acknowledgments

The author thanks the NAEOS Foundation contributors and maintainers, the early adopters whose feedback shaped the specification language and NEIR schema, and the open-source communities behind Go, wazero, Cobra, and the Model Context Protocol — this monograph stands on their work.
