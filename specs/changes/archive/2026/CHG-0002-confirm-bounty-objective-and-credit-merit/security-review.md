---
change_id: CHG-0002
artifact: security-review
status: planned
updated: 2026-08-27
security_role: resource-server
---

# Security Review

## Trust Boundaries

- Authenticated user → sending evidence and consulting standings.
- Authorized operator → publication of objectives, seasons and corrections.
- Reliable sources → goal confirmations.
- Missions → merit credits and projection of standings.

## Mandatory Policies

- Tenant is derived exclusively from the validated token.
- The client does not select the recipient, credit category or amount of merits.
- Only contract participants can receive attribution.
- External confirmations validate source, schema, tenant, contract, objective and uniqueness.
- Public tenant leaderboards do not expose evidence or sensitive identifiers.

## Abuse Cases

- Reuse the same target in several credits.
- Resend a confirmation to double merits.
- Change category or points from the client.
- Confirm in a contract from another tenant.
- Send evidence when the season is not `ACTIVE`.

## Pending Evidence

Executable threat model, capabilities matrix, cross-tenant tests, replay tests and rate limits are pending until Gate 2.
