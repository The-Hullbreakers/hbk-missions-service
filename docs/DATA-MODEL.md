# Conceptual Data Model

This document defines business identities and queries, not a provider schema.

## Records

- `MissionDefinition`: ID, version, taxonomy, objectives and scoring policy.
- `MissionOffer`: offer ID, definition version, window, capacity and repeat policy.
- `ActiveLimitPolicy`: tenant default and category overrides.
- `MissionContract`: offer snapshot, mode, participant/team and lifecycle.
- `ContractParticipant`: confirmed roster snapshot and participation state.
- `ObjectiveConfirmation`: unique unit, source, evidence and adjudication.
- `Season`: window, categories, tracks, tie-breaks, award tiers, optional `StartsAt` and `SeasonSchedule` (`active_weeks`, `off_weeks`, `series` prefix/iteration).
- `MeritEntry`: immutable credit or adjustment.
- `Standing`: rebuildable category/track projection.
- `Dispute`: challenged confirmation or merit entry and resolution.
- `PublicationIntent`: reliable external event intent.

## Required Queries

- Available eligible offers by tenant and filters.
- Available offers excluding active definition IDs for the player.
- Active contract count by tenant and player, including team participation.
- Contract history by player/team.
- Offer capacity and repeat history.
- Confirmation by unique business identity.
- Merit entries by season/category/track/recipient.
- Top standings and neighborhood around one recipient.
- Pending validations and disputes before freeze/finalize.
- Unpublished placement awards for retry.

## Integrity

- Every identity is tenant-scoped.
- Contract points to one immutable definition version.
- A player has at most one active contract per stable mission definition.
- Participant membership in an active team contract contributes to active count.
- Confirmation uniqueness precedes merit credit.
- Every credit references one confirmation or declared completion bonus.
- Every adjustment references a reason and actor/decision.
- Standing total equals sum of ledger entries for its exact bucket.
- Final standing snapshot is immutable.

## Retention

Offers and evidence may have separate retention policies, but contracts, merit ledger, final standings and placement awards must preserve competitive history. Destructive retention requires a later governed change.
