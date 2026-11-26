---
change_id: CHG-0051
artifact: critical-e2e
status: planned
updated: 2026-10-20
---

# Critical E2E — CHG-0051

## Journey (critical)

```gherkin
Scenario: Finalizing an official season emits durable placement facts
  Given an official season FREEZING with ranked standings
  When the season is finalized
  Then placement-awarded intents exist for every ranked entry
  And the publisher emits each fact once with its stable event_id
```

## Planned Evidence

- `TEST-E2E-AWD-001`: local stack — finalize through the real path, inspect
  emitted envelope on local SNS/SQS wiring. Deferred: no local SNS/SQS
  emulator is wired in this stack; envelope publication is covered by the
  manifest-publisher unit and integration suites instead.
- `TEST-E2E-AWD-002` (executed below) covers REQ-SCC-018 (+004) at the
  real-adapter boundary: 200 ranked entries finalize -> manifest -> 200
  publishable facts.

## Record

| Run | Command | Result | Date |
| --- | --- | --- | --- |
| TEST-E2E-AWD-002 | `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -v -run TestSeasonAdjudicationE2E ./tests/integration/infrastructure/dynamodb/` | PASS — official FREEZING season with 200 merit credits finalizes through the real two-phase pipeline (lock, consistent snapshot, stage, verify, constant-size commit); season persists FINALIZED; exactly one PENDING manifest declares `FactCount=200` with a non-empty `FactHash`; `ListRemaining` yields 200 intents, all with unique non-empty event IDs | 2026-11-16 |
