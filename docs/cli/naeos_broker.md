## naeos broker

Message broker management

### Synopsis

Manage message broker connections (Redis, NATS, Memory, etc.).

Example:
  naeos broker connect --type redis --name myredis --host localhost --port 6379
  naeos broker connect --type nats --name mynats --host localhost --port 4222
  naeos broker connect --type memory --name mymem
  naeos broker list
  naeos broker publish --name myredis --channel events --message '{"event":"created"}'
  naeos broker subscribe --name myredis --channel events
  naeos broker disconnect --name myredis

```
naeos broker [flags]
```

### Options

```
  -h, --help   help for broker
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos broker connect](naeos_broker_connect.md)	 - Connect to a message broker
* [naeos broker disconnect](naeos_broker_disconnect.md)	 - Disconnect from a broker
* [naeos broker list](naeos_broker_list.md)	 - List all broker connections
* [naeos broker publish](naeos_broker_publish.md)	 - Publish a message to a channel

