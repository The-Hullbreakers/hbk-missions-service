---
change_id: CHG-0001
artifact: security-review
status: planned
updated: 2026-08-26
security_role: resource-server
---

# Security Review

## Trust Boundaries

- Authenticated user → board, contracts and competitive profile.
- Authorized operator → publication, policies, allocation and season.
- Party Manager → roster snapshot requested by Missions.
- Reliable sources → goal confirmations.
- Missions → awarded positions sent to Achievements.

## Mandatory Policies

- Tenant is derived exclusively from the validated token.
- An ID from another tenant does not allow reading, accepting, confirming or scoring.
- The client does not select the recipient, credit category or amount of merits.
- Publishing or changing policies requires administrative capacity.
- Only contract participants can receive attribution.
- External confirmations validate source, schema, tenant, contract, objective and uniqueness.
- The tenant's public leaderboards do not expose evidence or sensitive identifiers.
- Limit reductions never silently cancel contracts.

## Abuse Cases

- Concurrent acceptances to exceed the limit.
- Consult by ID or directly accept an offer from an already active and hidden template.
- Reuse the same PvP target in several credits.
- Add members to the team after completing.
- Resend a confirmation or season award.
- Change category or points from the client.
- Adjust merits without auditable reason.
- End while eligible disputes exist.

## Pending Evidence

Executable threat model, capabilities matrix, cross-tenant tests, replay tests, rate limits and privacy review are pending until Gate 2.
