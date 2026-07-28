## naeos deploy

Deploy the pipeline output to a target environment

### Synopsis

Deploy generated artifacts to a target environment using configured deployment tools.

Supported targets:
  docker    — Build and push Docker images
  k8s       — Apply Kubernetes manifests
  compose   — Docker Compose up
  ssh       — Remote deployment via SSH
  rsync     — File sync via rsync
  local     — Local directory copy

Example:
  naeos deploy --target docker
  naeos deploy --target k8s --env staging
  naeos deploy --target compose --dry-run
  naeos deploy --target rsync --env production

```
naeos deploy [flags]
```

### Options

```
      --config string   path to config file
      --dry-run         preview deployment without executing
  -e, --env string      target environment (default "development")
  -h, --help            help for deploy
  -t, --target string   deployment target: docker, k8s, compose, ssh, rsync, local (default "local")
```

### Options inherited from parent commands

```
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

