---
change_id: CHG-0039
artifact: delta
status: implemented
updated: 2026-10-14
---

# Delta

## What Is Changing

### Mission Contract Fulfillment

- The active-contract limit is read from a persisted policy at acceptance; admins update it via PUT. Composition no longer hardcodes the default.

## Requirements

### REQ-MCF-023 — Manage active-contract limit (added)

An admin updates the global (and optional per-category) active-contract limit; the policy applies to subsequent acceptances without restart.

## BDD Scenarios

### SCN-MCF-POLICY-UPDATE — Admin sets limit

Given the global limit is 5, when an admin PUTs limit 3, then subsequent acceptances enforce 3.

### SCN-MCF-POLICY-DENIED — Member cannot update

Given a non-admin caller, when they PUT the policy, then it returns 403 FORBIDDEN.

## Error Vocabulary

- Reuses `FORBIDDEN`; `INVALID_POLICY` — 422 (new) for non-positive limits.
