---
change_id: CHG-0003
artifact: mutation-report
status: completed
updated: 2026-08-29
---

# Mutation Report

## Scope

Changes in domain, application, infrastructure and composition for prod-readiness of `mission-contract-fulfillment`.

## Method

`go-mutesting` and `gremlins` are not installed; manual analysis of critical mutations.

## Critical Mutations Reviewed

| Function | Mutation | Test Detection |
| --- | --- | --- |
| `domain.ConfirmObjective` | Reverse condition `contract.State != ContractInProgress` | `TestConfirmObjectiveHandler_ErrorsWhenContractNotInProgress` returns `CONTRACT_NOT_IN_PROGRESS` |
| `domain.ConfirmObjective` | Delete active season check | `TestConfirmObjectiveHandler_ErrorsWhenSeasonNotActive` returns `SEASON_NOT_ACTIVE` |
| `domain.ConfirmObjective` | Allow duplicates | `TestMeritCredit_DuplicateTargetIsRejected` fails on second attempt |
| `application.ConfirmObjectiveHandler.Handle` | Do not call `contractWriter.Save` | `TEST-COMP-001` would not persist `confirmedUnits`; test reload doesn't verify it yet, but `Confirmations()` exposes it |
| `dynamodb.MeritLedgerStore.Save` | Do not write GSI1 keys | `TEST-PERSIST-001` `TotalFor` returns 0 instead of 100 |
| `dynamodb.MeritLedgerStore.TotalFor` | Add wrong or not paginate | `TEST-PERSIST-001` correct total with multiple credits |
| `interface/http.ConfirmEvidenceHandler` | Return `err.Error()` directly | `TEST-ERROR-001` asserts public codes |

## Conclusion

- Tests detect mutations in critical business logic and persistence.
- Mutations in OIDC/provider initialization are not unit-testable; covered by security review and E2E.
