# NAEOS Third-Party Software Policy

Status: Active  
Version: 1.0

## 1. Purpose

NAEOS must maintain an auditable record of third-party software and the obligations attached to each dependency.

## 2. Required release evidence

Each release should retain:

1. dependency lockfiles;
2. generated SBOM;
3. license inventory;
4. attribution/NOTICE material;
5. vulnerability scan result;
6. exceptions and their approvals.

## 3. License classes

Dependencies must be classified as:

- Permissive: MIT, BSD, ISC, Apache-2.0 and similar;
- File-level/copyleft: e.g. MPL-2.0;
- Strong copyleft: GPL/AGPL and similar;
- Proprietary/restricted;
- Unknown/unresolved.

Unknown or unresolved licenses are release blockers until reviewed.

## 4. Existing NOTICE alignment

The current `NOTICE` records the repository's known third-party licensing position, including the MySQL driver under MPL-2.0 and website/build-time dependencies.

Any future dependency change must update the relevant attribution evidence.

The Supabase API authentication implementation pins `keyfunc/v3` v3.9.2
(Apache-2.0), `jwkset` v0.11.4 (Apache-2.0), `golang-jwt/jwt/v5` v5.3.1
(MIT), and `golang.org/x/time` v0.15.0 (BSD-3-Clause). These dependencies
are used unmodified; their upstream license texts are included in the
production container under `/usr/local/share/licenses/naeos/`. Their
SPDX identifiers are curated in `internal/sbom/gomod.go`.

## 5. Modifications

If a third-party component is modified, record:

- upstream version;
- modified files;
- modification summary;
- applicable license obligations;
- source availability requirements, where applicable.

## 6. Generated and vendored content

Generated output is not automatically first-party. The generating tool, template, source dataset, upstream package, and applicable license must be considered before distribution.

## 7. Release gate

A release must not be published when a dependency's applicable license or attribution obligation is materially unresolved.
