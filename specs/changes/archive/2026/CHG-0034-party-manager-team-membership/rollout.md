---
change_id: CHG-0034
artifact: rollout
updated: 2026-10-15
---

# Rollout — Party Manager Team Membership

## Activation

- New env var `PARTY_MANAGER_URL`. Unset/empty keeps `DenyAllReader`:
  TEAM/FLEX-as-team acceptance stays rejected — identical to current behavior.
- Setting the variable enables real roster validation on next deploy; no
  migration, no data change.

## Ordering

1. Deploy with `PARTY_MANAGER_URL` unset (safe no-op release).
2. Confirm Party Manager reachability from the service network.
3. Set `PARTY_MANAGER_URL`, redeploy.
4. Smoke: accept a TEAM offer with a real teammate → 201; with a non-member → 422.

## Failure Signals

- `MEMBERSHIP_UNAVAILABLE` (503) spike → Party Manager down or timeout budget
  too tight; contract creation for team offers pauses (intended fail-closed).

## Rollback

Unset `PARTY_MANAGER_URL` and redeploy — composition reverts to `DenyAllReader`.
No partial state to clean: failed acceptances write nothing.
