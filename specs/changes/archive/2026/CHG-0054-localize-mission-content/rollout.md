---
change_id: CHG-0054
artifact: rollout
status: pending
updated: 2026-11-01
---

# Rollout — CHG-0054

## Order

1. Prepare and review the version-bound English catalog drafts,
   verify the actual operator authoring client sends
   `translations.en`, and prove repeated seed preserves existing
   locales. Deployment stops if any prerequisite is absent.
2. Only then release the service with bilingual read and authoring
   support and the already-implemented new-version publish gate.
   Requiring `en` at publish is a behavioral change, not additive:
   the current code does not have a separate toggle for this rule.
   Do not claim that read support can ship first from the current
   binary. CI runs the racing seed after service deployment; the
   seed must already be safe against existing English. Already
   published Spanish-only versions remain valid and serve `es`.
3. Backfill (separately approved execution): apply reviewed
   translations to the current published catalog, then augment
   eligible frozen offer/contract records per definition version so
   historical reads can serve English. Rehearsed on DynamoDB Local
   first (T-0054-009) with a dry-run diff and an unresolved-record
   report. Reviewed input must include its source `es` snapshot;
   same-version frozen records with differing Spanish text are
   unresolved, not translated (amendment approved 2026-11-01).
   Records that stay `es`-only keep serving `es` declared `es`, so
   the backfill may run gradually and never blocks deploys.

### Catalog And Publish Cutover (Gate 1 Amendment 2026-11-01)

The source of truth is the 24 generated racing definitions and 12
explicit PvP definitions under `seeds/`. Version-bound English drafts
are saved separately in Git, paired with exact Spanish source text.
The operator must review the translations and the actual authoring
client must submit `translations.en` before any service release that
enforces English on a new publish. The API's optional draft input
does not prove that the operator's client is ready. PvP remains
local/manual; the deployed seed pipeline still contains racing only
(24 staging, eight production). Draft files and structural checks
are not human translation approval.

The safe seed adapter must ship before any actual historical
translation backfill: today's `seed_missions.py` compares full
published rows and rewrites current definitions on every deployment.
Before shipping a seed that performs exact-offer preflight reads,
attach the separately reviewed `GetItem`-on-`OFFER` policy to the
existing seed deployment role; the current role cannot read offers.
First validate the least-privilege policy contract and a loopback
dry-run/apply against new local tables. A missing role grant is a
deployment blocker, not a reason to skip offer checks or broaden
reads to contracts, `Query`, or `Scan`. Gate 1, Gate 2 IAM RED,
and scoped Gate 3-IAM were approved 2026-11-03; the source
policy's local static test passed 3/3. It has **not been
deployed**; any deployed read or real-table operation still
requires the separate rollout authorization.
First rehearse repeated seed on a new DynamoDB Local table with
Spanish-only rows and with English-enriched current/version/offer
rows; both must remain stable. Reject conflicting Spanish,
economics, version, or reviewed English; never overwrite an
accepted contract. Only a new version may seed both locales
atomically. Then, after operator readiness and L3 gates pass,
deploy the compatible reader/seed and inspect its dry-run.
The existing Spanish-only versions remain Spanish until a separate
approved local-to-staging-to-production backfill; do not smuggle
historical writes into CI's ordinary seed job.

### Offline Seed Evidence (2026-11-03)

The policy source grants `GetItem` for any row with PK=OFFER;
the seed uses computed seed-owned SKs to inspect exact offers
on dry-run and apply; no
OFFER partition query was added to the deployed script. Source-
bound English reaches both CLIs. Python unit 16/16, catalog
regression 8/8, unique-table DynamoDB Local integration 17/17,
and aligned legacy CI seed integration 1/1 passed. Repeated
seeds preserve prior `ContentEn` and Spanish-only history;
previous-version offer drift and concurrent translations block
the transaction. No shared IAM policy was deployed and no real
table was queried or updated. These results do **not** satisfy
authoring-client readiness, full Boundary, or final L3 gates.
The user approved English wording for all 36 records on
2026-11-03 and chose local-only publication; non-local deployment
remains blocked. Do not deploy the changed CI seed
while any prerequisite remains open.

The user approved read-only preflight against local application
`http://localhost:8000` / `hbk-missions-local`. The dry-run found
one active `local-training` season and no catalog conflicts.
Read-only PK queries found 36 current definitions, 36 immutable
version rows, and 36 active-season offers, **all Spanish-only**;
CONTRACT count is zero. The safe seed preserves same-version
Spanish-only rows, so `--apply` would not publish English.
Plan a separately gated, source-checked local backfill for the
108 existing rows, then obtain explicit apply approval and
verify both locales. Do not overwrite Spanish, historical offers,
or accepted contracts to expedite publication.

