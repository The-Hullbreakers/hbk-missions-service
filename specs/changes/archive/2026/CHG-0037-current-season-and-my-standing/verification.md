---
change_id: CHG-0037
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

- **T-037-001** (2026-10-17): `go test -run TestGetCurrentSeason` → build failed — `application.NewGetCurrentSeasonHandler`, `missionhttp.GetCurrentSeasonHandler` undefined. Application/interface RED.
- **T-037-002** (2026-10-17): `go test -run TestGetMyStanding` → build failed — `domain.ErrStandingNotFound`, `application.NewGetMyStandingHandler`, `missionhttp.GetMyStandingHandler` undefined. Domain/application/interface RED.

