---
title: Cookbook
description: Ready-to-use specification patterns for common architectures.
---

Copy-paste these spec templates to bootstrap your next project. Each recipe includes a complete NAEOS specification with explanations.

## Microservices API Gateway

Define a microservices architecture with API gateway, internal services, and shared database.

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
    kind: reverse-proxy
    port: 8080
  - name: user-api
    kind: rest
    port: 9001
  - name: product-api
    kind: grpc
    port: 9002
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]</code></pre>
</div>

**Key points:** Gateway handles routing, services are isolated, only gateway is publicly exposed.

---

## Serverless Event Pipeline

Build an event-driven serverless application with function isolation.

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: analytics-pipeline
modules:
  - name: event-ingestor
    path: ./functions/ingest
  - name: stream-processor
    path: ./functions/process
    dependencies: [event-ingestor]
  - name: data-analyzer
    path: ./functions/analyze
    dependencies: [stream-processor]
services:
  - name: ingest-api
    kind: lambda
  - name: process-worker
    kind: lambda
  - name: analyze-worker
    kind: lambda
architecture:
  pattern: serverless
deployment:
  strategy: serverless-framework
generation:
  languages: [python, typescript]</code></pre>
</div>

**Key points:** Each function is a separate module, dependencies form the event flow, no shared state.

---

## Hexagonal Clean Architecture

Implement domain-driven design with the hexagonal architecture pattern.

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
  - name: infrastructure
    path: ./internal/infrastructure
    dependencies: [adapters-outbound]
services:
  - name: rest-api
    kind: rest
    port: 8080
  - name: grpc-api
    kind: grpc
    port: 9090
architecture:
  pattern: hexagonal
generation:
  languages: [go, java]
  output_dir: ./src</code></pre>
</div>

**Key points:** Domain has zero dependencies, application depends on domain, adapters depend on application, infrastructure at the edge.

---

## Monolithic Application

A simpler starting point — single binary with layered modules.

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: monolith-app
modules:
  - name: core
    path: ./core
  - name: web
    path: ./web
    dependencies: [core]
  - name: database
    path: ./infra/db
    dependencies: [core]
services:
  - name: web-server
    kind: http
    port: 8080
architecture:
  pattern: monolithic
deployment:
  strategy: docker-compose
generation:
  languages: [go]
  output_dir: ./cmd</code></pre>
</div>

---

## AI Agent Platform

Build a GenAI service with multiple model providers and vector storage.

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
  - name: memory-store
    path: ./stores/memory
  - name: vector-db
    path: ./infra/vector
services:
  - name: api-gateway
    kind: reverse-proxy
    port: 8080
  - name: chat-api
    kind: rest
    port: 9001
  - name: streaming-ws
    kind: websocket
    port: 9002
architecture:
  pattern: microservices
ai:
  providers:
    - name: openai
      models: [gpt-4o, gpt-4o-mini]
    - name: anthropic
      models: [claude-opus-4, claude-sonnet-4]
generation:
  languages: [go, typescript, python]
  ai_instructions: true</code></pre>
</div>

---

## Event-Driven Architecture

Asynchronous messaging between services with workers and streams.

<div class="code-block">
<div class="code-block-header"><span>spec.yaml</span><button class="copy-btn" aria-label="Copy code"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>Copy</button></div>
<pre><code>project: event-platform
modules:
  - name: event-ingestor
    path: ./ingest
  - name: stream-processor
    path: ./process
    dependencies: [event-ingestor]
  - name: analytics
    path: ./analytics
    dependencies: [stream-processor]
  - name: notification
    path: ./notify
    dependencies: [stream-processor]
services:
  - name: ingestion-api
    kind: rest
    port: 8080
  - name: stream-worker
    kind: worker
    port: 9001
  - name: notification-ws
    kind: websocket
    port: 9002
architecture:
  pattern: event-driven
deployment:
  strategy: kubernetes
generation:
  languages: [go, typescript, python]
  output_dir: ./generated</code></pre>
</div>
