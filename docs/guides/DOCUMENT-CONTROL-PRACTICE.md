<!-- Copyright 2026 NAEOS contributors -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

Summary: A vendor-neutral operational guide for establishing document control in an engineering or delivery team. This guide is informational and does not define NAEOS product behavior or override repository specifications.

# Document control practice

Last updated: 2026-11-01

## Overview

Document control keeps working documents identifiable, reviewable, current, and retrievable. It reduces the risk of teams acting on obsolete instructions, unapproved changes, or records with unclear ownership.

This is a general operating practice. It is not a NAEOS runtime feature, compliance certification, legal opinion, or substitute for an organization's approved quality-management system.

## Concepts

Use a consistent set of metadata for each controlled document:

| Field | Purpose |
|---|---|
| Document ID | Stable identifier that does not depend on the filename |
| Title | Human-readable document name |
| Owner | Role accountable for accuracy and maintenance |
| Version | Revision identifier |
| Status | Draft, In review, Approved, Superseded, or Archived |
| Classification | Public, Internal, Confidential, or Restricted |
| Created / updated | Record of document lifecycle dates |
| Reviewer / approver | Roles that performed review and authorization |
| Source of truth | Canonical location of the current approved copy |
| Related records | Links to change requests, decisions, or evidence |

Avoid embedding personal data, credentials, customer secrets, or confidential commercial terms in public repositories.

## Architecture

A lightweight document-control arrangement has four logical areas:

1. **Register** — searchable metadata for controlled documents.
2. **Controlled storage** — canonical copies, separated by access classification.
3. **Change log** — append-only record of material revisions and decisions.
4. **Archive** — superseded records retained under a defined retention policy.

Keep public guidance separate from internal working records. A public repository should contain only material explicitly reviewed for public release. Internal registers, approval records, customer documents, contracts, invoices, and working templates belong in access-controlled storage.

## Workflow

### 1. Register a document

Assign a stable ID, owner, classification, version, and initial status. Record the canonical storage location. Do not use a filename as the sole identifier.

### 2. Draft and review

Mark new or materially changed documents as Draft. The owner requests review from the relevant subject-matter reviewer. Review comments and required changes should be traceable to the revision under review.

### 3. Approve and publish

Only an authorized approver may move a document to Approved. Record the approver role, decision date, and approved version. Publish the approved copy to its canonical location and make its status visible in the register.

### 4. Revise

Create a new revision rather than silently overwriting an approved record. Record the reason for change, summary of changes, author, reviewer, approver, and effective date. Reassess linked documents when a change affects them.

### 5. Supersede and archive

When a new version becomes effective, mark the previous version Superseded. Retain or archive it according to the organization's retention requirements. Do not delete records that are subject to legal hold, contractual obligations, or an active investigation.

## Configuration

### Status model

| Status | Meaning |
|---|---|
| Draft | Work in progress; not authoritative |
| In review | Submitted for review; not yet approved |
| Approved | Authorized current version |
| Superseded | Replaced by a later approved version |
| Archived | Retained for reference; not current |

### Access model

Apply least privilege:

- Public: approved for unrestricted publication.
- Internal: available to authorized team members.
- Confidential: restricted to named roles or groups with a business need.
- Restricted: tightly limited; access must be explicitly approved and periodically reviewed.

Use storage that supports identity-based access, version history, recovery, and auditability. A spreadsheet register is a coordination aid; by itself it is not an access-control system, immutable audit log, e-signature service, or records-management platform.

### Naming convention

Use stable, predictable names such as:

`<document-id>_<short-title>_v<major>.<minor>.<patch>.<extension>`

Keep the document ID and version in document metadata as well as the filename where practical.

## Examples

A generic register row might contain:

| Document ID | Title | Owner | Version | Status | Classification |
|---|---|---|---|---|---|
| DOC-OPS-001 | Release checklist | Operations owner | 1.0.0 | Approved | Internal |

This example is illustrative only. Do not treat it as an actual NAEOS-controlled record.

A change-log entry should capture the document ID, old and new versions, change reason, change summary, author, reviewer, approver, decision date, and link to the relevant change request.

## Troubleshooting

- **Two copies appear current:** designate one canonical source, compare revisions, and resolve the conflict through the owner and approver.
- **An obsolete copy is in circulation:** mark it Superseded, notify affected users, and point them to the canonical version.
- **An approval is missing:** keep the document in Draft or In review; do not infer approval from publication or informal acknowledgement.
- **A document is misclassified:** restrict access while the owner or information-security role assesses the correct classification and any exposure.
- **A record is missing:** check the register, archive, version history, and retention process before recreating it.

## Validation

Before adopting the process, verify that:

- Every controlled document has a unique ID and named owner.
- Status and classification values are consistently applied.
- Only authorized roles can approve documents.
- Approved copies have a canonical location and version history.
- Changes are traceable to a reason and review decision.
- Superseded copies are clearly marked and retained appropriately.
- Public release receives a separate information-classification review.
- Backup, recovery, retention, and access reviews are tested periodically.

## Limitations

This guide provides general operational guidance. Each organization must define its own approval authority, retention schedule, privacy obligations, contractual controls, and applicable regulatory requirements. Seek qualified legal, security, or compliance advice where required.
