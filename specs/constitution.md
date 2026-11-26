---
artifact_id: ART-CONSTITUTION-SDD
artifact_title: HBK Missions SDD constitution
status: active
updated: 2026-08-26
---

# HBK Missions SDD Constitution

This constitution complements `agent-rules` and does not weaken its gates.

## Domain-First Principle

Work begins with the business language: offer, contract, confirmed objective, merit, category, season, and standing. Tables, handlers, SDKs, and delivery mechanisms are derived afterward.

## Mandatory Sequence

```text
Business BDD
  → Domain RED / Gate 3-DOMAIN / GREEN / BLUE
  → Application RED / Gate 3-APPLICATION / GREEN / BLUE
  → Boundary RED / Gate 3-BOUNDARY
  → Infrastructure / Interface / Composition
  → L3 verification / consolidation / archive
```

Gate 1 authorizes the change artifacts. Gate 2 authorizes starting RED. Each Gate 3 opens only its corresponding layer. No documentation approval authorizes premature production implementation.

## Service Invariants

- All state and standings live in a single global scope.
- A confirmed objective produces at most one merit credit per recipient and policy.
- Merits belong to a season, category, and competitive track.
- A finalized season is never reopened or silently rewritten.
- Starting another season creates a new score; it does not erase the previous one.
- An accepted contract retains the published version of its rules and rewards.
- The active-contract limit also applies to confirmed team members.
- Missions determines placements; Achievements owns permanent achievements.

## Architecture

- Domain and Application use provider-neutral terms.
- Ports are small, consumer-owned interfaces with a single command or query behavior.
- DI belongs to the business module; the executable only assembles the module.
- Party Manager, Identity, and Achievements remain outside the Missions domain.
- Leaderboards are views derived from the merit ledger, not a second source of truth.

## Minimum Evidence

This system is classified as L3 because of isolation, authorization, competitive adjudication, concurrency, and integrations. It requires BDD, RED/GREEN/BLUE per layer, security validation, coverage, mutation testing, critical E2E testing, rollout, rollback, and independent review before consolidation.
