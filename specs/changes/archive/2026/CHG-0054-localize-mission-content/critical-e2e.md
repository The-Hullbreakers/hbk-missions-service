---
change_id: CHG-0054
artifact: critical-e2e
status: completed
updated: 2026-11-16
---

# Critical E2E — CHG-0054

## Journey (critical)

```gherkin
Scenario: Accepted bilingual offer freezes both locales
  Given a bilingual published offer readable in English
  When a member accepts it
  Then the contract serves frozen English and Spanish content
  And a later translation edit cannot rewrite the accepted snapshot
```

## Record

| Run | Command | Result | Date |
| --- | --- | --- | --- |
| TEST-E2E-LANG-001 | `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -v -run TestLocalizedAcceptanceE2E ./tests/integration/composition/` | PASS — `GET /offers/{id}?lang=en` serves English labeled `en`; member accepts → contract detail serves `en` then `es` with the frozen titles; after the stored definition and offer receive a translation edit the contract still serves the original English title | 2026-11-16 |
