## naeos init

Initialize a new NAEOS project or generate config

### Synopsis

Initialize a new NAEOS project with a configuration file.

Available templates:
  basic          — Minimal config with Go (default)
  microservices  — Multi-service microservices architecture
  rest-api       — Single REST API service
  fullstack      — Fullstack with backend + frontend + worker
  kubernetes     — Production-ready Kubernetes deployment
  supabase       — Supabase backend with pipeline + auth + storage
  hcl            — HCL format specification
 
Example:
  naeos init
  naeos init --template microservices
  naeos init --template rest-api --name my-api
  naeos init --list-templates

```
naeos init [flags]
```

### Options

```
  -h, --help              help for init
      --list-templates    list all available templates
  -n, --name string       project name (replaces default in template)
  -o, --output string     path for the generated config file (default "naeos.yaml")
  -t, --template string   template to use (basic, microservices, rest-api, fullstack, kubernetes, supabase, hcl) (default "basic")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

