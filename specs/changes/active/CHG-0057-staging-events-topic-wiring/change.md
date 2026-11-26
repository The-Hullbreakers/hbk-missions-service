---
change_id: CHG-0057
title: Wire the staging deployment to the shared events topic export
status: locally_verified
risk: L2
canonical_action: not_affected
canonical_paths:
  - specs/capabilities/seasonal-category-competition/spec.md
execution_mode: sequential
---

# Staging events topic wiring

User request authorizes repairing the staging deployment workflow so that
`sam deploy` receives the shared infrastructure events topic ARN resolved from
the CloudFormation export `hbk-shared-infrastructure-stg-EventsTopicArn`.
Canonical code behavior is not changed; this is operational configuration for
the existing `SharedEventsTopicArn` template parameter already consumed by the
SNS publisher boundary (see REQ-SCC-018 placement-awarded events).

## Acceptance scenarios

- Staging export present: the `SAM deploy` step resolves
  `hbk-shared-infrastructure-stg-EventsTopicArn` via
  `aws cloudformation list-exports --region us-east-2` and passes it as
  `ParameterKey=SharedEventsTopicArn` to `sam deploy`.
- Export missing or not matching
  `arn:aws:sns:us-east-2:<account>:hbk-domain-events-stg`: the step fails with
  `Missing or invalid staging events topic export` before `sam` runs.
- Production deploy: unchanged — no export lookup, no new parameter.

## Layer gates / verification

Domain and application: not affected. Boundary BDD: given the staging export
exists, SAM gets the exact ARN; if missing/invalid, stop before `sam`
(TEST-STG-EVENTS-001 in `tools/test-workflow-contract.sh`, isolated `aws`/`sam`
doubles, no shared writes). IAM: deployment role already allows
`cloudformation:ListExports`; verified matching live export/topic. No template
change — `SharedEventsTopicArn` is an existing parameter.
Rollback: revert the workflow step; the parameter is additive.
