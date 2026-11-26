---
capability_id: CAP-MISSION-CONTRACT-FULFILLMENT
title: Mission Contract Fulfillment
status: current
last_consolidated_change: CHG-0055-local-season-preview-scenarios
updated: 2026-11-23
---

# Mission Contract Fulfillment

## Intent

A participant or team completes verifiable objectives inside an accepted mission contract and receives fair seasonal merit in the right category and track.

## Scope

- Progress through sequential contract phases and required/optional objectives.
- Confirm countable units while the contract is `IN_PROGRESS` and the season is `ACTIVE`.
- Credit seasonal merit per season, category, track, and recipient.
- Reject duplicate confirmations, terminal contracts, and out-of-season credits.

## Non-Goals

- Generic objective rule engine.
- Synchronous integration with Achievements.
- Disputes or corrections (see CHG-0001). Manual review of contract completion is core to this capability, not an exception.
- Procedural contract generation.

## Key Concepts

- `MissionContract`: accepted offer, roster, phase, and terminal state.
- `Objective`: published, category-bound unit with merit value.
- `Confirmation`: an attempt to credit a concrete unit.
- `ConfirmationID`: composite value object (`contract:objective:unit:recipient`) for idempotent credit identity and duplicate detection.
- `MeritCredit`: immutable ledger fact; carries an optional `MediaURI`.
- `MediaURI`: validated evidence-object reference `evidence/{contract_id}/{object_id}`.
- `MeritLedger`: append-only credits per bucket.
- `Standing`: projection derived from the ledger, not a second source of truth.

## Requirements

### REQ-MCF-001 — Execute solo or team

`SOLO` credits the individual participant; `TEAM` credits the team; `FLEX` selects a mode at acceptance and cannot change later. `team_id` is meaningful only for `TEAM`/`FLEX` offers — supplying it on a `SOLO` offer is rejected with `TEAM_NOT_ALLOWED`. A team contract's roster is the requester plus the validated teammates from the declared team (`TEAMMATE_NOT_MEMBER` when a declared teammate is not in the team); each confirmed member consumes one active-contract slot in their personal limit (`TEAMMATE_LIMIT_REACHED` when any member is at their limit). Team credits land on the team track only.

Teammate validation reads the authoritative team roster from Party Manager (`GET {PARTY_MANAGER_URL}/teams/{team_id}/members/{player_id}`) at acceptance time — request input alone never proves membership. A definitive negative resolves as `422 TEAMMATE_NOT_MEMBER`; an unreachable, timed-out, or non-definitive source fails closed as `503 MEMBERSHIP_UNAVAILABLE` and no contract is created. When `PARTY_MANAGER_URL` is not configured, composition keeps the deny-all default, so TEAM/FLEX-as-team acceptance stays rejected rather than permissive.

#### SCN-MCF-TEAM — Freeze an eligible roster

Given all invitees are within their active-mission limit, when they confirm and the mission starts, then the roster is frozen and the contract uses the team track.

#### SCN-MCF-TEAM-SOLO-REJECT — Team id on a solo offer

Given a `SOLO` offer, when a member accepts it with a `team_id`, then the acceptance is rejected with `TEAM_NOT_ALLOWED` and no contract is created.

#### SCN-MCF-TEAM-ROSTER — Validate teammates and their limits

Given a `TEAM` offer and teammates `player-2`, `player-3` of the caller's team, when one teammate is already at their active limit, then the acceptance is rejected with `TEAMMATE_LIMIT_REACHED` and no contract is created.

#### SCN-MCF-TEAM-CONTRACT — Team contract consumes every member's slot

Given a valid team acceptance with two teammates, when the contract is created, then all three members consume one active slot and the contract uses the `TEAM` track.

#### SCN-MCF-MEMBER-VALID — Verified teammates accepted

Given a TEAM offer and teammates that Party Manager confirms as members of the caller team, when the member accepts, then the contract is created with the full roster.

#### SCN-MCF-MEMBER-UNAVAILABLE — Membership source down fails closed

Given Party Manager is unreachable or returns a non-definitive response, when a member accepts a TEAM offer with teammates, then the acceptance is rejected with `503 MEMBERSHIP_UNAVAILABLE` and no contract is created.

### REQ-MCF-002 — Progress by phases and objectives

A contract is a sequence of `phases`; each declares `mandatory` and `optional` objectives and the contract exposes `current_phase`. Only objectives of the current phase are confirmable — a unit for a later phase returns `422 PHASE_NOT_UNLOCKED`. Confirming the last mandatory objective of a phase advances `current_phase` lazily; optional objectives never gate progression. A completed final phase stays eligible — remaining units may still confirm. Completion (`COMPLETION` credit, REQ-MCF-020) requires all mandatory objectives across all phases. Flat legacy terms deserialize as a single all-mandatory phase, preserving prior behavior.

#### SCN-MCF-OBJECTIVES — Unlock next phase

Given a phase requires its mandatory objectives, when all are confirmed, then the next phase becomes available without requiring optional objectives.

#### SCN-MCF-PHASE-ORDER — Out-of-phase confirmation rejected

Given the contract is in phase 1, when a unit for a phase-2 objective is submitted, then it returns `422` `PHASE_NOT_UNLOCKED`.

#### SCN-MCF-PHASE-COMPLETE — Mandatory completion earns the reward

Given phase 1 has an optional objective unmet and all mandatory objectives across both phases are confirmed, when a reviewer approves, then the `COMPLETION` credit is emitted.

### REQ-MCF-003 — Confirm countable units

A unit is valid only when the contract was active, the objective was eligible, it occurred inside the window, the recipient was entitled, the evidence or source was accepted, and the supplied season matches the contract's bound season. When `media_uri` is supplied, it must reference the contract's `evidence/{contract_id}/` prefix and resolve to an existing stored object; the reference is persisted on the resulting `MeritCredit`. The confirmation returns the computed `MeritCredit` projection but does **not** persist it to the ledger — the credit stays staged on the contract until a reviewer approves it (`REQ-MCF-020`).

#### SCN-MCF-001 — Credit each confirmed bounty

Given a Bounty PvP mission offers 100 merits per unique target, when four valid objectives are confirmed, then the participant receives 400 merits in `BOUNTY_HUNTING` for the active season.

