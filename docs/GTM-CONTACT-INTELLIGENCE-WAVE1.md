# GTM Contact-Level Intelligence — Wave 1

**Research date:** 2026-09-25  
**Scope:** 10 Tier-A Wave 1 accounts  
**Purpose:** identify public, current role targets and produce account-specific first-touch drafts.

## Contact targeting rules

- Prefer a directly relevant engineering/platform/AI/security owner over the highest-ranking executive.
- Use a named person only when a current public source supports the role.
- A named person is a **research target**, not evidence of buying intent.
- Do not infer email addresses or private contact data.
- Every personalized claim must be traceable to a public source and dated.
- If a suitable current person cannot be verified, target the role rather than inventing a name.

## Contact map

| Account | Primary target | Secondary target | Why this role | Evidence confidence |
|---|---|---|---|---|
| Stripe | Staff/Principal engineering leader in ML Platform / agentic AI | Engineering Lead, Agentic Commerce | Stripe's ML Platform explicitly includes agentic AI capabilities; Stripe also publicly identifies an Engineering Lead for Agentic Commerce. | High |
| Datadog | Alexis Lê-Quôc — CTO & Co-Founder | Yadi Narayana — Field CTO, APJ | CTO owns technical direction; Field CTO role explicitly works with technology leaders on secure/scalable architectures. | High |
| Cloudflare | Dev Productivity leadership | Sam Rhea — CIO; internal AI engineering authors/team | Dev Productivity owns internal tooling, CI/CD, build systems and automation; CIO publicly documents agent access/control challenges. | High |
| GitLab | Siva Padisetty — CTO | VP/Director Engineering or AI/DevSecOps leader | Current CTO leads software engineering, operations and customer support and publicly frames AI around the software lifecycle. | High |
| Atlassian | Sherif Mansour — Head of AI | Engineering/platform leadership | Current Head of AI is explicitly focused on AI infrastructure and how AI changes team workflows; current engineering material emphasizes governed agentic workflows. | High |
| Shopify | Farhan Thawar — VP & Head of Engineering | Infrastructure/Security leadership | Shopify publicly identifies Thawar in engineering material and describes AI as used broadly across engineering. | High |
| Twilio | CTO / platform engineering leadership | Security/identity engineering | Public MCP/Skills work exposes a large API surface to coding agents, making platform/security ownership the relevant boundary. | Medium — current individual not verified in this pass |
| MongoDB | Jim Scharf — CTO | Pablo Stern — CPO, AI and Emerging Products | MongoDB publicly confirms both roles; product strategy is explicitly moving core applications and AI workloads onto the platform. | High |
| AMD | Brian Amick — SVP, Technology and Engineering | Andrej Zdravkovic — SVP, GPU Technologies and Engineering Software & Chief Software Officer | Amick owns centralized engineering and technology delivery; Zdravkovic owns engineering software/software leadership. | High |
| ASOS | Lesley Johnson — Director of Technology | Technology/platform engineering leadership | ASOS publicly identifies Johnson as Director of Technology; public IT automation work makes technology leadership the relevant first entry point. | High |

## Account-specific first-touch drafts

### Stripe
**Trigger:** Stripe's ML Platform job material explicitly includes agentic AI capabilities; Stripe's agentic-commerce engineering work shows agents moving beyond code generation into execution. citeturn0search24turn0search4

**Draft:**
> Hi — I’ve been following Stripe’s work around agentic engineering and the ML Platform, especially the move from agents writing code toward agents that can plan, execute, and evaluate outcomes. I’m working on NAEOS around one specific control-plane question: how do you keep agent intent, authorization, execution, and verification independently inspectable when a workflow crosses API/infrastructure boundaries? Curious how Stripe is approaching that boundary today.

### Datadog
**Trigger:** Datadog is shipping autonomous development/operations capabilities and explicitly frames operational control as important as AI capability. citeturn1search3turn0search10

**Draft:**
> Hi Alexis — Datadog’s DASH work caught my attention because you’re pushing AI from assistance toward autonomous development and operational workflows while emphasizing control around the resulting complexity. I’m building NAEOS around the authorization/evidence boundary: separating what an agent proposes from what policy permits, what runtime executes, and what telemetry proves actually happened. Is that boundary already explicit in your agent workflows, or still distributed across the toolchain?

**APJ alternative:** For an APJ-first conversation, Yadi Narayana is a current Field CTO focused on secure, scalable architectures and technology-leader engagement. citeturn1search5

### Cloudflare
**Trigger:** Cloudflare reports 93% of R&D used AI coding tools in the prior 30 days; Dev Productivity owns internal tooling including CI/CD, build systems and automation. Cloudflare also publicly described a production-access scenario involving AI agents and multiple systems of record. citeturn0search2turn0search23

**Draft:**
> Hi — Cloudflare’s internal AI engineering stack is unusually relevant to what I’m building. You’ve already put an MCP/access layer around agent capabilities, while Dev Productivity owns CI/CD, build systems and automation. I’m working on NAEOS as a control plane for the next boundary: making the policy decision, last-boundary authorization, execution result, and evidence independently inspectable. Would that complement the control model you’re developing internally, or is that already solved another way?

### GitLab
**Trigger:** Siva Padisetty became CTO in January 2026 and leads software engineering, operations and customer support; GitLab is simultaneously expanding agentic SDLC and AI-fluent engineering practices. citeturn1search0turn0search12

