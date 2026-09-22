# NAEOS

**Nusantara Engineering & Architecture Operating System**

> **The Engineering Control Plane for AI Coding Agents.**
>
> Build software with AI agents under explicit **architecture, policy, execution, evidence, and verification** controls.

<p>
  <a href="https://github.com/NAEOS-foundation/naeos/actions/workflows/ci.yml"><img src="https://github.com/NAEOS-foundation/naeos/actions/workflows/ci.yml/badge.svg" alt="CI status"></a>
  <a href="https://img.shields.io/badge/go-1.26.6+-00ADD8"><img src="https://img.shields.io/badge/go-1.26.6+-00ADD8?logo=go&logoColor=white" alt="Go 1.26.6+"></a>
  <a href="https://github.com/NAEOS-foundation/naeos/releases"><img src="https://img.shields.io/github/v/release/NAEOS-foundation/naeos" alt="Latest release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="Apache 2.0"></a>
</p>

NAEOS is an open-source engineering framework for building production software with AI coding agents. It provides the engineering layer between **what a team intends**, **what an agent proposes**, **what is authorized to execute**, and **what can be independently verified afterward**.

It is not another AI coding assistant. NAEOS is the control plane around the agent.

## Why NAEOS?

AI coding agents can generate code quickly. Production engineering still requires explicit answers to harder questions:

- What architecture is the agent operating within?
- Which actions are allowed?
- Who or what authorizes an external side effect?
- What actually happened during execution?
- What evidence proves the result?
- Can another system independently verify the claim?
- Can the same engineering intent be carried across different AI agents?

NAEOS treats these as engineering-system concerns rather than prompt-writing concerns.

## The model

Without an engineering control plane:

```text
Prompt → AI Agent → Code / Action → ?
```

With NAEOS:

```text
Specification
      ↓
NEIR
      ↓
Validation + Policy
      ↓
Agent Context / Intent
      ↓
Authorized Execution
      ↓
Observation → Evidence
      ↓
Independent Verification
```

Two architectural views describe the same system:

**Engineering model**

```text
Specification → NEIR → Validation → Policy → AI Context → Generation
```

**AI Engineering control plane**

```text
Agent Intent → Policy Decision → Authorized Execution
            → Observation → Evidence → Independent Verification
```

The first describes how engineering knowledge moves through NAEOS. The second describes how agent actions are controlled and made auditable.

## Core concepts

| Concept | Role |
|---|---|
| **Specification** | Declares system intent and engineering requirements |
| **NEIR** | Canonical engineering representation derived from specifications |
| **Policy** | Determines which proposed actions are permitted |
| **Runtime** | Executes only within the authorized boundary |
| **Evidence** | Records observable results and integrity-relevant metadata |
| **Verification** | Independently checks claims about outputs or side effects |
| **Handoffs** | Carries explicit contracts between agents, tools, and execution boundaries |
| **AI Compiler** | Translates engineering context into agent-specific instructions |
| **Extensions** | Adds profiles, plugins, adapters, and integrations without changing the core model |

## See it run

The fastest way to understand NAEOS is the CLI control-plane demo:

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos

go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

The demo exercises the specification-first pipeline and produces traceable run metadata.

For the Todo API demonstration:

```bash
go build -o naeos ./cmd/naeos
./examples/todo-api/run-demo.sh
```

See [START-HERE.md](START-HERE.md) for the supported onboarding path.

## A concrete run

The `naeos run` path exposes structured engineering-run metadata, including identifiers such as:

```text
run_id
specification_hash
neir_hash
validation_state
policy_evaluation
generated_context
artifacts
evidence
```

The goal is not merely to generate code. The goal is to preserve enough engineering context and evidence to understand **what was intended, what was authorized, what happened, and what can be verified**.

## AI coding agents

NAEOS is designed to remain vendor-neutral at the engineering layer.

NAEOS targets Go **1.26.6 or later** and can produce agent-specific instruction/context artifacts for tools such as:

- GitHub Copilot
- Claude Code
- OpenAI Codex
- Cursor
- Gemini CLI
- OpenCode
- Windsurf

The engineering model remains upstream of the individual agent. Changing agents should not require rebuilding the project's engineering rules from scratch.

## Evidence and verification

NAEOS includes engineering experiments and implementation paths around:

- policy boundaries and authorization
- durable audit/evidence records
- tamper detection
- handoff contracts
- independent verification
- artifact signing and verification
- SBOM generation
- security and vulnerability checks
- benchmark and fuzz gates

These experiments are evidence of specific mechanisms and behaviors. They should not be interpreted as blanket proof of every production property.

See [experiments/](experiments/) for the current experiment suite, including the **Evidence V5.6** milestone.

## Trust model

NAEOS is built around explicit boundaries:

