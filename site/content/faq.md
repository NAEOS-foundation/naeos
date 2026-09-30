---
title: Frequently Asked Questions
description: Common questions about NAEOS, the engineering control plane for AI coding agents, its evidence model, and the Golden Path.
---

<div class="faq-list">
<div class="faq-item">
<button class="faq-question"><span>What is NAEOS?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>NAEOS (Nusantara AI Engineering Operating System) is an open-source engineering control plane for AI coding agents. It provides explicit boundaries between engineering intent, specification, policy, authorized execution, observation, evidence, and independent verification.</p></div>
</div>

<div class="faq-item">
<button class="faq-question"><span>What problem does NAEOS address?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>AI coding agents can generate and propose changes quickly, but engineering teams still need to control what an agent is authorized to do and retain evidence of what actually happened. NAEOS provides a structured control boundary for that workflow.</p></div>
</div>

<div class="faq-item">
<button class="faq-question"><span>What is the NAEOS control model?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>The current model is: Specification → NEIR → Validation + Policy → Agent Intent → Authorized Execution → Observation → Evidence → Independent Verification. The important boundary is that an agent does not grant itself execution authority.</p></div>
</div>

<div class="faq-item">
<button class="faq-question"><span>What is the Golden Path?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>The Golden Path is the reproducible developer proof path for the control-plane workflow. The current adoption path is Control Plane → Golden Path → Reference Demo → Evidence → Independent Verification. P1.6–P1.10 form the primary proof sequence.</p></div>
</div>

<div class="faq-item">
<button class="faq-question"><span>What does P1.11 add?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>P1.11 adds an independent verifier CLI for serialized EvidenceBundle records. It verifies decision/execution identity bindings and evidence integrity without evaluating policy, executing actions, or requiring a live control-plane dependency.</p></div>
</div>

<div class="faq-item">
<button class="faq-question"><span>Is NAEOS production-ready?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>NAEOS contains substantial engineering and governance capabilities, but public proof and adoption are intentionally being strengthened before broader commercialization. Evaluate the documented scope, Golden Path, and evidence produced by the repository for your use case.</p></div>
</div>

<div class="faq-item">
<button class="faq-question"><span>Is NAEOS open source?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>Yes. NAEOS is licensed under Apache License 2.0. Contributions are governed through the repository's DCO-based contribution process.</p></div>
</div>

<div class="faq-item">
<button class="faq-question"><span>How do I evaluate NAEOS?</span><span class="faq-arrow">▾</span></button>
<div class="faq-answer"><p>Start with the public Control Plane, then run the repository Golden Path and Reference Demo from a fresh checkout. Inspect the resulting evidence and verify it independently. This is the recommended technical evaluation path.</p></div>
</div>
</div>