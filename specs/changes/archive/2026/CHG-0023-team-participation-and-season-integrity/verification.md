---
change_id: CHG-0023
artifact: verification
updated: 2026-10-13
---

# Verification — CHG-0023

## Status

`verified` — all suites green, L3 gates recorded.

## Evidence Plan

| Check | Command / Method | Result |
| --- | --- | --- |
| Unit suites | `go test ./tests/unit/...` | pass |
| Integration (http + infra) | `go test ./tests/integration/...` | pass |
| Composition E2E | `go test ./tests/integration/composition/` | pass — solo `team_id` exploit → `422 TEAM_NOT_ALLOWED`; duplicate `season_id` → `422 SEASON_ALREADY_EXISTS`, stored season untouched |
| Race | `go test -race ./tests/unit/... ./tests/integration/composition/ ./tests/integration/http/` | pass |
| SDD validation | `tools/validate-sdd-change.sh` | pass |
| Coverage | domain / application / dynamodb | 91.1% / 92.5% / 92.6% — meets ≥90% |
| Mutation gate | manual mutants | 4/4 killed: member-limit `>` vs `>=`; SOLO legality branch removed; `SuccessorPending` forced false; `attribute_not_exists` condition dropped |
| Critical E2E | `season_integrity_e2e_test.go`, `accept_team_violation_test.go` | pass — exploit rejected; season integrity holds end to end |
| Security gate | trust-boundary review | recorded below — no Critical/High findings |

## Security Gate

- **Requester identity** still comes only from session claims; `team_id` and
  `teammates` are body fields, never trusted for identity.
- **Teammate validation** goes through `TeamMembershipReader`. Party Manager
  is not yet integrated, so composition wires `membership.DenyAllReader`:
  every declared teammate is rejected (`TEAMMATE_NOT_MEMBER`) instead of
  accepting unverifiable input — a permissive stub would let any caller
  consume another player's active-contract slot.
- **Season integrity** is enforced at the store (`attribute_not_exists(PK)`),
  not just in the handler — a buggy caller cannot overwrite season state or
  `Official`.
- **Member slot consumption** remains a single DynamoDB transaction per
  roster member; no partial acceptance.

## Reproduction Evidence (pre-fix)

- `team_id` on a `SOLO` offer produced a team contract with a one-member roster.
- Creating a season with an existing finalized id rewrote it `ACTIVE` and
  changed `Official`.
- Injected next-season-creation failure → subsequent ticks never retried.
- `IN_PROGRESS` credited contract counted toward `CompletedMissions`.
