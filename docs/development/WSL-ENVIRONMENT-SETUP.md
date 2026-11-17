# WSL Environment Setup

Audit date: 2026-11-11 (Asia/Bangkok)

This audit covers the related repositories under `/home/bayu/workspace`: `naeos`, `naeos-crm`, and `naeos-intelligence`. Each repository had pre-existing uncommitted changes before validation. No existing files were reset or cleaned. The only intended repository change from this task is this report.

## Status summary

- **Completed:** WSL/Linux and resource audit; native Go version confirmed compatible; Node/npm confirmed compatible with the CRM; Supabase CLI confirmed; Compose files parse; Go module tidy checks, builds, vet, and formatting checks passed; CRM lint passed.
- **Failed:** Both Go test suites timed out; CRM test command exited 1 (45 failures, including sandbox loopback permission errors); Docker daemon access failed and the Docker Desktop context caused the installed CLI to panic. `golangci-lint` timed out; CRM production build was interrupted after Vite remained at `transforming...` without progress.
- **Blocked:** Docker-backed local stacks and Supabase local database/migration checks cannot run until Docker access is repaired. `naeos/supabase/.temp` contains local project-link metadata; remote migration status was intentionally not queried because the link could point to the hosted production project.
- **Requires user action:** Restore Docker Desktop WSL integration or daemon/socket access; reauthenticate GitHub CLI if needed; run the bounded-out tests in an unrestricted local WSL shell; confirm and safely handle the existing Supabase link before any remote operation.

## Initial environment findings

| Area | Finding |
| --- | --- |
| WSL | WSL2 kernel confirmed by `uname`: `6.6.114.1-microsoft-standard-WSL2`. `wsl.exe --version` was unavailable inside this shell, so the Windows-side WSL package version was not established. |
| Distribution | Ubuntu 24.04.4 LTS (Noble). |
| Kernel/architecture | Linux `6.6.114.1-microsoft-standard-WSL2`, `x86_64` / Go `linux/amd64`. |
| Workspace | `/home/bayu/workspace`, containing `naeos`, `naeos-crm`, and `naeos-intelligence`; there is no Git repository at the workspace root. |
| Git | Git 2.43.0. Branches: `naeos` `fix/pr782-dco-signed` (22 changed/untracked paths); `naeos-crm` `fix/m1-runtime-stabilization` (10); `naeos-intelligence` `docs/m2-provider-neutral-authorization` (12). These changes predated validation and were preserved. |
| Go | Go 1.26.10, installed at `/usr/local/go/bin/go` (native Linux). Both `naeos/go.mod` and `naeos-intelligence/go.mod` require `go 1.26.10`. `GOTOOLCHAIN=auto`, `GOPATH=/home/bayu/go`, `GOROOT=/usr/local/go`. No Go upgrade was needed. |
| Go cache | Default `GOCACHE` resolves under `/home/bayu/.cache/go-build`, which is read-only in this managed shell. Checks were rerun with `GOCACHE=/tmp/naeos-go-cache`; Go also emitted a nonfatal read-only stat-cache warning from the existing module cache. |
| Docker | Docker Engine CLI 29.6.1, Compose v5.3.1. Default `docker version` could not reach `/var/run/docker.sock` (`permission denied`). The `desktop-linux` context exists but `docker --context desktop-linux version` panicked in the CLI. No container/image inspection could be completed because daemon access failed. |
| Supabase | Supabase CLI 2.109.0. Its first version check failed when it tried to write telemetry under the read-only home directory; `HOME=/tmp/codex-supabase-home` allowed the version and local CLI help checks to succeed. |
| Supabase project state | `naeos/supabase/.temp` has link metadata and a project-ref file (values withheld). No `supabase/config.toml` was found. CRM and Intelligence have no local Supabase project directory. No link or remote migration command was run. |
| Node | Node.js v24.18.0 and npm 11.16.0. The CRM declares Node `>=24 <25` and npm `>=11`; its existing `node_modules` directory is present. |
| Disk/memory | Workspace filesystem: 1007G total, 901G available at audit time. Memory: 3.7Gi total, 2.9Gi available; swap 8Gi. |

## Dependencies and configuration changes

- No system dependency installation or upgrade was needed or performed.
- No Go, PATH, shell startup, Git, Docker, Supabase, or repository configuration was changed.
- No repository environment file was read into output or modified. `.env` files are present in the three repositories; the site `.env` is absent. Their contents were withheld. The pre-existing untracked `naeos/.naeos-api.token` path is also present; its value was not read or exposed. CRM's configured `DATABASE_URL` was checked without printing it and resolves to a local host.
- The Go cache override was command-scoped and pointed to `/tmp`; it was not persisted.
- Git identity was already set to `bayu911-source` / `bayu@users.noreply.github.com`. The NAEOS repository documents DCO and has sign-off workflows. SSH GitHub authentication could not resolve `github.com`; `gh auth status` reports its existing token is invalid. No credentials were changed.
- No `core.autocrlf` or `core.eol` global setting was found. No line-ending configuration was changed.

## Verification commands and results

Commands below were run from the corresponding repository unless stated otherwise.

