---
change_id: CHG-0033
title: Get contract detail for adjudicators
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0033-contract-detail/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
context_scope: existing
created: 2026-10-13
updated: 2026-10-13
---

# Change — Contract Detail for Adjudicators

## Intent

Reviewers picking an entry from `GET /review-queue` currently adjudicate
blind — they cannot inspect the contract's frozen terms, confirmed units,
staged credits, or evidence references without a dedicated read. This
change adds a reviewer-only contract-detail endpoint so adjudication can
be informed.

## Scope

- `GET /api/v1/missions/contracts/{contract_id}` for `admin`/`moderator`.
- Full detail projection: identity, offer, frozen `Terms`, state, roster,
  departures, confirmed units, and staged pending credits including each
  credit's `media_uri` (opened via the existing reviewer read grant).
- Members and outsiders receive `403 FORBIDDEN` — member-facing contract
  visibility stays roster-scoped through `GET /contracts` (REQ-MCF-006);
  exposing teammates' evidence references would leak reviewer-only data.
- Unknown contract ids resolve as `422 CONTRACT_NOT_FOUND` (repository
  convention).

## Out Of Scope

- Member-facing detail, `GET /definitions/{code}`, `GET /offers/{id}`.
- New persistence — `ContractReader.Get` and its adapter already exist.
- Pagination, filtering, or history views.

## Tenancy Note

Deployment/storage boundary only; `CallerIdentity` carries no tenant
claim, consistent with REQ-MCF-020 adjudication.
