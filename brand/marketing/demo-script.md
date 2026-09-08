# NAEOS Marketing Demo Script

## Goal
Show that NAEOS is not just a code generator. It is a declarative engineering platform that turns specifications into validated, AI-ready systems.

## Core narrative
"AI can generate code. NAEOS helps structure the engineering system around that code."

## Recommended durations
- 45 seconds: teaser / social clip
- 2 minutes: pitch / deck
- 5 minutes: technical demo

---

## 45-second teaser

### Hook
"The real problem with AI-assisted development is not code generation. It is engineering consistency."

### Script
"NAEOS starts with one specification and turns it into a shared engineering model."

"It parses the spec, normalizes it, resolves dependencies, builds NEIR, validates it, and generates artifacts and AI context."

"That means teams are not managing scattered prompts and drifted implementation. They are working from one source of truth."

"Specify once. Build anywhere."

### CTA
"Explore the NAEOS repository and try the quick start."

---

## 2-minute pitch demo

### On-screen flow
1. Title: "One specification. One engineering model."
2. YAML specification on left
3. Pipeline: Parse → Normalize → Resolve → NEIR → Validate → Generate → AI Context
4. Output: generated project, docs, manifests, and AI bundles

### Voiceover
"NAEOS is a declarative engineering platform. A team defines the system once and NAEOS turns that into a consistent engineering model."

"It reads the specification, validates the structure, builds a NEIR model, and produces artifacts that are traceable throughout the lifecycle."

"This is different from simple scaffolding. NAEOS treats engineering intent as a structured model, not just a template."

"When AI participates in the workflow, that matters. Teams need context, governance, and consistent architecture — not more prompt drift."

"NAEOS helps compile that context for GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode, and Windsurf."

### Closing line
"NAEOS brings structure, validation, and governance to AI-assisted software development."

---

## 5-minute technical demo

### Scenario
Use a simple application with modules: `auth`, `api`, and `gateway`.

### Step 1: Start from a spec

```yaml
project: my-app
modules:
  - name: auth
    path: ./auth
  - name: api
    path: ./api
    dependencies: [auth]
services:
  - name: gateway
    kind: http
    port: 8080
architecture:
  pattern: hexagonal
generation:
  languages: [go, typescript]
```

Narration:
"This is the source of truth. The team defines the system once, and NAEOS carries that intent through the entire engineering pipeline."

### Step 2: Show the internal pipeline
"NAEOS parses the spec, normalizes the structure, resolves dependencies, and builds the NEIR model."

Show:
- Parser
- Normalizer
- Resolver
- NEIR
- Validator
- Scheduler
- Generator

### Step 3: Validate before generation
Narration:
"Before emitting any output, NAEOS checks the shape of the project, dependency integrity, policy constraints, and architecture rules."

Show example warnings or validation messages:
- no circular dependency
- no conflicting ports
- module boundaries respected

### Step 4: Generated output
Show generated files such as:
- `auth/`
- `api/`
- deployment manifests
- CI/CD config
- documentation
- AI instruction bundle

Narration:
"The result is not just boilerplate. It is a generated system aligned to the original specification."

### Step 5: AI context output
Show generated context bundles for:
- GitHub Copilot
- Claude Code
- Cursor
- Gemini CLI
- Codex
- OpenCode
- Windsurf

Narration:
"NAEOS compiles engineering context so AI tools work from a consistent model, not from ad hoc prompts."

### Closing
"This is the difference between generating code and building an engineering system."

---

## Short quotes for slides

- "AI can generate code. NAEOS structures the engineering system around it."
- "Specify once. Build anywhere."
- "NAEOS turns intent into a validated, reusable engineering model."
- "From specification to AI context, governance, and generated artifacts."
- "The bottleneck in AI-assisted development is not generation — it is consistency."

---

## Presenter closing CTA
- Try NAEOS
- Explore GitHub
- Read the architecture
- Build your first specification

---

## Script for presenter

"What we are showing here is a different model for AI-assisted software development. Instead of treating AI as a code generator disconnected from the system, we treat engineering intent as a structured specification. NAEOS reads that spec, validates it, builds a normalized engineering model, and then generates the project artifacts and AI context that downstream tools need. The result is a more consistent and governable system, from initial design through implementation and evolution."