#### SCN-MCF-SEASON-MATCH — Reject confirmation for the wrong season

Given a contract is bound to `season-2026-q4` and a request supplies `season-2026-q3`, when the confirmation is processed, then it returns `SEASON_MISMATCH` and no credit is recorded.

#### SCN-MCF-EVIDENCE-MEDIA — Confirmation carries the media reference

Given `player-1` uploaded an object under `evidence/contract-1/{object_id}`, when they confirm `unit_id` with that `media_uri`, then the credit records `media_uri` and the `201` response echoes it.

#### SCN-MCF-EVIDENCE-MEDIA-PHANTOM — Phantom object rejected

Given no object exists at `media_uri`, when the confirmation is submitted, then it is rejected with `EVIDENCE_MEDIA_INVALID` and no credit is recorded.

#### SCN-MCF-EVIDENCE-MEDIA-FOREIGN — Foreign prefix rejected

Given `media_uri` points under another contract's prefix, when the confirmation is submitted, then it is rejected with `EVIDENCE_MEDIA_INVALID`.

### REQ-MCF-004 — Prevent duplicates and basic farming

The same unit, objective, contract, and recipient are credited at most once. Objective `unit_limit` and `bonus` persist and are enforced; duplicate detection uses a delimiter-safe `ConfirmationID` that always parses even when ids contain colons. A unit rejected at adjudication leaves the contract's confirmed set before fulfillment is evaluated — it can neither pay nor count toward a `unit_limit`.

#### SCN-MCF-DUPLICATE — Ignore a replay

Given a target already produced a credit, when the same confirmation arrives again, then the ledger and standing do not change.

#### SCN-MCF-UNIT-LIMIT — Enforce the unit cap

Given an objective with `unit_limit: 2` and two confirmations recorded, when a third confirmation arrives, then it is rejected with `UNIT_LIMIT_EXCEEDED`.

### REQ-MCF-008 — Progress a contract through its lifecycle

An `ACCEPTED` contract can be started by a roster member (`IN_PROGRESS`), submitted for review (`PENDING_VALIDATION`), adjudicated by a reviewer (`COMPLETED` or `DISQUALIFIED`), abandoned by the player (`ABANDONED`), or expired when the window closes (`EXPIRED`). Terminal states are irreversible and keep reason and evidence history. `PENDING_VALIDATION` is non-terminal but closed to evidence and media grants; `abandon` and `leave` still apply. A roster member may also **leave** a non-terminal contract individually without terminating it for the rest (see `REQ-MCF-013`). Member transitions (`start`, `submit`, `evidence`) are restricted to members still active on the contract — departures are excluded before the roster-member check passes.

#### SCN-MCF-START — Start an accepted contract

Given player `player-1` holds an `ACCEPTED` contract, when they start it, then the contract is `IN_PROGRESS` and objectives become confirmable.

#### SCN-MCF-SUBMIT — Member submits for review

Given an `IN_PROGRESS` contract for `player-1`, when they call `submit`, then the contract is `PENDING_VALIDATION` and their active slot is freed.

#### SCN-MCF-SUBMIT-PARTIAL — Partial submission allowed

Given a contract with one of two required objectives confirmed, when a member submits, then it is `PENDING_VALIDATION` awaiting review.

#### SCN-MCF-PENDING-CLOSED — No evidence while pending

Given a `PENDING_VALIDATION` contract, when a member submits evidence or requests an upload grant, then it is rejected with `CONTRACT_NOT_IN_PROGRESS`.

#### SCN-MCF-ABANDON — Abandon releases the slot

Given an `IN_PROGRESS` contract, when the player abandons it, then it is `ABANDONED` and no longer counts toward the active limit.

### REQ-MCF-009 — Score against frozen contract terms

Confirmation resolves the objective and merit value from the contract's `ContractTerms` snapshot captured at acceptance. Later definition or offer edits do not change what an existing contract pays.

#### SCN-MCF-FROZEN — Republished definition does not alter a live contract

Given a contract was accepted at version 1, when the operator republishes version 2 with different merits, then the live contract still scores version-1 objectives and merits.

#### SCN-MCF-WRONG-OBJECTIVE — Reject an objective not in the snapshot

Given a contract's snapshot has objective `obj-a`, when evidence references `obj-z` (same category, different objective), then it is rejected with `OBJECTIVE_NOT_IN_CONTRACT` and no credit is recorded.

### REQ-MCF-010 — Commit adjudication and credits atomically

A confirmation stages its `MeritCredit` on the contract item itself — never a partial ledger state. Approving a pending contract commits the new contract state and appends every materialized `MeritCredit` as one atomic operation. If the write fails, no partial state remains; a retry of the same logical confirmation or approval produces the same single credit set.

Every contract write — the approval write and the ordinary non-departure save — is fenced on the roster truth the writer saw: the contract put is conditional on the stored departed set exactly matching the candidate's already-departed members (length and membership; a missing departed attribute is tolerated only when the expected set is empty). A stale write whose snapshot predates a committed departure conflicts and writes nothing: no contract overwrite and no materialized credits. Credit-item conditional failures still resolve as the duplicate-confirmation rejection.

While a season's adjudication lock is engaged (see `seasonal-category-competition` REQ-SCC-004), every rank-changing credit write MUST be rejected atomically in the same write — contract and ledger untouched — surfacing `422 SEASON_ADJUDICATING`, never a false 200; a post-`FINALIZED` retry cannot add to the standing. Pre-lock, eligible `FREEZING` credit writes proceed unchanged. Zero-credit writes changing no tie-break-relevant state are NOT blocked.

#### SCN-MCF-ADJ-001 — Pre-lock approval with eligible credit succeeds
Given a `FREEZING` season whose lock is NOT yet engaged and a `PENDING_VALIDATION` contract, when a moderator approves, then credits commit normally — freeze holds until finalization starts.

#### SCN-MCF-ADJ-002 — Approval racing the adjudication cutoff
Given adjudication is starting, when an approval write races the lock, then it either commits fully before the cutoff or is rejected atomically — contract stays `PENDING_VALIDATION`, ledger unchanged.

