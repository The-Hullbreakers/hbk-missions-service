---
change_id: CHG-0028
artifact: delta
status: proposed
updated: 2026-10-09
---

# Delta — Evidence Media Upload

## What Is Changing

### Mission Contract Fulfillment

- A roster member of an `IN_PROGRESS` contract can request a short-lived,
  constrained upload grant and attach the resulting object reference to an
  evidence confirmation (`media_uri`). The uploaded object is stored in
  private object storage; only its reference lands on the merit credit.
- Reviewers (`admin`/`moderator`) can request a short-lived read grant for
  a confirmation's media; the bucket never becomes public.

## ADDED

### REQ-MCF-014 — Request an evidence media upload grant

`POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url` accepts
`{"content_type": "...", "content_length": N}` and returns `201` with
`{"upload_url", "fields", "media_uri", "expires_at"}`. The grant is a
presigned POST constrained by policy: key `evidence/{contract_id}/{uuid}`
(server-generated), content-type allowlist `image/*`/`video/*`,
`content-length-range` bounded by the configured maximum, expiry ~5 minutes.
Authority matches `REQ-MCF-007`; a non-`IN_PROGRESS` contract resolves as
`422 CONTRACT_NOT_IN_PROGRESS`; an invalid or disallowed `content_type`
resolves as `400 INVALID_REQUEST`.

#### SCN-MCF-UPLOAD-GRANT — Member receives a constrained grant

Given `player-1` is a roster member of an `IN_PROGRESS` contract, when they
request an upload grant with `content_type: image/png`, then the response is
`201` with an `upload_url`, POST `fields` encoding the size/type/key policy,
a `media_uri` under `evidence/{contract_id}/`, and an `expires_at` ~5 min.

#### SCN-MCF-UPLOAD-NOT-IN-PROGRESS — No grant on inactive contract

Given the contract is `ACCEPTED` or terminal, when a member requests an
upload grant, then the response is `422` with `CONTRACT_NOT_IN_PROGRESS`
and no grant is issued.

#### SCN-MCF-UPLOAD-OUTSIDER — Outsider cannot request a grant

Given `player-2` is not in the roster and holds no privileged role, when
they request an upload grant, then the response is `403 FORBIDDEN` and no
grant is issued.

#### SCN-MCF-UPLOAD-TYPE-REJECTED — Disallowed media type

Given a member requests an upload grant with `content_type:
application/x-msdownload`, when it is validated against the allowlist, then
the response is `400` with `INVALID_REQUEST` and no grant is issued.

### REQ-MCF-015 — Attach media to an evidence confirmation

`POST /api/v1/missions/contracts/{contract_id}/evidence` accepts an
optional `media_uri`. When present, it must reference this contract's
prefix (`evidence/{contract_id}/`) and the object must exist in storage;
otherwise the response is `422` with `EVIDENCE_MEDIA_INVALID`. On success
the `media_uri` is stored on the recorded `MeritCredit` and echoed in the
`201` response. Evidence without `media_uri` is unchanged.

#### SCN-MCF-EVIDENCE-MEDIA — Confirmation carries the media reference

Given `player-1` uploaded an object under `evidence/contract-1/{uuid}`, when
they confirm `unit_id` with that `media_uri`, then the credit records
`media_uri` and the `201` response echoes it.

#### SCN-MCF-EVIDENCE-MEDIA-PHANTOM — Phantom object rejected

Given no object exists at `media_uri`, when the confirmation is submitted,
then it is rejected with `EVIDENCE_MEDIA_INVALID` and no credit is recorded.

#### SCN-MCF-EVIDENCE-MEDIA-FOREIGN — Foreign prefix rejected

Given `media_uri` points under another contract's prefix, when the
confirmation is submitted, then it is rejected with `EVIDENCE_MEDIA_INVALID`.

### REQ-MCF-016 — Read grant for stored evidence media

`POST /api/v1/missions/contracts/{contract_id}/evidence/media-url` accepts
`{"media_uri": "..."}` and returns `200` with `{"media_url", "expires_at"}` —
a presigned GET with ~15-minute expiry — when the caller holds `admin` or
`moderator` and the `media_uri` belongs to this contract. A roster member
without a privileged role receives `403 FORBIDDEN`; an unknown or foreign
`media_uri` resolves `422` with `EVIDENCE_MEDIA_INVALID`.

#### SCN-MCF-MEDIA-READ-REVIEWER — Moderator opens the media

Given a credit recorded `media_uri` for `contract-1`, when a moderator
requests a media read grant for it, then the response is `200` with a
short-lived `media_url`.

#### SCN-MCF-MEDIA-READ-DENIED — Member cannot read reviewer media

Given a roster member without a privileged role requests a media read
grant, when it is evaluated, then the response is `403 FORBIDDEN`.

## MODIFIED

### REQ-MCF-003 — Confirm countable units

A unit is valid only when the contract was active, the objective was
eligible, it occurred inside the window, the recipient was entitled, the
evidence or source was accepted, and the supplied season matches the
contract's bound season. **When `media_uri` is supplied, it must reference
the contract's `evidence/{contract_id}/` prefix and resolve to an existing
stored object; the reference is persisted on the resulting `MeritCredit`.**

### REQ-MCF-007 — Confirm evidence only within caller authority

`POST /api/v1/missions/contracts/{contract_id}/evidence` resolves player
from the session. Submission — **and evidence media upload grants —** is
authorized only when the session player is in the contract roster or the
caller holds `admin`/`moderator`. **Media read grants additionally require
the `admin`/`moderator` role regardless of roster membership.**

## HTTP Boundary Changes

- `POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url` —
  new. `201` `{upload_url, fields, media_uri, expires_at}`; `401`
  unauthenticated; `403` outsider; `422` `CONTRACT_NOT_IN_PROGRESS`;
  `400` `INVALID_REQUEST` (bad JSON or disallowed `content_type`).
- `POST /api/v1/missions/contracts/{contract_id}/evidence` — request gains
  optional `media_uri`; `422` gains `EVIDENCE_MEDIA_INVALID`.
- `POST /api/v1/missions/contracts/{contract_id}/evidence/media-url` —
  new. `200` `{media_url, expires_at}`; `401`; `403`; `422`.

## Persistence Changes

- `MeritCredit` gains optional `media_uri` (string attribute on the credit
  item; absent for media-less confirmations). No migration needed.
- New private bucket resource in `template.yaml` with public access fully
  blocked, lifecycle rule for object expiry, and least-privilege function
  IAM (PutObject via presign is performed by the client; the function needs
  `s3:PutObject`/`GetObject` signing rights and `HeadObject`).
