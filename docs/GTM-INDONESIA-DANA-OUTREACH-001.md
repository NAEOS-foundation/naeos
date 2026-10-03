# DANA Founder Outreach Sequence v1

**Account:** DANA Indonesia  
**Primary public persona:** Norman Sasono, CTO  
**Purpose:** Founder-led technical discovery for a bounded AI-agent control-plane design-partner conversation.

## Public signal

DANA's September 2026 AI@Work Lab material says roughly 80–90% of its code is generated with AI support, while engineers remain responsible for problem formulation, architecture, validation, security, and final decisions. DANA also describes Smart Friction as adding verification layers to higher-risk activities. In 2024, DANA publicly described integrating GitHub Copilot into its SDLC, with nearly 300 developers using it at the time.

These signals support a discovery hypothesis, not a claim that DANA has an unresolved control gap.

## First touch — LinkedIn

Hi Norman — I’ve been following DANA’s recent AI@Work discussions, especially the emphasis on human judgment, accountability, and verification as AI adoption scales.

I’m building NAEOS, an open-source engineering control plane for AI coding agents. We’re testing a narrow question:

As AI-assisted development moves from generating code toward taking consequential actions, where should the boundary sit between what an agent proposes, what it is authorized to execute, and what must remain explicitly approved?

I’d like to test that question against one real engineering workflow at DANA — without replacing existing coding tools.

Would you be open to a 20-minute technical exchange? I’m mainly looking to learn whether there is a real control/evidence gap worth validating through a small bounded workflow.

## Follow-up 1 — 5–7 days later

Hi Norman — following up on this.

One concrete workflow is enough for the discussion. We can map:

Intent → Agent → Proposed Action → Authorization → Execution → Verification → Evidence

If DANA already has this boundary well controlled, that is useful for us to learn too. The goal is validation, not a product demo.

Would a 20-minute technical exchange be reasonable?

## Follow-up 2 — 7–10 days after Follow-up 1

Hi Norman — closing the loop for now.

The specific question I’m exploring is whether AI-assisted engineering needs a more explicit control boundary once agents can move beyond generating code into actions affecting repositories, CI/CD, or protected environments.

If this is already handled well at DANA, I’d still value understanding how you approach it. If there is a narrow gap, we can test it without changing the existing engineering stack.

Happy to reconnect whenever the topic is relevant.

## If Norman replies positively

Thanks, Norman. To keep the conversation practical, I suggest we spend the 20 minutes on one real workflow rather than a general AI discussion.

I’d like to understand:
1. What the agent is currently allowed to access or change.
2. Which actions require explicit authorization.
3. What happens when policy or task scope changes mid-workflow.
4. What evidence remains after execution.
5. How the result is independently verified.

If we find a concrete gap, we can define a small 2–4 week validation around one repository, one agent path, and a small number of capabilities.

## Founder operating rules

- Send individually; do not bulk-send.
- Keep the first message problem-led and low-pressure.
- Do not claim DANA is a customer, partner, or integration.
- Do not claim DANA has a control weakness before discovery.
- Do not lead with pricing or procurement.
- Do not ask for a product demo as the first CTA.
- Treat a positive reply as discovery, not pilot acceptance.
- Record the observed workflow, current control, evidence model, gap, owner, and next action.
- If there is no meaningful gap, record that outcome and do not force an NAEOS use case.
