---
change_id: CHG-0053
artifact: canonical-delta
status: archived
updated: 2026-10-26
---

# Canonical Delta — mission-board

## ADDED

None — this change corrects behavior `REQ-MB-008`/`REQ-MB-004` already
promise; the delta strengthens their wording to match the intended
guarantee.

## MODIFIED

### REQ-MB-004 — Consistent acceptance

Acceptance validates eligibility, window, repeatability, and limit in a
single decision, then commits the contract, each roster member's
active-set entry, and a per-player+definition reservation for every
roster member in **one conditional transaction**. Two concurrent
acceptances serialize: the first commits and the rest receive the
explainable rejection (`ACTIVE_MISSION_LIMIT_REACHED`,
`TEAMMATE_LIMIT_REACHED`, `SEASON_NOT_ACTIVE`, or
`MISSION_DEFINITION_ALREADY_ACTIVE`) with no partial writes. Reaching a
terminal state releases the contract from the player's active set and
frees the reservation of each member still active on it — never the
reservation of a departed member, which may already belong to a newer
attempt.

#### SCN-MB-CONCURRENT-ACCEPTANCE — Protect the last slot

Given the player has one active-set slot remaining, when two acceptances
compete, then exactly one obtains the contract and the other receives an
explainable rejection.

#### SCN-MB-UNIQUE-CONCURRENT — Concurrent same-definition accepts serialize

Given a player with free active-set slots holds no contract of definition
X, when two acceptances for offers of X compete for that player, then
exactly one contract is created and the other is rejected with
`MISSION_DEFINITION_ALREADY_ACTIVE`.

#### SCN-MB-UNIQUE-OTHERS — Different players are independent

Given players `player-1` and `player-2` both have free slots, when each
accepts an offer of the same definition concurrently, then both contracts
are created.

### REQ-MB-008 — Hide an active template

If a player is an active member of an `ACCEPTED`, `IN_PROGRESS`, or
`PENDING_VALIDATION` contract from a stable definition, the board
excludes every offer of the same `definition_code`, even from another
version. `PENDING_VALIDATION` still blocks the retake even when the
player's active-mission count has spare slots (its slot accounting is
governed by `REQ-MB-003`; the known submit-release defect is tracked
outside this change).

The exclusion binds to **active membership**, not roster history: once a
member leaves (`REQ-MCF-013`), offers of that template reappear on their
board, the old contract stops consuming their active-mission slot, and a
fresh acceptance is evaluated without it — while members still active on
the contract keep it hidden and counted.

Acceptance enforces the rule atomically, not only at decision time: the
commit reserves each roster member's (player, definition) pair, so a
concurrent duplicate of the same definition — via any offer version or as
a roster teammate in another acceptance — is rejected with
`MISSION_DEFINITION_ALREADY_ACTIVE` instead of creating a second
contract. The reservation belongs to the contract that created it: a
release succeeds only when the reservation is absent or owned by that
contract, so a stale or delayed release from an older attempt can never
free a newer one — finding a different owner conflicts and writes nothing
instead of deleting it. The reservation is released on terminal state
only for members still active on the contract, and for the departing
member alone — atomically with the leave write — when a member leaves.

The departure write itself is guarded the same way: it commits only when
the stored contract still matches the snapshot the member left from —
the same lifecycle state and the same set of previously departed members.
A leave committed on a stale snapshot conflicts and writes nothing, so
parallel departures on stale snapshots can never resurrect a member whose
leave already persisted.

#### SCN-MB-ACTIVE-TEMPLATE — Do not show or retake the template

Given a player has active template X, when they browse the board or
attempt to accept another offer of X, then they do not see it and direct
acceptance is rejected.

#### SCN-MB-TEMPLATE-RELEASED — Show after terminal state

Given the contract for template X ended, when cooldown and repeatability
allow another attempt, then the player sees eligible offers of X again
and may accept one.

#### SCN-MB-DEPART-VISIBLE — Departure re-shows the template to the leaver only

