## naeos scaffold

Generate a starter project scaffold

### Synopsis

Generate a starter project scaffold with all necessary files.

Example:
  naeos scaffold --name my-app
  naeos scaffold --name my-app --output ./my-app --language typescript
  naeos scaffold --name my-app --language go --language python

```
naeos scaffold [flags]
```

### Options

```
  -h, --help                   help for scaffold
      --language stringArray   target language for code generation (go, typescript, python, java, rust)
      --name string            project name for the scaffold
      --output string          directory where scaffold files will be created
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

