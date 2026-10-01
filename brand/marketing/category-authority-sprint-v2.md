# NAEOS Brand Content Calendar — v3

**Status:** Execution draft  
**Window:** 2026-10-01 → 2026-10-30  
**Owner:** NAEOS Foundation  
**Strategy:** Brand-native proof and adoption engine

## 1. Strategic shift

The founder's personal feed is now the primary channel for founder perspective, category thesis, build-in-public narrative, and personal authority.

The NAEOS brand channel should therefore **not repeat the founder feed**.

NAEOS brand content exists to make the product and engineering system inspectable:

**Documentation → Architecture → Experiment → Evidence → Repository → Community → Adoption**

### Brand role

NAEOS should communicate:

- what the system is
- how the architecture works
- what is implemented
- how to reproduce it
- what evidence exists
- what remains experimental
- how engineers can participate

### Founder role

The founder feed can communicate:

- why the problem matters
- lessons learned
- personal observations
- category thesis
- build-in-public narrative
- reactions to industry conversations
- invitations to challenge the thesis

When a topic appears on both channels, the **founder provides the thesis; NAEOS provides the artifact/proof**.

---

## 2. Brand content pillars

| Pillar | Purpose | Primary CTA |
|---|---|---|
| Product Proof | Show implemented capabilities | Run / Inspect |
| Architecture | Explain system design | Read architecture |
| Experiments | Test a concrete engineering hypothesis | Reproduce |
| Evidence | Show inputs, outputs, validation and provenance | Inspect evidence |
| Documentation | Make the project easier to understand/use | Read / Try |
| Community | Invite technical participation | Discuss / Contribute |
| Ecosystem | Profiles, plugins, adapters and integrations | Build / Contribute |
| Release Notes | Show concrete progress | Try the change |

**Target mix:** 70% proof/technical utility, 20% community/ecosystem, 10% announcements.

---

## 3. 30-day brand operating cadence

### Week 1 — Make the system inspectable

**Day 1 — Product map**  
"What is NAEOS?"  
Show the relationship between specification, NEIR, validation, generation, AI context, governance and artifacts.  
CTA: Inspect the architecture.

**Day 2 — Architecture flow**  
"From engineering intent to executable artifacts."  
Show the end-to-end pipeline with links to repository artifacts.  
CTA: Read the architecture.

**Day 3 — Quick Start**  
"Run the smallest NAEOS workflow."  
Provide exact prerequisites, command, expected output and evidence location.  
CTA: Run it.

**Day 4 — Evidence**  
"What does NAEOS actually prove?"  
Show one real input → decision/process → output → validation/evidence chain.  
CTA: Inspect the evidence.

**Day 5 — NEIR**  
"Why NAEOS uses an intermediate engineering representation."  
Explain the role of NEIR without turning it into a generic thought-leadership post.  
CTA: Read the implementation/docs.

**Day 6 — Repository navigation**  
"Where to find the important parts of NAEOS."  
Map docs, architecture, CLI, tests, profiles/plugins and experiments.  
CTA: Explore GitHub.

**Day 7 — Weekly proof recap**  
Summarize what was made reproducible this week.  
CTA: Pick one artifact to inspect.

### Week 2 — Demonstrate engineering controls

**Day 8 — Policy lifecycle**  
Show how policy/authority is represented and consumed.  
CTA: Inspect policy artifacts.

**Day 9 — Authority change experiment**  
Show the setup for a mid-task authority/policy change. Publish only verified repository behavior.  
CTA: Reproduce the experiment.

**Day 10 — Stale authorization failure mode**  
Explain the concrete failure mode and what the system needs to detect.  
CTA: Inspect the test/evidence.

**Day 11 — Validation**  
"Validation before generation."  
Show the actual validation path and failure behavior.  
CTA: Run validation.

**Day 12 — Governance + auditability**  
Show documented mechanisms and evidence surfaces. Avoid broad security/compliance claims.  
CTA: Inspect the implementation.

**Day 13 — Handoff Contract**  
Explain the versioned handoff concept: payload digest, capability boundary, policy/version, provenance, expiry/replay state where implemented. Clearly separate implemented vs experimental fields.  
CTA: Read / discuss the contract.

**Day 14 — Community technical question**  
"If an agent changes a repository, what evidence should survive?"  
Use the answers to select a future experiment.  
CTA: Join the discussion.

### Week 3 — Reproducibility and system integrity

**Day 15 — Cross-service integrity**  
Demonstrate the problem of a locally valid change creating a downstream contract break.  
CTA: Inspect the example.

**Day 16 — Contract verification**  
Show how API/event contracts can become explicit verification inputs where implemented.  
CTA: Reproduce.

**Day 17 — Before/after workflow**  
Compare an agent workflow with explicit engineering controls against the uncontrolled baseline. Use only measured/reproducible differences.  
CTA: Inspect the experiment.

**Day 18 — Failure-mode report**  
Publish one failure discovered during development, including what failed, why, the fix, and remaining limitation.  
CTA: Read the evidence.

**Day 19 — Adapter walkthrough**  
Show one adapter from input/context to the NAEOS engineering workflow.  
CTA: Inspect / contribute.

