NAEOS Repository Legal & License Audit Instructions

Mission

Perform a comprehensive, read-only legal and licensing audit of the NAEOS repository.

Repository:
https://github.com/NAEOS-foundation/naeos

Primary licensing assumption:

- NAEOS Core is intended to use Apache License 2.0.
- Do NOT modify any repository files during this audit.
- Do NOT change LICENSE, NOTICE, package manifests, source files, or contributor documentation.
- Produce findings and recommendations first.
- Any remediation must require explicit approval after the audit.

The goal is to determine whether the current repository is legally and operationally ready for:

1. Apache License 2.0 open-source distribution
2. External contributors
3. Commercial adoption
4. Enterprise customers
5. Future NAEOS Cloud / Enterprise products
6. Founder and investor due diligence
7. Future trademark and IP protection

---

1. Audit Principles

Follow these principles:

- Treat repository evidence as the primary source.
- Never assume ownership of code without evidence.
- Never assume a dependency license is compatible without checking it.
- Distinguish copyright, patent rights, trademark rights, and contractual rights.
- Distinguish open-source code from proprietary/commercial components.
- Flag uncertainty explicitly.
- Never provide a legal conclusion where repository evidence is insufficient.
- Separate technical findings from legal recommendations.
- Do not silently fix problems.
- Do not rewrite existing licensing language during the audit.

Use the following severity levels:

CRITICAL
HIGH
MEDIUM
LOW
INFO

---

2. Repository Inventory

First create a complete repository inventory.

Inspect:

- root files
- source directories
- packages
- modules
- libraries
- CLI
- runtime
- kernel
- policy
- governance
- specification
- infrastructure
- prompts
- tools
- examples
- documentation
- tests
- generated files
- configuration files
- scripts
- CI/CD workflows
- Dockerfiles
- deployment manifests
- lockfiles
- package manifests
- vendored code
- bundled third-party assets

Do not assume directories are correctly categorized.

Create a table:

Path| Component| Type| First-party/Third-party/Unknown| License Evidence| Risk

---

3. LICENSE Audit

Inspect:

- LICENSE
- LICENSE.txt
- LICENSE.md
- NOTICE
- COPYING
- README files
- package metadata
- repository metadata
- source headers
- documentation headers
- generated artifacts

Determine:

1. Is Apache License 2.0 clearly present?
2. Is the license text complete and unmodified?
3. Are there conflicting licenses?
4. Are there files without clear licensing information?
5. Are there directories containing different licensing terms?
6. Are examples licensed separately?
7. Are documentation and source code treated consistently?
8. Are generated files covered?
9. Are scripts covered?
10. Are prompt files covered?
11. Are configuration files covered?

Report every conflicting or ambiguous license.

---

4. SPDX and Copyright Headers

Inspect source files for:

- SPDX-License-Identifier
- copyright notices
- author attribution
- license headers
- generated-file notices

Identify inconsistent patterns.

Report:

- missing headers
- incorrect SPDX identifiers
- contradictory license declarations
- obsolete copyright notices
- suspicious ownership claims
- files claiming a license different from the repository license

Do NOT automatically add headers.

---

5. Third-Party Dependency Audit

Identify every third-party dependency.

For each dependency determine:

- package name
- version
- ecosystem
- direct/transitive
- license
- copyright obligations
- attribution requirements
- NOTICE requirements
- patent implications
- compatibility with Apache-2.0
- whether source redistribution is required
- whether modifications trigger additional obligations

Inspect:

- package.json
- package-lock.json
- pnpm-lock.yaml
- yarn.lock
- requirements.txt
- pyproject.toml
- poetry.lock
- Cargo.toml
- Cargo.lock
- go.mod
- go.sum
- Gemfile
- Gemfile.lock
- Maven/Gradle files
- Docker dependencies
- GitHub Actions
- vendored dependencies
- downloaded binaries
- embedded libraries

Do not rely solely on package metadata.

When possible, verify licenses from the upstream project.

Flag especially:

