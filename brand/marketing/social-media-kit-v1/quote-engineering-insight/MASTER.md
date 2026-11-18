# Quote & Engineering Insight Master — v1.0

**Status:** Draft asset master  
**Brand source of truth:** `brand/brand.json`, `brand/brand.css`, and Logo Master v2.0.0  
**Target channel:** LinkedIn  
**Primary canvas:** 1080 × 1350 px (4:5 portrait)  
**Companion canvas:** 1080 × 1080 px (1:1 square; adapt the same hierarchy, do not stretch)

## Purpose

A reusable editorial template for engineering principles, governance insights, and evidence-led statements. It must communicate NAEOS as the engineering control plane for AI coding agents—not as a generic AI productivity product.

## Master statement

> Trust is not a feature. It is an engineering outcome.

Supporting line:

> Define the boundary. Authorize execution. Preserve evidence. Verify independently.

## Canonical visual system

Use the repository tokens; do not introduce a parallel palette.

| Role | Token | Value |
|---|---|---|
| Background | Deep Navy | `#0B0F19` |
| Primary emphasis | NAEOS Blue | `#2563FF` |
| Secondary accent | Electric Cyan | `#60A5FA` |
| Body / muted text | Slate | `#64748B` |
| Primary text | White | `#FFFFFF` |
| Secondary light text | Silver | `#E5E7EB` |

Typography roles:
- Display: **Sora**
- Body: **Inter**
- Technical labels: **JetBrains Mono**

Logo:
- Use the approved Logo Master v2.0.0 vector from `brand/master/naeos-primary-symbol-master-v2.svg`.
- Do not redraw, approximate, recolor outside approved variants, or alter its geometry.
- The supplied quote SVG intentionally uses a text-only NAEOS footer placeholder. Replace it with the approved lockup only when the correct asset variant is inserted and visually verified.

## Layout rules

1. Keep the statement dominant and left aligned.
2. Use short stacked lines; preserve the message hierarchy when adapting canvas ratios.
3. Keep essential copy and brand elements inside a minimum 92 px horizontal inset, consistent with the LinkedIn Portrait Master.
4. Use NAEOS Blue and Electric Cyan as restrained emphasis—not as a replacement for primary text.
5. Keep the footer quiet: approved logo/wordmark, series label, and naeos.dev.
6. Maintain readable contrast and avoid placing text over complex decoration.
7. Never distort the mark or lockup; preserve its aspect ratio.

## Content variants

### Quote
**Trust is not a feature. It is an engineering outcome.**

Caption angle: Trust in AI-assisted engineering comes from explicit authorization, observable execution, durable evidence, and independent verification—not from model confidence.

### Engineering insight
**An agent can propose an action. It cannot authorize itself.**

Caption angle: Separate agent intent from policy decisions and runtime authority. Revalidate at the last controllable boundary before consequential side effects.

### Evidence principle
**If the outcome cannot be verified, the claim is not yet evidence.**

Caption angle: Preserve observations and artifact references, then verify independently of the component that performed the work.

## Export and review checklist

- [ ] SVG remains editable and uses a 1080 × 1350 viewBox.
- [ ] Text remains within the safe area and is readable at mobile feed size.
- [ ] Palette and typography match `brand/brand.json` and `brand/brand.css`.
- [ ] Logo placeholder is replaced only with an approved Logo Master v2.0.0 asset before publication.
- [ ] Copy is technically accurate and does not imply blanket security or production guarantees.
- [ ] Export PNG at 1080 × 1350 px after visual rendering review.
- [ ] Confirm the 1:1 adaptation separately; do not crop the portrait master blindly.

## Accuracy guardrails

Describe NAEOS using the canonical sequence:

**Specification → NEIR → Validation + Policy → Agent Intent → Authorized Execution → Observation → Evidence → Independent Verification**

Do not collapse policy, authorization, execution, observation, and verification into one step. Do not claim that an agent's own output independently verifies itself.
