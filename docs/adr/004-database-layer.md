# ADR-004: Database Layer

> Status: Accepted
> Date: 2026-07-20

## Context

NAEOS requires persistent storage for pipeline runs, artifacts, user configurations, and audit logs. The project needs a database abstraction that supports multiple backend engines (PostgreSQL, MySQL, SQLite) without coupling to any single driver. Key requirements:

- Support for embedded (SQLite) and server-based (PostgreSQL, MySQL) deployments
- Connection pooling with configurable limits and timeouts
- Retry logic for transient failures
- Transaction support for atomic operations
- Graceful migration from embedded to server-backed storage as projects scale

## Decision

Adopt a **database adapter pattern** with a common `Database` interface and separate implementations for PostgreSQL, MySQL, and SQLite.

Each adapter implements: `Connect`, `Close`, `Exec`, `Query`, `Begin`, `Ping`, and a transaction-aware `WithTx` helper. All adapters are registered in a central registry and selected via config.

Connection pooling is configured via `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` from `database/sql`. A `QueryLogger` decorator wraps all adapters for structured logging of slow queries.

## Consequences

### Positive

- Users can start with SQLite for local development and migrate to PostgreSQL for production without code changes
- The adapter pattern makes it straightforward to add new database backends (e.g., CockroachDB, SQL Server)
- All adapters share connection lifecycle, retry, and logging behavior via composable wrappers
- Health checks (`Ping`) are wired into the API `/healthz` endpoint automatically

### Negative

- Each database has subtle SQL dialect differences (e.g., `RETURNING` clause, index syntax, type mappings) requiring adapter-specific query logic
- Feature parity across adapters is not guaranteed; some PostgreSQL-only features (e.g., LISTEN/NOTIFY, JSONB operators) have no SQLite equivalent
- Testing must be run against all three backends to validate dialect correctness

### Mitigations

- Queries use a common DSL where possible; adapter-specific queries live in per-adapter files
- CI runs the full test suite against all three database backends in parallel
- The `database/sql` interface provides a baseline; extensions are opt-in and documented per adapter
- Migration files are backend-agnostic; dialect differences are resolved at the migration loader level

## Notes

- Implementation: `internal/database/` — adapters, registry, migration loader, query logger
- Related NES documents: NES-042 (Database Layer Specification)
- Migration engine in `internal/migration/` handles schema versioning independently
