---
change_id: CHG-0029
artifact: security-review
status: complete
updated: 2026-10-09
---

# Security Review — Evidence Media Upload Hardening

## Scope

CORS configuration added to the private evidence bucket; upload bound
moved to a deployment parameter; S3 emulator added to test compose.
Role: security reviewer for the bucket trust boundary widened by CHG-0028.

## Findings

### CORS origin allowlist (was: not configured)

- The bucket now exposes `CorsConfiguration` with `PUT`/`GET`/`HEAD` for
  `EvidenceCorsOrigins` — a `CommaDelimitedList` parameter, default
  `https://app-stg.thehullbreakers.com` + `http://localhost:3000`.
- **Not** `*`: a wildcard origin would let any website drive uploads
  through a leaked/stolen grant in a victim's browser. Explicit origins
  keep CORS enforcement meaningful.
- CORS is a browser control, not authorization — it gates *who may read
  the response in a browser*, and preflights *which origins may send*
  the request. A non-browser client ignores it. Grant authorization
  remains the roster/reviewer check upstream.
- `AllowedHeaders: ["*"]` is required because presigned PUT URLs sign
  `Content-Type` (and clients may send additional headers like
  `x-amz-*`); the signature still binds key + content-type + expiry.

### Size bound (unchanged residual from CHG-0028)

- `EvidenceMaxBytes` (default 256 MB) is enforced at **grant issue** only.
  Presigned PUT cannot bound actual bytes at S3 level (no POST policy in
  SDK v2). Raising 50 MB → 256 MB widens the oversize window but does not
  change the residual class (Medium, accepted): TTL 5min, contract-scoped
  prefix, 365-day expiry, SSE-S3, no public access.

### Test credentials

- `s3_client.go` static credentials (`test`/`testtest`) apply only when
  `S3_ENDPOINT_URL` is set — local/dev emulator path. Lambda/production
  never sets it; `cmd/lambda/main.go` only wires the endpoint branch when
  the env var is present. No real credentials in code.

### IAM / bucket posture (unchanged)

- Function IAM still scoped to `arn:aws:s3:::bucket/evidence/*`.
- `PublicAccessBlockConfiguration` all true; versioning, AES256,
  lifecycle unchanged.

## Residual Risks

| Risk | Level | Status |
| --- | --- | --- |
| Presigned PUT does not enforce byte size at S3 | Medium | Accepted (mitigations documented; reconsider when real uploads prove abuse) |
| `AllowedHeaders: *` on CORS rule | Low | Accepted (required by signed headers; signature binds key/type/TTL) |
| New origin values must be reviewed per environment | Low | Parameter default is stg+localhost; prod deploy supplies prod origin |

No Critical/High findings. Approved for consolidation.
