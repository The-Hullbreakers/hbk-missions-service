---
change_id: CHG-0054
artifact: security-review
status: pending
updated: 2026-11-01
---

# Security Review — CHG-0054

## Trust Boundaries

1. **Language labeling integrity**: `content_language` and the
   `Content-Language` header must describe the text actually served. A
   payload that serves Spanish under an `en` label — or a header
   claiming one language over a mixed collection — is a findings-class
   defect, not a cosmetic bug: clients render Markdown on that label.
2. **Fail-closed query validation**: `lang` is parsed to a fixed enum
   before any use case runs; invalid, empty, or conflicting repeats
   resolve `400 INVALID_QUERY_PARAMETER`. No locale string flows into
   queries, keys, or templates.
3. **Authorization unchanged**: `lang` selects rendering only. Roster
   membership, adjudicator role, member-vs-reviewer projection
   filtering (`pending_credits`, `media_uri`), and `403 FORBIDDEN`
   paths are exercised with `lang` present — a language parameter that
   widens or loosens access is a High finding.
4. **Authored content trust**: `en` translations enter only through
   the existing authenticated, authorized operator path — reviewed
   text supplied by operators holding the definition-authoring role,
   never by members or anonymous callers; the publish completeness
   rule prevents partial locale sets. Markdown rendering
   responsibility stays client-side as today — no new injection
   surface, but translation writes must flow through the authorized
   operator flow (or the reviewed backfill), not ad-hoc edits.
5. **Backfill integrity**: the augmentation path writes `en` attributes
   only and rejects records whose identity (version, objective ids,
   rule count) mismatches the translation. A backfill that mutates `es`
   content, economics, or unmatched records is a Critical finding; the
   rehearsal produces a diff + unresolved-record report as evidence.
   The backfill runs under operator control with no live-migration
   authorization from this change.
6. **Publish cutover**: the completeness gate rejects publishes that
   previously succeeded; enforcement must not activate before
   authoring tooling and reviewed translations are in place
   (`rollout.md`), or legitimate operator publishes fail — a
   deployment-sequencing control, verified at the staged cutover.
7. **Catalog redeploy** (Gate 1 amendment 2026-11-01): CI synchronizes
   racing after every deployment. A same-version `Put` that drops an
   already stored `ContentEn`, accepts changed Spanish/economics, or
   pairs an old offer locale with a new version breaks historical
   integrity. The translation sidecar is untrusted draft input until
   exact Spanish source/version binding and operator content review
   are proven. The original Gate 1 did not expand IAM permissions.
8. **Exact-offer seed read** (Gate 1 amendment 2026-11-03):
   safely detecting Spanish, merit, and English drift in frozen
   offers before a dry-run/apply requires reading the exact
   seed-owned offer row. The deployment role currently has
   `Query`/`ConditionCheckItem` on DEFINITION/SEASON but only
   `PutItem`/`UpdateItem` on OFFER. Grant `GetItem` on OFFER keys
   only, restricted by `dynamodb:LeadingKeys`, on the existing
   MissionsTable ARN. Query/Scan access to OFFER, any read of
   CONTRACT, and new write actions would exceed this approval.
   Tests reject a widened policy. The separately approved Gate
   3-IAM added only that statement in source and its static policy
   test passed 3/3; no live role deployment or shared-table read
   has been authorized. `LeadingKeys` constrains PK, **not SK**:
   IAM permits `GetItem` for any OFFER row; the seed code limits
   its own reads to exact `seed-<code>-<season hash>` keys.
   Do not present this as an IAM-enforced seed-only SK restriction.
   The user accepted this exact PK-only scope on 2026-11-03 for
   local publication; this does not deploy the role or waive
   the pending security gate for non-local rollout.
   The full security gate remains pending.
9. **Conditional seed offer update**: the offline seed reads only
   exact PK/SK OFFER items on a consistent read, then applies a
   conditional update that checks every observed field, including
   `ContentEn` when present or its absence otherwise. A conflicting
   prior-version Spanish/English/merit snapshot is rejected
   before a new version can replace an offer. The local race
   test proves an intervening English edit aborts without
   overriding it. This is local evidence, not a live IAM or
   full-service security-gate pass.
10. **Local backfill invocation** (Gate 1 2026-11-03): the
    approved English sidecars and a local-only operator action
    cross a data-mutation boundary. Gate 2 must prove remote
    endpoint refusal before client creation, manifest completeness,
    a read-only preview, and unresolved-record refusal before any
    apply. Gate 3 may reuse the existing conditional `ContentEn`
    adapter on isolated private tables. The user approved no
    local application-table write yet; apply needs its own
    explicit authorization and a fresh preview. The adapter
    is not globally atomic across rows: report partial successes
    and conflicts rather than inventing an unsafe rollback.
    The local CLI GREEN now enforces an explicit HTTP
    `localhost`/`127.0.0.1`/`::1` endpoint with port before
    creating a client; the SDK uses static test credentials.
    Unit endpoint tests and five isolated CLI scenarios pass.
    No non-local access, real-table write, or full L3 security
    signoff is implied by those results.
    After a separate exact-table user approval on 2026-11-04,
    the CLI conditionally added `ContentEn` to 108 source-matching
    local definition/offer rows. The fresh preflight had zero
    unresolved and post-apply raw comparisons verified no
    other field, season, or contract changed. No IAM policy
    was deployed and final L3 security signoff remains open.

## Findings

| Severity | Finding | Status |
| --- | --- | --- |
| — | none | closed 2026-11-16 — labeling-integrity and fail-closed `lang` validation verified by localized HTTP suites + TEST-E2E-LANG-001; backfill applied under separate exact-table approval with strongly consistent verification |

No Critical/High findings may remain open at consolidation.