#### SCN-MCF-ADJ-003 — Completion reward writes nothing after lock
Given the lock is engaged, when an approval would grant a completion reward, then it is rejected with `SEASON_ADJUDICATING` — no credit or contract transition is written.

#### SCN-MCF-ATOMIC-RETRY — Retry after ledger failure

Given a confirmation or approval failed mid-write, when it is retried, then it completes once and the standing reflects exactly one credit per staged unit.

#### SCN-MCF-STALE-SAVE — Stale ordinary save cannot restore a departed member

Given `player-1` left a contract and a snapshot taken before the leave is written through the ordinary save path, when the write runs, then it conflicts, persists nothing, and the stored contract keeps `player-1` departed.

#### SCN-MCF-STALE-APPROVAL — Stale approval cannot resurrect a member or mint credits

Given `player-1` left a `PENDING_VALIDATION` contract and a reviewer approves on a snapshot taken before the leave, when the approval write runs, then it conflicts atomically — the contract keeps `player-1` departed and no credit items are written.

### REQ-MCF-011 — Give every attempt its own identity

Each acceptance generates a unique `ContractID` (attempt identity), so re-accepting an offer after a terminal attempt creates a new record instead of overwriting history.

#### SCN-MCF-ATTEMPT — Re-accept keeps both attempts

Given `player-1` abandoned a contract for offer `off-1`, when they accept it again, then a second contract with a distinct `contract_id` is created and the abandoned attempt remains queryable.

### REQ-MCF-012 — Completion grants the frozen reward

When a reviewer approves a `PENDING_VALIDATION` contract, the same atomic write that materializes the staged credits also records a `MeritCredit` for the frozen `completion_reward` under `objective_id` `COMPLETION` with the contract's season, category, recipient, and track — but only when every mandatory objective across all phases is **fully satisfied**: its confirmed-and-not-rejected unit count reaches `unit_limit` (or at least one when `unit_limit` is `0`/unlimited). The check runs after rejected units are dropped, so a rejected unit always blocks the bonus. A `completion_reward` of `0` records no credit. A retried approval is idempotent — the deterministic completion identity credits the reward at most once. The `COMPLETION` credit is a positive-credit write fenced identically to REQ-MCF-010: post-lock it is rejected and writes nothing; pre-lock it commits per the atomic rule.

#### SCN-MCF-COMPLETION-REWARD — Approval credits the reward

Given a `TEAM` contract on a definition version with `completion_reward: 200` whose frozen objectives are fully satisfied, when a moderator approves the pending contract, then the team's season standing increases by `200` and the contract is `COMPLETED` in one write.

#### SCN-MCF-COMPLETION-ONCE — Reward credits at most once

Given a completed contract already recorded its completion credit, when the approval is retried, then the contract stays terminal and the standing keeps exactly one `COMPLETION` credit.

#### SCN-MCF-REJECT-BLOCKS-BONUS — Rejected unit forfeits the bonus

Given a pending contract whose objective requires `unit_limit: 10` and only nine staged units survive the verdict, when a moderator approves, then nine unit credits are written and no `COMPLETION` credit exists.

#### SCN-MCF-FULL-SATISFACTION-BONUS — Full satisfaction emits the bonus

Given a pending contract whose `unit_limit: 10` objective has all ten units staged and a verdict rejecting none, when a moderator approves, then ten unit credits plus one `COMPLETION` credit are written atomically.

#### SCN-MCF-PARTIAL-STAGE-NO-BONUS — Unstaged units also block the bonus

Given a pending contract whose objective requires ten units but only seven were ever staged, when a moderator approves with no rejects, then seven credits are written and no `COMPLETION` credit exists.

### REQ-MCF-013 — Leave a team contract without terminating it

`POST /api/v1/missions/contracts/{contract_id}/leave` removes the caller from the contract's active membership when the contract is non-terminal and the caller is a current (non-departed) roster member. The response is `200`. An outsider or an already-departed member resolves as `403 FORBIDDEN`; a terminal contract resolves as `422` with `CONTRACT_TERMINAL`. When the caller is the last active member, the contract transitions to `ABANDONED` — the solo-contract leave equals abandon. Departures are tracked separately from the accepted roster, which is preserved as history; already-recorded credits remain untouched.

Every member exit — including the last active member, whose leave transitions the contract to `ABANDONED` — commits through one focused departure write that atomically persists the contract and releases the departing member's guards: their per-definition reservation and their `PLAYER_ACTIVE` contract id. Guards of members who departed earlier were released at their own leaves and are never touched by a later departure. The accepted roster and the departed set stay preserved as history (`REQ-MB-008` governs the board consequence: the departed member may see and retake the definition under a free slot).

The departure write is optimistic: it commits only when the stored contract still matches the state the member left from — the same lifecycle state and the exact set of previously departed members. A leave attempted on a stale snapshot conflicts and writes nothing, so no departure can resurrect or silently drop another member's recorded exit.

#### SCN-MCF-LEAVE-MEMBER — Member exits, contract survives

Given a `TEAM` contract in `IN_PROGRESS` with roster `[p1, p2]`, when `p1` leaves, then the contract stays `IN_PROGRESS`, `p1` appears in the departed set, and `p2` may still submit evidence.

#### SCN-MCF-LEAVE-BLOCKED — Departed loses contract rights

Given `p1` left a team contract, when `p1` submits evidence, starts, submits for review, or leaves again, then every operation is rejected with `403 FORBIDDEN`.

#### SCN-MCF-LEAVE-LAST — Last member abandons

Given a team contract whose only remaining member is `p2`, when `p2` leaves, then the contract is `ABANDONED` and terminal.

#### SCN-MCF-LEAVE-OUTSIDER — Outsider cannot leave

Given `p3` was never in the roster, when `p3` calls leave, then the response is `403 FORBIDDEN` and the contract is unchanged.

#### SCN-MCF-LEAVE-LAST-RETAKE — Last member's leave frees their own guards

Given a contract of definition X whose only active member is `player-1`, when they leave, then the contract is `ABANDONED`, their reservation and active-set entry are released in the same atomic write, and a fresh acceptance of X is not blocked by the old attempt.

### REQ-MCF-005 — Preserve terminal outcomes

A contract ends `COMPLETED`, `FAILED`, `ABANDONED`, `EXPIRED`, `CANCELLED`, or `DISQUALIFIED`. No terminal state returns to active and the history keeps reasons and evidence.

