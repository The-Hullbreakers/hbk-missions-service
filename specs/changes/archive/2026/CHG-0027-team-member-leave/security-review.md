---
change_id: CHG-0027
artifact: security-review
status: approved
updated: 2026-10-09
---

# Security Review — Team Member Leave

## Role

Reviewer: self (TDD evidence). Focus: roster authority on scoring.

## Trust-Boundary Analysis

- **New endpoint** `POST /contracts/{id}/leave` sits behind
  `auth.Authenticate` — unauthenticated callers get `401` before reaching
  domain logic.
- **Caller identity** is resolved from the OIDC session context, never from
  the request body — a caller cannot leave on behalf of another player.
- **Authority gate:** `Leave` requires `CanSubmitEvidence` → roster member.
  An outsider gets `403 FORBIDDEN`; no information about contract existence
  leaks beyond the standard `CONTRACT_NOT_FOUND` path.
- **Post-departure authority:** `CanSubmitEvidence` excludes departed
  members, so a leaver can no longer confirm evidence, start, complete, or
  leave again — the departure actually revokes scoring authority.
- **Ledger integrity:** leaving writes no `MeritCredit`; the already-recorded
  credits are immutable and stay — no replay surface is added.
- **Denial-of-service:** a departed member cannot re-trigger state changes;
  the last-member rule terminalizes the contract, preventing zombie
  contracts that permanently consume capacity.

## Data Exposure

- `Departed` is persisted as a plain string set — same sensitivity class as
  `Roster`, already stored on the item. No PII beyond existing player ids.

## Findings

- **Verified:** departed caller rejected on evidence with `403` (e2e
  `team_leave_e2e_test.go`), lifecycle actions rejected with
  `ErrCallerNotAuthorized`/`ErrContractTerminal` (unit + HTTP tests), and
  `ListActiveByPlayer` no longer counts the leaver's contract — the slot is
  freed and the board template unhides.
- No Critical/High findings. **Approved**.
