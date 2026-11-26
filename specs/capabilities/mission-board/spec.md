---
capability_id: CAP-MISSION-BOARD
title: Mission Board
status: current
last_consolidated_change: CHG-0055-local-season-preview-scenarios
updated: 2026-11-23
---

# Mission Board

## Intent

Publish, discover, and accept eligible mission offers so members can take solo or team contracts while respecting active-mission limits and template-exclusion rules.

## Scope

- Publish versioned offers with issuer, category, type, subtype, mode, difficulty, window, objectives, merits, and repeatability.
- Discover available offers filtered by category, type, subtype, mode, difficulty, and availability.
- Enforce a per-player active-mission limit with optional category overrides.
- Hide any offer whose stable definition the player already has active.
- Preserve accepted contract terms even when the offer or definition is updated later.

## Non-Goals

- Generic rule engine for objectives.
- Procedural contract generation.
- Cross-season or global ranking.
- Manual reviews or disputes (see `mission-contract-fulfillment`).

## Key Concepts

- `MissionDefinition`: stable, versioned contract template.
- `MissionOffer`: published availability and repeatability.
- `MissionMode`: `SOLO`, `TEAM`, `FLEX`.
- `ActiveLimitPolicy`: platform default and optional category overrides.
- `DecideAcceptance`: single decision that validates eligibility, window, repeatability, and active limit.
- `ActiveContractSummary`: read-only value object consumed by `DecideAcceptance`.

## Requirements

### REQ-MB-001 — Publish a mission offer

An authorized operator publishes versioned offers with issuer, category, type, subtype, mode, difficulty, window, objectives, merits, and repeatability. Offers carry no acceptance cap — every eligible member may accept while the offer is published.

#### SCN-MB-POOL — Show a published offer

Given a published offer is inside its window, when an eligible member opens the board, then the offer appears with its current rules and availability.

### REQ-MB-002 — Discover eligible missions

A member queries `AVAILABLE` offers and filters by category, type, subtype, mode, difficulty, and availability. The result explains why an offer is not eligible. The board operates in a single scope.

#### SCN-MB-DISCOVERY — Isolate and filter the pool

Given offers exist across categories, when the member filters their board, then they only receive offers matching the filters.

### REQ-MB-003 — Limit active contracts

The active-contract limit is a single platform default per player with optional overrides per category. `ACCEPTED` and `IN_PROGRESS` count; `PENDING_VALIDATION` (slot released at `submit`), unconfirmed invitations, and terminal states do not. The effective limit (default or the offer-category override) applies to the player's **total** active set across all categories — category overrides raise or lower the threshold but the count is never sliced by category.

#### SCN-MB-LIMIT-TOTAL — Count the total across categories

Given a default limit of 3 and a player holding one active contract in each of three categories, when they accept a fourth offer in any category, then the acceptance is rejected with `ACTIVE_MISSION_LIMIT_REACHED`.

#### SCN-MB-001 — Reject acceptance over the limit

Given a player already reached their effective limit, when they attempt to accept another eligible offer, then the acceptance is rejected without creating a contract.

#### SCN-MB-002 — Preserve contracts when lowering the limit

Given an operator lowers the limit below a player's current contracts, when the policy becomes effective, then their contracts are not cancelled and they cannot accept another until they are below the limit.

### REQ-MB-004 — Consistent acceptance

Acceptance validates eligibility, window, repeatability, and limit in a single decision, then commits the contract, each roster member's active-set entry, and a per-player+definition reservation for every roster member in **one conditional transaction**. Two concurrent acceptances serialize: the first commits and the rest receive the explainable rejection (`ACTIVE_MISSION_LIMIT_REACHED`, `TEAMMATE_LIMIT_REACHED`, `SEASON_NOT_ACTIVE`, or `MISSION_DEFINITION_ALREADY_ACTIVE`) with no partial writes. Reaching a terminal state releases the contract from the player's active set and frees the reservation of each member still active on it — never the reservation of a departed member, which may already belong to a newer attempt.

