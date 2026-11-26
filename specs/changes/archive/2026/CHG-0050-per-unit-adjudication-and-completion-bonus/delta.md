---
change_id: CHG-0050
artifact: delta
status: proposed
updated: 2026-10-20
---

# Delta — Per-Unit Adjudication And Completion Bonus

## What Is Changing

### Mission Contract Fulfillment

- Reviewer `approve` adjudicates **per staged unit**: an optional
  `rejected_units` list names `{objective_id, unit_id}` pairs that fail
  review; every other staged unit's `MeritCredit` commits. All staged
  units pass when the list is empty or absent.
- A rejected unit is recorded on the contract (`rejected_units`) for
  audit and dispute context; its staged credit is discarded.
- The `COMPLETION` credit now requires **full fulfillment**: every
  mandatory objective's confirmed-and-not-rejected unit count reaches its
  `unit_limit` (an unlimited objective needs at least one unit),
  evaluated after rejected units are applied.
- Contract detail (reviewer and member projections) exposes
  `rejected_units` so a member sees which units failed and a dispute can
  reference them.
- New public error `UNIT_NOT_STAGED` (`422`): a verdict naming a unit
  that was never staged rejects the whole `approve` call.

### Mission Board

- No canonical requirement changes. The catalog (seeds) is remodeled to
  the unit model — one countable objective per mission plus a positive
  `completion_reward` — and rule lists are rewritten to the essential
  verifiable set. `merit_summary` already derives correctly from
  objectives + `completion_reward`.

### Seasonal Category Competition

- Unchanged — standings stay a pure ledger projection; credits simply
  arrive per approved unit.

## MODIFIED

### REQ-MCF-020 — Adjudicate a pending contract

`POST /api/v1/missions/contracts/{contract_id}/approve` accepts an
optional body `{"rejected_units": [{"objective_id", "unit_id"}]}`. The
reviewer adjudicates each staged unit: units not listed commit their
frozen credit; listed units are marked rejected on the contract and pay
nothing. An unknown `rejected_units` entry resolves `422
UNIT_NOT_STAGED` and writes nothing. `approve` still requires
`admin`/`moderator` (`403 FORBIDDEN`) and `PENDING_VALIDATION` (`422
CONTRACT_NOT_PENDING_VALIDATION`), returns `200` with the written
credits, transitions `COMPLETED`, and stays idempotent on retry.
`reject` is unchanged: terminal `DISQUALIFIED`, zero credits.

### REQ-MCF-012 — Completion grants the frozen reward

The `COMPLETION` credit is emitted at `approve` only when every
mandatory objective across all phases is **fully satisfied**: its
confirmed-and-not-rejected unit count reaches `unit_limit` (or is at
least one when `unit_limit` is `0`/unlimited). The check runs after
rejected units are dropped, so a rejected unit always blocks the bonus.
Identity, single-credit, and `completion_reward: 0` semantics are
unchanged.

### REQ-MCF-022 — Read own contract detail

The contract detail gains `rejected_units` — an ordered list of
`{objective_id, unit_id}` recorded by the adjudicating approve. Reviewers
see the full list; the member projection lists the units staged under
the caller's own recipient, consistent with the `pending_credits`
scoping rule.

### REQ-MCF-004 — Prevent duplicates and basic farming

Clarified: a unit rejected at adjudication is released from the
contract's confirmed set before fulfillment is evaluated. Duplicate
detection by `ConfirmationID` and `unit_limit` enforcement are
unchanged — a terminal contract cannot stage further confirmations.

## ADDED

None — the per-unit reward model reuses the existing countable-unit
objective machinery (`merit_per_unit`, `bonus`, `unit_limit`); the change
is in adjudication semantics, the completion predicate, and catalog
content.

## BDD Scenarios

### SCN-MCF-APPROVE-UNITS — Approval commits the approved subset

