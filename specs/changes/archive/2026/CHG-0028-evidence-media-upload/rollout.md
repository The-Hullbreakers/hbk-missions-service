---
change_id: CHG-0028
artifact: rollout
status: draft
updated: 2026-10-09
---

# Rollout — Evidence Media Upload

## Deployment

- Single service, additive change: two new endpoints + optional request
  field + new bucket resource.
- `template.yaml` creates the private evidence bucket and wires env vars;
  `sam deploy` creates it before/with the function update — no ordering
  risk since nothing reads it until the endpoints are called.
- Existing evidence confirmations without `media_uri` are unaffected; the
  attribute is absent on existing credit items and reads as empty.

## Backward Compatibility

- API is additive: new endpoints, one optional field on the evidence
  request. Clients ignoring media continue to work unchanged.
- New error code `EVIDENCE_MEDIA_INVALID` only surfaces when a client sends
  `media_uri` — clients that never send it never see it.
- Bucket lifecycle policy governs retention from creation; no migration.

## Rollback

- Revert the deployment. The bucket and any stored objects remain but are
  unreachable without the endpoints — dormant, not destructive.
- `media_uri` attributes on credit items are ignored by reverted readers;
  a later redeploy interprets them correctly again.
- Cost of keeping the bucket is near zero at expected volume; deleting it
  is a separate operator action if rollback is permanent.

## Verification On Deploy

- `POST .../evidence/upload-url` on an `IN_PROGRESS` contract returns `201`
  with policy fields; a real PUT/POST of an `image/*` object under the
  returned key succeeds in S3.
- `POST .../evidence` with the returned `media_uri` returns `201` and the
  credit echoes `media_uri`.
- `POST .../evidence/media-url` as `admin`/`moderator` returns a working
  presigned GET; as a member returns `403`.