#### SCN-MB-CONCURRENT-ACCEPTANCE — Protect the last slot

Given the player has one active-set slot remaining, when two acceptances compete, then exactly one obtains the contract and the other receives an explainable rejection.

#### SCN-MB-UNIQUE-CONCURRENT — Concurrent same-definition accepts serialize

Given a player with free active-set slots holds no contract of definition X, when two acceptances for offers of X compete for that player, then exactly one contract is created and the other is rejected with `MISSION_DEFINITION_ALREADY_ACTIVE`.

#### SCN-MB-UNIQUE-OTHERS — Different players are independent

Given players `player-1` and `player-2` both have free slots, when each accepts an offer of the same definition concurrently, then both contracts are created.

### REQ-MB-014 — Accept only within a live season

Acceptance is rejected when the offer's bound season is not `ACTIVE` (`FREEZING`, `FINALIZED`, `ARCHIVED`, or a dangling season reference), with code `SEASON_NOT_ACTIVE` and no contract.

#### SCN-MB-SEASON-CLOSED — Reject acceptance on finalized season

Given an offer bound to a `FINALIZED` season, when a member accepts it, then the response is `422` with `SEASON_NOT_ACTIVE`.

### REQ-MB-005 — Freeze promised rules

The contract keeps the version of objectives, merits, and policies active at acceptance. Editing a definition creates a new version and does not alter existing contracts. Acceptance embeds the full `ContractTerms` snapshot — objectives, merit values, unit limits, and bonuses — not only the version reference. The offer additionally freezes `title`, `description`, `completion_reward`, `objectives`, and `rules` of the published version.

#### SCN-RULES-FROZEN — Rules are frozen per version

Given a definition published with `rules`, when an offer is created from that version, then the offer and its contracts render the rules of that version, not later edits.

#### SCN-MB-VERSION-SNAPSHOT — Respect the accepted promise

Given a player accepted a published version, when the operator publishes another version, then the contract keeps the original objectives and merits.

The frozen snapshot carries both locales: an offer created from a published version freezes the `es` and `en` title, description, rules, and objective names/descriptions of that version; acceptance embeds the same bilingual `ContractTerms` snapshot. Later version edits — including translation-only edits — never rewrite frozen offers or contracts. A frozen record created before English existed keeps its Spanish-only snapshot and reads as `es`; the reviewed-translation backfill may add `en` to an eligible frozen record but never alters its Spanish content, economics, ids, or version identity. The reviewed translation includes the source Spanish snapshot; a frozen record whose Spanish differs is left untouched and reported unresolved.

#### SCN-MB-SNAPSHOT-BACKFILL-DRIFT — Reject a divergent frozen snapshot
Given a reviewed translation for a definition version and an older offer or contract whose frozen Spanish text differs despite the same version and objective ids, when the operator rehearses its backfill, then no English is added to that record and the report identifies it as unresolved; a matching record is eligible for English-only augmentation.

#### SCN-MB-SNAPSHOT-BILINGUAL — Offer freezes both locales
Given a definition version published with Spanish and English content, when an offer is created and a member accepts it, then the offer and the contract terms keep both language snapshots of that version.

#### SCN-MB-VERSION-EDIT — A translation edit never rewrites history
Given a member holds a contract accepted at version 1 and the operator publishes version 2 with edited translations, when the offer and contract are read in either language, then they still show the version-1 text in both languages.

### REQ-MB-013 — Publish definitions with mission content and reward rules

A `MissionDefinition` declares player-facing content — `title` and `description` are required — plus a `completion_reward` (non-negative, default `0`) of competitive points granted when a contract completes. `description` is stored verbatim and defined as Markdown content clients render. An optional `rules` field carries an ordered list of non-blank requirement strings for bullet rendering; a missing or empty list is allowed and a blank item resolves as `INVALID_MISSION_DEFINITION`. Its objectives carry a required readable `name`, optional `description`, type, category, countable unit, merit value, optional `bonus`, and optional `unit_limit`. `POST /api/v1/missions/definitions` accepts all of these fields; empty required fields resolve as `422` with `INVALID_MISSION_DEFINITION`. Acceptance copies the published version's objectives into the contract terms.

