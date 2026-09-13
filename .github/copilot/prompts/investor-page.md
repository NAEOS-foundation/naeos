# NAEOS Investor Page — Implementation Instructions

You are working inside the NAEOS repository:

https://github.com/NAEOS-foundation/naeos

Build a dedicated **Investor Page** for NAEOS.

The page must position NAEOS as:

> **The Engineering Control Plane for the AI-Native Software Era.**

Primary message:

> **AI coding agents generate code. NAEOS provides the engineering control plane around them.**

Tagline:

> **Specify once. Build anywhere.**

---

# 1. Repository-First Requirement

Before implementing anything, inspect the current repository.

Read and understand:

* README.md
* WHITEPAPER.md
* WHITEPAPER-EN.md
* ARCHITECTURE-OVERVIEW.md
* ROADMAP.md
* DEVELOPMENT_PLAN.md
* MARKETING-STRATEGY.md
* DOCUMENTATION-INDEX.md
* current site implementation
* brand assets
* existing website components
* existing documentation
* existing screenshots/demo assets
* current NAEOS technical capabilities

Do not invent technical capabilities that do not exist.

Use the GitHub repository as the primary source of truth.

If the existing website already contains reusable components, styling, navigation, typography, or layout systems, reuse them.

Do not create a second unrelated design system.

---

# 2. Investor Page Objective

The page has one primary goal:

> Convert technically sophisticated investors into qualified conversations with the NAEOS founder.

The page should make an investor understand NAEOS within approximately 60–90 seconds.

The investor should quickly understand:

1. What NAEOS is.
2. What problem it solves.
3. Why the problem exists now.
4. Why existing AI coding tools do not fully solve it.
5. How NAEOS works.
6. Why the architecture matters.
7. Why NAEOS can become infrastructure.
8. Current development stage.
9. Business model.
10. Fundraising objective.
11. Why this is a venture-scale opportunity.
12. How to contact the founder.

---

# 3. Positioning

Use this positioning consistently:

## Primary positioning

> NAEOS is the engineering control plane for the AI-native software era.

## Supporting statement

> AI coding agents make software implementation dramatically faster. NAEOS provides the persistent engineering model, validation, policy, governance, context, and execution layer around those agents.

## Avoid positioning NAEOS as:

* another AI coding assistant
* another Copilot alternative
* another code generator
* another prompt framework
* another project scaffolding tool
* an AI model
* a generic workflow automation platform

The page must clearly communicate that NAEOS operates at the **engineering-system/control-plane layer**.

---

# 4. Page Structure

Implement the following sections.

---

## Section 1 — Hero

Create a strong investor-focused hero.

Headline:

> The Engineering Control Plane for the AI-Native Software Era.

Subheadline:

> NAEOS transforms engineering specifications into validated, governed, AI-ready execution workflows.

Supporting line:

> Specify once. Build anywhere.

Primary CTA:

> Request Investor Briefing

Secondary CTA:

> Explore the Architecture

Optional tertiary CTA:

> View GitHub

The hero should immediately communicate that this is infrastructure, not another coding assistant.

---

# 5. Hero Visual

Create a technical visual showing:

```text
                    NAEOS
          Engineering Control Plane
                       │
       ┌───────────────┼───────────────┐
       │               │               │
 Specification     Governance       AI Agents
       │               │               │
       └───────────────┼───────────────┘
                       │
                      NEIR
                       │
       ┌───────────────┼───────────────┐
       │               │               │
 Validation        Context          Execution
       │               │               │
       └───────────────┼───────────────┘
                       │
               Software Systems
```

Use the actual NAEOS architecture where possible.

Do not create decorative diagrams that contradict the repository.

---

# 6. Section 2 — The AI Coding Revolution Has a Missing Layer

Headline:

> AI made coding faster. Engineering systems did not.

Explain the problem.

Show the current environment:

```text
Human
  ↓
Specification
  ↓
AI Coding Agent
  ↓
Code
```

