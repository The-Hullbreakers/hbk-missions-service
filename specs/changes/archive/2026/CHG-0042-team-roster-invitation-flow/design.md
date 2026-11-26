---
change_id: CHG-0042
artifact: design
updated: 2026-10-14
---

# Design — Roster Invitation Flow

## Decisions

- **Slot reservation**: the pending invitee's active-contract slot is reserved
  at invite time inside the same atomic acceptance write (CHG-0023 invariant).
  Decline or expiry releases it. This cannot oversubscribe and keeps limits
  enforced at acceptance.
- **Expiry**: resolved lazily — pending invitations are evaluated on the next
  command against the contract (confirm, decline, start, evidence, abandon).
  No scheduler coupling in the domain write path. Confirmation window: 7 days
  (domain constant).
- **Partial decline**: the contract activates without declined/expired members
  — the roster shrinks rather than vetoing the team.

## Member Confirmation State

Each roster member on a team contract carries an invitation state:
`PENDING` → `CONFIRMED` | `DECLINED` | `EXPIRED`. The requester is implicitly
`CONFIRMED` (they initiated). `rosterFrozen` becomes true when no member is
`PENDING`. Solo contracts have no invitations and freeze at creation.

## Error Surface

- `ErrNotInvited` → 403 `NOT_INVITED` — caller is not a pending invitee.
- `ErrInvitationNotPending` → 422 `INVITATION_NOT_PENDING` — already resolved
  or contract not in a confirmable state.

## Interaction With CHG-0023 Atomicity

Pending invitations and reserved slots live inside the same atomic acceptance
write — partial invitation state can never exist.