#### SCN-DESC-MARKDOWN — Markdown description round-trips verbatim

Given an operator creates a definition whose `description` contains Markdown headings and lists, when the definition or its offers are read, then the same Markdown text is returned unchanged.

#### SCN-RULES-STRUCTURED — Rules resolve as an ordered list

Given an operator sends `rules` with ordered requirement strings, when the definition is created, then the rules persist in order and a blank item resolves as `422` with `INVALID_MISSION_DEFINITION`.

#### SCN-MB-DEF-OBJECTIVES — Objectives ride the definition

Given a definition with two objectives, when a player accepts an offer bound to that version, then the contract terms contain exactly those objectives and rewards.

#### SCN-MB-DEF-CONTENT — Author readable missions

Given an authenticated operator creates a definition with `title`, `description`, `completion_reward` and named objectives, when the request is valid, then the definition is created in `DRAFT` echoing all content fields; missing `title`, `description` or an objective `name`, or a negative `completion_reward`, resolves as `422` with `INVALID_MISSION_DEFINITION`.

Definition content is locale-aware: `title`, `description` (Markdown, verbatim), ordered `rules`, and each objective's `name`/`description` exist per locale; `es` remains required. The create body stays the Spanish content and accepts an optional `translations` block — `translations.en` carries `{title, description, rules, objectives: [{objective_id, name, description}]}` stored atomically with the Spanish content; objective `description` is optional exactly as `es` allows and `name` is required per entry. Draft creation stays backward compatible — the `es`-only body still succeeds. Publishing a version requires both locales structurally complete: every `es` `objective_id` appears in `en` with a `name` (and a `description` wherever `es` carries one), and the `en` rule count equals the `es` rule count in the same order — completeness is structural; an incomplete `en` resolves as `422 INVALID_MISSION_DEFINITION`. There is no automatic machine translation: every locale is operator-supplied reviewed text. Previously published Spanish-only versions stay valid and keep serving `es`; new-version publishes require English.

#### SCN-MB-CONTENT-DRAFT-EN — Draft may carry an English translation
Given an operator authors a definition with Spanish content plus a complete English translation, when the draft is saved, then both languages are stored together on the same version.

#### SCN-MB-CONTENT-INCOMPLETE — Publish requires both locales complete
Given a definition draft whose English translation is missing or structurally incomplete, when the operator publishes a new version, then the publish is rejected as an invalid definition and no version is stored.

#### SCN-MB-CATALOG-TRANSLATED — Review complete bilingual missions
Given the Spanish racing and PvP missions are versioned, when an operator reviews their English text, then each mission has an English title, description, ordered rules, and objective text paired with the exact Spanish version, without changing its identity or rewards.

#### SCN-MB-CATALOG-REDEPLOY — Preserve previously approved text
Given a published Spanish version has gained reviewed English, when the same catalog is deployed again, then both languages and their promised rewards remain unchanged, including in offers from that version. A conflicting Spanish or English snapshot is rejected, not silently replaced.

#### SCN-MB-CATALOG-NEW-VERSION — Require English on first publication
Given an operator adds a new catalog version, when it is published, then both complete languages are stored together; if the English text is missing or incomplete, no new version or offer is published.

#### SCN-MB-CATALOG-LOCAL-PUBLISH — Add approved English to local history
Given the operator has approved the English for a published mission and its current definition, version, and active offer still contain the exact reviewed Spanish, when the operator publishes that English to the local catalog, then those records serve both languages without changing their Spanish, identity, rewards, or version. Already accepted contract terms remain unchanged.

#### SCN-MB-CATALOG-LOCAL-CONFLICT — Report divergent local history
Given a published mission's local Spanish history differs from the reviewed source or its English is already different, when the operator previews local publication, then the divergent record is reported, left unchanged, and publication does not begin.

