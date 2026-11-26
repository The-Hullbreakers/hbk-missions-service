---
change_id: CHG-0042
artifact: delta
status: proposed
updated: 2026-10-14
---

# Delta

## What Is Changing

### Mission Contract Fulfillment

- Team acceptance creates a contract with invitees in PENDING_CONFIRMATION.
- Invitees confirm or decline; roster freezes when all confirm (or window lapses and pending are dropped).
- Declined/expired invitations release the reserved slot.

## Requirements

### REQ-MCF-025 — Confirm team invitations before roster freeze (added)

Team contracts require invitee confirmation; the roster freezes on unanimous confirmation; pending slots are reserved and released on decline/expiry.

## BDD Scenarios

### SCN-MCF-INVITE-CONFIRM — All confirm

Given a TEAM acceptance with two invitees, when both confirm, then the roster freezes and the contract starts with all members.

### SCN-MCF-INVITE-DECLINE — Decline releases slot

Given an invitee declines, when the contract activates, then the declined member is absent and their slot released.

### SCN-MCF-INVITE-EXPIRE — Window lapses

Given an invitee never responds, when the confirmation window lapses, then the pending member is dropped and the slot released.

## Error Vocabulary

- `INVITATION_NOT_PENDING` — 422 (new); `NOT_INVITED` — 403 (new).
