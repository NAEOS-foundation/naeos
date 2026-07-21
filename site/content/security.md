---
title: Security
description: Security policy and vulnerability reporting for NAEOS.
---

## Security Policy

The NAEOS project takes security seriously. We appreciate your efforts in responsibly disclosing vulnerabilities.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, report them via one of the following methods:

1. **Email** — Send details to the maintainers at `security@naeos.dev`
2. **GitHub Private Vulnerability Reporting** — Use the "Report a vulnerability" feature under the repository's Security tab

### What to Include

When reporting a vulnerability, please include:

- Type of vulnerability
- Steps to reproduce
- Affected versions
- Potential impact
- Any suggested fixes (if known)

## Scope

The following are in scope for security reports:

- NAEOS CLI tool (`github.com/NAEOS-foundation/naeos`)
- NAEOS website (`naeos.dev`)
- Official NAEOS packages and distributions

## Disclosure Policy

We follow a coordinated disclosure process:

1. Reporter submits vulnerability details
2. Maintainers acknowledge receipt within 48 hours
3. Maintainers investigate and develop a fix
4. Fix is released according to severity:
   - **Critical** — 7 days
   - **High** — 14 days
   - **Medium** — 30 days
   - **Low** — 90 days
5. Public disclosure after the fix is released

## Security Measures

NAEOS incorporates several security measures:

- **No telemetry** — NAEOS CLI does not send any usage data
- **Local execution** — All code generation runs locally on your machine
- **Dependency scanning** — Automated vulnerability scanning via Dependabot
- **Code review** — All contributions are reviewed before merging
- **Signed releases** — Release artifacts are signed

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest release | ✅ |
| Previous release | 🚧 Security fixes only |
| Older releases | ❌ |
