---
title: Cookbook
description: Pola spesifikasi siap pakai untuk arsitektur umum.
---

Salin-tempel template spesifikasi ini untuk memulai proyek Anda berikutnya.

## Microservices API Gateway

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: microservices-app
modules:
  - name: api-gateway
    path: ./gateway
    dependencies: [user-service, product-service]
  - name: user-service
    path: ./services/users
    dependencies: [db]
  - name: product-service
    path: ./services/products
    dependencies: [db]
  - name: db
    path: ./infra/db
services:
  - name: gateway
    kind: http
    port: 8080
  - name: user-api
    kind: http
    port: 9001
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]</code></pre>
</div>

## Serverless Event Pipeline

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: analytics-pipeline
modules:
  - name: event-ingestor
    path: ./functions/ingest
  - name: stream-processor
    path: ./functions/process
    dependencies: [event-ingestor]
services:
  - name: ingest-api
    kind: worker
  - name: process-worker
    kind: worker
architecture:
  pattern: serverless
generation:
  languages: [python, typescript]</code></pre>
</div>

## Arsitektur Hexagonal

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: clean-arch-app
modules:
  - name: domain
    path: ./internal/domain
  - name: application
    path: ./internal/application
    dependencies: [domain]
  - name: adapters-inbound
    path: ./internal/adapters/inbound
    dependencies: [application]
  - name: adapters-outbound
    path: ./internal/adapters/outbound
    dependencies: [application]
architecture:
  pattern: hexagonal
generation:
  languages: [go, java]</code></pre>
</div>

## Platform AI Agent

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: ai-agent-platform
modules:
  - name: agent-orchestrator
    path: ./orchestrator
    dependencies: [llm-provider, memory-store]
  - name: llm-provider
    path: ./providers/llm
    dependencies: [vector-db]
services:
  - name: api-gateway
    kind: http
    port: 8080
  - name: chat-api
    kind: http
    port: 9001
architecture:
  pattern: microservices
generation:
  languages: [go, typescript, python]
  ai_instructions: true</code></pre>
</div>

## Arsitektur Event-Driven

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: event-platform
modules:
  - name: event-ingestor
    path: ./ingest
  - name: stream-processor
    path: ./process
    dependencies: [event-ingestor]
  - name: notification
    path: ./notify
    dependencies: [stream-processor]
services:
  - name: ingestion-api
    kind: http
    port: 8080
  - name: stream-worker
    kind: worker
architecture:
  pattern: event-driven
generation:
  languages: [go, typescript, python]</code></pre>
</div>
