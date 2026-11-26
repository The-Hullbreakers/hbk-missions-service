---
change_id: CHG-0043
title: Synchronize the racing mission catalog after deployment
status: locally_verified
risk: L2
canonical_action: modify
canonical_paths:
  - specs/capabilities/mission-board/spec.md
execution_mode: sequential
---

# Racing catalog deployment

User request authorizes adding the racing catalog, insertion scripts and deployment
integration on staging/main. Domain/application behavior is not changed.

## Acceptance scenarios

- Fresh active season: publish each enabled definition (current and immutable
  version) and one SOLO offer, with the five-consecutive-lap evidence requirement.
- Redeploy: preserve accepted contracts and historical seasons.
- Change content: require a higher integer definition version; refuse modification
  of an immutable version and refuse a downgrade of current.
- Concurrent acceptance: offer content updates leave in-flight acceptances unaffected.
- Concurrent season transition or catalog write: cancel the whole catalog transaction.
- No active season: synchronize definitions, report pending offers; never create a season.
- More than one active season: fail before writing.
- Production: omit missions marked staging-only pending timing calibration.
- CI: validate catalog and run unit and DynamoDB integration checks before deployment;
  synchronize only after successful SAM deploy, including empty changesets.

## Layer gates / verification

Domain and application: not affected. Boundary RED: seed tests precede the script.
Boundary GREEN: Python unit tests, DynamoDB Local integration and workflow checks.
IAM: deployment role receives only catalog writes and season/catalog reads/checks
on this stack's table. No contract/ledger writes or deletes.
Rollback: keep immutable history; publish corrected content with a new version.
No automatic deletion or rollback of live data.

## Evidence (2026-10-18)

- RED: six seed boundary tests failed to import the not-yet-created script.
- GREEN: six unit tests and the workflow contract passed.
- DynamoDB Local: isolated table lifecycle test passed, covering 24 definitions,
  immutable versions, 24 offers per season, dry run, corrected
  versions, contract preservation, downgrade rejection and atomic cancellation
  after a concurrent season transition. The temporary table was deleted by teardown.
- `sam validate --lint --template-file template.yaml --region us-east-2` passed.
- Workflow YAML parsed; both environment/table mappings and post-SAM ordering checked.
- No staging or production writes, commit, push or deployment performed here.