#### SCN-MCF-TERMINAL — Reject confirmation on terminal contract

Given a contract is in a terminal state, when a new confirmation arrives, then it is rejected without creating or modifying credits.

### REQ-MCF-006 — List player contracts

`GET /api/v1/missions/contracts` returns the contracts where the authenticated player is in the roster. Filters: `season_id` and `status` (`ALL` default, `ACTIVE` bucket for non-terminal, or an exact `ContractState`). Each item's embedded `Offer` carries the contract's frozen `objectives` and a `merit_summary` derived from `Terms` (`objective_merits = Σ (merit_per_unit + bonus) × unit_limit`, `total_merits = objective_merits + completion_reward` per REQ-MB-015). Contracts with no terms (legacy) report `objective_merits: 0` and `total_merits = completion_reward`.

#### SCN-MCF-LIST-MINE — Only roster membership

Given contracts for several players, when a member lists contracts, then the response contains only contracts where they are in the roster.

#### SCN-MCF-LIST-STATUS — Filter by lifecycle

Given contracts in `IN_PROGRESS` and `COMPLETED`, when a member requests `status=ACTIVE`, then only non-terminal contracts return; `status=COMPLETED` returns exact-state matches.

#### SCN-MCF-MERIT-SUMMARY — Contract shows the earnable total

Given a contract whose `Terms` freeze objectives `(merit_per_unit: 450, unit_limit: 1)` and `(merit_per_unit: 50, unit_limit: 1)` with `completion_reward: 0`, when the member lists `GET /api/v1/missions/contracts`, then the embedded offer shows `merit_summary: {objective_merits: 500, completion_reward: 0, total_merits: 500}` and the two frozen objectives.

#### SCN-MCF-MERIT-LEGACY — Contracts without terms degrade safely

Given a stored contract with no `Terms`/`Phases` attributes, when it is listed, then `offer.objectives` is empty and `total_merits` equals `completion_reward`.

The list accepts `lang` per REQ-MCF-026: each item's embedded `Offer` and its frozen `objectives` resolve the effective language from the contract's snapshot, and the item declares the served language at the contract level (covering the embedded offer) via `ContentLanguage`. `merit_summary` values, filters, ordering, and totals are language-independent.

### REQ-MCF-007 — Confirm evidence only within caller authority

`POST /api/v1/missions/contracts/{contract_id}/evidence` resolves the player from the session. Submission — and evidence media upload grants — is authorized only when the session player is in the contract roster or the caller holds `admin`/`moderator`; otherwise the response is `403` with `FORBIDDEN`. Media read grants additionally require the `admin`/`moderator` role regardless of roster membership.

#### SCN-SEC-EVIDENCE-ROSTER — Roster member confirms

Given `player-1` is in the contract roster, when they submit evidence for that contract, then the credit is recorded for `player-1` in the season, category, and track bucket.

#### SCN-SEC-EVIDENCE-OUTSIDER — Outsider cannot confirm

Given `player-2` is not in the contract roster and holds no privileged role, when they submit evidence for the contract, then the response is `403` with `FORBIDDEN` and no credit is recorded.

### REQ-MCF-014 — Request an evidence media upload grant

`POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url` accepts `{"content_type": "...", "content_length": N}` and returns `201` with `{"upload_url", "media_uri"}`. The grant is a presigned PUT bound to the server-generated key `evidence/{contract_id}/{object_id}` and declared content type, expiring ~5 minutes. Only `image/*`/`video/*` content types are accepted (`400 INVALID_REQUEST` otherwise); a declared size above the configured maximum returns `400 INVALID_REQUEST`. The bound is governed by the `EvidenceMaxBytes` deployment parameter (default `268435456` = 256 MB) passed through `EVIDENCE_MAX_BYTES`. Authority matches `REQ-MCF-007`; a non-`IN_PROGRESS` contract resolves as `422 CONTRACT_NOT_IN_PROGRESS`.

#### SCN-MCF-UPLOAD-GRANT — Member receives a constrained grant

Given `player-1` is a roster member of an `IN_PROGRESS` contract, when they request an upload grant with `content_type: image/png`, then the response is `201` with an `upload_url` bound to the contract-scoped key and content type, and a `media_uri` under `evidence/{contract_id}/`.

#### SCN-MCF-UPLOAD-NOT-IN-PROGRESS — No grant on inactive contract

Given the contract is `ACCEPTED` or terminal, when a member requests an upload grant, then the response is `422` with `CONTRACT_NOT_IN_PROGRESS` and no grant is issued.

#### SCN-MCF-UPLOAD-OUTSIDER — Outsider cannot request a grant

Given `player-2` is not in the roster and holds no privileged role, when they request an upload grant, then the response is `403 FORBIDDEN` and no grant is issued.

#### SCN-MCF-UPLOAD-TYPE-REJECTED — Disallowed media type or oversize

Given a member requests an upload grant with `content_type: application/x-msdownload` or a `content_length` above the bound, when it is validated, then the response is `400` with `INVALID_REQUEST` and no grant is issued.

### REQ-MCF-015 — Attach media to an evidence confirmation

`POST /api/v1/missions/contracts/{contract_id}/evidence` accepts an optional `media_uri`. When present, it must reference this contract's prefix (`evidence/{contract_id}/`) and the object must exist in storage; otherwise the response is `422` with `EVIDENCE_MEDIA_INVALID`. On success the `media_uri` is stored on the recorded `MeritCredit` and echoed in the `201` response. Evidence without `media_uri` is unchanged.

### REQ-MCF-016 — Read grant for stored evidence media

`POST /api/v1/missions/contracts/{contract_id}/evidence/media-url` accepts `{"media_uri": "..."}` and returns `200` with `{"download_url"}` — a presigned GET with ~15-minute expiry — when the caller holds `admin` or `moderator` and the `media_uri` belongs to this contract and exists. A roster member without a privileged role receives `403 FORBIDDEN`; an unknown or foreign `media_uri` resolves `422` with `EVIDENCE_MEDIA_INVALID`.

#### SCN-MCF-MEDIA-READ-REVIEWER — Moderator opens the media

