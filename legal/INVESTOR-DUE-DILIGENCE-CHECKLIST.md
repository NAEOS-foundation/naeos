# NAEOS Investor Due-Diligence Evidence Checklist

Status: Working checklist  
Version: 1.0

Use this checklist before a fundraising process, investor meeting, diligence request, or any public statement that references traction, ownership, product maturity, commercial readiness, or governance. It is an evidence-control tool and not legal advice, a warranty, or a substitute for qualified counsel.

Keep private contracts, founder employment records, customer information, credentials, and sensitive legal advice out of the public repository. Store sensitive diligence evidence in an approved access-controlled register and record only a safe reference, digest, review status, and owner here.

## 1. Objective

An investor should be able to verify from the repository and controlled evidence that:

- the project has a clear ownership and provenance baseline;
- the open-source license and dependency posture are understood;
- the operating company or contracting entity is identified and has authority to bind it;
- contributions and code rights are not ambiguous in a way that undermines future product or fundraising claims;
- the public product claims match what is actually implemented and released;
- commercial and cloud/enterprise plans are separated from Apache-2.0 repository code with clear boundaries;
- no material security, privacy, or customer-data issue remains hidden or unexplained.

## 2. Investor diligence questions to answer

| Question | Required evidence |
|---|---|
| Who owns the code? | Git history, contribution records, DCO workflow, IP ownership register, founder/employee/contractor assignment evidence |
| Who owns the brand? | Trademark clearance review, legal entity record, logo usage policy, product naming evidence |
| Are all contributors covered? | DCO or equivalent contribution control, AI-provenance policy, contributor identity records |
| Are dependencies license-compatible? | SBOM, lockfile review, third-party inventory, license scan |
| Are there GPL/AGPL contamination risks? | Dependency review, vendored code review, security and license scan results |
| Are contractor and founder IP rights assigned? | Executed agreements and evidence of chain of title |
| Is the open-source strategy documented? | Repository license policy, commercial boundary documentation, product/service boundary statements |
| Is the commercial strategy compatible with the OSS license? | Public license statements, commercial boundary policy, hosted/service terms, separate commercial review |
| Are there unresolved third-party IP issues? | Provenance review, attribution checks, vendor review, unknown-license disposition |
| Are secrets or customer data present? | Secret scan, data map review, public-disclosure checks, operational controls |
| Is the repo clean enough for diligence? | Current legal and security baselines, repository inventory, review approvals |

## 3. Ownership, entity, and governance evidence

- [ ] Confirm the legal entity that controls the project and any intended commercial offering.
- [ ] Confirm the jurisdiction and contracting authority of the entity before any public fundraising materials rely on it.
- [ ] Confirm the chain of rights for founder, employee, contractor, freelancer, agency, and pre-existing work.
- [ ] Check that the repository does not imply a legal conclusion not supported by the evidence.
- [ ] Maintain a private rights register that maps creator, asset, source, and review status.
- [ ] Review whether any code, data, or design was created under employer or customer work-for-hire terms that may limit ownership or distribution.
- [ ] Document unresolved ownership questions with a disposition and owner.

## 4. Licensing and dependency evidence

- [ ] Confirm the repository's Apache-2.0 baseline is present and unchanged.
- [ ] Review license and provenance records for all direct and transitive dependencies.
- [ ] Verify third-party code, assets, and vendored material have identified sources and compatible licenses.
- [ ] Reconcile dependency results with the SBOM, lockfiles, and the repository's third-party policy.
- [ ] Confirm there are no unreviewed GPL/AGPL/LGPL/SSPL or source-available surprises in the distribution path.
- [ ] Record unknown or unusual dependency licenses as unresolved until reviewed.
- [ ] Ensure public claims do not describe any code path as open-source or commercial without the precise boundary and evidence.

## 5. Contributor and provenance evidence

- [ ] Confirm the contribution model (for example DCO) is consistent with operating assumptions and project governance.
- [ ] Review contributor records for ambiguous or employee-owned submissions.
- [ ] Confirm AI-assisted contributions are documented in the relevant provenance policy and audit trail.
- [ ] Confirm generated code, docs, images, and configuration are traceable to a human owner or approved source.
- [ ] Review contributions for copied code, snippets, or third-party materials without attribution or a clear license basis.
- [ ] Confirm there is an evidence path from a contributor to a repository submission and to the final merged result.

