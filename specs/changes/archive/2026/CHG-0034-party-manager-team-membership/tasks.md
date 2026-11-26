---
change_id: CHG-0034
artifact: tasks
status: implemented
updated: 2026-10-15
---

# Tasks

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-034-001 | Add `ErrMembershipUnavailable` and HTTP adapter contract tests: 200→member, 404→not-member, error/timeout/5xx→unavailable, path escaping | infrastructure-adapter | done | infrastructure |
| T-034-002 | Implement `PartyManagerReader` HTTP adapter (2s timeout, `PARTY_MANAGER_URL` base) | infrastructure-adapter | done | infrastructure |
| T-034-003 | Map `ErrMembershipUnavailable` → 503 `MEMBERSHIP_UNAVAILABLE`; wire adapter in composition with `DenyAllReader` fallback; integration test with stubbed Party Manager | composition | done | composition |

## Requirement to Test to Production Map

- **REQ-MCF-001** → `tests/integration/infrastructure/membership/` contract tests, `tests/integration/composition/` team-accept test → `internal/missions/infrastructure/membership/party_manager.go`, `internal/missions/domain/errors.go`, `internal/missions/interface/http/errors.go`, `internal/missions/composition/server.go`.

## Gates

L3: security gate (external trust boundary — see `security-review.md`), critical E2E (team accept with real membership), coverage >= 90%, mutation.
