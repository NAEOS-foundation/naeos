## naeos evidence verify-bundle

Verify an exported control-plane evidence bundle

### Synopsis

Verify a canonical EvidenceBundle without contacting the control plane.

The verifier reads a serialized EvidenceBundle, validates its decision and
execution bindings, and recomputes the evidence digest. It is read-only and
does not evaluate policy, execute actions, or require a live control-plane
service.

Examples:
  naeos evidence verify-bundle --input-file evidence.json
  naeos evidence verify-bundle --input-file evidence.json --output json

```
naeos evidence verify-bundle [flags]
```

### Options

```
  -h, --help                help for verify-bundle
      --input-file string   path to exported EvidenceBundle JSON
      --output string       output format: table or json (default "table")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos evidence](naeos_evidence.md)	 - Governance evidence store — immutable audit trail
