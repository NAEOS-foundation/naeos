## naeos cloud

Cloud deployment commands

### Synopsis

Deploy NAEOS projects to AWS, GCP, or Azure.

### Options

```
  -e, --env string          Environment (dev, staging, prod) (default "dev")
  -h, --help                help for cloud
  -i, --input-file string   Spec file with cloud configuration (overrides flags)
  -j, --project string      Cloud project name
  -p, --provider string     Cloud provider (aws, gcp, azure) (default "aws")
  -r, --region string       Cloud region
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos cloud deploy](naeos_cloud_deploy.md)	 - Deploy to cloud provider
* [naeos cloud export](naeos_cloud_export.md)	 - Export Terraform configuration
* [naeos cloud plan](naeos_cloud_plan.md)	 - Plan cloud deployment
* [naeos cloud status](naeos_cloud_status.md)	 - Show deployed resource status
* [naeos cloud types](naeos_cloud_types.md)	 - List supported resource types

