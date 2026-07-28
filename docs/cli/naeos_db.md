## naeos db

Database connection and migration management

### Synopsis

Manage database connections, run migrations, and inspect schemas.

Example:
  naeos db connect --type sqlite --name mydb
  naeos db list
  naeos db migrate --name mydb
  naeos db disconnect --name mydb

```
naeos db [flags]
```

### Options

```
  -h, --help   help for db
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos db connect](naeos_db_connect.md)	 - Connect to a database
* [naeos db disconnect](naeos_db_disconnect.md)	 - Disconnect from a database
* [naeos db list](naeos_db_list.md)	 - List all database connections
* [naeos db migrate](naeos_db_migrate.md)	 - Run database migrations
* [naeos db status](naeos_db_status.md)	 - Show database connection status

