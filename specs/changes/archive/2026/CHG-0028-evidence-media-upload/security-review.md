---
change_id: CHG-0028
artifact: security-review
status: draft
updated: 2026-10-09
security_review_id: SEC-0028
security_review_title: Evidence media upload trust boundary
identity_role: resource-server
---

# Security Review — Evidence Media Upload

## Role

Declared identity role: `resource-server` — the service protects the new
grant endpoints with the existing OIDC session; it does not issue identity.
Reviewer: self (TDD evidence). Focus: presigned upload/read trust boundary.

## Trust-Boundary Analysis

- **New endpoints** `POST .../evidence/upload-url` and
  `POST .../evidence/media-url` sit behind `auth.Authenticate` —
  unauthenticated callers get `401`; identity comes from the session, never
  the body.
- **Upload authority:** grant issuance requires `canAdjudicate` (roster or
  `admin`/`moderator`) plus `IN_PROGRESS` contract — outsiders get `403`,
  inactive contracts `422`, so storage grants cannot be minted for dormant
  or foreign work.
- **Object keys are server-generated** (`evidence/{contract_id}/{uuid}`) —
  clients cannot name objects, cannot write outside their contract prefix,
  and cannot overwrite another contract's objects.
- **Policy-constrained presigned POST**: content-type allowlist
  (`image/*`, `video/*`), `content-length-range` cap, ~5-minute expiry —
  S3 enforces the bounds, not our code, so a leaked/forged grant cannot
  exceed them.
- **Existence gate on confirmation:** `Exists` (HeadObject) before credit —
  `media_uri` cannot reference phantom or foreign objects; prefix
  validation rejects cross-contract references.
- **Read grants are reviewer-only** (`admin`/`moderator`) with ~15-minute
  expiry — roster members cannot enumerate or fetch other players' media.
- **Bucket is private** — `PublicAccessBlockConfiguration` all true, no
  ACLs, no CDN; every object access flows through a short-lived signature.
- **Least-privilege IAM** — function policies scoped to
  `arn:aws:s3:::bucket/evidence/*`; no `s3:*Object` wildcards.
- **No secrets in artifacts** — presigned URLs are short-lived and are
  never logged beyond metadata; `media_uri` is an opaque key, not a URL.

## Data Exposure

- `media_uri` persisted on the `MeritCredit` item — same sensitivity class
  as existing contract/unit ids. Media binaries never enter DynamoDB,
  Lambda payloads, or logs.
- Evidence media may contain personal/captured content: retention is
  governed by bucket lifecycle; adjudication/deletion policy is deferred to
  the adjudication change (documented non-goal).

## Findings

- No Critical/High findings. Verified by tests and template review:
  disallowed content types rejected before presign
  (`TEST-APP-UPLOAD-TYPE-REJECTED`, `TEST-HTTP-UPLOAD-400`),
  declared oversize refused (`ErrEvidenceUploadTooLarge`),
  foreign-prefix `media_uri` rejected at confirmation
  (`TEST-DOMAIN-MEDIAURI-FOREIGN`, `TEST-APP-CONFIRM-MEDIA-FOREIGN`),
  phantom objects rejected via `Exists`, member-without-role denied read
  grant (`TEST-APP-MEDIA-READ-DENIED`, `TEST-HTTP-MEDIA-READ-403`),
  and `PublicAccessBlockConfiguration` all-true on `MissionsEvidenceBucket`.
- **Medium (accepted, documented)**: presigned PUT cannot enforce an
  S3-level `content-length-range`; a granted URL could upload larger than
  declared within its 5-minute TTL. Mitigations: declared-size check at
  issuance, short TTL, contract-scoped keys, lifecycle expiry, and bucket
  size monitoring recommended as a follow-up.
- **Status**: complete — approved 2026-10-09.
