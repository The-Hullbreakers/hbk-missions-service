---
change_id: CHG-0031
artifact: security-review
status: passed
updated: 2026-10-09
security_review_id: SEC-0031
security_review_title: Reviewer adjudication trust boundary
identity_role: resource-server
---

# Security Review — Mission Completion Moderation

## Role

Declared identity role: `resource-server` — the service protects the new
adjudication endpoints with the existing OIDC session; it does not issue
identity. Focus: the `admin`/`moderator`-only adjudication boundary and
credit materialization.

## Trust-Boundary Analysis

- **New endpoints** `POST .../submit`, `POST .../approve`,
  `POST .../reject` sit behind `auth.Authenticate` — unauthenticated
  callers get `401`; identity comes from the session, never the body.
- **Submit authority:** roster membership, non-departed, `IN_PROGRESS`
  contract — same rule as evidence submission.
- **Adjudication authority:** `approve`/`reject` require
  `admin`/`moderator` role *within the contract's tenant* — roster
  members without the role cannot self-approve (SCN-MCF-REVIEW-MEMBER-
  DENIED). Cross-tenant adjudication is impossible: the tenant is derived
  from the contract, and role checks ride on session claims.
- **Credit materialization is server-side:** approval writes credits the
  domain staged at confirm time; the reviewer can only release or discard
  — never edit amounts, recipients, or categories.
- **No ledger write on reject:** `DISQUALIFIED` transitions are contract-
  only writes; zero ledger items means no reversal path to abuse.
- **Idempotent approval:** conditional writes keyed by `ConfirmationID`
  make replayed approvals no-ops, not duplicate credits.

## Data Exposure

- `pending_credits` on the contract item carries the same attribute class
  as existing credit data (ids, amounts, media_uri) — tenant-scoped, no
  new sensitivity.
- Reviewer decisions are recorded as terminal state (`COMPLETED` /
  `DISQUALIFIED`) in contract history; no PII added.

## Findings

- Role checks verified: `isReviewer` admits only `admin`/`moderator`
  (`TestApproveContract_MemberCannotAdjudicate`,
  `TestRejectContract_MemberCannotAdjudicate`, boundary
  `TestContractReview_ApproveEndpointRejectsMember`); approve/reject routes
  additionally sit behind `adminOrModerator` middleware. Member
  self-adjudication denied — mutation evidence: removing the role guard
  kills both app tests.
- Wrong-state guards verified: `ErrContractNotPendingValidation` →
  `CONTRACT_NOT_PENDING_VALIDATION` (422); terminal contracts reject
  further adjudication (E2E `TestModerationJourney_RejectPaysNothing`).
- Atomicity verified: `SaveApproved` commits contract + roster release +
  N conditional credit puts in one `TransactWriteItems`; retried approval
  maps to `ErrDuplicateConfirmation` (`TestContractStore_SaveApprovedRejectsDuplicateRetry`).
- Zero-credit rejection verified end-to-end: staged credit never reaches
  the leaderboard before approval and stays empty after reject.
- **Residual (documented):** no review-queue listing endpoint — reviewers
  need contract ids from another channel; follow-up if moderation volume
  requires discovery.
- **Status**: passed — no Critical/High findings (2026-10-12).
