# Go package dependency map

Use the repository's Go toolchain to generate a current package-import graph before proposing package moves or consolidations.

## Generate

From the repository root:

```bash
./scripts/go-package-graph.sh > /tmp/naeos-go-packages.dot
```

If Graphviz is installed, render an SVG for inspection:

```bash
dot -Tsvg /tmp/naeos-go-packages.dot -o /tmp/naeos-go-packages.svg
```

The script can also write directly to a supplied output path:

```bash
./scripts/go-package-graph.sh /tmp/naeos-go-packages.dot
```

## Semantics and limitations

- Each edge points from the importing package to the imported package.
- Only direct imports between packages inside the NAEOS Go module are emitted. Third-party and standard-library dependencies are omitted to keep the graph focused on internal coupling.
- The graph reflects packages discovered by `go list ./...`; test-only import edges are not represented.
- The script requires the Go toolchain and a successful `go list ./...`. Graphviz is optional and is not needed to generate DOT.
- Output is generated on demand and is intentionally not checked in: a committed snapshot could become stale.
- A graph is evidence for review, not an automatic refactoring recommendation. Inspect package responsibilities, interfaces, tests, and runtime behavior before changing boundaries.

## Recorded CI baseline

The following observation was emitted by the CI graph generator at commit
`6093df018a3fb727e6fa66fb4b8ad47542002a51` (2026-10-13). It is a dated
baseline, not a substitute for regenerating the graph at the commit being
reviewed.

- Go packages discovered: **164**
- Direct in-module import edges: **387**

Top fan-in packages (incoming internal imports):

| Package | Fan-in | Fan-out |
|---|---:|---:|
| `internal/errors` | 75 | 0 |
| `internal/neir/model` | 20 | 16 |
| `internal/evidence` | 16 | 2 |
| `internal/governance/control` | 13 | 2 |
| `internal/securityext` | 11 | 1 |
| `internal/specification/parser` | 11 | 2 |
| `internal/neir/model/language` | 10 | 0 |
| `internal/governance/policy` | 9 | 2 |
| `internal/controlplane` | 8 | 0 |

Top fan-out packages (outgoing internal imports):

| Package | Fan-in | Fan-out |
|---|---:|---:|
| `cmd/naeos` | 0 | 75 |
| `pkg/pipeline` | 6 | 25 |
| `internal/api` | 2 | 19 |
| `internal/neir/builder` | 1 | 17 |
| `internal/testutil/adaptertest` | 0 | 17 |
| `internal/neir/model` | 20 | 16 |

### Interpretation

- `internal/errors` is a widely reused leaf package in this graph. High fan-in alone is not a reason to split it.
- `cmd/naeos` is a composition/entry-point package with broad fan-out; treat it as an orchestration surface, not an automatic extraction target.
- `pkg/pipeline`, `internal/api`, and `internal/neir/model` merit targeted caller and responsibility reviews because they combine comparatively high internal coupling with distinct responsibilities.
- `internal/evidence`, `internal/governance/control`, and `internal/controlplane` are visible governance/evidence boundaries. Their fan-in values do not establish duplication or justify consolidation.
- The graph counts production imports exposed by `go list`; it excludes test-only imports and does not encode runtime call direction, dependency injection, or semantic ownership.

## Safe review sequence

1. Generate the graph at the commit under review and retain the commit SHA with any findings.
2. Identify high-fan-in packages and verify their callers with targeted searches.
3. Check whether any proposed dependency inversion or package move changes authorization, runtime enforcement, evidence, persistence, CLI, or public documentation contracts.
4. Make one behavior-preserving change per PR and run the relevant package tests plus repository CI.
