# Product Specification

## Problem

HBK members coordinate activities, but do not have a structured board for discovering contracts, accepting objectives, demonstrating results, and competing for discipline over a season.
## Vision

Offer an experience similar to the Star Citizen contract board: available missions, active contracts and history; individual or team execution; verifiable objectives; and competition by category through Season Merits.
## Actors

- **Member**: Discover and accept individual missions.
- **Team leader**: share or start a team contract.
- **Team participant**: confirms your participation and contributes to the contract.
- **Mission operator**: design, version, publish and withdraw offers.
- **Reviewer**: validates evidence and resolves deliveries or disputes.
- **Competition manager**: configure season, categories, tiebreakers and prizes.
- **System**: expires contracts, credits merits and ends standings.
## Core Journey

1. An operator publishes an offer from a definition.
2. The member checks the board and their eligibility.
3. Missions validates window, quota and limit of active contracts.
4. The player accepts alone or confirms a team roster.
5. The contract progresses by phases and objectives.
6. Each valid unit is confirmed once.
7. Merits are credited to the correct category, season and track.
8. The user consults contract, history and leaderboard.
9. The season freezes, resolves pending issues and ends.
10. Missions communicates the awarded positions to Achievements.
## Mission Offer

An offer publishes issuer, narrative, category, type, subtype, difficulty, mode, team participation, location, availability, duration, objectives, merits, repeatability, cooldown and capacity.
It can be:

- `UNLIMITED`: multiple contracts while still available;
- `LIMITED`: a maximum of acceptances;
- `SINGLE_CLAIM`: special case with capacity one.
## Mission Definitions And Offers

- `MissionDefinition` is the reusable template that has basic objectives and merits.
- `MissionOffer` is the specific post generated from a published definition.
- Accepting an offer creates a `MissionContract` with frozen terms.
Publication mode:
- `MANUAL`: the operator publishes a special event.
## Active Template Visibility

The board is customized per player. If a player already has an `ACCEPTED`, `IN_PROGRESS`, or `PENDING_VALIDATION` contract for template X, they do not see any other offer for X and cannot accept one directly by ID.
The rule uses the template's stable identity, not its version. At the end of the contract, the template can be displayed again if cooldown and repeatability allow it. For teams, it applies to each confirmed participant.
## Active Contract Limit

- There is a default limit per tenant.
- There may be an override per category.
- The initial recommended value is three.
- Any contract accepted, in progress or pending validation counts.
- One team contract counts for each confirmed participant.
- An unconfirmed invitation does not consume a place.
- Reducing the limit does not cancel existing contracts.
- The next acceptance is blocked until it falls below the new limit.
## Competitive Vocabulary

- **Mission Category**: discipline that owns a classification.
- **Mission Type/Subtype**: specific classification of the contract.
- **Season Merit**: temporary competitive point credited by an action.
- **Track**: `INDIVIDUAL` or `TEAM` competition.
- **Standing**: position derived from merits.
- **Award Tier**: awardable result such as champion, top 3 or top 10.
## Initial Categories

`BOUNTY_HUNTING`, `MINING`, `SALVAGE`, `HAULING`, `RACING`, `MEDICAL_RESCUE`, `EXPLORATION`, and `SECURITY` are initial candidates. The final list is a manageable business catalog, not an irreversible enum.
## Success Criteria

- No player exceeds their limit due to race conditions.
- No player sees or resumes a template while it is active.
- No duplicate confirmation increases merits.
- Standings reconcile exactly with credits and adjustments.
- Team points are not multiplied between members.
- A new season starts from scratch without losing history.
- Achievements failure does not reverse the final result.
## Non-Goals V1

XP, levels, global ranking between categories, Inventory rewards, own multimedia files, procedural generation, direct synchronization with Star Citizen and advanced anti-fraud algorithms.
