# NAEOS v1.1.0 Development Plan

**20 tasks across 6 phases** — post-v1.0.0 quality, testing, and cleanup

---

## Phase 1: Critical Fixes (4 tasks)

### Task 1: Fix WebSocket Race Conditions
**Files:** `internal/websocket/server.go`, `internal/websocket/server_test.go`

Three distinct race conditions found by `-race`:

**Race A — Map mutation under RLock (server.go:93-102)**
- The `broadcast` case in `Run()` acquires `s.mu.RLock()` but then calls `delete(s.clients, client)` and `close(client.send)` — both are writes
- Fix: Acquire `RLock` to read+send, collect full clients into a slice, release `RLock`, then `Lock` to delete+close

**Race B — readPump vs writePump concurrent conn.WriteMessage (server.go:185 vs 206)**
- `readPump` writes pong reply (`c.conn.WriteMessage` line 185) while `writePump` writes broadcast messages (line 206)
- gorilla/websocket `Conn` is NOT safe for concurrent writes
- Fix: Add `writeMu sync.Mutex` to `Client` struct, wrap all `conn.WriteMessage` calls

**Race C — Stop() vs writePump concurrent conn.WriteMessage (server.go:150 vs 206)**
- `Stop()` sends close frame directly while `writePump` is still running
- Fix: Same `writeMu` mutex, or refactor `Stop()` to signal shutdown via channel instead of direct write

**Changes:**
1. Add `writeMu sync.Mutex` field to `Client` struct
2. Create helper method `(*Client).writeMessage(msgType int, data []byte) error` that locks `writeMu`
3. Replace all bare `c.conn.WriteMessage(...)` / `client.conn.WriteMessage(...)` with `c.writeMessage(...)`
4. Fix broadcast case: `RLock` → read+send, collect full clients, `RUnlock`, `Lock` → delete+close
5. Add test: `TestRaceConditions` that runs multiple goroutines broadcasting simultaneously

### Task 2: Sync OpenAPI Spec with Actual API
**File:** `docs/openapi.yaml`

The OpenAPI spec has **19 endpoints with major drift** out of 27 total. The spec describes an idealized reference-based API while the server accepts raw spec text inline.

**Strategy:** Update OpenAPI to match the actual server implementation (server.go is the source of truth).

Key changes needed:

| Endpoint | OpenAPI Says | Server Actually Does | Fix |
|----------|-------------|---------------------|-----|
| POST `/specs` | `name`+`content` object | `spec` string (raw text) | Rewrite request schema |
| POST `/specs/validate` | `spec_id` | `spec` string | Rewrite request schema |
| POST `/specs/compile` | `spec_id`+`target` enum | `spec`+`target` free-form | Rewrite request schema |
| POST `/pipeline/run` | `spec_id`+`pipeline`+`params` | `spec`+`target` | Rewrite request schema |
| POST `/context/generate` | `spec_id`+`context_type` | `spec`+`format` | Rewrite request schema |
| POST `/artifacts` | multipart `file`+`kind` | JSON `path`+`content`+`kind` | Rewrite content type + schema |
| POST `/cloud/plan` | `spec_id`+`resource_constraints` | `provider`+`project`+`region`+`resources` | Rewrite request schema |
| POST `/cloud/deploy` | `plan_id`+`environment`+`dry_run` | `provider`+`project`+`region`+`resources` | Rewrite request + 202→200 |
| POST `/cloud/destroy` | `deployment_id`+`force` | `provider`+`project`+`region`+`resources` | Rewrite request + 202→200 |
| POST `/plugins/execute` | `plugin`+`command`+`args` | `name`+`action`+`params` | Rename fields |
| GET `/healthz` | `{healthy: bool}` at `/api/v1/healthz` | `{status, timestamp}` at `/healthz` | Fix path + schema |
| GET `/readyz` | `{ready: bool}` at `/api/v1/readyz` | `{status}` at `/readyz` | Fix path + schema |
| GET `/metrics` | at `/api/v1/metrics` | at `/metrics` | Fix path |
| GET `/specs` | `offset` param, `specs` array | `page` param, `count` int | Fix query param + response |
| GET `/artifacts` | `spec_id`/`kind` filters, `total` | `page`/`limit` pagination, `count` | Fix query params + response |
| GET `/plugins` | `total` field | `count` + `page`/`limit` | Fix response schema |
| GET `/pipelines` | `runs` array | `pipelines` array + `page`/`limit` | Fix field name + add pagination |
| GET `/cloud/status` | single object, `deployment_id` query | paginated list | Rewrite response schema |
| GET `/version` | `go_version`, `git_commit` | `go` (no `git_commit`) | Fix field names |

**Add missing endpoint:** `/.well-known/openid-configuration`

**Add missing error response schemas:** 401, 403, 413, 429, 503

**Fix response envelope:** All responses wrapped in `{success: true, data: ...}` — add this to all response schemas.

### Task 3: Fix CHANGELOG Duplicate Section
**File:** `CHANGELOG.md`

Two `## [0.2.0]` sections at lines 368 and 392. Merge or remove the duplicate.

