---
change_id: CHG-0057
artifact: tasks
status: implemented
updated: 2026-11-26
---

# Tasks — Staging events topic wiring

- [x] Gate 1 — plan approved (local workflow/test only; no AWS/GitHub
      mutations, no deploy/seed/commit/push; IAM and table import already
      complete).
- [x] Active delta artifacts written (change.md, delta.md, design.md,
      tasks.md) — Gate 2.
- [x] RED — TEST-STG-EVENTS-001 added to `tools/test-workflow-contract.sh`
      before `echo "Workflow contract OK"`; run fails with
      `CONTRACT FAIL: staging SAM deploy does not receive the exported
      events topic ARN` (missing parameter, not a syntax error).
- [x] Gate 3-BOUNDARY — verified: extraction maps only the staging
      `SAM deploy` run body; `aws`/`sam` are isolated function doubles with no
      shared writes; the workflow file is only read, never written.
- [x] GREEN — `.github/workflows/ci-cd.yml` staging `SAM deploy` step:
      `shell: bash` added; export-resolution + regex-validation preamble
      inserted at the top of `run: |`;
      `ParameterKey=SharedEventsTopicArn,ParameterValue="$shared_events_topic_arn"`
      appended as last parameter override. Production block, template, and IAM
      untouched.
- [x] Verification — `bash tools/test-workflow-contract.sh` →
      `Workflow contract OK`; `bash -n tools/test-workflow-contract.sh` clean;
      `git diff --check` clean. `actionlint` not installed — skipped.
- [ ] Final gate / consolidation — not claimed; pending lead approval.
