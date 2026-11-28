---
change_id: CHG-0063
artifact: delta
status: implemented
updated: 2026-11-27
---

# Delta — Required mission evidence

## What Is Changing

### Domain

- `ErrContractEvidenceRequired` sentinel added to `domain/errors.go`.
- `MissionContract.Submit` keeps membership/terminal/state precedence and,
  immediately before assigning `ContractPendingValidation`, requires
  `hasStagedEvidence()` — at least one pending credit with a non-empty
  `MediaURI`. Contracts already in `PENDING_VALIDATION` (historic or staged
  by legacy credits) remain reviewable; the guard only gates new submits.
- `MissionContract.Confirm` (media-less) is retained for non-public and
  historical-fixture semantics.

### Application

- `ConfirmObjectiveHandler.HandleWithMedia` rejects an empty `mediaURI` with
  `ErrContractEvidenceRequired` after authorization and before media
  validation. The media-less `Handle` and standalone
  `NewConfirmObjectiveHandler` are removed; the public use case is
  `NewConfirmObjectiveHandlerWithEvidence` + `HandleWithMedia` only.

### Interface (HTTP)

- `ConfirmEvidenceHandler` always calls `HandleWithMedia`; the
  media-present if/else is gone.
- `PublicCode` maps `ErrContractEvidenceRequired` to `EVIDENCE_REQUIRED`
  (422 default; auth mapping unchanged).

## MODIFIED

`canonical_action: modify` against
`specs/capabilities/mission-contract-fulfillment/spec.md`:

- **REQ-MCF-003** (Confirm countable units) — `media_uri` is no longer
  optional: every HTTP evidence confirmation must reference a stored,
  same-contract media object; the media-less confirm path is retired.
- **REQ-MCF-015** (Attach media to an evidence confirmation) — the attach
  contract becomes the only public confirm variant; omission of
  `media_uri` resolves as `EVIDENCE_REQUIRED` rather than a media-less
  credit.
- **REQ-MCF-019** (Submit a contract for review) — submit additionally
  requires at least one staged credit carrying a media reference;
  contracts already in `PENDING_VALIDATION` and legacy staged credits
  stay reviewable.
- HTTP contract lines — `POST .../evidence` gains a mandatory
  `media_uri` field and a 422 `EVIDENCE_REQUIRED` response;
  `POST .../submit` gains the 422 `EVIDENCE_REQUIRED` response mode.

## BDD Scenarios

### SCN-0063-EVIDENCE — Media required to stage evidence

Given an in-progress contract and a roster member, when the member confirms
an objective without a media reference, then the use case answers
`ErrContractEvidenceRequired` and nothing is persisted.

### SCN-0063-SUBMIT — Evidence required to enter review

Given an in-progress contract with no media-backed staged credit, when the
member submits for review, then `ErrContractEvidenceRequired` is returned
and the contract stays `IN_PROGRESS`. Given at least one staged credit with
a media reference, submit transitions to `PENDING_VALIDATION` even with
unconfirmed objectives remaining.

## HTTP Boundary Changes

- `POST /api/v1/missions/contracts/{contract_id}/evidence` — `media_uri`
  required; omission is no longer a media-less confirm.
- `POST /api/v1/missions/contracts/{contract_id}/submit` — new
  `EVIDENCE_REQUIRED` rejection mode.

## Error Vocabulary

- `EVIDENCE_REQUIRED` (422) — new public code for
  `ErrContractEvidenceRequired`.