Given a team contract of definition X in which `player-1` departed while
`player-2` stays active, when each browses the board, then `player-1`
sees eligible offers of X and `player-2` does not.

#### SCN-MB-DEPART-RETAKE — Departed member retakes without the old contract

Given `player-1` left a contract of definition X and has a free slot,
when they accept an offer of X, then a new contract is created and the
old contract no longer counts toward their active-mission limit.

#### SCN-MB-DEPART-RACE — Stale leave write cannot restore a departed member

Given `player-1` and `player-2` each left a team contract on separate
snapshots, when `player-1`'s departure has already committed, then
`player-2`'s stale write conflicts, persists nothing, and releases no
guards — the stored contract keeps `player-1` departed.

#### SCN-MB-UNIQUE-VERSION — Another version is still the same template

Given `player-1` has an active contract of definition X version 1, when
they accept an offer of X version 2, then the acceptance is rejected with
`MISSION_DEFINITION_ALREADY_ACTIVE`.

#### SCN-MB-UNIQUE-TEAMMATE — Team accept cannot seat a member who already has it

Given `player-2` already has an active contract of definition X, when a
team acceptance of X names `player-2` in its roster, then the acceptance
is rejected and no contract is created.

#### SCN-MB-PENDING-BLOCKS — Awaiting review still holds the template

Given `player-1`'s contract of definition X is `PENDING_VALIDATION` and
their active-mission count has spare slots, when they attempt to accept
an offer of X, then the acceptance is rejected with
`MISSION_DEFINITION_ALREADY_ACTIVE`.

### REQ-MCF-013 — Leave a team contract without terminating it

Every member exit — including the last active member, whose leave
transitions the contract to `ABANDONED` — commits through one focused
departure write that atomically persists the contract and releases the
departing member's guards: their per-definition reservation and their
`PLAYER_ACTIVE` contract id. Guards of members who departed earlier were
released at their own leaves and are never touched by a later departure.
The accepted roster and the departed set stay preserved as history
(`REQ-MB-008` governs the board consequence: the departed member may see
and retake the definition under a free slot).

The departure write is optimistic: it commits only when the stored
contract still matches the state the member left from — the same
lifecycle state and the exact set of previously departed members. A
leave attempted on a stale snapshot conflicts and writes nothing, so no
departure can resurrect or silently drop another member's recorded exit.

#### SCN-MCF-LEAVE-LAST-RETAKE — Last member's leave frees their own guards

Given a contract of definition X whose only active member is `player-1`,
when they leave, then the contract is `ABANDONED`, their reservation and
active-set entry are released in the same atomic write, and a fresh
acceptance of X is not blocked by the old attempt.

### REQ-MCF-010 — Commit adjudication and credits atomically

Every contract write — the approval write and the ordinary non-departure
save — is fenced on the roster truth the writer saw: the contract put is
conditional on the stored departed set exactly matching the candidate's
already-departed members (length and membership; a missing departed
attribute is tolerated only when the expected set is empty). A stale
write whose snapshot predates a committed departure conflicts and writes
nothing: no contract overwrite and no materialized credits. Credit-item
conditional failures still resolve as the duplicate-confirmation
rejection.

#### SCN-MCF-STALE-SAVE — Stale ordinary save cannot restore a departed member

Given `player-1` left a contract and a snapshot taken before the leave
is written through the ordinary save path, when the write runs, then it
conflicts, persists nothing, and the stored contract keeps `player-1`
departed.

#### SCN-MCF-STALE-APPROVAL — Stale approval cannot resurrect a member or mint credits

Given `player-1` left a `PENDING_VALIDATION` contract and a reviewer
approves on a snapshot taken before the leave, when the approval write
runs, then it conflicts atomically — the contract keeps `player-1`
departed and no credit items are written.

### REQ-MCF-020 — Adjudicate a pending contract

Approval operates on an existing contract row: the contract put requires
the row to exist and its stored departed set to match the adjudicated
snapshot, so a stale review can neither resurrect a departed member nor
attach credits to a roster state that no longer holds.

## REMOVED

None.
