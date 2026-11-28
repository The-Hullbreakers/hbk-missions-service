---
change_id: CHG-0063
artifact: design
status: implemented
updated: 2026-11-27
---

# Design — Required mission evidence

## Domain

- `internal/missions/domain/errors.go`: add
  `ErrContractEvidenceRequired = errors.New("mission evidence is required")`.
- `internal/missions/domain/mission_contract_review.go`: in `Submit`, after
  the existing membership → terminal → state checks and immediately before
  `c.State = ContractPendingValidation`, add
  `if !c.hasStagedEvidence() { return ErrContractEvidenceRequired }`.
  Helper on `*MissionContract` scans `pendingCredits` for a non-empty
  `MediaURI`. Comment updated to state partial submission requires
  evidence. This preserves reviewability of contracts already pending —
  including ones staged by legacy media-less credits — because the guard
  only executes on the submit transition.

## Application

- `internal/missions/application/confirm_objective.go`: remove `Handle`
  and `NewConfirmObjectiveHandler`; `NewConfirmObjectiveHandlerWithEvidence`
  initializes the struct directly. In `HandleWithMedia`, after
  `canAdjudicate` and before `domain.NewMediaURI`, add
  `if mediaURI == "" { return domain.MeritCredit{}, domain.ErrContractEvidenceRequired }`.
  Foreign-prefix, missing-object and Exists-error handling is unchanged.

## Interface

- `internal/missions/interface/http/handler.go`: replace the
  media-present if/else with the single
  `credit, err := app.HandleWithMedia(identity, contractID, req.ObjectiveID, req.UnitID, req.MediaURI, req.SeasonID)`.
- `internal/missions/interface/http/errors.go`: add
  `errors.Is(err, domain.ErrContractEvidenceRequired) → "EVIDENCE_REQUIRED"`.

## Test strategy

- TEST-0063-APP-001: `HandleWithMedia(member, "contract-1", "objective-1",
  "unit-1", "", "season-1")` → `ErrContractEvidenceRequired`, writer not
  called. Existing media/foreign/phantom/outsider assertions preserved.
- TEST-0063-APP-002: submit on empty `IN_PROGRESS` → error, no `Save`,
  state unchanged.
- TEST-0063-DOM-001: `Submit` rejects empty and legacy-only staged credit
  (both unchanged); one real staged `MediaURI` with objectives remaining
  succeeds.
- TEST-0063-HTTP-001: POST evidence with omitted and empty `media_uri` →
  422 `EVIDENCE_REQUIRED`, no persistence; stored same-contract reference
  still 201; outsider still 403.
- TEST-0063-HTTP-002: submit route on evidence-empty contract → 422
  `EVIDENCE_REQUIRED`, unchanged; existing partial-success case re-backed
  by a media credit.
- Composition E2E: evidence journey extended with pre-upload rejection and
  submit/approve review; other journeys seeded through the upload grant
  flow (`blobStoreFake`).
- Fixture migration: application/HTTP/domain success fixtures stage media
  via `ConfirmWithMedia` / `HandleWithMedia` with a seeded
  `fakeEvidenceStore`/`blobStoreFake`; pending-only fixtures may use
  `NewMissionContractFromState` with legacy credits where the scenario is
  historical.

## Security / cost

- No IAM, routing, permission or storage changes; auth order preserved
  (403 before evidence validation). No migrations; no deploy or commit.
