---
title: Getting Started with NAEOS — From Spec to Code in 5 Minutes
description: A hands-on tutorial showing how to declare a microservices project with NAEOS and generate Go + TypeScript code in under five minutes.
date: 2026-07-26
author: "NAEOS Foundation"
categories: ["tutorial"]
---

NAEOS lets you describe your entire project — modules, services, APIs, dependencies — in a single YAML spec, then generates production-ready code for your chosen languages and platforms.

In this tutorial you'll build a simple microservices project with a Go API gateway and a TypeScript auth service.

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- A terminal

Install NAEOS:

```bash
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

## Step 1 — Create a Spec

Start by initializing a project:

```bash
naeos init --template microservices my-app
cd my-app
```

This creates a `naeos.yaml` file. Open it and replace the contents with:

```yaml
project: my-app
version: "1.0"
description: A microservices demo project

modules:
  - name: gateway
    path: ./gateway
  - name: auth
    path: ./auth
    dependencies: []
  - name: api
    path: ./api
    dependencies: [auth]

services:
  - name: gateway
    kind: http
    port: 8080
    module: gateway
  - name: auth-service
    kind: grpc
    port: 50051
    module: auth

generation:
  languages: [go, typescript]
```

## Step 2 — Validate

Check that your spec is valid:

```bash
naeos validate
```

You should see `✓ Specification is valid`.

## Step 3 — Generate Code

Run the full pipeline to generate Go and TypeScript code:

```bash
naeos run --input naeos.yaml --output-dir ./out
```

In seconds, NAEOS produces:

```
out/
├── go/
│   ├── gateway/
│   │   ├── main.go
│   │   ├── handler.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── auth/
│       ├── server.go
│       ├── proto/
│       ├── go.mod
│       └── Dockerfile
└── typescript/
    ├── api/
    │   ├── src/
    │   ├── package.json
    │   └── tsconfig.json
    └── auth/
        ├── src/
        ├── package.json
        └── tsconfig.json
```

## Step 4 — Preview NEIR

See the intermediate representation NAEOS builds from your spec:

```bash
naeos run --input naeos.yaml --output-dir ./out --verbose
```

The NEIR model includes resolved dependencies, service endpoints, and architecture metadata that the generators use to produce correct, consistent code across all target languages.

## What's Next

- Explore the [documentation](/docs/getting-started/) for advanced spec features
- Try different templates: `naeos init --template fullstack my-app`
- Add AI context compilation with `naeos ai compile`
- Join the [community on GitHub](https://github.com/NAEOS-foundation/naeos)