**Draft:**
> Hi Siva — your move into GitLab’s CTO role coincides with a particularly interesting shift: GitLab is moving AI from coding assistance toward a more end-to-end software lifecycle. I’m building NAEOS around the governance boundary for that shift — an agent can propose an action, but policy should independently authorize it, runtime should execute it, and the system should retain evidence of what actually happened. I’d be interested in how GitLab thinks about that boundary across CI/CD, MRs and security workflows.

### Atlassian
**Trigger:** Atlassian currently describes governed agent loops, shared context and accountability; Sherif Mansour is current Head of AI. citeturn0search11turn0search25

**Draft:**
> Hi Sherif — Atlassian’s recent work on governed agent loops gets very close to a problem I’m working on with NAEOS. Context can tell an agent what it needs to know, but I think the harder systems question is keeping context separate from authorization: what the agent may do, where that decision is made, and what evidence proves the resulting action. How are you thinking about that separation as agentic workflows move across the SDLC?

### Shopify
**Trigger:** Shopify says engineers use AI broadly and identifies Farhan Thawar as VP & Head of Engineering. citeturn0search9

**Draft:**
> Hi Farhan — Shopify’s engineering material describes AI being used reflexively across the engineering organization, which made me curious about the next control problem. I’m building NAEOS as a control plane for AI-assisted engineering: separating agent intent from policy authorization, execution, and independent verification. As AI becomes a normal engineering primitive at Shopify, where do you draw that boundary for consequential repository/API/infrastructure actions?

### Twilio
**Trigger:** Twilio's MCP/Skills work gives coding agents access to 1,800+ APIs; security/authentication is therefore a natural discovery boundary.

**Draft:**
> Hi — Twilio’s MCP/Skills work is an interesting example of agents moving from code generation into direct API interaction. I’m building NAEOS around the control-plane question that follows: when an agent can reach a large API surface, where is the authoritative authorization decision made, and how do you retain independent evidence of what was actually executed? I’d be interested in how Twilio approaches that boundary.

**Note:** Do not name a Twilio executive until current public role data is independently verified.

### MongoDB
**Trigger:** MongoDB's 2026 agentic capabilities include managed MCP and controls over which apps can access clusters and with what permissions; Jim Scharf remains CTO and Pablo Stern is CPO, AI and Emerging Products. citeturn0search13turn1search4

**Draft:**
> Hi Jim — MongoDB’s move toward managed MCP and explicit controls over which apps can access clusters is a strong example of agents becoming part of the execution surface. I’m building NAEOS around the layer above that permission boundary: making intent, policy decision, execution, and verification durable and independently inspectable. How are you thinking about that evidence layer as agents become capable of making decisions as well as executing them?

### AMD
**Trigger:** AMD's current leadership lists Brian Amick as SVP Technology and Engineering and Andrej Zdravkovic as SVP, GPU Technologies and Engineering Software and Chief Software Officer; AMD also publicly documents OpenHands and multi-step agentic development work. citeturn0search6turn0search17

**Draft:**
> Hi Brian — AMD’s work around coding agents and the shift toward multi-step agentic pipelines made me think about the control boundary that appears once an agent is no longer doing one isolated coding task. I’m building NAEOS to separate proposal, policy authorization, runtime execution, and verification evidence at that boundary. For centralized engineering at AMD, where would you want that decision/evidence layer to sit in a multi-step developer workflow?

### ASOS
**Trigger:** ASOS currently identifies Lesley Johnson as Director of Technology and has publicly described an AI-ready knowledge foundation where evidence and human approval are part of the operational workflow. citeturn0search0

**Draft:**
> Hi Lesley — I came across ASOS’s work on an AI-ready knowledge foundation in IT Operations, particularly the emphasis on evidence and human approval before changes go live. I’m building NAEOS around a related engineering control problem: separating an agent’s recommendation from policy authorization, execution, and independently verifiable evidence. I’m curious whether that same model is becoming relevant for agentic engineering/IT actions at ASOS.

## Recommended contact order

For the first research/outreach pass, use **role relevance rather than seniority** as the selection rule:

1. Stripe — ML Platform / Agentic AI engineering
2. Cloudflare — Dev Productivity
3. GitLab — Siva Padisetty / engineering leadership
4. Atlassian — Sherif Mansour / AI
5. MongoDB — Jim Scharf / AI & emerging products
6. Datadog — Alexis Lê-Quôc / APJ Field CTO
7. AMD — Brian Amick / engineering software
8. Shopify — Farhan Thawar / engineering
9. ASOS — Lesley Johnson / technology
10. Twilio — platform/security role pending current named-person verification

This is a **contact-research sequence, not a ranking of companies or prospects**.

## Evidence discipline

Before sending a personalized message:
- re-open the source;
- confirm the person's current role;
- confirm the trigger is still current;
- preserve the source URL and research date;
- remove any sentence that implies internal knowledge not present in the source;
- do not infer an email address;
- do not send until the recipient/channel is explicitly selected by the operator.

## Next step

After this artifact lands, the next execution unit is **Outbound Pack v1**: one final LinkedIn/email-length first touch per verified contact, one technical follow-up, one 30–45-day pilot CTA, and a tracking schema for sent/replied/discovery/pilot stages.
