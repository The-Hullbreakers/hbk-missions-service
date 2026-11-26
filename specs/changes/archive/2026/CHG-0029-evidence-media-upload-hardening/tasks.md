---
change_id: CHG-0029
artifact: tasks
status: planned
updated: 2026-10-09
---

# Tasks — Evidence Media Upload Hardening

Execution mode: sequential, one agent. No domain/application RED — the
production change is IaC configuration; the only code surface is the new
infrastructure contract test, which is GREEN-on-arrival against real
emulator behavior.

## Wave 1 — IaC + Test Infrastructure

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-01 | `CorsConfiguration` on evidence bucket + `EvidenceCorsOrigins` param | interface | — | `sam validate` / template lint |
| T-02 | `EvidenceMaxBytes` param → `EVIDENCE_MAX_BYTES` (default 268435456) | composition | — | same |
| T-03 | `localstack` (S3) service in `docker-compose.test.yml` + bucket bootstrap | infrastructure | — | `docker compose -f docker-compose.test.yml up -d --wait` |

## Wave 2 — Adapter Contract Test

| ID | Task | Owner | Depends | Verification |
| --- | --- | --- | --- | --- |
| T-04 | `tests/integration/infrastructure/s3/` contract test: presign PUT → real upload → `Exists` → presign GET → byte-equal download; phantom `Exists` false | infrastructure | T-03 | `go test ./tests/integration/infrastructure/s3/` |

RED evidence: none required (no production logic added); T-04 is a new
test against a previously untested adapter — initial run against the
emulator demonstrates real behavior rather than a fail-first stub.

## Wave 3 — Gates + Docs

| ID | Task |
| --- | --- |
| T-05 | BLUE — file sizes < 150 lines, gofmt/vet clean |
| T-06 | Security gate — CORS trust-boundary evidence in `security-review.md` |
| T-07 | Docs — `API-GUIDE-FRONTEND.md` CORS/origin note + size bound update |
| T-08 | `tools/validate-sdd-change.sh` + fill `verification.md` |
| T-09 | Consolidate canonical spec, index, archive |

Mutation gate: N/A (no non-trivial production logic). Critical E2E:
existing `TestEvidenceMediaJourney_EndToEnd` must stay green — regression
only, no new journey.
