# Appendix A — Artifact Evaluation and Reproduction Protocol

This appendix discloses the exact environment, commands, and data underlying every quantitative claim in Chapter 6, following artifact-evaluation conventions used at international software engineering venues (ICSE/FSE/ASE badging criteria: *available, functional, reused*).

## A.1 Artifact Availability

| Component | Location | License |
|---|---|---|
| Platform source | `github.com/NAEOS-foundation/naeos` (v3.1.0) | Apache 2.0 |
| Committed benchmark baseline | `bench/baseline.txt` in repository | Apache 2.0 |
| Example specification under test | `examples/spec-full.yaml` | Apache 2.0 |
| Monograph manuscript | `thesis/` in this repository | — |

## A.2 Environment

- **OS/toolchain:** Linux/amd64, Go 1.25
- **Replication hardware:** AMD EPYC 7763 (benchmark suite); committed baseline recorded on Intel Core i5-2520M @ 2.50 GHz
- **Binary:** built from source with `go build -o naeos ./cmd/naeos/`

## A.3 Determinism Experiment (Section 6.2.1)

Two isolated working directories, each containing the same specification and configuration:

```bash
# per directory (det1/, det2/):
cp examples/spec-full.yaml spec.yaml
cp config.example.yaml config.yaml   # output_dir: ./out

naeos run --config config.yaml --input-file spec.yaml --language go
cd out && find . -type f | sort | xargs sha256sum > hashes.txt
```

Pairwise comparison:

```bash
diff det1/out-hashes.txt det2/out-hashes.txt && echo IDENTICAL
```

**Observed result:** 79 artifacts per run; all SHA-256 digests equal across runs; emitted set includes `cmd/app/main.go`, `internal/<module>/{config,domain,handler}.go`, unit tests, Dockerfile, docker-compose.yml, `.github/workflows/ci.yml`, README.md, docs/architecture.md.

## A.4 Benchmark Suite Replication (Section 6.3.1)

```bash
go test -run '^$' -bench 'BenchmarkPipeline' -benchmem -count=1 ./pkg/pipeline/
```

**Observed result (AMD EPYC 7763):**

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| PipelineRun (small) | 5,833,778 | 174,388 | 1,281 |
| PipelineRun_Medium | 36,368,645 | 3,698,572 | 17,810 |
| PipelineRun_Large | 421,457,330 | 81,457,752 | 166,399 |
| PipelineRun_ParallelSmall | 5,603,323 | 176,572 | 1,281 |
| PipelineRun_ParallelMedium | 39,944,789 | 3,718,468 | 17,814 |
| PipelineRun_ParallelLarge | 397,082,472 | 81,618,301 | 166,383 |
| PipelineValidate | 127,384 | 25,773 | 267 |
| PipelineNew | 7,586 | 2,464 | 18 |

Total suite wall time: 22.0 s.

## A.5 CLI End-to-End Benchmark (Section 6.3.1)

```bash
naeos benchmark -n 50 --output text
```

**Observed result:** 50 iterations, 0 errors; average 31.971 ms; min 24.024 ms; max 103.676 ms.

## A.6 Warm-Cache and Mutation-Granularity Experiments (Section 6.2.1, 6.3.4)

**Warm cache:**

```bash
naeos run --config config.yaml --input-file spec.yaml --language go --cache-dir .naeos/cache   # cold
time naeos run --config config.yaml --input-file spec.yaml --language go --cache-dir .naeos/cache  # warm
```

**Observed:** cold 151 ms → warm 56 ms end-to-end wall time (2.7×); warm output byte-identical (79/79 vs.\ cold run).

**Mutation granularity** (full re-hash after each single-field edit of `spec.yaml`):

| Mutation | Command | Result |
|---|---|---|
| module description text | `sed 's/description: Authentication and authorization module/.../'` | 0 artifacts changed (field not emitted) |
| service port 8080→9090 | `sed 's/port: 8080/port: 9090/'` | 0 changed — port not propagated by Go adapter (recorded limitation) |
| project name +`-v2` | `sed 's/project: e-commerce-platform/project: e-commerce-platform-v2/'` | 10 changed + 1 package rename; 68 untouched |

## A.7 Remaining Protocol Step

Per-stage hit-rate breakdown via `--profile`: the flag requires profiling instrumentation to be enabled at build/run configuration; capture stage-level hit data once available and extend Table in Section 6.2.1.
