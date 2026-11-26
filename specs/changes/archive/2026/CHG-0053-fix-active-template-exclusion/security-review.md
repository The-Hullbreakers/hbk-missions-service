---
change_id: CHG-0053
artifact: security-review
status: approved-local-only
security_role: resource-server
updated: 2026-10-26
---

# Security Review — CHG-0053

## Review Scope

The service remains an OIDC-protected resource server. The changed public
paths are offer acceptance and contract leave; existing session-derived
identity, role validation, cookie handling, and token verification are
unchanged. Reviewed the affected domain/application policies, DynamoDB
acceptance and lifecycle writes, composition wiring, and real HTTP journeys.
No new dependency, secret, IAM permission, event field, or public response
shape is introduced by CHG-0053.

## Trust Boundaries

1. **Identity derivation (unchanged):** acceptance and leave resolve
   `player_id`/`team_ids` exclusively from the authenticated session;
   request-body identity fields stay ignored (`REQ-MB-011`). This change
   introduces no new identity input, endpoint, or role.
2. **Write integrity (acceptance transaction):** the reservation `Put`s
   join the existing conditional transaction. Fail-closed rule: any
   reservation conflict must resolve as
   `MISSION_DEFINITION_ALREADY_ACTIVE` — never as a silent pass or a
   limit error that invites retry confusion.
3. **Data exposure:** reservation items carry only internal ids
   (player, definition_code, contract_id); they are never projected to
   public responses or events. Error semantics are unchanged —
   `MISSION_DEFINITION_ALREADY_ACTIVE` is already a public `422` code and
   reveals nothing a board listing would not.
4. **Authorization surface:** a team accept naming a bound teammate now
   rejects instead of minting a duplicate contract — strictly tighter,
   no privilege change. Release paths delete only the owning contract's
   reservation (`contract_id` condition), so no member can free another's
   guard.
5. **Migration/backfill:** a naive conditional-put backfill is neither
   race-safe nor idempotent (see `rollout.md`). Production rollout of the
   strict invariant requires a separately approved, fenced reconciliation;
   this change performs no shared/live-data writes.
6. **Write integrity (departure):** `SaveDeparture`'s contract `Put`
   carries an optimistic-concurrency condition (`attribute_exists(PK)`,
   persisted `#state = :state`, and the persisted departed set exactly
   equal to the members departed before this leave — `size` plus a
   `contains` per prior member; a missing `Departed` attribute is
   tolerated only when that set is empty). A stale snapshot write
   cancels the transaction — it can no longer resurrect a departed
   member as an active participant (see FIND-0053-01 below). The same
   departed-set fence now covers every contract write path (Gate 1
   expansion, implemented T-15): `Save`'s non-terminal put and terminal
   transaction put require the stored departed set to equal the
   candidate's `Departures()` — an absent row passes only when the
   expected set is empty, preserving the legacy create path —
   `SaveApproved`'s put additionally requires `attribute_exists(PK)`,
   and `SaveDeparture` takes the caller's `previousState` so a
   last-member exit (ABANDONED with an emptied roster) commits through
   the same fenced write that releases the leaver's own guards. A stale
   snapshot cancels atomically: no contract overwrite, no credit items
   (see FIND-0053-02 below). Residual: the fence covers departed-set
   divergence only — other concurrent contract-state mutations on the
   same row (phases, confirmations, invitations) carry no optimistic
   versioning; that broader versioning is deliberately out of scope for
   this change. CHG-0051's planned adjudication-lock fence composes by
   appending conditions, not replacing these.

## Findings

