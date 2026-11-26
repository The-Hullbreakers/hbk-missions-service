---
change_id: CHG-0029
title: Harden evidence media upload for browser clients and video size
status: archived
risk: L2
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0029-evidence-media-upload-hardening/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
context_scope: existing
created: 2026-10-09
updated: 2026-10-09
---

# Change — Evidence Media Upload Hardening

## Intent

CHG-0028 shipped the evidence media flow (presigned PUT upload grant,
`media_uri` on confirmations, reviewer read grant). Two gaps block real
browser-based video/image usage: the evidence bucket has no CORS
configuration, so a webapp `PUT` to the presigned URL is rejected by the
browser before reaching S3; and the default `EVIDENCE_MAX_BYTES` (50 MB)
is too small for gameplay video. CI also lacks an S3 emulator, so the
adapter is only exercised through fakes.

## Business Value

- Browser clients can actually upload evidence media — without CORS the
  feature is unusable for the webapp.
- Video-sized uploads fit within the declared bound instead of failing
  validation at grant time.
- The S3 adapter is verified against a real emulator in CI, closing the
  "verify on deploy" residual from CHG-0028.

## User Stories

- **US-0029-01 — Browser upload works:** As a roster member using the
  webapp, my browser `PUT` to the presigned upload URL succeeds so I can
  attach video or image evidence.
- **US-0029-02 — Video evidence fits:** As a player, a typical gameplay
  clip fits within the configured size bound so I do not get
  `INVALID_REQUEST` for normal evidence.
- **US-0029-03 — Adapter verified:** As an operator, CI exercises the S3
  adapter against a faithful emulator so presign/exists/read behavior is
  proven without a deployed environment.

## Risk Classification

**L2** — modifies an existing trust boundary (bucket CORS widens who can
issue browser requests against it) and test infrastructure; no new public
endpoint and no domain/application logic change. Security gate mandatory
(CORS touches exposure). Mutation gate not applicable — no non-trivial
production logic. Critical E2E unaffected (journey shape unchanged).

## Approach

- `template.yaml`: add `CorsConfiguration` to the evidence bucket —
  `PUT`, `HEAD`, `GET` methods; allowed headers for content-type/
  content-length signing; origins parameterized per environment
  (`EvidenceCorsOrigins`, localhost default for dev).
- Raise the upload bound to a video-appropriate default (256 MB) via a
  `EvidenceMaxBytes` parameter feeding `EVIDENCE_MAX_BYTES`.
- `docker-compose.test.yml`: add a `localstack` service (S3 only).
- Infrastructure contract test: `S3_ENDPOINT_URL`-driven round-trip —
  presign PUT → real `PUT` object → `Exists` true → presign GET →
  download matches bytes; plus `Exists` false for a phantom key.

## Behavior Delta Summary

- ADDED: bucket CORS configuration; S3 emulator in test compose;
  infrastructure contract test for the evidence store.
- MODIFIED: `REQ-MCF-014` size bound becomes a configurable
  deployment parameter with a 256 MB default (was fixed 50 MB).
- REMOVED: none.

## Non-Goals

- No multipart upload support — single presigned PUT already handles
  objects up to 5 GB; revisit only if client reliability demands it.
- No post-upload content validation (malware/type sniffing) — deferred
  to the adjudication change, same as CHG-0028.
- No CloudFront/CDN or public delivery.

## Decisions

- **CORS origins parameterized, never `*`.** The evidence bucket is
  private; a wildcard origin would let any site drive uploads through a
  stolen grant. Origins come from a template parameter per environment.
- **Single-part PUT stays.** SDK v2 lacks POST-policy presign; PUT binds
  key + content-type, which CORS only needs to allow. Size remains
  declared-at-grant (documented residual, unchanged).
- **localstack, not moto/minio.** Consistent with `dynamodb-local`
  compose pattern; S3-only service keeps the image light.

## Sequencing

Independent; touches only `template.yaml`, `docker-compose.test.yml`,
and a new test file under `tests/integration/infrastructure/s3/`.
No overlap with CHG-0014/0017/0023 active deltas.

## Approvals

- Gate 1: approved 2026-10-09.
