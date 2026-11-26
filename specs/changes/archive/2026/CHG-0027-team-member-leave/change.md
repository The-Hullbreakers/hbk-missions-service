---
change_id: CHG-0027
title: Team member leave without terminating the contract
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/mission-board/spec.md
change_path: specs/changes/archive/2026/CHG-0027-team-member-leave/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
context_scope: existing
created: 2026-10-09
updated: 2026-10-09
---

# Change — Team Member Leave

## Intent

Today `POST /contracts/{id}/abandon` is contract-level: any roster member can
terminate the whole mission for everyone. There is no way for one player to
exit a team contract without ending it for the rest of the roster. This
change adds an individual `leave` so a member drops out while the remaining
roster keeps confirming, completing, and earning the reward.

## Business Value

- A player can quit a team mission without punishing the squad — the rest
  finish and get paid.
- Active-slot accounting stays correct: the leaver's slot frees and their
  board re-shows the template; the survivors keep their commitment.
- History is honest — the frozen roster records who accepted; the departed
  set records who left and when-sequence.

## User Stories

- **US-0027-01 — Leave a team contract:** As a roster member of an active
  team contract, I leave it so the contract stays alive for the remaining
  members and I stop counting toward it.
- **US-0027-02 — Continue after a departure:** As a surviving member, I
  confirm objectives and complete the contract after a teammate leaves so
  the team still earns the frozen reward.
- **US-0027-03 — Last member abandons:** As the last remaining member, when
  I leave, the contract transitions to `ABANDONED`.

## Risk Classification

**L3** — changes who may submit evidence on a live contract (trust boundary
on scoring authority), adds a public lifecycle endpoint, persists a new
`Departed` attribute in DynamoDB, and interacts with the active-slot and
board-exclusion policies in `mission-board`.

## Approach

- `MissionContract.Leave(identity)` — roster member exits a non-terminal
  contract. Guarded like `Abandon` (roster member, non-terminal).
- New `departed` set on the contract, disjoint from the frozen roster;
  `IncludesPlayer` keeps returning the accepted roster for history, while a
  new `ActiveRoster`/`CanSubmitEvidence` excludes departures.
- `Confirm`, `Start`, `Complete`, `Leave` reject departed players.
- Leaving the last active member transitions the contract to `ABANDONED`.
- Persisted via a `Departed` attribute (string set) on the contract item.
- HTTP: `POST /api/v1/missions/contracts/{contract_id}/leave` — `200` on
  success; `401` unauthenticated; `403` outsider; `422` departed-already or
  terminal.

## Behavior Delta Summary

- ADDED: individual leave on non-terminal contracts; departed exclusion
  from evidence/lifecycle; last-member leave abandons the contract.
- MODIFIED: `REQ-MCF-008` lifecycle gains the leave transition; the
  active-contract summary must stop counting a departed caller.
- REMOVED: none — `abandon` stays the "end it for everyone" escape hatch.

## Non-Goals

- No roster invites or replacements — departure is one-way.
- No reallocation of already-credited merits; the ledger stays immutable.
- No split of the completion reward across survivors/leaver — the credit
  still targets the contract recipient/track.

## Decisions

- **Departures are tracked, roster is immutable history.** The accepted
  roster is preserved; `departed` is the mutable exclusion set.
- **Leave is idempotent-safe:** a second leave by a departed player resolves
  as `422 CONTRACT_STATE_INVALID`, consistent with `repeat complete`.
- **Last member leaving abandons:** a contract with zero active members can
  never progress, so it terminates as `ABANDONED`.
- **Solo leave = abandon:** leaving a 1-member contract always terminates it.

## Sequencing

Independent; builds on the CHG-0026 completion-reward path (last-member
abandon reuses `Abandon` semantics; reward flow unaffected — the reward goes
to the contract recipient, not per-member).

## Approvals

- Gate 1: approved 2026-10-09.
