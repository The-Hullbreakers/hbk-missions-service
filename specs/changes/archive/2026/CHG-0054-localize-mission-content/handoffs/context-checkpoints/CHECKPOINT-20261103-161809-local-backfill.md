---
handoff_id: HANDOFF-0054-CONTEXT-002
handoff_title: Continue local publication of reviewed English mission content
change_id: CHG-0054
change_title: Localize mission content for Spanish and English readers
capability_id: CAP-MISSION-BOARD
capability_title: Mission Board
artifact_id: ART-0054-CONTEXT-HANDOFF
artifact_title: Context continuation handoff for localized mission content
checkpoint_id: CHECKPOINT-20261103-161809
checkpoint_title: Pause after local backfill invoker GREEN/BLUE
context_used_percent: "estimated >=60; host meter unavailable"
current_task_id: T-0054-021
current_task_title: Implement and refactor the local backfill invoker
next_task_id: T-0054-022
next_task_title: Publish approved English to local history safely
status: ready_for_next_context
---

# Context Handoff

This is a continuation checkpoint, not a release or migration approval.

## Current Task

- Objective: T-0054-021 — Implement and refactor the local backfill invoker.
- Done when: approved local-only reader and CLI pass focused unit,
  private-table integration, and existing Go backfill regression.
- Current step: done; T-0054-022 remains open.
- Risk level: L3 existing CHG-0054; offline data mutation boundary.
- Owner/track: one mission-board local publication track.

## Completed In This Context

- Gate 1 approved local-only invoker amendment on 2026-11-03.
  Read-only inspection of `http://localhost:8000` /
  `hbk-missions-local` found 36 current definitions, 36
  versioned definitions, 36 active offers, all Spanish-only,
  and zero contracts. Ordinary seed apply cannot add English
  to these existing versions.
- Gate 2 approved TEST-SEED-LOCAL-001/002 RED. Unit build failed
  solely for missing manifest reader/endpoint validator; five
  integration cases failed solely for missing Go CLI.
- Gate 3-LOCAL-BACKFILL approved for manifest reader and
  invoker only. Added
  `internal/missions/infrastructure/dynamodb/reviewed_translation_manifest.go`
  (99 lines), `cmd/local-translation-backfill/main.go` (35),
  and `cmd/local-translation-backfill/publish.go` (95).
  Added tests in `tests/unit/missions/infrastructure/dynamodb/reviewed_manifest_test.go`
  (137) and
  `tests/integration/infrastructure/dynamodb/test_local_translation_publish.py`
  (128). The CLI requires explicit HTTP loopback/port, table,
  all 36 reviewed source-bound records, and previews every
  candidate before any requested apply. Existing Go backfill
  owns conditional `ContentEn`-only updates.
- Updated `delta.md`, `design.md`, `tasks.md`,
  `traceability.yaml`, `verification.md`, `rollout.md`,
  `security-review.md`, and `change.md`. Canonical specs
  were not modified.

## Evidence

- Unit: `go test -count=1 ./tests/unit/missions/infrastructure/dynamodb -run '^TestReviewedTranslationManifest' -v`
  passed four methods and their subcases.
- CLI: `DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 python3 -B -m unittest discover -s tests/integration/infrastructure/dynamodb -p 'test_local_translation_publish.py' -v`
  passed 5/5. Only new private port-18000 tables received test writes.
- Backfill regression:
  `DYNAMODB_TEST_ENDPOINT=http://localhost:18000 go test -count=1 -tags=integration -p=1 ./tests/integration/infrastructure/dynamodb -run '^TestBackfill' -v`
  passed 12 top-level tests and five drift subtests. Tables
  left intact; no DeleteTable and no AWS call.
- GREEN/BLUE recorded in `verification.md`; Go files gofmt'd.
  The general `tools/validate-sdd-change.sh` is NOT run: it
  invokes `docker compose down --remove-orphans` and would
  remove existing containers without specific approval.

## Continuation

- Exact first action: read this file, check `git status`, and
  run the local CLI in READ-ONLY mode on the already confirmed
  app table: `go run ./cmd/local-translation-backfill --endpoint-url http://localhost:8000 --table hbk-missions-local --translations-dir seeds/translations --dry-run`.
  Check the actual eligible identities, count, unresolved
  reasons, and `applied: false`; do not assume the older 108
  observation still holds. Record the fresh result in
  `verification.md` / `rollout.md`.
- If the preview is fully eligible and has zero unresolved,
  ask the user for explicit approval of `--apply` on precisely
  `http://localhost:8000` / `hbk-missions-local`. After approval,
  capture before/after snapshots and verify only `ContentEn`
  changed; otherwise stop and resolve conflicts via a new gate.
- No staging/production/shared AWS access, CI/IAM deployment,
  accepted-contract overwrite, table deletion, or app-table
  apply is authorized yet.

## Risks And Blockers

- The backfill is per-record, not globally atomic. On a race,
  report partial updates and re-preview before any approved
  retry; never delete or overwrite Spanish to roll back.
- `CHG-0051` still blocks the ordinary full-tree boundary build;
  broader L3 security, coverage, mutation, E2E, documentation,
  and cleanup gates remain open.
- No blocker for the next READ-ONLY CLI preview. Exact-table
  data write still needs its own user decision.

## Resume Checklist

- [ ] Read this checkpoint and CHG-0054 active artifacts.
- [ ] Check worktree status; preserve unrelated edits.
- [ ] Run only the read-only local CLI preview first.
- [ ] Update evidence; ask for exact-table apply approval.
