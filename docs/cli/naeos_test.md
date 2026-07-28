## naeos test

Run tests for generated code

### Synopsis

Run tests across all detected or specified languages.

Automatically detects project languages and runs appropriate test commands:
  - Go: go test -v ./...
  - TypeScript/Node: npm test / pnpm test
  - Python: python -m pytest -v
  - Java: mvn test / ./gradlew test
  - Rust: cargo test --verbose

Example:
  naeos test
  naeos test --language go --language typescript
  naeos test --dir ./my-project --verbose
  naeos test --parallel --timeout 30
  naeos test --output json

```
naeos test [flags]
```

### Options

```
      --dir string             working directory for tests (default ".")
  -h, --help                   help for test
      --language stringArray   target language (go, typescript, python, java, rust)
  -o, --output string          output format: text, json (default "text")
      --parallel               run tests for different languages in parallel
      --timeout int            test timeout in seconds (0 = no timeout)
  -v, --verbose                verbose test output
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

