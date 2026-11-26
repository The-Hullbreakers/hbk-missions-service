---
change_id: CHG-0017
artifact: tasks
updated: 2026-10-05
---

# Tasks: CHG-0017 — List definitions and guard admin endpoints by role

| ID | Task | Status | Wave |
| --- | --- | --- | --- |
| T-017-001 | Add `DefinitionStore.List` (PK=DEFINITION query) | done | infra |
| T-017-002 | Add `ListDefinitionsHandler` + `DefinitionListReader` port | done | app |
| T-017-003 | Add `GET /api/v1/missions/definitions` handler | done | iface |
| T-017-004 | Add `RequireRole` middleware (403 FORBIDDEN) | done | iface |
| T-017-005 | Wire list endpoint + role guards in `server.go` | done | composition |
| T-017-006 | Unit + http integration tests | done | tests |
| T-017-007 | Fix BUG-0017-01: add `MissionsListDefinitionsRoute` to `template.yaml` | done | iac |

## Verification

- `go build ./...` — PASS
- `go test ./tests/unit/... ./tests/integration/http/...` — PASS
- `tools/validate-sdd-change.sh --risk L2` — PASS (coverage: domain 94%, application 93.5%, dynamodb 89.9%)

## Role mapping

- `admin` — create/freeze/finalize seasons.
- `admin`, `moderator` — create/publish definitions, publish offers, list definitions.