| Severity | Finding | Status |
| --- | --- | --- |
| High (candidate) | FIND-0053-01 — concurrent departures lost update: two members leaving on separate snapshots of one contract produce two `SaveDeparture` calls; the later, stale write's unconditional contract `Put` resurrects the first leaver as an active participant, restoring their evidence/submit surface and re-hiding the template while their reservation and `PLAYER_ACTIVE` slot were already released. Reproduced on DynamoDB Local by `TestContractStore_ConcurrentDeparturesDoNotRestoreFormerMember` — stale write returned nil (`evidence/t13-concurrent-departure-red.txt`). Proposed fix direction (implemented): a conditional `Put` on the contract item requiring the item to exist, the pre-leave state to match the persisted state, and the persisted departed set to equal exactly the members departed before this leave (`size` + `contains` per prior member; missing attribute allowed only when the expected prior set is empty); this yields an optimistic conflict for parallel departures while preserving the independent member-scoped reservation release. | resolved — same-set fence verified against the missing-attribute and same-count-different-set edges, `evidence/t13-departure-green.txt`, `evidence/t13-departure-set-green.txt` |
| High (candidate) | FIND-0053-02 — stale ordinary `Save` resurrects a departed member: `ContractStore.Save` wrote a non-terminal contract with an unconditional `PutItem` (`internal/missions/infrastructure/dynamodb/contract_store.go`, non-terminal branch). A snapshot fetched before a leave — e.g. an in-flight confirmation, invitation, or phase flow — overwrote the committed departure, returning the leaver to the active roster while their reservation and `PLAYER_ACTIVE` slot were already released (an active member with no guards). Reproduced on DynamoDB Local by `TestContractStore_StaleSaveDoesNotRestoreDepartedMember` — stale `Save` returned nil (`evidence/t13-stale-save-red.txt`). Fixed under the Gate 1 expansion: the shared `departedSetMatch` condition now fences `Save` (both branches) and `SaveApproved` (plus `attribute_exists(PK)`); stale `Save` and stale `SaveApproved` conflict at contract-put index 0 with no partial write — verified by `StaleSaveDoesNotRestoreDepartedMember`, `StaleSaveApprovedDoesNotRestoreOrCredit`, and the full standalone suite (`evidence/t15-persistence-green.txt`, `evidence/t16-infrastructure.txt`). | resolved — departed-set fence on all contract writes, `evidence/t15-persistence-green.txt`, `evidence/t16-infrastructure.txt` |

## Reviewer Role And Residual Risk

- **Role:** implementer-side trust-boundary audit; the user approved the
  security gate for local code only on 2026-10-26. This is not approval
  of a production deployment or of a migration on shared data.
- **Residual risks (recorded; owner decision pending):** legacy contracts without
  reservation items remain outside the strict uniqueness invariant until
  a separately approved race-safe reconciliation (`rollout.md`,
  boundary 5); the known out-of-scope `submit` → `PENDING_VALIDATION`
  `PLAYER_ACTIVE` slot defect is unchanged (recorded in `change.md`);
  the departed-set fence covers membership divergence only — other
  concurrent contract-state mutations on the same row carry no
  optimistic versioning.
- **Automated evidence:** `go vet ./...`, all unit/integration suites and
  race detector pass in the user-authorized isolated environment
  (`evidence/t17-isolated-validator.txt`); three public API journeys
  pass three repetitions (`evidence/t16-critical-e2e.txt`). Manual
  changed-scope scan found no embedded token, private key, or common
  credential prefix; no new dependencies or IAM changes. `gitleaks`,
  `govulncheck`, and `gosec` are not installed: no scanner result is
  claimed, and dependency/security scanning remains an operational
  follow-up for the CI owner. The validator lists two unrelated,
  pre-existing gofmt violations outside CHG-0053.
- **Production gate:** legacy reservation-less contracts require a
  separately approved, race-safe reconciliation before enabling the
  strict invariant in a shared environment. No shared-data migration or
  production deployment occurred in this change.

No Critical/High findings may remain open at consolidation.

## Human Decision

Approved by the user for local code on 2026-10-26. FIND-0053-01/02
are resolved with RED/GREEN evidence. The change consolidated and
archived on 2026-10-26 with this verdict; the missing external
security scanners and any reconciliation of legacy reservation-less contracts
remain required before a future shared-environment deployment.