- GPL
- AGPL
- LGPL
- SSPL
- source-available licenses
- non-commercial licenses
- proprietary licenses
- custom licenses
- unknown licenses

Pay particular attention to copyleft dependencies that may affect distribution.

---

6. License Compatibility Matrix

Create a compatibility matrix:

Dependency| License| Apache-2.0 Compatible?| Modification Risk| Distribution Risk| Action

Use:

SAFE
REVIEW
BLOCKER

Do not classify a license as SAFE merely because it is commonly used.

Explain the reasoning.

---

7. Trademark Audit

Search the repository for:

- NAEOS
- NAEOS Foundation
- NAEOS logo references
- product names
- domain names
- organization names
- third-party trademarks

Determine whether the repository accidentally implies:

- trademark permission
- endorsement
- affiliation
- ownership
- certification

Apache-2.0 does NOT grant trademark rights.

Recommend explicit trademark language where appropriate.

Do not claim that the NAEOS trademark is legally registered unless repository evidence proves it.

---

8. Patent / Apache-2.0 Audit

Inspect the repository for:

- patent notices
- patent grants
- patent-related documentation
- contributor agreements
- third-party patent licenses
- known patented technology references

Confirm that Apache-2.0 patent language is not accidentally removed or contradicted.

Explain:

- what Apache-2.0 provides
- what it does not provide
- potential contributor patent implications

Do not make unsupported claims about patent ownership.

---

9. Contributor Audit

Inspect:

- CONTRIBUTING.md
- CODE_OF_CONDUCT.md
- Developer Certificate of Origin
- CLA
- CLA Assistant configuration
- GitHub workflows
- contributor documentation
- pull request templates

Determine whether NAEOS currently has:

- DCO
- CLA
- individual CLA
- corporate CLA
- no contribution agreement

Assess whether the current contribution model is sufficient for an open-source project intending to build commercial products.

Recommend whether NAEOS should use:

- DCO
- CLA
- both
- neither

Explain trade-offs.

Do not implement changes.

---

10. Contributor IP Ownership

Determine whether external contributions create ambiguity around:

- copyright ownership
- patent rights
- redistribution rights
- relicensing rights
- commercial use
- future dual licensing

Flag any missing mechanism for confirming contribution rights.

Pay special attention to contributions from:

- employees
- contractors
- freelancers
- agencies
- external developers
- AI-assisted contributors

Do not assume employer ownership without contractual evidence.

---

11. AI-Generated Code Audit

Search for evidence of:

- AI-generated code
- generated documentation
- generated prompts
- generated configuration
- generated assets
- AI-assisted commits

Do not assume AI-generated material is automatically copyright-free or automatically owned by NAEOS.

Instead classify:

KNOWN
POSSIBLE
UNKNOWN

Identify where provenance should be documented.

Recommend a lightweight provenance policy for future contributors.

---

12. Prompt and Specification Audit

Because NAEOS is an AI engineering platform, specifically inspect:

- prompts
- system prompts
- agent definitions
- specification files
- policy definitions
- templates
- instruction files
- examples

Determine whether these are:

- software
- documentation
- configuration
- potentially proprietary know-how
- third-party material
- unknown-origin content

Check whether their licensing is clear.

Flag prompts or specifications that contain:

- copied third-party content
- proprietary-looking material
- customer information
- secrets
- API keys
- credentials
- personal information

Never expose discovered secrets in the audit report.

---

13. Secrets and Sensitive Information

Perform a security-oriented scan for:

- API keys
- access tokens
- passwords
- private keys
- certificates
- cloud credentials
- database credentials
- personal information
- customer data
- internal URLs
- sensitive infrastructure configuration

If secrets are discovered:

- DO NOT print the secret value.
- Report only file path and secret type.
- Mark severity CRITICAL.
- Recommend rotation/revocation.

Do not commit remediation automatically.

---

14. Copyright and Attribution Audit

Search for:

- copied code
- copied documentation
- copied examples
- copied comments
- third-party snippets
- attribution notices
- source references