Then show the problem:

```text
Context fragmentation
Architecture drift
Policy violations
Dependency problems
Inconsistent AI instructions
Limited auditability
Weak engineering memory
```

Explain that increasing agent capability increases the importance of system-level control.

---

# 7. Section 3 — The NAEOS Thesis

Headline:

> The next layer of software engineering is not another agent.

Explain:

> The future is an engineering system coordinating humans, AI agents, infrastructure, policies, and organizational knowledge.

Then present:

```text
Human Intent
     ↓
Specification
     ↓
NEIR
     ↓
Validation
     ↓
Policy
     ↓
AI Context
     ↓
AI Agents
     ↓
Execution
     ↓
Artifacts + Evidence
```

Highlight:

> NAEOS is designed to become that coordination layer.

---

# 8. Section 4 — How NAEOS Works

Headline:

> From engineering intent to validated execution.

Use the existing NAEOS pipeline.

Show:

```text
Specification
      ↓
Parse
      ↓
Normalize
      ↓
Resolve
      ↓
NEIR
      ↓
Validate
      ↓
Schedule
      ↓
Generate
      ↓
AI Context
      ↓
Governance
      ↓
Artifacts
```

For each major stage, provide a concise explanation.

Do not overload the investor with implementation details.

The visual should communicate that NAEOS is a system, not simply a prompt wrapper.

---

# 9. Section 5 — NEIR

Headline:

> NEIR is the persistent engineering model.

Explain:

> NAEOS Engineering Intermediate Representation gives engineering intent a structured, machine-processable representation that can be validated, governed, transformed, documented, and supplied to AI tooling.

Show:

```text
Specification
      ↓
     NEIR
      ↓
 ┌────┼────┬────┬────┐
 ↓    ↓    ↓    ↓    ↓
Validate Policy Context Generate Audit
```

This section should be one of the strongest technical moat sections.

---

# 10. Section 6 — Why NAEOS Is Different

Create a concise comparison.

Columns:

```text
                    AI Coding Tools
                    Project Generators
                    Workflow Tools
                    NAEOS
```

Rows:

* Persistent engineering model
* Specification as source of truth
* NEIR
* Deterministic validation
* Policy enforcement
* AI context compilation
* Multi-agent interoperability
* Governance
* Auditability
* Artifact traceability

Do not attack competitors.

The purpose is to show category differentiation.

---

# 11. Section 7 — Control Plane Architecture

Headline:

> AI agents execute. NAEOS governs the system around them.

Create a visual architecture:

```text
                 HUMAN
                   │
                   ▼
            SPECIFICATION
                   │
                   ▼
                  NEIR
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
   VALIDATION    POLICY     CONTEXT
        │          │          │
        └──────────┼──────────┘
                   ▼
              AI AGENTS
                   │
                   ▼
               EXECUTION
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
     ARTIFACTS   AUDIT     EVIDENCE
```

Important principle:

> Prompts are not security boundaries.

Explain that consequential actions should be evaluated through deterministic system controls rather than relying on model behavior.

---

# 12. Section 8 — Why Now

Headline:

> AI agents are becoming capable faster than engineering systems are adapting.

Explain the convergence of:

* increasingly capable coding agents
* AI-assisted software development
* autonomous development workflows
* MCP and tool interoperability
* infrastructure automation
* increasing software complexity
* enterprise requirements for governance
* security and compliance requirements
* multi-agent workflows

The conclusion:

> As AI agents become more autonomous, the control plane around them becomes increasingly important.

Do not make unsupported market-size claims.

---

# 13. Section 9 — Market Opportunity

Do not fabricate TAM numbers.

Instead establish the category logically:

```text
AI Coding
      ↓
AI-Native Development
      ↓
AI Engineering Systems
      ↓
Engineering Control Plane
```

Explain the potential market expansion from:

