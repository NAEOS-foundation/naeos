# NAEOS GTM — Wave 1 Account Intelligence

Research snapshot: 2026-09-25. This document records public evidence and working hypotheses for account-based discovery. It does not assert buying intent, internal priorities, or a relationship with NAEOS.

## 1. Stripe

**Public signal:** Stripe published an engineering benchmark in March 2026 testing whether AI agents can build real Stripe integrations end-to-end, explicitly emphasizing verification, testing, validation, persistent state, and recovery. Stripe later reported that agent traffic to Stripe documentation had grown more than 10x in 2025 and that 70% of Stripe CLI requests for API resources came from agents.  
**Evidence:** https://stripe.com/blog/can-ai-agents-build-real-stripe-integrations ; https://stripe.com/blog/stripe-projects-add-new-agents-providers-developer-controls

**Likely NAEOS discovery workflow:** governed agent execution for API/integration changes where correctness and verification are consequential.
**Discovery question:** What authorization and evidence boundary do you require when an agent moves from code generation into infrastructure, credentials, database, or deployment actions?
**Personalization hook:** Stripe's own benchmark frames the gap between code generation and autonomous engineering as a verification problem.
**Persona targets:** Engineering/platform leader; developer productivity; security/architecture.
**Evidence confidence:** High for agentic-development relevance; internal NAEOS fit remains a discovery hypothesis.

## 2. Datadog

**Public signal:** Datadog reports broad use of coding agents and has shipped native integrations for Claude Code, Codex, Cursor and other agents. Its 2026 announcements include Agent Console for adoption/impact visibility, AI Guard for runtime guardrails, and an Agent MCP providing secure, auditable infrastructure access. Datadog also published work on AI Golden Paths for standardized agent workflows.  
**Evidence:** https://www.datadoghq.com/blog/dash-2026-new-feature-roundup-ai/ ; https://www.datadoghq.com/blog/datadog-agent-console/ ; https://www.datadoghq.com/blog/ai-development-golden-paths/

**Likely NAEOS discovery workflow:** policy-controlled agent actions across developer tooling, infrastructure, observability and remediation.
**Discovery question:** Where does Datadog want authorization and evidence to live when agents can both inspect telemetry and take operational actions?
**Personalization hook:** Their public Golden Paths and Agent MCP work create a direct conversation around policy boundaries, authorization and durable evidence.
**Persona targets:** Platform engineering; developer productivity; security/AI security; SRE.
**Evidence confidence:** High.

## 3. Cloudflare

**Public signal:** Cloudflare reported that in the prior 30 days, 93% of its R&D organization used AI coding tools. It describes internal MCP servers, an access layer and AI tooling built by its Dev Productivity organization, which also owns CI/CD, build systems and automation.  
**Evidence:** https://blog.cloudflare.com/internal-ai-engineering-stack/

**Likely NAEOS discovery workflow:** centralized governance for AI coding agents connected to internal engineering systems and MCP services.
**Discovery question:** As agent adoption reaches most of R&D, how are tool permissions, consequential actions and evidence standardized across teams?
**Personalization hook:** Cloudflare already operates an internal AI engineering stack and access layer; NAEOS can be positioned as a policy/evidence control-plane experiment rather than another coding assistant.
**Persona targets:** Dev Productivity; platform engineering; security architecture; engineering leadership.
**Evidence confidence:** High.

## 4. GitLab

**Public signal:** GitLab Duo Agent Platform is GA and supports agentic work across the software lifecycle. GitLab 19.4 (September 2026) added governed agentic flows, MCP tools across CI/CD, merge requests, work items, vulnerabilities and projects, plus usage visibility. GitLab explicitly frames confidence and governance as constraints on scaling agentic automation.  
**Evidence:** https://about.gitlab.com/press/releases/2026-09-17-gitlab-19-4-brings-new-agentic-automation-at-a-lower-cost/ ; https://about.gitlab.com/press/releases/2026-08-20-gitlab-scales-agentic-ai-across-trusted-software-delivery-workflows/

**Likely NAEOS discovery workflow:** external governance/control-plane interoperability around multi-agent software delivery.
**Discovery question:** Which governance controls should remain platform-native, and where could an independent authorization/evidence layer be valuable across heterogeneous agents and repositories?
**Personalization hook:** Avoid pitching “agentic AI” as new; discuss interoperability, policy portability, evidence and independent verification.
**Persona targets:** AI/agent platform; DevSecOps; security/compliance; developer productivity.
**Evidence confidence:** High.

## 5. Atlassian

**Public signal:** Atlassian describes AI-native engineering as a shift where agents can own tasks end-to-end and has released tooling for planning, governing and measuring work across humans and agents. Its September 2026 material emphasizes context, guardrails, accountability and trusted systems of record.  
**Evidence:** https://www.atlassian.com/blog/company-news/the-agentic-pivot ; https://www.atlassian.com/blog/jira/governed-agent-loops ; https://www.atlassian.com/blog/company-news/ai-sdlc

