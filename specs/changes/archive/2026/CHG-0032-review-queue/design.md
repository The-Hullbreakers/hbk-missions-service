---
change_id: CHG-0032
artifact: design
status: final
updated: 2026-10-12
---

# Design — Review Queue for Pending Contracts

## Use Case

`ListPendingReviewHandler` (application query) → `identity` in,
`[]domain.MissionContract` out.

- `isReviewer(identity)` guards the use case (`403`-mapped
  `ErrCallerNotAuthorized`) — defense in depth on top of the route
  middleware.
- Calls the `PendingReviewLister` port: one CQRS operation
  `ListPendingReview() ([]domain.MissionContract, error)` returning
  contracts in `PENDING_VALIDATION`.

## Persistence

`ContractStore.ListPendingReview()` — Query `PK = CONTRACT` and filter
`State == PENDING_VALIDATION`, same in-memory filtering pattern as
`ListActiveByPlayer`/`ListByPlayer`. No new GSI, no item-shape change.

## Boundary

`GET /api/v1/missions/review-queue`

- Route behind `auth.Authenticate(adminOrModerator(handler))` —
  consistent with definitions/offers/season admin routes.
- `ReviewQueueHandler(app)` adapter: `401` unauthenticated (middleware),
  `403` non-reviewer, `200` with entries.
- Entry DTO: `contract_id`, `offer_id`, `definition_code`, `title`,
  `state`, `roster`, `season_id`, `team_id`, `confirmed_count`,
  `pending_credits` — mirrors the member summary plus review-relevant
  counts; no credit detail in the list.

## Sequence

1. Application unit RED: reviewer sees pending only; member denied;
   empty → `[]`.
2. Infrastructure RED: `ListPendingReview` vs DynamoDB local over mixed
   states.
3. Boundary RED: route 403 for member session, `200` + shape for
   reviewer; composition wiring.
