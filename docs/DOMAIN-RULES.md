# Domain Rules

## Mission Definition And Offer

A definition carries player-facing content: `title` and `description` are required, and every objective carries a readable `name`. It also declares `completion_reward` — the competitive points granted when a reviewer approves a fully confirmed contract, credited once to the season ledger under `COMPLETION` on the same track and recipient as objective credits. A zero reward records no credit.

A definition is editable in draft. Publishing it creates an immutable version. An offer references that version and freezes `title`, `description`, `completion_reward` and named objectives, so later definition edits never alter live offers or contracts. An offer's repeatability policy controls whether a player may accept it again; it is not a global acceptance cap.
```text
DRAFT → PUBLISHED → RETIRED
AVAILABLE → CLOSED
      ↘ WITHDRAWN
```

Withdrawing an offer prevents new acceptances, but does not change existing contracts unless explicitly canceled with reason.
## Eligibility And Acceptance

Acceptance requires simultaneously:
1. an available offer bound to a live season;
2. eligible member;
3. offer window open and repeatability/cooldown allowed;
4. active contracts below the player's effective limit.
The effective limit uses category override when it exists; otherwise use the global default. Two concurrent acceptances cannot approve both if the second would exceed the limit.
## Active Template Exclusion

- A contract in which the player still participates excludes from the board any offer with the same `definition_code`.
- The exclusion crosses versions of the template.
- Every non-terminal state blocks retake: `PENDING_VALIDATION` still hides the template even though it no longer consumes an active-limit slot.
- It is also validated when accepting by direct ID.
- As a team it applies to each confirmed participant.
- Departure releases the exclusion for the departing member only; the remaining active members still hide the template.
- Terminal states release the exclusion.
- Cooldown and repeatability are evaluated afterwards and may keep the offer hidden.
- Two concurrent acceptances of the same template produce a maximum of one active contract per player; the uniqueness is enforced at commit by a per-player definition reservation that is distinct from the active-slot count.
## Contract Lifecycle

```text
ACCEPTED → IN_PROGRESS → PENDING_VALIDATION → COMPLETED
    ├──────────────→ ABANDONED
    ├──────────────→ EXPIRED
    ├──────────────→ FAILED
    ├──────────────→ CANCELLED
    └──────────────→ DISQUALIFIED
```

All final states are irreversible. A new attempt creates another contract when repeatability and cooldown allow it.
## Teams

- `SOLO` has a participant.
- `TEAM` requires confirmed roster within size limits.
- `FLEX` choose one of those tracks before starting.
- The roster freezes at `IN_PROGRESS`.
- Joining later does not grant retroactive credit.
- The abandonment of a member does not change recipients of already confirmed credits.
- A roster member may leave a non-terminal contract individually (`POST /contracts/{id}/leave`); the contract keeps running for the remaining members.
- A departed member loses all contract rights — evidence, start, submit, and leave are rejected — and the contract stops counting toward their active limit and stops hiding the template on their board.
- The accepted roster is preserved as history; departures are tracked separately. The last active member leaving abandons the contract (solo leave equals abandon).
- Every member exit — including the last member's — commits through one departure write that also releases the leaver's definition reservation and active-slot entry; members who departed earlier already released their own.
- A contract write computed on a snapshot taken before a departure cannot restore the departed member: ordinary saves, approvals, and stale departures conflict instead of overwriting the committed exit.
- `abandon` remains the contract-level exit that ends the mission for the whole roster.
## Objectives And Confirmation

A phase may require all or some objectives depending on its published rule. Each confirmable unit has a stable identity within the contract.
A valid confirmation proves active contract, window, target, receiver, source and uniqueness. The same unit cannot be credited again by replay, retry or duplicate review.
## Review Adjudication

A submitted contract carries staged unit credits — one per confirmed unit.
The reviewer approves per unit: each staged unit either pays or is
recorded as rejected. Rejected units never reach the ledger and do not
count toward fulfillment. The completion reward pays only when every
mandatory objective reached its `unit_limit` in approved (not rejected)
units; an unlimited objective needs at least one. Adjudication is a
single decision per contract — a partial verdict pays the accepted units
and closes the contract as COMPLETED without the bonus.
## Season Merits

- A credit is positive and references its confirmation.
- An adjustment can be positive or negative and requires reason/authority.
- Credit and adjustment never change season, category or track.
- The total may go down due to correction, but the history remains.
- There is no conversion between categories in v1.
## Bounty PvP Example

An offer grants 100 merits per eligible target and 150 for completing five. Four unique targets produce 400; five produce 650. Repeating an already confirmed target does not produce another credit.
## Season Lifecycle

```text
DRAFT → SCHEDULED → ACTIVE → FREEZING → FINALIZED → ARCHIVED
MANUAL ────────────→ ACTIVE → FREEZING → FINALIZED → ARCHIVED
```

A scheduled season has a `StartsAt` timestamp and a `SeasonSchedule` (`active_weeks`, optional `off_weeks`, `series` prefix/iteration). Its `SeasonID` is derived as `{prefix}-{iteration}`. The scheduler, driven by a `Clock`, auto-freezes the season when the active window ends, auto-finalizes it when the off-season ends, and immediately creates the next `ACTIVE` season in the series. Manual seasons omit `schedule` and are never auto-transitioned.

`FREEZING` rejects new merits and allows only eligible work started on time to be resolved. `FINALIZED` freezes standings and award tiers. The following season has a new ledger with a total of zero.
## Rankings And Awards

Order v1: merits, completed contracts, unique objectives, time to reach the result and stable ID. The policy is published before `ACTIVE`.
Missions decides rank/tier. Achievements materializes the permanent achievement. Retrying the same award does not create another achievement.
