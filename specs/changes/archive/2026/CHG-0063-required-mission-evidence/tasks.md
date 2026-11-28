---
change_id: CHG-0063
artifact: tasks
status: implemented
updated: 2026-11-27
---

# Tasks — Required mission evidence

- [x] Gate 1 — plan approved by the user's literal `si` on the cross-repo
      plan; backend wave only, no commits, pushes, deploys, or server start;
      preserve dirty prior work (`composition/server.go` untouched).
- [x] Active delta artifacts written (change.md, delta.md, design.md,
      tasks.md) — Gate 2. L2; `canonical_action: modify`,
      `change_action: create_delta`, canonical path
      `specs/capabilities/mission-contract-fulfillment/spec.md`
      (MODIFIED REQ-MCF-003/015/019 + HTTP contract lines).
- [x] Behavioral RED replay (post-implementation `go test -overlay`
      against `git show HEAD` copies of `mission_contract_review.go` and
      `confirm_objective.go`, new `errors.go` retained so tests compile):
      `TestConfirmObjectiveHandler_WithMedia_RequiresMedia` fails with
      `ErrEvidenceMediaInvalid` instead of `ErrContractEvidenceRequired`;
      `TestSubmitContract_RejectsWithoutStagedEvidence`,
      `TestContractSubmit_RejectsInProgressWithoutEvidence` and
      `TestContractSubmit_RejectsLegacyOnlyStagedCredit` fail on `nil`
      error (old Submit transitions straight to pending).
- [x] Gate 3 — scoped review of the RED evidence before GREEN. The
      initial implementation-time RED for the unit partitions was
      compile-stage only (missing sentinel); the behavioral replay above
      is the definitive RED record for this change.
- [x] GREEN (domain + application) — `ErrContractEvidenceRequired`,
      `hasStagedEvidence` guard in `Submit`, empty-media check in
      `HandleWithMedia`, media-less `Handle`/`NewConfirmObjectiveHandler`
      removed; scoped run green:
      `go test ./tests/unit/missions/application ./tests/unit/missions/domain
      -run 'Test(ConfirmObjective|SubmitContract|ContractSubmit)'`.
- [x] Behavioral RED replay (boundary) — overlay additionally restoring
      HEAD `handler.go`/`errors.go` under the same integration tests:
      `TestConfirmEvidenceEndpoint_RejectsMissingMedia` and
      `RejectsEmptyMedia` got 201 vs 422;
      `TestContractReview_SubmitEndpointRejectsWithoutEvidence` got 200
      vs 422 (`DYNAMODB_TEST_ENDPOINT=http://localhost:55001`).
- [x] GREEN (boundary) — single `HandleWithMedia` call, `EVIDENCE_REQUIRED`
      public code;
      `go test ./tests/integration/http -run
      'Test(ConfirmEvidenceEndpoint|ContractReview_SubmitEndpoint)'` green.
- [x] Fixture migration — media-backed success fixtures across unit,
      http-integration, dynamodb fence and composition e2e suites
      (`ConfirmWithMedia`, `storedEvidence`, `blobStoreFake`,
      `stubEvidenceReader`); legacy no-media credit retained only in the
      TEST-0063-DOM-001 rejection fixture via `NewMissionContractFromState`.
- [x] Test hygiene — Arrange/Act/Assert markers added to all newly
      written rejection/success cases (APP-001/002, DOM-001 x3,
      HTTP-001 x2, HTTP-002).
- [x] Full verification —
      `go test ./tests/unit/...` all green;
      `go test ./tests/integration/http` green;
      `go test -p=1 -tags integration ./tests/integration/composition` green
      (DynamoDB local reused at http://localhost:55001 via
      `DYNAMODB_TEST_ENDPOINT`);
      `git diff --check` clean; `gofmt -l` clean; `go vet -tags integration`
      clean.
- [x] Mutation probes (actual, executed via `go test -overlay` against
      temp mutated copies — worktree production untouched):
      mutant removing the `mediaURI == ""` guard is killed by
      `TestConfirmObjectiveHandler_WithMedia_RequiresMedia`
      (`ErrEvidenceMediaInvalid` in chain instead);
      mutant removing the `hasStagedEvidence` guard is killed by
      `TestContractSubmit_RejectsInProgressWithoutEvidence`,
      `TestContractSubmit_RejectsLegacyOnlyStagedCredit` and
      `TestSubmitContract_RejectsWithoutStagedEvidence` (nil error, wrong
      transition).
- [x] Coverage gate — final run
      `DYNAMODB_TEST_ENDPOINT=http://localhost:55001
      S3_TEST_ENDPOINT=http://localhost:14566 go test -p=1 -tags integration
      -skip 'TestSeasonScenarioReaders'
      -coverpkg=./internal/missions/... -coverprofile=backend.cover
      ./tests/unit/... ./tests/integration/...`: all packages pass,
      total `90.0%` of statements (≥90 met). Touched code:
      `HandleWithMedia` 96.2%, `Submit`/`hasStagedEvidence` 100%,
      `ConfirmEvidenceHandler` 88.9%, `PublicCode` 82.7%.
      Dedicated LocalStack S3 brought up as
      `docker compose -f docker-compose.test.yml -p hbk-missions-test up -d
      localstack-s3` (separate test project, port 14566, test buckets only).
- [x] Exclusions/blockers — `TestSeasonScenarioReaders`
      (dynamodb infra) skipped: it seeds `--apply` into the user's
      preview table `hbk-missions-preview-mid-season` at
      `http://localhost:8000` (hard-coded, ignores
      `DYNAMODB_TEST_ENDPOINT`) and the preview rejected the seed;
      writing there is out of scope. Preexisting environment dependency,
      unrelated to this change.
- [x] Human-approved consolidation and archive completed; see change.md
      for scope and limitations.
