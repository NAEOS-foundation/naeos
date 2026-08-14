---
name: Good First Issue
about: A great entry point for first-time contributors to NAEOS
title: "[GOOD FIRST ISSUE] "
labels: ["good first issue"]
assignees: ""
---

## Summary

Briefly describe the task.

## Why It's a Good First Issue

Explain why this task is approachable for a first-time contributor (limited
scope, no complex architecture knowledge needed, well-contained changes).

## Tasks

- [ ] Task 1
- [ ] Task 2

## Where to Look

Point the contributor to the relevant files:

- `internal/...` — what to modify
- `cmd/naeos/...` — CLI wiring
- `internal/..._test.go` — tests to add or update

## Definition of Done

- [ ] Code change implemented
- [ ] Unit tests added or updated (table-driven, per project conventions)
- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run ./...` passes

## Resources

- [Contributing guide](../../CONTRIBUTING.md)
- [Project conventions (AGENTS.md)](../../AGENTS.md)
- [Development plan](../../DEVELOPMENT_PLAN.md)

## Additional Context

Any other context, links, or references that help get started.
