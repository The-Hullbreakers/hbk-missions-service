---
change_id: CHG-0023
artifact: delta
status: consolidated
updated: 2026-10-08
---

# Delta — Team Participation and Season Integrity

## What Is Changing

### Mission Contract Fulfillment

- `team_id` is meaningful only for `TEAM`/`FLEX` offers; supplying it on a
  `SOLO` offer is rejected.
- A team contract's roster is the requester plus validated teammates from the
  declared team; every confirmed member consumes one active-contract slot.
- Team credits land on the team track only.

### Seasonal Category Competition

- `create season` never overwrites an existing season id; the `Official` flag
  remains set-at-creation and immutable.
- The scheduler retries next-season creation on every tick until it succeeds.
- `CompletedMissions` tie-break counts only `COMPLETED` contracts.

## MODIFIED

### REQ-MCF-001 — Execute solo or team

`SOLO` offers cannot be taken as a team. For `TEAM`/`FLEX` offers accepted as a
team, the roster contains the requester and the validated teammates; each
member consumes one slot of their personal active-contract limit, and the
contract uses the team track.

### REQ-SCC-005 — Create a season with an official-statistics flag

Creation is refused when a season with the same `season_id` already exists in
the tenant (code `SEASON_ALREADY_EXISTS`); no state or `Official` flag of an
existing season can change through creation.

### REQ-SCC-008 — Apply published tie-breaks

`CompletedMissions` counts only contracts that reached `COMPLETED`; contracts
in `IN_PROGRESS` or any non-terminal state with credits do not count.

### REQ-SCC-010 — Auto-freeze, finalize, and start the next season

When next-season creation fails after finalization, subsequent scheduler ticks
retry it; the series eventually advances and the failed tick is observable.

## BDD Scenarios

### SCN-MCF-TEAM-SOLO-REJECT — Team id on a solo offer

Given a `SOLO` offer, when a member accepts it with a `team_id`, then the
acceptance is rejected with `TEAM_NOT_ALLOWED` and no contract is created.

### SCN-MCF-TEAM-ROSTER — Validate teammates and their limits

Given a `TEAM` offer and teammates `player-2`, `player-3` of the caller's team,
when one teammate is already at their active limit, then the acceptance is
rejected with `TEAMMATE_LIMIT_REACHED` and no contract is created.

### SCN-MCF-TEAM-CONTRACT — Team contract consumes every member's slot

Given a valid team acceptance with two teammates, when the contract is
created, then all three members consume one active slot and the contract uses
the `TEAM` track.

### SCN-SCC-CREATE-EXISTS — Reject overwrite of an existing season

Given `season-x` exists as `FINALIZED` with `Official: true`, when an operator
creates `season-x` again, then the response is `422` with
`SEASON_ALREADY_EXISTS` and the stored season is unchanged.

### SCN-SCC-RETRY-NEXT — Retry failed next-season creation

Given a finalized scheduled season whose next-season creation failed once,
when the scheduler ticks again, then the next season is created `ACTIVE` and
the series resumes.

### SCN-SCC-TIE-COMPLETED — Count only finished contracts

Given two participants tied on merits where one has an `IN_PROGRESS` credited
contract and the other a `COMPLETED` credited contract, when the leaderboard
applies tie-breaks, then the participant with the `COMPLETED` contract ranks
first.

## Error Vocabulary

- `TEAM_NOT_ALLOWED` — `422`
- `TEAMMATE_LIMIT_REACHED` — `422`
- `TEAMMATE_NOT_MEMBER` — `422`
- `SEASON_ALREADY_EXISTS` — `422`
