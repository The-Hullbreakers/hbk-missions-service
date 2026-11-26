# CHG-0055 — Create local season preview scenarios: security review

security_review_id: SEC-0055-001
security_review_title: Protect local identities and existing datasets
change_id: CHG-0055-local-season-preview-scenarios
status: reviewed
security_role: none

Trust boundaries: command-line input to local DynamoDB Local; read-only
identity metadata to synthetic mission references; script-generated
contract/credit rows to an isolated missions preview table.
No OAuth server, resource-server policy, tokens, identity records, or
web authentication contract changes.

Controls to verify: strict HTTP loopback endpoint; fixed target-table
allowlist; dummy credentials; no access to cloud AWS profiles; no
passwords, email, Discord identifiers, display names, or tokens in
fixtures/logs; active users only; no identity writes; comparison and
conditional inserts, not replacement/deletion; unchanged original
missions tables. Scope review and specific commands/findings must be
filled after executable evidence. Critical/High unresolved findings
block completion.

## Findings

- BUG-0055-001 — Reject empty-userinfo preview endpoints /
  REG-0055-001 — Empty userinfo must be rejected (2026-11-23):
  `validate_target` accepted `http://@localhost:8000` because an
  empty username is falsy, bypassing the "absence of userinfo"
  control on a trust boundary (CLI input -> local DynamoDB Local).
  Severity: guard defect, local-only scope. Fixed by scoped Gate 3
  approval: `url.username`/`url.password` now checked with
  `is not None` in `scripts/season_scenario_model.py`. Verified by
  regression subTest plus full unit (10/10) and local-resource
  integration (2/2) suites; see verification.md GREEN/BLUE
  BUG-0055-001 section. Resolved; no Critical/High findings
  outstanding from this defect.

## Final scoped review

Reviewed 2026-11-23 against the five new Python scripts, the
security regression and local integration evidence, and the
persistent preview tables. Role `none`: the tool changes no
OAuth/OIDC, API authorization or browser session. The CLI validates
HTTP loopback and fixed table names before invoking AWS CLI; it
overrides ambient AWS credentials with dummy local values, rejects
userinfo (including empty userinfo), paths, queries and fragments,
and never accepts a table name argument. Only active user IDs are
read from `hbk-users`; no names, emails, IDs, tokens or credentials
are printed. Requested rows are compared with every existing row;
transactions insert only absent keys conditionally. A conflicting
row aborts without overwrite, and the original missions and
identity tables have no write path.

Threats considered: remote endpoint/cloud exfiltration, hostile
endpoint syntax, accidental overwrite, inactive-member references
and sensitive-data logging. The endpoint regression was fixed;
Python unit 14/14, local integration 2/2 and real Go-reader and
HTTP preview checks passed. No known Critical or High security
finding remains in this scoped local-only tool. This is not an
independent security-owner sign-off and does not validate real OIDC.