* developer tooling
* AI developer infrastructure
* software engineering platforms
* enterprise engineering governance
* AI agent infrastructure

If verified market data exists elsewhere in the repository, use it.

Otherwise mark market numbers as requiring external validation.

---

# 14. Section 10 — Business Model

Headline:

> Open source for adoption. Control plane for monetization.

Explain the potential model:

```text
Open Source Core
       ↓
Developer Adoption
       ↓
Team Usage
       ↓
NAEOS Control Plane
       ↓
 ┌─────┼──────┬─────────┐
 ↓     ↓      ↓         ↓
Policy Governance Audit Enterprise
                     Integrations
```

Potential monetization areas:

* centralized governance
* policy management
* collaboration
* audit
* compliance
* managed infrastructure
* enterprise integrations
* organizational engineering state

Clearly distinguish:

**Current implementation**

from

**Future commercial opportunities**.

Never present planned monetization as current revenue.

---

# 15. Section 11 — Current Technical Foundation

Show what already exists in the repository.

Use actual capabilities discovered during repository inspection.

Potential categories include:

* Declarative specification
* NEIR
* Compiler
* Validation
* Governance
* Policy
* AI context
* MCP
* Artifacts
* Audit
* Profiles
* Migration
* Marketplace
* Plugins
* WASM integration
* Testing
* Documentation generation
* CLI
* Pipeline caching

Only display capabilities that are actually present.

Add:

> Built in public. Designed as open infrastructure.

Link to the GitHub repository.

---

# 16. Section 12 — Product Wedge

Headline:

> One workflow. One control plane.

Show the initial wedge:

```text
Specification
      ↓
NEIR
      ↓
Validation
      ↓
Policy
      ↓
AI Context
      ↓
Agent Execution
      ↓
Working Application
```

Explain:

> The immediate product objective is to make this workflow reliable, repeatable, and fast enough to become a developer's normal engineering loop.

This is more important than adding more features.

---

# 17. Section 13 — Traction

Do NOT invent traction.

Create a section that supports actual metrics.

Potential metrics:

```text
GitHub stars
Contributors
Forks
Commits
Releases
Developers using NAEOS
Design partners
Production deployments
Paying customers
Integrations
```

Only display metrics that can be verified.

If a metric is unavailable:

Do not display a fake number.

Instead use wording such as:

> Building in public.

or

> Seeking early design partners.

---

# 18. Section 14 — Roadmap

Use the actual repository roadmap.

Organize into:

### Now

Current engineering foundation and product wedge.

### Next

Focused AI engineering control-plane workflow.

### Later

Commercial control plane, team governance, enterprise capabilities, and broader ecosystem.

Do not promise dates unless dates exist in the official roadmap.

---

# 19. Section 15 — Fundraising

Headline:

> Building the control plane for AI-native engineering.

If the current fundraising plan is still $500K pre-seed, display:

> Raising: $500K Pre-Seed

Explain intended allocation:

```text
40% Product & Infrastructure
30% Developer Adoption
20% GTM
10% Operations
```

Clearly label this as a proposed allocation.

Primary objective:

> Convert the existing technical foundation into a focused developer wedge, measurable adoption, design partnerships, and early commercial validation.

Do not present proposed targets as existing traction.

---

# 20. Section 16 — 12-Month Objectives

Use a clearly labeled:

> 12-Month Targets

Potential targets:

```text
10–20 design partners
100+ active engineering teams
10+ paying organizations
3–5 enterprise pilots
8–10 AI/developer-tool integrations
50+ external contributors
10+ production deployments
```

These are targets, not current results.

Do not use language such as:

> We already have...

unless verified.

---

# 21. Section 17 — Investor Fit

Headline:

> Who should invest in NAEOS?

Target investors interested in:

* developer tools
* AI infrastructure
* engineering infrastructure
* open-source infrastructure
* enterprise software
* developer platforms
* AI-native workflows
* technical founders

Do not create a long investor directory on the public page.

