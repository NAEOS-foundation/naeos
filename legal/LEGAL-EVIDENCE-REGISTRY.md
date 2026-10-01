# NAEOS Legal Evidence Registry

Status: Active draft  
Version: 1.0

## Purpose

The Legal Evidence Registry defines the minimum evidence bundle that should accompany a NAEOS release or other legally significant distribution.

It turns legal/compliance assertions into inspectable engineering evidence.

## Evidence chain

```
Release
  ↓
Commit / Source
  ↓
Contributor provenance / DCO
  ↓
Dependencies / SBOM
  ↓
License & attribution review
  ↓
Security evidence
  ↓
Trademark/branding review
  ↓
Exceptions / approvals
  ↓
Release decision
```

## Registry record

Each release record should identify:

| Field | Requirement |
|---|---|
| release_version | exact release identifier |
| commit_sha | immutable source commit |
| release_date | UTC date |
| repository | canonical repository |
| license | effective software license |
| dco_status | contributor-signoff verification |
| sbom | artifact/reference and digest |
| dependency_license_status | pass/fail + evidence |
| attribution_status | NOTICE/third-party review |
| security_status | relevant scan/review |
| provenance_status | first-party and generated-content review |
| trademark_status | branding review |
| exceptions | unresolved items and disposition |
| reviewer | responsible reviewer |
| decision | PASS or HOLD |

## Evidence identity

Evidence artifacts should be content-addressable where practical. Prefer a cryptographic digest over a mutable URL alone.

Example:

```json
{
  "artifact": "sbom.cdx.json",
  "sha256": "<64-hex-digest>",
  "source_commit": "<40-hex-commit>"
}
```

## Decision semantics

### PASS

The release evidence satisfies all required controls for the release scope, with no unresolved material legal blocker.

### HOLD

At least one material legal, licensing, attribution, provenance, trademark, or contractual issue remains unresolved.

A PASS is not a legal opinion or warranty.

## Exceptions

Every exception must contain:

- identifier;
- affected release/component;
- issue;
- risk/impact description;
- mitigation;
- approving authority;
- expiration/review date.

Do not hide exceptions in free-form release notes.

## Retention

Release evidence should be retained for the lifetime of the release artifacts and for any longer period required by applicable agreements or law.

## Privacy

Do not store unnecessary personal data, private contracts, credentials, or confidential customer information in public evidence. Store sensitive evidence in an access-controlled system and retain only a reference/digest in the public registry.

## Relationship to NAEOS Evidence

Legal evidence is one evidence class within the broader NAEOS evidence model. It should remain distinguishable from runtime evidence and security claims.

```
Engineering Evidence
├── Runtime Evidence
├── Verification Evidence
├── Security Evidence
└── Legal / Provenance Evidence
```
