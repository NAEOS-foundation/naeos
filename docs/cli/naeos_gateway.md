## naeos gateway

API gateway management

### Synopsis

Manage API gateway routing, rate limiting, circuit breakers, and load balancing.

Example:
  naeos gateway status
  naeos gateway rate-status
  naeos gateway cb-status --name api
  naeos gateway lb-list --name api
  naeos gateway add-backend --lb api --name backend1 --url http://localhost:8080

```
naeos gateway [flags]
```

### Options

```
  -h, --help   help for gateway
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos gateway add-backend](naeos_gateway_add-backend.md)	 - Add a backend to load balancer
* [naeos gateway cb-status](naeos_gateway_cb-status.md)	 - Show circuit breaker status
* [naeos gateway lb-list](naeos_gateway_lb-list.md)	 - List load balancer backends
* [naeos gateway rate-status](naeos_gateway_rate-status.md)	 - Show rate limiter usage
* [naeos gateway status](naeos_gateway_status.md)	 - Show gateway status

