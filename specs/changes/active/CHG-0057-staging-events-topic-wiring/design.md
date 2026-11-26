---
change_id: CHG-0057
artifact: design
status: implemented
updated: 2026-11-26
---

# Design — Staging events topic wiring

## Domain / Application

- Not affected. No Go code changes.

## Infrastructure (template.yaml)

- No change. The `SharedEventsTopicArn` parameter (template.yaml line 40) is
  already declared and already feeds `SNS_TOPIC_ARN` and the Lambda's
  `sns:Publish` policy. Supplying a real ARN at deploy time replaces the
  empty-string default only in the staging stack.

## Boundary (CI/CD workflow)

- `deploy-staging` → `SAM deploy` step explicitly selects Bash for the
  Bash-specific `[[ =~ ]]` condition.
- Preamble resolves the export with
  `aws cloudformation list-exports --region us-east-2` and a JMESPath filter on
  the export name — the deploy role already holds `cloudformation:ListExports`
  (handled outside this change; IAM verified).
- Validation checks the region, account-ID format and topic name before SAM
  can mutate the stack. It does not compare account IDs; the value is taken
  from the named export in the authenticated deployment account.
- The parameter override is appended last so the diff stays additive and the
  production block is untouched.

## Test strategy

- `tools/test-workflow-contract.sh`, TEST-STG-EVENTS-001: the staging
  `SAM deploy` run body is extracted with `sed`, `${{ ... }}` expressions are
  replaced by fixtures, and the body is executed under `bash -e` with `aws`
  and `sam` function doubles — no shared state, no AWS calls, no writes to
  the workflow file.
- Assert 1: on a valid export, `sam` receives the exact
  `ParameterKey=SharedEventsTopicArn,ParameterValue=<arn>` line.
- Assert 2: on `None`, the step exits 1 with the exact stderr message and
  `sam` output is absent (fail-before-deploy).

## Verification commands

- `bash tools/test-workflow-contract.sh` — RED then GREEN evidence.
- `bash -n tools/test-workflow-contract.sh` — script syntax.
- `git diff --check` — whitespace/indent integrity.
- `actionlint` — not installed in this environment; skipped.
- Full repo validator intentionally not run: it starts/stops Docker and runs
  unrelated Go checks; this scope is local workflow/bash only.
