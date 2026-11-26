---
change_id: CHG-0014
title: Create CI/CD pipeline and shared API Gateway deployment
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-board/spec.md
  - specs/capabilities/mission-contract-fulfillment/spec.md
  - specs/capabilities/seasonal-category-competition/spec.md
change_path: specs/changes/archive/2026/CHG-0014-ci-cd-pipeline-shared-api-gateway/
execution_mode: sequential
primary_workflow: WORKFLOW-COMMON_CREATE_GITHUB_ACTIONS_PIPELINE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
context_scope: existing
created: 2026-09-22
updated: 2026-09-22
---

# Create CI/CD Pipeline and Shared API Gateway Deployment

## Intent

Add a GitHub Actions pipeline that enforces the SDD policy gate, runs the canonical `unit-tests` and `integration-tests` jobs, and deploys the service to the shared HBK API Gateway in `us-east-2` on `staging` and `main` pushes, following the `hbk-sc-tools-service` reference.

## Business Value

- Every pull request is validated by the same gates used across HBK services.
- The missions service becomes reachable through the shared API Gateway without manual deployment.

## User Stories

- **US-OPS-001**: As a maintainer, I want pull requests validated by `sdd-policy`, `unit-tests`, and `integration-tests`, so regressions are blocked before merge.
- **US-OPS-002**: As an operator, I want pushes to `staging`/`main` to deploy the Lambda behind the shared API Gateway, so the service is available per environment.

## Why L2

No domain or application behavior changes. The change adds a Lambda boundary entry, a SAM template, a Makefile, and a workflow. Deploy credentials and a new public trust boundary justify the security gate; no mutation or critical-E2E gate is required because no business logic changes.

## Scope

- **Boundary**: add `cmd/lambda/main.go` wrapping `composition.NewServer` via `aws-lambda-go-api-proxy`.
- **Composition/IaC**: add `template.yaml` (DynamoDB table, `MissionsAPIFunction`, shared API Gateway integration/routes via `Fn::ImportValue`), `Makefile` with `build-MissionsAPIFunction` and `unit-tests` targets.
- **CI**: add `.github/workflows/ci-cd.yml` with `sdd-policy`, `unit-tests`, `integration-tests`, `build`, `deploy-staging`, `deploy-production`.
- **Tests**: pipeline contract test asserting required jobs, permissions, and triggers.

## Layer Gates

- `LAYER-GATE-DOMAIN`: not_affected — no domain changes.
- `LAYER-GATE-APPLICATION`: not_affected — no application changes.
- `LAYER-GATE-INFRASTRUCTURE`: not_affected — no adapter changes.
- `LAYER-GATE-INTERFACE`: not_affected — existing HTTP handlers reused.
- Composition/boundary only: Lambda entry, SAM template, workflow.

## Gates

- Gate 1 (plan approval): approved by user 2026-09-22 (branches `staging`/`main`, `sdd-policy` runs full `validate-sdd-change.sh`, deploy to shared API Gateway in `us-east-2`).
- Gate 2 (test approval): pipeline contract test `tools/test-workflow-contract.sh` approved before RED.
- Gate 3-BOUNDARY: RED evidence recorded before workflow/template production changes.
- Security gate: selected — new deploy credentials and public API Gateway trust boundary.
- Coverage/mutation/critical-E2E: not selected — no production logic changes.
