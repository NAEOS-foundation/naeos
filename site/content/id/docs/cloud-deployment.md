---
title: Deployment Cloud
description: Deploy proyek yang di-generate NAEOS ke AWS, GCP, atau Azure dengan perencanaan dan provisioning cloud terintegrasi.
weight: 12
---

NAEOS menyertakan dukungan deployment cloud bawaan. Anda dapat merencanakan, mendeploy, dan menghancurkan infrastruktur cloud langsung dari spesifikasi Anda menggunakan perintah `cloud`.

## Ikhtisar

Integrasi cloud bekerja dalam tiga fase:

1. **Plan** — Hasilkan rencana deployment Terraform HCL dari spesifikasi Anda
2. **Deploy** — Eksekusi rencana terhadap provider cloud pilihan Anda
3. **Destroy** — Hancurkan resource yang di-deploy saat tidak diperlukan

## Provider yang Didukung

| Provider | Layanan | Status |
|----------|---------|--------|
| AWS | EC2, ECS, Lambda, RDS, S3, CloudFront, VPC | Stable |
| GCP | GCE, Cloud Run, Cloud SQL, Cloud Storage | Stable |
| Azure | AKS, Azure Functions, SQL Database, Blob Storage | Beta |

## Konfigurasi

Tambahkan konfigurasi cloud ke spesifikasi Anda:

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

## Perintah CLI

### Plan

Hasilkan rencana deployment tanpa menerapkannya:

```bash
naeos cloud plan --provider aws --region us-east-1 --input-file spec.yaml
```

Ini menghasilkan direktori `terraform/` berisi file HCL siap untuk `terraform init && terraform apply`.

### Deploy

Eksekusi deployment:

```bash
naeos cloud deploy --provider aws --region us-east-1 --input-file spec.yaml
```

Perintah deploy:
1. Menghasilkan rencana Terraform
2. Menjalankan `terraform init`
3. Menjalankan `terraform plan` dan menampilkan rencana eksekusi
4. Menerapkan rencana (dengan prompt konfirmasi)
5. Menghasilkan ID resource dan endpoint

### Status

Periksa status resource yang di-deploy:

```bash
naeos cloud status
```

### Destroy

Hancurkan semua resource yang di-deploy:

```bash
naeos cloud destroy --provider aws --region us-east-1
```

## Tipe Resource

| Tipe Resource | AWS | GCP | Azure |
|---------------|-----|-----|-------|
| `compute-instance` | EC2 | GCE | VM |
| `container-service` | ECS/Fargate | Cloud Run | AKS |
| `serverless-function` | Lambda | Cloud Functions | Azure Functions |
| `database` | RDS | Cloud SQL | SQL Database |
| `storage-bucket` | S3 | Cloud Storage | Blob Storage |
| `cdn` | CloudFront | Cloud CDN | Azure CDN |
| `queue` | SQS | Pub/Sub | Service Bus |
| `cache` | ElastiCache | Memorystore | Redis Cache |

## Estimasi Biaya

Perintah `plan` menyertakan estimasi biaya bulanan:

```bash
naeos cloud plan --provider aws --input-file spec.yaml --estimate-cost
```

Output:

```
Resource                        Est. Bulanan
─────────────────────────────────────────────
ecs-service/api (2 tasks)       $73.20
rds/postgres-db (db.t3.micro)   $12.40
s3/storage-bucket                $1.20
sqs/events-queue                 $0.40
─────────────────────────────────────────────
Total estimasi:                  $87.20/bulan
```

Lihat juga: [Pipeline Engine](/docs/pipeline-engine/), [Bahasa Spesifikasi](/docs/spec-language/), [Governance](/docs/governance/)
