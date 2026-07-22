# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.5.x   | :white_check_mark: |
| 1.4.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability within NAEOS, please send an email to **security@naeos.dev**. All security vulnerabilities will be promptly addressed.

**Please do not report security vulnerabilities through public GitHub issues.**

### What to include

When reporting a vulnerability, please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: within 48 hours
- **Initial assessment**: within 1 week
- **Fix timeline**: depends on severity

## Security Best Practices

When using NAEOS:

1. **Never commit secrets** to generated artifacts or specifications.
2. **Validate all inputs** — use NAEOS validation before processing specifications.
3. **Review generated code** — always review AI-generated artifacts before deployment.
4. **Keep dependencies updated** — use `go get -u` and monitor for vulnerabilities.
5. **Follow least privilege** — run NAEOS with minimum required permissions.
6. **Use generated passwords** — NAEOS generates random passwords for cloud databases. Always rotate them after deployment.

## Security Features

NAEOS includes the following security controls:

- **Policy Engine** — enforce security rules on specifications and artifacts.
- **Artifact Review** — check for TODOs, placeholders, and license headers.
- **Provenance Tracking** — trace artifacts back to source specifications.
- **Input Validation** — validate all specification inputs before processing.
- **WebSocket Origin Validation** — configurable allowed origins for WebSocket connections.
- **Rate Limiting** — IP-based and API key-based rate limiting.
- **Audit Logging** — track all API operations for security review.

For detailed security specifications, see [docs/NES-020-Security.md](docs/NES-020-Security.md).
