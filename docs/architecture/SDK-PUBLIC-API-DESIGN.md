# NAEOS Public Go SDK Boundary Design

Status: Proposal
Date: 2026-10-25
Scope: Existing signatures remain unchanged. The initial additive `pkg/sdk`
façade is implemented; this document records its boundary and remaining design
work.

## Purpose

Define a public Go integration surface that follows [NES-019](../NES-019-SDK.md): consumers import packages under `pkg/` and do not need types from `internal/`. Keep the existing runtime implementation behind adapters, and avoid turning implementation packages into permanent SDK contracts.

## Findings from the current tree

NES-019 is marked Draft v0.3 but says public APIs are stable and must not require knowledge of `internal/` (FR-001, NFR-001, NFR-002). The Go module path is `github.com/NAEOS-foundation/naeos`; repository tags include v3.x. Treat the published compatibility expectation as unresolved until maintainers confirm which tags/releases were distributed and what compatibility policy they promised.

| Public package/API | Internal types in public surface | Effect |
|---|---|---|
| `pkg/config.Pipeline.Policies` | `[]internal/governance/policy.Rule` | External consumers cannot construct policy values from the internal package. |
| `pkg/pipeline.Config` | Parser, normalizer, resolver, builder, validator, scheduler, generator, renderer, graph, registry, policy, review, profiling and other internal types | Configuration is coupled to implementation packages and cannot serve as a self-contained external extension contract. |
| `pkg/pipeline.Result` | NEIR model, artifacts, tasks, planner graph, review results and policy results from `internal/` | Results expose implementation representation and constrain refactors. |
| `pkg/pipeline.Pipeline` methods | Profiling types, registry, graph and renderer from `internal/` | Public methods return implementation objects. |
| `pkg/pipeline.PipelineAdapter` | Middleware and run snapshot from internal middleware/event-sourcing packages | Adapter extension and event inspection require internal packages. |
| `pkg/kernel.Kernel` | `internal/events` is used internally but does not appear in its current public signatures | Implementation dependency; no direct signature leak identified in the inspected methods. |

Useful source references: `pkg/config/config.go`; `pkg/pipeline/pipeline.go`; `pkg/pipeline/adapter.go`; `pkg/pipeline/run_policy.go`; `pkg/kernel/kernel.go`.

In-repository consumers also rely on these concrete types. `internal/pipelinecache` stores `*pipeline.Result`; CLI commands inspect result fields; observers in `internal/websocket` and `internal/monitoring` implement `pipeline.PipelineObserver`; the policy-bypass experiment constructs internal policy rules. These consumers are migration inputs, not evidence that external users can legally or practically depend on internal packages.

## Design decisions

1. Keep the current Go module and implementation packages in place. Do not split modules or move domain code as part of the API boundary work.
2. Introduce an additive public façade under `pkg/sdk` with public configuration, policy, result, artifact, and execution-summary types. The initial implementation provides configuration, policy, result, artifact, task/review/policy summaries, governance provenance, and graph counts. A public observer contract remains deferred until its supported semantics are designed.
3. Keep the first façade deliberately narrow. Expose scalar pipeline settings, policy DTOs, Run/Validate operations, public summaries, and artifacts. Do not expose constructors for internal parser/generator/graph objects until a corresponding public extension contract is designed. The pipeline checks cancellation at stage boundaries, but cannot interrupt a stage already running.
4. Convert between public DTOs and internal implementation types at the façade boundary. Keep conversions in one adapter layer; do not alias internal types into `pkg/` because aliases preserve the coupling.
5. Make public data types own their serialization and compatibility rules. The initial versioned DTO is `pkg/neir/v1`; it types project, module, service, and metadata data, keeps other sections opaque as `json.RawMessage`, and preserves unknown fields. Conversion stays in `internal/neir/publicadapter`, outside the public API.
6. Retain existing `pkg/config` and `pkg/pipeline` behavior during the additive migration. Mark them legacy/implementation-oriented in docs only after maintainers choose a release policy; do not change their signatures silently.
7. Do not promise a deprecation/removal schedule until maintainers confirm which v3 tags were released and whether the module path's major-version convention is intentional. A signature break may require a new major module path under Go module rules.

