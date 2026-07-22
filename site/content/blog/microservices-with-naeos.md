---
title: Building Microservices with NAEOS — A Step-by-Step Guide
description: Learn how to design, generate, and deploy microservices using NAEOS declarative specifications.
date: 2026-06-15
categories: ["tutorial"]
---

In this tutorial, we'll build a complete microservices application using NAEOS. You'll learn how to define services, manage dependencies, generate code, and deploy.

## Prerequisites

- NAEOS installed ([see installation guide](/docs/installation/))
- Basic knowledge of YAML and microservices concepts

## Step 1: Define the Specification

Create a `spec.yaml` file that describes your microservices architecture:

```yaml
project: ecommerce-platform
version: "1.0"
modules:
  - name: api-gateway
    path: ./api-gateway
    description: API Gateway entry point
    dependencies: [user-service, order-service, product-service]
  - name: user-service
    path: ./services/users
    description: User management and authentication
    dependencies: [shared-db]
  - name: order-service
    path: ./services/orders
    description: Order processing and management
    dependencies: [user-service, payment-service]
  - name: product-service
    path: ./services/products
    description: Product catalog management
    dependencies: [shared-db]
  - name: payment-service
    path: ./services/payments
    description: Payment processing
    dependencies: []
  - name: shared-db
    path: ./infra/db
    description: Shared database layer
services:
  - name: gateway
    kind: reverse-proxy
    port: 8080
  - name: user-api
    kind: rest
    port: 9001
  - name: order-api
    kind: rest
    port: 9002
  - name: product-api
    kind: rest
    port: 9003
architecture:
  pattern: microservices
  description: Event-driven microservices for e-commerce
generation:
  languages: [go, typescript]
  output_dir: ./generated
```

## Step 2: Run the Pipeline

```bash
naeos run --input-file spec.yaml --output-dir ./out
```

NAEOS will parse, validate, and generate your project structure with all the defined modules and services.

## Step 3: Compile for AI Assistance

```bash
naeos compile --all --input-file spec.yaml
```

This generates AI instruction sets so your coding assistants understand the architecture.

## Step 4: Generate Documentation

```bash
naeos docgen --input-file spec.yaml --output-dir ./docs
```

## Conclusion

You now have a complete microservices project with generated code, documentation, and AI context — all from a single specification file.