## 6. Trademark, brand, and public-claim evidence

- [ ] Confirm the legal and brand owners for NAEOS, NEIR, product names, and logos.
- [ ] Verify that public materials do not imply certification, endorsement, or affiliation without evidence.
- [ ] Ensure fundraising materials do not overstate trademark rights or registration status where proof is incomplete.
- [ ] Check that new product names, investor-facing materials, and launch language have been reviewed for brand risk.
- [ ] Separate open-source brand statements from any future commercial or enterprise service positioning.

## 7. Security, privacy, and product evidence

- [ ] Confirm there is no committed secret, personal data, customer material, or confidential infrastructure data in the public repository.
- [ ] Review the current security posture for actual production, demo, or hosted-service behavior, not only repository tests.
- [ ] Verify the actual data flow for website, newsletter, API, AI-provider calls, and cloud dependencies.
- [ ] Check whether public privacy terms match the current implementation and service boundaries.
- [ ] Review operational procedures for access control, incident handling, retention, deletion, and support access.
- [ ] Confirm any public launch, product, or trust claim has a clear source of evidence behind it.

## 8. Commercial and open-core boundary evidence

- [ ] Document which components are in the Apache-2.0 repository and which are expected to remain proprietary or hosted.
- [ ] Identify the exact boundary between open-source code, managed services, enterprise offerings, and private tooling.
- [ ] Verify that open-source distribution is not undermined by hidden commercial restrictions or assumption-based licensing.
- [ ] Confirm there is a documented rationale for any future commercial wedge, with evidence that it is supported by the technical architecture and product plan.
- [ ] Keep public statements conservative until the actual product and service boundaries are agreed and documented.

## 9. Fundraising materials and claims

- [ ] Each market claim, benchmark, product claim, feature statement, and adoption claim should be tied to evidence.
- [ ] Do not rely on roadmap items as current facts.
- [ ] Distinguish technical potential from realized product maturity.
- [ ] Do not state customer, partner, certification, or deployment claims unless verified and approved.
- [ ] Keep investor materials consistent with repository evidence and the current documented product stage.

## 10. Decision gate

Record the decision in the private register:

- **GO** — due-diligence evidence is complete, material gaps are resolved, and claims match repository and operational evidence.
- **HOLD** — any unresolved ownership, licensing, privacy, security, commercial-boundary, or public-claim issue remains open.
- **NOT APPLICABLE** — explain why no external fundraising diligence is being carried out; do not use this as a substitute for required review.

A GO is an internal due-diligence control point, not legal advice or a warranty.

## 11. Repository evidence to reconcile before fundraising

The repository includes concrete evidence that should be reconciled before any external or investor-facing claim is made:

| Repository evidence | What to verify before fundraising |
|---|---|
| `LICENSE` and the Apache-2.0 baseline in the root repository | Confirm the exact scope of the open-source code and whether any proprietary or hosted features are outside the repo boundary |
| `legal/README.md` and the legal architecture docs | Confirm the current legal entity, stewardship, and unresolved questions are not being overstated |
| `legal/IP-OWNERSHIP.md` | Confirm founder, employee, contractor, and external-contribution rights are evidenced and not assumed |
| `legal/COMMERCIAL-LAUNCH-CHECKLIST.md` | Confirm that any commercial or hosted plan is covered by its own evidence and service-boundary review |
| `site/content/privacy.md` and `site/content/terms.md` | Confirm public notices match actual analytics, hosting, AI-provider, and service behavior |
| `docs/ai-provenance.md` and contribution guidance | Confirm AI-assisted material is traceable to a human owner and subject to a documented review path |
| Third-party policy and lockfiles | Confirm all dependencies and generated assets are inventoried and reviewable |

Before external fundraising documentation is published, compare the repository and live service behavior with the actual operating company, product boundaries, and legal entity information. Do not treat the repository as a substitute for counsel-reviewed term sheets, founder equity documents, or jurisdiction-specific legal advice.
