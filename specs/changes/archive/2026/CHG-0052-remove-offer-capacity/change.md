---
change_id: CHG-0052
title: Remove the offer capacity slot limit entirely
status: proposed
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/active/CHG-0052-remove-offer-capacity/
execution_mode: sequential
depends_on: []
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_SDD_COVERAGE_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
context_scope: existing
created: 2026-10-20
updated: 2026-10-20
---

# Change — Remove Offer Capacity

## Intent

Offers are unlimited: a published mission accepts contracts until it is
unpublished or its season leaves `ACTIVE`. The `Capacity`/`remaining_capacity`
slot counter — publish-time seeding, the conditional decrement inside
`CommitAcceptance`, and the `OFFER_CAPACITY_EXHAUSTED` rejection — is
removed completely. The per-player active-mission limit (`PLAYER_ACTIVE`)
is a different feature and is preserved untouched.

## Business Value

- Product decision: mission slots add friction without payoff; every
  eligible member may take a mission while it is published.
- Fewer rejection paths: acceptance can only fail for limit, season,
  window, repeatability, or template-already-active reasons.
- Simpler transaction: `CommitAcceptance` writes the contract and the
  player active-set entry; one conditional branch disappears.

## Scope

- Domain: drop `MissionOffer.Capacity`, `MissionDefinition.Capacity`,
  `ReasonOfferCapacityExhausted`, `ErrOfferCapacityExhausted`, and the
  capacity branch in `DecideAcceptance`.
- Application: remove the `ReasonOfferCapacityExhausted` mapping in
  `AcceptMissionHandler`; publish handlers stop taking `capacity`.
- Interface: remove `capacity` from `POST /definitions`,
  `POST /definitions/{code}/publish`, `POST /offers`, and from offer /
  definition / contract projections; remove the
  `OFFER_CAPACITY_EXHAUSTED` error mapping. Breaking contract change —
  deploy with the webapp companion.
- Infrastructure: `CommitAcceptance` transaction loses the
  `remaining_capacity` update; offer store stops writing and querying
  `Capacity`/`remaining_capacity`. Existing DynamoDB items keep an
  inert leftover attribute — no data migration.
- Seeds: drop `Capacity` from `pvp-missions.json` and the racing
  generator; bump both catalogs to version `4`; redeploy
  `hbk-missions-local`.

## Non-Goals

- The per-player `PLAYER_ACTIVE` limit stays.
- No DynamoDB migration to scrub the orphaned attributes.
- No change to `AVAILABLE` filter semantics beyond "no capacity check":
  `AVAILABLE` still means offers listed for the resolved season scope
  minus active-template exclusions.
