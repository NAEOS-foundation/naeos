## naeos supabase init

Initialize Supabase project configuration

### Synopsis

Configure a Supabase project. Reads from environment variables if flags are not set.

Env vars:
  SUPABASE_PROJECT_REF, SUPABASE_URL, SUPABASE_PUBLISHABEL_KEY (anon),
  SUPABASE_SECRET_KEY (service role), SUPABASE_JWKS_URL

Examples:
  naeos supabase init --project-ref abcdefg
  naeos supabase init --project-ref abcdefg --anon-key "eyJ..." --service-role-key "eyJ..."
  SUPABASE_URL=https://abc.supabase.co SUPABASE_PUBLISHABEL_KEY=eyJ... naeos supabase init --project-ref abc

```
naeos supabase init [flags]
```

### Options

```
      --anon-key string           Supabase anon/public key
  -h, --help                      help for init
      --jwks-url string           Supabase JWKS URL for JWT verification
      --project-ref string        Supabase project reference (required)
      --service-role-key string   Supabase service role key
      --url string                Supabase project URL (default: https://<ref>.supabase.co)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos supabase](naeos_supabase.md)	 - Supabase backend management