### REQ-MB-015 — Expose mission content on the board and catalog

`GET /api/v1/missions/offers` and `GET /api/v1/missions/definitions` return `Title`, `Description` (Markdown), `Rules`, and `CompletionReward` alongside the existing fields. Stored items without a `Rules` attribute decode an empty list. Offers additionally return the frozen `objectives` (each with `objective_id`, `name`, `description`, `category`, `merit_per_unit`, `bonus`, `unit_limit`) and a derived `merit_summary` (`objective_merits`, `completion_reward`, `total_merits`). `objective_merits` is the maximum earnable from objectives — `Σ (merit_per_unit + bonus) × unit_limit` — and `total_merits` adds `completion_reward`. Publishing a new version may change `title`/`description`/`rules`; objectives, `completion_reward`, `rules`, and derived totals stay frozen per version so live offers and contracts never change.

#### SCN-RULES-LEGACY — Legacy items return empty rules

Given a stored definition or offer predates `rules`, when it is read, then `Rules` resolves as an empty list with no error. Offers stored without objectives serialize an empty list and a `total_merits` equal to `completion_reward`.

#### SCN-MB-OFFER-CONTENT — Offer shows frozen content

Given a published definition with `completion_reward: 200`, when an offer is created from that version, then the offer response and board list show the frozen title, description and reward.

#### SCN-MB-OFFER-MERITS — Offer shows the earnable breakdown

Given a published definition whose objectives pay `270` and `30` merits and `completion_reward: 0`, when a member lists offers, then the offer carries both structured objectives and `merit_summary` of `objective_merits: 300`, `completion_reward: 0`, `total_merits: 300`.

`Title`, `Description`, `Rules`, and each objective's `name` and `description` resolve to the request's effective language per REQ-MB-017, and every item declares its served language with the family's label key (`ContentLanguage`). Field names and shapes are unchanged; `objective_id`, `merit_summary`, rewards, and availability are language-independent. Stored items predating English resolve `es` content declared `es`.

### REQ-MB-009 — Bind offers to the active season

A published offer is bound to the single active season. If no season is `ACTIVE`, the publish operation is rejected. The offer and all contracts created from it carry the same `season_id`.

#### SCN-MB-PUBLISH-ACTIVE — Publish with active season

Given a season `season-2026-q4` is `ACTIVE`, when an operator calls `POST /api/v1/missions/offers`, then the response is `201 Created` and the returned offer contains `season_id: "season-2026-q4"`.

#### SCN-MB-PUBLISH-NO-ACTIVE — Reject publish without active season

Given no season is `ACTIVE`, when an operator calls `POST /api/v1/missions/offers`, then the response is `422 Unprocessable Entity` with code `NO_ACTIVE_SEASON` and no offer is persisted.

### REQ-MB-008 — Hide an active template

If a player is an active member of an `ACCEPTED`, `IN_PROGRESS`, or `PENDING_VALIDATION` contract from a stable definition, the board excludes every offer of the same `definition_code`, even from another version. `PENDING_VALIDATION` still blocks the retake even when the player's active-mission count has spare slots (its slot accounting is governed by `REQ-MB-003`).

The exclusion binds to **active membership**, not roster history: once a member leaves (`REQ-MCF-013`), offers of that template reappear on their board, the old contract stops consuming their active-mission slot, and a fresh acceptance is evaluated without it — while members still active on the contract keep it hidden and counted.

Acceptance enforces the rule atomically, not only at decision time: the commit reserves each roster member's (player, definition) pair, so a concurrent duplicate of the same definition — via any offer version or as a roster teammate in another acceptance — is rejected with `MISSION_DEFINITION_ALREADY_ACTIVE` instead of creating a second contract. The reservation belongs to the contract that created it: a release succeeds only when the reservation is absent or owned by that contract, so a stale or delayed release from an older attempt can never free a newer one — finding a different owner conflicts and writes nothing instead of deleting it. The reservation is released on terminal state only for members still active on the contract, and for the departing member alone — atomically with the leave write — when a member leaves.

