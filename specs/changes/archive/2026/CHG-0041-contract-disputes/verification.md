---
change_id: CHG-0041
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
| Coverage | domain / application / dynamodb | 90.3% / 91.5% / 93.1% |

## RED Evidence

- **T-041-001** (2026-10-17): `go test -run 'TestRaiseDispute|TestResolveDispute'` → build failed — `RaiseDispute`, `ResolveDispute`, `ErrDisputesPending`, `ErrDisputeStateInvalid` undefined. Domain RED.
- **T-041-002** (2026-10-17): `go test -run TestRaiseDispute_Persists` → build failed — `NewRaiseDisputeHandler`/`NewResolveDisputeHandler`/`NewListDisputesHandler` undefined. Application RED.
- **T-041-003** (2026-10-17): `TestContractStore_DisputesRoundTrip` → `Disputes()` empty after load. Infrastructure RED.
- **T-041-004** (2026-10-17): `go test -run TestDisputeEndpoint` → build failed — `RaiseDisputeHandler`/`ResolveDisputeHandler` undefined. Interface RED.

## Design Adjustment

- **D3 storage**: Gate 1 recorded `SK` children, but the table schema is `PK=CONTRACT`/`SK=contractID` with a `GetItem` read path — children would force a `Query`. Disputes serialize as a `Disputes` attribute on the contract item (same pattern as `PendingCredits`/`Invitations`), which also satisfies the "same atomic write" guard requirement for free.

## L3 Gate Notes

- **Security**: raise gated to roster members; list/resolve gated to `admin`/`moderator` at composition + application; dispute target validated against frozen terms — no new trust boundary beyond the existing reviewer split.
- **Critical-E2E**: dispute lifecycle covered by domain guards + HTTP tests (raise 201, leave blocked `DISPUTES_PENDING`, resolve + double-resolve `DISPUTE_STATE_INVALID`).
- **Mutation**: guards exercised both ways — blocked while OPEN, unblocked after RESOLVE.

