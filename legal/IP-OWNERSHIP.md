# NAEOS IP Ownership & Provenance

Status: Active draft  
Version: 1.0

## 1. Purpose

NAEOS must be able to establish where material came from, who had the right to submit it, and which license governs its use.

This document is an evidence policy. It is not an assignment agreement.

## 2. Asset classes

| Asset | Required evidence |
|---|---|
| Source code | Git history, DCO, license |
| Documentation/specifications | Git history, DCO, source/provenance record |
| Generated artifacts | Generator/source reference and license review |
| Logos/brand assets | Creation record and trademark review |
| Third-party code | Upstream source, license, attribution |
| Dependencies | Lockfile, SBOM, license scan |
| External contributions | Contributor identity + DCO |
| Commercial-only code | Separate repository/entity and explicit license |

## 3. Existing repository baseline

The repository currently uses Apache-2.0 and DCO-based contribution acceptance. Contributors retain copyright in their contributions while granting the project the rights provided by Apache-2.0.

NAEOS must not describe DCO as a copyright assignment.

## 4. AI-assisted material

AI assistance does not by itself establish ownership or provenance. The contributor remains responsible for the right to submit the material and for third-party license compliance.

See `docs/ai-provenance.md`.

## 5. Ownership register

Before a commercial launch, maintain a private or access-controlled register containing:

- asset name;
- creator/contributor;
- creation date;
- repository/commit;
- applicable license;
- third-party material;
- assignment/license agreement, if any;
- evidence location;
- review status.

Do not place confidential agreements or personal data in the public repository.
Use [`evidence/asset-provenance-register-template.csv`](evidence/asset-provenance-register-template.csv)
as a field checklist; populate it only in the approved private register, not in
the public repository.

## 6. Prohibited assumptions

The project must not assume that:

- public availability means public-domain status;
- AI-generated output is automatically free of third-party rights;
- a GitHub account proves legal ownership of an employer's work;
- DCO transfers copyright;
- an open-source license grants trademark rights.

## 7. Commercial boundary

Commercial services may be built around Apache-2.0 NAEOS code, but rights already granted under Apache-2.0 must not be retroactively withdrawn.

Any proprietary component must have a separately documented ownership and licensing basis.