Given a credit recorded `media_uri` for `contract-1`, when a moderator requests a media read grant for it, then the response is `200` with a short-lived `download_url`.

#### SCN-MCF-MEDIA-READ-DENIED — Member cannot read reviewer media

Given a roster member without a privileged role requests a media read grant, when it is evaluated, then the response is `403 FORBIDDEN`.

### REQ-MCF-017 — Browser clients can execute the upload grant

The evidence bucket allows cross-origin `PUT` and `HEAD` (and `GET` for reviewer read grants) only from the configured application origins via bucket CORS configuration. Credentials and object keys stay unchanged; CORS governs browser enforcement, not authorization.

#### SCN-MCF-CORS-UPLOAD — Webapp PUT reaches S3

Given a webapp origin listed in the bucket CORS configuration, when a browser issues the presigned `PUT` for `evidence/{contract_id}/{id}`, then the preflight and the upload succeed and the object lands under the grant key.

#### SCN-MCF-CORS-UNKNOWN-ORIGIN — Unknown origin blocked by browser

Given an origin not in the CORS allowlist, when a page on that origin attempts the presigned `PUT`, then the browser blocks the response (no `Access-Control-Allow-Origin` for that origin) and the request never reaches the service.

### REQ-MCF-019 — Submit a contract for review

`POST /api/v1/missions/contracts/{contract_id}/submit` transitions an `IN_PROGRESS` contract to `PENDING_VALIDATION` for a roster member still active on it (`403 FORBIDDEN` outsider/departed; `422` `CONTRACT_STATE_INVALID` or `CONTRACT_NOT_IN_PROGRESS` otherwise). No minimum confirmation threshold — partial fulfillment is administrable. Submitting releases the member's active-contract slot while the template stays hidden (`REQ-MB-008`).

#### SCN-MCF-SUBMIT-OUTSIDER — Outsider cannot submit

Given `player-2` outside the roster, when they call `submit`, then the response is `403 FORBIDDEN` and the contract is unchanged.

### REQ-MCF-020 — Adjudicate a pending contract

`POST /api/v1/missions/contracts/{contract_id}/approve` and `/reject` require `admin`/`moderator` (`403 FORBIDDEN` otherwise) and `PENDING_VALIDATION` state (`422 CONTRACT_NOT_PENDING_VALIDATION`). `approve` accepts an optional verdict body `{"rejected_units": [{"objective_id", "unit_id"}]}`: the reviewer adjudicates each staged unit — units not listed commit their frozen credit, listed units are recorded on the contract's `rejected_units` and pay nothing. A verdict naming a unit that was never staged resolves `422 UNIT_NOT_STAGED` and writes nothing. `approve` returns `200` with the written credits, transitions to `COMPLETED`, and atomically materializes every surviving staged unit credit plus the completion reward on full fulfillment; `reject` returns `200`, transitions to terminal `DISQUALIFIED`, and writes zero credits. Retried approvals are idempotent — duplicate `ConfirmationID`s make repeated writes no-ops. Approval operates on an existing contract row: the contract put requires the row to exist and its stored departed set to match the adjudicated snapshot, so a stale review can neither resurrect a departed member nor attach credits to a roster state that no longer holds.

#### SCN-MCF-APPROVE-UNITS — Approval commits the approved subset

Given a `PENDING_VALIDATION` contract with three staged units and a verdict rejecting `unit-2`, when a moderator approves, then exactly the credits for `unit-1` and `unit-3` are written, `unit-2` appears in `rejected_units`, and the contract is `COMPLETED`.

#### SCN-MCF-APPROVE-UNITS-DEFAULT — Empty verdict approves all

Given a pending contract with staged units and an approve call without a verdict body (or an empty `rejected_units`), when a moderator approves, then all staged credits are written — the legacy approve-all behavior.

#### SCN-MCF-APPROVE-UNKNOWN-UNIT — Unknown rejected unit fails the call

Given a pending contract and a verdict naming a unit never staged, when a moderator approves, then the response is `422 UNIT_NOT_STAGED`, the contract stays `PENDING_VALIDATION`, and no credit is written.

#### SCN-MCF-APPROVE-FULL — Approval credits everything

Given a `PENDING_VALIDATION` contract with all required objectives confirmed and `completion_reward: 200`, when a moderator approves, then one atomic write records all unit credits plus the `COMPLETION` credit and the contract is `COMPLETED`.

#### SCN-MCF-APPROVE-PARTIAL — Approval credits only confirmed units

Given a pending contract with one confirmed unit and no full completion, when a moderator approves, then only that unit's credit is written, no `COMPLETION` credit exists, and the contract is `COMPLETED`.

#### SCN-MCF-REJECT — Rejection credits nothing

Given a pending contract with confirmed units, when a moderator rejects, then the contract is `DISQUALIFIED` and the standing gains zero merits.

#### SCN-MCF-REVIEW-MEMBER-DENIED — Roster member cannot adjudicate

Given `player-1` owns the pending contract but holds no privileged role, when they call `approve`, then the response is `403 FORBIDDEN`.

#### SCN-MCF-REVIEW-WRONG-STATE — Adjudication needs pending state

Given an `IN_PROGRESS` contract, when a moderator calls `approve`, then it is rejected with `422 CONTRACT_NOT_PENDING_VALIDATION`.

#### SCN-MCF-REVIEW-IDEMPOTENT — Retried approval credits once

Given an approval whose write failed mid-transaction, when it is retried, then exactly one credit set exists — duplicate `ConfirmationID`s make repeated unit writes no-ops.

While a season's adjudication lock is engaged, `approve` paths writing ranking-affecting credits are fenced per REQ-MCF-010 (rejected atomically with `SEASON_ADJUDICATING`); `reject` writes zero credits and stays allowed unless the transition mutates tie-break-relevant state. Retried approvals are idempotent; a post-`FINALIZED` retry cannot add to the standing.

### REQ-MCF-021 — List the pending review queue

`GET /api/v1/missions/review-queue` returns `200` with the contracts in `PENDING_VALIDATION`, each carrying `contract_id`, `offer_id`, `definition_code`, `title`, `state`, `roster`, `season_id`, `team_id`, the confirmed-unit count, and `pending_credits` (staged credit count). Requires `admin` or `moderator` — any other caller receives `403 FORBIDDEN`; an empty queue returns `200` with `[]`. The member-facing `GET /contracts` stays roster-scoped (`REQ-MCF-006`) — reviewer discovery is a separate use case.

