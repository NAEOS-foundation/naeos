# NAEOS Legal Release Checklist

Status: Active  
Version: 1.0

Use this checklist before every public software release.

## A. License

- [ ] `LICENSE` is present and unchanged unless intentionally reviewed.
- [ ] Release artifacts preserve Apache-2.0 attribution.
- [ ] No first-party file was silently relicensed.
- [ ] Contributor changes passed DCO checks.

## B. Provenance

- [ ] New first-party material has identifiable provenance.
- [ ] AI-assisted contributions comply with `docs/ai-provenance.md`.
- [ ] No copied third-party material with unclear rights is included.
- [ ] Generated artifacts have an attributable source.

## C. Dependencies

- [ ] Dependency lockfiles are current.
- [ ] SBOM is generated.
- [ ] License inventory is reviewed.
- [ ] NOTICE/third-party attribution is current.
- [ ] Unknown or incompatible licenses are resolved.
- [ ] Security/dependency scans have been reviewed.

## D. Branding

- [ ] NAEOS/NEIR/logo usage is consistent with the branding policy.
- [ ] Release notes do not imply certification or endorsement.
- [ ] New product names have been routed through trademark clearance before public launch claims.

## E. Commercial boundary

- [ ] New proprietary components are outside the Apache-2.0 repository boundary or have explicit licensing.
- [ ] Existing Apache-2.0 rights are not restricted retroactively.
- [ ] Hosted/enterprise claims do not imply rights beyond the published terms.
- [ ] If the release enables or changes a paid offer, the
      [Commercial Launch Evidence Checklist](COMMERCIAL-LAUNCH-CHECKLIST.md)
      has been reviewed for that offering.
- [ ] If the release or repository is used in a fundraising, diligence, or
      investor-material context, the
      [Investor Due-Diligence Evidence Checklist](INVESTOR-DUE-DILIGENCE-CHECKLIST.md)
      has been reviewed.

## F. Evidence

Record:

- release version;
- release commit;
- SBOM identifier;
- license-scan result;
- security-scan result;
- reviewer;
- date;
- exceptions and disposition.

## Release decision

**PASS** — all required controls are satisfied.

**HOLD** — any material licensing, provenance, attribution, trademark, or contractual issue remains unresolved.

This checklist is an engineering control and does not replace jurisdiction-specific legal review.