1. **Intent is not authorization.**
2. **Policy decisions are distinct from runtime execution.**
3. **Observations are distinct from claims.**
4. **Evidence should outlive an agent's conversational memory.**
5. **Verification should not depend solely on the component making the claim.**
6. **Untrusted handoff input must not silently become downstream authority.**
7. **Controllable boundaries must be revalidated before consequential external actions.**
8. **Version and contract mismatches should fail closed where required by the governing specification.**

These are engineering principles, not guarantees that every integration is automatically safe.

## Architecture authority

NAEOS uses a normative documentation hierarchy:

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

Public guides such as README, START-HERE, and GETTING-STARTED explain and navigate the system; they do not override normative specifications.

Start with [DOCUMENTATION-AUTHORITY.md](DOCUMENTATION-AUTHORITY.md), then review:

- [NAEOS Reference Architecture](Reference%20Architecture/NAEOS-NRA-001.md)
- [NAEOS Master Technical Specification](NAEOS-MTS-001.md)
- [NAEOS Specification](specification/NAEOS-SPEC-001.md)
- [NEIR](docs/NES-023-NEIR.md)
- [CLI Reference](docs/NES-028-CLI-Reference.md)

**Architecture Drives Engineering.**

## Repository map

```text
naeos/
├── cmd/naeos/          # CLI
├── internal/
│   ├── specification/  # parsing, normalization, resolution
│   ├── neir/           # engineering representation
│   ├── compiler/       # AI instruction compilation
│   ├── governance/     # policy and governance
│   ├── runtime/        # execution gateway
│   ├── evidence/       # evidence storage
│   ├── verification/   # independent verification
│   ├── audit/          # audit trail
│   ├── signing/        # artifact signing
│   ├── pluginsdk/      # plugin SDK
│   └── ...             # supporting subsystems
├── pkg/                 # public Go packages
├── docs/                # engineering specifications
├── experiments/         # executable engineering experiments
├── constitution/        # project constitution
├── governance/          # governance material
└── examples/            # runnable examples
```

## Current release

**NAEOS 3.6.0** is the current documented software release.

NAEOS software releases and experiment milestones use different version sequences. For example, **NAEOS 3.6.0** is a software release, while **Evidence V5.6** identifies an engineering experiment milestone.

For release history and changes, see [CHANGELOG.md](CHANGELOG.md).

## Documentation

- [START-HERE.md](START-HERE.md) — onboarding and first contribution path
- [GETTING-STARTED.md](GETTING-STARTED.md) — developer setup
- [DOCUMENTATION-AUTHORITY.md](DOCUMENTATION-AUTHORITY.md) — normative documentation model
- [WHITEPAPER-EN.md](WHITEPAPER-EN.md) — English whitepaper
- [WHITEPAPER.md](WHITEPAPER.md) — Bahasa Indonesia whitepaper
- [DOCUMENTATION-INDEX.md](DOCUMENTATION-INDEX.md) — document index
- [docs/](docs/) — NAEOS engineering specifications
- [CHANGELOG.md](CHANGELOG.md) — release history

## Contributing

NAEOS is built in the open. Contributions are welcome across the engineering stack:

| Track | Examples |
|---|---|
| **Policy** | authorization, governance, policy evaluation |
| **Runtime** | execution boundaries, lifecycle, runtime controls |
| **Evidence** | audit, receipts, integrity, evidence storage |
| **Handoffs** | agent/tool contracts and boundary revalidation |
| **Verification** | independent checks and trust mechanisms |
| **Plugins** | SDK, adapters, profiles, extensions |
| **Experiments** | reproducible engineering experiments and benchmarks |
| **Documentation** | specifications, architecture, guides |

Read [CONTRIBUTING.md](CONTRIBUTING.md), then start with [START-HERE.md](START-HERE.md).

## Community and partnerships

- [GitHub Discussions](https://github.com/NAEOS-foundation/naeos/discussions) — technical discussion and ideas
- [Discord](https://discord.gg/naeos) — community discussion
- [NAEOS Partner Program](PARTNER-PROGRAM.md) — technology, infrastructure, AI, academic, community, pilot, and strategic collaboration

## Security and governance

- [SECURITY.md](SECURITY.md) — supported release lines and security reporting
- [Engineering Constitution](constitution/) — project-level engineering principles
- [Governance](governance/) — project governance material
- [AI-assisted provenance policy](docs/ai-provenance.md) — guidance for AI-assisted contributions
- [Open-core boundary](docs/open-core.md) — relationship between NAEOS Core and future offerings

## License

NAEOS is released under the [Apache License 2.0](LICENSE). See [NOTICE](NOTICE) for attribution and branding information.

Contributions are accepted under the [Developer Certificate of Origin](https://developercertificate.org/) as described in [CONTRIBUTING.md](CONTRIBUTING.md).

The **NAEOS**, **NEIR**, and NAEOS logo names are branding of the NAEOS Foundation. The Apache License grants no trademark rights; see [NOTICE](NOTICE) for the current branding and trademark policy.

---

**Active development · NAEOS 3.6.0**

**Architecture Drives Engineering.**
