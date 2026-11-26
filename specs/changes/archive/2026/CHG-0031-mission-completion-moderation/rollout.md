---
change_id: CHG-0031
artifact: rollout
status: draft
updated: 2026-10-09
---

# Rollout — Mission Completion Moderation

## Compatibility

- **Breaking API change:** `POST .../complete` is removed and replaced by
  `POST .../submit` (roster) + `POST .../approve|reject` (reviewer).
  Frontend and API consumers must migrate before or with deploy.
- **`POST .../evidence` contract:** still `201` with the computed credit
  shape, but the credit is *pending* — the standing only moves after
  approval. Consumers polling standings after evidence must now wait for
  adjudication.
- **New state value:** `PENDING_VALIDATION` appears in contract payloads
  and `status` filters; clients enumerating states must tolerate it.

## Data

- `pending_credits` attribute on the contract item is additive — existing
  items read as empty; no migration.
- In-flight `IN_PROGRESS` contracts at deploy time keep working: their
  next `submit` enters the review flow. There is no `PENDING_VALIDATION`
  data before this change ships.

## Rollback

- Reverting the deploy restores `complete` semantics, but contracts in
  `PENDING_VALIDATION` would be invisible to the old transition set.
  Before rollback, drain the queue: enumerate `PENDING_VALIDATION`
  contracts and approve/reject them under the new build first.
- `pending_credits` attributes are ignored by the old readers — dormant
  data, no corruption.

## Observability

- Log adjudication decisions (contract id, reviewer player id, decision)
  at application level — no PII beyond ids.
- Track pending-queue depth and review latency; a permanently pending
  contract silently withholds earned merits.

## Verification On Deploy

- Smoke: submit → approve happy path on a seeded contract; confirm the
  standing only increases after approval.
- Confirm `complete` returns `404` and `evidence` on a pending contract
  returns `422 CONTRACT_NOT_IN_PROGRESS`.
