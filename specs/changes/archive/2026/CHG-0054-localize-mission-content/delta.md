---
change_id: CHG-0054
artifact: canonical-delta
status: consolidated
updated: 2026-11-01
---

# Canonical Delta — mission-board

Gate 1 approved 2026-10-27 including the labeled Spanish fallback;
canonical merge deferred to post-verification consolidation.

## ADDED

### REQ-MB-017 — Select the language of mission content

`GET /api/v1/missions/definitions`, `GET /api/v1/missions/offers`, and
`GET /api/v1/missions/offers/{offer_id}` accept `?lang=es|en`. An
omitted `lang` resolves as `es` — the historical default. An
unsupported value, an empty value, or a repeated `lang` whose values
conflict resolves as `400 INVALID_QUERY_PARAMETER`; a repeated `lang`
whose values agree resolves as that value. Language is a URI query
variant: caches key on the URI and `Accept-Language` is never
negotiated.

Every mission representation declares the language actually served:
single-resource reads carry the label at top level and every item in a
collection carries its own. The label key follows each response
family's existing casing — `ContentLanguage` on the PascalCase
board/catalog projections and contract-list items,
`content_language` on the snake_case contract-detail and review-queue
DTOs. A `Content-Language` response header is emitted when — and only
when — every representation in a non-empty payload resolves the same
language; heterogeneous and empty collections omit it rather than
falsely describing the payload.

A record that stores only Spanish serves its Spanish content declared
as Spanish under any `lang`, including `lang=en` — never `406` and
never English-labeled Spanish. Offer search `q` matches each offer's
effective-language title and description — including the Spanish
fallback — before ordering and pagination, so `total` reflects
language-dependent matches while ordering and paging stay
deterministic. Identifiers, amounts (`merit_summary`,
`completion_reward`, merit values), and states are
language-independent.

`GET /api/v1/missions/offers/facets` does not accept `lang` and is
unchanged.

#### SCN-MB-LANG-EN — Member reads a bilingual mission in English

Given a published offer whose definition version has complete Spanish
and English content, when a member reads the board or the offer in
English, then the title, Markdown description, ordered rules, and
objective names and descriptions render in English — declared as
English — while the offer id, objective ids, and merits stay stable.

#### SCN-MB-LANG-DEFAULT — Omitted language serves Spanish

Given a bilingual offer, when a member reads the board without
choosing a language, then the content renders in Spanish and is
declared as Spanish.

#### SCN-MB-LANG-LEGACY — Spanish-only record falls back declared

Given an offer or definition stored before English existed, when a
member asks for English, then the Spanish content is served declared
as Spanish — not an error and not declared as English.

#### SCN-MB-LANG-INVALID — Invalid language choice rejected

Given a member asks for an unsupported, empty, or self-contradicting
language selection, when the request is evaluated, then it is rejected
as an invalid request and no mission content is served.

#### SCN-MB-LANG-MIXED — Mixed collection declares per item

Given a board page where some offers serve English and legacy offers
serve Spanish, when a member reads the board in English, then each
item declares the language it actually serves and the response claims
no single language for the whole page.

#### SCN-MB-LANG-HEADER — Homogeneous page declares one language

Given a single-offer read or a page where every item serves English,
when the member asks for English, then the response declares English
as its one language; an empty page declares no language at all.

#### SCN-MB-LANG-SEARCH — Search matches the effective language

Given an offer whose Spanish and English titles differ, when a member
searches the board in English, then matching runs against the English
text — and against the Spanish fallback for legacy records — before
results are ordered and paged.

## MODIFIED

### REQ-MB-013 — Publish definitions with mission content and reward rules

Extends: definition content becomes locale-aware. `title`,
`description` (Markdown, verbatim), ordered `rules`, and each
objective's `name`/`description` exist per locale; `es` remains
required. `POST /api/v1/missions/definitions` keeps its existing
snake_case body as the Spanish content and accepts an optional
`translations` block — `translations.en` carries `{title, description,
rules, objectives: [{objective_id, name, description}]}` stored
atomically with the Spanish content; `rules` may be empty (zero rules
stays valid), objective `description` is optional exactly as `es`
allows, `name` is required per entry. Draft creation stays backward
compatible — the `es`-only body still succeeds.

