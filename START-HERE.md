# Start Here

You do not need to read the whole NAEOS architecture before you can understand the value. The shortest path is: understand the control boundary, run one reproducible proof, inspect the evidence, and then decide whether to contribute.

NAEOS is an open-source engineering control plane for AI coding agents. It connects specification, policy, authorized execution, observation, evidence, and independent verification so agent actions can be inspected against explicit engineering intent.

## Coming from the NAEOS newsletter?

If you arrived here from a newsletter, article, or community post, start here:

1. Understand the problem.
2. Inspect one real example.
3. Run one experiment.
4. Challenge one assumption.
5. Make one small contribution.

This is the bridge between reading about NAEOS and participating in the repository.

## 1. What is NAEOS?

NAEOS connects engineering intent, policy, authorized execution, and evidence around a shared engineering model. The repository's Golden Path makes that control flow reproducible and inspectable before generated artifacts are treated as evidence of the run.

The practical question is simple: when an AI agent takes action, how do we know it is still operating against the current design, policy, and scope? NAEOS is built around that control problem.

## 2. Why should I care?

Imagine an AI coding agent is asked to make a change. It uses a prior plan, a stale policy, and a local understanding of the project. Then the system requirements change: a module boundary is tightened, a security rule is added, or an authorization policy is revised.

The engineering question is simple: should the old plan still run?

NAEOS is built around that question. It treats the specification as the source of truth and evaluates validation and policy before generating artifacts.

## 3. See the public control boundary