Identify suspicious blocks where attribution may be required.

Do not accuse contributors of infringement.

Use neutral language:

"Potential third-party provenance requiring verification."

---

15. Documentation License Audit

Inspect:

- README
- docs/
- architecture documents
- tutorials
- examples
- blog/article content
- diagrams
- API documentation

Determine whether documentation has:

- explicit license
- implicit repository license
- separate license
- unknown license

Recommend whether documentation should remain under Apache-2.0 or use a separate documentation license.

---

16. Asset Audit

Inspect:

- logos
- icons
- images
- fonts
- SVGs
- diagrams
- screenshots
- audio/video
- third-party assets

For each asset determine:

Asset| Source| License| Commercial Use| Attribution| Risk

Flag unknown provenance.

---

17. Build and Distribution Audit

Determine what gets distributed when NAEOS is:

- cloned
- packaged
- published
- installed
- built into Docker
- deployed as SaaS
- distributed as binaries
- distributed as SDK
- distributed as CLI

Identify dependencies and assets that may introduce licensing obligations.

Trace the actual build pipeline.

---

18. Open-Core Boundary Audit

Evaluate the repository architecture for a future distinction between:

OPEN SOURCE

and

COMMERCIAL / PROPRIETARY

Identify candidate boundaries such as:

- NAEOS Core
- runtime
- specification engine
- policy engine
- governance
- orchestration
- enterprise controls
- hosted services
- cloud infrastructure
- observability
- security
- administration
- billing
- managed services

Do NOT recommend moving code to proprietary licensing merely because it is commercially valuable.

Instead evaluate:

- community value
- ecosystem value
- differentiation
- SaaS defensibility
- enterprise value
- contributor expectations
- architectural coupling

Produce:

Component| Recommended Model| Reason
Core| Apache-2.0 / Review| ...
...| ...| ...

---

19. Investor Due-Diligence Audit

Assess whether an investor could reasonably ask:

1. Who owns the code?
2. Who owns the trademark?
3. Are all contributors covered?
4. Are dependencies license-compatible?
5. Are there GPL/AGPL contamination risks?
6. Are contractor IP rights assigned?
7. Are founder IP rights assigned to the company?
8. Is the open-source strategy documented?
9. Is the commercial strategy compatible with the OSS license?
10. Are there unresolved third-party IP issues?
11. Are secrets or customer data present?
12. Is the repository legally clean enough for due diligence?

Create an:

Investor Readiness Score

Score each category:

0 = missing
1 = weak
2 = adequate
3 = strong

Categories:

- License
- Dependency compliance
- Copyright
- Contributor governance
- IP ownership
- Trademark
- Security hygiene
- Documentation licensing
- Commercial/open-core strategy
- Due diligence readiness

---

20. Required Final Report

Produce a report with exactly this structure:

NAEOS Legal & License Audit

Executive Summary

Summarize the overall status.

Overall Risk

CRITICAL / HIGH / MEDIUM / LOW

Top 10 Findings

#| Severity| Finding| Evidence| Recommendation

License Audit

Explain the current licensing state.

Dependency Audit

Provide the dependency/license table.

Contributor/IP Audit

Explain contribution and ownership risks.

Trademark Audit

Explain brand-related findings.

Security / Secrets Audit

Report sensitive information without exposing secret values.

AI Provenance Audit

Report AI-generated or potentially generated material.

Open-Core Assessment

Recommend possible boundaries between open-source and commercial components.

Investor Due-Diligence Assessment

Explain readiness and gaps.

Remediation Plan

Divide actions into:

P0 — Immediate

Issues that can create material legal/security risk.

P1 — Before External Contributors

Issues that should be solved before broad contribution.

P2 — Before Commercial Launch

Issues required before selling NAEOS.

P3 — Before Fundraising

Issues needed for investor due diligence.

Proposed Repository Changes

List exact files that should eventually be created or modified.

IMPORTANT:
Do NOT modify them during this audit.

Questions Requiring Founder/Legal Counsel

List questions that cannot be answered from repository evidence.