The departure write itself is guarded the same way: it commits only when the stored contract still matches the snapshot the member left from — the same lifecycle state and the same set of previously departed members. A leave committed on a stale snapshot conflicts and writes nothing, so parallel departures on stale snapshots can never resurrect a member whose leave already persisted.

#### SCN-MB-ACTIVE-TEMPLATE — Do not show or retake the template

Given a player has active template X, when they browse the board or attempt to accept another offer of X, then they do not see it and direct acceptance is rejected.

#### SCN-MB-TEMPLATE-RELEASED — Show after terminal state

Given the contract for template X ended, when cooldown and repeatability allow another attempt, then the player sees eligible offers of X again and may accept one.

#### SCN-MB-DEPART-VISIBLE — Departure re-shows the template to the leaver only

Given a team contract of definition X in which `player-1` departed while `player-2` stays active, when each browses the board, then `player-1` sees eligible offers of X and `player-2` does not.

#### SCN-MB-DEPART-RETAKE — Departed member retakes without the old contract

Given `player-1` left a contract of definition X and has a free slot, when they accept an offer of X, then a new contract is created and the old contract no longer counts toward their active-mission limit.

#### SCN-MB-DEPART-RACE — Stale leave write cannot restore a departed member

Given `player-1` and `player-2` each left a team contract on separate snapshots, when `player-1`'s departure has already committed, then `player-2`'s stale write conflicts, persists nothing, and releases no guards — the stored contract keeps `player-1` departed.

#### SCN-MB-UNIQUE-VERSION — Another version is still the same template

Given `player-1` has an active contract of definition X version 1, when they accept an offer of X version 2, then the acceptance is rejected with `MISSION_DEFINITION_ALREADY_ACTIVE`.

#### SCN-MB-UNIQUE-TEAMMATE — Team accept cannot seat a member who already has it

Given `player-2` already has an active contract of definition X, when a team acceptance of X names `player-2` in its roster, then the acceptance is rejected and no contract is created.

#### SCN-MB-PENDING-BLOCKS — Awaiting review still holds the template

Given `player-1`'s contract of definition X is `PENDING_VALIDATION` and their active-mission count has spare slots, when they attempt to accept an offer of X, then the acceptance is rejected with `MISSION_DEFINITION_ALREADY_ACTIVE`.

### REQ-MB-010 — List mission offers

`GET /api/v1/missions/offers` returns the offers in the single scope. `status` accepts `AVAILABLE` (default, offers open for acceptance) or `ALL`; filters include `season_id` (defaults to the single active season), `category`, and `mode`. Offers whose stable `definition_code` the player already has active are excluded per `REQ-MB-008`. Each offer carries `definition_code` and `definition_version`. When no season is `ACTIVE` and no `season_id` is given, the response is `422 Unprocessable Entity` with `NO_ACTIVE_SEASON`; `ACTIVE_SEASON_EXISTS` signals a broken single-active-season invariant. Infrastructure failures remain `500 INTERNAL_ERROR`.

#### SCN-MB-LIST-ACTIVE — Default to active season

Given season `season-2026-q4` is `ACTIVE` and holds published offers, when a member calls `GET /api/v1/missions/offers`, then the response is `200 OK` with the available offers for `season-2026-q4`, excluding offers for definitions the member already has active.

#### SCN-MB-LIST-FILTER — Filter board entries

Given published offers across seasons and categories, when a member calls `GET /api/v1/missions/offers?season_id={id}&status=ALL&category=BOUNTY`, then the response contains only matching offers.

#### SCN-MB-LIST-INVALID — Reject unknown filters

Given an unknown `status` or `mode`, when a member lists offers, then the response is `400 Bad Request` with code `INVALID_QUERY_PARAMETER`.

