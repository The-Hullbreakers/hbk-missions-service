---
change_id: CHG-0021
artifact: verification
updated: 2026-10-08
---

# Verification — CHG-0021

## Status

`verified` — all L3 gates green.

## Evidence Plan

| Check | Command / Method | Result |
| --- | --- | --- |
| Unit suites | `go test ./tests/unit/...` | pass — 5 packages |
| HTTP integration | `go test ./tests/integration/http/...` | pass |
| Infrastructure integration | `go test ./tests/integration/infrastructure/...` | pass |
| Composition + E2E | `go test ./tests/integration/composition ./tests/integration/e2e` | pass — incl. `TestSoloContractJourney_FrozenTermsEndToEnd` |
| Race | `go test -race -p=1 ./tests/...` | pass |
| SDD validation | `tools/validate-sdd-change.sh` | pass |
| Coverage gate | validate script coverage section | pass — domain 90.5%, application 92.3%, dynamodb 91.4% |
| Mutation gate | manual mutants (gremlins not installed) | pass — 4/4 killed: `Start` guard, `Complete` terms gate, `Confirm` terms lookup, accept terms freeze |
| Critical E2E | `TestSoloContractJourney_FrozenTermsEndToEnd` | pass — accept→start→confirm(125 w/ bonus)→foreign-objective 422→complete→terminal repeat 422 |

## Reproduction Evidence (pre-fix)

- No public transition reaches `IN_PROGRESS`; tests start contracts manually.
- 999,999 merits credited via unrelated same-category objective.
- Injected ledger failure → duplicate response on retry, standing stays 0.
- `Bonus`/`UnitLimit` absent after persistence round-trip.
- Colon-containing contract id fails `ConfirmationID` parse.
- Re-acceptance after abandon overwrites prior attempt.
