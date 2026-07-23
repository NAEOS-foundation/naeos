## naeos supabase init

Initialize Supabase project configuration

### Synopsis

Configure a Supabase project. Reads SUPABASE_URL, SUPABASE_ANON_KEY,
and SUPABASE_SERVICE_ROLE_KEY from environment variables if flags are not set.

Examples:
  naeos supabase init --project-ref abcdefg
  naeos supabase init --project-ref abcdefg --anon-key "eyJ..." --service-role-key "eyJ..."
  SUPABASE_URL=https://abc.supabase.co SUPABASE_ANON_KEY=eyJ... naeos supabase init --project-ref abc

```
naeos supabase init [flags]
```

### Options

```
      --anon-key string           Supabase anon/public key
  -h, --help                      help for init
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

