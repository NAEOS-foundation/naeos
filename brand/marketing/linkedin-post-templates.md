# LinkedIn Post Templates

## Template 1 — “Why AI coding needs a specification layer”

Hook:
AI can generate code quickly. But code alone does not define the system.

Body:
We built NAEOS because engineering teams need more than fast generation. They need a shared model of intent.

NAEOS is a declarative engineering platform. It lets teams describe a system once, normalize it, validate it, and compile AI context and generated artifacts from the same source of truth.

That means fewer mismatches between architecture, implementation, and AI tooling.

For teams working with AI-assisted development, structure matters as much as speed.

CTA:
Explore the repo and try the quick start.

## Template 2 — “One source of truth”

Hook:
A lot of developer tooling optimizes for generation.

Body:
NAEOS is designed for the layer underneath generation: the engineering model itself.

It combines specification, validation, governance, AI context compilation, and artifact generation into one pipeline.

Instead of keeping architecture in docs, prompts, tickets, and code, NAEOS makes the model explicit and reusable.

That is the difference between generating code and engineering a system.

CTA:
Read the architecture and try the quick start.

## Template 3 — “From spec to system”

Hook:
The real challenge in AI-assisted development is not making code faster.

Body:
It is keeping the system coherent as it evolves.

NAEOS treats the specification as the source of truth. From there, it normalizes the model, resolves dependencies, validates the design, and generates downstream artifacts and AI context.

That gives teams a clearer path from intent to implementation.

Specify once. Build anywhere.

CTA:
See the repo and open the architecture docs.

## Template 4 — “We test our own guardrails” (governance/security story)

Hook:
We run an adversarial harness against NAEOS' own governance, in public.

Body:
We probe the enforcement layers — policy evaluation, control plane, artifact
review, prompt generation, and pipeline gates — the way an attacker would.
The findings are published as open issues, and critical bypasses are closed
before distribution.

This is grounded in the repository: `experiments/policy-bypass/` (17 scenarios),
the CI-generated findings report, and the issues that track remaining work.

Building a governance platform means proving it can be broken — repeatedly,
and in public.

CTA:
See the scan + the findings: github.com/NAEOS-foundation/naeos

Notes:
- Every number (scenarios, findings, adapters) must be read from the current
  report before posting — do not reuse stale counts.
