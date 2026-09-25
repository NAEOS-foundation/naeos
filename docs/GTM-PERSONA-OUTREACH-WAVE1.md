# GTM Persona & Outreach — Wave 1

**Status:** Working GTM artifact  
**Scope:** Tier-A Wave 1 target accounts  
**Research date:** 2026-09-25  
**Primary source:** public company material; names are included only where current public evidence supports them.

## Operating rule

This document is for research and outreach preparation. Inclusion does **not** imply that an account is evaluating NAEOS, has purchasing intent, or endorses NAEOS.

The outreach thesis remains:

> AI agents propose. Policy decides. Runtime executes. Observation confirms.

Outreach should start from a documented engineering problem, not from a generic AI pitch. Do not claim a compatibility, integration, partnership, or internal pain point unless verified in discovery.

## Wave 1 account playbook

| Account | Primary persona | Secondary persona | Documented trigger | First outreach angle | Pilot hypothesis | Confidence |
|---|---|---|---|---|---|---|
| Stripe | Engineering/platform leader for developer infrastructure or agentic engineering | API/platform engineering; security | Stripe is benchmarking agents on real integrations and highlights planning, state, recovery, testing and validation as gaps between coding and autonomous engineering. | “How are you separating agent intent from authorization and verification when an agent crosses API, database, test and infrastructure boundaries?” | Govern one bounded agent workflow from intent → policy → execution → verification with durable evidence. | High |
| Datadog | Platform/DevEx or engineering leader responsible for AI-assisted development | Security/SRE/AI observability | Datadog is shipping agentic investigations, MCP integrations and AI controls while emphasizing operational complexity and keeping engineers in control. | “You already have telemetry for agentic systems; where do you capture the authorization decision and proof that the agent actually performed the approved action?” | Correlate agent action authorization, runtime outcome and observability evidence for one workflow. | High |
| Cloudflare | Dev Productivity / internal tooling leader | Security/CIO/AI infrastructure | Cloudflare reports 93% of R&D using AI coding tools and describes internal MCP servers, access layers and WriteGuard-style controls before expanding write access. | “Your internal MCP access layer solves tool reach; how are you making the policy decision and execution receipt independently inspectable across the workflow?” | Bound write-enabled MCP/agent workflow with explicit policy, last-boundary revalidation and evidence. | High |
| GitLab | Engineering/AI engineering leader | Security/compliance; DevSecOps platform | GitLab is scaling agentic SDLC workflows and explicitly discusses governance, compliance, audit and AI fluency. | “When governed agents operate across CI/CD, merge requests and security workflows, what system is the authoritative policy boundary and evidence record?” | Map one agentic SDLC workflow to explicit policy decisions and auditable execution evidence. | High |
| Atlassian | AI/engineering platform leader | Security/governance; DevEx | Atlassian describes governed agent loops, shared context, accountability and controls for AI-native software development. | “The context layer can tell an agent what to know; what independently decides what it may do, and what records the resulting evidence?” | Add an authorization/evidence boundary around one governed agent loop without replacing the existing system of record. | High |
| Shopify | VP/Head of Engineering or platform engineering | Infrastructure/security; developer productivity | Shopify states that engineers use AI broadly and highlights AI integration across the engineering stack; its 2026 commerce work also shows rapid API/UCP development for AI-mediated commerce. | “As AI becomes a normal engineering primitive, where is the control boundary for consequential changes across repositories, APIs and infrastructure?” | One bounded engineering workflow with policy checks and execution evidence. | Medium |
| Twilio | Engineering/platform leader for developer infrastructure | Security/identity; API platform | Twilio exposes 1,800+ APIs to coding agents through MCP/Skills and publicly discusses agent authentication and security concerns. | “With agents able to reach a large API surface, how do you make authorization and post-action evidence explicit rather than implicit in the agent/tool layer?” | Govern a representative agent-to-API workflow with least privilege, authorization and receipts. | High |
| MongoDB | CTO/engineering or developer platform leader | Security/AI platform | MongoDB launched Agent Skills, managed MCP and controls for which apps can access clusters and with what permissions. | “You already control which apps can reach clusters; can the authorization decision and actual execution outcome be traced independently of the agent?” | Govern one agent-to-database workflow, including policy decision, permission scope and verification evidence. | High |
| AMD | Technology/engineering leader for developer platforms or software | AI software/engineering; security | AMD publicly documents OpenHands coding-agent deployment and an agentic shift toward multi-step pipelines; current leadership includes SVP Technology & Engineering and Chief Software Officer roles. | “As coding agents move into multi-step pipelines, where is the durable control/evidence boundary between proposed work, authorized execution and observed result?” | Govern one bounded multi-step coding-agent pipeline with deterministic policy and evidence. | Medium-High |
| ASOS | Director/VP Technology or platform engineering | IT automation/security; data/AI | ASOS is scaling AI across product and operations, launched an AI Stylist in ChatGPT, and documents an auditable AI-ready knowledge foundation with human approval before changes go live. | “Your AI automation work already emphasizes evidence and human approval; could the same evidence model govern consequential agent actions in engineering/IT workflows?” | Start with a bounded IT/engineering automation workflow where recommendation, authorization, execution and evidence are separately recorded. | High |

