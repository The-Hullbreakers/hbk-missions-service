# English catalog translations

Each JSON file contains `translations`: version-bound records with
`DefinitionCode`, `DefinitionVersion`, an exact Spanish `SourceSpanish`
snapshot, and an `English` snapshot. Each snapshot contains
`Title`, `Description`, ordered `Rules`, and `Objectives` with
`ObjectiveID`, `Name`, and `Description`. IDs and reward values never
translate.

The user approved the English wording of all 36 catalog missions on
2026-11-03 for local publication. This is content approval, not
authorization to deploy to staging/production or to backfill a
shared table. The
seed translator must verify every Spanish field against the source
catalog and reject a missing, duplicate, stale, or incomplete entry.
The seed now reads these files. Existing same-version Spanish-only
rows remain unchanged until a separately planned backfill; no
translation should be applied to a shared table without separate
execution approval.
