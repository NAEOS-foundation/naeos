---
name: Plugin / Profile / Integration Contribution
about: Contribute, fix, or extend a plugin, profile, adapter, or NAEOS integration
title: "[ECOSYSTEM] "
labels: enhancement
assignees: ""
---

## What are you contributing?

A plugin, profile, adapter, or integration for NAEOS:

- [ ] Plugin (WASM or Go) — see [`internal/pluginsdk`](../../internal/pluginsdk)
- [ ] Profile — see [`profile/`](../../profile) and [`internal/profile`](../../internal/profile)
- [ ] AI tool adapter / MCP integration — see [`internal/mcp`](../../internal/mcp)
- [ ] Template — see [`templates/`](../../templates) and [`examples/templates/`](../../examples/templates)
- [ ] Other (describe)

## Feature area

Which engineering area does it plug into?

- Specification / NEIR / compiler / pipeline
- Governance / policy / evidence / verification
- Generation / adapters / output targets
- Runtime / gateway / observability
- Marketplace / ecosystem
- Other

## Why it exists

The problem it solves and the developer benefit.

## Minimal example

A short specification, configuration, or command showing how it is used.

## For the maintainers

- [ ] I have read [CONTRIBUTING.md](../CONTRIBUTING.md)
- [ ] Tests added (unit + integration where relevant)
- [ ] `go test -race ./...` passes locally
- [ ] `golangci-lint run ./...` passes locally
- [ ] Documentation/example (`README`, `docs/`, `examples/`) updated
- [ ] No secrets or environment-specific absolute paths in committed files

## Additional Context

Links, references, or related discussions/PRs.