## Persona notes and public evidence

### Stripe
**Persona:** platform/developer infrastructure, engineering systems, or agentic engineering leadership.

**Evidence:** Stripe's engineering team published a benchmark for real AI-agent integrations and explicitly identifies long-horizon planning, persistent state, recovery, cross-domain work, testing and validation as important parts of real engineering. Stripe also has public work on agentic commerce and machine payments.

**Discovery questions**
1. Which agent actions currently cross repository, API, database or infrastructure boundaries?
2. Where is authorization evaluated independently from model output?
3. What evidence is required to prove an agent action was actually executed and verified?
4. Which failure mode currently requires the most human intervention?

**Source:** https://stripe.com/blog/can-ai-agents-build-real-stripe-integrations

### Datadog
**Persona:** platform/DevEx, AI engineering, SRE, or security leader.

**Evidence:** Datadog's 2026 AI releases cover coding agents, MCP integrations, autonomous investigations and AI controls. Its State of AI Engineering research identifies operational complexity as a major scaling constraint.

**Discovery questions**
1. Do agent sessions have a stable identity and authorization context?
2. Can an operator reconstruct proposal → policy decision → execution → observation?
3. Which agent actions require human approval today?
4. Where does observability end and authorization evidence begin?

**Source:** https://www.datadoghq.com/blog/dash-2026-new-feature-roundup-ai/

### Cloudflare
**Persona:** Dev Productivity/internal tooling first; security and CIO/AI infrastructure as secondary stakeholders.

**Evidence:** Cloudflare reports 93% R&D usage of AI coding tools and describes internal MCP servers, an access layer and Dev Productivity ownership of CI/CD/build/automation. Its public WriteGuard material describes fine-grained controls before expanding MCP write access.

**Discovery questions**
1. What is the canonical policy source for internal agent tool access?
2. Is the authorization decision durable and queryable after execution?
3. How is last-boundary revalidation handled before write actions?
4. What constitutes a trustworthy execution receipt?

**Sources:** https://blog.cloudflare.com/internal-ai-engineering-stack/ ; https://blog.cloudflare.com/author/scott-roe-meschke/

### GitLab
**Persona:** engineering/AI engineering plus security/compliance.

**Evidence:** GitLab publicly describes governed agentic workflows, audit/compliance controls and internal AI fluency. Its handbook also defines a broad engineering leadership structure suitable for role-based discovery.

**Discovery questions**
1. Which agent actions need organization-level policy rather than repository-level rules?
2. How are compliance controls applied at the final execution boundary?
3. Can evidence connect intent, authorization and execution across an agent workflow?
4. Where do AI pilots stall when governance is introduced?

**Sources:** https://about.gitlab.com/blog/how-gitlab-fosters-ai-fluent-teams/ ; https://handbook.gitlab.com/job-description-library/engineering/engineering-management/

### Atlassian
**Persona:** AI/engineering platform leadership; governance/security as secondary.

**Evidence:** Atlassian's current material explicitly frames governed agent loops around shared context, controls and accountability. Its AI engineering page describes production-grade AI and agentic workflows at enterprise scale.

**Discovery questions**
1. Which component is authoritative for authorization when agents act across the SDLC?
2. How is accountability preserved across chained agent actions?
3. Which evidence is required before an agent-created change can be considered verified?
4. How do you prevent shared context from becoming implicit authorization?

**Sources:** https://www.atlassian.com/blog/jira/governed-agent-loops ; https://www.atlassian.com/company/careers/engineering/ai

### Shopify
**Persona:** engineering/platform leadership, with infrastructure/security as secondary.

**Evidence:** Shopify's engineering careers material says AI is used broadly across engineering and identifies VP & Head of Engineering Farhan Thawar. Shopify's 2026 Catalog API/UCP work shows rapid development of AI-mediated commerce primitives.

**Discovery questions**
1. Which engineering actions are currently agent-assisted?
2. Where are repository, API and infrastructure permissions evaluated?
3. How do you retain evidence across rapidly changing AI-assisted workflows?
4. What is the smallest workflow where deterministic authorization would materially reduce risk?

**Sources:** https://www.shopify.com/careers/disciplines/engineering-data ; https://www.shopify.com/news/spring-26-edition-design

