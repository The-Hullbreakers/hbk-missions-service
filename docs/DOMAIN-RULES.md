# Domain Rules

## Mission Definition And Offer

A definition is editable in draft. Publishing it creates an immutable version. An offer references that version and controls how many times it can be accepted.
```text
DRAFT → PUBLISHED → RETIRED
AVAILABLE → CLOSED
      ↘ WITHDRAWN
```

Withdrawing an offer prevents new acceptances, but does not change existing contracts unless explicitly canceled with reason.
## Eligibility And Acceptance

Acceptance requires simultaneously:
1. same tenant;
2. available offer;
3. eligible member;
4. available space;
5. repeatability/cooldown allowed;
6. active contracts below the effective limit.
The effective limit uses category override when it exists; otherwise use the tenant default. Two concurrent acceptances cannot approve both if the second would exceed limit or capacity.
## Active Template Exclusion

- An active contract excludes from the board any offer with the same `mission_definition_id`.
- The exclusion crosses versions of the template.
- It is also validated when accepting by direct ID.
- As a team it applies to each confirmed participant.
- Terminal states release the exclusion.
- Cooldown and repeatability are evaluated afterwards and may keep the offer hidden.
- Two concurrent acceptances of the same template produce a maximum of one active contract.
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
- Each confirmed must have capacity within their personal limit.
- The roster freezes at `IN_PROGRESS`.
- Joining later does not grant retroactive credit.
- The abandonment of a member does not change recipients of already confirmed credits.
## Objectives And Confirmation

A phase may require all or some objectives depending on its published rule. Each confirmable unit has a stable identity within the contract.
A valid confirmation proves active contract, window, target, receiver, source and uniqueness. The same unit cannot be credited again by replay, retry or duplicate review.
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
