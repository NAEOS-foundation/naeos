## naeos broker connect

Connect to a message broker

```
naeos broker connect [flags]
```

### Options

```
      --db int            Redis database number
  -h, --help              help for connect
      --host string       broker host (default "localhost")
      --name string       connection name (required)
      --password string   broker password
      --port int          broker port (default 6379)
      --type string       broker type (redis, nats, memory) (default "redis")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos broker](naeos_broker.md)	 - Message broker management