| Check | Result |
| --- | --- |
| `go version`; `go env GOTOOLCHAIN GOPATH GOROOT GOOS GOARCH` | **Passed.** Go 1.26.10, native `linux/amd64`; `auto`, `/home/bayu/go`, `/usr/local/go`. |
| `GOCACHE=/tmp/naeos-go-cache GOPATH=/home/bayu/go go mod tidy -diff` in both Go repositories | **Passed** (exit 0 for both); no repository module edits. |
| `GOCACHE=/tmp/naeos-go-cache GOPATH=/home/bayu/go go test -timeout=120s ./...` in `naeos`, wrapped with `timeout 150s` | **Failed/timed out** (exit 124). Only `cmd/gentest` reported before timeout. |
| Same bounded `go build ./...` in `naeos` | **Passed** (exit 0), with a nonfatal warning that Go could not write its module stat cache under the read-only existing GOPATH cache. |
| Same bounded `go vet ./...` in `naeos` | **Passed** (exit 0). |
| Same bounded `go test -timeout=120s ./...` in `naeos-intelligence`, wrapped with `timeout 150s` | **Failed/timed out** (exit 124); no package results were emitted before timeout. |
| Same bounded `go build ./...` in `naeos-intelligence` | **Passed** (exit 0), with a nonfatal read-only module stat-cache warning. |
| Same bounded `go vet ./...` in `naeos-intelligence` | **Passed** (exit 0). |
| `gofmt -l $(git ls-files '*.go')` in both Go repositories | **Passed** (no unformatted tracked Go files reported). |
| `golangci-lint run ./...` in `naeos`, wrapped with `timeout 150s` | **Failed/timed out** (exit 124); no diagnostics were emitted. |
| `npm run lint` in `naeos-crm` | **Passed** (exit 0). |
| `npm test` in `naeos-crm` | **Failed** (exit 1): 61 passed, 2 skipped, 45 failed across 3 files; 44 setup/runtime errors include `listen EPERM` for `0.0.0.0` from `apps/api/src/integration.spec.ts`. |
| `npm run build` in `naeos-crm` | **Interrupted** after reaching Vite production build `transforming...` with no further progress; not a pass/fail build result. |
| `docker compose config --quiet` in all three repositories | **Passed** (exit 0 for each). This validates configuration syntax only; no containers were started. |
| `docker version`, `docker --context desktop-linux version` | **Failed**: default socket permission denied; Desktop context CLI panicked. |
| `HOME=/tmp/codex-supabase-home supabase --version` | **Passed:** 2.109.0. `supabase init --help`, `supabase migration --help`, and `supabase status --help` also exited 0. |
| `ssh -T -o BatchMode=yes -o ConnectTimeout=5 git@github.com`; `gh auth status` | **Failed/blocked:** DNS resolution unavailable in this shell; existing `gh` token invalid. |

## Safe local development start/stop commands

Use only after Docker access is restored and you verify each repository's local `.env` values and target database hosts. These commands do not remove volumes. Do not run against an unverified project link or production database.

```bash
# NAEOS core (from /home/bayu/workspace/naeos)
docker compose up -d naeos-core
docker compose ps
docker compose logs -f naeos-core
docker compose down

# CRM and its local PostgreSQL/Redis (from /home/bayu/workspace/naeos-crm)
docker compose up -d
docker compose ps
docker compose logs -f
docker compose down

# Intelligence and its local PostgreSQL (from /home/bayu/workspace/naeos-intelligence)
docker compose up -d
docker compose ps
docker compose logs -f
docker compose down
```

`docker compose down` stops/removes the compose containers and network but does not request volume deletion. Do not add `-v`. CRM documents local ports 5434/6380; the service/API ports and migration job are described in each repository's `DOCKER.md`. Review the compose files and environment before starting because the daemon was unavailable for runtime inspection.

## Outstanding blockers and next steps

1. **Docker / Requires user action:** Restore Docker Desktop WSL integration or grant the WSL user normal access to the intended daemon. Then run `docker version`, `docker compose version`, and inspect existing containers/images before starting anything. Do not remove Docker resources.
2. **Go tests / Requires user action:** The main repository test command timed out. Rerun `go test ./...` in a normal WSL session with writable Go build/module caches; investigate any package that remains stuck. Do not weaken tests.
3. **CRM tests / Requires user action:** Rerun from a local shell that permits test listeners on loopback; then distinguish any remaining functional failures from sandbox-only bind errors.
4. **Lint / Requires user action:** `golangci-lint` exceeded the 150 second limit without diagnostics. Rerun after Docker/tests and resource constraints are resolved.
5. **Supabase / Requires user action:** Confirm what project the existing `naeos/supabase/.temp` link refers to. The CLI's local database features require Docker and a local Supabase project configuration; no production link, migration, or database status query was attempted. Treat `naeos-crm-production` as the hosted Auth identity provider and `naeos.db` as the Intelligence database; keep their credentials and migrations separate.
6. **GitHub auth / Requires user action:** Reauthenticate GitHub CLI and verify SSH/DNS from WSL if remote Git operations are needed. Existing Git remotes and credentials were not changed.

## Readiness assessment

**Environment partially ready; Definition of Done not met.** Go matches both module requirements; both Go builds/vet and Compose syntax validation pass; CRM lint passes; Supabase CLI is available. Docker daemon access, test suites, `golangci-lint`, and CRM production build remain unresolved. Do not treat the workspace as ready for the local database-backed NAEOS development stack until Docker access and the test blockers are addressed.
