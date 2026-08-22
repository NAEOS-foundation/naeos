# Chapter 2 — Background and Related Work

This chapter situates NAEOS within the research and practice landscape. Each subsection reviews a body of prior work, identifies its contribution, and states the gap that NAEOS addresses.

## 2.1 Model-Driven Engineering (MDE/MDA)

Model-Driven Engineering proposes that models — not code — should be the primary engineering artifact, with transformations refining abstract models into implementations. The Object Management Group's Model-Driven Architecture (MDA) formalized this as a stack of Computation-Independent Models (CIM), Platform-Independent Models (PIM), and Platform-Specific Models (PSM), connected by automated transformations.

**Contributions.** MDE established the intellectual foundation for specification-first construction: the idea that a platform-independent description can be systematically refined into platform-specific artifacts through chains of model-to-model (M2M) and model-to-text (M2T) transformations. It also produced industrial-strength tooling — the Eclipse Modeling Framework (EMF) for metamodeling and model persistence, ATL and QVT for transformation languages, JetBrains MPS for projectional editing — demonstrating that DSLs and transformation chains are practical at scale. Academic evaluation frameworks for MDE (e.g., benchmarking of transformation engines) matured the field's empirical standards.

**Why classical MDE underdelivered.** Retrospective analyses of MDA adoption identify recurring failure factors: heavyweight notation (XMI serialization hostile to diff and merge); metamodel rigidity (changing a metamodel invalidates tooling and models); poor round-tripping (hand-edited code diverging from models with no reconciliation path); and tool monocultures that locked projects into specific vendor stacks. These are precisely the failure modes NAEOS's design counters point by point: YAML substrate (diffable, mergeable), semver-versioned schema with migration engine (governed evolution), derivation-only code generation with no hand-edit expectation (no round-trip problem), and adapters over open formats (no tool lock-in).

**Gap.** Industrial MDE adoption has been hampered by heavyweight tooling, metamodel rigidity, and poor integration with the everyday toolchain of developers (git, CI, text editors). Crucially, classical MDE predates both cloud-native infrastructure and AI coding agents; it has no story for generating *AI context* or for enforcing *governance* as part of the transformation chain. NAEOS inherits the MDE philosophy but deliberately uses lightweight, diff-friendly YAML rather than XMI-style metamodels, integrates with ordinary developer workflows, and extends the transformation chain to a new class of targets: AI instruction sets.

## 2.2 Infrastructure as Code and Declarative Configuration

Terraform (HCL), Kubernetes manifests, Pulumi, and CloudFormation demonstrated at scale that declarative specifications plus reconciliation engines can manage complex real-world systems. Their key insight is that **desired state**, expressed declaratively, can be compiled or converged into actual state by a deterministic mechanism.

**Contributions.** IaC normalized the idea that infrastructure is versioned, reviewed, and derived rather than hand-maintained; content-addressed plan caching, drift detection, and structural diffs are now standard expectations.

**Gap.** IaC tools scope their declarations to *infrastructure only*. Application code, documentation, API contracts, and AI context remain outside the declared surface. NAEOS generalizes the IaC pattern from infrastructure to the **entire engineering artifact set**: one specification yields services, storage, deployment manifests (including Terraform HCL for AWS/GCP/Azure), tests, documentation, and AI context. In this framing, NAEOS treats "the system" the way Terraform treats "the datacenter."

### 2.2.1 From Desired State to Derived State

The IaC conceptual model rests on three pillars that NAEOS generalizes:

1. **Declarative desired state** — the user states *what* should exist, not the steps to create it;
2. **Deterministic convergence** — a tool computes and applies the difference between declared and actual state;
3. **Plan inspection** — proposed changes are rendered reviewable before application (Terraform's plan/apply split).

NAEOS adopts all three, with one decisive widening of scope: its "actual state" is not deployed infrastructure but the artifact set of a repository, and its "plan" is a NEIR diff plus governance evaluation. The pipeline's stage caching (SHA-256 keyed) plays the role of Terraform's state tracking: re-running an unchanged specification is provably a no-op at every stage whose input hash matches.

### 2.2.2 Limitations Carried into This Work

IaC tooling also surfaces problems NAEOS must avoid: schema sprawl across provider versions, secrets handling in declarative files (`$env{}` interpolation exists precisely to keep environment-specific values out of normative specs), and the "configuration language Turing-completeness trap" — which motivates keeping the Specification Language deliberately non-Turing-complete: interpolation, references, includes, conditionals, and pure functions, but no unbounded iteration or recursion.

## 2.3 Intermediate Representations in Compiler Design

Compilers solved the N×M translation problem decades ago by factoring translation through an intermediate representation: N frontends produce a shared IR consumed by M backends. LLVM's IR is the canonical modern example — a well-versioned, serializable, analyzable form on which optimization passes operate once, independent of source language and target.

**Contribution.** The IR pattern is the direct design precedent for NEIR. As argued in ADR-002, without a shared IR every spec-to-language path requires a dedicated compiler; with NEIR, N generators each implement one IR-to-language backend, validation happens once at IR level, and source maps enable precise error reporting back to specification locations.

**Gap.** Classical compiler IRs encode *computation*. No mainstream IR encodes the full breadth of an engineered *system* — architecture patterns, domains, modules, services, APIs, storage, security policies, deployment topology, testing strategy, and even AI integration directives — as a single canonical, diff-able model. NEIR's contribution is precisely this breadth, together with semver schema governance so that generators can track schema drift (Section 4.2).

## 2.4 Scaffolding Generators and Low-Code Platforms

Project generators (`create-react-app`, Spring Initializr, Yeoman) and low-code platforms both attempt to reduce boilerplate. Generators run once and exit; their output immediately begins to drift from any documented intent. Low-code platforms retain a model but typically bind users to proprietary runtimes and UI-driven editing, trading away code ownership and vendor neutrality.

**Gap.** NAEOS is explicitly *not* a scaffolder: the specification remains live, the pipeline remains re-runnable, and generated artifacts stay aligned across the lifecycle ("keeps projects aligned with their spec"). Unlike low-code platforms, output is plain multi-language source in standard toolchains, and the runtime itself is Apache-2.0 open source with no runtime lock-in — satisfying what the project calls *technological sovereignty*.

The distinction can be stated as a test: **can the tool regenerate your project after a year of drift?** Generators fail this test by design (they run once); low-code platforms fail it differently (the "regeneration" never ends — you live inside the vendor's runtime forever). A declarative engineering runtime must pass both halves: full regeneration on demand, and full ownership of the generated material in between. NAEOS's lifecycle model — generate, validate against spec, structurally diff NEIR, migrate deliberately — is constructed around passing this test indefinitely.

## 2.5 AI-Assisted Development

LLM-based assistants (Copilot, Claude, Cursor, Gemini, Codex) have moved from autocompletion to agentic participation: reading repositories, planning changes, and writing multi-file patches — a shift documented across 395 primary studies in the most comprehensive survey of the field [R27]. Research and practitioner literature converge on two observations relevant here: (i) results depend heavily on the *context* provided to the model; and (ii) outputs are *nondeterministic*, which conflicts with reproducibility requirements in regulated environments.

**Contributions.** The Model Context Protocol (MCP) standardizes how agents discover and invoke external tools, decoupling agent vendors from capability providers. Repository-level instruction files (e.g., `CLAUDE.md`, `.cursorrules`) emerged as a de facto way to steer agents.

**Gap.** Instruction files are hand-written per tool and decay; context quality depends on manual curation; and there is no architectural principle confining AI nondeterminism. Chapter 5 shows how NAEOS addresses this by *compiling* AI context from NEIR (one source → six formats), exposing capabilities through MCP rather than vendor APIs, and codifying AI's permitted role in a dedicated constitution (NAEOS-CON-002) aligned with the human-accountability and reproducibility articles.

### 2.5.1 Context Engineering as a First-Class Concern

Practitioner experience converges on the observation that agent output quality is bounded by context quality: relevant architectural constraints, naming conventions, module boundaries, and security rules must be present in the prompt window. This elevates "project context" from documentation to an engineering artifact with its own lifecycle — and therefore, we argue, with its own *derivation* requirement. Hand-maintained context files violate that requirement structurally, not accidentally.

### 2.5.2 Nondeterminism as an Architectural Input

Classical toolchain design assumes tools are functions: same input, same output. LLM agents break this assumption. Prior responses in the literature range from treating AI output as untrusted suggestion (human review gates) to post-hoc verification (generated tests, static analysis). NAEOS's contribution (Chapter 5) is a *role-separation architecture*: deterministic derivation for normative artifacts; audited, reviewed AI participation only inside explicitly assistive roles; human accountability constitutionalized. This reframes nondeterminism from a hazard to be suppressed into a capability to be sandboxed — analogous to how the WASM sandbox treats untrusted plugin code.

## 2.6 Governance, Compliance, and Auditability

Compliance frameworks (SOC 2, HIPAA, GDPR) demand demonstrable controls and audit trails. Traditional compliance evidence is assembled manually from tickets, screenshots, and documents — expensive and fragile. Policy-as-code systems (e.g., OPA/Rego) made rules machine-evaluable for infrastructure and access control.

**Gap.** Policy-as-code has not been applied to the full engineering lifecycle — from specification admission ("no specification, no implementation") through artifact review to release accountability. NAEOS contributes a normative hierarchy (Constitution → Standards → Project Rules → Local Rules) in which top-level principles are *compiled* into executable rules evaluated by validators, policy engines, and AI review, with SHA-256 chained, optionally encrypted audit logs and framework-mapped control reporting (Chapter 4.3).

### 2.6.1 Traceability Research

Requirements-traceability research has long identified the core obstacle as *manual link maintenance*: trace matrices are created at project start and rot immediately. Automated approaches (trace recovery via information retrieval, artifact-link heuristics) recover links post hoc with imperfect precision. NAEOS takes the orthogonal route: because every artifact is *generated* through a pipeline whose inputs are known, the trace matrix is a by-product of construction rather than a maintained document — links cannot rot because they are recomputed on every run. This positions generation-based traceability as the strongest available form of the property: exact, current, and complete over the derived artifact space.

## 2.7 Extensibility and Sandboxing

Plugin architectures must balance extensibility against host safety. Approaches range from in-process dynamic loading (fast, unsafe) to container isolation (safe, heavy). WebAssembly offers a middle path: near-native speed, memory isolation by construction, platform neutrality, and small footprint. The wazero runtime adds zero-dependency, pure-Go execution — significant for a tool distributed as a single static binary.

**Positioning.** ADR-008 selects WASM/wazero for NAEOS plugins with JSON-over-WASI communication, capability grants declared in plugin manifests, memory/time limits, signature verification, and hot reload. This follows the emerging industry pattern (Envoy/Fermyon/Spin, Shopify Functions) and applies it to engineering-runtime extensions.

## 2.8 Reproducible Builds and Software Provenance

The reproducible-builds movement demonstrated that byte-level determinism is achievable even for complex toolchains, and that determinism converts builds from acts of faith into verifiable claims. Supply-chain security work (SLSA framework, in-toto attestations) extends this with provenance: signed records of *how* an artifact was produced.

NAEOS internalizes both ideas at design time rather than retrofitting them: determinism is a constitutional article (VIII) enforced by content hashing; provenance is structural — every artifact's derivation path (requirement → spec → NEIR → adapter) is recorded in the reasoning graph and audit chain. Where supply-chain tooling attests to provenance post hoc, NAEOS makes provenance a queryable property of the pipeline itself.

## 2.9 Domain-Specific Languages: Embedding vs. Inventing

The choice between an external DSL, an embedded/internal DSL, and plain configuration formats shapes adoption more than expressive power. Reports from the DSL literature consistently identify tooling — editing, validation, refactoring support — as the dominant adoption factor. NAEOS's decision chain — YAML substrate + directive vocabulary (`$ref`, `$include`, `$env`, `$fn`, `$if`) + LSP backed by the real parser — is a pragmatic middle path: no new grammar to learn beyond directives, yet semantic validation far beyond JSON-schema checking. The LSP's use of the production parser (`internal/specification`) means editor diagnostics cannot disagree with pipeline diagnostics — a small but telling application of single-source-of-truth thinking to the developer experience itself.

## 2.10 Operational Patterns: Middleware, Observability, and Event-Driven Platforms

Modern platform engineering converged on a set of operational patterns NAEOS adopts wholesale: **middleware chains** for cross-cutting concerns (the pipeline's composable log/metrics/auth/cache chain mirrors HTTP middleware culture); **OpenTelemetry-style observability** (spans, batched export, Prometheus metrics) rather than ad-hoc logging; **event-driven internal architecture** (bus + observers) so features attach without core modification; and **content-addressed caching** as popularized by build systems (Bazel) and registries. None of these is novel individually; their disciplined application to an *engineering-generation* pipeline — with stage purity making each pattern directly applicable — is part of the system's contribution. The monograph treats them as established substrate (cited accordingly) and reserves analysis for what is new: the IR breadth, executable governance, and AI-target compilation.

## 2.11 Summary Table

| Approach | Single source of truth | Multi-language derivation | Executable governance | AI-native context | Vendor-neutral |
|---|---|---|---|---|---|
| Classical MDE/MDA | Partial (models) | Yes | No | No | Partial |
| IaC (Terraform et al.) | Infra only | Infra targets | Partial (plans/policies) | No | Yes |
| Compilers (LLVM-style) | Code | Via IR | No | No | Yes |
| Scaffolding generators | No (run once) | Yes (once) | No | No | Yes |
| Low-code platforms | Proprietary model | Runtime-bound | Partial | No | **No** |
| AI assistants alone | No | No | No | Per-tool, manual | Partial |
| **NAEOS** | **Specification** | **Yes (5 languages)** | **Yes (constitution→rules)** | **Yes (6 tools + MCP)** | **Yes** |

The table's rightmost row is not claimed as uniformly superior on every dimension — scaffolding generators remain simpler for throwaway prototypes, and IaC tools remain the right scope when only infrastructure varies. The claim is narrower and stronger: **no prior single approach covers all six columns**, and it is precisely their combination that the fragmentation problems of Section 1.1 demand. Chapters 3–5 present the architecture realizing this combination; Chapter 6 evaluates whether the combination holds together in practice.

The remainder of this monograph presents the architecture (Chapter 3) and design (Chapters 4–5) that realize the rightmost row, followed by evaluation (Chapter 6).
