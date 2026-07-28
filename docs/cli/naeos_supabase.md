## naeos supabase

Supabase backend management

### Synopsis

Manage Supabase projects: authentication, storage, and SQL queries.

Examples:
  naeos supabase init
  naeos supabase auth signup --email user@example.com --password secret
  naeos supabase auth signin --email user@example.com --password secret
  naeos supabase auth user
  naeos supabase storage list-buckets
  naeos supabase sql "SELECT * FROM users"

```
naeos supabase [flags]
```

### Options

```
  -h, --help   help for supabase
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos supabase auth](naeos_supabase_auth.md)	 - Supabase authentication management
* [naeos supabase init](naeos_supabase_init.md)	 - Initialize Supabase project configuration
* [naeos supabase sql](naeos_supabase_sql.md)	 - Execute SQL query via Supabase Management API
* [naeos supabase status](naeos_supabase_status.md)	 - Check Supabase connection status
* [naeos supabase storage](naeos_supabase_storage.md)	 - Supabase storage management

