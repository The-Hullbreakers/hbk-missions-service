# CHG-0055 — Create local season preview scenarios: rollout

status: verified_local
environment: DynamoDB Local on HTTP loopback only

1. Dry run selected scenario and inspect target, row counts and active
   identity count; no writes by default.
2. Focused unit, local-resource integration, security and real
   service checks passed. Fresh-start and mid-season were applied
   to separate fixed preview tables by approved integration tests;
   off-season was explicitly applied by lead. No table was deleted.
3. Select that table in the local missions service environment;
   inspect off-season/fresh-start by selecting their own tables once
   loaded. Do not alter the web app's credentials or user identities.
4. Restore the previous service table setting to roll back the view.
   No table deletion or reset is part of rollout.
5. Verify original `hbk-missions-local` season/offers/definitions
   counts and exact key set are unchanged. Report generated preview
   counts without exposing individual member data.

If the local identity table changes between preview and apply, stop
and re-review the plan. If a batch is interrupted, rerun only against
matching scenario-owned rows; a divergent row blocks further writes.
Never drop or truncate a table as a recovery shortcut. The
fresh-start and mid-season preview tables remain in place for user
review after the tests; nothing in rollout deletes them.
