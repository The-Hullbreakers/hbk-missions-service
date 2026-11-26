---
change_id: CHG-0034
artifact: design
updated: 2026-10-14
---

# Design — Party Manager Team Membership

## Decisions

- **Transport**: synchronous HTTP lookup to Party Manager per declared
  teammate, one call per member per acceptance. No shared read model or
  replica (CHG-0034 non-goal: no caching).
- **Endpoint contract**: `GET {PARTY_MANAGER_URL}/teams/{team_id}/members/{player_id}`
  → `200` member / `404` not member / other or timeout → unavailable.
- **Timeout**: 2s per lookup, no retries (fail fast inside the request path).
- **Error code**: transport/timeout/5xx failures wrap `ErrMembershipUnavailable`
  → `MEMBERSHIP_UNAVAILABLE` (503). A definitive `404` is a business negative
  → `TEAMMATE_NOT_MEMBER` (422).
- **team_id existence**: not validated separately — membership lookup is the
  authoritative answer; a nonexistent team yields `404` → `TEAMMATE_NOT_MEMBER`.
- **Configuration**: `PARTY_MANAGER_URL` env var; unset/empty → composition
  keeps `membership.DenyAllReader` (fail-closed default).

## Fail-Closed Contract

`TeamMembershipReader.IsMember` returning an error propagates as
`ErrMembershipUnavailable` (503) — never treated as membership, never as
`TEAMMATE_NOT_MEMBER`. Composition keeps `membership.DenyAllReader` as the
safe default when `PARTY_MANAGER_URL` is not configured.