Publishing a version requires both locales structurally complete:
every `es` `objective_id` appears in `en` with a `name` (and a
`description` wherever `es` carries one), and the `en` rule count
equals the `es` rule count in the same order — completeness is
structural, never a guess at translation equality. An incomplete `en`
resolves as `422 INVALID_MISSION_DEFINITION`. There is no automatic
machine translation: every locale is operator-supplied reviewed text.

This completeness rule is a deliberate behavioral change to the
publish flow, not additive: previously published Spanish-only versions
stay valid and keep serving `es`, but new-version publishes require
English. Authoring clients must provide `en` before enforcement takes
effect (see `rollout.md`).

Catalog amendment (Gate 1 approved 2026-11-01): the 24 racing and 12
PvP definitions in the repository gain version-bound English drafts
for their Spanish title, description, ordered rules, and objective
names/descriptions. The exact Spanish source accompanies every draft
so a changed definition cannot silently inherit the wrong English.
IDs, category, mode, version, reward, and objective economics are not
translated. A human reviews text before it becomes an approved
translation. The racing seed is already run on deploy; repeated
synchronization must preserve an existing same-version English
snapshot without replacing Spanish or approved English. Missing
English on a historical version remains a Spanish fallback until a
separately authorized backfill. Newly created published versions
must include structurally complete English; they cannot use the
legacy exception. PvP is not added to the deployed racing catalog.

#### SCN-MB-CONTENT-DRAFT-EN — Draft may carry an English translation

Given an operator authors a definition with Spanish content plus a
complete English translation, when the draft is saved, then both
languages are stored together on the same version.

#### SCN-MB-CONTENT-INCOMPLETE — Publish requires both locales complete

Given a definition draft whose English translation is missing or
structurally incomplete, when the operator publishes a new version,
then the publish is rejected as an invalid definition and no version
is stored.

#### SCN-MB-CATALOG-TRANSLATED — Review complete bilingual missions

Given the Spanish racing and PvP missions are versioned, when an
operator reviews their English text, then each mission has an English
title, description, ordered rules, and objective text paired with
the exact Spanish version, without changing its identity or rewards.

#### SCN-MB-CATALOG-REDEPLOY — Preserve previously approved text

Given a published Spanish version has gained reviewed English, when
the same catalog is deployed again, then both languages and their
promised rewards remain unchanged, including in offers from that
version. A conflicting Spanish or English snapshot is rejected,
not silently replaced.

Reviewing the next deployment against a conflicting frozen offer
must stop it without rewriting its Spanish, English, or economics.
The automated deployment must not inspect unrelated offers or
contracts to achieve this outcome. Gate 1 amendment approved
2026-11-03; no real-table read is approved by this paragraph.

#### SCN-MB-CATALOG-NEW-VERSION — Require English on first publication

Given an operator adds a new catalog version, when it is published,
then both complete languages are stored together; if the English
text is missing or incomplete, no new version or offer is published.

#### SCN-MB-CATALOG-LOCAL-PUBLISH — Add approved English to local history

Given the operator has approved the English for a published mission
and its current definition, version, and active offer still contain
the exact reviewed Spanish, when the operator publishes that English
to the local catalog, then those records serve both languages
without changing their Spanish, identity, rewards, or version.
Already accepted contract terms remain unchanged.

#### SCN-MB-CATALOG-LOCAL-CONFLICT — Report divergent local history

Given a published mission's local Spanish history differs from the
reviewed source or its English is already different, when the operator
previews local publication, then the divergent record is reported,
left unchanged, and publication does not begin.

### REQ-MB-005 — Freeze promised rules

Extends: the frozen snapshot now carries both locales. An offer
created from a published version freezes the `es` and `en` title,
description, rules, and objective names/descriptions of that version;
acceptance embeds the same bilingual `ContractTerms` snapshot. Later
version edits — including translation-only edits — never rewrite
frozen offers or contracts. A frozen record created before English
existed keeps its Spanish-only snapshot and reads as `es`; the
reviewed-translation backfill may add `en` to an eligible frozen
record but never alters its Spanish content, economics, ids, or
version identity. The reviewed translation includes the source
Spanish snapshot; a frozen record whose Spanish title, description,
ordered rules, or objective ids/text differ is left untouched and
reported unresolved even if its definition code and version match
(amendment approved 2026-11-01).

#### SCN-MB-SNAPSHOT-BACKFILL-DRIFT — Reject a divergent frozen snapshot

