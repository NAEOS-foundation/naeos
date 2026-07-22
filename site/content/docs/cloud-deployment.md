---
title: Cloud Deployment
description: Deploy NAEOS-generated projects to AWS, GCP, or Azure with integrated cloud planning and provisioning.
weight: 12
---

NAEOS includes built-in cloud deployment support. You can plan, deploy, and destroy cloud infrastructure directly from your specifications using the `cloud` commands.

## Overview

The cloud integration works in three phases:

1. **Plan** — Generate a Terraform HCL deployment plan from your spec
2. **Deploy** — Execute the plan against your chosen cloud provider
3. **Destroy** — Tear down deployed resources when no longer needed

## Supported Providers

| Provider | Services | Status |
|----------|----------|--------|
| AWS | EC2, ECS, Lambda, RDS, S3, CloudFront, VPC | Stable |
| GCP | GCE, Cloud Run, Cloud SQL, Cloud Storage | Stable |
| Azure | AKS, Azure Functions, SQL Database, Blob Storage | Beta |

## Configuration

Add cloud configuration to your specification:

```yaml
project: my-service
modules:
  - name: api
    path: ./api
  - name: worker
    path: ./worker
services:
  - name: api-server
    kind: rest
    port: 8080
  - name: queue-worker
    kind: worker
deployment:
  provider: aws
  region: us-east-1
  strategy: ecs-fargate
  resources:
    - type: ecs-service
      name: api
      module: api
      config:
        cpu: 512
        memory: 1024
        desired_count: 2
    - type: sqs-queue
      name: events
      config:
        visibility_timeout: 300
        retention_period: 1209600
```

## CLI Commands

### Plan

Generate a deployment plan without applying it:

```bash
naeos cloud plan --provider aws --region us-east-1 --input-file spec.yaml
```

This produces a `terraform/` directory containing HCL files ready for `terraform init && terraform apply`.

### Deploy

Execute the deployment:

```bash
naeos cloud deploy --provider aws --region us-east-1 --input-file spec.yaml
```

The deploy command:
1. Generates the Terraform plan
2. Runs `terraform init`
3. Runs `terraform plan` and shows the execution plan
4. Applies the plan (with confirmation prompt)
5. Outputs resource IDs and endpoints

### Status

Check the status of deployed resources:

```bash
naeos cloud status
```

### Destroy

Tear down all deployed resources:

```bash
naeos cloud destroy --provider aws --region us-east-1
```

## Resource Types

| Resource Type | AWS | GCP | Azure |
|---------------|-----|-----|-------|
| `compute-instance` | EC2 | GCE | VM |
| `container-service` | ECS/Fargate | Cloud Run | AKS |
| `serverless-function` | Lambda | Cloud Functions | Azure Functions |
| `database` | RDS | Cloud SQL | SQL Database |
| `storage-bucket` | S3 | Cloud Storage | Blob Storage |
| `cdn` | CloudFront | Cloud CDN | Azure CDN |
| `queue` | SQS | Pub/Sub | Service Bus |
| `cache` | ElastiCache | Memorystore | Redis Cache |

## Environment Variables

The cloud commands use these environment variables for authentication:

| Variable | Description | Required |
|----------|-------------|----------|
| `AWS_ACCESS_KEY_ID` | AWS access key | For AWS |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | For AWS |
| `AWS_REGION` | AWS region | For AWS |
| `GOOGLE_PROJECT_ID` | GCP project ID | For GCP |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to GCP service account key | For GCP |
| `AZURE_SUBSCRIPTION_ID` | Azure subscription ID | For Azure |
| `AZURE_TENANT_ID` | Azure tenant ID | For Azure |
| `AZURE_CLIENT_ID` | Azure client ID | For Azure |
| `AZURE_CLIENT_SECRET` | Azure client secret | For Azure |

## Cost Estimation

The `plan` command includes an estimated monthly cost breakdown:

```bash
naeos cloud plan --provider aws --input-file spec.yaml --estimate-cost
```

Output:

```
Resource                        Monthly Est.
─────────────────────────────────────────────
ecs-service/api (2 tasks)       $73.20
rds/postgres-db (db.t3.micro)   $12.40
s3/storage-bucket                $1.20
sqs/events-queue                 $0.40
─────────────────────────────────────────────
Total estimated:                $87.20/month
```

See also: [Pipeline Engine](/docs/pipeline-engine/), [Spec Language](/docs/spec-language/), [Governance](/docs/governance/)
