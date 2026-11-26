---
change_id: CHG-0034
artifact: security-review
updated: 2026-10-15
---

# Security Review — Party Manager Team Membership

## Role

Reviewer: AI implementation agent (self-review); human sign-off pending at verify gate.

## Trust Boundary

New synchronous HTTP dependency: `missions-service` → Party Manager. The
membership answer authorizes whether a declared teammate consumes an
active-contract slot. A wrong "yes" lets a caller draft a stranger's slot; a
wrong fail-open on error does the same silently.

## Threats And Controls

| Threat | Control | Evidence |
| --- | --- | --- |
| Fail-open on transport error grants phantom membership | Adapter returns `ErrMembershipUnavailable`; handler aborts acceptance; no contract written | T-034-001 contract test + T-034-003 integration |
| 404 (definitive non-member) misclassified as member | Adapter maps only `200` to `true`; `404` → `false, nil` | T-034-001 contract test |
| Indefinite hang blocks accept path | 2s client timeout, no retries | T-034-001 contract test (timeout case) |
| URL injection via `team_id`/`player_id` | Path segments are URL-escaped by the adapter | T-034-001 contract test (escaping case) |
| Missing `PARTY_MANAGER_URL` silently allows all | Composition falls back to `DenyAllReader`, never allow-all | T-034-003 composition test |
| Secrets exposure | No credentials required for the internal endpoint; if auth is later added it must come from env, never hardcoded | N/A |

## Findings

- No Critical/High findings. Residual risk: availability dependency on Party
  Manager at accept time — accepted, documented in `rollout.md`.
