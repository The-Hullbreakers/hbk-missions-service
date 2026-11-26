---
change_id: CHG-0021
artifact: security-review
updated: 2026-10-08
---

# Security Review — CHG-0021

## Trust Boundaries

- Lifecycle endpoints mutate contract state: only roster members (or privileged
  roles per CHG-0020 policy) may transition a contract.
- Frozen `ContractTerms` prevents reward manipulation via definition edits or
  foreign objective ids.
- Atomic write prevents merit loss/duplication; dedup marker binds credit to
  `ConfirmationID`.

## Findings Addressed

- Arbitrary merit amounts through unbound objectives (BUG-0021-02).
- Merit loss on partial write (BUG-0021-03).
- History overwrite via contract-id reuse (BUG-0021-06).

## Evidence

- Non-roster transition attempts → `403`: `TestContractLifecycle_StartEndpointRejectsOutsider` (http integration), `Test*Lifecycle*` outsider cases (application unit).
- Wrong-objective evidence → `OBJECTIVE_NOT_IN_CONTRACT`: `TestConfirmObjective_RejectsObjectiveOutsideFrozenTerms` + E2E `not-in-terms` step.
- Injected transaction failure → no partial state; retry → single credit: `TestConfirmObjective_RetryAfterFailureIsIdempotent`, `TestContractStore_SaveConfirmedCommitsContractAndCredit` (conditional put maps `TransactionCanceledException/ConditionalCheckFailed` → `ErrDuplicateConfirmation`).
- Lifecycle authorization is tenant + roster-scoped via `CallerIdentity`; no body-supplied identity trusted (CHG-0020 policy preserved).
- No Critical/High residual findings. Role: implementer self-review with boundary tests.