## Compatibility policy recommendation

The repository has an actual v3 release signal: the local tag set includes
`v3.7.0`, and the [GitHub release page](https://github.com/NAEOS-foundation/naeos/releases)
lists v3.7.0 as a stable release dated 2026-10-06. The repository's `CHANGELOG.md`
and `VERSION` identify the checkout as 3.7.1, but its local tag set stops at
`v3.7.0`; the GitHub release list inspected here also showed v3.7.0 as latest.
Release engineering should reconcile the source version, changelog, Git tags,
and GitHub releases before the next publication. In particular, decide whether
the additive SDK API belongs in the 3.7.1 patch release or requires a compatible
minor version. Until the release state is reconciled, assume the published v3
API has external consumers.

The module directive is `github.com/NAEOS-foundation/naeos` without a major
suffix, despite the v3 tags. The Go module documentation requires a `/vN`
suffix for module paths at v2 and above ([Go module versioning](https://go.dev/ref/mod#major-version-suffixes)).
Do not change the current import path in place: that would break consumers of
the released path. Preserve it for the current line and resolve the migration
with release engineering; use the proper `/v4` path for any future v4 module.

Use this conservative add-only policy for the current release line:

- Treat existing exported `pkg/` signatures and documented behavior as compatibility commitments. Do not remove or change them in the current v3 line; continue the façade additively.
- Keep implementation packages under `internal/` outside the compatibility promise. Do not expand their types into the new SDK.
- Mark `pkg/sdk` as the preferred integration API, while describing it as an initial release until its public data schema, plugin story, and compatibility policy are complete. After the first release containing it, make only compatible additions and bug fixes on the current line.
- Put any necessary breaking SDK redesign in a new major module path (`/v4`) with migration notes. Coordinate the path correction, package import updates, release artifacts, and consumer migration as one explicit major-version plan.
- Treat the published v3.7.0 release as evidence of external compatibility expectations; do not use absence of known consumers as permission to break compatibility.

This recommendation preserves compatibility with the verified v3.7.0 release. It does not establish which support terms or additional public commitments apply; release policy and distribution scope still require maintainer confirmation.

## Typed NEIR boundary

The mapping is now explicit: NEIR model metadata follows the stable NES-023 version (`2.0.0`), while `SchemaVersion` records the current input-spec parser version (`0.5.0`). The registry artifact at `site/public/schemaregistry/v1/neir.json` uses JSON Schema draft-07 and is generated from `internal/neir/model.NEIR`; its `v1` path versions the registry publication interface, not either metadata value. `internal/neir/version.Default()` is the single source used by both eager and lazy builders, and its default is covered by validation tests.

A builder regression test serializes generated NEIR and validates it against the checked-in model v1 schema with a full draft-07 validator. Schema loading uses only embedded resources; unresolved external references fail closed without network or file access. The generator now publishes a separate input-spec v1 schema at `site/public/schemaregistry/spec/v1/spec.json`; it describes source-only forms such as string project names, deployment environment names, and coverage strings. `ValidateSpec` and the CLI use the input schema, while the builder test uses the model schema. Repository example specs pass the input schema checks.

The public `pkg/neir/v1` DTO and internal adapter are implemented. Tests verify that a builder-shaped internal model converts to the DTO, preserves JSON semantics through decode/encode, and validates against the checked-in model schema. Complex top-level sections remain opaque raw JSON in v1, with unknown fields preserved. `Pipeline.ValidateContext` now materializes module and service slices before computing the result hash, matching the `RunContext` output. The SDK exposes that model through `Result.NEIR` using the public DTO, with focused coverage for both operations.

This corrects the previous metadata values (`0.1.0` and `1.0`); the latter was invalid under the three-component SemVer parser. Generated NEIR metadata therefore changes for consumers who inspect serialized output. Existing serialized files are not rewritten. Release notes should call out the metadata correction before distributing a release containing it.

Recommended boundary:

1. Keep the separate authorities and field meanings for NEIR model version, input-spec version, and each schema artifact version synchronized when any of them changes.
2. Keep the builder/model-schema regression test, input-spec example checks, and DTO round-trip tests. Expand typed complex sections only with schema-backed compatibility tests.
3. Publish a migration note for the metadata value correction before release.
4. Maintain the public `pkg/neir/v1` contract against the canonical model schema; its opaque sections and unknown-field behavior are explicit compatibility commitments.
5. Keep materialization and hashing in `Run` and `Validate` aligned as the internal NEIR lifecycle evolves.
6. Keep `sdk.Result.NEIR` on the versioned public DTO; do not return raw JSON of the internal representation as though it were a stable public contract.

## Plugin SDK boundary

`docs/plugin-sdk.md` tells plugin authors to import `internal/pluginhost`, and `internal/pluginhost.Plugin` uses `*PluginContext` plus `any` request/response values. The runtime, sandbox, and scaffold implementation also live under `internal/pluginsdk/`. This is not an external Go SDK contract: a separate module cannot import those paths.

Keep the plugin host/manager internal. The proposed versioned WASM request/response protocol is the intended portable external boundary, but is not implemented or a compatibility promise yet. [`PLUGIN-API-DESIGN.md`](PLUGIN-API-DESIGN.md) records the current host behavior, proposed envelope, version negotiation, authorization, error, and timeout requirements. Add a public author package (for example `pkg/pluginapi`) only for shared manifest and wire DTOs; a Go helper should serialize those types rather than expose the host manager interface. Use context-aware calls and capability-scoped host services rather than exposing raw host directories or internal event bus objects. Keep native Go `.so` loading as an implementation-specific, same-process option, not the portable compatibility promise for external plugins.

Before publishing that contract, implement and test the requirements in [`PLUGIN-API-DESIGN.md`](PLUGIN-API-DESIGN.md), then resolve the open maintainer decisions listed there. The public plugin API should not mirror `internal/pluginhost.Plugin` until there is a demonstrated external use case for in-process Go interfaces.

## Proposed first public contract

Illustrative shape only; names and exact fields remain subject to API review:

```go
package sdk

type Config struct {
    Name              string
    Mode              string
    Verbose           bool
    DryRun            bool
    OutputDir         string
    Languages         []string
    RequireGovernance bool
    Policies          []PolicyRule
}

type PolicyRule struct {
    ID        string
    Condition string
    Priority  int
    Action    string
    Scope     string
    Enabled   bool
}

type Result struct {
    RunID             string
    SpecificationHash string
    NEIRHash          string
    Artifacts         []Artifact
    Tasks             []TaskSummary
    Reviews           []ReviewSummary
    PolicyResults     []PolicyEvaluation
    Governance        GovernanceSummary
    Graph             GraphSummary
}

type Artifact struct { Path string; Content []byte }
type TaskSummary struct { ID, Name string; Dependencies []string; Priority int }
type ReviewSummary struct { Status, RuleID, Message string }
type PolicyEvaluation struct { Passed bool; RuleID, Message, Action string; Priority int }
type GovernanceSummary struct { Mode, Status string; EffectivePolicyCount int; PolicyContextVersion, PolicyContextDigest string }
type GraphSummary struct { Nodes, Edges int }

func New(Config) (*Client, error)
func (*Client) Run(context.Context, string) (*Result, error)
func (*Client) Validate(context.Context, string) (*Result, error)
```

The sketch intentionally omits implementation injection, arbitrary internal graph access, and unstable profiler objects. Any omitted capability needed by real consumers must be added as a designed public interface before the façade is declared stable.

The current implementation treats `Mode: "governed"` as requiring at least
one enabled policy, matching the fail-closed behavior of file-loaded pipeline
configuration. A caller that needs a different governance-mode contract must
resolve that behavior explicitly before the SDK is declared stable.

## Migration sequence

### Phase A — Establish compatibility policy

- Confirm release/tag distribution and any existing SDK consumers.
- Decide whether the additive façade becomes the stable SDK now, and how existing `pkg/config` and `pkg/pipeline` are described during transition.
- Resolve whether the existing v3 tags and unsuffixed module path represent a compatibility promise or internal release labels.

### Phase B — Complete the public contract

- Review and extend DTOs and documentation under `pkg/sdk`.
- Implement conversions to/from internal policy and pipeline types in a narrow adapter.
- Ensure public signatures import only standard library, other approved public `pkg/` packages, and external dependencies that are intentionally part of the contract.
- Keep configuration loading in `pkg/config`; add conversion to the new DTO rather than exposing internal policy values.

### Phase C — Migrate internal consumers

- Update CLI and `internal/pipelinecache` to consume the façade or a private internal result model, whichever preserves cache compatibility and avoids unnecessary DTO round trips.
- Keep observers internal where they are used only inside NAEOS; add a distinct public observer only if supported as an SDK capability.
- Update experiments and examples to use the public policy DTO and API.
- Keep plugin support as a separate boundary decision. `docs/plugin-sdk.md` now labels the current interface and JSON-over-stdio wire shape experimental/in-repository, explains the Go `internal` import constraint, and avoids promising a stable external plugin API. A public plugin contract still requires a separate design.

### Phase D — Freeze and maintain

- Publish an external-consumer example using only supported public imports and verify it from outside the NAEOS module.
- Document API stability, serialization/versioning, errors, context cancellation, and deprecation rules.
- Add API compatibility checks and a public-package dependency rule to CI after the contract is accepted.
- Remove or change legacy signatures only under the confirmed compatibility/versioning policy.

## Acceptance criteria before calling the SDK stable

- An external module can construct configuration, define policies, run and validate a specification, and consume results without importing `internal/`.
- Every exported SDK field, parameter, return type, callback, and interface uses a public or standard-library type.
- Public serialization has explicit compatibility and schema-version rules.
- The plugin story is either supported through a public SDK or clearly documented as an in-repository/experimental API. The current documentation meets the latter condition; it does not make the plugin API externally stable.
- Documentation examples compile from outside the NAEOS module.
- Maintainers have written down release compatibility and deprecation policy.

## Initial implementation verification

- `go build ./pkg/sdk ./pkg/config ./pkg/pipeline` passed.
- `go test ./pkg/sdk ./pkg/config ./pkg/pipeline ./internal/pipelinecache` passed. Tests cover generated public results, enabled policies, fail-closed governed mode, pre-canceled contexts, cancellation at a stage boundary, and hashes from Validate.
- A temporary separate Go module importing only `pkg/sdk` passed integration tests for Run results, governance behavior, and canceled contexts. Its files were kept under `/tmp`, outside the repository.
- The pipeline checks cancellation before processing and at stage boundaries. An active stage is not interrupted; cancellation is observed when that stage returns. Cancellation during artifact writing may leave artifacts already written earlier in that stage.
- Both `Run` and `Validate` populate specification and NEIR hashes after successful completion.

## Open maintainer decisions

1. Is 3.7.1 intended to be a published release? `VERSION` and `CHANGELOG.md` say 3.7.1, while the local tags and GitHub release list inspected here stop at v3.7.0.
2. Is the unsuffixed module path intentional for the existing line, and who owns the `/v4` import-path migration?
3. Which capabilities must third-party SDK users configure directly: custom policy evaluators, generators, validators, middleware, cache, graph, profiler, or only supported options?
4. Who owns both schema contracts and maintenance of the public NEIR compatibility contract?
5. Should plugin support expose a Go interface, a WASM protocol, or both, and who maintains the public compatibility contract?

## Non-goals

- No changes to existing signatures, aliases, deprecations, module split, or dependency changes in this step. The new `pkg/sdk` surface is additive.
- No claim that the current API is legally or contractually stable; repository tags and source show release signals, but external distribution and user-facing promises require maintainer evidence.
