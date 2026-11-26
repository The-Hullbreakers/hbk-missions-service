---
change_id: CHG-0035
artifact: verification
updated: 2026-10-17
---

# Verification

## Status

`implemented — pending final review`

## Evidence Plan

| Check | Command / Method | Result |
| --- | --- | --- |
| Unit suites | `go test ./tests/unit/...` | pass |
| Integration | `go test -p=1 ./tests/integration/...` | pass |
| SDD validation | `tools/validate-sdd-change.sh` | pass |
| Coverage | domain / application / dynamodb | 91.8% / 91.9% / 92.9% |

## RED Evidence

- **T-035-001** (2026-10-17): `TestGetContract_RosterMemberGetsDetail` → `caller is not authorized for this contract` — application RED.
- **T-035-002/003** (2026-10-17): `TestContractDetail_MemberGetsMinimizedDetail` → member received `media_uri` in `pending_credits` — interface RED.

## Security Notes

- Member projection filters `pending_credits` to the caller's `recipient_id` and strips `media_uri`; reviewers retain the full view.
- Departed members and outsiders return `403`; no store-side or route-side bypass — `adminOrModerator` middleware removed from the route.

