---
change_id: CHG-0025
title: Rename definition identifier to definition code
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
archive_path: specs/changes/archive/2026/CHG-0025-rename-definition-id-to-code/
verified_at: 2026-10-09
source_revision: efa4223
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
context_scope: existing
created: 2026-10-09
updated: 2026-10-09
---

# Rename Definition Identifier To Definition Code

## Intent

Rename the operator-chosen mission-definition identifier from `definition_id`
to `definition_code` across the public REST contract, domain vocabulary, and
persistence mapping, so the field communicates a human-readable business code
the operator picks (e.g. `def-kills-100`) instead of a system-generated ID.

## Business Value

- The admin/operator surface reads in business language: a "code" is something
  a human chooses and recognizes; an "id" implies an opaque system value.
- One identifier only — no parallel `id` + `code` pair, no autogeneration.

## User Stories

- **US-DEF-CODE-CONTRACT**: As an API consumer, I want the definition
  identifier field named `definition_code` in requests, responses, and paths,
  so the contract reads as a business code.
- **US-DEF-CODE-VOCAB**: As a maintainer, I want domain and persistence to use
  `DefinitionCode` consistently, so code and docs share one vocabulary.

## Why L2

Breaking rename of a public contract field on a verified capability. No
behavior change: same value, same validation rules, same lookup semantics.
Requires updated HTTP/integration evidence and doc convergence.

## Scope

- **Domain**: `MissionDefinition.DefinitionID`, `MissionOffer.DefinitionID`,
  `ActiveContractSummary.DefinitionID` renamed to `DefinitionCode`; all
  constructor params and lookup helpers follow.
- **Application**: handler params and port signatures (`DefinitionReader.Get`,
  `GetVersion`, `DefinitionWriter`) rename to `definitionCode`.
- **Interface**: body field `definition_id` → `definition_code` in
  `POST /definitions` and `POST /offers`; path `{definition_id}` →
  `{definition_code}` in publish; response `DefinitionID` → `DefinitionCode`.
- **Infrastructure**: DynamoDB attribute `DefinitionID` → `DefinitionCode`
  with a read fallback to the legacy attribute (mirrors the existing
  `legacy` version fallback in `contract_store_mapping.go`).
- **Docs**: `API-CONTRACT.md`, `API-GUIDE-FRONTEND.md`, `DOMAIN-RULES.md`.

## Non-Goals

- Autogenerating the identifier server-side; the operator still supplies it.
- Adding a second identifier (`id` + `code` pair).
- Migrating stored items eagerly; the read fallback covers legacy attributes.
- Changing `definition_version` semantics.

## Sequencing

Independent intent from CHG-0024 (tenant removal) but touches the same
handlers. If CHG-0024 lands first this delta rebases onto it; either order is
valid, recorded per merge order in `verification.md`.

## Approvals

- Gate 1: approved 2026-10-09.
- Gate 2: approved 2026-10-09.
