---
change_id: CHG-0028
artifact: verification
status: verified
updated: 2026-10-09
---

# Verification — Evidence Media Upload

## Gate 3-DOMAIN — Wave 1 (T-01/T-02/T-03)

- **gate_scope**: domain; prior layer gates: `not_affected` (first layer).
- **Tests**: `TestMediaURI_*` (TEST-DOMAIN-MEDIAURI-VALID,
  TEST-DOMAIN-MEDIAURI-FOREIGN), `TestMissionContract_MediaPrefix`,
  `TestAllowedMediaType_*`, `TestMissionContract_ConfirmWithMedia_*` —
  `tests/unit/missions/domain/media_uri_test.go`.
- **test_type/layer**: domain unit; lowest level for VO validation and
  credit field behavior. No doubles — real domain objects only.
- **standalone_test_command**:
  `go test ./tests/unit/missions/domain/ -run 'TestMediaURI|TestAllowedMediaType|TestMissionContract_MediaPrefix|TestMissionContract_Confirm'`
- **RED result**: FAIL — build failed; `undefined: domain.NewMediaURI`,
  `domain.ErrEvidenceMediaInvalid` (compile-time RED for missing symbols).
- **Mapped production files**: `internal/missions/domain/errors.go`,
  `media_uri.go` (new), `mission_contract.go`, `merit_credit.go`,
  `mission_contract_confirm.go`.
- **No production code changed before RED.** Harness files: none new.
- **Decision**: approved 2026-10-09 — GREEN authorized for
  `internal/missions/domain/` scope.
- **GREEN**: `media_uri.go` (new: `MediaURI`, `NewMediaURI`,
  `AllowedMediaType`, `MissionContract.MediaPrefix`),
  `ErrEvidenceMediaInvalid` in `errors.go`, `MeritCredit.MediaURI`,
  `ConfirmWithMedia` extracted from `Confirm`. Design note: new method
  instead of signature change — existing callers untouched.
  `go test ./tests/unit/missions/domain/` → ok.
- **BLUE**: gofmt/vet clean; sizes 35–114 lines; `go build ./...` ok.

## Gate 3-APPLICATION — Wave 2 (T-04/T-05/T-06/T-07)

- **gate_scope**: application; prior layer gate: `passed` (domain).
- **Tests**: `TEST-APP-UPLOAD-GRANT-MEMBER`, `TEST-APP-UPLOAD-NOT-IN-PROGRESS`,
  `TEST-APP-UPLOAD-OUTSIDER`, `TEST-APP-UPLOAD-TYPE-REJECTED`,
  `TEST-APP-CONFIRM-MEDIA-STORED`, `TEST-APP-CONFIRM-MEDIA-PHANTOM`,
  `TEST-APP-CONFIRM-MEDIA-FOREIGN`, `TEST-APP-MEDIA-READ-MODERATOR`,
  `TEST-APP-MEDIA-READ-DENIED` + foreign-URI case —
  `tests/unit/missions/application/evidence_media_test.go`.
- **test_type/layer**: application unit with hand-written port double
  (`fakeEvidenceStore`); no assertions or rules inside the double.
- **standalone_test_command**:
  `go test ./tests/unit/missions/application/`
- **RED result**: FAIL — build failed; `undefined: application.UploadGrant`,
  `application.ReadGrant`, `application.NewRequestEvidenceUploadHandler`,
  `application.NewConfirmObjectiveHandlerWithEvidence`,
  `application.NewRequestEvidenceReadHandler`, `domain.ErrMediaTypeNotAllowed`.
- **Mapped production files**:
  `internal/missions/application/evidence_media_ports.go` (new),
  `request_evidence_upload.go` (new), `request_evidence_read.go` (new),
  `confirm_objective.go`; plus `domain/errors.go` for
  `ErrMediaTypeNotAllowed` (maps to `400 INVALID_REQUEST` at boundary).
- **Design note**: `ConfirmObjectiveHandler` gains a variant constructor +
  `HandleWithMedia`; `Handle` stays unchanged for existing callers.
- **No production code changed before RED.**
- **Decision**: approved 2026-10-09 — GREEN authorized for
  `internal/missions/application/` scope.
- **GREEN**: `evidence_media_ports.go` (`UploadGrant`, `ReadGrant`,
  `EvidenceUploadPresigner`, `EvidenceObjectReader`),
  `request_evidence_upload.go`, `request_evidence_read.go`,
  `ConfirmObjectiveHandler` + `evidence`/`HandleWithMedia`,
  `domain.ErrMediaTypeNotAllowed`. `go test
  ./tests/unit/missions/application/` → ok.
- **BLUE**: split test file into `evidence_upload_test.go` +
  `evidence_confirm_media_test.go` (<150 lines each); gofmt/vet clean;
  sizes 27–130 lines; `go build ./...` ok.

## Gate 3-BOUNDARY — Wave 3 (T-08/T-09/T-10)

- **gate_scope**: boundary; prior layer gates: `passed` (domain,
  application).
