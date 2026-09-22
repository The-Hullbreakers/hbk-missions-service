---
change_id: CHG-0003
artifact: rollout
status: proposed
updated: 2026-08-29
---

# Rollout Plan

## Deployment Stages

1. **Local**: `docker compose up` raises DynamoDB Local + service.
2. **CI**: GitHub Actions runs unit, integration with DynamoDB Local, E2E, coverage, mutation, security.
3. **Staging**: manual deployment with staging database; smoke tests of `/healthz`, `/readyz`, confirmed.
4. **Production**: blue/green or canary; `readyz` includes DB check.

## Migrations

- DynamoDB is schema-free; no SQL migrations required.
- Table `hbk-missions` with `PK`/`SK` and GSI per credit bucket.
- Entities: `contract`, `season`, `objective`, `merit_credit`.

## Observability

- Logs structured in composition root.
- Metrics: request count/latency, DynamoDB latency/error rate, error rate.
- Health: `/healthz` (live), `/readyz` (DynamoDB available and OIDC issuer accessible).

## Rollback

- Rollback maintains existing credits.
- Acceptance of new confirmations through config is disabled.
- Tables and credits are not deleted.
