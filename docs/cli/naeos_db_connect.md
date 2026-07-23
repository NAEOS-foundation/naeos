## naeos db connect

Connect to a database

```
naeos db connect [flags]
```

### Options

```
      --database string   database name
  -h, --help              help for connect
      --host string       database host (default "localhost")
      --name string       connection name (required)
      --pass string       database password
      --port int          database port (default 5432)
      --sslmode string    SSL mode (default "disable")
      --type string       database type (sqlite, postgresql, mysql) (default "sqlite")
      --user string       database username
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos db](naeos_db.md)	 - Database connection and migration management