- **Tests**: `TEST-HTTP-UPLOAD-201`, `TEST-HTTP-UPLOAD-400`,
  `TEST-HTTP-UPLOAD-403`, `TEST-HTTP-UPLOAD-422`,
  `TEST-HTTP-MEDIA-READ-200`, `TEST-HTTP-MEDIA-READ-403`,
  `TEST-HTTP-EVIDENCE-MEDIA-201`, `TEST-HTTP-EVIDENCE-MEDIA-422` —
  `tests/integration/http/evidence_media_test.go`. Exercises the real HTTP
  entry point, `auth.Authenticate`, and a hand-written blob double.
- **test_type/layer**: public-boundary integration; identity derives from
  the session double (`roleSessionFake`), never the body.
- **standalone_test_command**: `go test ./tests/integration/http/`
- **RED result**: FAIL — build failed;
  `undefined: missionhttp.RequestEvidenceUploadHandler`,
  `missionhttp.RequestEvidenceReadHandler`.
- **Mapped production files**:
  `internal/missions/interface/http/evidence_upload.go` (new),
  `evidence_read.go` (new), `handler.go` (`media_uri` on evidenceRequest +
  `HandleWithMedia` dispatch), `errors.go` (`EVIDENCE_MEDIA_INVALID`).
- **No production code changed before RED.**
- **Decision**: approved 2026-10-09 — GREEN authorized for
  `internal/missions/interface/http/` scope.
- **GREEN**: `evidence_upload.go` (201 grant, 400/403/422 mapping),
  `evidence_read.go` (200 grant, 403/422 mapping), `handler.go` dispatches
  to `HandleWithMedia` when `media_uri` present, `errors.go` maps
  `EVIDENCE_MEDIA_INVALID`. `go test ./tests/integration/http/` → ok
  (after fixing test double: roles must live in `claims["roles"]` for
  `CallerIdentityFromContext`).
- **BLUE**: `newObjectID` simplified to `crypto/rand.Text()` (no error
  branch); error-branch coverage tests moved to
  `evidence_media_errors_test.go`; all files <150 lines; gofmt/vet clean.

## Wave 3 — Infrastructure (T-09/T-10)

- **T-10**: `creditItem` writes `MediaURI` only when non-empty.
  `TestCreditItem_PersistsMediaURI` + `TestCreditItem_OmitsMediaURIWhenEmpty`
  against dynamodb-local → ok.
- **T-09**: `internal/missions/infrastructure/s3/evidence_store.go` +
  `s3_client.go` — presigned PUT (key + ContentType bound), presigned GET
  15min, `HeadObject` for `Exists`. aws-sdk-go-v2 has no POST-policy
  presign → declared-size check `ErrEvidenceUploadTooLarge` at grant
  issuance (400). New dep: `service/s3 v1.114.2`.

## Wave 4 — Composition + E2E (T-11/T-12)

- `Config.Evidence`/`EvidenceMaxBytes`; routes registered when store
  provided; `cmd/*` build `EvidenceStore` when `EVIDENCE_BUCKET_NAME` set.
- `template.yaml`: `MissionsEvidenceBucket` (private, AES256, versioning,
  lifecycle: abort multipart 7d + expire evidence/ 365d, Retain),
  `EVIDENCE_BUCKET_NAME`/`EVIDENCE_MAX_BYTES` env, IAM scoped to
  `evidence/*` (Put/Get/Delete), 2 API Gateway routes.
- **TEST-E2E-EVIDENCE-MEDIA-JOURNEY**
  (`TestEvidenceMediaJourney_EndToEnd`): season→definition→publish→offer→
  accept→start→upload-url(201)+400-bad-type→evidence+media(201, echo)+
  422-phantom→media-url 200-reviewer/403-member → ok with in-memory
  `blobStoreFake`.

## Gate Results

- **GATE-VALIDATE** `tools/validate-sdd-change.sh`: PASS (gofmt, vet,
  unit, integration, race).
- **GATE-COVERAGE**: domain 91.4%, application 91.9%, dynamodb 88.1% —
  all new-code functions ≥90% (newObjectID 100% after BLUE).
- **GATE-CRITICAL-E2E**: evidence media journey passed.
- **GATE-MUTATION**: no mutation tooling installed (gremlins/ooze
  absent); manual spot-mutation performed 2026-10-09 — three mutants,
  all killed by tests, all reverted to green:
  `NewMediaURI` prefix check (`!ok`→`ok`) → FAIL;
  `AllowedMediaType` (`||`→`&&`) → FAIL;
  `isReviewer` (`||`→`&&`) → FAIL on moderator-grant test.
- **GATE-SECURITY**: see security-review.md — no Critical/High findings;
  one Medium documented (PUT presign lacks S3-level length cap).

## Known Evidence Gaps

- S3 presign POST policy is not supported by aws-sdk-go-v2; the adapter
  will emit presigned PUT URLs (key + Content-Type bound in the signature).
  Residual gap: no S3-level `content-length-range` enforcement; declared
  size is checked at grant issuance. Documented in security-review.
- No localstack S3 in `docker-compose.test.yml`; adapter contract tests
  cover generated grant fields; real-S3 verification is on-deploy
  (rollout.md).
