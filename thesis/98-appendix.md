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

## A.6 Remaining Protocol Steps (Future Runs)

1. **Warm-cache run:** third execution with persistent cache directory (`--cache-dir .naeos/cache`); capture per-stage hit rates via `--profile`.
2. **Mutation granularity:** change one field in `spec.yaml`; re-run; verify only downstream artifacts' hashes change.
3. **Three-adapter micro-benchmark:** isolate generation stage timing to replicate the whitepaper's ≈1.4 ms vs ≈3 ms parallel claim directly.
