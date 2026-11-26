---
change_id: CHG-0040
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
| Coverage | domain / application / dynamodb | 90.4% / 91.9% / 93.1% |

## RED Evidence

- **T-040-001** (2026-10-17): `go test -run TestConfirm_Phase` → build failed — `NewContractTermsPhased`, `Phase`, `ErrPhaseNotUnlocked` undefined. Domain RED.
- **T-040-002** (2026-10-17): `TestContractApprove_PhasedMandatoryCompletionGrantsReward` → no `COMPLETION` credit with optional unmet (`AllConfirmed` gated on every objective). Domain RED.
- **T-040-003** (2026-10-17): `TestContractStore_PhasedTermsRoundTrip` → `currentPhase` restored as 0; phases not serialized. Infrastructure RED.
- **T-040-004** (2026-10-17): `TestContractDetail_ExposesPhases` + `TestConfirmEvidence_Phase2ObjectiveRejected` → projection lacked `phases`/`current_phase`; `PHASE_NOT_UNLOCKED` unmapped. Interface RED (tests written before the projection edits).

## Regression Notes

- Flat-contract confirm regression (`advancePhases` ran past the last phase) was caught by existing unit tests and fixed by clamping at `PhaseCount()-1`.
- 150-line limit: `contract_terms.go` split into `contract_terms_phases.go`; `contract_detail.go` split into `contract_detail_projection.go`.

## L3 Gate Notes

- Security: progression is domain-side; `PHASE_NOT_UNLOCKED` is a 422, not an auth change — no new trust boundary.
- Critical-E2E: phase gating is covered by domain + HTTP tests; no new user journey beyond confirm evidence.
- Mutation: invariant coverage exercised via the flat-contract regression the unit suite caught.

