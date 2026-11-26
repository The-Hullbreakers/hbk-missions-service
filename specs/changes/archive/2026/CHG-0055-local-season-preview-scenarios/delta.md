# CHG-0055 — Create local season preview scenarios: behavioral delta

## US-0055-001 — Preview seasonal experiences

### REQ-0055-001 — Keep existing local data safe

ADDED: Selecting a preview loads only a dedicated, named local dataset. A
reviewer can return to the prior local dataset without losing its contents.
An unknown destination, non-loopback endpoint, inactive user, conflicting
item, or a repeated load with different contents cannot silently alter an
existing dataset. A repeated load with identical contents is harmless.

#### SCN-0055-001 — Preserve the reviewer's existing work

Given an existing local season and its offers, when a reviewer prepares a
mid-season preview, then the existing season and offers remain unchanged and
the reviewer can return to them.

#### SCN-0055-002 — Repeat without changing the picture

Given a mid-season preview already exists, when it is loaded again with the
same active participants, then its standings and contracts remain unchanged.

#### SCN-0055-003 — Reject unsafe selection

Given a destination outside the approved local previews or conflicting
existing data, when a reviewer attempts to load a preview, then no existing
data is replaced.

### REQ-0055-002 — Preview the phases of a season

ADDED: A reviewer may select off-season, fresh-start, or mid-season independently.
Mid-season uses only presently active local members. A fresh start has
available missions but no progress; mid-season has varied progress, including
contracts in progress and completed and nonempty category standings.

#### SCN-0055-004 — Off-season is visibly quiet

Given no season is active in the selected preview, when a reviewer visits
the mission screens, then they see the off-season experience without
current-season standings or new contracts.

#### SCN-0055-005 — Fresh start has opportunities without standings

Given a newly active preview season, when a reviewer visits missions and
leaderboards, then offers are available and no one has earned merits yet.

#### SCN-0055-006 — Mid-season has believable progress

Given at least three active local members, when a reviewer selects
mid-season, then several members have distinct category standings and
the mission screens show both ongoing and completed contracts.

## Existing requirements affected

- `REQ-SCC-001`, `REQ-SCC-002`, `REQ-SCC-003` and the current-season query:
  scenario merits must retain season/category/track separation.
- `REQ-MB-001`, `REQ-MB-002` and frozen offer terms: previews reuse
  existing publication rules without changing product behavior.
- Existing contract lifecycle/listing requirements: preview contracts
  must be readable as normal contracts, without adding new API behavior.

No product rules, API contracts, real season transitions, identity records,
or official rankings are modified. Preview seasons are unofficial.
