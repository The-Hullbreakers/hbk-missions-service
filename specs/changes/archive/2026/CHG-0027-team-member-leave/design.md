---
change_id: CHG-0027
artifact: design
status: proposed
updated: 2026-10-09
---

# Design — Team Member Leave

## Domain

`MissionContract` gains a `departed map[string]bool` alongside the frozen
`roster`. The roster keeps recording who accepted (history); `departed`
records who left. Derived membership:

```go
// ActiveMembers returns roster members who have not departed.
func (c *MissionContract) ActiveMembers() []string

// IsDeparted reports whether the player left this contract.
func (c *MissionContract) IsDeparted(playerID string) bool
```

`CanSubmitEvidence` is tightened to `IncludesPlayer(id) && !IsDeparted(id)` —
departed members cannot confirm, start, complete, or leave again. Because
`Abandon`, `Start`, `Complete`, and `Confirm` all gate on
`CanSubmitEvidence`, the departure propagates without duplicating checks.

New lifecycle operation:

```go
// Leave removes the caller from the contract's active membership.
// The last active member leaving abandons the contract.
func (c *MissionContract) Leave(identity CallerIdentity) error {
	if !c.CanSubmitEvidence(identity) { return ErrCallerNotAuthorized }
	if c.State.IsTerminal()          { return ErrContractTerminal }
	c.departed[identity.PlayerID()] = true
	if len(c.ActiveMembers()) == 0   { c.State = ContractAbandoned }
	return nil
}
```

Reconstructors `NewMissionContractFromState` take an extra `departed []string`
param. `Departures()` returns a defensive copy for persistence.

## Application

New `LeaveContractHandler` — same ports as `AbandonContractHandler`
(`ContractReader`, `ContractWriter`): `Get` → `contract.Leave(identity)` →
`Save`. No ledger write is needed; leaving is not a scoring event.

## HTTP Boundary

New handler `LeaveContractHandler` in `interface/http/contract_lifecycle.go`:
`POST /api/v1/missions/contracts/{contract_id}/leave`. Identity from context;
`403` when the domain returns `ErrCallerNotAuthorized`; `422` for
`ErrContractTerminal`/`ErrContractStateInvalid` (departed callers hit the
403 path first — they are no longer evidence-capable).

Error mapping: `ErrCallerNotAuthorized` → `403 FORBIDDEN`;
`ErrContractTerminal` → `422 CONTRACT_TERMINAL`. A departed caller's leave
retry surfaces `403` — consistent with outsider evidence rejection and no
new error code is introduced.

## Persistence

Contract item adds `"Departed"` string-set (`SS`/`L` following
`stringSliceToAttribute`). `mapToContract` reads it into the reconstituted
domain state.

## Composition

`server.go` wires `LeaveContractHandler` with the same `contractStore`
ports and mounts `POST /api/v1/missions/contracts/{contract_id}/leave` behind
`auth.Authenticate`, next to `abandon`.

## Slot/Policy Effect

`Summary()`/active-limit and `REQ-MB-008` board exclusion operate per
caller. Since the summary is contract-level, exclusion must consult the
caller: `ActiveContractSummary` gains an `Excludes(playerID)`-style check or
the list filter consults `contract.IsDeparted(caller)` — simplest: the
acceptance/list policy skips contracts where the caller departed. This is a
small filter change where active contracts are computed for a caller.
