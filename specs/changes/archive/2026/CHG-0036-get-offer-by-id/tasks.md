---
change_id: CHG-0036
artifact: tasks
status: implemented
updated: 2026-10-16
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-036-001 | `ErrOfferNotAvailable` sentinel + `OfferStore.Get` mapping + `OFFER_NOT_AVAILABLE` code | persistence-adapter | done | infrastructure |
| T-036-002 | GetOfferHandler + `GET /offers/{offer_id}` route + tests | rest-endpoint | done | application→interface |

## Requirement to Test to Production Map

- **REQ-MB-016** → handler unit + HTTP tests → `application/` (new `get_offer.go`), `interface/http/` (new `get_offer.go`), `composition/server.go`.

## Gates

Standard per risk level.
