# Security Policy

## Supported Versions

NAEOS follows the current release line for security support. The versions below reflect the repository's current public support policy.

| Version | Support |
|---|---|
| 3.6.x | :white_check_mark: Current release line |
| 3.5.x | :white_check_mark: Security support |
| 3.4.x | :warning: Legacy / upgrade recommended |
| < 3.4.x | :x: Not supported |

The current NAEOS release is **NAEOS 3.6.0**. Security support is maintained against the current release line unless a release-specific advisory states otherwise.

## Reporting a Vulnerability

If you discover a security vulnerability within NAEOS, please send an email to **security@naeos.dev**. All security vulnerabilities will be promptly addressed.

**Please do not report security vulnerabilities through public GitHub issues.**

### What to include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix, if any
- Affected version or commit

### Response Timeline

- **Acknowledgment:** within 48 hours
- **Initial assessment:** within 1 week
- **Fix timeline:** depends on severity

## Security Engineering Scope

NAEOS includes security-oriented controls across policy, runtime, handoff, evidence, and verification boundaries, including:

- **Policy evaluation** — deterministic authorization decisions separate from agent intent.
- **Runtime enforcement** — execution is checked against authorization and constraints.
- **Artifact integrity** — artifact-sensitive authorization can bind execution to a digest.
- **Handoff governance** — recipient, capability, provenance, replay, version, and signature checks.
- **Independent observation** — evidence construction can consume independently observed runtime events.
- **Durable runtime evidence** — append-only runtime ledger and durable receipt/recovery experiments.
- **Audit/evidence** — durable records intended to outlive agent memory.
- **Dependency scanning** — automated vulnerability checks in CI.
- **Fuzz testing** — deterministic fuzz targets for parser and migration boundaries.

Detailed security specifications are documented in [docs/NES-020-Security.md](docs/NES-020-Security.md).

## Security Best Practices

1. Never commit secrets to source code, generated artifacts, specifications, or examples.
2. Validate all inputs before processing specifications or execution requests.
3. Review AI-generated artifacts before deployment.
4. Keep dependencies updated and review vulnerability reports.
5. Follow least privilege.
6. Treat AI-agent intent as untrusted input until authorized by policy.
7. Do not treat repository experiments as proof of security for every production deployment.

## Security Claim Discipline

NAEOS experiments demonstrate behavior of named deterministic harnesses and repository components. They are not a substitute for a production penetration test, immutable infrastructure, remote attestation, distributed consensus, or externally trusted telemetry unless those capabilities are explicitly implemented and independently verified.

Security claims should identify the exact enforcement boundary, evidence source, test/experiment, and known limitations.
