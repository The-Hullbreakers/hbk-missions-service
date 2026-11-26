---
change_id: CHG-0033
artifact: design
status: final
updated: 2026-10-13
---

# Design — Contract Detail for Adjudicators

## Use Case

`GetContractHandler` (application query) → `(identity, contractID)` in,
`*domain.MissionContract` out.

- Reuses the existing `ContractReader` port (`Get`) — no new port, no new
  adapter; `ContractStore` already satisfies it.
- Policy: `isReviewer(identity)` before the store read — same guard as
  `ListPendingReviewHandler` and adjudication handlers. Members get
  `ErrCallerNotAuthorized` (`403`), which also prevents existence
  probing by non-reviewers.

## Boundary

`ContractDetailHandler` in `interface/http/contract_detail.go` —
`GET /api/v1/missions/contracts/{contract_id}` behind
`auth.Authenticate(adminOrModerator(...))`.

- `contractDetail` DTO: summary fields plus `terms` (frozen objectives /
  units / completion reward), `departures`, `confirmations`, and
  `pending_credits[]` entries with `objective_id`, `unit_id`,
  `recipient_id`, `amount`, `media_uri` (omitted when empty).
- `ErrContractNil` → `422 CONTRACT_NOT_FOUND`; `ErrCallerNotAuthorized` →
  `403 FORBIDDEN`; other errors → `500 INTERNAL_ERROR`.

## Sequence

1. Application: `GetContractHandler` + unit tests (reviewer ok, member
   denied without store call, missing contract propagates `ErrContractNil`).
2. Boundary: handler + DTO + http tests (reviewer 200 detail incl. staged
   credit + media_uri, member 403, unknown id 422).
3. Composition: wire `reviewDetailApp` on `contractStore`, mount route,
   update `API-CONTRACT.md` / `API-GUIDE-FRONTEND.md`.

## Notes

- Read-only: no domain change, no persistence work.
- The detail intentionally exposes `media_uri` references — bytes still
  require the reviewer-only `media-url` read grant.
