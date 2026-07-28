## naeos ai

AI-powered assistance commands

### Synopsis

AI-powered commands for specification improvement and concept explanation.

Example:
  naeos ai suggest --input-file spec.yaml
  naeos ai explain pipeline
  naeos ai enrich --input-file spec.yaml --stream --provider anthropic

### Options

```
  -h, --help   help for ai
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos ai compile](naeos_ai_compile.md)	 - Compile a specification for a target AI agent using AI
* [naeos ai enrich](naeos_ai_enrich.md)	 - Enrich a specification with AI-powered best practices
* [naeos ai explain](naeos_ai_explain.md)	 - Explain a NAEOS concept
* [naeos ai suggest](naeos_ai_suggest.md)	 - Get AI suggestions for improving a specification

