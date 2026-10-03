# NAEOS Commercial Launch Evidence Checklist

Status: Working checklist  
Version: 1.0

Use this checklist before offering a paid assessment, pilot, hosted service, or
enterprise product. It is an evidence and launch-control tool, not a contract,
privacy notice, legal opinion, or substitute for qualified counsel.

Keep private contracts, personal data, credentials, customer data, and
confidential legal advice out of this public repository. Store sensitive
evidence in an approved access-controlled register and record only a safe
reference, digest, review status, and owner here.

## 1. Define the offering

Complete one scope record for each distinct offer. Do not treat a software
release, architecture assessment, time-limited pilot, and hosted service as the
same product.

| Field | Required record |
|---|---|
| Offering name and version | Exact public name and revision |
| Offering type | Open-source software, assessment, pilot, hosted SaaS, support, or enterprise product |
| Customer and users | Intended contracting customer and user roles |
| Included components | Core code, hosted components, integrations, support, and third-party services |
| Excluded components | Capabilities not supplied or not part of the offer |
| Deployment boundary | Customer-hosted, NAEOS-hosted, or split responsibility |
| Data boundary | Data supplied, generated, observed, retained, or sent to providers |
| Stage | Proposed, design partner, pilot, generally available, or retired |
| Evidence owner | Role accountable for keeping this record current |

## 2. Ownership, licensing, and brand evidence

- [ ] Identify the legal entity that will offer the product and confirm its
  jurisdiction and authority to contract.
- [ ] Confirm the chain of rights for founder, employee, contractor, freelancer,
  agency, and pre-existing work with the relevant rights holder and counsel.
- [ ] Use the private asset register described in
  [IP-OWNERSHIP.md](IP-OWNERSHIP.md); record evidence references, not agreement
  contents, in public artifacts.
- [ ] Identify code or assets outside the Apache-2.0 Core and record their
  owner, license, and distribution boundary.
- [ ] Verify that product names, logos, and third-party marks have appropriate
  clearance and do not imply affiliation, certification, or endorsement.
- [ ] Keep trademark registration status unclaimed unless supported by
  jurisdiction-specific evidence.
- [ ] Review open-source and third-party obligations for each shipped binary,
  container, SDK, website artifact, and runtime asset using
  [THIRD-PARTY.md](THIRD-PARTY.md) and the applicable SBOM.

## 3. Data-flow and privacy evidence

Create a data-flow record for every surface in the offer, including website,
newsletter, assessment intake, pilot environment, API/control plane, hosted
service, support channel, and operational telemetry.

| Data-flow item | Evidence to collect |
|---|---|
| Data categories | Input, account, specification, code, execution, log, evidence, and support data actually processed |
| Purpose and necessity | Product function served by each category |
| System boundary | Source, destination, storage location, and whether data leaves the customer's environment |
| Service providers | Provider, service, purpose, data involved, and current provider terms |
| Retention and deletion | Actual retention period, deletion trigger, backup behavior, and deletion evidence |
| Access and security | Roles, support access, credentials, encryption, audit events, and incident process |
| Transfer and region | Actual processing and storage locations; counsel review where cross-border transfer may apply |
| User requests | Operational owner and tested route for access, correction, deletion, or other applicable requests |
| Public notice | Exact public page and revision that matches the implemented behavior |

- [ ] Compare the data map with the implementation and deployment configuration;
  do not rely on planned behavior.
- [ ] Reconcile the map with the public English and Indonesian privacy pages,
  cookie notice, newsletter behavior, and any service-specific notice.
- [ ] Have qualified counsel determine applicable controller/processor roles,
  legal bases, notice content, data-subject rights, and cross-border obligations.
- [ ] Verify the actual analytics, hosting, newsletter, authentication, logging,
  AI-provider, and support subprocessors for the offer.
- [ ] Test deletion, retention, access restriction, and incident escalation
  against the documented operating procedure.

## 4. Customer and service terms

Do not use this checklist as contract text. Counsel should review terms against
the defined offer, actual system behavior, contracting entity, and target
jurisdictions.

- [ ] Applicable website/software terms distinguish the Apache-2.0 software
  license from any separate hosted-service agreement.
- [ ] Assessment or pilot scope, participants, duration, deliverables,
  acceptance/evaluation method, fees, expenses, and exit process are documented.
- [ ] Customer data, specifications, source code, outputs, logs, evidence,
  feedback, and pre-existing materials have clearly assigned handling and
  ownership terms.
- [ ] Confidentiality, permitted support access, security incident
  notifications, retention, return, and deletion are addressed.
