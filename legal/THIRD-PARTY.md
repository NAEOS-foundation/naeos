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