Gate 1 for a local-only invocation of the already rehearsed
translation backfiller was approved 2026-11-03. T-0054-020
test RED, T-0054-021 invoker GREEN/BLUE, and T-0054-022
specific local-table apply are separate gates. The invoker
must refuse non-loopback endpoints, validate all 36 reviewed
sidecars, preview every candidate before any apply, and use
conditional `ContentEn`-only updates. The adapter is per-record,
not a 108-row atomic transaction: a conflict may leave eligible
earlier rows enriched. Report actual updated/unresolved IDs,
rerun preview, and request a reviewed retry; do not roll back
by overwriting Spanish or deleting data. No staging/production
deploy, CI change, or local-table write is authorized by Gate 1.
Gate 2 test-only RED is reviewed: unit compile fails only for
the absent manifest/loopback API; five isolated integration
cases fail only for the absent CLI. Gate 3 to implement it,
and T-0054-022 permission to apply on the local app table,
remain separate decisions.
Gate 3-LOCAL-BACKFILL was subsequently approved; the
local-only invoker is GREEN/BLUE on fresh private test
tables (unit four methods, CLI integration 5/5, existing
backfill 12 top-level + five drift subtests). No preview
of the app table has been run with this CLI yet. Before
local apply, run its `--dry-run` against the confirmed
loopback port 8000 and `hbk-missions-local`, inspect all
eligible and unresolved IDs, and ask for specific approval
for that table and operation. Never interpret test-table
apply as permission for app-table apply.

Read-only deployment review on 2026-11-04: the CLI preview
found 108 unique eligible Spanish-only records (36 current
definitions, 36 versions, 36 active offers) and zero
unresolved. All 108 still lack `ContentEn`; no contracts
exist. The gateway is up but its missions upstream at
`host.docker.internal:3006` is not listening: mission health,
readiness, and GET routes return 502. The current service
tree fails a compile-only check at the pre-existing
CHG-0051 `NewPlacementFactPublisher` wiring (3 args,
5 now required). Do not claim local publication or HTTP
readiness. Apply English only after exact-table permission;
resolve the separately owned composition blocker and
review the backend startup separately. No restart or write
occurred during this review.

The user subsequently authorized `--apply` exclusively on
`http://localhost:8000` / `hbk-missions-local`. The fresh
source-bound manifest test passed 3/3 and the target preview
returned 108 unique eligible rows with zero unresolved.
Conditional apply added `ContentEn` to all 72 definition
and 36 offer rows; before/after comparisons proved every
other attribute, the season, and zero contracts unchanged.
Each English snapshot matches its reviewed version-bound
manifest. Follow-up preview has zero eligible rows; both
catalogs remain synchronized under the ordinary seed's
read-only dry-run. Local **data** publication is complete.
The local HTTP backend still has no listener on port 3006
and gateway missions health returns 502; resolve the
separately owned CHG-0051 composition blocker and
revalidate `?lang=es|en` only after an approved startup.
No AWS, CI, staging, or production rollout was authorized.

Stop if the operator client is not ready, a full-tree Boundary test
still needs the CHG-0051 overlay, any catalog translation lacks
human review, or a dry-run reports changed Spanish, economics, or
existing English. No AWS account/table query or write is authorized
by this Gate 1 amendment; each non-local rehearsal or apply needs
its own reviewed plan and explicit approval.

### Local Rehearsal Evidence (2026-11-01)

`TEST-INF-LANG-002` passed 12 top-level tests plus five source-text
drift subtests against fresh DynamoDB Local tables. Dry-run reported
eligible matching offers/contracts without changing their raw items;
same-version Spanish drift (including rule order and phased terms)
and wrong-version records were reported unresolved. The apply tests
enriched only matching definition, offer, and contract items with
`ContentEn`; raw item comparisons showed every other attribute
unchanged. Existing authored English was not overwritten, and
incomplete reviewed English was rejected before any update. These
are synthetic reviewed fixtures, not approved translations for the
real catalog. No shared table, live migration, or production CLI was
used. Separate approval and actual reviewed inputs are required
before any real execution.

## Rollback

- Rolling the service back restores the pre-`lang` contract entirely:
  stored `en` attributes are inert to old readers, `es` content is
  never modified, and the language label simply disappears.
- A failed or partial backfill leaves records untouched or enriched;
  the unresolved-record report is the recovery input for the next run.
- If the publish gate causes rejections after cutover, the resolution
  is operator-side (supply `en` for the affected draft) — the data
  model needs no rollback since drafts are untouched.

## Compatibility

- Read paths: additive query parameter and response field; no
  versioned or breaking change. `INVALID_QUERY_PARAMETER` already
  exists on list endpoints — one new `400` reach on `lang` misuse.
- Publish path: deliberately breaking for es-only draft→publish
  clients — mitigated by the staged cutover above, not claimed
  additive.
- `Content-Language` is emitted only on homogeneous non-empty
  payloads; heterogeneous and empty collections omit it, so no client
  can be misled.
- Economics (`merit_summary`, `completion_reward`), identifiers,
  states, ordering, and pagination are language-independent.
- `GET /api/v1/missions/offers/facets` unchanged.