#### SCN-MCF-QUEUE-REVIEWER — Reviewer sees the pending queue

Given contracts `c1` (PENDING_VALIDATION) and `c2` (IN_PROGRESS) exist, when a moderator requests the review queue, then the response is `200` with `c1` only, including its staged-credit count.

#### SCN-MCF-QUEUE-MEMBER-DENIED — Member cannot list the queue

Given `player-1` holds no privileged role, when they request the review queue — even while owning a pending contract — then the response is `403 FORBIDDEN`.

#### SCN-MCF-QUEUE-EMPTY — Empty queue returns an empty list

Given no `PENDING_VALIDATION` contracts exist, when a moderator requests the review queue, then the response is `200` with `[]`.

#### SCN-MCF-QUEUE-DRAINS — Queue reflects adjudication

Given a pending contract, when a moderator approves or rejects it, then a subsequent queue request no longer contains it.

The queue accepts `lang` per REQ-MCF-026: each entry's `title` resolves the effective language from the contract snapshot with a `content_language` label per entry; role enforcement, fields, and counts are unchanged.

### REQ-MCF-018 — Evidence store adapter verified against an emulator

CI runs a `localstack` S3 service in `docker-compose.test.yml`; the infrastructure contract test performs presign-PUT → upload → `Exists` → presign-GET round-trips against it, including the phantom-key negative.

#### SCN-MCF-ADAPTER-ROUNDTRIP — Grant, upload, read back

Given the emulator bucket exists, when the adapter issues an upload grant and a client PUTs bytes to it, then `Exists` returns true and a read grant downloads identical bytes.

### REQ-MCF-025 — Confirm team invitations before roster freeze

A TEAM acceptance creates the contract with the requester `CONFIRMED` and every other invitee `PENDING`; each pending member's active-contract slot is reserved in the same atomic acceptance write. Invitees answer via `POST /api/v1/missions/contracts/{contract_id}/confirm-invitation` or `decline-invitation` (200 on success). A non-roster caller is rejected `403 NOT_INVITED`; a resolved invitation answers `422 INVITATION_NOT_PENDING`. The invitation window is `7` days and resolves lazily on the next contract command — a `PENDING` member past the window becomes `EXPIRED`. Declined and expired members leave the roster and release their reserved slot; the contract starts with the remaining roster rather than being cancelled. `roster_frozen` becomes true once no member is pending. The reviewer contract detail exposes `invitations` per member. Solo contracts carry no invitations and freeze at creation.

#### SCN-MCF-INVITE-CONFIRM — All confirm

Given a TEAM acceptance with two invitees, when both confirm, then the roster freezes and the contract starts with all members.

#### SCN-MCF-INVITE-DECLINE — Decline releases slot

Given an invitee declines, when the contract activates, then the declined member is absent and their slot released.

#### SCN-MCF-INVITE-EXPIRE — Window lapses

Given an invitee never responds, when the confirmation window lapses, then the pending member is dropped and the slot released.

### REQ-MCF-022 — Read own contract detail

`GET /api/v1/missions/contracts/{contract_id}` is readable by adjudicators and by active roster members. Reviewers receive the full projection; members receive a minimized view whose `pending_credits` are filtered to their own `recipient_id` and never carry `media_uri` — evidence media stays reachable only through the reviewer read-grant flow. The projection also exposes `rejected_units` — `{objective_id, unit_id, recipient_id}` entries recorded by the adjudicating approve; reviewers see all of them, members only those staged under their own recipient. The embedded offer exposes the same frozen `objectives` and `merit_summary` as the list projection. Departed members and outsiders receive `403 FORBIDDEN`; unknown ids return `404`.

#### SCN-MCF-DETAIL-MEMBER — Member reads own contract

Given a contract with roster [`player-1`], when `player-1` requests `GET /contracts/{id}`, then it returns `200` with the member projection — their own staged credits without media references.

#### SCN-MCF-DETAIL-OUTSIDER — Non-member denied

Given a contract with roster [`player-1`], when `player-2` requests it, then it returns `403` `FORBIDDEN`.

#### SCN-MCF-DETAIL-DEPARTED — Departed member denied

Given `player-1` left the contract, when they request it, then it returns `403` `FORBIDDEN`.

The detail accepts `lang` per REQ-MCF-026: `title` and frozen objective `name`/`description` resolve the effective language with a top-level `content_language`. The member-vs-reviewer projections, `pending_credits`/`rejected_units` filtering, and access rules are unchanged.

### REQ-MCF-023 — Manage active-contract limit

`PUT /api/v1/missions/policies/active-contract-limit` persists the global active-contract `default` plus optional per-category `overrides` (`{"default": N, "overrides": {"CATEGORY": M}}`). Only `admin`/`moderator` may update it (`403 FORBIDDEN` otherwise); every value must be positive (`422 INVALID_POLICY`). Acceptances read the persisted policy per request — tuning takes effect without restart. An absent document reads as the shipped default of 5.

#### SCN-MCF-POLICY-UPDATE — Admin sets limit

Given the global limit is 5, when an admin PUTs limit 3, then subsequent acceptances enforce 3.

#### SCN-MCF-POLICY-DENIED — Member cannot update

Given a non-admin caller, when they PUT the policy, then it returns `403` `FORBIDDEN`.

#### SCN-MCF-POLICY-INVALID — Non-positive limit rejected

Given an admin PUTs `default: 0`, then it returns `422` `INVALID_POLICY`.

### REQ-MCF-024 — Dispute an adjudication outcome

A roster member opens a dispute (`POST /api/v1/missions/contracts/{contract_id}/disputes` with `{"objective_id", "reason"}`) targeting a specific objective from the contract's frozen terms; the dispute starts `OPEN`. While any dispute is `OPEN`, `leave`/`abandon` on the contract returns `422 DISPUTES_PENDING`. Moderators resolve via `POST .../disputes/{dispute_id}/resolve` (`admin`/`moderator` only); resolution is immutable — re-resolving returns `422 DISPUTE_STATE_INVALID`. Reviewers list disputes via `GET .../disputes` (`403 FORBIDDEN` otherwise); unknown ids return `404 DISPUTE_NOT_FOUND`. Outsiders raising a dispute get `403 FORBIDDEN`; an objective outside the terms gets `422 OBJECTIVE_NOT_IN_CONTRACT`.