### Task 4: Fix README Go Version Badge
**File:** `README.md`

Badge says `go-1.22+` but project requires Go 1.25. Update badge.

---

## Phase 2: Test Coverage — Public API & Infrastructure (4 tasks)

### Task 5: Test `pkg/plugin` (0% → ~80%)
**Files:** `pkg/plugin/plugin_test.go` (new)

The package is 100% deprecated wrapper with 0 imports. Test the `Execute()` convenience function and type aliases.

Tests:
- Test type aliases resolve correctly (compile-time check via assignment)
- Test `Execute()` delegates to manager
- Test `NewSandbox()` returns correct type
- Test `NewManager()` returns correct type

### Task 6: Test Generation Adapters — Java/Python/Rust/TypeScript (0% → ~70%)
**Files:** `internal/generation/adapters/java_test.go`, `python_test.go`, `rust_test.go`, `typescript_test.go` (new)

All 4 language adapters are 100% untested. Follow the pattern from `go_test.go` and `actixweb_test.go`.

Tests per adapter:
- `Test{Language}GenerateProject` — verify output contains expected file structure
- `Test{Language}GenerateModule` — verify module code generation
- `Test{Language}GenerateService` — verify service code generation
- `Test{Language}GenerateDockerfile` — verify Dockerfile content
- `Test{Language}GenerateCI` — verify CI config content
- `Test{Language}GenerateDockerCompose` — verify compose file

### Task 7: Test Cloud Destroy Methods (0% → ~60%)
**File:** `internal/cloud/terraform_test.go`

AWS/GCP/Azure `Destroy()` methods are 0% tested. They just call `t.Run("terraform", "destroy", ...)`.

Tests:
- `TestAWSDestroy` — verify TerraformRunner.Run called with correct args
- `TestGCPDestroy` — same pattern
- `TestAzureDestroy` — same pattern
- Use existing `mockRunner` pattern from `TestTerraformPlan`

### Task 8: Add Integration Test Build Tags
**Files:** `internal/integration/e2e_test.go`, `internal/integration/pipeline_test.go`, `pkg/pipeline/pipeline_e2e_test.go`, `internal/websocket/websocket_integration_test.go`

These integration tests run in every CI cycle. Add build tags so they only run when explicitly requested.

Changes:
- Add `//go:build integration` to each file
- Create `Makefile` target or script: `go test -tags integration ./...`
- Update CI workflow to optionally run integration tests on schedule or manual trigger

---

## Phase 3: Test Coverage — CLI & Internal (4 tasks)

### Task 9: Test CLI Commands (37% → ~60%)
**Files:** `cmd/naeos/` test files

Focus on the 0% functions:
- `TestNewCompileCommand` — verify command setup
- `TestRunDeploy` — verify deploy workflow
- `TestRunDistributed` — verify distributed command
- `TestCheckSpec` — verify spec validation
- `TestFormatOutput` / `TestFormatTable` — verify output formatting
- `TestRenderHealthReport` — verify health report rendering
- `TestNewCloudDeployCommand` / `TestNewCloudPlanCommand` / `TestNewCloudStatusCommand` / `TestNewCloudExportCommand` / `TestNewCloudTypesCommand`

Pattern: Use `bytes.Buffer` for stdout/stderr, create temp files for input, verify output strings.

### Task 10: Test Real Connectors (mock-only)
**Files:** `internal/broker/nats_real_test.go`, `internal/database/postgres_real_test.go` (new)

These are 0% because they need real NATS/PostgreSQL. Add tests that:
- Test connection failure handling (invalid host)
- Test `Name()` returns correct string
- Use build tags: `//go:build integration`
- Test `IsNotFoundError` / `IsConnectionError` sentinel errors

### Task 11: Test Sandbox (41.5% → ~75%)
**Files:** `internal/pluginsdk/sandbox/sandbox_test.go`

Focus on untested paths:
- Test `Execute` with various timeout values
- Test `Execute` with invalid command
- Test `ValidatePath` edge cases
- Test rate limiting behavior
- Test concurrent execution

### Task 12: Test Marketplace & Diff (37.6% / 53.2% → ~70%)
**Files:** `internal/marketplace/marketplace_test.go`, `internal/diff/diff_test.go`

Marketplace:
- Test `FetchPlugin` with mock HTTP server
- Test `SearchPlugins` with mock HTTP server
- Test error handling (404, timeout, invalid JSON)

Diff:
- Test `DiffNEIR` with various NEIR structures
- Test `DiffModules` / `DiffServices` edge cases
- Test empty/nil inputs

---

## Phase 4: Code Quality (4 tasks)

### Task 13: Migrate `interface{}` → `any`
**Files:** 27 non-test files + 12 test files (247 occurrences total)

Use `sed -i 's/interface{}/any/g'` across all `.go` files. Go 1.18+ supports `any` as alias for `interface{}`.

Priority files (by count):
1. `internal/api/server.go` (26)
2. `internal/database/database.go` (21)
3. `internal/graphql/engine.go` (19)
4. `internal/configreload/config.go` (18)
5. `internal/specification/parser/types.go` (12)
6. `internal/cloud/gcp.go` (12)
7. `internal/observability/observability.go` (11)
8. `internal/cloud/azure.go` (11)
9. `internal/cloud/aws.go` (11)
10. `internal/database/postgres_real.go` (9)

