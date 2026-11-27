---
change_id: CHG-0058
title: Add the six-mission bilingual competitive LIVE catalog
status: archived
risk: L1
canonical_action: not_affected
canonical_paths:
  - specs/capabilities/mission-board/spec.md
execution_mode: sequential
validation_status: passed
verified_at: 2026-11-26
source_revision: reviewed working-tree catalog diff based on 43bb599
archive_path: specs/changes/archive/2026/CHG-0058-competitive-live-catalog/
context_scope: none
---

# Competitive LIVE catalog

User request authorizes creating an explicit six-mission competitive catalog
(`seeds/competitive-missions.json`) with its versioned English manifest
(`seeds/translations/competitive-missions-v1.en.json`) and adding it to the
current local active season through the approved separate local process. No
API, domain, UI, Go, pipeline, or default-bootstrap change; existing
REQ-MB-001/013/015/017 behavior is reused as-is.

## Acceptance scenarios

- Six offers (MINING ×2, ARENA_COMMANDER ×2, SALVAGE ×1, HAULING ×1)
  are published to the requested local active season, `local-training`.
- The user sees English and Spanish content for every mission with the same
  merits (per-unit ×3 units + completion = 200/300/500/400/300/300).
- The production selector rejects the catalog (`environments: ["staging"]`);
  `staging` is the local render/load selector only — never the default seed.

## Layer gates / verification

Domain and application: not affected. Boundary Gate 3 (catalog seed scope):
- RED: `python3 -B -m unittest discover -s tests/unit/seeds -p
  test_competitive_mission_catalog.py` → 3 failures, all
  `FileNotFoundError: seeds/competitive-missions.json` — proves the catalog,
  not the harness/imports, is the missing behavior.
- GREEN: same command → 3 tests pass; all 33 seeds unit tests pass; offline
  render produces 18 items (3 rows × 6 codes) with identical `ContentEn`
  triplets and preserved Spanish fields; production render exits non-zero
  with "Catalog is not enabled for this environment".
- BLUE: none required — content-only addition; no production code changed.
- Local verification: `--dry-run` and `--apply` with `--require-active-season`
  targeted `hbk-missions-local` on `http://localhost:8000` and selected
  `local-training`. Consistent reads matched all 18 rendered items exactly
  (six definitions, six versions, six offers), including both languages.

Manual moderation (video evidence, duplicate-result rules) stays outside the
service; recorded in catalog Rules text only.

## Final Review

The requester approved archival on 2026-11-26. The scoped SDD policy
validator passed at L1 against the catalog, translation, seed README, seed
test and this change record. No canonical behavior changed. The existing seed
workflow is documented in `seeds/README.md`, so no AI context snapshot is
needed. LIVE eligibility and duplicate-evidence checks remain manual.

## Rollback

Leave unapplied; no automatic cleanup. Removal of applied rows is a separate
approved local administration process. Accepted contracts preserve their
terms; definitions are immutable versioned additions.
