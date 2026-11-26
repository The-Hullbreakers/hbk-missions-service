---
change_id: CHG-0042
artifact: tasks
status: implemented
updated: 2026-10-16
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-042-001 | Add member invitation state (PENDING/CONFIRMED/DECLINED/EXPIRED), requester auto-confirmed, `rosterFrozen` on all-resolved, `ErrNotInvited`/`ErrInvitationNotPending` | domain-rule | done | domain |
| T-042-002 | `ConfirmInvitation`/`DeclineInvitation` use cases; invite-at-accept pending members; lazy 7-day expiry on contract commands | application-command | done | application |
| T-042-003 | Persist pending members + reserved slots in atomic acceptance write; load/save invitation state | persistence-adapter | done | infrastructure |
| T-042-004 | `POST /contracts/{id}/confirm-invitation`, `POST /contracts/{id}/decline-invitation`; `NOT_INVITED`/`INVITATION_NOT_PENDING` codes; pending members in contract detail | rest-endpoint | done | interface |

## Requirement to Test to Production Map

- **REQ-MCF-025** → invitation lifecycle tests → `domain/`, `application/`, `infrastructure/dynamodb/`, `interface/http/`.

## Gates

L3: security gate (invitee consent trust boundary), critical E2E (confirm→freeze journey), coverage ≥ 90%, mutation gate on new domain rules.
