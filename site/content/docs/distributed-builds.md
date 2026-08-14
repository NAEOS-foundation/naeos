---
title: Distributed Builds
description: Parallelize pipeline execution across multiple workers for faster builds.
weight: 13
---

NAEOS supports distributed pipeline execution: the pipeline stages are
scheduled as independent tasks and processed in parallel across configurable
worker pools.

## When to Use It

Distributed mode helps when:

- Specifications contain many modules or services that can be processed
  independently
- You want to reduce end-to-end build time on large specs
- You are benchmarking pipeline stage throughput

For small specs the in-process pipeline is faster — distribution overhead
only pays off once there is enough work to parallelize.

## Usage

```bash
# Run the pipeline with 4 workers (default)
naeos distributed --config naeos.yaml

# Use more workers for larger specs
naeos distributed --config naeos.yaml --workers 8
```

| Flag | Description | Default |
|------|-------------|---------|
| `--input` | Path to the specification file | — |
| `--config` | Path to a NAEOS config file | — |
| `-w, --workers` | Number of parallel workers | `4` |

## How It Works

1. The pipeline configuration is loaded (same config resolution as
   `naeos run`: `naeos.yaml`, `naeos.yml`, `naeos.json`, or
   `.naeos/config.yaml`).
2. Each pipeline stage (`parse`, `normalize`, `resolve`, `build-neir`,
   `validate`, `schedule`, `generate`, `review`) becomes a task.
3. Tasks are submitted to a coordinator that load-balances them across the
   worker pool using round-robin distribution.
4. Results are aggregated; failed tasks are reported per worker.

## Output

The command prints a summary and any failed tasks:

```text
Distributed pipeline: 4 workers, 8 tasks
Results: completed: 8, failed: 0
Pipeline: my-project
```

## Related

- [Pipeline Engine](/docs/pipeline-engine/) — the 11-stage DAG pipeline
- [Dashboard](/docs/dashboard/) — monitor pipeline activity in real time
