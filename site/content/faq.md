---
title: Frequently Asked Questions
description: Common questions about NAEOS and declarative engineering.
---

<div class="faq-list">
<div class="faq-item">
<button class="faq-question">
<span>What is NAEOS?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>NAEOS (Nusantara Engineering & Architecture Operating System) is a declarative engineering platform that transforms YAML/JSON specifications into validated, multi-language software systems. It is an engineering runtime — not just a project generator — that understands specifications, builds an internal model (NEIR), orchestrates execution plans, generates artifacts, validates results, and keeps projects aligned with specifications throughout their lifecycle.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>How is NAEOS different from project generators?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>Unlike static project generators (like `create-react-app` or cookiecutters), NAEOS is a full engineering runtime. It doesn't just create files once — it maintains an ongoing relationship between your specification and your code. NAEOS validates, compiles to AI instruction sets, generates documentation, and adapts as your specification evolves.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>What languages are supported?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>NAEOS currently supports code generation for Go, TypeScript, Python, Java, and Rust. Each language has a dedicated adapter that follows best practices and idiomatic patterns for that language.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>What AI coding assistants are supported?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>NAEOS compiles specifications into AI instruction sets for 6 platforms: GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, and OpenCode. Each adapter generates platform-specific context files that help your AI assistant understand your project architecture.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Do I need to know Go to use NAEOS?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>No. NAEOS is a CLI tool written in Go, but you only need to write YAML/JSON specifications. The output can be in any of the 5 supported languages. You don't need any Go knowledge to use NAEOS effectively.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Can NAEOS integrate with existing projects?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>Yes. NAEOS can be configured to generate artifacts into existing project structures. You can start using NAEOS incrementally — define specs for new modules while keeping your existing code intact. The diff engine helps you understand what changed between spec revisions.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Is NAEOS free?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>Yes. NAEOS is completely free and open source under the Apache License 2.0. You can use it for personal projects, commercial applications, or enterprise deployments without any licensing fees.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>How do I get started?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>Check out our <a href="/docs/getting-started/">Getting Started guide</a>. You can install NAEOS via Go, Docker, or download a binary from GitHub Releases. Create a YAML spec, run <code>naeos run</code>, and within minutes you'll have generated code.</p>
</div>
</div>
</div>