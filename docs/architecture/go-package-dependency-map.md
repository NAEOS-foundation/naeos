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

### Source-level follow-up from the baseline

The following source inspection used `main` at the time of review. These are
review targets, not approved package moves:

- **`pkg/pipeline`** — fan-in 6, fan-out 25. `pipeline.go` imports evidence,
  generation, governance policy/review, NEIR model/builder/validator, planner,
  profiling, registry/schema registry, specification parsing/resolution, and
  kernel/config packages. This is a cross-domain orchestration surface. Before
  splitting it, identify cohesive stages and existing public types; do not move
  code solely to reduce its import count.
- **`internal/api`** — fan-in 2, fan-out 19. `server.go` is a large server
  composition file and imports multiple service domains including audit,
  authentication, database, plugin hosting, schema registry, WebSocket, and
  pipeline. A future review should distinguish route/handler wiring from server
  lifecycle and shared dependencies. Any extraction needs API/server integration
  tests and must preserve route behavior.
- **`internal/neir/model`** — fan-in 20, fan-out 16. `model.go` aggregates
  types from multiple `internal/neir/model/*` subpackages. This may be an
  intentional facade, so retain it until callers and compatibility expectations
  are understood.
- **`internal/errors`** — fan-in 75, fan-out 0. It is a widely imported leaf;
  high fan-in alone is not evidence of a defect or a reason to split it.

**Current recommendation:** do not consolidate `internal/controlplane` with
`internal/governance/control`. The graph measures imports, not equivalent
authorization semantics. First map the decision types, policy-version checks,
runtime call sites, evidence records, persistence guarantees, and targeted tests.
The first source refactor should target a demonstrated responsibility problem
and be isolated in its own PR.

## Safe review sequence

1. Generate the graph at the commit under review and retain the commit SHA with any findings.
2. Identify high-fan-in packages and verify their callers with targeted searches.
3. Check whether any proposed dependency inversion or package move changes authorization, runtime enforcement, evidence, persistence, CLI, or public documentation contracts.
4. Make one behavior-preserving change per PR and run the relevant package tests plus repository CI.


### Caller and enforcement-boundary audit — 2026-10-15

This pass searched source call sites and inspected the implementation and
focused tests on `main` at base commit
`82215b2596bc9ca46c34a9c252106c44b0e84841`. It maps active usage; it does
not establish that every integration is production-deployed.

| Surface | Observed callers and contracts | Boundary conclusion |
|---|---|---|
| `internal/governance/control` | CLI runtime paths in `cmd/naeos/runtime_cmd.go`, `runtime_bridge_cmd.go`, `agent_request_cmd.go`, and `evidence_cmd.go` construct this control plane and pass it to `internal/runtime/gateway`. The control plane binds a grant digest, supports grant revocation/expiry, and implements `ValidateDecision`. | Canonical governance contract for the CLI runtime gateway path observed in this audit. |
| `internal/runtime/gateway` | `ExecutionGateway.Authorize` evaluates a request, applies restrictions, optionally calls `DecisionRevalidator.ValidateDecision`, and atomically claims an invocation ID before sandbox execution when replay protection is enabled. The concrete `control.ControlPlane` implements the revalidation interface. | Runtime side-effect boundary. Keep revalidation and replay-claim ordering intact during future changes. The interface permits non-revalidating implementations, so callers/configuration must be reviewed before claiming all deployments enforce second-phase validation. |
| `internal/evidence` | `EvidenceRecord` imports `governance/control.Decision`; completion validation checks the evidence chain and runtime ledger. `pkg/pipeline` calls `ValidateCompletionWithRuntimeLedger`. | Run-completion evidence contract tied to the governance decision model; distinct from the control-plane authorization ledger. |
| `internal/controlplane` | Used by `internal/integrations/sandbase` and `internal/investordemo`, plus control-plane examples. It defines its own `Policy`, `Grant`, `DecisionGateway`, `Ledger`, and `EvidenceBundle`. | Separate contract used by SandBase integration and the control-plane/investor-demo flow; not proven to be a duplicate of `governance/control`. |
| SandBase freshness tests | `internal/integrations/sandbase/http_handler_freshness_test.go` checks that a v1 grant against active policy v2 returns `reauthorize`. `internal/controlplane/policy_change_revocation_test.go` checks that stale authorization does not invoke the side effect and emits a blocked event. | Existing tests cover important stale-policy behavior in this contract, but they do not justify merging the two control-plane models. |

#### Concrete implications

1. Do not consolidate the two packages as a repository-cleanup task. They serve
   different call paths and have different decision and evidence types.
2. Keep runtime gateway authorization revalidation immediately before replay
   claiming and side-effect execution; changing that sequence requires targeted
   concurrency and stale-policy regression tests.
3. Keep SandBase policy refresh/version checks and the adapter's active-policy
   lookup in the integration path. Audit the complete grant lifecycle separately
   before changing its policy/grant model.
4. Keep `internal/evidence` completion/provenance validation distinct from the
   `internal/controlplane.Ledger` until a contract-level comparison proves
   which records are authoritative for each flow.
5. This was source inspection only. It does not claim local tests ran, nor that
   all deployment configurations enable the same enforcement options.