The page should instead communicate investor thesis alignment.

---

# 22. Section 18 — Founder

Create a concise founder section.

Focus on:

* technical founder journey
* why NAEOS exists
* building in public
* long-term infrastructure thesis

Use only factual information available in the repository or existing official founder materials.

Do not invent credentials, previous companies, funding, or achievements.

---

# 23. Section 19 — Investor CTA

Create a strong closing section.

Headline:

> Help build the engineering control plane for the AI-native era.

Copy:

> NAEOS is building infrastructure for a world where software is increasingly created by humans and AI agents together. We are looking for investors and strategic partners who understand developer infrastructure, AI systems, and the next generation of software engineering.

CTA:

> Request Investor Briefing

Secondary:

> Start a Conversation

Optional:

> View GitHub

---

# 24. Investor Briefing CTA

The primary CTA should lead to a real contact mechanism.

Inspect the repository for an existing:

* contact form
* email
* investor form
* Calendly
* contact page

Reuse the existing mechanism.

Do not invent a fake email address.

If no investor contact mechanism exists, create a clean contact section that can later be connected to the actual destination.

---

# 25. Visual Design

The visual language should communicate:

* infrastructure
* engineering
* AI systems
* technical credibility
* precision
* long-term platform ambition

Avoid:

* generic AI robot imagery
* excessive gradients
* stock photography
* crypto-style aesthetics
* fake dashboards
* meaningless 3D graphics
* excessive animation
* hype-driven startup visuals

Prefer:

* architecture diagrams
* system graphs
* structured typography
* code/specification snippets
* pipeline visualization
* subtle motion
* technical UI patterns
* restrained color palette
* generous whitespace

The page should feel closer to:

> infrastructure platform + developer tooling

than:

> generic AI startup landing page.

---

# 26. Animation

Use animation only when it communicates architecture.

Good examples:

```text
Specification
    ↓
NEIR
    ↓
Validation
    ↓
AI Context
    ↓
Agent
```

Animate the flow subtly.

Do not animate every component.

Respect:

```text
prefers-reduced-motion
```

---

# 27. Responsive Design

The page must work properly on:

* desktop
* tablet
* mobile

Architecture diagrams must remain understandable on mobile.

Do not simply shrink desktop diagrams.

Create responsive versions where necessary.

---

# 28. Performance

Investor pages must load quickly.

Optimize:

* images
* SVG
* fonts
* JavaScript
* animation
* third-party scripts

Avoid unnecessary dependencies.

Do not introduce a large UI framework if the existing site does not use one.

---

# 29. SEO

Add appropriate metadata.

Suggested title:

> NAEOS — The Engineering Control Plane for the AI-Native Software Era

Suggested description:

> NAEOS is an open engineering platform that transforms software specifications into validated, governed, AI-ready workflows.

Use appropriate:

* Open Graph
* Twitter/X metadata
* canonical URL
* structured metadata where appropriate

Do not keyword-stuff.

---

# 30. Analytics

If the existing site has analytics, integrate the investor CTA with the existing analytics system.

Track events such as:

```text
investor_page_view
investor_briefing_click
github_click
architecture_click
contact_click
```

Reuse the existing analytics architecture.

Do not introduce another analytics provider unless necessary.

---

# 31. Accessibility

Follow WCAG-oriented practices.

Ensure:

* semantic HTML
* keyboard navigation
* sufficient contrast
* accessible buttons
* meaningful alt text
* reduced motion support
* logical heading hierarchy
* mobile readability

---

# 32. Content Rules

Every claim must be categorized as one of:

### Current

Already implemented and verifiable.

### Target

Future objective.

### Vision

Long-term direction.

Do not mix these categories.

Avoid claims such as:

* “industry standard”
* “revolutionary”
* “10x”
* “millions of developers”
* “enterprise-ready”

unless they can be substantiated.

---

# 33. Competitive Positioning

