# Storyboard Video Demo NAEOS

## Format
- Durasi: 90 detik
- Gaya: clean technical explainer
- Target: YouTube, LinkedIn, GitHub announcement, website promo

---

## Scene 1 — Hook (0:00–0:08)

Visual:
- Close-up layar editor kode
- Prompt AI muncul berulang
- Chat windows stack di layar
- Teks overlay: "AI can generate code fast"

Narration:
"AI can generate code fast. But most teams still lose architecture, governance, and consistency."

Camera / motion:
- quick cuts, dynamic zoom
- harsh but modern editing style

---

## Scene 2 — Problem (0:08–0:20)

Visual:
- Tiga panel: tickets, prompt, docs, code
- Kursor bergerak antar panel
- Muncul label: "Prompt drift", "Spec drift", "Architecture drift"

Narration:
"Engineering intent gets scattered across tickets, prompts, docs, and AI tools."

On-screen text:
"The problem is not generation. It's consistency."

---

## Scene 3 — Product reveal (0:20–0:35)

Visual:
- Logo NAEOS muncul
- Ditengah: pipeline diagram
- Label: Parse → Normalize → Resolve → NEIR → Validate → Generate → AI Context

Narration:
"NAEOS changes that. It starts with one specification and builds a shared engineering model."

On-screen text:
"Specify once. Build anywhere."

---

## Scene 4 — Demo spec (0:35–0:46)

Visual:
- Editor menampilkan file `spec.yaml`
- Simple YAML example

Example spec shown:
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
"A team defines the system once: modules, services, architecture, and generation targets."

---

## Scene 5 — Internal pipeline (0:46–1:00)

Visual:
- Pipeline animates forward step by step
- Each stage lights up sequentially
- Include labels: parse, normalize, resolve, NEIR, validate, generate

Narration:
"NAEOS reads that specification, normalizes it, resolves dependency structure, builds the NEIR model, and validates the design before generation."

On-screen text:
"One source of truth"

---

## Scene 6 — Generated output (1:00–1:12)

Visual:
- Folder tree appears
- Generated project structure with modules and docs
- AI context bundle / instruction set generated

Narration:
"Then it generates artifacts, documentation, and AI-ready context from the same model."

On-screen text:
"Generated artifacts + AI context"

---

## Scene 7 — Why it matters (1:12–1:22)

Visual:
- Two molecules: left = fragmented AI workflow, right = NAEOS workflow
- Clear contrast with labels

Narration:
"This is different from scaffolding. It gives engineering teams a structured, validated system model that AI tools can actually reason over."

On-screen text:
"From specification to system"

---

## Scene 8 — Closing CTA (1:22–1:30)

Visual:
- NAEOS GitHub repo and branding
- Buttons: Explore GitHub, Try NAEOS, Read the Architecture

Narration:
"NAEOS is open source and built in public. Explore the repository, try the quick start, and see how specification-driven engineering can bring more consistency to AI-assisted development."

Final text:
"Specify once. Build anywhere."

---

## Narration text (full script)

"AI can generate code fast. But most teams still lose architecture, governance, and consistency as they scale.

Engineering intent gets scattered across tickets, prompts, docs, and AI tools.

NAEOS changes that. It starts with one specification and builds a shared engineering model.

A team defines the system once: modules, services, architecture, and generation targets.

NAEOS reads that specification, normalizes it, resolves dependency structure, builds NEIR, and validates the design before generation.

Then it generates artifacts, documentation, and AI-ready context from the same source of truth.

This is different from scaffolding. It gives teams a structured, validated system model that AI tools can reason over.

NAEOS is open source and built in public. Explore the repository, try the quick start, and see how specification-driven engineering can bring more consistency to AI-assisted development.

Specify once. Build anywhere."

---

## Production notes

### Camera style
- Clean product-demo aesthetic
- Crisp motion graphics
- Minimal but technical UI
- Avoid overuse of flashy effects

### Audio style
- Modern, confident voiceover
- Clean background music with low volume
- Short, precise pacing

### Subtitle style
- Big, high-contrast text
- 2–5 words per line max
- Keep on screen 2–3 seconds each

---

## Suggested title cards

- "AI can generate code. NAEOS structures the system."
- "Specify once. Build anywhere."
- "From specification to system."
- "Why AI-assisted engineering needs structure"