- [ ] Hosted-service terms cover account access, acceptable use, service
  suspension, service changes, termination, and data export/deletion.
- [ ] A data-processing agreement or equivalent is reviewed if required for the
  actual processing relationship.
- [ ] Support scope, supported versions, response targets, service availability,
  maintenance, exclusions, and escalation are documented before making service
  level commitments.
- [ ] Fees, invoicing, tax, refunds, renewal, and payment failure handling match
  the contracting and billing process.
- [ ] Warranty, liability, indemnity, governing law, dispute handling, and
  consumer/business terms are approved by counsel for relevant jurisdictions.
- [ ] Third-party AI, cloud, identity, analytics, and communications terms are
  checked against the service's data and use boundaries.

## 5. Operational readiness

- [ ] Name a service owner, security contact, privacy contact, and escalation
  path for the offering.
- [ ] Define production access control, key management, audit logging, backup,
  restore, vulnerability response, and incident handling.
- [ ] Run and retain a deployment-specific security review; do not infer hosted
  service security from repository tests alone.
- [ ] Record actual availability, recovery, and performance evidence before
  publishing reliability or service-level claims.
- [ ] Test onboarding, support request, customer exit, export, and data
  deletion paths end to end.
- [ ] Separate production, pilot, staging, and development data and access.
- [ ] Document what happens when the service or a required third-party provider
  is unavailable.

## 6. Public claims and launch gate

- [ ] Public product pages identify the offer's current stage and do not present
  a proposed feature as available.
- [ ] Each security, privacy, availability, benchmark, compatibility, and
  compliance claim links to evidence that supports its exact scope.
- [ ] No customer, partner, certification, deployment, or endorsement is named
  without verification and permission.
- [ ] Terms, privacy pages, product documentation, onboarding, and sales
  materials describe the same service boundary.
- [ ] Required legal and operational reviews have an identified approver and
  review date.

Record the decision in the private register:

- **GO** — required evidence is complete, operational behavior matches the
  approved terms/notices, and no material blocker is open.
- **HOLD** — any material ownership, licensing, privacy, security, contract,
  operational, or public-claim question remains unresolved.
- **NOT APPLICABLE** — explain the offer-specific reason; do not use this value
  to bypass a review that is still needed.

A GO is an internal launch-control decision, not a legal opinion, compliance
certification, or warranty.

## 7. Repository evidence to reconcile before launch

The current repository includes English and Indonesian
[Terms of Service](../site/content/terms.md) and
[Privacy Policy](../site/content/privacy.md) pages. They describe the website,
open-source CLI, newsletter, and community, but this checklist does not
establish that their statements match every current deployment or any future
paid assessment, pilot, hosted, or enterprise service.

The repository contains concrete points that require verification and
reconciliation:

| Repository evidence | Follow-up before a paid offer |
|---|---|
| The privacy page says CLI specifications, projects, and generated code never leave the local machine. The CLI reads specification files in `cmd/naeos/ai_cmd.go`; `internal/ai/llm.go` sends prompts to the configured OpenAI-compatible, OpenAI, or Anthropic endpoint, with Ollama also supported. | Map which commands transmit which inputs to which configured endpoint. Confirm provider and user configuration behavior, then ask counsel to review accurate English and Indonesian disclosures. This evidence does not mean every CLI command transmits data. |
| The privacy page names Netlify for website hosting and the newsletter database. The deployment workflow builds and deploys the site on Cloudflare Workers, and `site/app/api/newsletter/route.ts` writes subscriber records to a Cloudflare KV binding when configured. | Verify the production provider, KV configuration, actual stored fields, retention/deletion path, and any other processing providers. Reconcile the notices only after confirming production behavior. |
| The website is built with a Railway control-plane endpoint and can load hosted Umami analytics when its website ID is configured. | Include the live control-plane and optional analytics behavior in the data map and verify the active production configuration. |
| The Terms identify “The NAEOS Foundation” in a liability clause, while `legal/README.md` lists the exact contracting entity and jurisdiction as an open legal question. | Confirm the legal contracting party, authority, and jurisdiction with the founder and counsel before offering paid services. |
| The public Terms cover the website, documentation, and open-source software, and mention newsletter subscriptions; they do not define separate Cloud, enterprise, assessment, pilot, data-processing, or support terms. | Keep paid offerings out of scope of these terms until counsel approves the relevant service terms and the actual offer matches them. |

Before a paid offer is announced, compare the pages with its data map and actual
service behavior, then obtain counsel review. Do not edit or extend public legal
terms based on this checklist alone.