Before running locally, you can inspect the public Control Plane at [naeos.dev/control-plane](https://naeos.dev/control-plane/). It evaluates authorization requests against the deployed NAEOS control-plane engine and is intentionally side-effect-free.

Use it to understand the public control boundary first; then reproduce the engineering workflow locally with the Golden Path below.

## 4. Run the Verified Golden Path

The canonical onboarding spine is [Verified Golden Path](docs/VERIFIED-GOLDEN-PATH.md). It connects the public control boundary, local Golden Path, evidence inspection, independent verification, governance experiment, and handoff experiment without creating a second implementation path.

Start with the public boundary:

- [NAEOS Control Plane](https://naeos.dev/control-plane/)

Then run the local proof:

```bash
go build -o naeos ./cmd/naeos
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

The underlying acceptance contract remains [docs/GOLDEN-PATH.md](docs/GOLDEN-PATH.md). The repository-level readiness checklist is [docs/READINESS-CONTRACT.md](docs/READINESS-CONTRACT.md). After the run, inspect the evidence under `/tmp/naeos-demo`, then use the read-only [P1.11 Independent Verifier](docs/control-plane/p1-11-independent-verifier-cli.md) for a canonical `EvidenceBundle`.

### What you just proved

A successful local run establishes:

```text
Specification
  → NEIR
  → Validation
  → Policy
  → AI context
  → Authorized generation
  → Artifacts
  → Traceable evidence
```

The Verified Golden Path adds independent verification and challenge-oriented experiments around that core proof. It does **not** by itself prove production readiness, customer adoption, enterprise compliance, or the safety of every external agent integration.

## 5. Run the governance experiment

After the five-minute CLI demo, run the flagship governance lifecycle experiment:

```bash
go run ./experiments/governance-lifecycle
```

It tests the chain `agent intent → policy decision → execution → observation → evidence → independent verification` and deliberately treats an agent's claim as different from observed evidence. See [experiments/governance-lifecycle/README.md](experiments/governance-lifecycle/README.md).

## 6. Understand the architecture

The repository’s conceptual architecture is summarized in [README.md](README.md) and [ARCHITECTURE-OVERVIEW.md](ARCHITECTURE-OVERVIEW.md). For normative authority and conflict resolution, use [DOCUMENTATION-AUTHORITY.md](DOCUMENTATION-AUTHORITY.md).

The short version is:

```text
Specification
  ↓
Parse → Normalize → Resolve
  ↓
Build NEIR
  ↓
Validate → Policy → AI context
  ↓
Generate artifacts and engineering outputs
```

If you want the authoritative references, use [DOCUMENTATION-AUTHORITY.md](DOCUMENTATION-AUTHORITY.md) first. For repository navigation and domain boundaries, see [REPOSITORY-ARCHITECTURE.md](REPOSITORY-ARCHITECTURE.md). Then:

- [NAEOS-NRA-001](Reference%20Architecture/NAEOS-NRA-001.md)
- [NAEOS-MTS-001](NAEOS-MTS-001.md)
- [specification/NAEOS-SPEC-001.md](specification/NAEOS-SPEC-001.md)
- [docs/NES-023-NEIR.md](docs/NES-023-NEIR.md)
- [docs/NES-028-CLI-Reference.md](docs/NES-028-CLI-Reference.md)

## 7. Pick your path

### I want to understand NAEOS

Start with:

- [README.md](README.md)
- [GETTING-STARTED.md](GETTING-STARTED.md)
- [specification/NAEOS-SPEC-001.md](specification/NAEOS-SPEC-001.md)
- [ARCHITECTURE-OVERVIEW.md](ARCHITECTURE-OVERVIEW.md)

### I want to run or verify NAEOS

Start with:

- [docs/VERIFIED-GOLDEN-PATH.md](docs/VERIFIED-GOLDEN-PATH.md)
- [docs/GOLDEN-PATH.md](docs/GOLDEN-PATH.md)
- [docs/REFERENCE-DEMO.md](docs/REFERENCE-DEMO.md)
- [docs/control-plane/p1-11-independent-verifier-cli.md](docs/control-plane/p1-11-independent-verifier-cli.md)
- [docs/EXTERNAL-VALIDATION.md](docs/EXTERNAL-VALIDATION.md)

Verified local path:

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

### I want to inspect the architecture

Read:

- [ARCHITECTURE-OVERVIEW.md](ARCHITECTURE-OVERVIEW.md)
- [docs/NES-023-NEIR.md](docs/NES-023-NEIR.md)
- [docs/NES-026-Pipeline.md](docs/NES-026-Pipeline.md)
- [docs/NES-027-Governance.md](docs/NES-027-Governance.md)

### I want to challenge the design

This is a valuable contribution. Ask questions such as:

- What assumption does the current pipeline make that fails in the real world?
- Where does stale context survive validation?
- Where does policy enforcement rely on incomplete metadata?
- What happens when authority or scope changes mid-run?

The repository’s discussion workflow is described in [docs/community/discussions.md](docs/community/discussions.md). A focused technical question is often the best way to start.

### I want to contribute code

Start with [CONTRIBUTING.md](CONTRIBUTING.md), then look for an issue or a concrete gap. The repository includes issue templates in [.github/ISSUE_TEMPLATE](.github/ISSUE_TEMPLATE) and a contributor ladder in [docs/community/contributor-ladder.md](docs/community/contributor-ladder.md).

### I want to contribute documentation

Good first documentation work includes clarifying the onboarding path, fixing broken or confusing links, and improving the examples used by new visitors. This is often the easiest first step.

### I want to experiment with AI agents

Look at the pipeline and compiler docs, the demo, and the AI context generation flow. The best concrete starting point is the local demo plus the compiler references in [README.md](README.md) and [docs/NES-028-CLI-Reference.md](docs/NES-028-CLI-Reference.md).

## 8. Start with a small contribution

The easiest successful contribution is not “build a large feature.” It is “improve one small, concrete thing that makes the project easier to understand or use.”

### 15 minutes

Examples of good early work:

- clarify a confusing section in onboarding docs
- identify a broken link or stale path
- reproduce a limitation and document the exact steps
- review a policy scenario and ask whether the behavior is consistent
- improve one example or one command snippet

### 1 hour

Examples:

- add or improve a test around a pipeline or validation behavior
- improve a minimal example or demo output
- document an edge case in the CLI workflow
- tighten a small piece of architecture or policy documentation

### Deeper contribution

These are real areas of the repository and are appropriate once the basics are understood:

- architecture and pipeline work
- validation and policy work
- NEIR and compiler behavior
- runtime and execution semantics
- migration or schema work
- plugin and profile integration
- security and evidence-oriented review

The contributor ladder in [docs/community/contributor-ladder.md](docs/community/contributor-ladder.md) is the best guide for how these contributions fit together.

## 9. Challenge NAEOS

A healthy NAEOS contribution is not just “add more features.” It is also challenging the assumptions the project currently makes.

The most valuable questions are about failure modes and edge cases:

- policy bypass or stale authorization
- inconsistent or delayed validation
- replay or stale artifact reuse
- authority changes mid-run
- handoff manipulation between tools or agents
- verification gaps or evidence gaps

This is not anti-project work. It is exactly the kind of work that makes a governance and AI engineering system more trustworthy.

If you see a gap, ask: “What assumption does NAEOS currently make that may not hold in the real world?”

## 10. Good first issues

The repository's current contributor/adoption work is tracked in [#447 — External contributor and adoption readiness](https://github.com/NAEOS-foundation/naeos/issues/447). Use it as the coordination point for onboarding gaps and bounded external-contributor work.

For a concrete contribution, search the repository's open issues for a scoped task, or open a focused report using [.github/ISSUE_TEMPLATE](.github/ISSUE_TEMPLATE). Prefer small changes that can be reproduced and reviewed independently.

If you are unsure where a finding belongs, start with [GitHub Discussions](docs/community/discussions.md) for a design/question discussion, or open the closest issue-template category with the exact reproduction steps and evidence.

## 11. Contribution workflow

The repository already defines the engineering workflow in [CONTRIBUTING.md](CONTRIBUTING.md). For the current external-contributor path, use this sequence:

1. Run the canonical Golden Path from this page.
2. Inspect the evidence and record the exact command, commit, and observed result.
3. Choose one small improvement, challenge, or documentation gap.
4. Report the finding through an issue or discussion before taking on a larger change.
5. Follow [CONTRIBUTING.md](CONTRIBUTING.md) for the implementation and PR requirements.

The simplest contribution flow is:

1. Clone the repository and read [CONTRIBUTING.md](CONTRIBUTING.md).
2. Pick a concrete issue, question, or documentation gap.
3. Make a small change with a clear explanation.
4. Run the relevant tests or validation commands.
5. Open a pull request and explain what changed and why.

Good first work includes clarifying onboarding, reproducing a limitation, improving an example, strengthening a test, or challenging a governance assumption. See [.github/ISSUE_TEMPLATE](.github/ISSUE_TEMPLATE) and [docs/community/contributor-ladder.md](docs/community/contributor-ladder.md).

## 12. Documentation map

| I want to... | Read... |
|---|---|
| Understand the project | [README.md](README.md), [GETTING-STARTED.md](GETTING-STARTED.md), [specification/NAEOS-SPEC-001.md](specification/NAEOS-SPEC-001.md) |
| Run the CLI | [docs/GOLDEN-PATH.md](docs/GOLDEN-PATH.md), [docs/REFERENCE-DEMO.md](docs/REFERENCE-DEMO.md), [GETTING-STARTED.md](GETTING-STARTED.md), [examples/demo-cli/README.md](examples/demo-cli/README.md), [docs/NES-028-CLI-Reference.md](docs/NES-028-CLI-Reference.md) |
| Understand the architecture | [ARCHITECTURE-OVERVIEW.md](ARCHITECTURE-OVERVIEW.md), [docs/NES-023-NEIR.md](docs/NES-023-NEIR.md), [docs/NES-026-Pipeline.md](docs/NES-026-Pipeline.md) |
| Understand governance | [constitution/NAEOS-CON-001.md](constitution/NAEOS-CON-001.md), [governance/NAEOS-GOV-001.md](governance/NAEOS-GOV-001.md) |
| Understand policy | [policy/NAEOS-POL-001.md](policy/NAEOS-POL-001.md) |
| Contribute | [CONTRIBUTING.md](CONTRIBUTING.md), [docs/community/contributor-ladder.md](docs/community/contributor-ladder.md) |
| See the roadmap | [ROADMAP.md](ROADMAP.md) |
| Join discussion | [docs/community/discussions.md](docs/community/discussions.md) |

## 13. Join the discussion

If you have a technical question, a design idea, or a project you built with NAEOS, use GitHub Discussions and the issue templates rather than silently watching. The repository already defines the community structure in [docs/community/discussions.md](docs/community/discussions.md).

A strong first discussion usually contains:

- a short problem statement
- the exact command or workflow you tried
- the actual output or failure mode
- the question you are asking

This is more useful than a vague “this seems broken” note.

## 14. The NAEOS principle

NAEOS should not merely make AI agents more capable. It should make their actions more understandable, governable, verifiable, and trustworthy.

That is the project’s real starting point: not a faster code generator, but a more disciplined engineering layer for AI-assisted work.

## AI-agent instruction trust boundary

Generated agent instructions such as `AGENTS.md` and `.opencode` are **advisory context, not an authorization boundary**. An AI agent with write access can alter or remove those files, so their contents must not be treated as tamper-evident policy evidence.

Consequential authorization is enforced by NAEOS governance and runtime controls, not by instruction-file presence or wording. In particular:

- policy evaluation determines whether a requested action is permitted;
- the control plane fails closed when no policy matches;
- runtime authorization remains separate from agent-provided instructions;
- evidence and independent verification should rely on recorded execution/evidence artifacts rather than an agent's claim that its instruction file was followed.

If tamper-evidence for instruction content is required in a future workflow, it must be introduced as a separately specified integrity mechanism; this onboarding guide does not imply that such a mechanism currently exists.
