---
title: Ecosystem
description: Profiles, plugins, integrations, and community extensions for NAEOS.
---

## Built-in Profiles

NAEOS comes with 5 industry profiles that provide pre-configured patterns, rules, and templates for common domains.

<div class="eco-card" style="margin-bottom:1rem;">
<h3>SaaS Profile</h3>
<p>Multi-tenant architecture, subscription management, API rate limiting, and RBAC patterns for Software-as-a-Service applications.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#00ff88;"></span> Multi-tenant database patterns</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Subscription & billing integration</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> API key management & rate limiting</li>
</ul>
</div>

<div class="eco-card" style="margin-bottom:1rem;">
<h3>AI Agent Profile</h3>
<p>Agent-based architectures, LLM integration patterns, tool-use scaffolding, and context management for AI-powered applications.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#00ff88;"></span> Agent orchestration patterns</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> LLM provider abstraction layer</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> Tool-use and function calling</li>
</ul>
</div>

<div class="eco-card" style="margin-bottom:1rem;">
<h3>FinTech Profile</h3>
<p>Financial domain patterns, transaction processing, audit trails, and compliance rules for financial technology applications.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#00ff88;"></span> Transaction processing & ledger</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Audit trail & compliance logging</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> Regulatory rule enforcement</li>
</ul>
</div>

<div class="eco-card" style="margin-bottom:1rem;">
<h3>Healthcare Profile</h3>
<p>HIPAA-compliant patterns, FHIR API integration, patient data management, and security controls for healthcare applications.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#00ff88;"></span> HIPAA compliance scaffolding</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> FHIR resource definitions</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> PHI data handling patterns</li>
</ul>
</div>

<div class="eco-card">
<h3>Government Profile</h3>
<p>Government system patterns, regulatory compliance, document management, and security standards for public sector applications.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#00ff88;"></span> Regulatory compliance framework</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Document workflow automation</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> Security standard enforcement</li>
</ul>
</div>

## Plugin System

Extend NAEOS with WASM and native plugins. The Plugin SDK makes it easy to create custom:

- **Code generators** — Add new language adapters
- **Validators** — Custom validation rules
- **Deployers** — Deploy to any platform
- **Analyzers** — Custom analysis and reporting

## Marketplace

Publish and discover profiles, plugins, and templates through the NAEOS Marketplace.

- Search and install with `naeos marketplace`
- SHA-256 verified for security
- Version management and dependency resolution
- Community ratings and reviews