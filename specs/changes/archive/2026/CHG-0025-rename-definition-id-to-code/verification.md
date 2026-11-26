---
change_id: CHG-0025
artifact: verification
status: proposed
updated: 2026-10-09
---

# Verification — Rename Definition Identifier To Definition Code

## Evidence Plan

| Gate | Evidence | Status |
| --- | --- | --- |
| Gate 1 | `change.md` approved 2026-10-09 | done |
| Gate 2 | delta/design/tasks approved before RED | done |
| Domain RED→GREEN | `go test ./tests/unit/missions/domain/` | done — rename-RED (compile: `def.DefinitionCode undefined`) then green |
| Application RED→GREEN | `go test ./tests/unit/missions/application/` | done — RED (compile) then green |
| Boundary RED→GREEN | `go test ./tests/unit/missions/interface/http/ ./tests/integration/http/` | done — RED (`422` on `definition_code` body) then green |
| Infrastructure RED→GREEN | `go test ./tests/integration/infrastructure/dynamodb/` | done — green; legacy `DefinitionID` attribute still resolves via read fallback |
| E2E | create → publish → offer → accept carries `definition_code` | done — `tests/integration/composition/publish_mission_test.go`, `solo_journey_e2e_test.go` |
| Coverage | `>= 90%` affected production scope | done — domain 91.6%, application 91.1%; dynamodb 87.7% (unchanged error paths, same as baseline) |
| Validation | `tools/validate-sdd-change.sh` | done — gofmt/vet/unit/integration/race all pass |

## Residual Risk

- **Breaking contract change**: clients sending `definition_id` in bodies get
  `422`/ignored-field behavior depending on endpoint; `docs/` and
  `template.yaml` RouteKey are updated in this change.
- **Legacy stored items** keep the `DefinitionID` attribute; reads fall back
  (`definitionFromItem`, `mapToContract`, offer `Get`). New writes emit
  `DefinitionCode` only.

## Sign-Off

All gates green 2026-10-09: full unit + integration suites green under race
detector via `tools/validate-sdd-change.sh`. Ready for human review and
consolidation into `specs/capabilities/`.
