---
change_id: CHG-0060
artifact: tasks
status: implemented
updated: 2026-11-26
---

# Tasks — Lambda stage-prefix routing

- [x] Gate 1 — plan approved (local boundary fix only; no AWS/GitHub changes,
      no commit/push/deploy; preserve unrelated worktree changes).
- [x] Active delta artifacts written (change.md, delta.md, design.md,
      tasks.md) — Gate 2. L2 operational; `canonical_action: not_affected`;
      cites the existing season-current contract.
- [x] Behavior-free harness — `internal/missions/interface/lambdahttp/
      proxy.go` `NewProxy` created as a pass-through to
      `httpadapter.NewV2(handler).ProxyWithContext`; `cmd/lambda` not yet
      wired before RED.
- [x] RED — `tests/integration/composition/lambda_stage_routing_test.go`
      TEST-LAMBDA-STAGE-ROUTING-001;
      `go test ./tests/integration/composition -run TestGivenGatewayStageWhenInvokingSeasonRouteThenOnlyExactPrefixIsNormalized -count=1`
      failed on the named-stage and fallback cases (404 vs expected 401),
      proving detection before normalization.
- [x] Gate 3-BOUNDARY — verified: test owns synthetic events and the real
      composition root with `UseExistingTable` (zero AWS calls), no Docker,
      no shared writes, no IAM/template/auth changes.
- [x] GREEN — stage-exact prefix normalization implemented in `proxy.go`
      (`strings` import); `cmd/lambda/main.go` wired to
      `missionlambda.NewProxy(handler)` (httpadapter import replaced with
      the aliased seam import).
- [x] Verification — focused tests GREEN, `go build ./cmd/lambda`,
      `bash tools/test-workflow-contract.sh`, `git diff --check`, gofmt on
      touched files. Full Docker validator intentionally not run for this
      boundary-only scope.
- [x] BLUE — no refactor needed; exact-segment matching cannot cross stages;
      auth unchanged.
- [ ] Final gate / consolidation — not claimed; pending lead approval.