#### SCN-MCF-DISPUTE-RAISE — Member raises dispute

Given a rejected contract of player-1, when they raise a dispute with a reason, then it is OPEN and visible to reviewers.

#### SCN-MCF-DISPUTE-BLOCK — Leave blocked while pending

Given a contract with an OPEN dispute, when a member leaves, then it returns `422 DISPUTES_PENDING`.

#### SCN-MCF-DISPUTE-RESOLVE — Moderator resolves

Given an OPEN dispute, when a moderator resolves it, then the dispute is RESOLVED and exit is unblocked.

### REQ-MCF-026 — Select the language of contract content

`GET /api/v1/missions/contracts`, `GET /api/v1/missions/contracts/{contract_id}`, and `GET /api/v1/missions/review-queue` accept `?lang=es|en` under the same rules as `REQ-MB-017`: omitted resolves `es`; invalid, empty, or conflicting repeated values resolve `400 INVALID_QUERY_PARAMETER`; URI variant, no `Accept-Language` negotiation.

The localized fields are the frozen content the contract captured at acceptance: the offer `title`/`description`/`rules` and the frozen objective `name`/`description` (contract detail keeps its existing lowercase field names). The contract detail carries the label at top level; each contract-list item carries one `ContentLanguage` label at the contract level that covers its embedded `Offer` — both render the same frozen snapshot in the same language; each review-queue entry carries `content_language`. `Content-Language` header and homogeneous rules follow `REQ-MB-017`. A contract whose frozen snapshot has no English serves `es` declared `es`.

Authorization is unchanged — roster membership, adjudicator role, and member-vs-reviewer projections resolve exactly as before; `lang` changes only which language renders.

#### SCN-MCF-LANG-EN — Member reads contracts in English
Given a contract accepted from a bilingual offer version, when the member reviews their contracts in English, then the embedded offer and frozen objectives render English declared as English with identical ids and merit figures.

#### SCN-MCF-LANG-DETAIL — Contract detail renders the selected language
Given a bilingual contract, when an active member or an adjudicator opens it in English, then the title and each frozen objective's name and description render English — while a member still sees only their own staged credits without media references and an adjudicator sees the full view.

#### SCN-MCF-LANG-QUEUE — Reviewer queue titles localize without loosening access
Given a bilingual contract awaits adjudication, when a moderator reads the review queue in English, then its title renders English — while a caller without the adjudicator role is still refused regardless of language.

#### SCN-MCF-LANG-LEGACY — Spanish-only snapshot falls back declared
Given a contract frozen before English existed, when any authorized reader asks for English, then the frozen Spanish content is served declared as Spanish.

### REQ-MCF-027 — Preview local contract progress without performing fulfillment

A local mid-season preview contains deterministic examples of in-progress and completed solo contracts for existing active members, with frozen offer terms and ledger-shaped credits that reference the matching season, category, objective, unit, contract, and recipient. It does not create real confirmation events, modify identity data, or award official results. Off-season and fresh-start previews have no contracts or credits.

#### SCN-MCF-LOCAL-PROGRESS — Compare member contract and merit screens

Given the reviewer selects the local mid-season dataset, when a participating member reads their contracts and standings, then the existing readers show sample ongoing and completed work with matching merit totals; the reviewer's prior contracts remain untouched in the original dataset.

## Related Capability

- `seasonal-category-competition` owns the seasonal ledger and standings.

## Acceptance Summary

- A contract preserves the accepted `definition_code` and `definition_version` from the offer, plus a frozen `ContractTerms` snapshot of objectives, merits, bonuses, and unit limits.
- A valid confirmation credits exactly the frozen terms' merits once.
- The same unit, objective, contract, and recipient do not generate a second credit.
- The credit belongs to the right season, category, and track.
- Contract state and merit credit persist atomically; a retried confirmation after a partial failure is idempotent.
- Every acceptance mints a distinct attempt contract id; terminal attempts remain queryable.
- A `FREEZING` or `FINALIZED` season rejects new credits except for eligible work started in time.
- Individual standing shows the derived ledger total.
- A team credits the team track, not each member's individual standing.

## HTTP Boundary

- `GET /api/v1/missions/contracts` returns the caller's contracts (filters: `season_id`, `status`; `401` without a valid token, `400` `INVALID_QUERY_PARAMETER` for unknown filters, `200` with the contract list including frozen `Terms` and an embedded `Offer` whose `objectives`/`merit_summary` are derived from those terms).
- `POST /api/v1/missions/contracts/{contract_id}/evidence`
- Request body: `objective_id`, `unit_id`, `season_id`, optional `media_uri`.
- `POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url` — `201 {upload_url, media_uri}`; `400` disallowed type/oversize; `403` outsider; `422` not-in-progress.
- `POST /api/v1/missions/contracts/{contract_id}/evidence/media-url` — `200 {download_url}`; `403` non-reviewer; `422` `EVIDENCE_MEDIA_INVALID`.
- `POST /api/v1/missions/contracts/{contract_id}/start` — roster member transitions `ACCEPTED` to `IN_PROGRESS`.
- `POST /api/v1/missions/contracts/{contract_id}/submit` — roster member transitions `IN_PROGRESS` to `PENDING_VALIDATION`; the active slot is released.
- `POST /api/v1/missions/contracts/{contract_id}/approve` — `admin`/`moderator` transitions `PENDING_VALIDATION` to `COMPLETED`; one atomic write materializes the staged credits plus the frozen `completion_reward` credit (`objective_id` `COMPLETION`) when non-zero and fully confirmed; `200` returns the written credits.
- `POST /api/v1/missions/contracts/{contract_id}/reject` — `admin`/`moderator` transitions `PENDING_VALIDATION` to terminal `DISQUALIFIED` with zero credits.
- `GET /api/v1/missions/review-queue` — `admin`/`moderator` lists `PENDING_VALIDATION` contracts; entries carry the contract summary plus `confirmed_count`/`pending_credits`; `403` non-reviewer, `200` `[]` when empty.
- `POST /api/v1/missions/contracts/{contract_id}/abandon` — transitions a non-terminal contract to `ABANDONED`.
- Requires `Authorization: Bearer <OIDC token>` from HBK Identity, or the HttpOnly `access_token` cookie carrying the same token (the header takes precedence when both are present).
- `201 Created` with the credited `MeritCredit` on evidence success; `200` on lifecycle transitions.
- `400 Bad Request` for invalid JSON.
- `401 Unauthorized` for missing or invalid token.
- `403 Forbidden` for caller without required role.
- `422 Unprocessable Entity` for domain errors with a stable `code` (`CONTRACT_NOT_FOUND`, `CONTRACT_NOT_IN_PROGRESS`, `CONTRACT_NOT_PENDING_VALIDATION`, `CONTRACT_STATE_INVALID`, `CONTRACT_TERMINAL`, `OBJECTIVE_NOT_IN_CONTRACT`, `UNIT_LIMIT_EXCEEDED`, `CONTRACT_EXPIRED`, `SEASON_NOT_ACTIVE`, `SEASON_MISMATCH`, `CATEGORY_MISMATCH`, `ROSTER_EMPTY`, `DUPLICATE_CONFIRMATION`, `INVALID_REQUEST`).
- `503 Service Unavailable` with `MEMBERSHIP_UNAVAILABLE` when the Party Manager membership source cannot answer a teammate check; acceptance always fails closed.