### Twilio
**Persona:** platform/API engineering plus security/identity.

**Evidence:** Twilio's MCP/Skills work exposes a large API surface to coding agents, while its public agent guidance emphasizes authentication, secure APIs and risks such as prompt injection and information disclosure.

**Discovery questions**
1. How is agent identity mapped to API authorization?
2. Are permissions evaluated per action or inherited from the tool/session?
3. What evidence is retained for consequential API calls?
4. Where would you want an independent policy decision before an agent invokes an API?

**Source:** https://www.twilio.com/en-us/blog/developers/introducing-twilio-mcp-skills

### MongoDB
**Persona:** engineering/developer platform and security.

**Evidence:** MongoDB launched Agent Skills, managed MCP and controls for application access to clusters and permissions. Its current leadership page lists Jim Scharf as CTO and CJ Desai as CEO.

**Discovery questions**
1. Can an agent's database permissions be traced to a durable policy decision?
2. How are write actions revalidated before execution?
3. What evidence distinguishes proposed database changes from executed changes?
4. Which workflow would be safe and useful for a 30–45 day controlled pilot?

**Sources:** https://www.mongodb.com/company/blog/product-release-announcements/mongodb-for-agentic-era-built-for-developers-ai-agents ; https://www.mongodb.com/company/leadership

### AMD
**Persona:** software/developer platform leadership, with AI software and security as secondary.

**Evidence:** AMD documents OpenHands coding-agent deployment on Instinct GPUs and an agentic shift toward multi-step pipelines. Current executive leadership lists Brian Amick as SVP Technology & Engineering and Andrej Zdravkovic as SVP GPU Technologies and Engineering Software and Chief Software Officer.

**Discovery questions**
1. Which agentic engineering workflows are moving from experiments to repeatable pipelines?
2. What is the authorization boundary for multi-step agent execution?
3. How are policy/schema/version mismatches handled?
4. What evidence is required to independently verify pipeline outcomes?

**Source:** https://www.amd.com/en/corporate/leadership.html

### ASOS
**Persona:** Director/VP Technology, platform engineering or IT automation.

**Evidence:** ASOS identifies Lesley Johnson as Director of Technology and has public material on an AI-ready knowledge foundation in IT Operations where recommendations are linked to verifiable evidence and humans approve changes before production. ASOS also launched ASOS Stylist in ChatGPT and has described broader AI adoption.

**Discovery questions**
1. Which engineering or IT actions are candidates for agentic execution rather than recommendation?
2. Can the same evidence model used for knowledge automation be applied to execution?
3. Where does human approval become the bottleneck?
4. What bounded action would benefit from explicit policy and an execution receipt?

**Sources:** https://www.asosplc.com/inside-asos/building-an-ai-ready-knowledge-foundation-at-asos/ ; https://www.asosplc.com/news-and-media/latest-news/asos-launches-asos-stylist-app-in-chatgpt/

## Outreach sequence

### Touch 1 — problem-led
Do not lead with a product tour. Reference the account's public engineering/AI work and ask one boundary question.

**Template**
> I saw [documented initiative]. The part I find interesting is the boundary between what an agent proposes and what the engineering system actually authorizes and executes. How are you handling that boundary today?

### Touch 2 — technical follow-up
Use only after engagement.

> One pattern we're testing in NAEOS is to make proposal, policy decision, execution and observation separate evidence events, with authorization revalidated at the last controllable boundary. Is that close to how your current workflow is structured?

### Touch 3 — bounded pilot
Offer a narrow technical evaluation, not an enterprise transformation.

> If useful, we could take one bounded agentic workflow and map the authorization/evidence path end to end — including one deliberate deny case — without replacing your existing CI/CD or system of record.

## Qualification gate

An account advances from research to active outreach only when at least three of these are true:

- Public evidence of agentic/AI-assisted engineering or operational automation.
- A plausible platform, DevEx, engineering, security or AI engineering owner.
- A consequential action that crosses a permission or policy boundary.
- A requirement for auditability, verification or durable evidence.
- A bounded workflow that can be evaluated in 30–45 days.
- A technical stakeholder willing to validate the workflow.

## Anti-claims

Do not state that an account:
- is evaluating NAEOS;
- has a known NAEOS-compatible architecture;
- has a specific internal pain not disclosed publicly;
- is likely to buy;
- is a partner or customer;
- has approved an agentic workflow.

These require direct discovery or explicit public evidence.

## Next execution step

Convert the 10 account playbooks into a contact-level research sheet only after verifying current public role/name data. Store source URL and evidence date for every named person. Then prepare account-specific first-touch drafts, keeping the technical claim anchored to the documented trigger.
