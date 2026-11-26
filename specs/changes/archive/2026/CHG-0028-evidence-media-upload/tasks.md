---
change_id: CHG-0028
artifact: tasks
status: implemented
updated: 2026-10-09
---

# Tasks — Evidence Media Upload

Execution mode: sequential, one agent. Each wave opens only after the
previous layer's RED evidence is reviewed (scoped Gate 3).

## Wave 1 — Domain

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-01 | Add `ErrEvidenceMediaInvalid` sentinel | domain | — | `go test ./tests/unit/missions/domain/` |
| T-02 | Add `MediaURI` VO + `NewMediaURI` validation + `MissionContract.MediaPrefix()` | domain | T-01 | same suite |
| T-03 | `MeritCredit.MediaURI` field + `ConfirmWithMedia` attaches validated reference | domain | T-02 | same suite — done |

RED evidence: `TestMediaURI_*` — valid prefix accepted; foreign contract,
traversal, absolute URL, empty rejected. `TestConfirm_MediaURI` — credit
carries the reference.

## Wave 2 — Application

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-04 | `EvidenceUploadPresigner`/`EvidenceObjectReader` ports + grant DTOs | application | T-02 | `go test ./tests/unit/missions/application/` |
| T-05 | `RequestEvidenceUploadHandler` — auth, `IN_PROGRESS`, content-type allowlist | application | T-04 | same suite |
| T-06 | `ConfirmObjectiveHandler` `WithEvidence` variant + `HandleWithMedia` + `Exists` check | application | T-04 | same suite — done |
| T-07 | `RequestEvidenceReadHandler` — `admin`/`moderator` only | application | T-04 | same suite |

RED evidence: `TestRequestEvidenceUploadHandler_*` — grant issued to
member; `ACCEPTED`/terminal rejected; outsider 403-equivalent domain error;
disallowed content type rejected. `TestConfirmObjective_Media*` — phantom
and foreign `media_uri` rejected, no credit. `TestRequestEvidenceReadHandler_*`
— member denied, moderator granted.

## Wave 3 — Boundary + Infrastructure

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-08 | HTTP `evidence/upload-url` + `evidence/media-url` handlers + `media_uri` on evidence request + error mapping | interface | T-05, T-07 | `go test ./tests/unit/missions/interface/http/` + `tests/integration/http/` |
| T-09 | S3 adapter `evidence_store.go` (presigned PUT/GET + HeadObject) | infrastructure | T-04 | done — real-S3 verify on deploy |
| T-10 | `MeritCredit` item persists optional `media_uri` | infrastructure | T-03 | `go test ./tests/integration/infrastructure/dynamodb/` |

## Wave 4 — Composition + E2E + IaC

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-11 | Wire handlers + routes in `server.go`; bucket resource, IAM, env vars in `template.yaml` | composition | T-08, T-09 | `go test ./tests/integration/composition/` |
| T-12 | Critical E2E: grant → upload (localstack) → confirm with media → reviewer read grant | e2e | T-11 | `tests/integration/e2e/` |

## Wave 5 — Gates

| ID | Task |
| --- | --- |
| T-13 | BLUE — file sizes < 150 lines, gofmt/vet clean |
| T-14 | Security gate — trust-boundary evidence in `security-review.md` |
| T-15 | Mutation gate on `NewMediaURI`/`Confirm`/handler authorization |
| T-16 | Coverage gate — production ≥90% |
| T-17 | Docs — `API-CONTRACT.md`, `API-GUIDE-FRONTEND.md`, `DATA-MODEL.md`, `AUTHENTICATION.md` |
| T-18 | `tools/validate-sdd-change.sh` + fill `verification.md` |
| T-19 | Consolidate canonical spec, index, archive |