Given a `PENDING_VALIDATION` contract with three staged units and a
verdict rejecting `unit-2`, when a moderator approves, then exactly the
credits for `unit-1` and `unit-3` are written, `unit-2` appears in the
contract's `rejected_units`, and the contract is `COMPLETED`.

### SCN-MCF-APPROVE-UNITS-DEFAULT — Empty verdict approves all

Given a pending contract with staged units and an approve call without a
body (or an empty `rejected_units`), when a moderator approves, then all
staged credits are written — the legacy approve-all behavior.

### SCN-MCF-APPROVE-UNKNOWN-UNIT — Unknown rejected unit fails the call

Given a pending contract and a verdict naming a unit never staged, when
a moderator approves, then the response is `422 UNIT_NOT_STAGED`, the
contract stays `PENDING_VALIDATION`, and no credit is written.

### SCN-MCF-REJECT-BLOCKS-BONUS — Rejected unit forfeits the bonus

Given a pending contract whose objective requires `unit_limit: 10` and
only nine staged units remain after the verdict, when a moderator
approves, then nine unit credits are written and no `COMPLETION` credit
exists.

### SCN-MCF-FULL-BONUS — Full satisfaction emits the bonus

Given a pending contract whose `unit_limit: 10` objective has all ten
units staged and a verdict rejecting none, when a moderator approves,
then ten unit credits plus one `COMPLETION` credit are written
atomically.

### SCN-MCF-PARTIAL-STAGE-NO-BONUS — Unstaged units also block the bonus

Given a pending contract whose objective requires ten units but only
seven were ever staged, when a moderator approves with no rejects, then
seven credits are written and no `COMPLETION` credit exists.

### SCN-MCF-REJECTED-VISIBLE — Rejected units are listed on the detail

Given a contract completed with two rejected units, when a reviewer or
the owning member reads the detail, then `rejected_units` names both
`{objective_id, unit_id}` entries.

### SCN-MCF-REJECTED-UNIT-RESPONSE — Rejects do not credit

Given a pending contract with one staged unit rejected by verdict, when
a moderator approves, then the `200` response lists only the approved
credits and the standing increases only by those amounts.

## Catalog Remodel (implementation scope, not canonical text)

- `seeds/pvp-missions.json` — each mission keeps code/title/category and
  declares **one** countable objective with `UnitLimit` = required
  results and `MeritPerUnit` = old package share ÷ limit, plus
  `CompletionReward` = old verification share; `Version: "3"`.

  | definition | unit objective | limit | per unit | completion |
  | --- | --- | --- | --- | --- |
  | pvp-live-fps-5/10/20/50 | `bajas-fps-verificadas` | 5/10/20/50 | 54 | 30/60/120/300 |
  | pvp-live-air-5/10/20/50 | `derribos-verificados` | 5/10/20/50 | 90 | 50/100/200/500 |
  | pvp-live-duel-1/3/5/10 | `victorias-verificadas` | 1/3/5/10 | 180 | 20/60/100/200 |

- `seeds/racing-missions.json` + `scripts/seed_missions.py` — the
  generator emits one `vueltas-verificadas` objective with `UnitLimit: 5`
  and `MeritPerUnit` = tanda share ÷ 5 (36/72/126/198), plus
  `CompletionReward` = verification share (20/40/70/110). The two
  artificial objectives disappear. Objective-count validation relaxes to
  allow the single-objective shape.
- Rules are rewritten to ~5 essential statements per mission (PU/LIVE +
  identity on video; post-acceptance continuous evidence; distinct human
  victims/results, no NPC/allies/alts/pacts; result confirmed and
  attributable in the video; one evidence credits one mission — plus the
  family-specific identification rule for AIR/DUEL and the consecutive-
  lap rule for RACING). `Version: "3"`; the mission description keeps the
  required-result count.

## Error Vocabulary

- `UNIT_NOT_STAGED` — `422` (approve verdict names an unknown unit)
- existing: `CONTRACT_NOT_PENDING_VALIDATION`, `FORBIDDEN`,
  `CONTRACT_TERMINAL`, `INVALID_REQUEST`.
