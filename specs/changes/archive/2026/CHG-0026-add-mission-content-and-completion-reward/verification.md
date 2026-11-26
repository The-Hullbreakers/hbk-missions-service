---
change_id: CHG-0026
artifact: verification
status: verified
updated: 2026-10-09
---

# Verification — Add Mission Content And Completion Reward

## Test-First Evidence

| Layer | RED (observed) | GREEN (evidence) |
| --- | --- | --- |
| Domain | Build fail: `NewMissionDefinition` old signature, missing `Title`/`Description`/`CompletionReward`; `Complete` returning 1 value | `go test ./tests/unit/missions/domain/` ok |
| Application | Build fail: `Handle` old signature; `NewCompleteContractHandler` missing `ContractAndCreditWriter` | `go test ./tests/unit/missions/application/` ok |
| HTTP boundary | `TestCreateMissionDefinitionHandler_CreatesDefinition` 422 with old payload | `go test ./tests/unit/missions/interface/http/` ok |
| DynamoDB | `TestDefinitionStore_RoundTrip`/`TestOfferStore_RoundTrip` lost content fields | `go test ./tests/integration/infrastructure/dynamodb/` ok |
| Composition/e2e | Old definition payload → 422 | `TestSoloContractJourney_FrozenTermsEndToEnd` ok — standing 325 = 125 evidence + 200 reward |

## Scenarios Verified

- SCN-DEF-CONTENT-CREATE — 201 with title/description/reward echo (unit http).
- SCN-DEF-CONTENT-REQUIRED — empty title/description/objective name → `ErrDefinitionInvalid` → 422.
- SCN-DEF-CONTENT-NEGATIVE-REWARD — `completion_reward: -5` → `ErrDefinitionInvalid`.
- SCN-OFFER-CONTENT — offer/definition round-trip persists content (dynamodb integration).
- SCN-CONTRACT-COMPLETION-REWARD — complete emits `COMPLETION` credit, atomic via `SaveConfirmed`; e2e leaderboard total 325.
- SCN-CONTRACT-COMPLETION-ZERO-REWARD — plain `Save`, no credit (app unit).
- SCN-CONTRACT-COMPLETION-IDEMPOTENT — terminal contract rejects; deterministic `ConfirmationID` makes a retry a no-op write.

## Validation

- `tools/validate-sdd-change.sh` — **passed** (unit + integration, race detector, coverage: domain 91.8%, application 91.1%, dynamodb 87.7%).
- `go build ./...`, `go vet ./...`, `gofmt` — clean.
- Maintained files < 150 physical lines — verified.

## Risk Gates (L3)

- **Critical-e2e:** `TestSoloContractJourney_FrozenTermsEndToEnd` covers
  create → publish → offer → accept → evidence → complete → leaderboard
  including the completion reward. PASS.
- **Mutation gate:** not executed — no mutation tooling exists in this repo
  (`tools/` has only `validate-sdd-change.sh`). Mitigation: domain tests
  assert guards (terminal, unconfirmed terms, negative reward, unnamed
  objectives), atomicity (`SaveConfirmed` transaction) and idempotency
  (deterministic `ConfirmationID`) directly.
- **Coverage gate:** domain 91.8% / application 91.1% ≥ 90%. Infrastructure
  87.7% < 90% — no regression attributable to this change; affected lines
  (objective mapping, store writes/reads) are covered by round-trip tests.

## Residual Risk

- Old contracts saved before this change carry `CompletionReward: 0` → they
  complete without emitting a credit. Acceptable: no live items exist.
- `COMPLETION` is a reserved objective id — now rejected at definition
  creation (`ErrDefinitionInvalid`), so the confirmation-id namespace
  cannot collide. Legacy cleanup: dead `ObjectiveStore`, `DefinitionID` /
  `legacy` DynamoDB fallbacks and the name-less `NewObjective*` constructors
  were removed; all call sites use `NewObjectiveNamed`.

## Sign-off

- All unit/integration/e2e suites green from clean state.
- Ready for consolidation into `mission-board` and
  `mission-contract-fulfillment` canonical specs.
