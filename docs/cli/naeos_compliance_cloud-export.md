## naeos compliance cloud-export

Export audit log to cloud storage (S3, GCS, Azure Blob)

### Synopsis

Upload the audit log directly to cloud storage.

Examples:
  naeos compliance cloud-export --provider s3 --bucket my-bucket --access-key XXX --secret-key YYY --region us-east-1
  naeos compliance cloud-export --provider gcs --bucket my-bucket --access-key GOOG1XXX --secret-key YYY
  naeos compliance cloud-export --provider azure --account-name myaccount --account-key YYY --container my-container

```
naeos compliance cloud-export [flags]
```

### Options

```
      --access-key string     access key ID (S3, GCS)
      --account-key string    Azure storage account key
      --account-name string   Azure storage account name
  -a, --audit-file string     path to audit log file (default: ~/.naeos/audit.log)
      --bucket string         bucket or container name (required)
      --endpoint string       S3 custom endpoint (for MinIO)
  -h, --help                  help for cloud-export
      --prefix string         key prefix/path (default "audit/")
      --provider string       cloud provider: s3, gcs, azure (default "s3")
      --region string         S3 region (default: us-east-1)
      --secret-key string     secret access key (S3, GCS)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos compliance](naeos_compliance.md)	 - Compliance reporting and audit log export

