## naeos cicd

Generate CI/CD pipeline configuration

### Synopsis

Generate CI/CD pipeline configuration for GitHub Actions, GitLab CI, or Jenkins.

```
naeos cicd [flags]
```

### Options

```
  -h, --help                help for cicd
      --input-file string   Path to YAML/JSON spec file to override config
      --languages string    Comma-separated list of languages (go, python, node, etc.) (default "go")
  -o, --output string       Output file path
  -p, --platform string     CI/CD platform (github, gitlab, jenkins) (default "github")
      --project string      Project name for pipeline config (default "myapp")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