Do not build a hostile competitor comparison.

The core distinction should be:

```text
AI Coding Tools
        ↓
Generate / Modify Code

NAEOS
        ↓
Model
Validate
Govern
Contextualize
Coordinate
Audit
Execute
```

The message:

> NAEOS does not need to replace AI coding agents. It can become the engineering system that coordinates them.

---

# 34. Technical Credibility

Include at least one section that lets a technical investor inspect the architecture.

CTA:

> Explore the Architecture

Link to the most relevant existing NAEOS architecture documentation.

Do not create fake diagrams that cannot be reconciled with the repository.

---

# 35. GitHub CTA

Use the official repository:

https://github.com/NAEOS-foundation/naeos

CTA:

> View NAEOS on GitHub

Make it obvious that the project is open source and being built publicly.

---

# 36. Implementation Rules

Before coding:

1. Inspect the current website.
2. Identify the framework.
3. Identify routing.
4. Identify components.
5. Identify design tokens.
6. Identify typography.
7. Identify existing responsive patterns.
8. Identify existing navigation/footer.
9. Identify existing analytics.
10. Identify existing CTA/contact implementation.

Then implement the investor page using the existing system.

Do not rebuild the website.

---

# 37. Suggested Route

Prefer:

```text
/investors
```

or:

```text
/investor
```

Use whichever routing convention already exists in the repository.

Do not introduce both.

---

# 38. Testing

Verify:

### Functional

* page loads
* navigation works
* CTA works
* GitHub link works
* architecture links work
* contact flow works

### Responsive

Test:

* desktop
* tablet
* mobile

### Accessibility

Run existing accessibility checks if available.

### Build

Run the site's existing build command.

### Tests

Run the existing test suite.

Do not stop at visual implementation.

---

# 39. Final Verification

After implementation, review the page as three different people.

## Investor

Can I understand:

> What is NAEOS?

within 10 seconds?

Can I understand:

> Why is this important?

within 30 seconds?

Can I understand:

> Why can this become a large company?

within 90 seconds?

## Technical Founder

Can I understand:

> What is actually being built?

Can I inspect the architecture?

## Potential Customer

Can I understand:

> What would NAEOS do for my engineering team?

If any answer is unclear, improve the page.

---

# 40. Definition of Done

The implementation is complete only when:

* [ ] `/investors` or equivalent route exists.
* [ ] Existing website design system is reused.
* [ ] Hero clearly communicates the control-plane thesis.
* [ ] Problem is clearly explained.
* [ ] NAEOS architecture is visually explained.
* [ ] NEIR is clearly positioned.
* [ ] AI-agent relationship is clear.
* [ ] Governance/policy boundary is explained.
* [ ] Current technical foundation is shown accurately.
* [ ] Business model is clearly separated from future plans.
* [ ] Traction contains only verified information.
* [ ] Targets are explicitly labeled as targets.
* [ ] Fundraising information is accurate.
* [ ] Investor CTA works.
* [ ] GitHub CTA works.
* [ ] Mobile layout works.
* [ ] Accessibility is acceptable.
* [ ] SEO metadata exists.
* [ ] Existing tests pass.
* [ ] Existing build passes.
* [ ] No unsupported claims were introduced.

---

# Final Product Principle

The page should make one idea unforgettable:

> **AI coding agents generate code. NAEOS provides the engineering control plane around them.**

And the investor should leave with this mental model:

```text
          AI Coding Agents
                 │
                 ▼
        ┌─────────────────┐
        │      NAEOS      │
        │ Engineering     │
        │ Control Plane   │
        └─────────────────┘
                 │
       ┌─────────┼─────────┐
       ▼         ▼         ▼
    Humans    Policies   Systems
       │         │         │
       └─────────┼─────────┘
                 ▼
        AI-Native Engineering
```

Build the page around this thesis.

Do not optimize for hype.

Optimize for **technical credibility, investor clarity, and evidence**.
