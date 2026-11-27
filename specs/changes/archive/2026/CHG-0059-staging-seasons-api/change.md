---
change_id: CHG-0059
title: Restore the staging current-season API and stop table provisioning at init
status: archived
validation_status: passed
verified_at: 2026-11-26
source_revision: 212088dec8c29556a3c12ca2f2833d51799156f2
archive_path: specs/changes/archive/2026/CHG-0059-staging-seasons-api/
context_scope: none
risk: L2
canonical_action: not_affected
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
execution_mode: sequential
---

# Staging seasons API repair

User request authorizes two operational fixes so the staging deployment serves
the existing seasonal-category-competition current-season contract. Canonical
behavior is unchanged; this restores the deployed boundary to match the
already-implemented, authenticated Go handler. Diagnosis is settled:

- Live `OPTIONS` returns correct CORS (204); live `GET
  /api/v1/missions/seasons/current` returned a generic API Gateway 404 — the
  route was never wired in `template.yaml` despite the authenticated handler
  existing.
- Live `GET /api/v1/seasons` returned 500 with proper CORS; CloudWatch shows a
  Lambda init panic: `CreateTable` AccessDenied. SAM owns the existing
  `hbk-missions-stg` table; the role must NOT gain provisioning permission.

Non-goals: health-path mismatch and other missing routes are out of scope.

## Acceptance scenarios

- `GET /api/v1/missions/seasons/current` has an `AWS::ApiGatewayV2::Route`
  wired to the shared Lambda integration in `template.yaml`
  (`MissionsCurrentSeasonRoute`, `RouteKey: GET
  /api/v1/missions/seasons/current`).
- Lambda startup on a managed (SAM-owned) table performs no `CreateTable`
  call: with `UseExistingTable`, `NewServer` builds the handler without
  touching DynamoDB, and a current-season request reaches auth
  (`401 Unauthorized` unauthenticated) instead of an init panic.
- Local/dev behavior unchanged: without `UseExistingTable`, `NewServer` still
  creates the table and fails on `CreateTable` errors as before.
- No production deployment is in scope. If released to production later, the
  shared SAM template and Lambda entry point get the same fixes without IAM
  or authentication changes.

## Layer gates / verification

Domain and application: not affected. Boundary BDD:
TEST-STG-CURRENT-SEASON-ROUTE-001 asserts the template route contract in
`tools/test-workflow-contract.sh`; TEST-STG-MANAGED-TABLE-001 is a Go
composition integration test with a loopback `httptest` DynamoDB stub that
denies `CreateTable` — the Lambda must start (no provisioning call) and the
request must reach authentication. Security/cost: existing route auth
unchanged, no IAM changes, one skipped API call at init. Rollback: revert the
route block and the config flag.

## Final outcome — 2026-11-26

- `validation_status: passed`; explicit user approval received for final
  closure/archive and the commit/push/PR request to staging.
- Verified commit: `212088dec8c29556a3c12ca2f2833d51799156f2` on
  `The-Hullbreakers/hbk-missions-service`, byte-identical to the prior local
  implementation `261bd4f` for all CHG-0059 paths.
- Verified CI run:
  `https://github.com/The-Hullbreakers/hbk-missions-service/actions/runs/45893405142`
  — all jobs successful: SDD validation, unit tests, integration tests, SAM
  build, staging deploy; production job skipped as designed.
- Independent security review of the exact implementation found no blockers:
  the current-season handler retains existing in-Lambda authentication, no
  IAM changes were introduced, and the loopback integration test asserted
  zero DynamoDB calls at init and `401 Unauthorized` for unauthenticated
  requests.
- Closure commands (this worktree): `git diff --check` clean;
  `bash tools/test-workflow-contract.sh` OK; focused Go composition test
  green; seeds unittest discovery green. No runtime, Docker, deployment,
  seeding, or shared-AWS actions performed by this closure.
- Full-repo local validator intentionally skipped (it manipulates Docker;
  not required for archival packaging). CI evidence is the recorded run;
  new PR CI pending the lead's git actions.
- `canonical_action: not_affected` retained — the existing
  seasonal-category-competition contract (REQ-SCC-014) is unchanged; this
  restored deployed wiring only. No canonical spec update and no durable
  context snapshot (operational restoration, no reusable knowledge).