## Stable Production Files

- `internal/missions/domain/confirmation.go`
- `internal/missions/domain/confirmation_id.go`
- `internal/missions/domain/contract_terms.go`
- `internal/missions/domain/merit_credit.go`
- `internal/missions/domain/merit_ledger.go`
- `internal/missions/domain/mission_contract.go`
- `internal/missions/domain/mission_contract_membership.go`
- `internal/missions/domain/mission_contract_projection.go`
- `internal/missions/domain/mission_contract_lifecycle.go`
- `internal/missions/domain/mission_contract_confirm.go`
- `internal/missions/domain/season.go`
- `internal/missions/domain/objective.go`
- `internal/missions/application/confirm_objective.go`
- `internal/missions/application/start_contract.go`
- `internal/missions/application/submit_contract.go`
- `internal/missions/application/review_contract.go`
- `internal/missions/application/list_pending_review.go`
- `internal/missions/domain/mission_contract_review.go`
- `internal/missions/application/abandon_contract.go`
- `internal/missions/interface/http/handler.go`
- `internal/missions/interface/http/errors.go`
- `internal/missions/interface/http/middleware.go`
- `internal/missions/interface/http/contract_lifecycle.go`
- `internal/missions/interface/http/review_queue.go`
- `internal/missions/interface/http/contract_detail.go`
- `internal/missions/interface/http/evidence_upload.go`
- `internal/missions/interface/http/evidence_read.go`
- `internal/missions/domain/media_uri.go`
- `internal/missions/application/evidence_media_ports.go`
- `internal/missions/application/request_evidence_upload.go`
- `internal/missions/application/request_evidence_read.go`
- `internal/missions/infrastructure/merit_ledger_repository.go`
- `internal/missions/infrastructure/dynamodb/*.go`
- `internal/missions/infrastructure/session/*.go`
- `internal/missions/infrastructure/s3/*.go`
- `internal/missions/infrastructure/membership/*.go`
- `internal/missions/composition/*.go`
- `cmd/hbk-missions-service/main.go`
- `cmd/lambda/main.go`

## Test Suites

- `tests/unit/missions/domain`
- `tests/unit/missions/application`
- `tests/unit/missions/interface/http`
- `tests/unit/missions/infrastructure/dynamodb`
- `tests/unit/missions/infrastructure/session`
- `tests/integration/http`
- `tests/integration/infrastructure`
- `tests/integration/infrastructure/dynamodb`
- `tests/integration/infrastructure/session`
- `tests/integration/infrastructure/membership`
- `tests/integration/composition`
- `tests/integration/e2e`

## Implementation Notes

- `MissionContract.Confirm` owns the confirmation policy against the frozen `ContractTerms` and stages the computed `MeritCredit` in the contract's `pendingCredits` map; `ConfirmWithMedia` attaches a validated `MediaURI` on the staged credit.
- Evidence media uses presigned S3 URLs: `EvidenceUploadPresigner` issues contract-prefixed PUT grants (key + content-type bound, ~5min); `EvidenceObjectReader` performs `HeadObject` existence checks and issues reviewer-only GET grants (~15min). The `media_uri` attribute on the `MeritCredit` item is optional and additive — no migration.
- `ConfirmObjectiveHandler` persists the contract with its staged credits via `ContractWriter.Save`; `ApproveContractHandler` commits the adjudicated contract plus every materialized credit via `ContractCreditsWriter.SaveApproved`, and the DynamoDB adapter runs one `TransactWriteItems` whose conditional credit puts map `TransactionCanceledException/ConditionalCheckFailed` to `ErrDuplicateConfirmation`.
- Lifecycle transitions live in `MissionContract` (`Start`, `Submit`, `Approve`, `Reject`, `Abandon`, `Expire`) — `Submit` guarded by roster membership, adjudication guarded by the `admin`/`moderator` role at the application boundary plus the `adminOrModerator` route middleware.
- `ConfirmationID` centralizes the `contract~objective~unit~recipient` identity using a `~` delimiter that cannot collide with colon-containing contract ids.
- Contract ids embed the attempt number (`offer:player:attempt`), so re-acceptance never overwrites prior history.
- `MissionContract.Summary` produces an `ActiveContractSummary` used by the acceptance policy in `mission-board`.
- Adapters persist `MissionContract`, `Season`, `Objective`, and `MeritCredit` in DynamoDB using a single table with `PK`/`SK` and a `GSI1` for bucketed ledger queries.
- `cmd/hbk-missions-service/main.go` wires configuration, DynamoDB client, OIDC session, stores, and HTTP routes.
- `AuthMiddleware` validates OIDC ID tokens and injects `callerID` and claims into the request context.
- Domain errors are mapped to stable public `code` values in the HTTP boundary; internal error strings are never exposed.
