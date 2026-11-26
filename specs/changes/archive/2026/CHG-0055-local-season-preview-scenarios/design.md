# CHG-0055 — Create local season preview scenarios: design

## Boundary and dataset

Use Python 3.9+ and the existing AWS CLI v2 transport and catalog helpers.
The preview script validates its endpoint using parsed scheme, hostname,
port, and absence of userinfo/path/query/fragment; permit HTTP loopback
only. Force dummy local credentials and disable metadata. Permit only
the three fixed `hbk-missions-preview-*` table names; no arbitrary table
or production/staging target. `hbk-users` is read-only and only its
`USER#<id>` / `METADATA#<id>` rows with `Status=active` are eligible.
Collect all pages with consistent reads, sort IDs, and use a deterministic
bounded selection. Never emit names, emails, roles, or entire user rows.
Only mid-season requires at least three active members; off-season and
fresh-start do not need users.

Build desired rows in memory before any writes, validate uniqueness of
PK/SK, maximum transaction sizes and expected shapes. Create a missing
table with PK/SK plus GSI1 and GSI2 indexes matching existing missions
table definitions; wait until ACTIVE. A present table must match the
schema. Refuse any item not in the exact expected scenario dataset.
Compare existing rows to desired rows before writing: exact matches are
no-ops, any differing row is an error. Apply missing rows with conditional
transactional puts, grouped in <=100 operations. Preflight conflicts
prevent partial writes from known drift; interruption may leave a partial
preview and retry safely fills only missing identical rows. Do not delete
or overwrite. The preview output is counts and target name only.

Off-season has no active season, offers, contracts, or credits (catalog
definitions may exist). Fresh-start has one ACTIVE unofficial season and
offers from the existing 36-mission staging catalog but no contract/credit.
Mid-season has one ACTIVE unofficial season, those offers, deterministic
individual solo contracts tied to selected active members, and credits
using the service's `MERIT_CREDIT` / GSI1 / GSI2 shape. Credits reference
real seeded contracts and objective IDs, match the contract's category
and season, and award plausible catalog-defined merit values; completed
contracts account for the existing leaderboard tie-break. No fake media,
team membership, identity profile, confirmation events, or official award.
Reuse existing published offer and definition content, not a second catalog.

The local service reads `DYNAMODB_TABLE_NAME` (Lambda) or `DYNAMODB_TABLE`
(local executable); documentation shows how to select a preview and restore
the original. Python local-resource tests verify persisted fixture
shapes; a focused Go infrastructure integration test uses the existing
real standings and contract readers against a dedicated persistent preview table.
No changes to Domain, Application, HTTP, web app, or production IaC.

## Layer and test strategy

Application/Domain product logic: not affected. First RED is at the
script's pure scenario-builder/guard boundary; no fictitious Go
Application test. Next RED uses Python local-resource persistence
checks and a Go infrastructure test against existing adapter queries.
Each suite runs independently: Python applies fresh-start and Go
applies mid-season to their own fixed, persistent local preview
tables. These tables are the requested review outcomes, not
disposable test fixtures; tests never delete them. Each applies its
own dataset and does not require another test layer to run first.
Evidence review precedes the
corresponding script change. Existing HTTP tests remain regression checks.
L3 E2E uses a real local service plus DynamoDB Local and checks
current season, offers, contracts and leaderboard without claiming a
browser screenshot if none was actually captured.

Safety: never take credentials from cloud profiles, never print user data,
never treat a GSI count as a source of truth, and never reset a table.
Security role `none`; trust boundary is local CLI -> local identity read
and dedicated local missions table write.
