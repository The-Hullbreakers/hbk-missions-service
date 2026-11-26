---
change_id: CHG-0025
artifact: design
status: proposed
updated: 2026-10-09
---

# Design — Rename Definition Identifier To Definition Code

Pure rename; no behavior change. Every layer uses `DefinitionCode` /
`definition_code` for the operator-chosen identifier.

## Domain

- `MissionDefinition.DefinitionID` → `DefinitionCode`;
  `NewMissionDefinition(definitionCode, category, capacity, mode)` — same
  non-empty validation via `ErrDefinitionInvalid`.
- `MissionOffer.DefinitionID` → `DefinitionCode`;
  `NewMissionOfferFromDefinition` guard unchanged (`ErrOfferInvalid`).
- `ActiveContractSummary.DefinitionID` → `DefinitionCode`;
  `NewActiveContractSummary` param renamed.
- `MissionContract.Summary()` and `acceptance.go` comparisons read
  `DefinitionCode`; the `== ""` → fallback-to-`OfferID` rule is preserved.

## Application

- `CreateMissionDefinitionHandler.Handle(definitionCode, ...)`.
- `PublishMissionDefinitionHandler.Handle(definitionCode, version)`;
  `DefinitionReader.Get/GetVersion(definitionCode, ...)`.
- `PublishMission` offer handler takes `definitionCode`.
- `list_offers.go` and `accept_mission_attempt.go` hidden/lookup maps keyed by
  `DefinitionCode`; legacy `""` → `OfferID` fallback preserved.

## Interface (HTTP)

- `createMissionDefinitionRequest.DefinitionID` →
  `DefinitionCode string \`json:"definition_code"\``.
- `publishMissionRequest.DefinitionID` → `DefinitionCode` with
  `json:"definition_code"`.
- Publish route pattern `POST /api/v1/missions/definitions/{definition_code}/publish`;
  `r.PathValue("definition_code")`.
- Responses serialize `DefinitionCode` (Go field name drives the JSON shape
  today; no custom marshalling needed).

## Infrastructure (DynamoDB)

- Item attribute `DefinitionID` → `DefinitionCode` on write for definitions,
  offers, and contracts.
- Reads (`definition_store_queries.go`, `contract_store_mapping.go`,
  `contract_store_confirm.go`, `offer_store.go`) fall back to legacy
  `DefinitionID` when `DefinitionCode` is absent — same pattern as the
  existing `"legacy"` version fallback.
- No eager data migration (pre-production, consistent with CHG-0024 rollout
  decision).

## Composition

`server.go` wiring renames identifiers only; no structural change.

## What Survives Unchanged

- Value semantics: operator-chosen, required, non-empty; no format rules.
- `definition_version` name and versioning behavior.
- Board exclusion by same-code offers across versions.
- Error codes and HTTP statuses.
