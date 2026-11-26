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
- [x] Follow-up RED — root cause of the failed staging deploy job: passing
      `SeedDeploymentRoleName` made the stack manage the deploy role's own
      IAM policy; denied (`iam:GetRolePolicy`). TEST-STG-CATALOG-IAM-001
      planned as boundary test + L2 operational map before RED, then added
      after the SharedEventsTopicArn assert; run fails with
      `CONTRACT FAIL: staging SAM deploy still asks CloudFormation to manage
      the deployment role policy`.
- [x] Follow-up Gate 3-BOUNDARY — verified: test reads the workflow only,
      `aws`/`sam` are function doubles, no shared writes.
- [x] Follow-up GREEN — removed the staging `SEED_DEPLOYMENT_ROLE_ARN` env
      binding and `SeedDeploymentRoleName` override (workflow + unused test
      export). Production, template default `''`, `HasSeedDeploymentRole`,
      and post-deploy seed unchanged. `bash tools/test-workflow-contract.sh`
      → `Workflow contract OK`; `bash -n` clean; `git diff --check` clean;
      `git diff 9058729 -- ci-cd.yml test-workflow-contract.sh` shows only
      the intended removals, previous SNS fix preserved.
- [x] BLUE — no refactor needed; recorded security finding: staging managed
      policy already grants `dynamodb:*` on `*` (pre-existing, out-of-band);
      no new grant, no scoped-rights claim, narrowing needs separate approval.
- [ ] Final gate / consolidation — not claimed; pending lead approval.
