---
change_id: CHG-0061
artifact: tasks
status: implemented
updated: 2026-11-26
---

# Tasks — Staging season statistics route

- [x] Gate 1 — plan approved (template-only staging repair; no commits,
      pushes, external writes, Docker, or server; preserve unrelated worktree
      changes).
- [x] Active delta artifacts written (change.md, delta.md, design.md,
      tasks.md) — Gate 2. L2 boundary; `canonical_action: not_affected`;
      cites canonical REQ-SCC-017.
- [x] RED — TEST-STG-SEASON-STATISTICS-ROUTE-001 added to
      `tools/test-workflow-contract.sh` after the current-season check;
      `bash tools/test-workflow-contract.sh` failed with
      `template missing season statistics GET route or shared integration`.
- [x] Gate 3-BOUNDARY — verified: static `sed`/`grep` of the template only,
      no AWS calls, no shared writes, no Go edits.
- [x] GREEN — `MissionsSeasonStatisticsRoute` inserted before
      `MissionsFreezeSeasonRoute`; `bash tools/test-workflow-contract.sh` →
      `Workflow contract OK`;
      `go test ./tests/unit/missions/application -run '^TestGetSeasonStatistics' -count=1` → pass;
      `sam validate --template-file template.yaml --region us-east-2` →
      valid; `git diff --check` clean.
- [x] BLUE — none needed; root cause recorded (route never declared), auth
      unchanged (Go middleware), rollback = remove route block.
- [ ] Final gate / consolidation — not claimed; pending lead approval.
