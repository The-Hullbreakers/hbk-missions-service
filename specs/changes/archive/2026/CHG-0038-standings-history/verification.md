---
change_id: CHG-0038
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
| Coverage | domain / application / dynamodb | 91.8% / 91.8% / 92.9% |

## RED Evidence

- **T-038-001** (2026-10-17): `go test -run TestListStandingsHistory` → build failed — `application.NewListStandingsHistoryHandler`, `domain.ErrSeasonStateInvalid`, `domain.StandingBucket` undefined. Domain/application RED.
- **ListBuckets adapter** (2026-10-17): `go test -run TestMeritLedgerStore_ListBuckets` → build failed — `store.ListBuckets` undefined. Infrastructure RED.
- **T-038-002** (2026-10-17): `go test -run TestStandingsHistory` → build failed — `missionhttp.StandingsHistoryHandler` undefined. Interface RED.

## BLUE Notes

- `merit_ledger_readers.go` exceeded 150 lines after `ListBuckets`; split into `merit_ledger_buckets.go` (56 lines).

