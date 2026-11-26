---
change_id: CHG-0057
artifact: delta
status: implemented
updated: 2026-11-26
---

# Delta — Staging events topic wiring

## What Is Changing

### Deployment configuration (non-canonical, operational)

- `.github/workflows/ci-cd.yml`, `deploy-staging` → `SAM deploy` step only:
  gains `shell: bash` and a preamble that resolves the CloudFormation export
  `hbk-shared-infrastructure-stg-EventsTopicArn` into
  `shared_events_topic_arn`, validates it against
  `^arn:aws:sns:us-east-2:[0-9]{12}:hbk-domain-events-stg$`, and exits 1 with
  `Missing or invalid staging events topic export` on stderr when the export
  is absent or malformed.
- The resolved ARN is appended as the last `--parameter-overrides` entry:
  `ParameterKey=SharedEventsTopicArn,ParameterValue="$shared_events_topic_arn"`.
- `deploy-production` is untouched; `template.yaml` is untouched (the
  `SharedEventsTopicArn` parameter already exists).

## MODIFIED

None — `canonical_action: not_affected`. No canonical requirement text changes.
This delta only documents the operational wiring of the existing
`SharedEventsTopicArn` parameter that the SNS publisher boundary (REQ-SCC-018,
seasonal-category-competition) already consumes at runtime.

## BDD Scenarios

### SCN-STG-EVENTS-WIRED — Export resolved and passed to SAM

Given the staging export `hbk-shared-infrastructure-stg-EventsTopicArn` exists
with value `arn:aws:sns:us-east-2:123456789012:hbk-domain-events-stg`, when the
staging `SAM deploy` step runs, then `sam deploy` is invoked with
`ParameterKey=SharedEventsTopicArn,ParameterValue=arn:aws:sns:us-east-2:123456789012:hbk-domain-events-stg`.

### SCN-STG-EVENTS-MISSING — Missing export stops before deploy

Given the export lookup returns `None` (absent or invalid), when the staging
`SAM deploy` step runs, then it exits 1 with
`Missing or invalid staging events topic export` and `sam` is never invoked.

## HTTP Boundary Changes

- None.

## Error Vocabulary

- New operational stderr message (workflow only, not an API code):
  `Missing or invalid staging events topic export`.
