## naeos supabase sql

Execute SQL query via Supabase Management API

### Synopsis

Execute a SQL query against the Supabase database using the service role key.

Examples:
  naeos supabase sql "SELECT * FROM users"
  naeos supabase sql "CREATE TABLE test (id SERIAL PRIMARY KEY, name TEXT)"

```
naeos supabase sql [flags]
```

### Options

```
  -h, --help   help for sql
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos supabase](naeos_supabase.md)	 - Supabase backend management

