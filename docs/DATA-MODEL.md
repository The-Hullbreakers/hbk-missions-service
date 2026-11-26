# Conceptual Data Model

This document defines business identities and queries, not a provider schema.

## Records

- `MissionDefinition`: ID, version, taxonomy, objectives and scoring policy.
- `MissionOffer`: offer ID, definition version, window and repeat policy.
- `ActiveLimitPolicy`: global default and category overrides.
- `MissionContract`: offer snapshot, mode, participant/team and lifecycle.
- `ContractParticipant`: confirmed roster snapshot and participation state.
- `ObjectiveConfirmation`: unique unit, source, evidence and adjudication.
- Evidence media: `MeritCredit` items carry an optional `MediaURI`
  attribute — an opaque reference (`evidence/{contract_id}/{object_id}`)
  to a private S3 object. Objects are written via short-lived presigned
  PUT grants and read via presigned GET grants (reviewers only).
- `Season`: window, categories, tracks, tie-breaks, award tiers, optional `StartsAt` and `SeasonSchedule` (`active_weeks`, `off_weeks`, `series` prefix/iteration).
- `MeritEntry`: immutable credit or adjustment.
- `Standing`: rebuildable category/track projection.
- `Dispute`: challenged confirmation or merit entry and resolution.
- `PublicationIntent`: reliable external event intent.

## Required Queries

- Available eligible offers by filter.
- Available offers excluding active definition IDs for the player.
- Active contract count by player, including team participation.
- Contract history by player/team.
- Repeat history.
- Confirmation by unique business identity.
- Merit entries by season/category/track/recipient.
- Top standings and neighborhood around one recipient.
- Pending validations and disputes before freeze/finalize.
- Unpublished placement awards for retry.

## Integrity

- Every identity addresses a single global scope.
- Contract points to one immutable definition version.
- A player has at most one active contract per stable mission definition.
- The same-definition uniqueness is held by a reservation owned by the creating contract; it is written with the contract in one transaction and released only by its owner when the contract ends or the member departs.
- The definition reservation is distinct from the player's active-slot count, which the same transaction maintains separately.
- Every contract write is fenced on the stored departed-member set matching the candidate's, and the departure write additionally requires the persisted pre-leave state; a stale snapshot can never overwrite a committed departure or mint credits from an outdated roster.
- Participant membership in an active team contract contributes to active count.
- Confirmation uniqueness precedes merit credit.
- Every credit references one confirmation or declared completion bonus.
- Every adjustment references a reason and actor/decision.
- Standing total equals sum of ledger entries for its exact bucket.
- Final standing snapshot is immutable.

## Retention

Offers and evidence may have separate retention policies, but contracts, merit ledger, final standings and placement awards must preserve competitive history. Destructive retention requires a later governed change.
