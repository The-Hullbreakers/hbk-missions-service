---
change_id: CHG-0029
artifact: delta
status: consolidated
updated: 2026-10-09
---

# Delta — Evidence Media Upload Hardening

## What Is Changing

### Mission Contract Fulfillment

- The evidence bucket gains a CORS configuration so browser clients can
  execute the presigned `PUT` upload; origins are explicit per environment.
- The declared-size bound becomes a deployment parameter defaulting to
  256 MB instead of a fixed 50 MB, so video evidence fits.
- CI gains an S3 emulator; the evidence-store adapter is verified against
  real presign/`HeadObject`/download behavior.

## ADDED

### REQ-MCF-017 — Browser clients can execute the upload grant

The evidence bucket allows cross-origin `PUT` and `HEAD` (and `GET` for
reviewer read grants) only from the configured application origins via
bucket CORS configuration. Credentials and object keys stay unchanged;
CORS governs browser enforcement, not authorization.

#### SCN-MCF-CORS-UPLOAD — Webapp PUT reaches S3

Given a webapp origin listed in the bucket CORS configuration, when a
browser issues the presigned `PUT` for `evidence/{contract_id}/{id}`, then
the preflight and the upload succeed and the object lands under the grant
key.

#### SCN-MCF-CORS-UNKNOWN-ORIGIN — Unknown origin blocked by browser

Given an origin not in the CORS allowlist, when a page on that origin
attempts the presigned `PUT`, then the browser blocks the response
(no `Access-Control-Allow-Origin` for that origin) and the request never
reaches the service.

### REQ-MCF-018 — Evidence store adapter verified against an emulator

CI runs a `localstack` S3 service in `docker-compose.test.yml`; the
infrastructure contract test performs presign-PUT → upload → `Exists` →
presign-GET round-trips against it, including the phantom-key negative.

#### SCN-MCF-ADAPTER-ROUNDTRIP — Grant, upload, read back

Given the emulator bucket exists, when the adapter issues an upload grant
and a client PUTs bytes to it, then `Exists` returns true and a read
grant downloads identical bytes.

## MODIFIED

### REQ-MCF-014 — Request an evidence media upload grant

The declared-size bound is now governed by the `EvidenceMaxBytes`
deployment parameter (default `268435456` = 256 MB) passed through
`EVIDENCE_MAX_BYTES`, replacing the fixed 50 MB default. All other grant
semantics are unchanged.

## Boundary Changes

- `template.yaml`: `CorsConfiguration` on the evidence bucket; new
  `EvidenceCorsOrigins` and `EvidenceMaxBytes` parameters.
- `docker-compose.test.yml`: new `localstack` service exposing S3 on
  a dedicated port; test bootstrap creates the bucket from
  `EVIDENCE_BUCKET_NAME`.
- No HTTP contract change; no domain/application change.

## Persistence Changes

None — bucket configuration only; no schema or item-shape change.