Given a reviewed translation for a definition version and an older
offer or contract whose frozen Spanish text differs despite the same
version and objective ids, when the operator rehearses its backfill,
then no English is added to that record and the report identifies it
as unresolved; a matching record is eligible for English-only
augmentation.

#### SCN-MB-SNAPSHOT-BILINGUAL — Offer freezes both locales

Given a definition version published with Spanish and English content,
when an offer is created and a member accepts it, then the offer and
the contract terms keep both language snapshots of that version.

#### SCN-MB-VERSION-EDIT — A translation edit never rewrites history

Given a member holds a contract accepted at version 1 and the operator
publishes version 2 with edited translations, when the offer and
contract are read in either language, then they still show the
version-1 text in both languages.

### REQ-MB-015 — Expose mission content on the board and catalog

Extends: `Title`, `Description`, `Rules`, and each objective's `name`
and `description` resolve to the request's effective language per
`REQ-MB-017`, and every item declares its served language with the
family's label key. Field names and shapes are unchanged;
`objective_id`, `merit_summary`, rewards, and availability are
language-independent. Stored items predating English resolve `es`
content declared `es`.

### REQ-MB-016 — Read a single offer

Extends: `GET /api/v1/missions/offers/{offer_id}` accepts `lang` per
`REQ-MB-017` and returns the same localized projection as the list
items, with the language label at top level.

# Canonical Delta — mission-contract-fulfillment

## ADDED

### REQ-MCF-026 — Select the language of contract content

`GET /api/v1/missions/contracts`,
`GET /api/v1/missions/contracts/{contract_id}`, and
`GET /api/v1/missions/review-queue` accept `?lang=es|en` under the same
rules as `REQ-MB-017`: omitted resolves `es`; invalid, empty, or
conflicting repeated values resolve `400 INVALID_QUERY_PARAMETER`; URI
variant, no `Accept-Language` negotiation.

The localized fields are the frozen content the contract captured at
acceptance: the offer `title`/`description`/`rules` and the frozen
objective `name`/`description` (contract detail keeps its existing
lowercase field names). The contract detail carries the label at top
level; each contract-list item carries one `ContentLanguage` label at
the contract level that covers its embedded `Offer` — both render the
same frozen snapshot in the same language; each review-queue entry
carries `content_language`. `Content-Language` header and homogeneous
rules follow `REQ-MB-017`. A contract whose frozen snapshot has no
English serves `es` declared `es`.

Authorization is unchanged — roster membership, adjudicator role, and
member-vs-reviewer projections resolve exactly as before; `lang`
changes only which language renders.

#### SCN-MCF-LANG-EN — Member reads contracts in English

Given a contract accepted from a bilingual offer version, when the
member reviews their contracts in English, then the embedded offer and
frozen objectives render English declared as English with identical
ids and merit figures.

#### SCN-MCF-LANG-DETAIL — Contract detail renders the selected language

Given a bilingual contract, when an active member or an adjudicator
opens it in English, then the title and each frozen objective's name
and description render English — while a member still sees only their
own staged credits without media references and an adjudicator sees
the full view.

#### SCN-MCF-LANG-QUEUE — Reviewer queue titles localize without loosening access

Given a bilingual contract awaits adjudication, when a moderator reads
the review queue in English, then its title renders English — while a
caller without the adjudicator role is still refused regardless of
language.

#### SCN-MCF-LANG-LEGACY — Spanish-only snapshot falls back declared

Given a contract frozen before English existed, when any authorized
reader asks for English, then the frozen Spanish content is served
declared as Spanish.

## MODIFIED

### REQ-MCF-006 — List player contracts

Extends: the list accepts `lang`; each item's embedded `Offer` and its
frozen `objectives` resolve the effective language from the contract's
snapshot, and the item declares the served language at the contract
level (covering the embedded offer). `merit_summary` values, filters,
ordering, and totals are language-independent.

### REQ-MCF-021 — List the pending review queue

Extends: the queue accepts `lang` and each entry's `title` resolves
the effective language from the contract snapshot, with a
`content_language` label per entry; role enforcement, fields, and
counts are unchanged.

### REQ-MCF-022 — Read own contract detail

Extends: the detail accepts `lang`; `title` and frozen objective
`name`/`description` resolve the effective language with a top-level
`content_language`. The member-vs-reviewer projections,
`pending_credits`/`rejected_units` filtering, and access rules are
unchanged.