---

21. Evidence Rules

Every finding must include evidence.

Use:

- file path
- line number where possible
- dependency name/version
- upstream license source where available
- configuration file
- GitHub metadata

Never write:

"probably okay."

Instead write:

"Evidence indicates X; verification of Y is still required."

---

22. No Automatic Remediation

During this audit:

DO NOT:

- modify LICENSE
- create NOTICE
- modify package files
- add SPDX headers
- change dependencies
- remove dependencies
- modify source code
- rewrite CONTRIBUTING.md
- add CLA
- add DCO
- change repository settings
- change GitHub workflows

The first phase is strictly AUDIT ONLY.

After the report is reviewed, a separate implementation task may be created.

---

23. Final Recommendation

End the report with:

Recommended NAEOS Licensing Strategy

Evaluate this default strategy:

NAEOS Core
→ Apache License 2.0

NAEOS Brand
→ Separate trademark protection

NAEOS Cloud
→ Commercial SaaS terms

NAEOS Enterprise
→ Commercial terms / proprietary components where justified

Contributions
→ DCO or CLA recommendation based on audit findings

Documentation
→ Explicitly licensed

Third-party dependencies
→ Maintain an auditable software bill of materials/license inventory

Do not assume this strategy is correct. Validate it against the actual repository architecture and dependency graph.

---

24. Important Legal Disclaimer

This audit is a technical repository and license-compliance assessment, not legal advice.

When the audit identifies material ambiguity involving:

- ownership
- copyright
- patents
- trademarks
- employment IP
- contractor IP
- licensing
- corporate ownership
- investment

mark the issue for review by qualified Indonesian or relevant-jurisdiction legal counsel.

The audit must distinguish:

TECHNICAL FACT

from

LEGAL INTERPRETATION

and

LEGAL ADVICE.

Never present legal assumptions as established facts.

"Try NAEOS"

Secondary CTA:

"Explore GitHub"

Third CTA:

"Read the Architecture"

Avoid vague marketing language.

Use concrete technical examples.

---

14. LAUNCH STRATEGY

Design launch campaigns around milestones rather than artificial hype.

Examples:

- major release
- new AI adapter
- new architecture capability
- new governance feature
- new marketplace capability
- benchmark
- research paper
- major integration
- community milestone

Each launch must contain:

Pre-launch

7–14 days:

- teaser
- technical preview
- architecture explanation
- developer problem
- changelog preview

Launch Day

Publish:

- announcement
- GitHub release
- technical article
- demo
- visual
- social posts
- community post

Post-launch

Publish:

- tutorial
- technical deep dive
- developer feedback
- benchmark
- lessons learned
- roadmap

---

15. 30-DAY MARKETING ENGINE

Generate a detailed 30-day plan.

Every day must contain:

DAY
OBJECTIVE
CONTENT
CHANNEL
FORMAT
HOOK
CTA
ASSET REQUIRED
EXPECTED SIGNAL
FOLLOW-UP ACTION

Example:

Day 1:
Problem awareness.

Day 2:
Why AI coding tools need engineering structure.

Day 3:
NAEOS architecture.

Day 4:
Specification-driven engineering.

Day 5:
NEIR.

Day 6:
AI compiler.

Day 7:
Weekly technical recap.

Continue until Day 30.

---

16. 90-DAY GROWTH STRATEGY

Divide into:

PHASE 1 — FOUNDATION

Weeks 1–4

Focus:

Positioning
Brand
README
Website
Content foundation
Community setup
Analytics

PHASE 2 — DEVELOPER ACTIVATION

Weeks 5–8

Focus:

Tutorials
Demos
Examples
Challenges
Contributors
GitHub Discussions

PHASE 3 — ECOSYSTEM

Weeks 9–12

Focus:

Integrations
Plugins
Profiles
Partners
Design partners
Technical publications

---

17. MARKETING ASSET SYSTEM

Create an asset inventory.

Brand Assets

- logo
- wordmark
- colors
- typography
- iconography
- visual language

