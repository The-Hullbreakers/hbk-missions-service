---
change_id: CHG-0005
title: Enforce single active season rule
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
change_path: specs/changes/archive/2026/CHG-0005-single-active-season/
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
archive_path: specs/changes/archive/2026/CHG-0005-single-active-season/
verified_at: 2026-08-30
source_revision: HEAD
context_scope: existing
created: 2026-08-30
updated: 2026-08-30
---

# Enforce Single Active Season Rule

## Intent

Ensure the Missions Platform never has more than one `ACTIVE` season at a time, so standings, offers, and leaderboards always resolve to a single current competitive window.

## Business Value

- Avoid ambiguity when publishing mission offers or confirming evidence.
- Simplify the mission board and player UI by pointing to one active season.
- Prevent accidental overlapping standings.

## User Story

- **US-SCC-005**: As an operator, I want there to be only one active season, to avoid overlapping offers and scores.

## Why L2

Adds a business rule to the domain, requires a new query port (`ListActive`), and affects the response of the `POST /api/v1/seasons` endpoint. It does not change identity, authorization or external communication.

## Scope

- Domain: add `ErrActiveSeasonAlreadyExists` and a domain check that rejects creation when another season is `ACTIVE`.
- Application: `CreateSeasonHandler` queries `SeasonReader.ListActive()` and validates before saving.
- Infrastructure: `SeasonStore.ListActive()` uses DynamoDB `PK=SEASON` and `State=ACTIVE`.
- Tests: unit domain, unit application, integration HTTP.

## Non-Goals

- Force the closing of a season to automatically create the next one.
- Change the status of the existing season.
- Modify `ConfirmObjective` or `AcceptMission` to validate active season in this change.

## Acceptance Summary

- Creating a season when none is `ACTIVE` succeeds with `201 Created`.
- Creating a season when one is `ACTIVE` returns `ACTIVE_SEASON_EXISTS` with `422`.
- A `FREEZING`, `FINALIZED`, or `ARCHIVED` season does not block creation.

## Current Outcome

Gate 2 approved on 2026-08-30. T-001, T-002, and T-003 completed with RED/GREEN/BLUE evidence.
`gofmt`, `go vet`, and `go test -race ./...` all pass.
`tools/validate-sdd-change.sh` is not present in this repository.
The active change is ready for `common-sdd-verify-spec` or user review.