After migration: `go build ./...` and `go vet ./...` to verify.

### Task 14: Fix `fmt.Errorf` Without `%w` (selective)
**Files:** `internal/broker/broker.go`, `internal/database/database.go`, `internal/database/postgres_real.go`, `internal/broker/nats_real.go`

Focus on the ~30 instances in `internal/` where an underlying error exists but isn't wrapped:
- `internal/broker/broker.go`: 9× `"not connected"` → wrap with `%w` where an error is available
- `internal/database/database.go`: 16× `"not connected"` → same
- `internal/database/postgres_real.go`: 7× → same
- `internal/broker/nats_real.go`: 4× → same

Leave CLI validation errors (cmd/naeos/) as-is — they are intentional sentinel errors.

### Task 15: Fix `return nil, nil` Ambiguity
**Files:** `internal/eventsourcing/eventsourcing.go`, `internal/marketplace/remote.go`, `internal/rollback/rollback.go`, `internal/generation/adapters/adapter.go`, `internal/specification/parser/parser.go`

9 instances across 5 files. For each:
- If "not found" semantics: return `nil, ErrNotFound` (create sentinel if needed)
- If "no-op" semantics: keep as-is but add comment explaining intent
- If "empty input": return explicit error

### Task 16: Add Missing Godoc Comments
**Files:** All `internal/` packages

822 exported symbols missing doc comments. Key packages:
- `internal/ai/` — `AIService`, `NewService`, `Suggestion`, etc.
- `internal/api/` — `Server`, `NewServer`, `AuthConfig`, etc.
- `internal/artifacts/` — `ArtifactKind`, `Artifact`, `Store`, etc.
- `internal/audit/` — `AuditEvent`, `Auditor`, etc.
- `internal/auth/` — `User`, `OAuth2Provider`, etc.
- `internal/cloud/` — `CloudProvider`, `DeployConfig`, etc.
- `internal/specification/parser/` — `IncludeResolver`, `FuncRegistry`, etc.

Pattern: `// FunctionName does X.` (Go convention, starts with symbol name).

---

## Phase 5: Architecture Improvements (2 tasks)

### Task 17: Remove Deprecated Packages
**Files:** `internal/pluginsdk/sdk.go` (delete), `pkg/plugin/plugin.go` (delete)

Both packages have **zero external consumers**. Safe to remove.

Changes:
1. Delete `internal/pluginsdk/sdk.go`
2. Delete `pkg/plugin/plugin.go`
3. Check if `internal/pluginsdk/` directory can be removed (keep `sandbox/` and `wasm/` — they are NOT deprecated)
4. Verify: `go build ./...` and `go vet ./...`

### Task 18: Add `t.Parallel()` to Tests
**Files:** All `*_test.go` files

Currently 0% usage of `t.Parallel()`. Adding it speeds up CI significantly.

Strategy:
- Add `t.Parallel()` at the start of each test function
- For tests that use `t.Setenv()`, `t.TempDir()`, or shared state: skip parallel (or use test-specific fixtures)
- For table-driven tests: add `t.Run` with parallel subtests
- Priority: packages with longest test times (`internal/testrunner` 12.6s, `internal/watch` 2.5s)

---

## Phase 6: Documentation & Polish (2 tasks)

### Task 19: Update CHANGELOG for v1.1.0
**Files:** `CHANGELOG.md`, `VERSION`

- Add `## [1.1.0]` section with all changes from Tasks 1-18
- Update `VERSION` to `1.1.0`
- Follow Keep a Changelog format

### Task 20: Final Verification
- `go build ./...` — clean
- `go vet ./...` — clean
- `go test -race ./...` — no race conditions
- `go test ./...` — all pass
- `go test -coverprofile=cover.out ./... && go tool cover -func=cover.out | tail -1` — verify coverage improved
- `golangci-lint run` — clean (if available)

---

## Execution Order

```
Phase 1 (Critical):    Task 1 → Task 2 → Task 3 → Task 4
Phase 2 (Test API):    Task 5 → Task 6 → Task 7 → Task 8
Phase 3 (Test CLI):    Task 9 → Task 10 → Task 11 → Task 12
Phase 4 (Quality):     Task 13 → Task 14 → Task 15 → Task 16
Phase 5 (Architecture): Task 17 → Task 18
Phase 6 (Docs):        Task 19 → Task 20
```

Tasks within each phase can be parallelized where no file dependencies exist.

## Expected Impact

| Metric | v1.0.0 | v1.1.0 Target |
|--------|--------|---------------|
| Test Coverage | 58.4% | ~75% |
| Race Conditions | 1 (WebSocket) | 0 |
| OpenAPI Drift | 19 major | 0 |
| `interface{}` | 247 | 0 |
| Deprecated Code | 2 packages | 0 |
| Missing Godoc | 822 symbols | ~0 |
| Integration Tests | No build tags | Build-tag gated |
