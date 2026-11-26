# Tasks — CHG-0014 Create CI/CD pipeline and shared API Gateway deployment

| Task | Title | Wave | Ownership | Depends on | Status |
| --- | --- | --- | --- | --- | --- |
| T-001 | Add pipeline contract test asserting required jobs and permissions | 1 | agent | Gate 2 | done |
| T-002 | Add Makefile with `build-MissionsAPIFunction` and `unit-tests` targets | 2 | agent | T-001 RED | done |
| T-003 | Add `cmd/lambda/main.go` Lambda HTTP adapter entry | 2 | agent | T-001 RED | done |
| T-004 | Add `template.yaml` with table, function, shared API Gateway routes | 2 | agent | T-001 RED | done |
| T-005 | Add `.github/workflows/ci-cd.yml` with canonical jobs and deploy | 2 | agent | T-001 RED | done |
| T-006 | Run contract test GREEN, `sam validate`, and `validate-sdd-change.sh` | 3 | agent | T-002..T-005 | done |
| T-007 | Security gate review of deploy credentials and trust boundary | 3 | agent | T-006 | done |
| T-008 | Update docs and consolidate | 4 | agent | T-007 | pending |

## Outcomes

- T-001: RED confirmed — `CONTRACT FAIL: .github/workflows/ci-cd.yml does not exist`. GREEN after T-005.
- T-002..T-005: `Makefile`, `cmd/lambda/main.go`, `template.yaml`, `.github/workflows/ci-cd.yml` created; `go build` and `gofmt`/`go vet` clean.
- T-006: `tools/test-workflow-contract.sh` passes; `sam validate` passes; unit suite green. Integration suite not run locally — Docker daemon unavailable; CI runs it via `docker-compose.test.yml`.
- T-007: no Critical/High findings — OIDC `role-to-assume` (no long-lived keys), `id-token: write` only on deploy jobs, `persist-credentials: false`, deploy gated on `push`/`workflow_dispatch` (never PR), IAM policies scoped to `hbk-missions-*` resources, Lambda `DynamoDBCrudPolicy` on own table, API Gateway invoke permission scoped to the shared API ARN.
- Provisioned: S3 buckets `hbk-missions-deploy-{staging,production}-us-east-2` (versioned), IAM roles `GitHubActions-{Staging,Production}-HBKMissionsService` (GitHub OIDC trust for `repo:xChiro/hbk-missions-service:*`), policies `GitHubActions-Deploy-HBKMissionsService-{Staging,Production}`, GitHub environments `staging`/`production`, secrets `AWS_ROLE_ARN_*` and `OIDC_*`.