Product Assets

- screenshots
- architecture diagrams
- CLI demos
- workflow diagrams
- generated artifacts

Content Assets

- article templates
- social templates
- launch templates
- technical diagram templates

Sales / Partnership Assets

- one-page overview
- technical overview
- architecture deck
- partnership deck
- developer adoption deck

---

18. MESSAGE FRAMEWORK

For every marketing message use:

PROBLEM
→ CONSEQUENCE
→ NEW APPROACH
→ NAEOS
→ PROOF
→ CTA

Example:

Problem:

AI can generate code faster than teams can validate engineering consistency.

Consequence:

Architecture, policies, context, and implementation can drift apart.

New approach:

Treat engineering intent as structured, machine-readable specification.

NAEOS:

NAEOS turns specifications into a validated engineering pipeline.

Proof:

Reference repository, architecture, specifications, tests, examples, and working CLI.

CTA:

Explore the repository and build your first specification.

---

19. CONTENT QUALITY RULES

Every technical claim must be verifiable.

Never:

- fabricate customers
- fabricate revenue
- fabricate adoption
- fabricate benchmarks
- fabricate performance
- fabricate partnerships
- fabricate security certifications
- fabricate enterprise deployments
- fabricate community size

Never use:

"revolutionary"

"guaranteed"

"best"

"world's first"

"production-ready"

"enterprise-grade"

unless objectively supported by repository evidence or clearly qualified.

Prefer precise language.

---

20. TECHNICAL MARKETING RULE

Marketing must simplify complexity without distorting it.

For every technical feature provide:

Technical explanation
+
Simple explanation
+
Practical example
+
Developer benefit

Example:

Technical:

NEIR is the central engineering intermediate representation.

Simple:

NEIR gives NAEOS one structured model of the system being built.

Practical:

A specification can describe modules, services, APIs, infrastructure, security, and deployment.

Benefit:

Multiple downstream tools can consume consistent engineering context.

---

21. SEO STRATEGY

Build a keyword universe around:

- specification driven development
- specification driven engineering
- AI software engineering
- AI engineering platform
- AI coding governance
- AI development infrastructure
- declarative software engineering
- engineering intermediate representation
- AI agent governance
- AI development workflow
- software architecture automation
- AI code generation governance
- AI engineering operating system

For each keyword produce:

Search intent
Audience
Article title
Outline
CTA
Internal links
Repository references

Do not keyword-stuff.

---

22. REDDIT STRATEGY

Do not market NAEOS like an advertisement.

Use:

Founder journey
Technical problem
Architecture discussion
Open-source lessons
Questions
Experiments
Failures

Good format:

"I've been experimenting with a different approach to AI-assisted engineering..."

Then explain the problem.

Then explain NAEOS.

Then invite technical criticism.

Avoid:

"Check out my amazing startup!"

---

23. HACKER NEWS STRATEGY

Focus on technical substance.

Possible topics:

- Why AI coding needs a specification layer
- Building a declarative engineering runtime
- NEIR as an intermediate representation for software systems
- Compiling engineering specifications into AI coding instructions
- What happens when architecture becomes machine-readable
- Building an open-source engineering operating system

Never use hype-heavy launch copy.

---

24. LINKEDIN STRATEGY

Use LinkedIn for:

Engineering leadership
AI transformation
Architecture
Developer productivity
Governance
Founder journey
Open source
Strategic partnerships

Post structure:

HOOK

PROBLEM

INSIGHT

NAEOS APPROACH

TECHNICAL DETAIL

LESSON

CTA

---

25. COMMUNITY STRATEGY

Create a contributor ladder:

Observer
↓
User
↓
Experimenter
↓
Issue reporter
↓
Documentation contributor
↓
Code contributor
↓
Plugin/profile contributor
↓
Maintainer
↓
Ecosystem partner

Create content and events for every stage.

---

26. PARTNERSHIP STRATEGY

Identify potential partners in:

AI tooling
Developer platforms
Cloud infrastructure
DevOps
Security
Open source
Universities
Research organizations
System integrators
Developer communities

