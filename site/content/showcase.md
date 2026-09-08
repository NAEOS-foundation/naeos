---
title: Showcase
description: Real projects and systems built with NAEOS.
---

These examples show representative patterns that can be modeled with NAEOS. They are illustrative, not customer references. Want to add your project? [Submit a showcase](https://github.com/NAEOS-foundation/naeos/discussions/new?category=show-and-tell).

<div class="showcase-grid">
  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Microservices</span>
      <span class="showcase-badge">Spec-driven</span>
    </div>
    <h3>Microservices Platform</h3>
    <p>One NAEOS specification can describe service boundaries, API contracts, and deployment configuration for a multi-service platform.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>One</strong> spec</span>
      <span class="showcase-stat"><strong>Go, TS</strong> outputs</span>
      <span class="showcase-stat"><strong>Shared</strong> model</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">AI</span>
      <span class="showcase-badge">Context</span>
    </div>
    <h3>AI Product Architecture</h3>
    <p>NAEOS can compile a single NEIR model into AI context bundles for coding assistants, reducing drift between architecture and implementation.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>NEIR</strong> model</span>
      <span class="showcase-stat"><strong>AI</strong> context</span>
      <span class="showcase-stat"><strong>Shared</strong> intent</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Serverless</span>
      <span class="showcase-badge">Event-driven</span>
    </div>
    <h3>Event-Driven System</h3>
    <p>NAEOS can model asynchronous flows, dependency boundaries, and generated artifacts for event-driven services and integrations.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>Async</strong> flow</span>
      <span class="showcase-stat"><strong>Generated</strong> artifacts</span>
      <span class="showcase-stat"><strong>Policy</strong> checks</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Governance</span>
      <span class="showcase-badge">Compliance</span>
    </div>
    <h3>Policy-Aware Platform</h3>
    <p>Governance rules, audit trails, and compliance-oriented policy templates can be applied through the NAEOS validation and review pipeline.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>Rules</strong> engine</span>
      <span class="showcase-stat"><strong>Audit</strong> trail</span>
      <span class="showcase-stat"><strong>Policy</strong> review</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Architecture</span>
      <span class="showcase-badge">NEIR</span>
    </div>
    <h3>Clean Architecture Project</h3>
    <p>Domain models, adapter boundaries, and service contracts can be represented in a single NEIR model and propagated to multiple output targets.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>One</strong> model</span>
      <span class="showcase-stat"><strong>Multiple</strong> outputs</span>
      <span class="showcase-stat"><strong>Clear</strong> boundaries</span>
    </div>
  </div>

  <div class="showcase-card showcase-card-add">
    <h3>Your Project Here</h3>
    <p>Built something with NAEOS? Share your story with the community.</p>
    <a href="https://github.com/NAEOS-foundation/naeos/discussions/new?category=show-and-tell" class="btn btn-primary btn-sm">Submit Project</a>
  </div>
</div>

## Try It Yourself — Demo Project

The official demo project lives in the repository: [`cmd/naeos/demo-app/`](https://github.com/NAEOS-foundation/naeos/tree/main/cmd/naeos/demo-app). It is a hexagonal-architecture Go app described entirely by a spec.

```bash
# 1. Build artifacts from the demo spec
naeos build --config cmd/naeos/demo-app/config.yaml --input cmd/naeos/demo-app/spec.yaml

# 2. Run the full pipeline with traceable output
naeos run --config cmd/naeos/demo-app/config.yaml --input cmd/naeos/demo-app/spec.yaml

# 3. Distribute the build across workers
naeos distributed --config cmd/naeos/demo-app/config.yaml --workers 4
```

Prefer to start from a template? Scaffold a full microservices starter:

```bash
naeos template init microservices-go -o .
naeos build --config naeos.yaml --input spec.yaml
```

Both projects are open source and small enough to read in minutes — a great way to see NAEOS in action.

## Why Teams Explore NAEOS

NAEOS is useful when a team wants one source of truth for architecture, code generation, policy checks, and AI context. It is especially relevant when specification drift, duplicated prompts, and fragmented governance create avoidable rework.