### REQ-MB-011 — Accept an offer only as the authenticated member

`POST /api/v1/missions/offers/{offer_id}/accept` resolves `player_id` (and `team_ids`) from the caller session. Identity fields in the request body are ignored. A caller can never produce a contract for another player. `team_id` remains an option field validated against the caller's `team_ids` claim (`NOT_TEAM_MEMBER` when it is not).

#### SCN-SEC-ACCEPT-DERIVE — Identity comes from the session

Given an authenticated member `player-1`, when they accept an offer while sending `player_id: "player-9"` in the body, then the contract is created for `player-1`.

All definition lookups, including internal lookups performed during acceptance and confirmation, resolve by `definition_code` in the single shared catalog — an unknown code resolves as `DEFINITION_NOT_FOUND`.

#### SCN-REM-TENANT-AUTH — Authenticate without a tenant claim

Given an OIDC token carrying `player_id`, `roles`, and `team_ids` but no `tenant_id`, when the caller hits any authenticated route, then the request succeeds and `CallerIdentity` is built without a tenant.

### REQ-MB-016 — Read a single offer

`GET /api/v1/missions/offers/{offer_id}` returns `200` with the published offer — the same projection as the list items. An authenticated member may read it; unknown ids return `404` with `OFFER_NOT_AVAILABLE`. It accepts `lang` per REQ-MB-017 and returns the same localized projection as the list items, with the language label at top level.

#### SCN-MB-OFFER-GET — Fetch existing offer

Given offer `offer-1` exists for the active season, when a member calls `GET /offers/offer-1`, then it returns `200` with the offer.

#### SCN-MB-OFFER-404 — Unknown offer

Given `offer-x` does not exist, when it is requested, then it returns `404` `OFFER_NOT_AVAILABLE`.

### REQ-MB-017 — Select the language of mission content

`GET /api/v1/missions/definitions`, `GET /api/v1/missions/offers`, and `GET /api/v1/missions/offers/{offer_id}` accept `?lang=es|en`. An omitted `lang` resolves as `es` — the historical default. An unsupported value, an empty value, or a repeated `lang` whose values conflict resolves as `400 INVALID_QUERY_PARAMETER`; a repeated `lang` whose values agree resolves as that value. Language is a URI query variant: caches key on the URI and `Accept-Language` is never negotiated.

Every mission representation declares the language actually served: single-resource reads carry the label at top level and every item in a collection carries its own (`ContentLanguage`). A `Content-Language` response header is emitted when — and only when — every representation in a non-empty payload resolves the same language; heterogeneous and empty collections omit it rather than falsely describing the payload.

A record that stores only Spanish serves its Spanish content declared as Spanish under any `lang`, including `lang=en` — never `406` and never English-labeled Spanish. Offer search `q` matches each offer's effective-language title and description — including the Spanish fallback — before ordering and pagination. Identifiers, amounts (`merit_summary`, `completion_reward`, merit values), and states are language-independent. `GET /api/v1/missions/offers/facets` does not accept `lang` and is unchanged.

#### SCN-MB-LANG-EN — Member reads a bilingual mission in English
Given a published offer whose definition version has complete Spanish and English content, when a member reads the board or the offer in English, then the title, Markdown description, ordered rules, and objective names and descriptions render in English — declared as English — while the offer id, objective ids, and merits stay stable.

#### SCN-MB-LANG-DEFAULT — Omitted language serves Spanish
Given a bilingual offer, when a member reads the board without choosing a language, then the content renders in Spanish and is declared as Spanish.

#### SCN-MB-LANG-LEGACY — Spanish-only record falls back declared
Given an offer or definition stored before English existed, when a member asks for English, then the Spanish content is served declared as Spanish — not an error and not declared as English.

#### SCN-MB-LANG-INVALID — Invalid language choice rejected
Given a member asks for an unsupported, empty, or self-contradicting language selection, when the request is evaluated, then it is rejected as an invalid request and no mission content is served.

