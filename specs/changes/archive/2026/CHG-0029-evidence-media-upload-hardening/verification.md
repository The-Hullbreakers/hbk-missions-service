---
change_id: CHG-0029
artifact: verification
status: verified
updated: 2026-10-09
---

# Verification — Evidence Media Upload Hardening

## Wave 1 — IaC + Test Infrastructure

- T-01 done: `CorsConfiguration` on `MissionsEvidenceBucket` — `PUT`/`GET`/`HEAD`, origins from `EvidenceCorsOrigins` (CommaDelimitedList, default stg app + localhost:3000), `AllowedHeaders: ["*"]`, `ExposedHeaders: [ETag]`, `MaxAge: 3000`.
- T-02 done: `EvidenceMaxBytes` Number param (default 268435456 = 256 MB) → `EVIDENCE_MAX_BYTES` env var replaces the hardcoded 50 MB.
- T-03 done: `localstack-s3` service in `docker-compose.test.yml` (`localstack/localstack:3.0.2`, `SERVICES=s3`, port 14566→4566). Note: `latest` is license-gated — pinned 3.0.2 community tag.

## Wave 2 — Adapter Contract Test

- T-04 done: `tests/integration/infrastructure/s3/evidence_store_test.go` — `TestEvidenceStore_RoundTrip` (presign PUT → real HTTP PUT with signed Content-Type → `Exists` true → presign GET → byte-equal download) and `TestEvidenceStore_ExistsPhantom` (phantom key → false, no error). Both PASS against localstack: `ok tests/integration/infrastructure/s3`.
- `s3_client.go` static test credentials aligned to emulator (`test`/`testtest`); only used when `S3_ENDPOINT_URL` is set.

## Gates

- GATE-SECURITY (CORS trust boundary): PASS — `security-review.md`, no Critical/High findings; origins never `*`, CORS-as-browser-control documented.
- GATE-CRITICAL-E2E: PASS — `TestEvidenceMediaJourney_EndToEnd` green (regression only, journey unchanged).
- GATE-VALIDATE (`tools/validate-sdd-change.sh`): PASS — gofmt clean, vet clean, unit + integration + race all green.
- Coverage: PASS — domain 91.4%, application 91.9%, dynamodb 88.1% (no regression; new s3 adapter exercised by contract test).
- Mutation: N/A (no non-trivial production logic per L2 classification).
- BLUE: PASS — all touched files < 150 lines; no unformatted files.
