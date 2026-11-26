---
change_id: CHG-0050
artifact: security-review
status: pending
updated: 2026-10-20
---

# Security Review — Per-Unit Adjudication And Completion Bonus

## Trust Boundary

`POST /contracts/{id}/approve` is a privileged adjudication boundary
(`admin`/`moderator` only). This change widens its input surface with a
`rejected_units` verdict list.

## Risks To Evaluate

- **Privilege**: verdict path must keep the same role check; roster
  members can never adjudicate (existing SCN-MCF-REVIEW-MEMBER-DENIED
  coverage must still pass).
- **Input integrity**: `rejected_units` entries are validated against the
  contract's staged set — unknown units fail the whole write so a forged
  or mistyped verdict cannot silently alter payouts.
- **Ledger integrity**: rejected units write no credit; the contract item
  records them for audit. Atomic `SaveApproved` transaction unchanged.
- **Bonus abuse**: completion requires full post-verdict satisfaction —
  a rejected unit blocks the bonus; re-attempts freeze fresh terms and
  their own bar (existing retake semantics, no new exposure).
- **Information exposure**: member detail projection scopes
  `rejected_units` like `pending_credits` (own recipient only, no media).

## Findings

Pending — to be filled by the security gate run before consolidation.
