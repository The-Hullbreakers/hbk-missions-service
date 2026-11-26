---
change_id: CHG-0059
artifact: tasks
status: implemented
updated: 2026-11-26
---

# Tasks — Staging seasons API repair

- [x] Gate 1 — plan approved (two local staging fixes; no AWS/GitHub
      mutations, commit/push/deploy/seed; IAM denial diagnosis settled, SAM
      owns table, no new permissions).
- [x] Active delta artifacts written (change.md, delta.md, design.md,
      tasks.md) — Gate 2. L2 operational boundary; `canonical_action:
      not_affected`.
- [x] RED (route) — TEST-STG-CURRENT-SEASON-ROUTE-001 added to
      `tools/test-workflow-contract.sh`; `bash tools/test-workflow-contract.sh`
      failed with `template missing current-season GET route` before the
      template edit.
- [x] RED (managed table) — `tests/integration/composition/
      managed_table_test.go` + behavior-free `UseExistingTable` config field
      (harness contract only);
      `go test ./tests/integration/composition -run TestGivenManagedTableWhenSeasonRequestArrivesThenLambdaStartsWithoutProvisioning -count=1`
      failed at init with the denied `CreateTable`, reproducing the observed
      staging panic.
- [x] Gate 3-BOUNDARY — verified: template check is a static grep; Go test
      owns a loopback `httptest` stub (CreateTable denied, call counter), no
      Docker, no shared writes, no IAM/AWS calls.
- [x] GREEN — `MissionsCurrentSeasonRoute` added before
      `MissionsLeaderboardRoute`; `if !cfg.UseExistingTable` guard around
      `CreateMissionsTable` in `server.go`; `UseExistingTable: true` in
      `cmd/lambda/main.go` Config literal.
- [x] Verification — both focused commands GREEN, `go build ./cmd/lambda
      ./cmd/hbk-missions-service`, `git diff --check`, gofmt on touched Go
      files. `sam validate --template-file template.yaml --region us-east-2`
      passed (basic SAM validation). `sam validate --lint` reported one
      pre-existing warning W1030 (`Ref SharedEventsTopicArn` does not match
      the ARN pattern when resolved — the parameter default is `''`; the
      policy line predates this change and is untouched by it). Full repo
      validator intentionally not run: it manipulates Docker for this
      no-Docker scope.
- [x] BLUE — no refactor needed; security/cost noted: existing route auth
      reused, no IAM changes, one fewer unauthorized init call.
- [ ] Final gate / consolidation — not claimed; pending lead approval.