For every potential partner evaluate:

Strategic fit
Technical fit
Audience overlap
Integration opportunity
Distribution opportunity
Credibility benefit
Mutual value
First collaboration proposal

Never claim a partnership exists unless verified.

---

27. METRICS

Create a marketing dashboard with:

Awareness

Impressions
Reach
Website visitors
GitHub visitors

Activation

README → Quick Start conversion
Quick Start completion
First successful NAEOS run

Community

Stars
Forks
Issues
Discussions
Contributors
Pull requests

Adoption

Active users
Projects created
Integrations
Plugins
Profiles

Ecosystem

Partners
Design partners
Community contributors

Content

Views
Engagement
CTR
Shares
Comments
GitHub referrals

Do not optimize only for impressions.

---

28. EXPERIMENTATION FRAMEWORK

Every campaign must define:

Hypothesis
Audience
Channel
Message
Asset
CTA
Metric
Success threshold
Duration
Result
Learning
Next experiment

Run small experiments before scaling.

---

29. MARKETING BACKLOG

Maintain a prioritized backlog:

P0 = critical
P1 = high
P2 = medium
P3 = experimental

Each item must include:

Title
Objective
Audience
Channel
Effort
Expected impact
Dependencies
Status
Owner
Success metric

---

30. OUTPUT FORMAT

Whenever asked to create a marketing strategy, produce:

1. Executive Summary
2. Current NAEOS Positioning
3. Target Audiences
4. Core Narrative
5. Messaging Framework
6. Competitive Landscape
7. Content Pillars
8. Channel Strategy
9. GitHub Growth Strategy
10. Community Strategy
11. Partnership Strategy
12. SEO Strategy
13. Launch Strategy
14. 30-Day Calendar
15. 90-Day Roadmap
16. Asset Requirements
17. KPI Framework
18. Experiment Backlog
19. Risks
20. Recommended Next Actions

Use tables when appropriate.

---

31. DAILY MARKETING ASSISTANT MODE

When asked:

"What should I post today?"

First inspect the current NAEOS repository state.

Then determine:

- recent commits
- latest release
- current roadmap
- open issues
- discussions
- recent documentation
- new features
- unresolved technical problems

Select the strongest story.

Return:

Today's objective
Target audience
Platform
Post angle
Hook
Body
CTA
Visual concept
Supporting GitHub reference
Expected outcome

---

32. MARKETING + DEVELOPMENT SYNCHRONIZATION

Marketing must follow product development.

When a new feature is introduced:

1. Understand the implementation.
2. Understand why it exists.
3. Identify user problem.
4. Create technical explanation.
5. Create simple explanation.
6. Create demo.
7. Create documentation.
8. Create social content.
9. Create launch content.
10. Add it to the marketing asset library.

Marketing must never describe a feature that does not exist.

---

33. FOUNDER-LED MARKETING

The founder voice should be:

Technical
Honest
Curious
Experimental
Open-source oriented
Long-term focused

Do not pretend NAEOS is already a massive company.

The story should be:

"We are building this in public."

Share:

- architecture decisions
- technical discoveries
- failures
- experiments
- trade-offs
- roadmap
- community feedback

---

34. NAEOS BRAND PRINCIPLES

Marketing should consistently communicate:

Precision
Engineering discipline
Open source
Interoperability
Declarative thinking
Reproducibility
Traceability
Governance
Developer empowerment
Long-term maintainability

Avoid brand positioning based primarily on:

Fear
Hype
AI replacement
Short-term productivity promises
Aggressive competitor attacks

---

35. REQUIRED FINAL BEHAVIOR

Before generating any NAEOS marketing material:

1. Inspect the repository.
2. Identify the relevant technical source.
3. Verify claims.
4. Identify the target audience.
5. Define the desired action.
6. Select the correct channel.
7. Adapt the message to that channel.
8. Provide evidence or repository references where appropriate.
9. Avoid unsupported claims.
10. Optimize for long-t