#### SCN-MB-LANG-MIXED — Mixed collection declares per item
Given a board page where some offers serve English and legacy offers serve Spanish, when a member reads the board in English, then each item declares the language it actually serves and the response claims no single language for the whole page.

#### SCN-MB-LANG-HEADER — Homogeneous page declares one language
Given a single-offer read or a page where every item serves English, when the member asks for English, then the response declares English as its one language; an empty page declares no language at all.

#### SCN-MB-LANG-SEARCH — Search matches the effective language
Given an offer whose Spanish and English titles differ, when a member searches the board in English, then matching runs against the English text — and against the Spanish fallback for legacy records — before results are ordered and paged.

### REQ-MB-018 — Preview mission-board phases locally

Local off-season previews have no available offers; fresh-start and mid-season previews reuse the published catalog's frozen definitions and offers without creating a second catalog or changing the public offer API. The preview datasets remain separate from the reviewer's existing missions dataset.

#### SCN-MB-LOCAL-BOARD — Show available work at the selected phase

Given the reviewer selects a local seasonal preview, when they open the mission board, then off-season has no active offers and fresh-start or mid-season shows the catalog offers for that preview season.

## HTTP Boundary

- `POST /api/v1/missions/definitions` — body accepts `objectives` (`objective_id`, `category`, `merit_per_unit`, `bonus`, `unit_limit`) and `rules` (ordered non-blank strings); `description` is Markdown.
- `POST /api/v1/missions/definitions/{definition_code}/publish`
- `GET /api/v1/missions/offers` — paginated envelope `{offers,total,limit,offset}`; multi-value `category`/`mode`, `q`, `limit`/`offset`, `sort`. `lang=es|en` selects the served language (default `es`).
- `GET /api/v1/missions/offers/facets` — distinct `categories`/`modes` present under the resolved season and status.
- `GET /api/v1/missions/offers/{offer_id}` — single offer read; `404 OFFER_NOT_AVAILABLE` for unknown ids.
- `POST /api/v1/missions/offers` — body accepts `offer_id`, `definition_code`, `definition_version`; unrecognized fields are ignored.
- `POST /api/v1/missions/offers/{offer_id}/accept` — `201` returns the contract including its attempt-scoped `ContractID` and frozen `Terms`.

## Invariants

- A player never exceeds their effective active-mission limit, counted against the total active set across all categories; concurrent acceptances serialize through a conditional transaction — a check-then-write race cannot overspend the limit.
- Reaching a terminal contract state releases the player's active-set slot.
- An accepted contract preserves the frozen objective and reward snapshot that existed at acceptance; every acceptance mints a new attempt-scoped contract id.
- The board operates in a single scope; offer ids and definition codes resolve globally or not at all.
- Caller identity (player, team membership) is derived exclusively from the authenticated session; request-body identity fields are ignored.
- A stable definition already active for a player is hidden from that player.
- Language never alters economics, identity, ordering, or authorization: `?lang` selects only which content renders, and every frozen snapshot preserves both locales immutably.

## Implementation Notes

- Offer validation lives in the domain constructor `NewValidMissionOffer`; the application layer no longer checks offer fields manually.
- `AcceptMissionHandler` loads the offer's bound `Season`, builds `ActiveContractSummary` values via `MissionContract.Summary` for the `DecideAcceptance` policy, then commits through the `AtomicAcceptanceWriter` port.
- `ContractStore.CommitAcceptance` runs `TransactWriteItems`: `Put` contract + `Update` the per-player `PLAYER_ACTIVE` StringSet item (`ADD contract_ids` conditioned `size() < limit`). `ConditionalCheckFailed` maps to `ErrActiveMissionLimitReached` by cancellation index.
- `ContractStore.Save` on a terminal contract `DELETE`s the contract id from each roster member's active set — an idempotent release that keeps the race guard accurate without drift.

## Related Capabilities

- `mission-contract-fulfillment` executes accepted contracts.
- `seasonal-category-competition` owns seasons and standings.
