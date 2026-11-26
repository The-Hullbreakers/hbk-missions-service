---
change_id: CHG-0052
artifact: tasks
status: in-progress
updated: 2026-10-20
---

# Tasks — CHG-0052

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-052-001 | Application RED: publish/accept handlers without capacity; no `ErrOfferCapacityExhausted` mapping | application | done | application-red |
| T-052-002 | Domain: drop `Capacity` from offer/definition, capacity branch in `DecideAcceptance`, `ReasonOfferCapacityExhausted`, `ErrOfferCapacityExhausted` | domain | done | domain |
| T-052-003 | Application GREEN: publish handlers stop taking `capacity` | application | done | application |
| T-052-004 | Interface: remove `capacity` from create/publish/offer payloads + projections; remove `OFFER_CAPACITY_EXHAUSTED` mapping | interface | done | boundary |
| T-052-005 | Infrastructure: `CommitAcceptance` loses `remaining_capacity`; offer store drops `Capacity`/`remaining_capacity` write+read | infrastructure | done | infrastructure |
| T-052-006 | Seeds: drop `Capacity`, bump both catalogs to version `4`, render+dry-run+apply on `hbk-missions-local` | infrastructure | done | infrastructure |
| T-052-007 | Gates: full suite, coverage, mutation spot-check on acceptance decision, docs, `validate-sdd-change.sh`, consolidate + archive | verification | done | boundary |
