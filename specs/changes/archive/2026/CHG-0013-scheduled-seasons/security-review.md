---
change_id: CHG-0013
artifact: security-review
status: proposed
updated: 2026-09-02
---

# Security Review — Schedule Seasons With Automatic Transitions

## Trust Boundaries

- **Public boundary**: `POST /api/v1/seasons` still goes through `AuthMiddleware` and only the operator role can create a season.
- **Internal boundary**: `SeasonScheduler` is an internal background goroutine. It does not expose a public endpoint and does not accept external input.
- **Data boundary**: `SeasonStore` continues to use tenant-scoped access through the caller. The scheduler uses the same store and does not bypass tenant checks.

## Threats and Mitigations

| Threat | Risk | Mitigation |
| --- | --- | --- |
| Unauthorized season creation | High | Reuse existing `AuthMiddleware` and admin role checks in `CreateSeasonHandler`. |
| Scheduler race creates two active seasons | Medium | `CreateScheduledSeasonHandler` checks `ListActive`; `PutItem` is idempotent for the same `SeasonID`. |
| Scheduler finalizes season too early | Medium | `ShouldFreeze` and `ShouldFinalize` use a configurable `Clock` and are unit-tested against arbitrary `now`. |
| Tenant isolation break | Low | `Season` remains stored per `PK` and the scheduler does not expose cross-tenant queries. |
| Denial of service via frequent ticks | Low | `scheduler_tick_minutes` is bounded (default 1, min 1). |

## Findings

- No new secrets or credentials.
- No new external APIs.
- No changes to OIDC or session handling.
- Background scheduler has no public surface and cannot be triggered by unauthenticated callers.

## Residual Risk

None identified for v1. Multi-instance scheduler race is acceptable for an infrequent daily transition; if the service scales horizontally, a distributed lock should be added in a later change.
