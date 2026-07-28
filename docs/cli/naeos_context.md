## naeos context

Generate AI context bundles from specifications

### Synopsis

Generate context bundles optimized for LLM consumption.
Produces structured markdown or plain text summaries of your project.

Example:
  naeos context --input-file spec.yaml
  naeos context --input 'project: myapp' --output json
  naeos context --input-file spec.yaml --output markdown

```
naeos context [flags]
```

### Options

```
  -h, --help                 help for context
      --input string         specification input to process
      --input-file string    path to a specification file
      --output string        output format: markdown, plain, json, or yaml (default "markdown")
      --output-file string   optional file path to write the output
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

