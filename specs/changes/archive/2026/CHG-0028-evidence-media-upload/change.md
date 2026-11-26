---
change_id: CHG-0028
title: Evidence media upload via object storage
status: archived
risk: L3
canonical_action: modify
change_action: create_delta
canonical_paths:
  - specs/capabilities/mission-contract-fulfillment/spec.md
change_path: specs/changes/archive/2026/CHG-0028-evidence-media-upload/
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_REST_API_DESIGN_WORKFLOW
  - WORKFLOW-COMMON_AWS_LAMBDA_REST_WORKFLOW
  - WORKFLOW-COMMON_SDD_MUTATION_GATE_WORKFLOW
  - WORKFLOW-COMMON_SDD_CRITICAL_E2E_WORKFLOW
  - WORKFLOW-COMMON_SDD_SECURITY_GATE_WORKFLOW
context_scope: existing
created: 2026-10-09
updated: 2026-10-09
---

# Change — Evidence Media Upload

## Intent

Today evidence is a bare JSON confirmation (`objective_id`, `unit_id`,
`season_id`); there is no way for a player to attach the video or image that
proves the unit happened. This change adds object-storage-backed media to
evidence: the client requests a short-lived upload grant, sends the binary
straight to storage, and confirms the unit referencing the stored object.
Reviewers (`admin`/`moderator`) get a read grant to inspect the media.

## Business Value

- A player can prove a bounty/objective with a video or image instead of an
  unverifiable unit id — the roadmap's "external multimedia evidence".
- Storage cost stays near zero for the service: binaries never pass through
  Lambda or the database; only opaque references land in DynamoDB.
- Reviewers can open submitted media without the bucket ever becoming
  public — every access is short-lived and authorized.

## User Stories

- **US-0028-01 — Upload evidence media:** As a roster member of an
  `IN_PROGRESS` contract, I request an upload grant so I can send my video
  or image straight to storage.
- **US-0028-02 — Confirm with media:** As a roster member, I submit
  evidence referencing the uploaded object so my confirmation carries the
  proof with it.
- **US-0028-03 — Review media:** As a moderator, I obtain a short-lived
  read grant for a confirmation's media so I can adjudicate without public
  access to the bucket.

## Risk Classification

**L3** — introduces a new public trust boundary (short-lived write grants to
object storage), extends authorization to upload/read paths, stores
user-supplied binary metadata, and changes the evidence request contract.
Security gate, mutation gate, and critical E2E are mandatory.

## Approach

- `POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url`
  returns a presigned POST (policy-constrained) with a server-generated key
  `evidence/{contract_id}/{uuid}`, a content-type allowlist
  (`image/*`, `video/*`), `content-length-range` cap, and ~5-minute expiry.
  Authorization mirrors `REQ-MCF-007` (roster or `admin`/`moderator`) and the
  contract must be `IN_PROGRESS`.
- `POST /api/v1/missions/contracts/{contract_id}/evidence` accepts an
  optional `media_uri`. The reference must use the
  `evidence/{contract_id}/` prefix and the object must exist in storage;
  it is persisted on the resulting `MeritCredit` (`media_uri`, optional).
- `POST /api/v1/missions/contracts/{contract_id}/evidence/media-url`
  (reviewer path) returns a presigned GET for a stored `media_uri`;
  restricted to `admin`/`moderator`.
- Domain stays provider-neutral: `MediaURI` is a validated string reference;
  the storage capability lives behind an `EvidenceBlobStore` port
  (`PresignUpload`, `PresignRead`, `Exists`) implemented by an S3 adapter.
- Bucket provisioned in `template.yaml`: private, blocked public access,
  lifecycle expiry for staging/quarantine objects; least-privilege IAM for
  the function.

## Behavior Delta Summary

- ADDED: media upload grant, `media_uri` on evidence confirmation and on
  the recorded credit, reviewer read grant.
- MODIFIED: `REQ-MCF-003` accepts an optional validated media reference;
  `REQ-MCF-007` authority extends to the upload/read grant endpoints.
- REMOVED: none — evidence without `media_uri` keeps working unchanged.

## Non-Goals

- No adjudication workflow: media does not gate credit; rejection/revocation
  of evidence remains a separate governed change.
- No media processing (thumbnails, transcoding, scanning) — raw storage only.
- No client-supplied object keys; the server always generates them.
- No public or CDN delivery of evidence media.

## Decisions

- **Presigned POST, not PUT.** POST policies enforce
  `content-length-range`, so oversized uploads are rejected by S3 itself;
  a PUT signature cannot bound size.
- **Server-generated keys.** The key is `evidence/{contract_id}/{uuid}`;
  the client never names objects, which prevents cross-contract writes.
- **Existence checked at confirmation.** `EvidenceBlobStore.Exists`
  (HeadObject) runs before crediting so `media_uri` cannot point at a
  phantom object.
- **Credit stays immediate.** `media_uri` is stored metadata on the credit;
  a later adjudication change may revoke it.
- **Private bucket only.** All access flows through short-lived presigned
  URLs; no public ACL, no CDN.

## Sequencing

Independent of CHG-0023/0024/0027; touches only the evidence path in
`mission-contract-fulfillment`. If CHG-0027 lands first, departed members
are already excluded by `CanSubmitEvidence` and the new endpoints inherit
the exclusion for free.

## Approvals

- Gate 1: approved 2026-10-09.
