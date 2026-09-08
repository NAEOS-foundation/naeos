---
title: Feature Comparison
description: How NAEOS compares to other project generation and scaffolding tools.
---

NAEOS is not only a code generator. It is a declarative engineering platform with a pipeline, AI compiler, governance hooks, and multi-language support. Here is a practical comparison with similar tools.

## Comparison Table

<div class="comparison-scroll">
<table class="comparison-table">
  <thead>
    <tr>
      <th>Feature</th>
      <th class="highlight-col">NAEOS</th>
      <th>Cookie Cutter</th>
      <th>Copier</th>
      <th>OpenAPI Gen</th>
      <th>Hygen</th>
      <th>Yeoman</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Declarative YAML/JSON spec</td>
      <td class="highlight-col"><span class="check-yes">Yes</span></td>
      <td><span class="check-partial">Partial</span></td>
      <td><span class="check-yes">Yes</span></td>
      <td><span class="check-yes">Yes</span></td>
      <td><span class="check-no">No</span></td>
      <td><span class="check-no">No</span></td>
    </tr>
    <tr>
      <td>Multi-language code gen</td>
      <td class="highlight-col"><span class="check-yes">5 languages</span></td>
      <td><span class="check-yes">Any</span></td>
      <td><span class="check-yes">Any</span></td>
      <td><span class="check-partial">API only</span></td>
      <td><span class="check-yes">Any</span></td>
      <td><span class="check-yes">Any</span></td>
    </tr>
    <tr>
      <td>AI context generation</td>
      <td class="highlight-col"><span class="check-yes">7 platforms</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
    </tr>
    <tr>
      <td>Pipeline engine (DAG)</td>
      <td class="highlight-col"><span class="check-yes">11-stage</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
    </tr>
    <tr>
      <td>Built-in governance & policy</td>
      <td class="highlight-col"><span class="check-yes">RBAC + audit</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
    </tr>
    <tr>
      <td>Plugin system</td>
      <td class="highlight-col"><span class="check-yes">WASM + native</span></td>
      <td><span class="check-partial">Jinja only</span></td>
      <td><span class="check-partial">Jinja only</span></td>
      <td><span class="check-partial">Mustache</span></td>
      <td><span class="check-yes">JS plugins</span></td>
      <td><span class="check-yes">JS plugins</span></td>
    </tr>
    <tr>
      <td>Template marketplace</td>
      <td class="highlight-col"><span class="check-yes">Built-in</span></td>
      <td><span class="check-partial">Cookie Cutter</span></td>
      <td><span class="check-partial">Copier</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-yes">npm registry</span></td>
    </tr>
    <tr>
      <td>Interactive dashboard</td>
      <td class="highlight-col"><span class="check-yes">WebSocket</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
    </tr>
    <tr>
      <td>Watch mode (hot-reload)</td>
      <td class="highlight-col"><span class="check-yes">Yes</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-yes">Yes</span></td>
      <td><span class="check-yes">Yes</span></td>
    </tr>
    <tr>
      <td>CI/CD integration</td>
      <td class="highlight-col"><span class="check-yes">Native</span></td>
      <td><span class="check-partial">Manual</span></td>
      <td><span class="check-partial">Manual</span></td>
      <td><span class="check-yes">Yes</span></td>
      <td><span class="check-partial">Manual</span></td>
      <td><span class="check-partial">Manual</span></td>
    </tr>
    <tr>
      <td>SSO (OIDC / SAML / LDAP)</td>
      <td class="highlight-col"><span class="check-yes">Built-in</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
    </tr>
    <tr>
      <td>Compliance (SOC 2 / HIPAA / GDPR)</td>
      <td class="highlight-col"><span class="check-yes">Policy templates</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
      <td><span class="check-no">—</span></td>
    </tr>
    <tr>
      <td>Language</td>
      <td class="highlight-col"><span class="check-yes">Go</span></td>
      <td><span class="check-yes">Python</span></td>
      <td><span class="check-yes">Python</span></td>
      <td><span class="check-yes">Java/TS</span></td>
      <td><span class="check-yes">Node.js</span></td>
      <td><span class="check-yes">Node.js</span></td>
    </tr>
    <tr>
      <td>Open source license</td>
      <td class="highlight-col"><span class="check-yes">Apache 2.0</span></td>
      <td><span class="check-yes">BSD</span></td>
      <td><span class="check-yes">MIT</span></td>
      <td><span class="check-yes">Apache 2.0</span></td>
      <td><span class="check-yes">MIT</span></td>
      <td><span class="check-yes">BSD</span></td>
    </tr>
  </tbody>
</table>
</div>

## Why NAEOS?

Several things make NAEOS fundamentally different from traditional project generators:

### Spec-Driven, Not Template-Driven

Cookie Cutter and Copier start with templates — you define placeholder variables inside file templates. NAEOS starts with a **specification** — a structured YAML/JSON document that describes the architecture, modules, services, and dependencies. Code generation is just one output of the pipeline.

### Architectural Knowledge

NAEOS builds the **NEIR model** — a canonical intermediate representation of your entire system. This model understands dependencies, architecture patterns, service boundaries, and governance rules. Templates don't have this awareness.

### AI-Native Design

AI context generation is a first-class feature, not an afterthought. NAEOS compiles NEIR into instruction sets for seven documented AI platforms. The comparison focuses on the capabilities documented for each tool.

### Built for Teams

Governance, RBAC, audit trails, and policy evaluation are documented parts of the pipeline. NAEOS does more than generate code: it provides mechanisms for reviewing engineering policies and artifacts.

## When to Use Each Tool

| Tool | Best For |
|------|----------|
| **NAEOS** | Full-stack projects, microservices, AI-assisted development, policy-aware workflows |
| **Cookie Cutter** | Quick Python project scaffolding from templates |
| **Copier** | Python projects with template updates (data-driven) |
| **OpenAPI Generator** | Generating API clients and server stubs from OpenAPI specs |
| **Hygen** | Generating individual files within an existing project |
| **Yeoman** | Web application scaffolding with interactive prompts |

## Get Started

Ready to try NAEOS? [Install NAEOS](/download/) and create your first project using the documented quick start.

See also: [Features](/features/), [Documentation](/docs/getting-started/), [Use Cases](/use-cases/)
