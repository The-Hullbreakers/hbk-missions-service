---
change_id: CHG-0028
artifact: design
status: implemented
updated: 2026-10-09
---

# Design — Evidence Media Upload

## Domain

New value object in `internal/missions/domain/` — provider-neutral string
reference, no AWS vocabulary:

```go
// MediaURI is a validated reference to a stored evidence object.
// Valid form: evidence/{contract_id}/{object_id}; never client-shaped URLs.
type MediaURI string

// NewMediaURI validates the reference against the owning contract.
func NewMediaURI(raw, contractID string) (MediaURI, error)
```

- `NewMediaURI` rejects empty, non-`evidence/{contractID}/`-prefixed,
  traversal (`..`), absolute-URL, or otherwise malformed references with
  `ErrEvidenceMediaInvalid` (new sentinel in `errors.go`).
- `MissionContract.MediaPrefix()` returns `evidence/{contract_id}/` so
  application code never rebuilds the convention.
- `MeritCredit` gains `MediaURI MediaURI` (zero value = no media; the
  DynamoDB item omits the attribute when empty).
- As-built: `Confirm` keeps its signature; a new `ConfirmWithMedia`
  accepts the validated `MediaURI` so existing callers stay untouched —
  media validation is input-shape only; existence is an application
  concern via the port.
- `AllowedMediaType` domain policy allows `image/*` and `video/*`;
  violations return `ErrMediaTypeNotAllowed` (`400 INVALID_REQUEST`); a
  declared size above the bound returns `ErrEvidenceUploadTooLarge`.

## Application

New port (one operation per consumer, CQRS-shaped; defined beside its use
case, not a grouped repository):

```go
// EvidenceUploadPresigner grants a constrained upload to evidence storage.
type EvidenceUploadPresigner interface {
	PresignUpload(key string, contentType string, maxBytes int64) (UploadGrant, error)
}

// EvidenceObjectReader verifies and grants reads of stored evidence.
type EvidenceObjectReader interface {
	Exists(mediaURI domain.MediaURI) (bool, error)
	PresignRead(mediaURI domain.MediaURI) (ReadGrant, error)
}
```

`UploadGrant`/`ReadGrant` are application-owned structs carrying `URL`
(`MediaURI` on the upload grant) — no SDK types leak upward. A combined
`EvidenceBlobStore` interface exists solely for composition injection.

- `RequestEvidenceUploadHandler` — `Get` contract → authorize
  (`canAdjudicate` reuse) → require `ContractInProgress` → validate
  `content_type` via `AllowedMediaType` → declared-size check → build key
  via `contract.MediaPrefix() + rand.Text()` → `PresignUpload`.
- `ConfirmObjectiveHandler` gains a variant constructor
  (`NewConfirmObjectiveHandlerWithEvidence`) and `HandleWithMedia`:
  validate `NewMediaURI`, `Exists` via `EvidenceObjectReader`, then
  `ConfirmWithMedia` and atomic `SaveConfirmed` unchanged.
- `RequestEvidenceReadHandler` — `Get` contract → require
  `admin`/`moderator` role → `NewMediaURI` against this contract →
  `Exists` → `PresignRead`.

## HTTP Boundary

`interface/http/` additions (files stay <150 lines):

- `evidence_upload.go`: `POST .../evidence/upload-url` — request
  `{content_type, content_length}`; `201` response `{upload_url,
  media_uri}`; `403` on `ErrCallerNotAuthorized`, `422` on
  contract-state errors, `400` on decode/`ErrMediaTypeNotAllowed`.
- `handler.go`: `evidenceRequest` gains `media_uri` (optional string) and
  dispatches to `HandleWithMedia`; `EVIDENCE_MEDIA_INVALID` added to the
  stable error-code map (`ErrMediaTypeNotAllowed`/`ErrEvidenceUploadTooLarge`
  fall through to `INVALID_REQUEST`).
- `evidence_read.go`: `POST .../evidence/media-url` — request
  `{media_uri}`; `200` response `{download_url}`; `403` for callers
  without `admin`/`moderator`, `422` on invalid/phantom reference.

## Infrastructure

`internal/missions/infrastructure/s3/`:

- `evidence_store.go` — S3 adapter implementing `EvidenceBlobStore` via
  `s3.PresignClient` (`PresignPutObject` binding key + Content-Type,
  5min; `HeadObject` for `Exists`; `PresignGetObject`, 15min). As-built:
  aws-sdk-go-v2 has no POST-policy presign, so no S3-level
  `content-length-range`; declared size is validated at grant issuance.
- `s3_client.go` — client factory mirroring the DynamoDB pattern
  (`S3_ENDPOINT_URL` + static test creds → local emulator; else
  `LoadDefaultConfig`).

`template.yaml`:

- `AWS::S3::Bucket` private (`PublicAccessBlockConfiguration` all true),
  `BucketEncryption` SSE-S3, versioning, `LifecycleConfiguration` (abort
  multipart 7d, expire `evidence/` after 365d), `DeletionPolicy: Retain`.
- Function `Policies` gain least-privilege `s3:PutObject`, `s3:GetObject`,
  `s3:DeleteObject` scoped to `arn:aws:s3:::bucket/evidence/*`.
- Env vars: `EVIDENCE_BUCKET_NAME`, `EVIDENCE_MAX_BYTES`
  (`S3_ENDPOINT_URL` for local emulators, read in `cmd/*`).

## Persistence

- `MeritCredit` item adds optional `media_uri` attribute; absent = no media.
  Ledger readers map it back; no migration (attribute-additive).

## Composition

`server.go` wires `RequestEvidenceUploadHandler`,
`ConfirmObjectiveHandlerWithEvidence`, `RequestEvidenceReadHandler` when
`Config.Evidence` is set; both new routes mount behind
`auth.Authenticate`; reviewer-only access is enforced inside the
application handler (`isReviewer`), keeping role semantics with the use
case.

## Test Strategy

- Domain unit: `NewMediaURI` validation, `MediaPrefix`, `Confirm` stores
  media reference on the credit.
- Application unit: hand-written port doubles for presigner/reader —
  authorization, state, content-type, phantom/foreign `media_uri` paths.
- HTTP integration: full request paths for the three endpoints incl. error
  vocabulary.
- Infrastructure: DynamoDB item test for `MediaURI` persistence (real
  dynamodb-local); **gap:** no localstack S3 in CI — presign fields are
  pure SDK output; real-S3 verification is on-deploy (rollout.md).
- E2E (critical journey): upload grant → confirm with `media_uri` →
  reviewer read grant, through the real HTTP boundary and composition
  (`tests/integration/composition/evidence_media_e2e_test.go`).

## Rollback

Additive endpoints + optional field: revert deploy; bucket and `media_uri`
data remain dormant and re-readable by a later redeploy. No flag needed.
