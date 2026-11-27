---
change_id: CHG-0063
title: Require stored media evidence for confirmations and review submission
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_path: specs/capabilities/mission-contract-fulfillment/spec.md
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
archive_path: specs/changes/archive/2026/CHG-0063-required-mission-evidence/
verified_at: 2026-11-27
source_revision: reviewed working diff atop 43bb599 (not a change commit)
context_scope: none
context_scope_reason: no durable cross-session context change; canonical spec
  plus this archive are the record
execution_mode: sequential
---

Validation outcome: consolidated into
`specs/capabilities/mission-contract-fulfillment/spec.md` after explicit human
approval (`listo todo se ve bien puedes archivar y cerrar el cambio`) following
disclosure of the gateway limitation. Scope acceptance: unit, HTTP-integration
and composition suites green; behavioral RED verified via post-implementation
`go test -overlay` replay (recorded as replay, not a pre-implementation run).
Known limitations approved at closure: authenticated end-to-end through the
local gateway was not executed (gateway returns 401 without credentials; user's
preview stack untouched); `TestSeasonScenarioReaders` remains skipped because
it writes to the user's preview table; a real deployment needs a configured S3
evidence bucket (the AWS template already binds it). Independent review
disposition: no new security finding.

# Required mission evidence

Approved by the user's literal `si` on the cross-repo plan; this artifact set
covers the backend wave only (hbk-missions-service).

Each new HTTP evidence confirmation must reference an already stored
same-contract image/video object through the existing upload grant flow. A
contract may only be submitted for review when at least one staged credit
carries a media reference; partial progress still submits. Historic data,
draft credits and existing pending contracts are not rewritten and remain
reviewable. No permission, routing, storage or migration changes.

## Acceptance scenarios

- `POST /api/v1/missions/contracts/{contract_id}/evidence` with a missing or
  empty `media_uri` answers 422 `EVIDENCE_REQUIRED` and stages nothing.
- The same request with a stored, same-contract media reference stages the
  credit as before; foreign or phantom references keep `EVIDENCE_MEDIA_INVALID`.
- `POST /api/v1/missions/contracts/{contract_id}/submit` on an in-progress
  contract with no media-backed staged credit answers 422
  `EVIDENCE_REQUIRED` and persists no state change; a contract with at least
  one media-backed credit still submits, including partially fulfilled terms.
- Authorization precedes the evidence requirement: outsiders still get 403.

## Layer gates / verification

- TEST-0063-APP-001/002 — application: empty media rejected without save;
  submit on an evidence-empty in-progress contract rejected without save.
- TEST-0063-DOM-001 — domain: `Submit` rejects empty and legacy-only staged
  credits unchanged; succeeds with one media-backed credit under partial
  objectives.
- TEST-0063-HTTP-001/002 — boundary: omitted/empty media and bare submit
  return 422 `EVIDENCE_REQUIRED`; stored-reference confirm and
  evidence-backed submit stay green; outsider still 403.
- Composition E2E `TestEvidenceMediaJourney_EndToEnd` extended with
  pre-upload rejection and post-confirmation review.
