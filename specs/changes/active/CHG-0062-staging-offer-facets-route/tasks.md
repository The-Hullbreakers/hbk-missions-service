---
change_id: CHG-0062
artifact: tasks
status: implemented
updated: 2026-11-26
---

# Tasks — Staging offer facets route

- [x] Gate 1 — plan approved (template-only staging repair; no commits,
      pushes, cloud writes, Docker, or server; preserve dirty prior work).
- [x] Active delta artifacts written (change.md, delta.md, design.md,
      tasks.md) — Gate 2. L2 boundary; `canonical_action: not_affected`;
      cites mission-board facets contract (spec.md:319).
- [x] RED — TEST-STG-OFFER-FACETS-ROUTE-001 added to
      `tools/test-workflow-contract.sh` after the statistics check;
      `bash tools/test-workflow-contract.sh` failed with
      `template missing offer facets GET route or shared integration`.
- [x] Gate 3-BOUNDARY — verified: static `sed`/`grep` of the template only,
      no AWS calls, no shared writes, no Go edits.
- [x] GREEN — `MissionsOfferFacetsRoute` inserted between
      `MissionsListOffersRoute` and `MissionsListContractsRoute`;
      `bash tools/test-workflow-contract.sh` → `Workflow contract OK`;
      `go test ./tests/unit/missions/application ./tests/unit/missions/interface/http -run '^TestListOfferFacets' -count=1` → pass;
      `sam validate --template-file template.yaml --region us-east-2` →
      valid; `git diff --check` clean.
- [x] BLUE — none needed; root cause recorded (route never declared), auth
      unchanged (Go middleware), rollback = remove route block.
- [ ] Final gate / consolidation — not claimed; pending lead approval.
