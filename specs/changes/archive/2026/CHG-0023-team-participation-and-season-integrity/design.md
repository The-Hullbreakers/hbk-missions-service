---
change_id: CHG-0023
artifact: design
updated: 2026-10-08
---

# Design — Team Participation and Season Integrity

## Canonical Layer Scope

Domain → Application → Boundary → Infrastructure → Interface → Composition.

## Domain

- `MissionMode` rule: `team_id` valid only for `TEAM`/`FLEX`; `SOLO` rejects.
- `Roster` VO: requester + teammates; validates non-empty, distinct, and
  membership evidence supplied by application (Party Manager stays external —
  a `TeamMembershipReader` port answers "is P member of team T in tenant").
- Season aggregate: create-vs-existing invariant; `Official` immutable.
- Scheduler result type records pending next-season creation so a later tick
  retries.

## Application

- `AcceptMissionHandler`: for team acceptance, resolves teammates via
  `TeamMembershipReader`, then checks each member's effective limit inside the
  same atomic write path delivered by CHG-0022.
- `CreateSeasonHandler`: `SEASON_ALREADY_EXISTS` on any existing id (tenant
  scope from CHG-0020).
- `SeasonScheduler`: finalize-and-create becomes two steps; failure to create
  leaves a retry marker consumed by subsequent ticks.

## Infrastructure

- `SeasonStore.PutSeason` uses `ConditionExpression attribute_not_exists(PK)`
  — overwrite impossible regardless of caller.
- `merit_ledger_readers` tie-break: count credits grouped by contract joined to
  contract state = `COMPLETED`.

## Boundary — HTTP

- New `422` codes: `TEAM_NOT_ALLOWED`, `TEAMMATE_LIMIT_REACHED`,
  `TEAMMATE_NOT_MEMBER`, `SEASON_ALREADY_EXISTS`.

## Test Strategy

- Domain: mode/team matrix; roster VO partitions.
- Application: membership-reader stub; teammate-limit partitions; scheduler
  retry with injected creation failure; create-existing rejection.
- Infrastructure: conditional `Put` cannot overwrite finalized season;
  tie-break projection with mixed contract states.
- HTTP: team violation codes; create-existing `422`.
- Critical E2E: solo-offer `team_id` exploit rejected; scheduled series
  recovers after injected failure.

## Risk And Rollback

- **Risk**: `TeamMembershipReader` source (Party Manager) may not exist yet —
  fallback: identity session's `TeamIDs` claim validates the requester's own
  team; teammate validation uses the minimal port with a stub adapter until
  Party Manager integrates (decision recorded at Gate 2).
- **Rollback**: revert; conditional Put is additive safety.
