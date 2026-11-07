# NAEOS Investor / Partner Deck — Outline

Bahasa Indonesia, 10 slides. This outline is paired with the presentation brief in [NAEOS-PRESENTATION-SIDE-BY-SIDE.md](NAEOS-PRESENTATION-SIDE-BY-SIDE.md).

## Slide 1 — AI-native engineering, with explicit control
**Message:** NAEOS is building an engineering control plane for AI coding agents.
**Narrative:** AI-assisted coding needs more than code generation: it needs explicit intent, governance boundaries, observable execution, and reviewable evidence.

## Slide 2 — Not another coding assistant
**Contrast:** Without a control plane: Prompt → Agent → Code/Action → unknowns. With NAEOS: Specification → Policy → Execution → Evidence.
**Source:** [README](../README.md).

## Slide 3 — Generation is not the whole problem
Introduce four questions: intent, authority, observability, and evidence.
**Caveat:** These are engineering-system concerns; do not imply all agents or external integrations are automatically safe.

## Slide 4 — One inspectable engineering flow
**Diagram:** Specification → NEIR → Validation + Policy → Context / Generation → Evidence.
**Sources:** [README](../README.md), [Architecture](../site/content/docs/architecture.md), [Glossary](../GLOSSARY.md).

## Slide 5 — Intent is not authorization
Separate intent, policy decision, execution, observation, and independent verification.
**Source:** [README trust model](../README.md).

## Slide 6 — Open the artifacts, not just the slides
Show the reference demo evidence: `spec.yaml`, `inspect.json`, `validate.json`, `invalid-policy.log`, `run.json`, `generated/`, and `summary.md`.
**Source:** [Reference Demo & Evidence Story](REFERENCE-DEMO.md).

## Slide 7 — Reproducible live demo
```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build -o naeos ./cmd/naeos
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
find /tmp/naeos-demo -maxdepth 3 -type f | sort
```
Record the repository commit SHA, Go toolchain version, command, exit status, and output evidence. Follow the current demo documentation if prerequisites change.

## Slide 8 — Start with a measurable partner pilot
Four steps: choose one use case, define boundaries, run the demo/pilot, evaluate against agreed acceptance criteria.
**Source:** [Reference Demo](REFERENCE-DEMO.md), [External Validation](EXTERNAL-VALIDATION.md).

## Slide 9 — Stage and priorities
Frame priorities as a proposed sequence, not as achieved traction: strengthen technical evidence; validate with design partners; test commercial model and willingness to pay.
Targets are future goals unless supported by measured results.

## Slide 10 — Call to action
Invite the audience to evaluate one real engineering workflow with explicit boundaries and jointly reviewable evidence.

## Claim discipline

A successful local reference run supports claims about the tested workflow and its observable outputs. It does not, by itself, prove production readiness, customer adoption, enterprise compliance, performance at scale, or the safety of every external agent integration.

## Source of truth
- [Repository README](../README.md)
- [Architecture](../site/content/docs/architecture.md)
- [Reference Demo](REFERENCE-DEMO.md)
- [Golden Path](GOLDEN-PATH.md)
- [External Validation](EXTERNAL-VALIDATION.md)
