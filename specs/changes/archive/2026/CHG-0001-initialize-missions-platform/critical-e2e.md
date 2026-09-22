---
change_id: CHG-0001
artifact: critical-e2e
status: planned
updated: 2026-08-26
workflow: WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
---

# Critical E2E Plan

There is no executed evidence: Gate 2 is pending.

## Journey E2E-001 — From Board To Seasonal Award

1. A competition manager sets active limit of two and creates a season.
2. Publish a PvP Bounty offer with 100 merits per unique target.
3. A player accepts two contracts.
4. The accepted template disappears from your board and a direct acceptance of another offer of the same template fails.
5. A third template is rejected by the limit without consuming capacity.
6. A valid target is confirmed and the same confirmation is resend.
7. The ledger contains a single credit and the leaderboard shows 100.
8. The player ends a contract and is eligible for his roster again according to cooldown.
9. Accept the third offer successfully.
10. Standings freeze and end.
11. The placement award is published and its retry does not duplicate the achievement.
12. The next season shows zero and the previous one keeps the result.

## Required Variants

- Another tenant does not observe or alter offers, contracts or standings.
- A team contract counts against the limit of each confirmed member.
- Team merits appear only in the team track.
- Achievements not available leaves the award pending without reopening the season.
- An eligible dispute prevents completion until resolved.

## Planned Evidence

- Execute using the public boundary and real composition.
- Persist command, timestamp, revision and result.
- Do not substitute direct calls to use cases.
- Keep the business scenario separate from the technical setup.