**Likely NAEOS discovery workflow:** policy and evidence boundary spanning planning, agent execution, code changes and verification.
**Discovery question:** How should an organization independently verify that an agent's proposed action respected architecture, policy and authorization constraints before execution?
**Personalization hook:** Atlassian's own “governed agent loops” framing overlaps directly with NAEOS's control-plane thesis; differentiate through independent authorization/evidence rather than competing for planning context.
**Persona targets:** Developer experience; platform engineering; engineering productivity; AI engineering.
**Evidence confidence:** High.

## 6. Shopify

**Public signal:** Current public evidence was not sufficiently verified in this research pass to make a strong claim about Shopify's internal AI-agent engineering adoption.
**Likely NAEOS discovery workflow:** large-scale product engineering with bounded agent-assisted changes.
**Discovery question:** Which agent actions are currently permitted automatically, which require human approval, and how are those decisions evidenced?
**Personalization hook:** Lead with a discovery conversation, not an adoption assumption.
**Persona targets:** Developer productivity; platform engineering; engineering leadership; security.
**Evidence confidence:** Low until validated.

## 7. Twilio

**Public signal:** Twilio launched an MCP Server and Skills giving coding agents native access to more than 1,800 Twilio APIs. Twilio's public developer material also stresses authentication, secure tools and protection against prompt injection and information disclosure for agent systems.  
**Evidence:** https://www.twilio.com/en-us/blog/developers/introducing-twilio-mcp-skills ; https://www.twilio.com/en-us/blog/developers/ai-agents-explained

**Likely NAEOS discovery workflow:** governed agent access to API/tool ecosystems where authorization and tool scope matter.
**Discovery question:** How do you define and audit the boundary between an agent discovering an API and an agent being authorized to invoke consequential operations?
**Personalization hook:** Twilio's MCP work creates a concrete bridge from agent context to NAEOS authorization and evidence.
**Persona targets:** Developer platform; security; AI infrastructure; developer experience.
**Evidence confidence:** High.

## 8. MongoDB

**Public signal:** MongoDB launched Agent Skills for coding agents and an MCP Server that can connect agents to MongoDB capabilities. Its August 2026 announcement describes controls over which applications can access clusters and with what permissions.  
**Evidence:** https://www.mongodb.com/company/blog/product-release-announcements/introducing-mongodb-agent-skills ; https://www.mongodb.com/company/blog/product-release-announcements/mongodb-for-agentic-era-built-for-developers-ai-agents

**Likely NAEOS discovery workflow:** authorization and evidence for agents interacting with data infrastructure and schema/database workflows.
**Discovery question:** How should engineering policy distinguish read-only agent context from agent actions that alter schema, data or production infrastructure?
**Personalization hook:** NAEOS can complement MongoDB's agent connectivity by focusing on cross-system authorization, policy and evidence rather than database-specific agent skills.
**Persona targets:** Developer platform; data infrastructure; security; AI engineering.
**Evidence confidence:** High.

## 9. AMD

**Public signal:** AMD publicly documents deploying OpenHands coding agents on AMD Instinct GPUs and describes agentic multi-step workflows. Its 2026 AI DevDay material also highlights the shift toward multi-step agentic pipelines and expanded software validation.  
**Evidence:** https://www.amd.com/en/developer/resources/technical-articles/2026/deploying-openhands-coding-agents-on-amd-instinct-gpus.html ; https://www.amd.com/en/developer/resources/technical-articles/2026/amd-ai-devday-2026.html

**Likely NAEOS discovery workflow:** controlled agent execution and verification across a large software/AI stack with performance and compatibility constraints.
**Discovery question:** How are agent-generated changes validated across ROCm, model integrations, CI and upstream dependencies before they become trusted artifacts?
**Personalization hook:** Focus on evidence and independent verification across a fast-moving open-source/software stack.
**Persona targets:** ROCm/software engineering; developer infrastructure; AI platform; security.
**Evidence confidence:** Medium-high for agentic engineering relevance; internal governance fit needs validation.

## 10. ASOS

**Public signal:** Current public evidence was not sufficiently verified in this research pass to make a strong claim about ASOS's internal AI-agent engineering adoption.
**Likely NAEOS discovery workflow:** bounded agent-assisted product engineering and CI/CD workflows.
**Discovery question:** What controls are required before AI-generated changes can cross from developer workflow into production delivery?
**Personalization hook:** Use a measurable design-partner proposition rather than assuming existing agent scale.
**Persona targets:** Engineering productivity; platform engineering; security; engineering leadership.
**Evidence confidence:** Low until validated.

## Wave 1 qualification model

Do not score or rank accounts by perceived attractiveness. Qualify each independently against the same factual fields:

1. Public evidence of agentic engineering.
2. Existence of a platform/developer-productivity owner.
3. Consequential agent actions in the workflow.
4. Existing policy/authorization boundary.
5. Evidence/audit requirement.
6. Ability to define a bounded 30–45 day pilot.
7. Technical stakeholder willing to validate the workflow.

## Immediate next actions

- Verify current leadership/persona contacts through public company sources or professional profiles before outreach.
- For each account, identify one concrete engineering workflow suitable for a bounded design-partner experiment.
- Record evidence date and source for every non-trivial claim.
- Do not claim NAEOS compatibility, partnership, endorsement or purchasing intent without direct evidence.
- Use the same discovery questions across all ten accounts so results are comparable.
