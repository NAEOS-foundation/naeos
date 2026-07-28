## naeos kernel

Inspect the NAEOS kernel and service registry

### Synopsis

Inspect and interact with the NAEOS kernel services, metrics, and event bus.

Example:
  naeos kernel services --config config.yaml
  naeos kernel metrics --config config.yaml --output json
  naeos kernel publish --topic my-topic --payload hello
  naeos kernel subscribe --topic my-topic --payload hello

### Options

```
      --config string    path to JSON or YAML config file (auto-detected if omitted)
  -h, --help             help for kernel
      --output string    output format: text, json, or yaml (default "text")
      --payload string   event payload to publish
      --topic string     kernel event topic
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos kernel events](naeos_kernel_events.md)	 - List active kernel event topics
* [naeos kernel metrics](naeos_kernel_metrics.md)	 - Show kernel telemetry metrics
* [naeos kernel publish](naeos_kernel_publish.md)	 - Publish an event to the kernel event bus
* [naeos kernel services](naeos_kernel_services.md)	 - List registered kernel services
* [naeos kernel subscribe](naeos_kernel_subscribe.md)	 - Subscribe to a kernel event topic and optionally publish a sample payload

