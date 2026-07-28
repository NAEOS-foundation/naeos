## naeos gateway add-backend

Add a backend to load balancer

```
naeos gateway add-backend [flags]
```

### Options

```
  -h, --help          help for add-backend
      --lb string     load balancer name (default "default")
      --name string   backend name (required)
      --url string    backend URL (required)
      --weight int    backend weight (default 1)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos gateway](naeos_gateway.md)	 - API gateway management