**Day 20 — Verification report**  
Publish a compact evidence report containing commands, results, tests and artifact references.  
CTA: Reproduce.

**Day 21 — Contributor challenge**  
Offer one concrete good-first issue tied to documentation, tests, evaluator/reviewer heuristics, adapters or evidence UX.  
CTA: Open the issue / PR.

### Week 4 — Turn proof into participation

**Day 22 — Plugin/profile architecture**  
Explain extension boundaries and what a plugin/profile is allowed to contribute.  
CTA: Explore the extension model.

**Day 23 — Ecosystem challenge**  
Invite engineers to propose a profile, adapter or plugin against a concrete use case.  
CTA: Open a discussion/issue.

**Day 24 — Design-partner research**  
Invite engineering teams to reproduce a defined workflow and report gaps. Do not imply existing customers or deployments.  
CTA: Participate in research.

**Day 25 — Architecture decision record**  
Publish one important engineering trade-off or decision from the repository.  
CTA: Challenge the decision.

**Day 26 — Community showcase**  
Highlight verified contributions, experiments, discussions or useful external reproductions.  
CTA: Participate.

**Day 27 — Roadmap from evidence**  
Show what the latest experiments imply for the next engineering work. Distinguish roadmap from current capability.  
CTA: Comment / open an issue.

**Day 28 — AMA / office hour**  
Technical Q&A focused on architecture, experiments and implementation.  
CTA: Ask / reproduce.

**Day 29 — Monthly evidence report**  
Summarize shipped artifacts, experiments, failures, fixes, open questions and contribution opportunities.  
CTA: Inspect the evidence.

**Day 30 — Next experiment**  
Announce the next concrete hypothesis to test.  
CTA: Follow/reproduce/challenge the experiment.

---

## 4. Channel operating model

### LinkedIn — NAEOS brand

Use for:
- product proof
- architecture diagrams
- experiment results
- release notes
- evidence summaries
- contributor calls

Do not duplicate the founder's personal thesis post.

### GitHub

Canonical source for:
- implementation
- architecture
- experiments
- issues
- discussions
- reproducibility

### Reddit / Hacker News

Use selectively for:
- technical experiments
- failure reports
- architecture questions
- reproducible workflows

Lead with the engineering problem and evidence, not product promotion.

### Indie Hackers

Use for:
- verified build milestones
- distribution lessons backed by actual results
- founder/build system learnings that have not already been covered by the personal feed

### Website / Docs

Canonical conversion layer:

**Understand → Run → Inspect → Reproduce → Contribute**

---

## 5. Founder ↔ brand content rule

Avoid publishing the same idea twice.

Use this transformation:

**Founder:** "Why this engineering problem matters."  
↓  
**NAEOS:** "Here is the implementation."  
↓  
**GitHub:** "Here is the artifact."  
↓  
**Experiment:** "Here is what happened."  
↓  
**Community:** "Can you reproduce or challenge it?"

Example:

Founder post:
> "What happens when engineering authority changes mid-task?"

NAEOS brand:
> "Authority Change Experiment: setup, command, expected behavior, observed result."

GitHub:
> test + implementation + evidence

Community:
> "Can you reproduce this failure mode in your agent workflow?"

---

## 6. Content production rule

Every major brand post should map to at least one repository artifact.

Preferred chain:

**Artifact → Evidence → Brand content → Community discussion → Learning**

Do not create a brand post merely because the calendar says so.

If no verified artifact exists, convert the slot into:
- documentation improvement
- experiment proposal
- open technical question
- contributor issue

---

## 7. Primary CTAs

Prefer concrete CTAs:

1. Run the Quick Start
2. Inspect the architecture
3. Reproduce the experiment
4. Inspect the evidence
5. Open the relevant GitHub issue/discussion
6. Contribute a fix/test/doc
7. Propose a profile/plugin/adapter
8. Participate in design-partner research

Avoid generic:
- Learn more
- Follow us
- We are building the future
- Revolutionary AI
- Enterprise-ready

unless supported by a specific artifact or action.

---

## 8. KPI

Primary:
- Quick Start completion
- successful example runs
- GitHub referrals
- experiment reproductions
- issues/discussions
- first-time contributors
- contributor PRs
- profile/plugin activity
- qualified design-partner research participation

Secondary:
- reach
- impressions
- followers
- engagement rate

The primary question is:

> **Did the content cause someone to inspect, run, reproduce, or contribute to NAEOS?**

---

## 9. Publication gate

Before publishing:

1. Verify the repository artifact exists.
2. Verify current behavior on publication date.
3. Link the exact artifact where possible.
4. Separate implemented capability from roadmap.
5. State experimental limitations.
6. Do not claim customer adoption, production deployment, compliance certification, benchmark superiority, or security guarantees without evidence.
7. Prefer measured/reproducible results over adjectives.

---

## 10. Strategic outcome

The NAEOS brand feed should become the **public evidence layer of the project**.

The founder builds attention and frames the problem.

NAEOS demonstrates the engineering system.

GitHub provides the source of truth.

Experiments create evidence.

Community challenges the assumptions.

Contributors extend the system.

This creates a cleaner flywheel:

**Founder attention → NAEOS proof → GitHub reproduction → Community participation → Contribution → Ecosystem → Adoption**
