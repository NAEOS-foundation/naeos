---
title: NEIR Model
description: The NAEOS Engineering Intermediate Representation — the canonical system model.
---

## Overview

NEIR (NAEOS Engineering Intermediate Representation) is the canonical model that represents the entire system being engineered. It is the single source of truth that flows through the pipeline, enabling all downstream processing — code generation, AI compilation, documentation, and deployment.

## NEIR Architecture

```text
┌─────────────────────────────────────────────┐
│                  NEIR Model                    │
│  ┌─────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ Project │ │ Modules  │ │ Services     │   │
│  │ Metadata│ │ & Deps   │ │ & APIs       │   │
│  └─────────┘ └──────────┘ └──────────────┘   │
│  ┌─────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ Storage │ │Infra     │ │ Security     │   │
│  │ & DB    │ │& Cloud   │ │ & Policies   │   │
│  └─────────┘ └──────────┘ └──────────────┘   │
│  ┌─────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ AI      │ │ Docs     │ │ Deployment   │   │
│  │Config   │ │& Specs   │ │ & CI/CD      │   │
│  └─────────┘ └──────────┘ └──────────────┘   │
└─────────────────────────────────────────────┘
```

## Core Components

### Project Metadata

Top-level information about the system being built:

- Name, version, and description
- Architecture pattern (microservices, serverless, monolithic, hexagonal)
- Domain model and bounded contexts
- Team and ownership information

### Module Structure

The module graph captures all code components and their relationships:

- Module definitions with paths and types
- Dependency edges with version constraints
- Module groups and namespaces
- Entry points and exports

### Service Definitions

Services represent runnable components with:

- Service type (REST, GraphQL, WebSocket, gRPC, HTTP)
- Port mappings and protocols
- Endpoint definitions with request/response schemas
- Middleware and interceptor chains
- Health check configurations

### API Contracts

API definitions including:

- RESTful endpoints with methods and paths
- GraphQL schema types and resolvers
- WebSocket event types
- Request/response schemas
- Authentication and authorization rules

### Storage & Database

Data layer configuration:

- Database engines (PostgreSQL, Redis, MongoDB, etc.)
- Table/collection schemas
- Migration configurations
- Connection pooling settings
- Backup and replication policies

### Infrastructure & Cloud

Infrastructure as code definitions:

- Kubernetes resources and manifests
- Docker container configurations
- Cloud provider resources (AWS, GCP, Azure)
- Network and VPC configuration
- Load balancer and auto-scaling rules

### Security & Policies

Security model:

- Authentication providers and methods
- RBAC role definitions
- Policy rules and constraints
- Encryption settings
- Audit trail configuration

### AI Integration

AI-specific configuration:

- LLM provider configurations
- Model selections and parameters
- Prompt templates and context bundling
- Tool and function definitions
- Agent orchestration settings

### Documentation

Documentation requirements:

- API documentation specifications
- Architecture decision records
- README and onboarding guides
- changelog templates

### Deployment & CI/CD

Deployment configuration:

- Environment definitions (dev, staging, production)
- CI/CD pipeline templates
- Blue/green and canary deployment settings
- Rollback and health check policies

## Accessing the NEIR Model

```bash
# Export generated artifacts from the NEIR model
naeos export --input spec.yaml --output-dir ./out

# Inspect the model as JSON
naeos inspect --input-file spec.yaml --output json --output-file neir.json

# Validate the model
naeos validate --input-file spec.yaml
```

## NEIR in AI Context

The NEIR model is also used to generate AI context bundles. When compiled, it produces architecture-aware instruction sets that help AI coding assistants understand the full system context:

```bash
naeos context --input-file spec.yaml --output json
```

## Benefits

- **Single source of truth** — One model drives all outputs
- **Traceability** — Every artifact links back to the spec
- **Consistency** — Cross-language and cross-platform alignment
- **Analysis** — Query and analyze the model for insights
- **Evolution** — Track changes and migrate between versions
