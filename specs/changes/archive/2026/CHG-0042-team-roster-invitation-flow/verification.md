---
change_id: CHG-0042
artifact: verification
updated: 2026-10-14
---

# Verification

## Status

`implemented — pending final review`

## Evidence Plan

| Check | Command / Method | Result |
| --- | --- | --- |
| Unit suites | `go test ./tests/unit/...` | pass |
| Integration | `go test -p=1 ./tests/integration/...` | pass |
| Race | `go test -race -p=1 ./tests/...` | pass |
| SDD validation | `tools/validate-sdd-change.sh` | pass |
| Coverage | unit domain/app + dynamodb infra | 91.8% / 92.1% / 92.9% |

## Gate Results

- **Security gate**: invitee consent is enforced at the domain boundary — `ErrNotInvited` (403) for non-roster callers, `ErrInvitationNotPending` (422) for resolved invitations; declined/expired members leave the roster so they cannot submit or start. No Critical/High.
- **Critical E2E**: confirm→freeze journey covered by domain, application, interface, and DynamoDB round-trip tests; `SCN-MCF-INVITE-*` all exercised.
- **Mutation gate**: no mutation tool installed. Manual review — flipping confirm/decline targets, removing `updateRosterFrozen`, skipping `removeFromRoster`, or dropping the window check each fails a dedicated test.

## RED Evidence

- **T-042-001** (2026-10-16): `go test ./tests/unit/missions/domain/` → build failed — `InvitationState`, `ConfirmInvitation`, `DeclineInvitation`, `ResolveExpirations`, `InvitationPending/Confirmed/Declined/Expired`, `ErrNotInvited`, `ErrInvitationNotPending`, `InvitationWindow` all undefined. Correctly classified domain RED; Gate 3 scoped to `internal/missions/domain/`.

